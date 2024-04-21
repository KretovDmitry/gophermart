package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/KretovDmitry/gophermart-loyalty-service/internal/config"
	"github.com/KretovDmitry/gophermart-loyalty-service/internal/jwt"
	"github.com/KretovDmitry/gophermart-loyalty-service/internal/models/errs"
	"github.com/KretovDmitry/gophermart-loyalty-service/internal/models/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestRegisterOperationMiddleware(t *testing.T) {
	path := "/api/user/register"

	type want struct {
		response   string
		statusCode int
	}

	tests := []struct {
		name        string
		contentType string
		payload     io.Reader
		repo        Repository
		want        want
		wantErr     bool
	}{
		{
			name:        "OK",
			contentType: "application/json",
			payload:     strings.NewReader(`{"login":"login","password":"password"}`),
			repo:        &mockRepository{},
			want: want{
				statusCode: http.StatusOK,
				response:   "",
			},
			wantErr: false,
		},
		{
			name:        "invalid content type",
			contentType: "text/plain; charset=utf-8",
			payload:     strings.NewReader(""),
			repo:        &mockRepository{},
			want: want{
				statusCode: http.StatusBadRequest,
				response:   fmt.Sprintf("%s: text/plain; charset=utf-8", errs.ErrInvalidContentType),
			},
			wantErr: true,
		},
		{
			name:        "empty body",
			contentType: "application/json",
			payload:     strings.NewReader(""),
			repo:        &mockRepository{},
			want: want{
				statusCode: http.StatusBadRequest,
				response:   fmt.Sprintf("%v: empty body", errs.ErrInvalidPayload),
			},
			wantErr: true,
		},
		{
			name:        "empty login",
			contentType: "application/json",
			payload:     strings.NewReader(`{"login":"","password":"password"}`),
			repo:        &mockRepository{},
			want: want{
				statusCode: http.StatusBadRequest,
				response:   fmt.Sprintf("%s: login", errs.ErrRequiredBodyParam),
			},
			wantErr: true,
		},
		{
			name:        "empty password",
			contentType: "application/json",
			payload:     strings.NewReader(`{"login":"login","password":""}`),
			repo:        &mockRepository{},
			want: want{
				statusCode: http.StatusBadRequest,
				response:   fmt.Sprintf("%s: password", errs.ErrRequiredBodyParam),
			},
			wantErr: true,
		},
		{
			name:        "invalid data type: login is number",
			contentType: "application/json",
			payload:     strings.NewReader(`{"login":123,"password":"password"}`),
			repo:        &mockRepository{},
			want: want{
				statusCode: http.StatusBadRequest,
				response: fmt.Sprintf("%v: login must be of type string, got number",
					errs.ErrInvalidPayload),
			},
			wantErr: true,
		},
		{
			name:        "invalid data type: login is bool",
			contentType: "application/json",
			payload:     strings.NewReader(`{"login":true,"password":"password"}`),
			repo:        &mockRepository{},
			want: want{
				statusCode: http.StatusBadRequest,
				response: fmt.Sprintf("%v: login must be of type string, got bool",
					errs.ErrInvalidPayload),
			},
			wantErr: true,
		},
		{
			name:        "invalid data type: login is object",
			contentType: "application/json",
			payload:     strings.NewReader(`{"login":{},"password":"password"}`),
			repo:        &mockRepository{},
			want: want{
				statusCode: http.StatusBadRequest,
				response: fmt.Sprintf("%v: login must be of type string, got object",
					errs.ErrInvalidPayload),
			},
			wantErr: true,
		},
		{
			name:        "invalid data type: login is array",
			contentType: "application/json",
			payload:     strings.NewReader(`{"login":[],"password":"password"}`),
			repo:        &mockRepository{},
			want: want{
				statusCode: http.StatusBadRequest,
				response: fmt.Sprintf("%v: login must be of type string, got array",
					errs.ErrInvalidPayload),
			},
			wantErr: true,
		},
		{
			name:        "invalid data type: password is number",
			contentType: "application/json",
			payload:     strings.NewReader(`{"login":"login","password":123}`),
			repo:        &mockRepository{},
			want: want{
				statusCode: http.StatusBadRequest,
				response: fmt.Sprintf("%v: password must be of type string, got number",
					errs.ErrInvalidPayload),
			},
			wantErr: true,
		},
		{
			name:        "invalid data type: password is bool",
			contentType: "application/json",
			payload:     strings.NewReader(`{"login":"login","password":true}`),
			repo:        &mockRepository{},
			want: want{
				statusCode: http.StatusBadRequest,
				response: fmt.Sprintf("%v: password must be of type string, got bool",
					errs.ErrInvalidPayload),
			},
			wantErr: true,
		},
		{
			name:        "invalid data type: password is object",
			contentType: "application/json",
			payload:     strings.NewReader(`{"login":"login","password":{}}`),
			repo:        &mockRepository{},
			want: want{
				statusCode: http.StatusBadRequest,
				response: fmt.Sprintf("%v: password must be of type string, got object",
					errs.ErrInvalidPayload),
			},
			wantErr: true,
		},
		{
			name:        "invalid data type: password is array",
			contentType: "application/json",
			payload:     strings.NewReader(`{"login":"login","password":[]}`),
			repo:        &mockRepository{},
			want: want{
				statusCode: http.StatusBadRequest,
				response: fmt.Sprintf("%v: password must be of type string, got array",
					errs.ErrInvalidPayload),
			},
			wantErr: true,
		},
		{
			name:        "password too long: gt 72 bytes",
			contentType: "application/json",
			payload:     strings.NewReader(`{"login":"login","password":"123456789_123456789_123456789_123456789_123456789_123456789_123456789_123"}`),
			repo:        &mockRepository{},
			want: want{
				statusCode: http.StatusBadRequest,
				response: fmt.Sprintf(
					"%v: password must not exceed 72 characters in length",
					errs.ErrInvalidPayload),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r := httptest.NewRequest(http.MethodPost, path, tt.payload)
			r.Header.Set("Content-Type", tt.contentType)

			w := httptest.NewRecorder()

			siw := ServerInterfaceWrapper{
				Handler:          &mockAuthService{},
				ErrorHandlerFunc: ErrorHandlerFunc,
			}

			siw.Register(w, r)

			res := w.Result()

			errorResponse := new(errs.JSON)

			if tt.wantErr {
				err := json.NewDecoder(res.Body).Decode(&errorResponse)
				require.NoError(t, err, "failed to decode JSON response")
			}
			r.Body.Close()
			res.Body.Close()

			assert.Equal(t, tt.want.statusCode, res.StatusCode, "status mismatch")
			if tt.wantErr {
				assert.Equal(t, errorResponse.Error, tt.want.response, "error message mismatch")
			}
		})
	}
}

