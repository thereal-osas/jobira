package recentviews

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

func (r *SQLRepository) RecordView(ctx context.Context, clientID uint, cleanerID uint) error {
	query := `
		INSERT INTO recently_viewed_cleaners (
			client_id,
			cleaner_id,
			viewed_at
		)
			VALUES ($1, $2, $3)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		clientID,
		cleanerID,
		time.Now(),
	)

	return err
}

func (r *SQLRepository) ListByClientID(ctx context.Context, clientID uint) ([]RecentlyViewedCleaner, error) {
	query := `
		SELECT
			id,
			client_id,
			cleaner_id,
			viewed_at 
		FROM recently_viewed_cleaners 
		WHERE client_id = $1 
		ORDER BY viewed_at DESC 
		LIMIT 20 	
	`

	rows, err := r.db.QueryContext(ctx, query, clientID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var views []RecentlyViewedCleaner

	for rows.Next() {
		var view RecentlyViewedCleaner

		err := rows.Scan(
			&view.ID,
			&view.ClientID,
			&view.CleanerID,
			&view.ViewedAt,
		)

		if err != nil {
			return nil, err
		}

		views = append(views, view)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return views, nil
}

func (r *SQLRepository) CountByCleanerID(ctx context.Context, cleanerID uint) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM recently_viewed_cleaners
		WHERE cleaner_id = $1
	`

	var count int

	err := r.db.QueryRowContext(
		ctx,
		query,
		cleanerID,
	).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *SQLRepository) CountUniqueViewersByCleanerID(ctx context.Context, cleanerID uint) (int, error) {
	query := `
		SELECT COUNT(DISTINCT client_id)
		FROM recently_viewed_cleaners
		WHERE cleaner_id = $1
	`

	var count int

	err := r.db.QueryRowContext(
		ctx,
		query,
		cleanerID,
	).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *SQLRepository) CountByCleanerIDSince(ctx context.Context, cleanerID uint, since time.Time) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM recently_viewed_cleaners
		WHERE cleaner_id = $1
		  AND viewed_at >= $2
	`

	var count int

	err := r.db.QueryRowContext(
		ctx,
		query,
		cleanerID,
		since,
	).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *SQLRepository) CountByClenaerIDBetween(ctx context.Context, cleanerID uint, from time.Time, to time.Time) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM recently_viewed_cleaners
		WHERE cleaner_id = $1
		  AND viewed_at >= $2
		  AND viewed_at < $3
	`

	var count int

	err := r.db.QueryRowContext(
		ctx,
		query,
		cleanerID,
		from,
		to,
	).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *SQLRepository) CountByCleanerIDBetween(ctx context.Context, cleanerID uint, from time.Time, to time.Time) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM recently_viewed_cleaners
		WHERE cleaner_id = $1
		  AND viewed_at >= $2
		  AND viewed_at < $3
	`

	var count int

	err := r.db.QueryRowContext(
		ctx,
		query,
		cleanerID,
		from,
		to,
	).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}
