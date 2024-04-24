package services

import (
	"context"
	"errors"

	"github.com/KretovDmitry/gophermart-loyalty-service/internal/application/interfaces"
	"github.com/KretovDmitry/gophermart-loyalty-service/internal/application/params"
	"github.com/KretovDmitry/gophermart-loyalty-service/internal/config"
	"github.com/KretovDmitry/gophermart-loyalty-service/internal/domain/entities"
	"github.com/KretovDmitry/gophermart-loyalty-service/internal/domain/entities/user"
	"github.com/KretovDmitry/gophermart-loyalty-service/internal/domain/repositories"
	"github.com/KretovDmitry/gophermart-loyalty-service/pkg/logger"
	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
)

type AccountService struct {
	accountRepo repositories.AccountRepository
	orderRepo   repositories.OrderRepository
	trm         *manager.Manager
	logger      logger.Logger
	config      *config.Config
}

func NewAccountService(
	accountRepository repositories.AccountRepository,
	orderRepository repositories.OrderRepository,
	trm *manager.Manager,
	logger logger.Logger,
	config *config.Config,
) (*AccountService, error) {
	if config == nil {
		return nil, errors.New("nil dependency: config")
	}
	if trm == nil {
		return nil, errors.New("nil dependency: transaction manager")
	}
	return &AccountService{
		accountRepo: accountRepository,
		orderRepo:   orderRepository,
		trm:         trm,
		logger:      logger,
		config:      config,
	}, nil
}

var _ interfaces.AccountService = (*AccountService)(nil)

func (s *AccountService) GetAccount(ctx context.Context, id user.ID) (*entities.Account, error) {
	account, err := s.accountRepo.GetAccountByUserID(ctx, id)
	if err != nil {
		return nil, err
	}

	return account, nil
}

func (s *AccountService) Withdraw(ctx context.Context, params *params.Withdraw) error {
	return s.trm.Do(ctx, func(ctx context.Context) error {
		var err error

		if err = s.orderRepo.CreateOrder(ctx, params.UserID, params.Order); err != nil {
			return err
		}

		// Withdraw funds to the account of the new order.
		if err = s.accountRepo.Withdraw(ctx, params.Sum, params.UserID); err != nil {
			return err
		}

		// Write withdrawal to the operations history table.
		withdrawalOperation := entities.NewWithdrawOperation(params.UserID, params.Order, params.Sum)

		if err = s.accountRepo.SaveAccountOperation(ctx, withdrawalOperation); err != nil {
			return err
		}

		return nil
	})
}

func (s *AccountService) GetWithdrawals(ctx context.Context, id user.ID) ([]*entities.Withdrawal, error) {
	withdrawals, err := s.accountRepo.GetWithdrawals(ctx, id)
	if err != nil {
		return nil, err
	}

	return withdrawals, nil
}
