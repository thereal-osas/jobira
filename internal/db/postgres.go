package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"github.com/rodrigueghenda/jobira/internal/config"
)

const (
	maxOpenConnections    = 25
	maxIdleConnections    = 10
	connectionMaxLifetime = 30 * time.Minute
	connectionMaxIdleTime = 5 * time.Minute
	pingTimeout           = 5 * time.Second
)

func NewPostgres(
	cfg config.Config,
) (*sql.DB, error) {
	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf(
			"DATABASE_URL is required",
		)
	}

	database, err := sql.Open(
		"postgres",
		cfg.DatabaseURL,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"open postgres: %w",
			err,
		)
	}

	configurePool(database)

	ctx, cancel := context.WithTimeout(
		context.Background(),
		pingTimeout,
	)
	defer cancel()

	if err := database.PingContext(ctx); err != nil {
		_ = database.Close()

		return nil, fmt.Errorf(
			"ping postgres: %w",
			err,
		)
	}

	return database, nil
}

func configurePool(
	database *sql.DB,
) {
	database.SetMaxOpenConns(
		maxOpenConnections,
	)

	database.SetMaxIdleConns(
		maxIdleConnections,
	)

	database.SetConnMaxLifetime(
		connectionMaxLifetime,
	)

	database.SetConnMaxIdleTime(
		connectionMaxIdleTime,
	)
}
