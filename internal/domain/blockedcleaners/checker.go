package blockedcleaners

import (
	"context"
	"database/sql"
)

type Checker struct {
	db *sql.DB
}

func NewChecker(db *sql.DB) *Checker {
	return &Checker{db: db}
}

func (c *Checker) IsBlocked(ctx context.Context, clientID uint, cleanerID uint) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT 1 
			FROM blocked_cleaners
			WHERE client_id = $1 
			AND cleaner_id = $2 
		)
	`

	var exists bool

	err := c.db.QueryRowContext(
		ctx,
		query, 
		clientID,
		cleanerID, 
	).Scan(&exists)

	return exists, err
}