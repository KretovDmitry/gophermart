package entities

import (
	"github.com/shopspring/decimal"
)

type Account struct {
	ID        int
	UserID    int
	Balance   decimal.Decimal
	Withdrawn decimal.Decimal
}
