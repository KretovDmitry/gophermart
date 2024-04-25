package rest

import (
	"net/http"

	"github.com/KretovDmitry/gophermart-loyalty-service/pkg/accesslog"
	"github.com/KretovDmitry/gophermart-loyalty-service/pkg/logger"
	"github.com/KretovDmitry/gophermart-loyalty-service/pkg/unzip"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/nanmu42/gzip"
)

func InitChi(logger logger.Logger) *chi.Mux {
	router := chi.NewRouter()
	router.Use(accesslog.Handler(logger))
	router.Use(middleware.Recoverer)
	router.Use(gzip.DefaultHandler().WrapHandler)
	router.Use(unzip.Middleware(logger))

	return router
}

type (
	MiddlewareFunc func(http.Handler) http.Handler

	ChiServerOptions struct {
		BaseRouter  chi.Router
		BaseURL     string
		Middlewares []MiddlewareFunc
	}
)