func TestLoginOperationMiddleware(t *testing.T) {
	path := "/api/user/login"

	type want struct {
		response   string
		statusCode int
	}

	tests := []struct {
		name        string
		contentType string
		payload     io.Reader
		repo        Repository
		want        want
		wantErr     bool
	}{
		{
			name:        "OK",
			contentType: "application/json",
			payload:     strings.NewReader(`{"login":"login","password":"password"}`),
			repo:        &mockRepository{},
			want: want{
				statusCode: http.StatusOK,
				response:   "",
			},
			wantErr: false,
		},
		{
			name:        "invalid content type",
			contentType: "text/plain; charset=utf-8",
			payload:     strings.NewReader(""),
			repo:        &mockRepository{},
			want: want{
				statusCode: http.StatusBadRequest,
				response:   fmt.Sprintf("%s: text/plain; charset=utf-8", errs.ErrInvalidContentType),
			},
			wantErr: true,
		},
		{
			name:        "empty body",
			contentType: "application/json",
			payload:     strings.NewReader(""),
			repo:        &mockRepository{},
			want: want{
				statusCode: http.StatusBadRequest,
				response:   fmt.Sprintf("%v: empty body", errs.ErrInvalidPayload),
			},
			wantErr: true,
		},
		{
			name:        "empty login",
			contentType: "application/json",
			payload:     strings.NewReader(`{"login":"","password":"password"}`),
			repo:        &mockRepository{},
			want: want{
				statusCode: http.StatusBadRequest,
				response:   fmt.Sprintf("%s: login", errs.ErrRequiredBodyParam),
			},
			wantErr: true,
		},
		{
			name:        "empty password",
			contentType: "application/json",
			payload:     strings.NewReader(`{"login":"login","password":""}`),
			repo:        &mockRepository{},
			want: want{
				statusCode: http.StatusBadRequest,
				response:   fmt.Sprintf("%s: password", errs.ErrRequiredBodyParam),
			},
			wantErr: true,
		},
		{
			name:        "invalid data type: login is number",
			contentType: "application/json",
			payload:     strings.NewReader(`{"login":123,"password":"password"}`),
			repo:        &mockRepository{},
			want: want{
				statusCode: http.StatusBadRequest,
				response: fmt.Sprintf("%v: login must be of type string, got number",
					errs.ErrInvalidPayload),
			},
			wantErr: true,
		},
		{
			name:        "invalid data type: login is bool",
			contentType: "application/json",
			payload:     strings.NewReader(`{"login":true,"password":"password"}`),
			repo:        &mockRepository{},
			want: want{
				statusCode: http.StatusBadRequest,
				response: fmt.Sprintf("%v: login must be of type string, got bool",
					errs.ErrInvalidPayload),
			},
			wantErr: true,
		},
		{
			name:        "invalid data type: login is object",
			contentType: "application/json",
			payload:     strings.NewReader(`{"login":{},"password":"password"}`),
			repo:        &mockRepository{},
			want: want{
				statusCode: http.StatusBadRequest,
				response: fmt.Sprintf("%v: login must be of type string, got object",
					errs.ErrInvalidPayload),
			},
			wantErr: true,
		},
		{
			name:        "invalid data type: login is array",
			contentType: "application/json",
			payload:     strings.NewReader(`{"login":[],"password":"password"}`),
			repo:        &mockRepository{},
			want: want{
				statusCode: http.StatusBadRequest,
				response: fmt.Sprintf("%v: login must be of type string, got array",
					errs.ErrInvalidPayload),
			},
			wantErr: true,
		},
		{
			name:        "invalid data type: password is number",
			contentType: "application/json",
			payload:     strings.NewReader(`{"login":"login","password":123}`),
			repo:        &mockRepository{},
			want: want{
				statusCode: http.StatusBadRequest,
				response: fmt.Sprintf("%v: password must be of type string, got number",
					errs.ErrInvalidPayload),
			},
			wantErr: true,
		},
		{
			name:        "invalid data type: password is bool",
			contentType: "application/json",
			payload:     strings.NewReader(`{"login":"login","password":true}`),
			repo:        &mockRepository{},
			want: want{
				statusCode: http.StatusBadRequest,
				response: fmt.Sprintf("%v: password must be of type string, got bool",
					errs.ErrInvalidPayload),
			},
			wantErr: true,
		},
		{
			name:        "invalid data type: password is object",
			contentType: "application/json",
			payload:     strings.NewReader(`{"login":"login","password":{}}`),
			repo:        &mockRepository{},
			want: want{
				statusCode: http.StatusBadRequest,
				response: fmt.Sprintf("%v: password must be of type string, got object",
					errs.ErrInvalidPayload),
			},
			wantErr: true,
		},
		{
			name:        "invalid data type: password is array",
			contentType: "application/json",
			payload:     strings.NewReader(`{"login":"login","password":[]}`),
			repo:        &mockRepository{},
			want: want{
				statusCode: http.StatusBadRequest,
				response: fmt.Sprintf("%v: password must be of type string, got array",
					errs.ErrInvalidPayload),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r := httptest.NewRequest(http.MethodPost, path, tt.payload)
			r.Header.Set("Content-Type", tt.contentType)

			w := httptest.NewRecorder()

			siw := ServerInterfaceWrapper{
				Handler:          &mockAuthService{},
				ErrorHandlerFunc: ErrorHandlerFunc,
			}

			siw.Login(w, r)

			res := w.Result()

			errorResponse := new(errs.JSON)

			if tt.wantErr {
				err := json.NewDecoder(res.Body).Decode(&errorResponse)
				require.NoError(t, err, "failed to decode JSON response")
			}
			r.Body.Close()
			res.Body.Close()

			assert.Equal(t, tt.want.statusCode, res.StatusCode, "status mismatch")
			if tt.wantErr {
				assert.Equal(t, errorResponse.Error, tt.want.response, "error message mismatch")
			}
		})
	}
}

