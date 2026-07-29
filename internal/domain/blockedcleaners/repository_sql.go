package blockedcleaners

import (
	"context"
	"database/sql"
	"time"
)

type SQLRepository struct {
	db *sql.DB
}

func NewSQLRepository(db *sql.DB) *SQLRepository {
	return &SQLRepository{db: db}
}

func (r *SQLRepository) Create(ctx context.Context, block *BlockedCleaner) error {
	query := `
		INSERT INTO blocked_cleaners (
			client_id,
			cleaner_id,
			reason,
			created_at
		)
			VALUES ($1, $2, $3, $4)
			RETURNING id, created_at 
	`

	now := time.Now()

	return r.db.QueryRowContext(
		ctx,
		query, 
		block.ClientID,
		block.CleanerID,
		block.Reason,
		now,
	).Scan(
		&block.ID,
		&block.CreatedAt,
	)
}
func (r *SQLRepository) Delete(ctx context.Context, clientID uint, cleanerID uint) error {
	query := `
		DELETE FROM blocked_cleaners
		WHERE client_id = $1 
		AND cleaner_id = $2 
	`
	result, err := r.db.ExecContext(ctx, query, clientID, cleanerID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrBlockedNotFound
	}
	
	return err
}


func (r *SQLRepository) ListByClientID(ctx context.Context, clientID uint) ([]BlockedCleaner, error) {
	query := `
		SELECT 
			id,
			client_id,
			cleaner_id,
			reason,
			created_at 
		FROM blocked_cleaners 
		WHERE client_id = $1 
		ORDER BY created_at DESC 	
	`

	rows, err := r.db.QueryContext(ctx, query, clientID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var blocks []BlockedCleaner

	for rows.Next() {
		var block BlockedCleaner

		err := rows.Scan(
			&block.ID,
			&block.ClientID,
			&block.CleanerID,
			&block.Reason,
			&block.CreatedAt,
		)

		if err != nil {
			return nil, err
		}

		blocks = append(blocks, block)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return blocks, nil
}
func (r *SQLRepository) Exists(ctx context.Context, clientID uint, cleanerID uint) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT 1 
			FROM blocked_cleaners
			WHERE client_id = $1  
			AND cleaner_id = $2 
		)
	`

	var exists bool

	err := r.db.QueryRowContext(ctx, query, clientID, cleanerID).Scan(&exists)

	return exists, err
}