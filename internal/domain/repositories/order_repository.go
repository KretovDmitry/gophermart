package repositories

import (
	"context"

	"github.com/KretovDmitry/gophermart/internal/domain/entities"
	"github.com/KretovDmitry/gophermart/internal/domain/entities/user"
)

type OrderRepository interface {
	CreateOrder(context.Context, user.ID, entities.OrderNumber) error
	GetOrdersByUserID(context.Context, user.ID) ([]*entities.Order, error)
	GetUnprocessedOrders(ctx context.Context, limit, offset int) ([]*entities.Order, error)
	UpdateOrder(context.Context, *entities.UpdateOrderInfo) (user.ID, error)
}
