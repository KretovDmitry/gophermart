package reward

import (
	"context"
	"database/sql"
	"errors"

	"github.com/KretovDmitry/gophermart-loyalty-service/internal/models/errs"
	"github.com/KretovDmitry/gophermart-loyalty-service/internal/models/order"
	"github.com/KretovDmitry/gophermart-loyalty-service/pkg/logger"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

type Repository interface {
	CreateOrder(ctx context.Context, order *order.Order) error
	GetOrderByNumber(ctx context.Context, num int) (*order.Order, error)
}

type Repo struct {
	db     *sql.DB
	logger logger.Logger
}

func NewRepository(db *sql.DB, logger logger.Logger) (*Repo, error) {
	if db == nil {
		return nil, errors.New("nil dependency: database")
	}

	return &Repo{
		db:     db,
		logger: logger,
	}, nil
}

var _ Repository = (*Repo)(nil)

func (r *Repo) CreateOrder(ctx context.Context, order *order.Order) error {
	const query = "INSERT INTO orders (user_id, number) VALUES ($1, $2);"

	_, err := r.db.ExecContext(ctx, query, order.UserID, order.Number)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == pgerrcode.UniqueViolation {
				return errs.ErrDataConflict
			}
		}
		return err
	}

	return nil
}

func (r *Repo) GetOrderByNumber(ctx context.Context, num int) (*order.Order, error) {
	const query = "SELECT * FROM orders WHERE number = $1"

	order := new(order.Order)

	err := r.db.QueryRowContext(ctx, query, num).Scan(
		&order.ID,
		&order.UserID,
		&order.Number,
		&order.Status,
		&order.UploadetAt,
	)
	if err != nil {
		return nil, err
	}

	return order, nil
}
