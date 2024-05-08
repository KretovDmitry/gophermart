package rest

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/KretovDmitry/gophermart/internal/application/errs"
	"github.com/KretovDmitry/gophermart/internal/application/interfaces"
	"github.com/KretovDmitry/gophermart/internal/domain/entities"
	"github.com/KretovDmitry/gophermart/internal/domain/entities/user"
	"github.com/KretovDmitry/gophermart/internal/interface/api/rest/header"
	"github.com/KretovDmitry/gophermart/internal/interface/api/rest/response"
	"github.com/KretovDmitry/gophermart/pkg/logger"
	"github.com/go-chi/chi/v5"
)

type OrderController struct {
	service interfaces.OrderService
	logger  logger.Logger
}

// NewOrderController registers http.Handlers with additional options.
func NewOrderController(
	service interfaces.OrderService, logger logger.Logger, options ChiServerOptions,
) {
	r := options.BaseRouter

	if r == nil {
		r = chi.NewRouter()
	}

	c := OrderController{
		service: service,
		logger:  logger,
	}

	r.Group(func(r chi.Router) {
		for _, middleware := range options.Middlewares {
			r.Use(middleware)
		}
		r.Post(options.BaseURL+"/orders", c.CreateOrder)
		r.Get(options.BaseURL+"/orders", c.GetOrders)
	})
}

// Create new order (POST /api/user/orders HTTP1.1).
func (c *OrderController) CreateOrder(w http.ResponseWriter, r *http.Request) {
	// Check content type.
	if !header.IsTextPlainContentType(r) {
		c.ErrorHandlerFunc(w, r, fmt.Errorf("%w: invalid content type", errs.ErrInvalidRequest))
		return
	}

	// Read and close request body.
	defer r.Body.Close()

	bytes, err := io.ReadAll(r.Body)
	if err != nil {
		if errors.Is(err, io.EOF) {
			c.ErrorHandlerFunc(w, r, fmt.Errorf("%w: empty body", errs.ErrInvalidRequest))
			return
		}
		c.ErrorHandlerFunc(w, r, err)
		return
	}

	// Create selfvalidating order number entity.
	orderNumber, err := entities.NewOrderNumber(string(bytes))
	if err != nil {
		c.ErrorHandlerFunc(w, r, err)
		return
	}

	// Get user from context.
	user, found := user.FromContext(r.Context())
	if !found {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Create order.
	if err = c.service.CreateOrder(r.Context(), user.ID, orderNumber); err != nil {
		c.ErrorHandlerFunc(w, r, err)
		return
	}

	// Return 202 if everything is fine.
	w.WriteHeader(http.StatusAccepted)
}

// Get user orders (GET /api/user/orders HTTP/1.1).
func (c *OrderController) GetOrders(w http.ResponseWriter, r *http.Request) {
	// Get user from context.
	user, found := user.FromContext(r.Context())
	if !found {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Get all orders for the user.
	orders, err := c.service.GetOrders(r.Context(), user.ID)
	if err != nil {
		c.ErrorHandlerFunc(w, r, err)
		return
	}

	// Convert entities to handler response representation.
	res := make([]*response.GetOrders, len(orders))
	for i, order := range orders {
		res[i] = response.NewGetOrdersFromOrderEntity(order)
	}

	w.Header().Set("Content-Type", "application/json")

	// Encode and return them. Status 200.
	if err = json.NewEncoder(w).Encode(res); err != nil {
		c.ErrorHandlerFunc(w, r, err)
		return
	}
}

// ErrorHandlerFunc handles sending of an error in the JSON format,
// writing appropriate status code and handling the failure to marshal that.
func (c *OrderController) ErrorHandlerFunc(w http.ResponseWriter, _ *http.Request, err error) {
	errJSON := errs.JSON{Error: err.Error()}
	code := http.StatusInternalServerError

	switch {
	// Status OK (200).
	// I used this error to distinguish between an attempt to create the same
	// order by the same user in which case we should return 200 OK and
	// actually the first time new order creation when status 202
	// Accepted is returned (method returns nil err in this case).
	case errors.Is(err, errs.ErrAlreadyExists):
		code = http.StatusOK

	// Status No Content (204).
	case errors.Is(err, errs.ErrNotFound):
		code = http.StatusNoContent

	// Status Bad Request (400).
	case errors.Is(err, errs.ErrInvalidRequest):
		code = http.StatusBadRequest

	// Status Payment Required (402).
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

	c.logger.Errorf("order controller [%d]: %s", code, err)

	if err = json.NewEncoder(w).Encode(errJSON); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
