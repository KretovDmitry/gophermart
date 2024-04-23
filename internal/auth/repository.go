package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/KretovDmitry/gophermart-loyalty-service/internal/models/errs"
	"github.com/KretovDmitry/gophermart-loyalty-service/internal/models/user"
	"github.com/KretovDmitry/gophermart-loyalty-service/pkg/logger"
	trmsql "github.com/avito-tech/go-transaction-manager/drivers/sql/v2"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

type Repository interface {
	GetUserByID(ctx context.Context, userID int) (*user.User, error)
	GetUserByLogin(ctx context.Context, login string) (*user.User, error)
	CreateUser(ctx context.Context, login, password string) (id int, err error)
	CreateAccount(ctx context.Context, userID int) error
}

type Repo struct {
	db     *sql.DB
	getter *trmsql.CtxGetter
	logger logger.Logger
}

func NewRepository(db *sql.DB, getter *trmsql.CtxGetter, logger logger.Logger) (*Repo, error) {
	if db == nil {
		return nil, errors.New("nil dependency: database")
	}
	if getter == nil {
		return nil, errors.New("nil dependency: transaction getter")
	}

	return &Repo{db: db, getter: getter, logger: logger}, nil
}

var _ Repository = (*Repo)(nil)

func (r *Repo) GetUserByID(ctx context.Context, userID int) (*user.User, error) {
	const query = "SELECT * FROM users WHERE id = $1"

	u := new(user.User)

	err := r.db.QueryRowContext(ctx, query, userID).Scan(
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

func (r *Repo) GetUserByLogin(ctx context.Context, login string) (*user.User, error) {
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

func (r *Repo) CreateUser(ctx context.Context, login, password string) (int, error) {
	const query = "INSERT INTO users (login, password) VALUES ($1, $2) RETURNING id"

	var id int

	err := r.getter.DefaultTrOrDB(ctx, r.db).QueryRowContext(ctx, query, login, password).Scan(&id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == pgerrcode.UniqueViolation {
				return -1, errs.ErrDataConflict
			}
		}
		return -1, fmt.Errorf("create user: %w", err)
	}

	return id, nil
}

func (r *Repo) CreateAccount(ctx context.Context, userID int) error {
	const query = "INSERT INTO accounts (user_id) VALUES ($1)"

	_, err := r.getter.DefaultTrOrDB(ctx, r.db).ExecContext(ctx, query, userID)
	if err != nil {
		return err
	}

	return nil
}
