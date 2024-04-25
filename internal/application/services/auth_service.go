package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/KretovDmitry/gophermart-loyalty-service/internal/application/errs"
	"github.com/KretovDmitry/gophermart-loyalty-service/internal/config"
	"github.com/KretovDmitry/gophermart-loyalty-service/internal/domain/entities"
	"github.com/KretovDmitry/gophermart-loyalty-service/internal/domain/entities/user"
	"github.com/KretovDmitry/gophermart-loyalty-service/internal/domain/repositories"
	"github.com/KretovDmitry/gophermart-loyalty-service/pkg/logger"
	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo    repositories.UserRepository
	accountRepo repositories.AccountRepository
	trm         *manager.Manager
	logger      logger.Logger
	config      *config.Config
}

func NewAuthService(
	userRepo repositories.UserRepository,
	accountRepo repositories.AccountRepository,
	trm *manager.Manager,
	logger logger.Logger,
	config *config.Config,
) (*AuthService, error) {
	if config == nil {
		return nil, errors.New("nil dependency: config")
	}
	if trm == nil {
		return nil, errors.New("nil dependency: transaction manager")
	}
	return &AuthService{
		userRepo:    userRepo,
		accountRepo: accountRepo,
		trm:         trm,
		logger:      logger,
		config:      config,
	}, nil
}

// Registr user.
func (s *AuthService) Register(ctx context.Context, login, password string) (user.ID, error) {
	var userID user.ID = -1

	// Careate password hash.
	hashPassword, err := bcrypt.GenerateFromPassword([]byte(password), s.config.PasswordHashCost)
	if err != nil {
		return userID, fmt.Errorf("hash password: %w", err)
	}

	// Create user and his account.
	err = s.trm.Do(ctx, func(ctx context.Context) error {
		userID, err = s.userRepo.CreateUser(ctx, login, string(hashPassword))
		if err != nil {
			return err
		}

		if err = s.accountRepo.CreateAccount(ctx, userID); err != nil {
			return err
		}

		return nil
	})

	return userID, err
}

// Authenticate user by login.
func (s *AuthService) Login(ctx context.Context, login, password string) (*user.User, error) {
	// Retrieve user from the database with provided login.
	user, err := s.userRepo.GetUserByLogin(ctx, login)
	if err != nil {
		return nil, fmt.Errorf("get user %q: %w", login, err)
	}

	// Compare stored and provided passwords.
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return nil, fmt.Errorf("%w: password", errs.ErrInvalidCredentials)
		}
		return nil, fmt.Errorf("compare passwords: %w", err)
	}

	return user, nil
}

// BuildAuthToken creates a JWT string for the given user ID.
func (s *AuthService) BuildAuthToken(userID user.ID) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, entities.AuthClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.config.JWT.Expiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		UserID: userID,
	})

	tokenString, err := token.SignedString([]byte(s.config.JWT.SigningKey))
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Bearer %s", tokenString), nil
}

// GetUserFromToken extracts the user ID from a JWT token and returns user.
func (s *AuthService) GetUserFromToken(ctx context.Context, tokenString string) (*user.User, error) {
	claims := new(entities.AuthClaims)

	tokenString = strings.TrimPrefix(tokenString, "Bearer ")

	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(token *jwt.Token) (interface{}, error) {
			// Verify that the token method is HS256.
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf(
					"%w: unexpected signing method: %v",
					errs.ErrInvalidCredentials, token.Header["alg"],
				)
			}

			// Return the secret key.
			return []byte(s.config.JWT.SigningKey), nil
		})
	// Check for errors.
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errs.ErrInvalidCredentials, err)
	}

	// Check if the token is valid.
	if !token.Valid {
		return nil, fmt.Errorf("%w: %w", errs.ErrInvalidCredentials, err)
	}

	// Get user from repo by id.
	user, err := s.userRepo.GetUserByID(ctx, claims.UserID)
	if err != nil {
		return nil, err
	}
	// Return the user.
	return user, nil
}
