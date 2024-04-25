package interfaces

import (
	"context"

	"github.com/KretovDmitry/gophermart/internal/domain/entities"
	"github.com/KretovDmitry/gophermart/internal/domain/entities/user"
)

// OrderService represents all service actions.
type OrderService interface {
	CreateOrder(context.Context, user.ID, entities.OrderNumber) error
	GetOrders(context.Context, user.ID) ([]*entities.Order, error)
}
