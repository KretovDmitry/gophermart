package auth

import (
	"encoding/json"
	"io"
	"net/http"

	appErrors "github.com/KretovDmitry/gophermart-loyalty-service/internal/models/errors"
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

// Register operation middleware.
func (siw *ServerInterfaceWrapper) Register(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var err error

	// Parameter object where we will unmarshal all parameters from the context.
	var params RegisterParams

	data, err := io.ReadAll(r.Body)
	if err != nil {
		siw.ErrorHandlerFunc(w, r, err)
	}
	r.Body.Close()

	if err = json.Unmarshal(data, &params); err != nil {
		siw.ErrorHandlerFunc(w, r, err)
	}

	// ------------- Required JSON body parameter "login" -------------

	if params.Login == "" {
		siw.ErrorHandlerFunc(w, r,
			&appErrors.RequiredJSONBodyParamError{ParamName: "login"})
		return
	}

	// ------------- Required JSON body parameter "password" ----------

	if params.Password == "" {
		siw.ErrorHandlerFunc(w, r,
			&appErrors.RequiredJSONBodyParamError{ParamName: "password"})
		return
	}

	handler := http.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.Register(w, r, params)
	}))

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler.ServeHTTP(w, r.WithContext(ctx))
}

// Login operation middleware.
func (siw *ServerInterfaceWrapper) Login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var err error

	// Parameter object where we will unmarshal all parameters from the context.
	var params LoginParams

	data, err := io.ReadAll(r.Body)
	if err != nil {
		siw.ErrorHandlerFunc(w, r, err)
	}
	r.Body.Close()

	if err = json.Unmarshal(data, &params); err != nil {
		siw.ErrorHandlerFunc(w, r, err)
	}

	// ------------- Required JSON body parameter "login" -------------

	if params.Login == "" {
		siw.ErrorHandlerFunc(w, r,
			&appErrors.RequiredJSONBodyParamError{ParamName: "login"})
		return
	}

	// ------------- Required JSON body parameter "password" ----------

	if params.Password == "" {
		siw.ErrorHandlerFunc(w, r,
			&appErrors.RequiredJSONBodyParamError{ParamName: "password"})
		return
	}

	handler := http.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.Login(w, r, params)
	}))

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler.ServeHTTP(w, r.WithContext(ctx))
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
		options.ErrorHandlerFunc = func(w http.ResponseWriter, r *http.Request, err error) {
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
