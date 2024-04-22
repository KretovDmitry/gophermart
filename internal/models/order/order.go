package order

import (
	"time"

	"github.com/shopspring/decimal"
)

type OrderStatus string

const (
	INVALID    OrderStatus = "INVALID"
	PROCESSED  OrderStatus = "PROCESSED"
	NEW        OrderStatus = "NEW"
	PROCESSING OrderStatus = "PROCESSING"
)

type Order struct {
	ID         int             `json:"id"`
	UserID     int             `json:"user_id"`
	Number     string          `json:"number"`
	Status     OrderStatus     `json:"status"`
	Accrual    decimal.Decimal `json:"accrual"`
	UploadetAt time.Time       `json:"uploadet_at"`
}
