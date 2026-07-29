package availability

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

func (r *SQLRepository) Create(ctx context.Context, availability *CleanerAvailability) error {
	query := `
		INSERT INTO cleaner_availability (
			cleaner_id, 
			available_date,
			start_time,
			end_time,
			status,
			notes,
			created_at,
			updated_at 
		)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $7)
			RETURNING id, created_at, updated_at
	`

	now := time.Now()

	err := r.db.QueryRowContext(
		ctx, query,
		availability.CleanerID,
		availability.AvailableDate,
		availability.StartTime,
		availability.EndTime,
		availability.Status,
		availability.Notes,
		now,
	).Scan(
		&availability.ID,
		&availability.CreatedAt,
		&availability.UpdatedAt,
	)

	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "duplicate key") {
			return ErrAvailabilityExists
		}

		return err
	}

	return nil
}

func (r *SQLRepository) GetByID(ctx context.Context, id uint) (*CleanerAvailability, error) {
	query := `
		SELECT 
			id, 
			cleaner_id,
			available_date::text,
			start_time,
			end_time,
			status,
			COALESCE(notes, ''),
			created_at, 
			updated_at
		FROM cleaner_availability
		WHERE id = $1
		LIMIT 1	
	`

	var availability CleanerAvailability

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&availability.ID,
		&availability.CleanerID,
		&availability.AvailableDate,
		&availability.StartTime,
		&availability.EndTime,
		&availability.Status,
		&availability.Notes,
		&availability.CreatedAt,
		&availability.UpdatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrAvailabilityNotFound
	}

	if err != nil {
		return nil, err
	}

	return &availability, nil
}

func (r *SQLRepository) ListByCleanerID(ctx context.Context, cleanerID uint) ([]CleanerAvailability, error) {
	query := `
		SELECT
			id, 
			cleaner_id,
			available_date::text,
			start_time,
			end_time,
			status,
			COALESCE(notes, ''),
			created_at, 
			updated_at
		FROM cleaner_availability
		WHERE cleaner_id = $1
		ORDER BY available_date ASC, start_time ASC	
	`

	rows, err := r.db.QueryContext(ctx, query, cleanerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []CleanerAvailability

	for rows.Next() {
		var availability CleanerAvailability

		err := rows.Scan(
			&availability.ID,
			&availability.CleanerID,
			&availability.AvailableDate,
			&availability.StartTime,
			&availability.EndTime,
			&availability.Status,
			&availability.Notes,
			&availability.CreatedAt,
			&availability.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		records = append(records, availability)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return records, nil
}

func (r *SQLRepository) Update(ctx context.Context, availability *CleanerAvailability) error {
	query := `
		UPDATE cleaner_availability
		SET 
			available_date = $1,
			start_time = $2,
			end_time = $3,
			status = $4,
			notes = $5,
			updated_at = $6
		WHERE id = $7	
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		availability.AvailableDate,
		availability.StartTime,
		availability.EndTime,
		availability.Status,
		availability.Notes,
		time.Now(),
		availability.ID,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrAvailabilityNotFound
	}

	return nil
}

func (r *SQLRepository) Delete(ctx context.Context, id uint) error {
	query := ` 
		DELETE FROM cleaner_availability
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
		return ErrAvailabilityNotFound
	}

	return nil
}
