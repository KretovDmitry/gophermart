package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/KretovDmitry/gophermart-loyalty-service/internal/models/errs"
	"github.com/go-chi/chi/v5"
)

// RegisterParams defines parameters for Register.
type RegisterParams struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// LoginParams defines parameters for Login.
type LoginParams struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// ServerInterface represents all server handlers.
type ServerInterface interface {
	// Registration (POST /api/user/register)
	Register(w http.ResponseWriter, r *http.Request, params RegisterParams)
	// Authentication (POST /api/user/login)
	Login(w http.ResponseWriter, r *http.Request, params LoginParams)
}

// ServerInterfaceWrapper converts payloads to parameters.
type ServerInterfaceWrapper struct {
	Handler            ServerInterface
	ErrorHandlerFunc   func(w http.ResponseWriter, r *http.Request, err error)
	HandlerMiddlewares []MiddlewareFunc
}

type MiddlewareFunc func(http.Handler) http.Handler

const MaxPasswordLength = 72

// Register operation middleware.
func (siw *ServerInterfaceWrapper) Register(w http.ResponseWriter, r *http.Request) {
	// ------------- Required JSON content type -----------------------

	contentType := r.Header.Get("Content-Type")
	if strings.ToLower(strings.TrimSpace(contentType)) != "application/json" {
		siw.ErrorHandlerFunc(w, r, fmt.Errorf("%w: %s", errs.ErrContentType, contentType))
		return
	}

	// Decode request body params.
	var params RegisterParams

	defer r.Body.Close()

	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		var e *json.UnmarshalTypeError
		if errors.As(err, &e) {
			siw.ErrorHandlerFunc(w, r, fmt.Errorf(
				"%w: %s must be of type %s, got %s",
				errs.ErrInvalidPayload, e.Field, e.Type, e.Value),
			)
			return
		}
		siw.ErrorHandlerFunc(w, r, err)
		return
	}

	// ------------- Required JSON body parameter "login" -------------

	if params.Login == "" {
		siw.ErrorHandlerFunc(w, r, fmt.Errorf("%w: login", errs.ErrRequiredJSONBodyParam))
		return
	}

	// ------------- Required JSON body parameter "password" ----------

	if params.Password == "" {
		siw.ErrorHandlerFunc(w, r, fmt.Errorf("%w: password", errs.ErrRequiredJSONBodyParam))
		return
	}

	// Password must not exceed 72 characters in length [bcrypt.ErrPasswordTooLong]
	if len(params.Password) > MaxPasswordLength {
		ErrorHandlerFunc(w, r, fmt.Errorf(
			"%w: password must not exceed 72 characters in length",
			errs.ErrInvalidPayload))
		return
	}

	handler := http.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.Register(w, r, params)
	}))

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler.ServeHTTP(w, r)
}

// Login operation middleware.
func (siw *ServerInterfaceWrapper) Login(w http.ResponseWriter, r *http.Request) {
	// ------------- Required JSON content type -----------------------

	contentType := r.Header.Get("Content-Type")
	if strings.ToLower(strings.TrimSpace(contentType)) != "application/json" {
		siw.ErrorHandlerFunc(w, r, fmt.Errorf("%w: %s", errs.ErrContentType, contentType))
		return
	}

	// Decode request body params.
	var params LoginParams

	defer r.Body.Close()

	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		var e *json.UnmarshalTypeError
		if errors.As(err, &e) {
			siw.ErrorHandlerFunc(w, r, fmt.Errorf(
				"%w: %s must be of type %s, got %s",
				errs.ErrInvalidPayload, e.Field, e.Type, e.Value),
			)
			return
		}
		siw.ErrorHandlerFunc(w, r, err)
		return
	}

	// ------------- Required JSON body parameter "login" -------------

	if params.Login == "" {
		siw.ErrorHandlerFunc(w, r, fmt.Errorf("%w: login", errs.ErrRequiredJSONBodyParam))
		return
	}

	// ------------- Required JSON body parameter "password" ----------

	if params.Password == "" {
		siw.ErrorHandlerFunc(w, r, fmt.Errorf("%w: password", errs.ErrRequiredJSONBodyParam))
		return
	}

	handler := http.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.Login(w, r, params)
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
		r.Post(options.BaseURL+"/register", wrapper.Register)
	})
	r.Group(func(r chi.Router) {
		r.Post(options.BaseURL+"/login", wrapper.Login)
	})

	return r
}
