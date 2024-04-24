package interfaces

import (
	"context"

	"github.com/KretovDmitry/gophermart-loyalty-service/internal/application/params"
	"github.com/KretovDmitry/gophermart-loyalty-service/internal/domain/entities"
	"github.com/KretovDmitry/gophermart-loyalty-service/internal/domain/entities/user"
)

type AccountService interface {
	GetAccount(context.Context, user.ID) (*entities.Account, error)
	Withdraw(context.Context, *params.Withdraw) error
	GetWithdrawals(context.Context, user.ID) ([]*entities.Withdrawal, error)
}
