package reward

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/KretovDmitry/gophermart-loyalty-service/internal/luhn"
	"github.com/KretovDmitry/gophermart-loyalty-service/internal/models/errs"
	"github.com/go-chi/chi/v5"
)

type PostOrderParams struct {
	Number int
}

// ServerInterface represents all server handlers.
type ServerInterface interface {
	// Order upload (POST /api/user/orders)
	CreateOrder(w http.ResponseWriter, r *http.Request, params PostOrderParams)
}

// ServerInterfaceWrapper converts payloads to parameters.
type ServerInterfaceWrapper struct {
	Handler            ServerInterface
	ErrorHandlerFunc   func(w http.ResponseWriter, r *http.Request, err error)
	HandlerMiddlewares []MiddlewareFunc
}

type MiddlewareFunc func(http.Handler) http.Handler

// Login operation middleware.
func (siw *ServerInterfaceWrapper) UploadOrder(w http.ResponseWriter, r *http.Request) {
	// ------------- Required text/plain content type -----------------

	contentType := r.Header.Get("Content-Type")
	if !isTextPlainContentType(contentType) {
		siw.ErrorHandlerFunc(w, r, fmt.Errorf("%w: %s", errs.ErrInvalidContentType, contentType))
		return
	}

	// ------------- Parse and validate request body params ----------

	var params PostOrderParams

	defer r.Body.Close()
	data, err := io.ReadAll(r.Body)
	if err != nil {
		siw.ErrorHandlerFunc(w, r, err)
		return
	}

	if err = luhn.Validate(string(data)); err != nil {
		siw.ErrorHandlerFunc(w, r, fmt.Errorf("%w: %w", errs.ErrInvalidPayload, err))
		return
	}

	number, err := strconv.Atoi(string(data))
	if err != nil {
		siw.ErrorHandlerFunc(w, r, err)
		return
	}

	params.Number = number

	handler := http.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.CreateOrder(w, r, params)
	}))

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler.ServeHTTP(w, r)
}

// Handler creates http.Handler with routing matching spec.
func Handler(si ServerInterface) http.Handler {
	return HandlerWithOptions(si, ChiServerOptions{})
}

type ChiServerOptions struct {
	ErrorHandlerFunc func(w http.ResponseWriter, r *http.Request, err error)
	BaseRouter       chi.Router
	BaseURL          string
	Middlewares      []MiddlewareFunc
}

// HandlerFromMux creates http.Handler with routing matching spec.
func HandlerFromMux(si ServerInterface, r chi.Router) http.Handler {
	return HandlerWithOptions(si, ChiServerOptions{
		BaseRouter: r,
	})
}

func HandlerFromMuxWithBaseURL(si ServerInterface, r chi.Router, baseURL string) http.Handler {
	return HandlerWithOptions(si, ChiServerOptions{
		BaseURL:    baseURL,
		BaseRouter: r,
	})
}

// HandlerWithOptions creates http.Handler with additional options.
func HandlerWithOptions(si ServerInterface, options ChiServerOptions) http.Handler {
	r := options.BaseRouter

	if r == nil {
		r = chi.NewRouter()
	}
	if options.ErrorHandlerFunc == nil {
		options.ErrorHandlerFunc = func(w http.ResponseWriter, _ *http.Request, err error) {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	}
	wrapper := ServerInterfaceWrapper{
		Handler:            si,
		HandlerMiddlewares: options.Middlewares,
		ErrorHandlerFunc:   options.ErrorHandlerFunc,
	}

	r.Group(func(r chi.Router) {
		r.Post(options.BaseURL+"/orders", wrapper.UploadOrder)
	})

	return r
}

// isTextPlainContentType returns true if the content type is text/plain.
func isTextPlainContentType(contentType string) bool {
	contentType = strings.ToLower(strings.TrimSpace(contentType))
	if i := strings.Index(contentType, ";"); i > -1 {
		contentType = contentType[0:i]
	}
	return contentType == "text/plain"
}
