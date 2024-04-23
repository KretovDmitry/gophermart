package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/KretovDmitry/gophermart-loyalty-service/internal/auth"
	"github.com/KretovDmitry/gophermart-loyalty-service/internal/config"
	"github.com/KretovDmitry/gophermart-loyalty-service/internal/reward"
	"github.com/KretovDmitry/gophermart-loyalty-service/pkg/accesslog"
	"github.com/KretovDmitry/gophermart-loyalty-service/pkg/logger"
	"github.com/KretovDmitry/gophermart-loyalty-service/pkg/unzip"
	trmsql "github.com/avito-tech/go-transaction-manager/drivers/sql/v2"
	trmcontext "github.com/avito-tech/go-transaction-manager/trm/v2/context"
	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/nanmu42/gzip"
	sqldblogger "github.com/simukti/sqldb-logger"
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

	db, err := sql.Open("pgx", cfg.DSN)
	if err != nil {
		return fmt.Errorf("failed to open the database: %w", err)
	}

	// Log every query to the database.
	db = sqldblogger.OpenDriver(cfg.DSN, db.Driver(), logger)

	// Check connectivity and DSN correctness.
	if err = db.Ping(); err != nil {
		return fmt.Errorf("failed to connect to the database: %w", err)
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

	// Init repository for auth service.
	authRepo, err := auth.NewRepository(db, trmsql.DefaultCtxGetter, logger)
	if err != nil {
		return fmt.Errorf("failed to init auth repository: %w", err)
	}

	// Init auth service.
	authService, err := auth.NewService(authRepo, trManager, logger, cfg)
	if err != nil {
		return fmt.Errorf("failed to init auth service: %w", err)
	}

	// Init repository for reward service.
	rewardRepo, err := reward.NewRepository(db, trmsql.DefaultCtxGetter, logger)
	if err != nil {
		return fmt.Errorf("failed to init reward repository: %w", err)
	}

	// Init reward service.
	rewardService, err := reward.NewService(rewardRepo, trManager, logger, cfg)
	if err != nil {
		return fmt.Errorf("failed to init banner service: %w", err)
	}

	// Create root router.
	router := initRootRouter(logger)

	// Init and group handlers for auth routes.
	authHandlers := auth.HandlerWithOptions(authService, auth.ChiServerOptions{
		BaseURL:          "/api/user",
		BaseRouter:       router,
		ErrorHandlerFunc: auth.ErrorHandlerFunc,
	})

	// Init handlers for reward routes.
	rewHandlers := reward.HandlerWithOptions(rewardService, reward.ChiServerOptions{
		BaseURL:          "/api/user",
		BaseRouter:       router,
		Middlewares:      []reward.MiddlewareFunc{authService.Middleware},
		ErrorHandlerFunc: reward.ErrorHandlerFunc,
	})

	router.Handle("/api/user", authHandlers)
	router.Handle("/api/user", rewHandlers)

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

func initRootRouter(logger logger.Logger) *chi.Mux {
	router := chi.NewRouter()
	router.Use(accesslog.Handler(logger))
	router.Use(middleware.Recoverer)
	router.Use(gzip.DefaultHandler().WrapHandler)
	router.Use(unzip.Middleware(logger))

	return router
}
