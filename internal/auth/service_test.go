package auth

import "net/http"

type mockAuthService struct{}

func (m *mockAuthService) Register(w http.ResponseWriter, r *http.Request, params RegisterParams) {}

func (m *mockAuthService) Login(w http.ResponseWriter, r *http.Request, params LoginParams) {}
