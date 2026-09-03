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

func (r *SQLRepository) ListByCleanerIDRange(ctx context.Context, cleanerID uint, fromDate string, toDate string) ([]CleanerAvailability, error) {
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
			AND available_date >= $2 
			AND available_date <= $3 
		ORDER BY available_date ASC, start_time ASC		
		`

	rows, err := r.db.QueryContext(
		ctx,
		query,
		cleanerID,
		fromDate,
		toDate,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []CleanerAvailability

	for rows.Next() {
		var record CleanerAvailability

		if err := rows.Scan(
			&record.ID,
			&record.CleanerID,
			&record.AvailableDate,
			&record.StartTime,
			&record.EndTime,
			&record.Status,
			&record.Notes,
			&record.CreatedAt,
			&record.UpdatedAt,
		); err != nil {
			return nil, err
		}

		records = append(records, record)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return records, nil
}

func (r *SQLRepository) ListBlocksByCleanerIDRange(ctx context.Context, cleanerID uint, startAt time.Time, endAt time.Time) ([]AvailabilityBlock, error) {
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
			AND start_at < $3 
			AND end_at > $2 
		ORDER BY start_at ASC 		
	`

	rows, err := r.db.QueryContext(
		ctx,
		query,
		cleanerID,
		startAt,
		endAt,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var blocks []AvailabilityBlock

	for rows.Next() {
		var block AvailabilityBlock

		if err := rows.Scan(
			&block.ID,
			&block.CleanerID,
			&block.StartAt,
			&block.EndAt,
			&block.Reason,
			&block.CreatedAt,
			&block.UpdatedAt,
		); err != nil {
			return nil, err
		}

		blocks = append(blocks, block)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return blocks, nil
}

func (r *SQLRepository) CreateRecurring(ctx context.Context, recurring *RecurringAvailability) error {
	query := ` 
	INSERT INTO recurring_availability (
		cleaner_id, 
		weekday,
		start_time,
		end_time,
		status, 
		created_at, 
		updated_at
	) 
	VALUES ($1, $2, $3, $4, $5, $6, $6)
	RETURNING id, created_at, updated_at	
	`

	now := time.Now()

	return r.db.QueryRowContext(
		ctx,
		query,
		recurring.CleanerID,
		recurring.Weekday,
		recurring.StartTime,
		recurring.EndTime,
		recurring.Status,
		now,
	).Scan(
		&recurring.ID,
		&recurring.CreatedAt,
		&recurring.UpdatedAt,
	)
}

func (r *SQLRepository) ListRecurringByCleanerID(ctx context.Context, cleanerID uint) ([]RecurringAvailability, error) {
	query := `
		SELECT 	
			id, 
			cleaner_id, 
			weekday, 
			start_time, 
			end_time, 
			status, 
			created_at,
			updated_at 
		FROM recurring_availability 
		WHERE cleaner_id = $1 
		ORDER BY weekday ASC, start_time ASC 	
	`

	rows, err := r.db.QueryContext(
		ctx,
		query,
		cleanerID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []RecurringAvailability

	for rows.Next() {
		var record RecurringAvailability

		if err := rows.Scan(
			&record.ID,
			&record.CleanerID,
			&record.Weekday,
			&record.StartTime,
			&record.EndTime,
			&record.Status,
			&record.CreatedAt,
			&record.UpdatedAt,
		); err != nil {
			return nil, err
		}

		records = append(records, record)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return records, nil
}

func (r *SQLRepository) DeleteRecurring(ctx context.Context, recurringID uint, cleanerID uint) error {

	result, err := r.db.ExecContext(
		ctx,
		`
		DELETE FROM recurring_availability
		WHERE id = $1
			AND cleaner_id = $2
		`,
		recurringID,
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
		return ErrRecurringAvailabilityNotFound
	}

	return nil
}

func (r *SQLRepository) UpsertSettings(ctx context.Context, settings *AvailabilitySettings) error {
	query := `
		INSERT INTO availability_settings (
			cleaner_id,
			min_notice_minutes,
			min_booking_minutes,
			max_booking_minutes,
			buffer_minutes,
			booking_horizon_days,
			timezone,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $8)
		ON CONFLICT (cleaner_id)
		DO UPDATE SET 
			min_notice_minutes = EXCLUDED.min_notice_minutes, 
			min_booking_minutes = EXCLUDED.min_booking_minutes,
			max_booking_minutes = EXCLUDED.max_booking_minutes,
			buffer_minutes = EXCLUDED.buffer_minutes,
			booking_horizon_days = EXCLUDED.booking_horizon_days,
			timezone = EXCLUDED.timezone, 
			updated_at = EXCLUDED.updated_at
		RETURNING created_at, updated_at	
	`

	now := time.Now()

	return r.db.QueryRowContext(
		ctx,
		query,
		settings.CleanerID,
		settings.MinNoticeMinutes,
		settings.MinBookingMinutes,
		settings.MaxBookingMinutes,
		settings.BufferMinutes,
		settings.BookingHorizonDays,
		settings.Timezone,
		now,
	).Scan(
		&settings.CreatedAt,
		&settings.UpdatedAt,
	)
}

func (r *SQLRepository) GetSettings(ctx context.Context, cleanerID uint) (*AvailabilitySettings, error) {
	query := `
		SELECT 
			cleaner_id,
			min_notice_minutes,
			min_booking_minutes,
			max_booking_minutes,
			buffer_minutes,
			booking_horizon_days, 
			timezone,
			created_at,
			updated_at 
		FROM availability_settings
		WHERE cleaner_id = $1 
		LIMIT 1 	
	`

	var settings AvailabilitySettings

	err := r.db.QueryRowContext(
		ctx,
		query,
		cleanerID,
	).Scan(
		&settings.CleanerID,
		&settings.MinNoticeMinutes,
		&settings.MinBookingMinutes,
		&settings.MaxBookingMinutes,
		&settings.BufferMinutes,
		&settings.BookingHorizonDays,
		&settings.Timezone,
		&settings.CreatedAt,
		&settings.UpdatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrAvailabilitySettingsNotFound
	}

	if err != nil {
		return nil, err
	}

	return &settings, nil
}

func (r *SQLRepository) CreateOverride(ctx context.Context, override *AvailabilityOverride) error {
	query := `
		INSERT INTO availability_overrides (
			cleaner_id,
			available_date,
			start_time,
			end_time, 
			status, 
			reason, 
			created_at, 
			updated_at
		)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $7)
			RETURNING id, created_at, updated_at
	`

	now := time.Now()

	return r.db.QueryRowContext(
		ctx,
		query,
		override.CleanerID,
		override.AvailabilityDate,
		override.StartTime,
		override.EndTime,
		override.Status,
		override.Reason,
		now,
	).Scan(
		&override.ID,
		&override.CreatedAt,
		&override.UpdatedAt,
	)
}

func (r *SQLRepository) ListOverridesByCleanerIDRange(ctx context.Context, cleanerID uint, fromDate string, toDate string) ([]AvailabilityOverride, error) {
	query := ` 
		SELECT 
			id,
			cleaner_id,
			available_date::text, 
			start_time,
			end_time,
			status,
			COALESCE(reason, ''), 
			created_at,
			updated_at
		FROM availability_overrides
		WHERE cleaner_id = $1 
			AND available_date >= $2 
			AND available_date <= $3 
		ORDER BY available_date ASC, start_time ASC 		
	`

	rows, err := r.db.QueryContext(
		ctx,
		query,
		cleanerID,
		fromDate,
		toDate,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var overrides []AvailabilityOverride

	for rows.Next() {
		var override AvailabilityOverride

		if err := rows.Scan(
			&override.ID,
			&override.CleanerID,
			&override.AvailabilityDate,
			&override.StartTime,
			&override.EndTime,
			&override.Status,
			&override.Reason,
			&override.CreatedAt,
			&override.UpdatedAt,
		); err != nil {
			return nil, err
		}

		overrides = append(overrides, override)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return overrides, nil
}

func (r *SQLRepository) DeleteOverride(ctx context.Context, overrideID uint, cleanerID uint) error {
	result, err := r.db.ExecContext(
		ctx,
		`
			DELETE FROM availability_overrides
			WHERE id = $1 
				AND cleaner_id = $2 
		`,
		overrideID,
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
