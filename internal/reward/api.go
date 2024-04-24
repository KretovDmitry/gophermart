package reward

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/KretovDmitry/gophermart-loyalty-service/internal/models/errs"
	"github.com/KretovDmitry/gophermart-loyalty-service/pkg/luhn"
	"github.com/go-chi/chi/v5"
)

type PostOrderParams struct {
	Number string
}

type WithdrawParams struct {
	Order string  `json:"order"`
	Sum   float64 `json:"sum"`
}

// ServerInterface represents all server handlers.
type ServerInterface interface {
	// Order upload (POST /api/user/orders HTTP/1.1).
	CreateOrder(w http.ResponseWriter, r *http.Request, params PostOrderParams)
	// Get user orders (DET /api/user/orders HTTP/1.1).
	GetOrders(w http.ResponseWriter, r *http.Request)
	// Get user account data (GET /api/user/balance HTTP/1.1).
	GetAccount(w http.ResponseWriter, r *http.Request)
	// Withdraw (POST /api/user/balance/withdraw HTTP/1.1).
	Withdraw(w http.ResponseWriter, r *http.Request, params WithdrawParams)
	// Get all user withdrawals (GET /api/user/withdrawals HTTP/1.1).
	GetWithdrawals(w http.ResponseWriter, r *http.Request)
}

// ServerInterfaceWrapper converts payloads to parameters.
type ServerInterfaceWrapper struct {
	Handler          ServerInterface
	ErrorHandlerFunc func(w http.ResponseWriter, r *http.Request, err error)
}

type MiddlewareFunc func(http.Handler) http.Handler

// Create order operation middleware.
func (siw *ServerInterfaceWrapper) CreateOrder(w http.ResponseWriter, r *http.Request) {
	// ------------- Required text/plain content type -----------------

	contentType := r.Header.Get("Content-Type")
	if !isTextPlainContentType(contentType) {
		siw.ErrorHandlerFunc(w, r, fmt.Errorf("%w: invalid content type", errs.ErrInvalidRequest))
		return
	}

	// ------------- Parse and validate request body params ----------

	var params PostOrderParams

	defer r.Body.Close()
	bytes, err := io.ReadAll(r.Body)
	if err != nil {
		if errors.Is(err, io.EOF) {
			siw.ErrorHandlerFunc(w, r, fmt.Errorf("%w: empty body", errs.ErrInvalidRequest))
			return
		}
		siw.ErrorHandlerFunc(w, r, err)
		return
	}

	number := string(bytes)

	if err = luhn.Validate(number); err != nil {
		siw.ErrorHandlerFunc(w, r, errs.ErrInvalidOrderNumber)
		return
	}

	params.Number = number

	siw.Handler.CreateOrder(w, r, params)
}

// Withdraw operation middleware.
func (siw *ServerInterfaceWrapper) Withdraw(w http.ResponseWriter, r *http.Request) {
	// ------------- Required JSON content type -----------------------

	contentType := r.Header.Get("Content-Type")
	if strings.ToLower(strings.TrimSpace(contentType)) != "application/json" {
		siw.ErrorHandlerFunc(w, r, fmt.Errorf("%w: invalid content type", errs.ErrInvalidRequest))
		return
	}

	// ------------- Dacode and validate request body params ---------
	var params WithdrawParams

	defer r.Body.Close()

	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		var e *json.UnmarshalTypeError
		if errors.As(err, &e) {
			siw.ErrorHandlerFunc(w, r, fmt.Errorf(
				"%w: %s must be of type %s, got %s",
				errs.ErrInvalidRequest, e.Field, e.Type, e.Value),
			)
			return
		}
		if errors.Is(err, io.EOF) {
			siw.ErrorHandlerFunc(w, r, fmt.Errorf("%w: empty body", errs.ErrInvalidRequest))
			return
		}
		siw.ErrorHandlerFunc(w, r, err)
		return
	}

	if err := luhn.Validate(params.Order); err != nil {
		siw.ErrorHandlerFunc(w, r, errs.ErrInvalidOrderNumber)
		return
	}

	if params.Sum <= 0 {
		siw.ErrorHandlerFunc(w, r, fmt.Errorf("%w: invalid sum", errs.ErrInvalidRequest))
		return
	}

	siw.Handler.Withdraw(w, r, params)
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
		Handler:          si,
		ErrorHandlerFunc: options.ErrorHandlerFunc,
	}

	r.Group(func(r chi.Router) {
		for _, middleware := range options.Middlewares {
			r.Use(middleware)
		}
		r.Post(options.BaseURL+"/orders", wrapper.CreateOrder)
		r.Get(options.BaseURL+"/orders", si.GetOrders)
		r.Get(options.BaseURL+"/balance", si.GetAccount)
		r.Post(options.BaseURL+"/balance/withdraw", wrapper.Withdraw)
		r.Get(options.BaseURL+"/withdrawals", si.GetWithdrawals)
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
