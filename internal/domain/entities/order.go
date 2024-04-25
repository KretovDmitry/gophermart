package entities

import (
	"time"

	"github.com/KretovDmitry/gophermart/internal/application/errs"
	"github.com/KretovDmitry/gophermart/internal/domain/entities/user"
	"github.com/KretovDmitry/gophermart/internal/interface/api/rest/response/accrual"
	"github.com/KretovDmitry/gophermart/pkg/luhn"
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

type UpdateOrderInfo struct {
	Number  OrderNumber
	Status  OrderStatus
	Accrual decimal.Decimal
}

func NewUpdateInfoFromResponse(r *accrual.UpdateOrderInfo) *UpdateOrderInfo {
	var newStatus OrderStatus
	switch r.Status {
	case accrual.PROCESSED:
		newStatus = PROCESSED
	case accrual.INVALID:
		newStatus = INVALID
	case accrual.PROCESSING, accrual.REGISTERED:
		newStatus = PROCESSING
	}

	return &UpdateOrderInfo{
		Number:  OrderNumber(r.Order),
		Status:  newStatus,
		Accrual: r.Accrual,
	}
}
