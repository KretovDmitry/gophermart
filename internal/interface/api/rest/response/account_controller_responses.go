package response

import (
	"time"

	"github.com/KretovDmitry/gophermart/internal/domain/entities"
	"github.com/shopspring/decimal"
)

type GetBalance struct {
	Balance   decimal.Decimal `json:"current"`
	Withdrawn decimal.Decimal `json:"withdrawn"`
}

func NewGetBalance(e *entities.Account) GetBalance {
	return GetBalance{Balance: e.Balance, Withdrawn: e.Withdrawn}
}

type GetWithdrawals struct {
	ProcessedAt time.Time            `json:"processed_at"`
	Order       entities.OrderNumber `json:"order"`
	Sum         decimal.Decimal      `json:"sum"`
}

func NewGetWithdrawals(e *entities.Withdrawal) *GetWithdrawals {
	return &GetWithdrawals{
		Order:       e.Order,
		Sum:         e.Sum,
		ProcessedAt: e.ProcessedAt,
	}
}
