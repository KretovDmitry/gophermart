package interfaces

import (
	"context"

	"github.com/KretovDmitry/gophermart/internal/application/params"
	"github.com/KretovDmitry/gophermart/internal/domain/entities"
	"github.com/KretovDmitry/gophermart/internal/domain/entities/user"
)

// AccountService represents all service actions.
type AccountService interface {
	GetAccount(context.Context, user.ID) (*entities.Account, error)
	Withdraw(context.Context, *params.Withdraw) error
	GetWithdrawals(context.Context, user.ID) ([]*entities.Withdrawal, error)
}
