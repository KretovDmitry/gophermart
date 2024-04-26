package response

import (
	"time"

	"github.com/KretovDmitry/gophermart/internal/domain/entities"
)

type GetOrders struct {
	Number     entities.OrderNumber `json:"number"`
	Status     entities.OrderStatus `json:"status"`
	Accrual    float64              `json:"accrual,omitempty"`
	UploadetAt time.Time            `json:"uploadet_at"`
}

func NewGetOrdersFromOrderEntity(e *entities.Order) *GetOrders {
	return &GetOrders{
		Number:     e.Number,
		Status:     e.Status,
		Accrual:    e.Accrual.InexactFloat64(),
		UploadetAt: e.UploadetAt,
	}
}
