package repositories

import (
	"context"

	"github.com/KretovDmitry/gophermart/internal/domain/entities/user"
)

type UserRepository interface {
	GetUserByID(context.Context, user.ID) (*user.User, error)
	GetUserByLogin(ctx context.Context, login string) (*user.User, error)
	CreateUser(ctx context.Context, login, password string) (user.ID, error)
}
