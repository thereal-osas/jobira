package reviews

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type SQLRepository struct {
	db *sql.DB
}

func NewSQLRepository(db *sql.DB) *SQLRepository {
	return &SQLRepository{
		db: db,
	}
}

func (r *SQLRepository) Create(ctx context.Context, review *Review) error {
	query := `
		INSERT INTO reviews (
			booking_id,
			job_id,
			cleaner_id,
			client_id,
			rating,
			comment,
			created_at,
			updated_at
		)
		VALUES (
			$1,
			$2,
			$3,
			$4,
			$5,
			$6,
			$7,
			$7
		)
		RETURNING
			id,
			created_at,
			updated_at
	`

	now := time.Now()

	return r.db.QueryRowContext(
		ctx,
		query,
		review.BookingID,
		review.JobID,
		review.CleanerID,
		review.ClientID,
		review.Rating,
		review.Comment,
		now,
	).Scan(
		&review.ID,
		&review.CreatedAt,
		&review.UpdatedAt,
	)
}

func (r *SQLRepository) ListByCleanerID(ctx context.Context, cleanerID uint) ([]Review, error) {
	query := `
		SELECT
			id,
			cleaner_id,
			client_id,
			job_id,
			rating,
			comment,
			created_at,
			updated_at
		FROM reviews
		WHERE cleaner_id = $1 
		ORDER BY created_at DESC 	
	`

	rows, err := r.db.QueryContext(ctx, query, cleanerID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var reviews []Review

	for rows.Next() {
		var review Review

		err := rows.Scan(
			&review.ID,
			&review.CleanerID,
			&review.ClientID,
			&review.JobID,
			&review.Rating,
			&review.Comment,
			&review.CreatedAt,
			&review.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		reviews = append(reviews, review)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return reviews, nil
}

func (r *SQLRepository) ListByClientID(ctx context.Context, clientID uint) ([]Review, error) {
	query := `
		SELECT
			id,
			cleaner_id,
			client_id, 
			job_id,
			rating, 
			comment,
			created_at,
			updated_at
		FROM reviews
		WHERE client_id = $1 
		ORDER BY created_at DESC 	
	`

	rows, err := r.db.QueryContext(ctx, query, clientID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var reviews []Review

	for rows.Next() {
		var review Review

		err := rows.Scan(
			&review.ID,
			&review.CleanerID,
			&review.ClientID,
			&review.JobID,
			&review.Rating,
			&review.Comment,
			&review.CreatedAt,
			&review.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		reviews = append(reviews, review)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return reviews, nil
}

func (r *SQLRepository) GetJobClientID(ctx context.Context, jobID uint) (uint, error) {
	query := `
		SELECT client_id
		FROM jobs 
		WHERE id  = $1 
		LIMIT 1
	`

	var clientID uint

	err := r.db.QueryRowContext(ctx, query, jobID).Scan(&clientID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrInvalidInput
	}

	if err != nil {
		return 0, err
	}

	return clientID, nil
}

func (r *SQLRepository) GetAcceptedCleanerID(ctx context.Context, jobID uint) (uint, error) {
	query := `
		SELECT cleaner_id
		FROM applications 
		WHERE job_id = $1 
		AND status = 'accepted'
		LIMIT 1
	`

	var cleanerID uint

	err := r.db.QueryRowContext(ctx, query, jobID).Scan(&cleanerID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrInvalidInput
	}

	if err != nil {
		return 0, err
	}

	return cleanerID, nil
}

func (r *SQLRepository) GetJobStatus(ctx context.Context, jobID uint) (string, error) {
	query := `
		SELECT status
		FROM jobs
		WHERE id = $1
		LIMIT 1
	`

	var status string

	err := r.db.QueryRowContext(ctx, query, jobID).Scan(&status)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrInvalidInput
	}

	if err != nil {
		return "", err
	}

	return status, nil
}
