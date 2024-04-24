package repositories

import (
	"context"

	"github.com/KretovDmitry/gophermart-loyalty-service/internal/domain/entities"
	"github.com/KretovDmitry/gophermart-loyalty-service/internal/domain/entities/user"
	"github.com/shopspring/decimal"
)

type AccountRepository interface {
	GetAccountByUserID(context.Context, user.ID) (*entities.Account, error)
	Withdraw(ctx context.Context, sum decimal.Decimal, userID user.ID) error
	GetWithdrawals(context.Context, user.ID) ([]*entities.Withdrawal, error)
	SaveAccountOperation(context.Context, *entities.Operation) error
}
