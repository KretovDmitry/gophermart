package response

import (
	"time"

	"github.com/KretovDmitry/gophermart/internal/domain/entities"
)

type GetBalance struct {
	Balance   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

func NewGetBalance(e *entities.Account) GetBalance {
	return GetBalance{
		Balance:   e.Balance.InexactFloat64(),
		Withdrawn: e.Withdrawn.InexactFloat64(),
	}
}

type GetWithdrawals struct {
	ProcessedAt time.Time            `json:"processed_at"`
	Order       entities.OrderNumber `json:"order"`
	Sum         float64              `json:"sum"`
}

func NewGetWithdrawals(e *entities.Withdrawal) *GetWithdrawals {
	return &GetWithdrawals{
		Order:       e.Order,
		Sum:         e.Sum.InexactFloat64(),
		ProcessedAt: e.ProcessedAt,
	}
}
