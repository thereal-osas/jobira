package db

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/rodrigueghenda/jobira/internal/config"
)

func NewPostgres(cfg config.Config) (*sql.DB, error) {
	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, err
	}

	return db, nil

}