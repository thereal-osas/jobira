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

func (r *SQLRepository) HasConflict(ctx context.Context, cleanerID uint, availableDate string, startTime string, endTime string, excludeID uint) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT 1 
			FROM cleaner_availability
			WHERE cleaner_id = $1 
				AND available_date = $2 
				AND id <> $3 
				AND status <> 'unavailable'
				AND start_time < $4 
				AND end_time > $5
		)
	`

	var conflict bool

	err := r.db.QueryRowContext(
		ctx,
		query,
		cleanerID,
		availableDate,
		excludeID,
		endTime,
		startTime,
	).Scan(&conflict)

	if err != nil {
		return false, err
	}

	return conflict, nil
}

func (r *SQLRepository) CreateBlock(ctx context.Context, block *AvailabilityBlock) error {
	query := `
		INSERT INTO availability_blocks (
			cleaner_id,
			start_at, 
			end_at, 
			reason, 
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $5)
		RETURNING id 	
	`

	now := time.Now()

	block.CreatedAt = now
	block.UpdatedAt = now

	err := r.db.QueryRowContext(
		ctx,
		query,
		block.CleanerID,
		block.StartAt,
		block.EndAt,
		block.Reason,
		now,
	).Scan(&block.ID)

	return err
}

func (r *SQLRepository) ListBlocksByCleanerID(ctx context.Context, cleanerID uint) ([]AvailabilityBlock, error) {
	query := `
		SELECT
			id, 
			cleaner_id,
			start_at,
			end_at, 
			reason, 
			created_at,
			updated_at
		FROM availability_blocks
		WHERE cleaner_id = $1 
		ORDER BY start_at ASC 	
	`

	rows, err := r.db.QueryContext(ctx, query, cleanerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var blocks []AvailabilityBlock

	for rows.Next() {
		var block AvailabilityBlock

		err := rows.Scan(
			&block.ID,
			&block.CleanerID,
			&block.StartAt,
			&block.EndAt,
			&block.Reason,
			&block.CreatedAt,
			&block.UpdatedAt,
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

func (r *SQLRepository) DeleteBlock(ctx context.Context, blockID uint, cleanerID uint) error {
	query := ` 
		DELETE FROM availability_blocks
		WHERE id = $1 
			AND cleaner_id = $2
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		blockID,
		cleanerID,
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
