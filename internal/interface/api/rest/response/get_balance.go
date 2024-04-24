package response

import "github.com/shopspring/decimal"

type GetBalance struct {
	Balance   decimal.Decimal `json:"current"`
	Withdrawn decimal.Decimal `json:"withdrawn"`
}

func NewGetBalance(balance, withdrawn decimal.Decimal) GetBalance {
	return GetBalance{Balance: balance, Withdrawn: withdrawn}
}
