package response

import "github.com/KretovDmitry/gophermart/internal/domain/entities"

type OrderUpdateInfo struct {
	Order  entities.OrderNumber
	Status string
}
