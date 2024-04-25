package rest

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/KretovDmitry/gophermart/internal/application/errs"
	"github.com/KretovDmitry/gophermart/internal/application/interfaces"
	"github.com/KretovDmitry/gophermart/internal/application/params"
	"github.com/KretovDmitry/gophermart/internal/domain/entities"
	"github.com/KretovDmitry/gophermart/internal/domain/entities/user"
	"github.com/KretovDmitry/gophermart/internal/interface/api/rest/header"
	"github.com/KretovDmitry/gophermart/internal/interface/api/rest/request"
	"github.com/KretovDmitry/gophermart/internal/interface/api/rest/response"
	"github.com/go-chi/chi/v5"
	"github.com/shopspring/decimal"
)

type AccountController struct {
	service interfaces.AccountService
}

// NewAccountController registers http.Handlers with additional options.
func NewAccountController(service interfaces.AccountService, options ChiServerOptions) {
	r := options.BaseRouter

	if r == nil {
		r = chi.NewRouter()
	}

	c := AccountController{
		service: service,
	}

	r.Group(func(r chi.Router) {
		for _, middleware := range options.Middlewares {
			r.Use(middleware)
		}
		r.Get(options.BaseURL+"/balance", c.GetBalance)
		r.Post(options.BaseURL+"/balance/withdraw", c.Withdraw)
		r.Get(options.BaseURL+"/withdrawals", c.GetWithdrawals)
	})
}

// Get user balance (GET /api/user/balance HTTP/1.1).
func (c *AccountController) GetBalance(w http.ResponseWriter, r *http.Request) {
	// Get user from context.
	user, found := user.FromContext(r.Context())
	if !found {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Get user's account.
	account, err := c.service.GetAccount(r.Context(), user.ID)
	if err != nil {
		c.ErrorHandlerFunc(w, r, err)
		return
	}

	// Create response payload.
	response := response.NewGetBalance(account)

	// Encode and return. Status 200.
	if err = json.NewEncoder(w).Encode(response); err != nil {
		c.ErrorHandlerFunc(w, r, err)
		return
	}
}

// Withdraw (POST /api/user/balance/withdraw HTTP/1.1).
func (c *AccountController) Withdraw(w http.ResponseWriter, r *http.Request) {
	// Check content type.
	if !header.IsApplicationJSONContentType(r) {
		c.ErrorHandlerFunc(w, r, fmt.Errorf("%w: invalid content type", errs.ErrInvalidRequest))
		return
	}

	// Read, decode and close request body.
	defer r.Body.Close()

	var payload request.Withdraw

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		c.ErrorHandlerFunc(w, r, checkJSONDecodeError(err))
		return
	}

	// Create selfvalidating order number entity.
	orderNumber, err := entities.NewOrderNumber(payload.Order)
	if err != nil {
		c.ErrorHandlerFunc(w, r, err)
		return
	}

	// Check if sum is meaningful.
	if payload.Sum.LessThanOrEqual(decimal.NewFromInt(0)) {
		c.ErrorHandlerFunc(w, r, fmt.Errorf("%w: invalid sum", errs.ErrInvalidRequest))
		return
	}

	// Get user from context.
	u, found := user.FromContext(r.Context())
	if !found {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Create params for Withdraw interface method.
	params := params.NewWithraw(u.ID, orderNumber, payload.Sum)

	// Return 200 OK if there is no error.
	if err = c.service.Withdraw(r.Context(), params); err != nil {
		c.ErrorHandlerFunc(w, r, err)
		return
	}
}

// Get all user withdrawals (GET /api/user/withdrawals HTTP/1.1).
func (s *AccountController) GetWithdrawals(w http.ResponseWriter, r *http.Request) {
	// Get user from context.
	user, found := user.FromContext(r.Context())
	if !found {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Get all withdrawals made by the user.
	withdrawals, err := s.service.GetWithdrawals(r.Context(), user.ID)
	if err != nil {
		s.ErrorHandlerFunc(w, r, err)
		return
	}

	// Convert entities to handler response representation.
	res := make([]*response.GetWithdrawals, len(withdrawals))
	for i, w := range withdrawals {
		res[i] = response.NewGetWithdrawals(w)
	}

	// Encode them. Status 200 OK.
	if err = json.NewEncoder(w).Encode(res); err != nil {
		s.ErrorHandlerFunc(w, r, err)
		return
	}
}

// ErrorHandlerFunc handles sending of an error in the JSON format,
// writing appropriate status code and handling the failure to marshal that.
func (c *AccountController) ErrorHandlerFunc(w http.ResponseWriter, _ *http.Request, err error) {
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