func TestRegisterHandler(t *testing.T) {
	path := "/api/user/register"

	config := &config.Config{
		PasswordHashCost: 14,
		JWT: config.JWT{
			Expiration: 3 * time.Hour,
			SigningKey: "Kyoto",
		},
	}

	type want struct {
		response   string
		statusCode int
	}

	tests := []struct {
		name    string
		params  RegisterParams
		repo    Repository
		want    want
		wantErr bool
	}{
		{
			name: "OK",
			params: RegisterParams{
				Login:    "gopher",
				Password: "gopher",
			},
			repo: &mockRepository{},
			want: want{
				statusCode: http.StatusOK,
				response:   "",
			},
			wantErr: false,
		},
		{
			name: "login already exists",
			params: RegisterParams{
				Login:    "gopher",
				Password: "gopher",
			},
			repo: &mockRepository{
				items: []user.User{
					{
						ID:       0,
						Login:    "gopher",
						Password: "$2a$14$exSjgqssYnKcJdJY0wJcTeqdpdrH7e4tz8wM3brPZaVtoDV/75UW6",
					},
				},
			},
			want: want{
				statusCode: http.StatusConflict,
				response:   fmt.Sprintf(`%v: login "gopher" already exists`, errs.ErrDataConflict),
			},
			wantErr: true,
		},
		{
			name: "failed to create user",
			params: RegisterParams{
				Login:    "panic",
				Password: "oh-my-zsh",
			},
			repo: &mockRepository{},
			want: want{
				statusCode: http.StatusInternalServerError,
				response:   "create user: don't panic!",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r := httptest.NewRequest(http.MethodPost, path, http.NoBody)

			w := httptest.NewRecorder()

			authHandler, err := NewService(tt.repo, nil, config)
			require.NoError(t, err, "failed to init service")

			authHandler.Register(w, r, tt.params)

			res := w.Result()

			errorResponse := new(errs.JSON)

			if tt.wantErr {
				err = json.NewDecoder(res.Body).Decode(&errorResponse)
				require.NoError(t, err, "failed to decode JSON response")
			}
			r.Body.Close()
			res.Body.Close()

			assert.Equal(t, tt.want.statusCode, res.StatusCode, "status mismatch")
			switch {
			case tt.wantErr:
				assert.Equal(t, errorResponse.Error, tt.want.response, "error message mismatch")

			case !tt.wantErr:
				var token string

				for _, c := range res.Cookies() {
					if c.Name == "Authorization" {
						token = c.Value
						break
					}
				}

				require.NotEmpty(t, token, "the call was successful, but the authorization cookie was not set")

				id, err := jwt.GetUserID(token, config.JWT.SigningKey)
				require.NoError(t, err, "jwt: get user id")
				user, err := tt.repo.GetUserByID(context.TODO(), id)
				require.NoError(t, err, "never errors, but just in case")
				assert.Equal(t, user.ID, id, "token user id mismatch")
			}
		})
	}
}

