package response

import (
	"time"

	"github.com/KretovDmitry/gophermart-loyalty-service/internal/domain/entities"
	"github.com/shopspring/decimal"
)

type GetOrders struct {
	Number     entities.OrderNumber `json:"number"`
	Status     entities.OrderStatus `json:"status"`
	Accrual    string               `json:"accrual,omitempty"`
	UploadetAt time.Time            `json:"uploadet_at"`
}

func NewGetOrdersFromOrderEntity(e *entities.Order) *GetOrders {
	var accrual string
	if !e.Accrual.Equal(decimal.NewFromInt(0)) {
		accrual = e.Accrual.String()
	}
	return &GetOrders{
		Number:     e.Number,
		Status:     e.Status,
		Accrual:    accrual,
		UploadetAt: e.UploadetAt,
	}
}
