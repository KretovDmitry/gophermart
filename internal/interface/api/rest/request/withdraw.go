package request

import "github.com/shopspring/decimal"

type Withdraw struct {
	Order string          `json:"order"`
	Sum   decimal.Decimal `json:"sum"`
}
