package order

import (
	"time"

	"github.com/shopspring/decimal"
)

type Status string

const (
	INVALID    Status = "INVALID"
	PROCESSED  Status = "PROCESSED"
	NEW        Status = "NEW"
	PROCESSING Status = "PROCESSING"
)

type Order struct {
	ID         int             `json:"id"`
	UserID     int             `json:"user_id"`
	Number     string          `json:"number"`
	Status     Status          `json:"status"`
	Accrual    decimal.Decimal `json:"accrual"`
	UploadetAt time.Time       `json:"uploadet_at"`
}