func TestLoginHandler(t *testing.T) {
	path := "/api/user/login"

	config := &config.Config{
		PasswordHashCost: 14,
		JWT: config.JWT{
			Expiration: 3 * time.Hour,
			SigningKey: "Kyoto",
		},
	}

	type want struct {
		response   string
		statusCode int
	}

	tests := []struct {
		name    string
		params  LoginParams
		repo    Repository
		want    want
		wantErr bool
	}{
		{
			name: "OK",
			params: LoginParams{
				Login:    "gopher",
				Password: "gopher",
			},
			repo: &mockRepository{
				items: []user.User{
					{
						ID:       0,
						Login:    "gopher",
						Password: "$2a$14$exSjgqssYnKcJdJY0wJcTeqdpdrH7e4tz8wM3brPZaVtoDV/75UW6",
					},
				},
			},
			want: want{
				statusCode: http.StatusOK,
				response:   "",
			},
			wantErr: false,
		},
		{
			name: "no such user",
			params: LoginParams{
				Login:    "gopher",
				Password: "gopher",
			},
			repo: &mockRepository{},
			want: want{
				statusCode: http.StatusUnauthorized,
				response: fmt.Sprintf(`%v: user with login "gopher" not found`,
					errs.ErrInvalidCredentials),
			},
			wantErr: true,
		},
		{
			name: "failed to get user by login from database",
			params: LoginParams{
				Login:    "panic",
				Password: "oh-my-zsh",
			},
			repo: &mockRepository{},
			want: want{
				statusCode: http.StatusInternalServerError,
				response:   `get user "panic": don't panic!`,
			},
			wantErr: true,
		},
		{
			name: "wrong password",
			params: LoginParams{
				Login:    "gopher",
				Password: "no_gopher",
			},
			repo: &mockRepository{
				items: []user.User{
					{
						ID:       0,
						Login:    "gopher",
						Password: "$2a$14$exSjgqssYnKcJdJY0wJcTeqdpdrH7e4tz8wM3brPZaVtoDV/75UW6",
					},
				},
			},
			want: want{
				statusCode: http.StatusUnauthorized,
				response:   fmt.Sprintf("%v: password", errs.ErrInvalidCredentials),
			},
			wantErr: true,
		},
		{
			name: "internal error: wrong hash saved to db",
			params: LoginParams{
				Login:    "gopher",
				Password: "gopher",
			},
			repo: &mockRepository{
				items: []user.User{
					{
						ID:       0,
						Login:    "gopher",
						Password: "too_short_hash_LT_59_bytes",
					},
				},
			},
			want: want{
				statusCode: http.StatusInternalServerError,
				response:   fmt.Sprintf("compare passwords: %v", bcrypt.ErrHashTooShort),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r := httptest.NewRequest(http.MethodPost, path, http.NoBody)

			w := httptest.NewRecorder()

			authHandler, err := NewService(tt.repo, nil, config)
			require.NoError(t, err, "failed to init service")

			authHandler.Login(w, r, tt.params)

			res := w.Result()

			errorResponse := new(errs.JSON)

			if tt.wantErr {
				err = json.NewDecoder(res.Body).Decode(&errorResponse)
				require.NoError(t, err, "failed to decode JSON response")
			}
			r.Body.Close()
			res.Body.Close()

			assert.Equal(t, tt.want.statusCode, res.StatusCode, "status mismatch")
			switch {
			case tt.wantErr:
				assert.Equal(t, errorResponse.Error, tt.want.response, "error message mismatch")

			case !tt.wantErr:
				var token string

				for _, c := range res.Cookies() {
					if c.Name == "Authorization" {
						token = c.Value
						break
					}
				}

				require.NotEmpty(t, token, "the call was successful, but the authorization cookie was not set")

				id, err := jwt.GetUserID(token, config.JWT.SigningKey)
				require.NoError(t, err, "jwt: get user id")
				user, err := tt.repo.GetUserByID(context.TODO(), id)
				require.NoError(t, err, "never errors, but just in case")
				assert.Equal(t, user.ID, id, "token user id mismatch")
			}
		})
	}
}
