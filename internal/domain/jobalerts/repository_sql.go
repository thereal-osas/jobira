package jobalerts

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
	return &SQLRepository{db: db}
}

func (r *SQLRepository) Create(ctx context.Context, alert *JobAlert) error {
	query := `
		INSERT INTO job_alerts (
			user_id,
			location, 
			job_type,
			minimum_budget,
			is_active, 
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $6)
		RETURNING id, created_at, updated_at	
	`

	now := time.Now()

	err := r.db.QueryRowContext(
		ctx,
		query,
		alert.UserID,
		alert.Location,
		alert.JobType,
		alert.MinimumBudget,
		alert.IsActive,
		now,
	).Scan(
		&alert.ID,
		&alert.CreatedAt,
		&alert.UpdatedAt,
	)

	if err != nil {
		return err
	}

	return nil
}

func (r *SQLRepository) GetByID(ctx context.Context, id uint) (*JobAlert, error) {
	query := `
		SELECT 
			id, 
			user_id,
			location,
			job_type,
			minimum_budget, 
			is_active, 
			created_at, 
			updated_at
		FROM job_alerts
		WHERE id = $1
		LIMIT 1	
	`

	var alert JobAlert

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&alert.ID,
		&alert.UserID,
		&alert.Location,
		&alert.JobType,
		&alert.MinimumBudget,
		&alert.IsActive,
		&alert.CreatedAt,
		&alert.UpdatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrAlertNotFound
	}

	if err != nil {
		return nil, err
	}

	return &alert, nil
}

func (r *SQLRepository) ListByUserID(ctx context.Context, userID uint) ([]JobAlert, error) {
	query := `
		SELECT 
			id,
			user_id,
			location, 
			job_type, 
			minimum_budget, 
			is_active, 
			created_at,
			updated_at 
		FROM job_alerts
		WHERE user_id = $1
		ORDER BY created_at DESC 	
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var alerts []JobAlert

	for rows.Next() {
		var alert JobAlert

		err := rows.Scan(
			&alert.ID,
			&alert.UserID,
			&alert.Location,
			&alert.JobType,
			&alert.MinimumBudget,
			&alert.IsActive,
			&alert.CreatedAt,
			&alert.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		alerts = append(alerts, alert)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return alerts, nil
}

func (r *SQLRepository) Update(ctx context.Context, alert *JobAlert) error {
	query := `
		UPDATE job_alerts
		SET
			location = $1,
			job_type = $2,
			minimum_budget = $3,
			is_active = $4,
			updated_at = $5 
		WHERE id = $6	
	`

	now := time.Now()

	result, err := r.db.ExecContext(
		ctx,
		query,
		alert.Location,
		alert.JobType,
		alert.MinimumBudget,
		alert.IsActive,
		now,
		alert.ID,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrAlertNotFound
	}

	alert.UpdatedAt = now

	return nil
}

func (r *SQLRepository) Delete(ctx context.Context, id uint) error {
	query := `
		DELETE FROM job_alerts
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrAlertNotFound
	}

	return nil
}
