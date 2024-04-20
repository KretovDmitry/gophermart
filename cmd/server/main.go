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
	"github.com/KretovDmitry/gophermart-loyalty-service/pkg/logger"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/jackc/pgx/v5/stdlib"
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

	// // Init repository for banner service
	// repo, err := banner.NewRepository(db, rdb, logger, cfg)
	// if err != nil {
	// 	logger.Errorf("failed to create banner repository: %s", err)
	// 	os.Exit(1)
	// }
	//
	// // Init service
	// bannerService, err := banner.NewService(repo, logger, cfg)
	// if err != nil {
	// 	logger.Error("failed to init banner service")
	// 	os.Exit(1)
	// }
	// // Do not loose banners being asynchronously deleted
	// defer bannerService.Stop()
	//
	// Init repository for auth service
	authRepo, err := auth.NewRepository(db, logger)
	if err != nil {
		return fmt.Errorf("failed to create auth repository: %w", err)
	}

	authService, err := auth.NewService(authRepo, logger, cfg)
	if err != nil {
		return fmt.Errorf("failed to init auth service: %w", err)
	}

	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	handler := auth.HandlerWithOptions(authService, auth.ChiServerOptions{
		BaseURL:          "/api/user",
		BaseRouter:       router,
		ErrorHandlerFunc: auth.ErrorHandlerFunc,
	})

	// Build HTTP server.
	hs := &http.Server{
		Addr:              cfg.HTTPServer.Address,
		ReadHeaderTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:       cfg.HTTPServer.IdleTimeout,
		Handler:           handler,
	}

	// Graceful shutdown.
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT,
			syscall.SIGTERM, syscall.SIGQUIT, os.Interrupt)

		<-sig

		logger.Infof("Shutting down server with %s timeout",
			cfg.HTTPServer.ShutdownTimeout)

		if err = hs.Shutdown(serverCtx); err != nil {
			logger.Errorf("graceful shutdown failed: %w", err)
		}
		serverStopCtx()
	}()

	// Start the HTTP server with graceful shutdown.
	logger.Infof("server %v is running at %v", Version, cfg.HTTPServer.Address)
	if err = hs.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("run server failed: %w", err)
	}

	// Wait for server context to be stopped.
	select {
	case <-serverCtx.Done():
	case <-time.After(cfg.HTTPServer.ShutdownTimeout):
		return errors.New("graceful shutdown timed out.. forcing exit")
	}

	return nil
}
