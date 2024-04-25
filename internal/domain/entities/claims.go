package entities

import (
	"github.com/KretovDmitry/gophermart/internal/domain/entities/user"
	"github.com/golang-jwt/jwt/v4"
)

type AuthClaims struct {
	jwt.RegisteredClaims
	UserID user.ID
}
