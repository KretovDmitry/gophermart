package postgres

import (
	"database/sql"
	"fmt"

	"github.com/KretovDmitry/gophermart/pkg/logger"
	_ "github.com/jackc/pgx/v5/stdlib"
	sqldblogger "github.com/simukti/sqldb-logger"
)

func Connect(dsn string, logger logger.Logger) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open the database: %w", err)
	}

	// Log every query to the database.
	db = sqldblogger.OpenDriver(dsn, db.Driver(), logger)

	// Check connectivity and DSN correctness.
	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to connect to the database: %w", err)
	}

	return db, nil
}
