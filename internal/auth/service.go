package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/KretovDmitry/gophermart-loyalty-service/internal/config"
	"github.com/KretovDmitry/gophermart-loyalty-service/internal/jwt"
	"github.com/KretovDmitry/gophermart-loyalty-service/internal/models/errs"
	"github.com/KretovDmitry/gophermart-loyalty-service/internal/models/user"
	"github.com/KretovDmitry/gophermart-loyalty-service/pkg/logger"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	repo   Repository
	logger logger.Logger
	config *config.Config
}

func NewService(repo Repository, logger logger.Logger, config *config.Config) (*Service, error) {
	if config == nil {
		return nil, errors.New("nil dependency: config")
	}
	return &Service{repo: repo, logger: logger, config: config}, nil
}

var _ ServerInterface = (*Service)(nil)

// Registration (POST /api/user/register).
func (s *Service) Register(w http.ResponseWriter, r *http.Request, params RegisterParams) {
	// Careate password hash.
	hashPassword, err := bcrypt.GenerateFromPassword([]byte(params.Password), s.config.PasswordHashCost)
	if err != nil {
		ErrorHandlerFunc(w, r, fmt.Errorf("hash password: %w", err))
		return
	}

	// Create user.
	id, err := s.repo.CreateUser(r.Context(), params.Login, string(hashPassword))
	if err != nil {
		if errors.Is(err, errs.ErrConflict) {
			ErrorHandlerFunc(w, r, fmt.Errorf("%w: login %q already exists", err, params.Login))
			return
		}
		ErrorHandlerFunc(w, r, fmt.Errorf("create user: %w", err))
		return
	}

	// Build authentication token.
	authToken, err := jwt.BuildString(id, s.config.JWT.SigningKey, s.config.JWT.Expiration)
	if err != nil {
		ErrorHandlerFunc(w, r, fmt.Errorf("build token: %w", err))
		return
	}

	// Set the "Authorization" cookie with the JWT authentication token.
	http.SetCookie(w, &http.Cookie{
		Name:     "Authorization",
		Value:    authToken,
		Expires:  time.Now().Add(s.config.JWT.Expiration),
		HttpOnly: true,
	})

	w.WriteHeader(http.StatusOK)
}

// Authentication (POST /api/user/login).
func (s *Service) Login(w http.ResponseWriter, r *http.Request, params LoginParams) {
	// Retrieve user from the database with provided login.
	u, err := s.repo.GetUserByLogin(r.Context(), params.Login)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			ErrorHandlerFunc(w, r, fmt.Errorf("%w: user with login %q not found",
				errs.ErrInvalidCredentials, params.Login))
			return
		}
		ErrorHandlerFunc(w, r, fmt.Errorf("get user %q: %w", params.Login, err))
		return
	}

	// Compare stored and provided passwords.
	err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(params.Password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			ErrorHandlerFunc(w, r, fmt.Errorf("%w: password", errs.ErrInvalidCredentials))
			return
		}
		ErrorHandlerFunc(w, r, fmt.Errorf("compare passwords: %w", err))
		return
	}

	// Build authentication token.
	authToken, err := jwt.BuildString(u.ID, s.config.JWT.SigningKey, s.config.JWT.Expiration)
	if err != nil {
		ErrorHandlerFunc(w, r, fmt.Errorf("build token: %w", err))
		return
	}

	// Set the "Authorization" cookie with the JWT authentication token.
	http.SetCookie(w, &http.Cookie{
		Name:     "Authorization",
		Value:    authToken,
		Expires:  time.Now().Add(s.config.JWT.Expiration),
		HttpOnly: true,
	})

	w.WriteHeader(http.StatusOK)
}

// Authorization middleware.
func (s *Service) Middleware(next http.Handler) http.Handler {
	f := func(w http.ResponseWriter, r *http.Request) {
		authCookie, err := r.Cookie("Authorization")
		if err != nil {
			if errors.Is(err, http.ErrNoCookie) {
				ErrorHandlerFunc(w, r, fmt.Errorf("authorization token: %w", errs.ErrNotFound))
				return
			}
			ErrorHandlerFunc(w, r, fmt.Errorf("authorization token: %w", err))
			return
		}

		userID, err := jwt.GetUserID(authCookie.Value, s.config.JWT.SigningKey)
		if err != nil {
			ErrorHandlerFunc(w, r, fmt.Errorf("parse token: %w", err))
			return
		}

		u, err := s.repo.GetUserByID(r.Context(), userID)
		if err != nil {
			ErrorHandlerFunc(w, r, fmt.Errorf("get user %q: %w", userID, err))
			return
		}

		r = r.WithContext(user.NewContext(r.Context(), u))

		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(f)
}

// ErrorHandlerFunc handles sending of an error in the JSON format,
// writing appropriate status code and handling the failure to marshal that.
func ErrorHandlerFunc(w http.ResponseWriter, _ *http.Request, err error) {
	errJSON := errs.JSON{Error: err.Error()}
	code := http.StatusInternalServerError

	switch {
	// Status Bad Request.
	case errors.Is(err, errs.ErrRequiredJSONBodyParam) ||
		errors.Is(err, errs.ErrInvalidPayload) ||
		errors.Is(err, errs.ErrContentType) ||
		errors.Is(err, io.EOF):
		code = http.StatusBadRequest

	// Status Unauthorized.
	case errors.Is(err, errs.ErrNotFound) ||
		errors.Is(err, errs.ErrInvalidCredentials):
		code = http.StatusUnauthorized

	// Status Conflict.
	case errors.Is(err, errs.ErrConflict):
		code = http.StatusConflict
	}

	w.WriteHeader(code)

	if err = json.NewEncoder(w).Encode(errJSON); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
