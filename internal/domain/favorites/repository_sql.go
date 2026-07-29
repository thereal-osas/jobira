package favorites

import (
	"context"
	"database/sql"
)

type SQLRepository struct {
	db *sql.DB
}

func NewSQLRepository(db *sql.DB) *SQLRepository {
	return &SQLRepository{
		db: db,
	}
}

func (r *SQLRepository) Create(ctx context.Context, favorite *FavoriteCleaner) error {
	query := `
		INSERT INTO favorite_cleaners (
			client_id,
			cleaner_id
		)
		VALUES ($1, $2)
		RETURNING id, created_at  
	`

	return r.db.QueryRowContext(
		ctx,
		query,
		favorite.ClientID,
		favorite.CleanerID,
	).Scan(
		&favorite.ID,
		&favorite.CreatedAt,
	)
}

func (r *SQLRepository) Delete(ctx context.Context, clientID uint, cleanerID uint) error {
	query:= `
		DELETE FROM favorite_cleaners
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

func (r *SQLRepository) ListByClientID(ctx context.Context, clientID uint) ([]FavoriteCleaner, error) {
	query := `
		SELECT 
			id, 
			client_id,
			cleaner_id, 
			created_at 
		FROM favorite_cleaners 
		WHERE client_id = $1 
		ORDER BY created_at DESC 	
	`

	rows, err := r.db.QueryContext(ctx, query, clientID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var favorites []FavoriteCleaner

	for rows.Next() {
		var favorite FavoriteCleaner

		err := rows.Scan(
			&favorite.ID,
			&favorite.ClientID,
			&favorite.CleanerID,
			&favorite.CreatedAt,
		)

		if err != nil {
			return nil, err
		}

		favorites = append(favorites, favorite)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return favorites, nil
}

func (r *SQLRepository) Exists(ctx context.Context, clientID uint, cleanerID uint) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT 1
			FROM favorite_cleaners
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

