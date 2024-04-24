package account

import (
	"time"

	"github.com/shopspring/decimal"
)

type OperationType string

const (
	ACCRUAL    OperationType = "ACCRUAL"
	WITHDRAWAL OperationType = "WITHDRAWAL"
)

type Account struct {
	ID        int             `json:"id"`
	UserID    int             `json:"user_id"`
	Balance   decimal.Decimal `json:"current"`
	Withdrawn decimal.Decimal `json:"withdrawn"`
}

type Operation struct {
	UserID int
	Type   OperationType
	Order  string
	Sum    decimal.Decimal
}

type Withdrawal struct {
	Order       string          `json:"order"`
	Sum         decimal.Decimal `json:"sum"`
	ProcessedAt time.Time       `json:"processed_at"`
}
