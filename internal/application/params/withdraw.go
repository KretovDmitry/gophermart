package params

import (
	"github.com/KretovDmitry/gophermart-loyalty-service/internal/domain/entities"
	"github.com/KretovDmitry/gophermart-loyalty-service/internal/domain/entities/user"
	"github.com/shopspring/decimal"
)

type Withdraw struct {
	UserID user.ID
	Order  entities.OrderNumber
	Sum    decimal.Decimal
}

func NewWithraw(userID user.ID, order entities.OrderNumber, sum decimal.Decimal) *Withdraw {
	return &Withdraw{UserID: userID, Order: order, Sum: sum}
}
