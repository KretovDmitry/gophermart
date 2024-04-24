package entities

import (
	"time"

	"github.com/shopspring/decimal"
)

type Withdrawal struct {
	Order       OrderNumber
	Sum         decimal.Decimal
	ProcessedAt time.Time
}
