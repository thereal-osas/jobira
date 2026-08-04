package preferredcleaners

import (
	"context"
	"database/sql"
)

type SQLRepository struct {
	db *sql.DB
}

func NewSQLRepository(db *sql.DB) *SQLRepository {
	return &SQLRepository{db: db}
}

func (r *SQLRepository) Create(ctx context.Context, preferred *PreferredCleaners) error {
	query := `
		INSERT INTO preferred_cleaners (
			client_id,
			cleaner_id
		)
			VALUES ($1, $2)
			RETURNING id, created_at 
	`

	return r.db.QueryRowContext(
		ctx,
		query,
		preferred.ClientID,
		preferred.CleanerID,
	).Scan(
		&preferred.ID,
		&preferred.CreatedAt,
	)
}

func (r *SQLRepository) Exists(ctx context.Context, clientID uint, cleanerID uint) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT 1 
			FROM preferred_cleaners
			WHERE client_id = $1 
			AND cleaner_id = $2 
		)
	`

	var exists bool

	err := r.db.QueryRowContext(
		ctx,
		query,
		clientID,
		cleanerID,
	).Scan(&exists)

	return exists, err
}

func (r *SQLRepository) Delete(ctx context.Context, clientID uint, cleanerID uint) error {
	query := `
		DELETE FROM preferred_cleaners
		WHERE client_id = $1 
		AND cleaner_id = $2 
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		clientID,
		cleanerID,
	)

	return err
}

func (r *SQLRepository) ListByClientID(ctx context.Context, clientID uint) ([]PreferredCleaners, error) {
	query := `
		SELECT 
			id,
			client_id,
			cleaner_id,
			created_at 
		FROM preferred_cleaners 
		WHERE client_id = $1 
		ORDER BY created_at DESC 	
	`

	rows, err := r.db.QueryContext(
		ctx,
		query,
		clientID,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var cleaners []PreferredCleaners

	for rows.Next() {
		var cleaner PreferredCleaners

		err := rows.Scan(
			&cleaner.ID,
			&cleaner.ClientID,
			&cleaner.CleanerID,
			&cleaner.CreatedAt,
		)

		if err != nil {
			return nil, err
		}

		cleaners = append(cleaners, cleaner)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return cleaners, nil

}
