package middleware

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/KretovDmitry/gophermart/internal/application/errs"
	"github.com/KretovDmitry/gophermart/internal/application/interfaces"
	"github.com/KretovDmitry/gophermart/internal/domain/entities/user"
)

// Authorization middleware.
func Middleware(service interfaces.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		f := func(w http.ResponseWriter, r *http.Request) {
			authCookie, err := r.Cookie("Authorization")
			if err != nil {
				if errors.Is(err, http.ErrNoCookie) {
					errorHandlerFunc(w, r, fmt.Errorf("authorization token: %w", errs.ErrNotFound))
					return
				}
				errorHandlerFunc(w, r, fmt.Errorf("authorization token: %w", err))
				return
			}

			u, err := service.GetUserFromToken(r.Context(), authCookie.Value)
			if err != nil {
				errorHandlerFunc(w, r, err)
				return
			}

			r = r.WithContext(user.NewContext(r.Context(), u))

			next.ServeHTTP(w, r)
		}

		return http.HandlerFunc(f)
	}
}

// errorHandlerFunc handles sending of an error in the JSON format,
// writing appropriate status code and handling the failure to marshal that.
func errorHandlerFunc(w http.ResponseWriter, _ *http.Request, err error) {
	errJSON := errs.JSON{Error: err.Error()}
	code := http.StatusInternalServerError

	if errors.Is(err, errs.ErrNotFound) || errors.Is(err, errs.ErrInvalidCredentials) {
		code = http.StatusUnauthorized
	}

	w.WriteHeader(code)

	if err = json.NewEncoder(w).Encode(errJSON); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
