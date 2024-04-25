package repositories

import (
	"context"

	"github.com/KretovDmitry/gophermart/internal/domain/entities"
	"github.com/KretovDmitry/gophermart/internal/domain/entities/user"
)

type OrderRepository interface {
	CreateOrder(context.Context, user.ID, entities.OrderNumber) error
	GetOrdersByUserID(context.Context, user.ID) ([]*entities.Order, error)
}
