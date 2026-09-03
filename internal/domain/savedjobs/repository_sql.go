package savedjobs

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"
)

type SQLRepository struct {
	db *sql.DB
}

func NewSQLRepository(db *sql.DB) *SQLRepository {
	return &SQLRepository{db: db}
}

func (r *SQLRepository) Save(ctx context.Context, savedJob *SavedJob) error {
	query := `
		INSERT INTO saved_jobs (
			user_id,
			job_id,
			created_at
		)
		VALUES ($1, $2, $3)
		RETURNING id, created_at	
	`

	err := r.db.QueryRowContext(
		ctx,
		query,
		savedJob.UserID,
		savedJob.JobID,
		time.Now(),
	).Scan(
		&savedJob.ID,
		&savedJob.CreatedAt,
	)

	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "duplicate key") {
			return ErrAlreadySved
		}

		return err
	}

	return nil
}

func (r *SQLRepository) ListByUserID(ctx context.Context, userID uint) ([]SavedJob, error) {
	query := `
		SELECT
			id,
			user_id,
			job_id,
			created_at
		FROM saved_jobs
		WHERE user_id = $1 
		ORDER BY created_at DESC 	
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var savedJobs []SavedJob

	for rows.Next() {
		var savedJob SavedJob

		err := rows.Scan(
			&savedJob.ID,
			&savedJob.UserID,
			&savedJob.JobID,
			&savedJob.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		savedJobs = append(savedJobs, savedJob)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return savedJobs, nil
}

func (r *SQLRepository) Delete(ctx context.Context, userID uint, jobID uint) error {
	query := `
		DELETE FROM saved_jobs
		WHERE user_id = $1 
		AND job_id = $2
	`

	result, err := r.db.ExecContext(ctx, query, userID, jobID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrSavedJobNotFound
	}

	return nil
}

func (r *SQLRepository) GetByID(ctx context.Context, id uint) (*SavedJob, error) {
	query := `
		SELECT 
			id,
			user_id,
			job_id,
			created_at
		FROM saved_jobs
		WHERE id = $1 
		LIMIT 1	
	`

	var savedJob SavedJob

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&savedJob.ID,
		&savedJob.UserID,
		&savedJob.JobID,
		&savedJob.CreatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrSavedJobNotFound
	}

	if err != nil {
		return nil, err
	}

	return &savedJob, nil
}
