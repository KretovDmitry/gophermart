package repositories

import (
	"context"

	"github.com/KretovDmitry/gophermart/internal/domain/entities"
	"github.com/KretovDmitry/gophermart/internal/domain/entities/user"
	"github.com/shopspring/decimal"
)

type AccountRepository interface {
	CreateAccount(context.Context, user.ID) error
	GetAccountByUserID(context.Context, user.ID) (*entities.Account, error)
	Withdraw(context.Context, user.ID, decimal.Decimal) error
	GetWithdrawalsByUserID(context.Context, user.ID) ([]*entities.Withdrawal, error)
	SaveAccountOperation(context.Context, *entities.Operation) error
	AddToAccount(context.Context, user.ID, decimal.Decimal) error
}
