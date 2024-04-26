package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/KretovDmitry/gophermart/internal/application/errs"
	"github.com/KretovDmitry/gophermart/internal/domain/entities"
	"github.com/KretovDmitry/gophermart/internal/domain/entities/user"
	"github.com/KretovDmitry/gophermart/internal/domain/repositories"
	"github.com/KretovDmitry/gophermart/pkg/logger"
	trmsql "github.com/avito-tech/go-transaction-manager/drivers/sql/v2"
)

type OrderRepository struct {
	db     *sql.DB
	getter *trmsql.CtxGetter
	logger logger.Logger
}

func NewOrderRepository(db *sql.DB, getter *trmsql.CtxGetter, logger logger.Logger) (*OrderRepository, error) {
	if db == nil {
		return nil, errors.New("nil dependency: database")
	}
	if getter == nil {
		return nil, errors.New("nil dependency: transaction getter")
	}

	return &OrderRepository{db: db, getter: getter, logger: logger}, nil
}

var _ repositories.OrderRepository = (*OrderRepository)(nil)

func (r *OrderRepository) CreateOrder(ctx context.Context, id user.ID, num entities.OrderNumber) error {
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
	var userID user.ID

	err := r.getter.DefaultTrOrDB(ctx, r.db).
		QueryRowContext(ctx, query, id, num).
		Scan(&alreadyExists, &userID)
	if err != nil {
		return err
	}

	switch {
	case !alreadyExists && userID == id:
		return nil
	case alreadyExists && userID == id:
		return errs.ErrAlreadyExists
	case alreadyExists && userID != id:
		return errs.ErrDataConflict
	}

	return nil
}

func (r *OrderRepository) GetOrdersByUserID(ctx context.Context, id user.ID) ([]*entities.Order, error) {
	const query = "SELECT * FROM orders WHERE user_id = $1 ORDER BY uploadet_at DESC"

	rows, err := r.db.QueryContext(ctx, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errs.ErrNotFound
		}
		return nil, err
	}

	orders := make([]*entities.Order, 0)

	for rows.Next() {
		order := new(entities.Order)
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

	return orders, nil
}

func (r *OrderRepository) GetUnprocessedOrders(ctx context.Context, limit, offset int) ([]*entities.Order, error) {
	const query = "SELECT * FROM orders WHERE status > 'PROCESSED' ORDER BY id LIMIT $1 OFFSET $2;"

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errs.ErrNotFound
		}
		return nil, err
	}

	orders := make([]*entities.Order, 0, limit)

	for rows.Next() {
		order := new(entities.Order)
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

	return orders, nil
}

func (r *OrderRepository) UpdateOrder(ctx context.Context, info *entities.UpdateOrderInfo) (user.ID, error) {
	const query = "UPDATE orders SET status = $1, accrual = $2 WHERE number = $3 RETURNING user_id;"

	var userID user.ID = -1

	err := r.getter.DefaultTrOrDB(ctx, r.db).
		QueryRowContext(ctx, query, info.Status, info.Accrual, info.Number).
		Scan(&userID)
	if err != nil {
		return userID, err
	}

	return userID, nil
}
