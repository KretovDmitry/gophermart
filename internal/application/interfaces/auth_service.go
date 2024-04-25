package interfaces

import (
	"context"

	"github.com/KretovDmitry/gophermart/internal/domain/entities/user"
)

// AuthService represents all service actions.
type AuthService interface {
	Register(ctx context.Context, login, password string) (user.ID, error)
	Login(ctx context.Context, login, password string) (*user.User, error)
	BuildAuthToken(user.ID) (string, error)
	GetUserFromToken(ctx context.Context, token string) (*user.User, error)
}
