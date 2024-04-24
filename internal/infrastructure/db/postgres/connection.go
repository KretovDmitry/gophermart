package postgres

import (
	"database/sql"
	"fmt"

	"github.com/KretovDmitry/gophermart-loyalty-service/internal/config"
	"github.com/KretovDmitry/gophermart-loyalty-service/pkg/logger"
	_ "github.com/jackc/pgx/v5/stdlib"
	sqldblogger "github.com/simukti/sqldb-logger"
)

func Connect(cfg *config.Config, logger logger.Logger) (*sql.DB, error) {
	db, err := sql.Open("pgx", cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to open the database: %w", err)
	}

	// Log every query to the database.
	db = sqldblogger.OpenDriver(cfg.DSN, db.Driver(), logger)

	return db, nil
}
