package entities

import (
	"time"

	"github.com/KretovDmitry/gophermart-loyalty-service/internal/application/errs"
	"github.com/KretovDmitry/gophermart-loyalty-service/internal/domain/entities/user"
	"github.com/KretovDmitry/gophermart-loyalty-service/pkg/luhn"
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
	ID         int
	UserID     user.ID
	Number     OrderNumber
	Status     OrderStatus
	Accrual    decimal.Decimal
	UploadetAt time.Time
}

func NewOrder(id user.ID, order OrderNumber) *Order {
	return &Order{
		UserID: id,
		Number: order,
		Status: NEW,
	}
}

type OrderNumber string

func NewOrderNumber(num string) (OrderNumber, error) {
	if err := luhn.Validate(num); err != nil {
		return "", errs.ErrInvalidOrderNumber
	}

	return OrderNumber(num), nil
}
