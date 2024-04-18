package claims

import "github.com/golang-jwt/jwt/v4"

type Auth struct {
	jwt.RegisteredClaims
	UserID int
}
