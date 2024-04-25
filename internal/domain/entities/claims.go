package entities

import (
	"github.com/KretovDmitry/gophermart-loyalty-service/internal/domain/entities/user"
	"github.com/golang-jwt/jwt/v4"
)

type AuthClaims struct {
	jwt.RegisteredClaims
	UserID user.ID
}
