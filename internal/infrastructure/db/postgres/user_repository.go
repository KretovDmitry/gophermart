package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/KretovDmitry/gophermart/internal/application/errs"
	"github.com/KretovDmitry/gophermart/internal/domain/entities/user"
	"github.com/KretovDmitry/gophermart/internal/domain/repositories"
	"github.com/KretovDmitry/gophermart/pkg/logger"
	trmsql "github.com/avito-tech/go-transaction-manager/drivers/sql/v2"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

type UserRepository struct {
	db     *sql.DB
	getter *trmsql.CtxGetter
	logger logger.Logger
}

func NewUserRepository(db *sql.DB, getter *trmsql.CtxGetter, logger logger.Logger) (*UserRepository, error) {
	if db == nil {
		return nil, errors.New("nil dependency: database")
	}
	if getter == nil {
		return nil, errors.New("nil dependency: transaction getter")
	}

	return &UserRepository{db: db, getter: getter, logger: logger}, nil
}

var _ repositories.UserRepository = (*UserRepository)(nil)

func (r *UserRepository) GetUserByID(ctx context.Context, id user.ID) (*user.User, error) {
	const query = "SELECT * FROM users WHERE id = $1"

	u := new(user.User)

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&u.ID,
		&u.Login,
		&u.Password,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errs.ErrNotFound
		}
		return nil, err
	}

	return u, nil
}

func (r *UserRepository) GetUserByLogin(ctx context.Context, login string) (*user.User, error) {
	const query = "SELECT * FROM users WHERE login = $1"

	u := new(user.User)

	err := r.db.QueryRowContext(ctx, query, login).Scan(
		&u.ID,
		&u.Login,
		&u.Password,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errs.ErrNotFound
		}
		return nil, err
	}

	return u, nil
}

func (r *UserRepository) CreateUser(ctx context.Context, login, password string) (user.ID, error) {
	const query = "INSERT INTO users (login, password) VALUES ($1, $2) RETURNING id"

	var id user.ID

	err := r.getter.DefaultTrOrDB(ctx, r.db).QueryRowContext(ctx, query, login, password).Scan(&id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == pgerrcode.UniqueViolation {
				return -1, fmt.Errorf("%w: login %q already exists", errs.ErrDataConflict, login)
			}
		}
		return -1, fmt.Errorf("create user: %w", err)
	}

	return id, nil
}
