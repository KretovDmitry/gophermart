package reward

import (
	"context"
	"database/sql"
	"errors"

	"github.com/KretovDmitry/gophermart-loyalty-service/internal/models/errs"
	"github.com/KretovDmitry/gophermart-loyalty-service/internal/models/order"
	"github.com/KretovDmitry/gophermart-loyalty-service/pkg/logger"
)

type Repository interface {
	CreateOrder(ctx context.Context, order *order.Order) error
	GetOrderByNumber(ctx context.Context, num string) (*order.Order, error)
	GetOrdersByUserID(ctx context.Context, id int) ([]*order.Order, error)
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
		SELECT FALSE AS source, user_id FROM ins UNION ALL
		SELECT TRUE AS source, c.user_id FROM input_rows 
		JOIN orders c USING (number);
	`

	var alreadyExists bool
	var userID int

	err := r.db.QueryRowContext(ctx, query, order.UserID, order.Number).
		Scan(&alreadyExists, &userID)
	if err != nil {
		return err
	}

	switch {
	case !alreadyExists && userID == order.UserID:
		return nil
	case alreadyExists && userID == order.UserID:
		return errs.ErrAlreadyExists
	case alreadyExists && userID != order.UserID:
		return errs.ErrDataConflict
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
		&order.Accrual,
		&order.UploadetAt,
	)
	if err != nil {
		return nil, err
	}

	return order, nil
}

func (r *Repo) GetOrdersByUserID(ctx context.Context, id int) ([]*order.Order, error) {
	const query = "SELECT * FROM orders WHERE user_id = $1 ORDER BY uploadet_at DESC"

	rows, err := r.db.QueryContext(ctx, query, id)
	if err != nil {
		return nil, err
	}

	orders := make([]*order.Order, 0)

	for rows.Next() {
		order := new(order.Order)
		err = rows.Scan(
			&order.ID,
			&order.UserID,
			&order.Number,
			&order.Status,
			&order.Accrual,
			&order.UploadetAt,
		)
		if err != nil {
			return nil, err
		}

		orders = append(orders, order)
	}

	defer func() {
		if err = rows.Close(); err != nil {
			r.logger.Errorf("close rows: %s", err)
		}
	}()

	// Rows.Err will report the last error encountered by Rows.Scan.
	if err = rows.Err(); err != nil {
		return nil, err
	}

	if len(orders) == 0 {
		return nil, errs.ErrNotFound
	}

	return orders, nil
}
