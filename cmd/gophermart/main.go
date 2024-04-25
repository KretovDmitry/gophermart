package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/KretovDmitry/gophermart-loyalty-service/internal/application/services"
	"github.com/KretovDmitry/gophermart-loyalty-service/internal/config"
	"github.com/KretovDmitry/gophermart-loyalty-service/internal/infrastructure/db/postgres"
	rest "github.com/KretovDmitry/gophermart-loyalty-service/internal/interface/api/rest/chi"
	"github.com/KretovDmitry/gophermart-loyalty-service/internal/interface/api/rest/middleware"
	"github.com/KretovDmitry/gophermart-loyalty-service/migrations"
	"github.com/KretovDmitry/gophermart-loyalty-service/pkg/logger"
	trmsql "github.com/avito-tech/go-transaction-manager/drivers/sql/v2"
	trmcontext "github.com/avito-tech/go-transaction-manager/trm/v2/context"
	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// Version indicates the current version of the application.
var Version = "1.0.0"

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	// Server run context.
	serverCtx, serverStopCtx := context.WithCancel(context.Background())
	defer serverStopCtx()

	// Load application configurations.
	cfg := config.MustLoad()

	// Create root logger tagged with server version.
	logger := logger.New(cfg).With(serverCtx, "version", Version)

	// Connect to postgres.
	db, err := postgres.Connect(cfg.DSN, logger)
	if err != nil {
		return fmt.Errorf("postgres connect: %w", err)
	}

	// Up all migrations for github tests.
	if err = migrations.Up(db, cfg); err != nil {
		return fmt.Errorf("run migrations: %w", err)
	}

	// Close connection.
	defer func() {
		if err = db.Close(); err != nil {
			logger.Error(err)
		}
		_ = logger.Sync()
	}()

	// Create default transaction manager for database/sql package.
	trManager := manager.Must(
		trmsql.NewDefaultFactory(db),
		manager.WithCtxManager(trmcontext.DefaultManager),
	)

	// Init repos.
	userRepo, err := postgres.NewUserRepository(db, trmsql.DefaultCtxGetter, logger)
	if err != nil {
		return fmt.Errorf("failed to init user repository: %w", err)
	}
	accountRepo, err := postgres.NewAccountRepository(db, trmsql.DefaultCtxGetter, logger)
	if err != nil {
		return fmt.Errorf("failed to init account repository: %w", err)
	}
	orderRepo, err := postgres.NewOrderRepository(db, trmsql.DefaultCtxGetter, logger)
	if err != nil {
		return fmt.Errorf("failed to order account repository: %w", err)
	}

	// Init services.
	authService, err := services.NewAuthService(userRepo, accountRepo, trManager, logger, cfg)
	if err != nil {
		return fmt.Errorf("failed to init auth service: %w", err)
	}
	accountService, err := services.NewAccountService(accountRepo, orderRepo, trManager, logger)
	if err != nil {
		return fmt.Errorf("failed to init account service: %w", err)
	}
	orderService, err := services.NewOrderService(orderRepo, logger)
	if err != nil {
		return fmt.Errorf("failed to init order service: %w", err)
	}

	// Create root router.
	router := rest.InitChi(logger)

	// Init and group handlers for auth routes.
	rest.NewAuthController(authService, cfg.JWT.Expiration, rest.ChiServerOptions{
		BaseURL:    "/api/user",
		BaseRouter: router,
	})

	// Init and group handlers for order routes.
	rest.NewOrderController(orderService, rest.ChiServerOptions{
		BaseURL:     "/api/user",
		BaseRouter:  router,
		Middlewares: []rest.MiddlewareFunc{middleware.Middleware(authService)},
	})

	// Init and group handlers for account routes.
	rest.NewAccountController(accountService, rest.ChiServerOptions{
		BaseURL:     "/api/user",
		BaseRouter:  router,
		Middlewares: []rest.MiddlewareFunc{middleware.Middleware(authService)},
	})

	// Build HTTP server.
	hs := &http.Server{
		Addr:              cfg.HTTPServer.Address,
		ReadHeaderTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:       cfg.HTTPServer.IdleTimeout,
		Handler:           router,
	}

	// Graceful shutdown.
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT,
			syscall.SIGTERM, syscall.SIGQUIT, os.Interrupt)

		signal := <-sig

		logger.With(serverCtx, "signal", signal.String()).
			Infof("Shutting down server with %s timeout",
				cfg.HTTPServer.ShutdownTimeout)

		if err = hs.Shutdown(serverCtx); err != nil {
			logger.Errorf("graceful shutdown failed: %s", err)
		}
		serverStopCtx()
	}()

	// Start the HTTP server with graceful shutdown.
	logger.Infof("Server %v is running at %v", Version, cfg.HTTPServer.Address)
	if err = hs.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("run server failed: %w", err)
	}

	// Wait for server context to be stopped or force exit if timeout exceeded.
	select {
	case <-serverCtx.Done():
	case <-time.After(cfg.HTTPServer.ShutdownTimeout):
		return errors.New("graceful shutdown timed out.. forcing exit")
	}

	return nil
}
