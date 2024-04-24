package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/KretovDmitry/gophermart-loyalty-service/internal/application/errs"
	"github.com/KretovDmitry/gophermart-loyalty-service/internal/domain/entities"
	"github.com/KretovDmitry/gophermart-loyalty-service/internal/domain/entities/user"
	"github.com/KretovDmitry/gophermart-loyalty-service/internal/domain/repositories"
	"github.com/KretovDmitry/gophermart-loyalty-service/pkg/logger"
	trmsql "github.com/avito-tech/go-transaction-manager/drivers/sql/v2"
	"github.com/shopspring/decimal"
)

type AccountRepository struct {
	db     *sql.DB
	getter *trmsql.CtxGetter
	logger logger.Logger
}

func NewAccountRepository(db *sql.DB, getter *trmsql.CtxGetter, logger logger.Logger) (*AccountRepository, error) {
	if db == nil {
		return nil, errors.New("nil dependency: database")
	}
	if getter == nil {
		return nil, errors.New("nil dependency: transaction getter")
	}

	return &AccountRepository{db: db, getter: getter, logger: logger}, nil
}

var _ repositories.AccountRepository = (*AccountRepository)(nil)

func (r *AccountRepository) GetAccountByUserID(ctx context.Context, id user.ID) (*entities.Account, error) {
	const query = "SELECT * FROM accounts WHERE user_id = $1"

	account := new(entities.Account)

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

func (r *AccountRepository) Withdraw(ctx context.Context, sum decimal.Decimal, userID user.ID) error {
	const query = `
		UPDATE accounts SET 
			balance = balance - $1,
			withdrawn = withdrawn + $1
		WHERE user_id = $2 
			RETURNING balance;
	`

	var updatedBalance decimal.Decimal

	err := r.getter.DefaultTrOrDB(ctx, r.db).QueryRowContext(ctx, query, sum, userID).Scan(&updatedBalance)
	if err != nil {
		return err
	}

	if updatedBalance.LessThan(decimal.NewFromInt(0)) {
		return errs.ErrNotEnoughFunds
	}

	return nil
}

func (r *AccountRepository) SaveAccountOperation(ctx context.Context, op *entities.Operation) error {
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

func (r *AccountRepository) GetWithdrawals(ctx context.Context, id user.ID) ([]*entities.Withdrawal, error) {
	const query = `
		SELECT order_number, sum, processed_at FROM account_operations
		WHERE operation = 'WITHDRAWAL' AND account_id = (
			SELECT id FROM accounts WHERE user_id = $1
		)
		ORDER BY processed_at DESC;
	`

	withdrawals := make([]*entities.Withdrawal, 0)

	rows, err := r.getter.DefaultTrOrDB(ctx, r.db).QueryContext(ctx, query, id)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		w := new(entities.Withdrawal)
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
