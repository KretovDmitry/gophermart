package reward

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/KretovDmitry/gophermart-loyalty-service/internal/config"
	"github.com/KretovDmitry/gophermart-loyalty-service/internal/models/account"
	"github.com/KretovDmitry/gophermart-loyalty-service/internal/models/errs"
	"github.com/KretovDmitry/gophermart-loyalty-service/internal/models/order"
	"github.com/KretovDmitry/gophermart-loyalty-service/internal/models/user"
	"github.com/KretovDmitry/gophermart-loyalty-service/pkg/logger"
	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
	"github.com/shopspring/decimal"
)

type Service struct {
	repo   Repository
	trm    *manager.Manager
	logger logger.Logger
	config *config.Config
}

func NewService(repo Repository, trm *manager.Manager, logger logger.Logger, config *config.Config) (*Service, error) {
	if config == nil {
		return nil, errors.New("nil dependency: config")
	}
	if trm == nil {
		return nil, errors.New("nil dependency: transaction manager")
	}
	return &Service{repo: repo, trm: trm, logger: logger, config: config}, nil
}

var _ ServerInterface = (*Service)(nil)

// Create new order (POST /api/user/orders).
func (s *Service) CreateOrder(w http.ResponseWriter, r *http.Request, params PostOrderParams) {
	u, found := user.FromContext(r.Context())
	if !found {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	order := &order.Order{
		UserID: u.ID,
		Number: params.Number,
		Status: order.NEW,
	}

	if err := s.repo.CreateOrder(r.Context(), order); err != nil {
		ErrorHandlerFunc(w, r, err)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

// Get user orders (GET /api/user/orders HTTP/1.1).
func (s *Service) GetOrders(w http.ResponseWriter, r *http.Request) {
	u, found := user.FromContext(r.Context())
	if !found {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	orders, err := s.repo.GetOrdersByUserID(r.Context(), u.ID)
	if err != nil {
		ErrorHandlerFunc(w, r, err)
		return
	}

	if err = json.NewEncoder(w).Encode(orders); err != nil {
		ErrorHandlerFunc(w, r, err)
		return
	}
}

// Get user account data (GET /api/user/balance HTTP/1.1).
func (s *Service) GetAccount(w http.ResponseWriter, r *http.Request) {
	u, found := user.FromContext(r.Context())
	if !found {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	account, err := s.repo.GetAccountByUserID(r.Context(), u.ID)
	if err != nil {
		ErrorHandlerFunc(w, r, err)
		return
	}

	if err = json.NewEncoder(w).Encode(account); err != nil {
		ErrorHandlerFunc(w, r, err)
		return
	}
}

// Withdraw (POST /api/user/balance/withdraw HTTP/1.1).
func (s *Service) Withdraw(w http.ResponseWriter, r *http.Request, params WithdrawParams) {
	u, found := user.FromContext(r.Context())
	if !found {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	order := &order.Order{
		UserID: u.ID,
		Number: params.Order,
		Status: order.NEW,
	}

	accountOperation := &account.Operation{
		UserID: u.ID,
		Type:   account.WITHDRAWAL,
		Order:  params.Order,
		Sum:    decimal.NewFromFloat(params.Sum),
	}

	var err error
	err = s.trm.Do(r.Context(), func(ctx context.Context) error {
		if err = s.repo.CreateOrder(ctx, order); err != nil {
			return err
		}
		if err = s.repo.Witdraw(ctx, params.Sum, u.ID); err != nil {
			return err
		}
		if err = s.repo.SaveAccountOperation(ctx, accountOperation); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		ErrorHandlerFunc(w, r, err)
		return
	}
}

// Get all user withdrawals (GET /api/user/withdrawals HTTP/1.1).
func (s *Service) GetWithdrawals(w http.ResponseWriter, r *http.Request) {
	u, found := user.FromContext(r.Context())
	if !found {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	withdrawals, err := s.repo.GetWithdrawals(r.Context(), u.ID)
	if err != nil {
		ErrorHandlerFunc(w, r, err)
		return
	}

	if err = json.NewEncoder(w).Encode(withdrawals); err != nil {
		ErrorHandlerFunc(w, r, err)
		return
	}
}

// ErrorHandlerFunc handles sending of an error in the JSON format,
// writing appropriate status code and handling the failure to marshal that.
func ErrorHandlerFunc(w http.ResponseWriter, _ *http.Request, err error) {
	errJSON := errs.JSON{Error: err.Error()}
	code := http.StatusInternalServerError

	switch {
	// Status OK (200).
	case errors.Is(err, errs.ErrAlreadyExists):
		code = http.StatusOK

	// Status No Content (204).
	case errors.Is(err, errs.ErrNotFound):
		code = http.StatusNoContent

	// Status Bad Request (400).
	case errors.Is(err, errs.ErrInvalidRequest):
		code = http.StatusBadRequest

	// Stats Payment Required (402).
	case errors.Is(err, errs.ErrNotEnoughFunds):
		code = http.StatusPaymentRequired

	// Status Conflict (409).
	case errors.Is(err, errs.ErrDataConflict):
		code = http.StatusConflict

	// Status Unproccessable Entity (422).
	case errors.Is(err, errs.ErrInvalidOrderNumber):
		code = http.StatusUnprocessableEntity
	}

	w.WriteHeader(code)

	if err = json.NewEncoder(w).Encode(errJSON); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
