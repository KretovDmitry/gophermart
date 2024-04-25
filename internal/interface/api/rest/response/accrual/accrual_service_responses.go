package accrual

import (
	"github.com/shopspring/decimal"
)

type OrderStatus string

const (
	REGISTERED OrderStatus = "REGISTERED" // DB => "PROCESSING"
	PROCESSING OrderStatus = "PROCESSING" // DB => "PROCESSING"
	INVALID    OrderStatus = "INVALID"    // DB => "INVALID"
	PROCESSED  OrderStatus = "PROCESSED"  // DB => "PROCESSED"
)

type UpdateOrderInfo struct {
	Order   string          `json:"order"`
	Status  OrderStatus     `json:"status"`
	Accrual decimal.Decimal `json:"accrual"`
}
