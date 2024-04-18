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
		if errors.Is(err, bcrypt.ErrPasswordTooLong) {
			ErrorHandlerFunc(w, r,
				fmt.Errorf(
					"%w: must not exceed 72 characters in length",
					errs.ErrInvalidPassword),
			)
			return
		}
		ErrorHandlerFunc(w, r, err)
		return
	}

	// Create user.
	id, err := s.repo.CreateUser(r.Context(), params.Login, string(hashPassword))
	if err != nil {
		// TODO: SENTINEL errors.Is
		ErrorHandlerFunc(w, r, err)
		return
	}

	// Build authentication token.
	authToken, err := jwt.BuildString(id, s.config.JWT.SigningKey, s.config.JWT.Expiration)
	if err != nil {
		ErrorHandlerFunc(w, r, err)
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
	// Retrieve user from the database by provided login.
	u, err := s.repo.GetUserByLogin(r.Context(), params.Login)
	if err != nil {
		ErrorHandlerFunc(w, r, err)
		return
	}

	// Compare stored and provided passwords.
	err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(params.Password))
	if err != nil {
		return
	}

	// Build authentication token.
	authToken, err := jwt.BuildString(u.ID, s.config.JWT.SigningKey, s.config.JWT.Expiration)
	if err != nil {
		ErrorHandlerFunc(w, r, err)
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
				ErrorHandlerFunc(w, r, fmt.Errorf("%w: Authorization", http.ErrNoCookie))
				return
			}
			ErrorHandlerFunc(w, r, err)
			return
		}

		userID, err := jwt.GetUserID(authCookie.Value, s.config.JWT.SigningKey)
		if err != nil {
			ErrorHandlerFunc(w, r,
				&errs.InvalidAuthorizationError{Message: err.Error()})
			return
		}

		u, err := s.repo.GetUserByID(r.Context(), userID)
		if err != nil {
			if errors.Is(err, errs.ErrNotFound) {
				ErrorHandlerFunc(w, r, fmt.Errorf("%w: user id: %d", errs.ErrNotFound, userID))
				return
			}
			ErrorHandlerFunc(w, r, err)
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
	appError := errs.JSON{Error: err.Error()}
	code := http.StatusInternalServerError

	switch err.(type) {
	case *errs.RequiredJSONBodyParamError:
		code = http.StatusBadRequest
	case *errs.InvalidAuthorizationError:
		code = http.StatusUnauthorized
	case *errs.AlreadyExistsError:
		code = http.StatusConflict
	}

	switch {
	// Status Unauthorized.
	case errors.Is(err, errs.ErrNotFound):
		code = http.StatusUnauthorized
	case errors.Is(err, http.ErrNoCookie):
		code = http.StatusUnauthorized

	// Status Bad Request.
	case errors.Is(err, errs.ErrInvalidPassword):

	}

	// Empty body.
	if errors.Is(err, io.EOF) {
		code = http.StatusBadRequest
	}

	w.WriteHeader(code)

	if err = json.NewEncoder(w).Encode(appError); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
