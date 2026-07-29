package db

import (
	"context"
	"database/sql"
)

func WithTx(ctx context.Context, database *sql.DB, fn func(*sql.Tx) error) error {
	tx, err := database.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	if err := fn(tx); err != nil {
		_ = tx.Rollback() 
		return err
	}

	return tx.Commit()
}