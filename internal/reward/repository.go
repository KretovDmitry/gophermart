package reward

import (
	"context"
	"database/sql"
	"errors"

	"github.com/KretovDmitry/gophermart-loyalty-service/internal/models/order"
	"github.com/KretovDmitry/gophermart-loyalty-service/pkg/logger"
)

type Repository interface {
	CreateOrder(ctx context.Context, order *order.Order) error
	GetOrderByNumber(ctx context.Context, num string) (*order.Order, error)
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
	const query = `
		WITH input_rows(user_id, number) AS (
			VALUES ($1::integer, $2::text)
		),
		ins AS (
			INSERT INTO orders (user_id, number)
			SELECT * FROM input_rows
			ON CONFLICT (number) DO NOTHING
			RETURNING user_id
		) 
		SELECT user_id FROM ins UNION ALL
		SELECT c.user_id FROM input_rows 
		JOIN orders c USING (number);
	`

	err := r.db.QueryRowContext(ctx, query, order.UserID, order.Number).Scan(&order.UserID)
	if err != nil {
		return err
	}

	return nil
}

func (r *Repo) GetOrderByNumber(ctx context.Context, num string) (*order.Order, error) {
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
