package auth

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/KretovDmitry/gophermart-loyalty-service/internal/models/errs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
				response:   fmt.Sprintf("%s: invalid content type", errs.ErrInvalidRequest),
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
				response:   fmt.Sprintf("%v: empty body", errs.ErrInvalidRequest),
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
				response:   fmt.Sprintf("%s: login required", errs.ErrInvalidRequest),
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
				response:   fmt.Sprintf("%s: password required", errs.ErrInvalidRequest),
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
					errs.ErrInvalidRequest),
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
					errs.ErrInvalidRequest),
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
					errs.ErrInvalidRequest),
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
					errs.ErrInvalidRequest),
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
					errs.ErrInvalidRequest),
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
					errs.ErrInvalidRequest),
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
					errs.ErrInvalidRequest),
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
					errs.ErrInvalidRequest),
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
					errs.ErrInvalidRequest),
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
				response:   fmt.Sprintf("%s: invalid content type", errs.ErrInvalidRequest),
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
				response:   fmt.Sprintf("%v: empty body", errs.ErrInvalidRequest),
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
				response:   fmt.Sprintf("%s: login required", errs.ErrInvalidRequest),
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
				response:   fmt.Sprintf("%s: password required", errs.ErrInvalidRequest),
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
					errs.ErrInvalidRequest),
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
					errs.ErrInvalidRequest),
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
					errs.ErrInvalidRequest),
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
					errs.ErrInvalidRequest),
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
					errs.ErrInvalidRequest),
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
					errs.ErrInvalidRequest),
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
					errs.ErrInvalidRequest),
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
					errs.ErrInvalidRequest),
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
