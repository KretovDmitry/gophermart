package entities

import (
	"github.com/KretovDmitry/gophermart/internal/domain/entities/user"
	"github.com/shopspring/decimal"
)

type OperationType string

const (
	ACCRUAL    OperationType = "ACCRUAL"
	WITHDRAWAL OperationType = "WITHDRAWAL"
)

type Operation struct {
	UserID user.ID
	Type   OperationType
	Order  OrderNumber
	Sum    decimal.Decimal
}

func NewWithdrawOperation(id user.ID, order OrderNumber, sum decimal.Decimal) *Operation {
	return &Operation{
		UserID: id,
		Type:   WITHDRAWAL,
		Order:  order,
		Sum:    sum,
	}
}
