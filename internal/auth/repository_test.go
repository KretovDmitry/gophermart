package auth

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/KretovDmitry/gophermart-loyalty-service/internal/models/errs"
	"github.com/KretovDmitry/gophermart-loyalty-service/internal/models/user"
)

// Lock in case of t.Parallel call.
type mockRepository struct {
	items []user.User
	mu    sync.RWMutex
}

func (m *mockRepository) GetUserByID(_ context.Context, userID int) (*user.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, item := range m.items {
		if item.ID == userID {
			return &item, nil
		}
	}
	return &user.User{}, errs.ErrNotFound
}

func (m *mockRepository) GetUserByLogin(_ context.Context, login string) (*user.User, error) {
	if login == "panic" {
		return &user.User{}, errors.New("don't panic!")
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, item := range m.items {
		if item.Login == login {
			return &item, nil
		}
	}
	return &user.User{}, errs.ErrNotFound
}

func (m *mockRepository) CreateUser(_ context.Context, login, password string) (int, error) {
	if login == "panic" {
		return -1, errors.New("don't panic!")
	}
	m.mu.Lock()
	maxID := -1
	for _, item := range m.items {
		if item.Login == login {
			return -1, errs.ErrDataConflict
		}
		maxID = max(maxID, item.ID)
	}
	m.items = append(m.items, user.User{
		ID:        maxID + 1,
		Login:     login,
		Password:  password,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})
	m.mu.Unlock()
	return maxID + 1, nil
}

func (m *mockRepository) CreateAccount(_ context.Context, userID int) error {
	return nil
}
