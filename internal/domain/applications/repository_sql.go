package applications

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
	return &SQLRepository{
		db: db,
	}
}

func (r *SQLRepository) Create(ctx context.Context, application *Application) error {
	query := `
		INSERT INTO applications (
			job_id,
			cleaner_id,
			cover_message,
			proposed_rate,
			status,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id	
	`

	now := time.Now()

	application.Status = "pending"
	application.CreatedAt = now
	application.UpdatedAt = now

	err := r.db.QueryRowContext(
		ctx,
		query,
		application.JobID,
		application.CleanerID,
		application.CoverMessage,
		application.ProposedRate,
		application.Status,
		application.CreatedAt,
		application.UpdatedAt,
	).Scan(&application.ID)

	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "duplicate key") {
			return ErrAlreadyApplied
		}

		return err
	}

	return nil
}

func (r *SQLRepository) GetByID(ctx context.Context, id uint) (*Application, error) {
	query := `
		SELECT
			id, 
			job_id,
			cleaner_id,
			cover_message,
			proposed_rate,
			status,
			created_at,
			updated_at
		FROM applications
		WHERE id = $1
		LIMIT 1	
	`

	var application Application

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&application.ID,
		&application.JobID,
		&application.CleanerID,
		&application.CoverMessage,
		&application.ProposedRate,
		&application.Status,
		&application.CreatedAt,
		&application.UpdatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrApplicationNotFound
	}

	if err != nil {
		return nil, err
	}

	return &application, nil
}

func (r *SQLRepository) ListByJobID(ctx context.Context, jobID uint) ([]Application, error) {
	query := `
		SELECT
			id,
			job_id,
			cleaner_id,
			cover_message,
			proposed_rate,
			status,
			created_at,
			updated_at
		FROM applications
		WHERE job_id = $1
		ORDER BY created_at DESC	
	`

	rows, err := r.db.QueryContext(ctx, query, jobID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var applications []Application

	for rows.Next() {
		var application Application

		err := rows.Scan(
			&application.ID,
			&application.JobID,
			&application.CleanerID,
			&application.CoverMessage,
			&application.ProposedRate,
			&application.Status,
			&application.CreatedAt,
			&application.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		applications = append(applications, application)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return applications, nil
}

func (r *SQLRepository) ListByCleanerID(ctx context.Context, cleanerID uint) ([]Application, error) {
	query := `
		SELECT
			id,
			job_id,
			cleaner_id,
			cover_message,
			proposed_rate,
			status,
			created_at,
			updated_at
		FROM applications
		WHERE cleaner_id = $1
		ORDER BY created_at DESC	
	`

	rows, err := r.db.QueryContext(ctx, query, cleanerID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var applications []Application

	for rows.Next() {
		var application Application

		err := rows.Scan(
			&application.ID,
			&application.JobID,
			&application.CleanerID,
			&application.CoverMessage,
			&application.ProposedRate,
			&application.Status,
			&application.CreatedAt,
			&application.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		applications = append(applications, application)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return applications, nil
}

func (r *SQLRepository) UpdateStatus(ctx context.Context, id uint, status string) error {
	query := `
		UPDATE applications
		SET status = $1, updated_at = $2
		WHERE id = $3
	`

	result, err := r.db.ExecContext(ctx, query, status, time.Now(), id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrApplicationNotFound
	}

	return nil
}

func (r *SQLRepository) GetJobClientID(ctx context.Context, jobID uint) (uint, error) {
	query := `
		SELECT client_id
		FROM jobs
		WHERE id = $1
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

func (r *SQLRepository) GetJobTitle(ctx context.Context, jobID uint) (string, error) {
	query := `
		SELECT title
		FROM jobs
		WHERE id = $1 
		LIMIT 1
	`

	var title string

	err := r.db.QueryRowContext(ctx, query, jobID).Scan(&title)
	if err != nil {
		return "", err
	}

	return title, nil
}

func (r *SQLRepository) GetUserEmail(ctx context.Context, userID uint) (string, error) {
	query := `
		SELECT email
		FROM users
		WHERE id = $1 
		LIMIT 1
	`

	var email string

	err := r.db.QueryRowContext(ctx, query, userID).Scan(&email)
	if err != nil {
		return "", err
	}

	return email, nil
}

func (r *SQLRepository) GetCleanerEmailByID(ctx context.Context, cleanerID uint) (string, error) {
	return r.GetUserEmail(ctx, cleanerID)
}
