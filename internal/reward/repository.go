package reward

import (
	"context"
	"database/sql"
	"errors"

	"github.com/KretovDmitry/gophermart-loyalty-service/internal/models/account"
	"github.com/KretovDmitry/gophermart-loyalty-service/internal/models/errs"
	"github.com/KretovDmitry/gophermart-loyalty-service/internal/models/order"
	"github.com/KretovDmitry/gophermart-loyalty-service/pkg/logger"
	trmsql "github.com/avito-tech/go-transaction-manager/drivers/sql/v2"
	"github.com/shopspring/decimal"
)

type Repository interface {
	CreateOrder(ctx context.Context, order *order.Order) error
	GetOrdersByUserID(ctx context.Context, id int) ([]*order.Order, error)
	GetAccountByUserID(ctx context.Context, id int) (*account.Account, error)
	Witdraw(ctx context.Context, sum float64, userID int) error
	SaveAccountOperation(ctx context.Context, op *account.Operation) error
	GetWithdrawals(ctx context.Context, userID int) ([]*account.Withdrawal, error)
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

	err := r.getter.DefaultTrOrDB(ctx, r.db).
		QueryRowContext(ctx, query, order.UserID, order.Number).
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

func (r *Repo) GetAccountByUserID(ctx context.Context, id int) (*account.Account, error) {
	const query = "SELECT * FROM accounts WHERE user_id = $1"

	account := new(account.Account)

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&account.ID,
		&account.UserID,
		&account.Balance,
		&account.Withdrawn,
	)
	if err != nil {
		return nil, err
	}

	return account, nil
}

func (r *Repo) Witdraw(ctx context.Context, sum float64, userID int) error {
	const query = `
		UPDATE accounts SET 
			balance = balance - $1,
			withdrawn = withdrawn + $1
		WHERE user_id = $2 
			RETURNING balance;`

	var updatedBalance decimal.Decimal
	decimalSum := decimal.NewFromFloat(sum)

	err := r.getter.DefaultTrOrDB(ctx, r.db).
		QueryRowContext(ctx, query, decimalSum, userID).
		Scan(&updatedBalance)
	if err != nil {
		return err
	}

	if updatedBalance.LessThan(decimal.NewFromInt(0)) {
		return errs.ErrNotEnoughFunds
	}

	return nil
}

func (r *Repo) SaveAccountOperation(ctx context.Context, op *account.Operation) error {
	const query = `
		INSERT INTO account_operations (account_id, operation, order_number, sum)
		VALUES ((SELECT id FROM accounts WHERE user_id = $1), $2, $3, $4);
	`

	_, err := r.getter.DefaultTrOrDB(ctx, r.db).
		ExecContext(ctx, query, op.UserID, op.Type, op.Order, op.Sum)
	if err != nil {
		return err
	}

	return nil
}

func (r *Repo) GetWithdrawals(ctx context.Context, userID int) ([]*account.Withdrawal, error) {
	const query = `
		SELECT order_number, sum, processed_at FROM account_operations
		WHERE operation = 'WITHDRAWAL' AND account_id = (
			SELECT id FROM accounts WHERE user_id = $1
		)
		ORDER BY processed_at DESC;
	`

	withdrawals := make([]*account.Withdrawal, 0)

	rows, err := r.getter.DefaultTrOrDB(ctx, r.db).QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		w := new(account.Withdrawal)
		err = rows.Scan(
			&w.Order,
			&w.Sum,
			&w.ProcessedAt,
		)
		if err != nil {
			return nil, err
		}

		withdrawals = append(withdrawals, w)
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

	if len(withdrawals) == 0 {
		return nil, errs.ErrNotFound
	}

	return withdrawals, nil
}
