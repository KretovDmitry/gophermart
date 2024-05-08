package services

import (
	"context"

	"github.com/KretovDmitry/gophermart/internal/application/interfaces"
	"github.com/KretovDmitry/gophermart/internal/domain/entities"
	"github.com/KretovDmitry/gophermart/internal/domain/entities/user"
	"github.com/KretovDmitry/gophermart/internal/domain/repositories"
	"github.com/KretovDmitry/gophermart/pkg/logger"
)

type OrderService struct {
	repo   repositories.OrderRepository
	logger logger.Logger
}

func NewOrderService(repo repositories.OrderRepository, logger logger.Logger) (*OrderService, error) {
	return &OrderService{repo: repo, logger: logger}, nil
}

var _ interfaces.OrderService = (*OrderService)(nil)

// Create new order for user.
func (s *OrderService) CreateOrder(ctx context.Context, id user.ID, num entities.OrderNumber) error {
	return s.repo.CreateOrder(ctx, id, num)
}

// Get user's orders.
func (s *OrderService) GetOrders(ctx context.Context, id user.ID) ([]*entities.Order, error) {
	return s.repo.GetOrdersByUserID(ctx, id)
}
