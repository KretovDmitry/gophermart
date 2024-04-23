package account

import "github.com/shopspring/decimal"

type Account struct {
	ID        int             `json:"id"`
	UserID    int             `json:"user_id"`
	Balance   decimal.Decimal `json:"current"`
	Withdrawn decimal.Decimal `json:"withdrawn"`
}
