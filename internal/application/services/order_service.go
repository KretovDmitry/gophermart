package services

import (
	"context"
	"errors"

	"github.com/KretovDmitry/gophermart-loyalty-service/internal/application/interfaces"
	"github.com/KretovDmitry/gophermart-loyalty-service/internal/domain/entities"
	"github.com/KretovDmitry/gophermart-loyalty-service/internal/domain/entities/user"
	"github.com/KretovDmitry/gophermart-loyalty-service/internal/domain/repositories"
	"github.com/KretovDmitry/gophermart-loyalty-service/pkg/logger"
	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
)

type OrderService struct {
	repo   repositories.OrderRepository
	trm    *manager.Manager
	logger logger.Logger
}

func NewOrderService(
	repo repositories.OrderRepository,
	trm *manager.Manager,
	logger logger.Logger,
) (*OrderService, error) {
	if trm == nil {
		return nil, errors.New("nil dependency: transaction manager")
	}
	return &OrderService{
		repo:   repo,
		trm:    trm,
		logger: logger,
	}, nil
}

var _ interfaces.OrderService = (*OrderService)(nil)

// Create new order for user.
func (s *OrderService) CreateOrder(ctx context.Context, id user.ID, num entities.OrderNumber) error {
	if err := s.repo.CreateOrder(ctx, id, num); err != nil {
		return err
	}

	return nil
}

// Get user's orders.
func (s *OrderService) GetOrders(ctx context.Context, id user.ID) ([]*entities.Order, error) {
	orders, err := s.repo.GetOrdersByUserID(ctx, id)
	if err != nil {
		return nil, err
	}

	return orders, nil
}
