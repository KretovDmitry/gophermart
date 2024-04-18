package jwt

import (
	"fmt"
	"strings"
	"time"

	"github.com/KretovDmitry/gophermart-loyalty-service/internal/models/claims"
	"github.com/golang-jwt/jwt/v4"
)

// BuildString creates a JWT string for the given user ID and token expiration time.
func BuildString(userID int, secret string, tokenExp time.Duration) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims.Auth{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenExp)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		UserID: userID,
	})

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Bearer %s", tokenString), nil
}

// GetUserID extracts the user ID from a JWT token.
func GetUserID(tokenString, secret string) (int, error) {
	claims := new(claims.Auth)

	tokenString = strings.TrimPrefix(tokenString, "Bearer ")

	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(token *jwt.Token) (interface{}, error) {
			// Verify that the token method is HS256
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return 0, fmt.Errorf(
					"unexpected signing method: %v", token.Header["alg"],
				)
			}

			// Return the secret key
			return []byte(secret), nil
		})
	// Check for errors
	if err != nil {
		return 0, fmt.Errorf("error parsing token: %w", err)
	}

	// Check if the token is valid
	if !token.Valid {
		return 0, fmt.Errorf("invalid token: %w", err)
	}

	// Return the user ID
	return claims.UserID, nil
}
