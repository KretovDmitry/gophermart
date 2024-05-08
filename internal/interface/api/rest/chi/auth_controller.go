package rest

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/KretovDmitry/gophermart/internal/application/errs"
	"github.com/KretovDmitry/gophermart/internal/application/interfaces"
	"github.com/KretovDmitry/gophermart/internal/interface/api/rest/header"
	"github.com/KretovDmitry/gophermart/internal/interface/api/rest/request"
	"github.com/KretovDmitry/gophermart/pkg/logger"
	"github.com/go-chi/chi/v5"
)

type AuthController struct {
	service         interfaces.AuthService
	logger          logger.Logger
	tokenExpiration time.Duration
}

// NewAuthController registers http.Handlers with additional options.
func NewAuthController(
	service interfaces.AuthService,
	tokenExpiration time.Duration,
	logger logger.Logger,
	options ChiServerOptions,
) {
	r := options.BaseRouter

	if r == nil {
		r = chi.NewRouter()
	}

	c := AuthController{
		service:         service,
		tokenExpiration: tokenExpiration,
		logger:          logger,
	}

	r.Group(func(r chi.Router) {
		for _, middleware := range options.Middlewares {
			r.Use(middleware)
		}
		r.Post(options.BaseURL+"/register", c.Register)
		r.Post(options.BaseURL+"/login", c.Login)
	})
}

const MaxPasswordLength = 72

// Register user.
func (c *AuthController) Register(w http.ResponseWriter, r *http.Request) {
	// Check content type.
	if !header.IsApplicationJSONContentType(r) {
		c.ErrorHandlerFunc(w, r, fmt.Errorf("%w: invalid content type", errs.ErrInvalidRequest))
		return
	}

	// Read, decode payload and close request body.
	var p request.Register

	defer r.Body.Close()

	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		c.ErrorHandlerFunc(w, r, checkJSONDecodeError(err))
		return
	}

	// Check payload.
	if p.Login == "" {
		c.ErrorHandlerFunc(w, r, fmt.Errorf("%w: login required", errs.ErrInvalidRequest))
		return
	}
	if p.Password == "" {
		c.ErrorHandlerFunc(w, r, fmt.Errorf("%w: password required", errs.ErrInvalidRequest))
		return
	}

	// Password must not exceed 72 characters in length [bcrypt.ErrPasswordTooLong]
	if len(p.Password) > MaxPasswordLength {
		c.ErrorHandlerFunc(w, r, fmt.Errorf(
			"%w: password must not exceed 72 characters in length",
			errs.ErrInvalidRequest))
		return
	}

	// Register user.
	userID, err := c.service.Register(r.Context(), p.Login, p.Password)
	if err != nil {
		c.ErrorHandlerFunc(w, r, fmt.Errorf("register user: %w", err))
		return
	}

	// Build authentication token.
	authToken, err := c.service.BuildAuthToken(userID)
	if err != nil {
		c.ErrorHandlerFunc(w, r, fmt.Errorf("build token: %w", err))
		return
	}

	// Set the "Authorization" cookie with the JWT authentication token.
	http.SetCookie(w, &http.Cookie{
		Name:     "Authorization",
		Value:    authToken,
		Expires:  time.Now().Add(c.tokenExpiration),
		HttpOnly: true,
	})

	w.WriteHeader(http.StatusOK)
}

// Login user.
func (c *AuthController) Login(w http.ResponseWriter, r *http.Request) {
	// Check content type.
	if !header.IsApplicationJSONContentType(r) {
		c.ErrorHandlerFunc(w, r, fmt.Errorf("%w: invalid content type", errs.ErrInvalidRequest))
		return
	}

	// Read, decode payload and close request body.
	var p request.Login

	defer r.Body.Close()

	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		c.ErrorHandlerFunc(w, r, checkJSONDecodeError(err))
		return
	}

	// Check payload.
	if p.Login == "" {
		c.ErrorHandlerFunc(w, r, fmt.Errorf("%w: login required", errs.ErrInvalidRequest))
		return
	}
	if p.Password == "" {
		c.ErrorHandlerFunc(w, r, fmt.Errorf("%w: password required", errs.ErrInvalidRequest))
		return
	}

	// Login user.
	user, err := c.service.Login(r.Context(), p.Login, p.Password)
	if err != nil {
		c.ErrorHandlerFunc(w, r, fmt.Errorf("login: %w", err))
		return
	}

	// Build authentication token.
	authToken, err := c.service.BuildAuthToken(user.ID)
	if err != nil {
		c.ErrorHandlerFunc(w, r, fmt.Errorf("build token: %w", err))
		return
	}

	// Set the "Authorization" cookie with the JWT authentication token.
	http.SetCookie(w, &http.Cookie{
		Name:     "Authorization",
		Value:    authToken,
		Expires:  time.Now().Add(c.tokenExpiration),
		HttpOnly: true,
	})

	w.WriteHeader(http.StatusOK)
}

// ErrorHandlerFunc handles sending of an error in the JSON format,
// writing appropriate status code and handling the failure to marshal that.
func (c *AuthController) ErrorHandlerFunc(w http.ResponseWriter, _ *http.Request, err error) {
	errJSON := errs.JSON{Error: err.Error()}
	code := http.StatusInternalServerError

	switch {
	// Status Bad Request (400).
	case errors.Is(err, errs.ErrInvalidRequest):
		code = http.StatusBadRequest

	// Status Unauthorized (401).
	case errors.Is(err, errs.ErrNotFound) ||
		errors.Is(err, errs.ErrInvalidCredentials):
		code = http.StatusUnauthorized

	// Status Conflict (409).
	case errors.Is(err, errs.ErrDataConflict):
		code = http.StatusConflict
	}

	w.WriteHeader(code)

	c.logger.Errorf("auth controller [%d]: %s", code, err)

	if err = json.NewEncoder(w).Encode(errJSON); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
