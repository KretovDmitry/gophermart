package main

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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
	// Server run context.
	serverCtx, serverStopCtx := context.WithCancel(context.Background())

	// Load application configurations.
	cfg := config.MustLoad()

	// Create root logger tagged with server version.
	logger := logger.New(cfg).With(serverCtx, "version", Version)

	db, err := sql.Open("pgx", cfg.DSN)
	if err != nil {
		logger.Errorf("failed to open the database: %s", err)
		os.Exit(1)
	}

	// Log every query to the database.
	db = sqldblogger.OpenDriver(cfg.DSN, db.Driver(), logger)

	// Check connectivity and DSN correctness.
	if err = db.Ping(); err != nil {
		logger.Errorf("failed to connect to the database: %s", err)
		os.Exit(1)
	}

	// Close connection.
	defer func() {
		if err = db.Close(); err != nil {
			logger.Error(err)
		}
		logger.Sync()
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
	// // Init repository for auth service
	// authRepo, err := auth.NewRepository(db, logger)
	// if err != nil {
	// 	logger.Errorf("failed to create auth repository: %s", err)
	// 	os.Exit(1)
	// }
	//
	// authService, err := auth.NewService(authRepo, logger, cfg)
	// if err != nil {
	// 	logger.Error("failed to init auth service")
	// 	os.Exit(1)
	// }

	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	// handler := banner.HandlerWithOptions(bannerService, banner.ChiServerOptions{
	// 	BaseRouter:       router,
	// 	ErrorHandlerFunc: banner.ErrorHandlerFunc,
	// })

	// Build HTTP server.
	hs := &http.Server{
		Addr:              cfg.HTTPServer.Address,
		ReadHeaderTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:       cfg.HTTPServer.IdleTimeout,
		// Handler: handler,
	}

	// Graceful shutdown.
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT,
		syscall.SIGTERM, syscall.SIGQUIT, os.Interrupt)
	go func() {
		<-sig

		logger.Infof("shutting down server with %s timeout",
			cfg.HTTPServer.ShutdownTimeout)

		if err = hs.Shutdown(serverCtx); err != nil {
			logger.Errorf("graceful shutdown failed: %v", err)
		}
		serverStopCtx()
	}()

	// Start the HTTP server with graceful shutdown.
	logger.Infof("server %v is running at %v", Version, cfg.HTTPServer.Address)
	if err = hs.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Errorf("run server failed: %v", err)
	}

	// Wait for server context to be stopped.
	select {
	case <-serverCtx.Done():
	case <-time.After(cfg.HTTPServer.ShutdownTimeout):
		logger.Error("graceful shutdown timed out.. forcing exit")
	}
}
