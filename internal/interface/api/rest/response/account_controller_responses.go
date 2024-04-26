package response

import (
	"time"

	"github.com/KretovDmitry/gophermart/internal/domain/entities"
)

type GetBalance struct {
	Balance   string `json:"current"`
	Withdrawn string `json:"withdrawn"`
}

func NewGetBalance(e *entities.Account) GetBalance {
	return GetBalance{
		Balance:   e.Balance.StringFixed(2),
		Withdrawn: e.Withdrawn.StringFixed(2),
	}
}

type GetWithdrawals struct {
	ProcessedAt time.Time            `json:"processed_at"`
	Order       entities.OrderNumber `json:"order"`
	Sum         string               `json:"sum"`
}

func NewGetWithdrawals(e *entities.Withdrawal) *GetWithdrawals {
	return &GetWithdrawals{
		Order:       e.Order,
		Sum:         e.Sum.StringFixed(2),
		ProcessedAt: e.ProcessedAt,
	}
}
