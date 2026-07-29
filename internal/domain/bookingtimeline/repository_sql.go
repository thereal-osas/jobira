package bookingtimeline

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
func (r *SQLRepository) CreateHistory(
	ctx context.Context,
	bookingID uint,
	changedBy uint,
	fromStatus string,
	toStatus string,
	note string,
) error {
	query := `
		INSERT INTO booking_status_history (
			booking_id,
			changed_by,
			from_status,
			to_status,
			note,
			created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	var changedByValue any

	if changedBy != 0 {
		changedByValue = changedBy
	}

	_, err := r.db.ExecContext(
		ctx,
		query,
		bookingID,
		changedByValue,
		fromStatus,
		toStatus,
		note,
		time.Now(),
	)

	return err
}

func (r *SQLRepository) ListByBookingID(ctx context.Context, bookingID uint) ([]StatusHistory, error) {
	query := `
		SELECT
			id, 
			booking_id,
			changed_by,
			COALESCE(from_status, ''),
			to_status,
			COALESCE(note, ''),
			created_at
		FROM booking_status_history
		WHERE booking_id = $1
		ORDER BY created_at ASC, id ASC	
	`

	rows, err := r.db.QueryContext(ctx, query, bookingID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []StatusHistory

	for rows.Next() {
		var item StatusHistory
		var changedBy sql.NullInt64

		err := rows.Scan(
			&item.ID,
			&item.BookingID,
			&changedBy,
			&item.FromStatus,
			&item.ToStatus,
			&item.Note,
			&item.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		if changedBy.Valid {
			value := uint(changedBy.Int64)
			item.ChangedBy = &value
		}

		history = append(history, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return history, nil
}

func (r *SQLRepository) GetBookingAccess(ctx context.Context, bookingID uint) (uint, uint, string, error) {
	query := `
		SELECT 
			client_id,
			cleaner_id,
			status
		FROM bookings
		WHERE id = $1
		LIMIT 1	
	`

	var clientID uint
	var cleanerID uint
	var currentStatus string

	err := r.db.QueryRowContext(ctx, query, bookingID).Scan(
		&clientID,
		&cleanerID,
		&currentStatus,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return 0, 0, "", ErrBookingNotFound
	}

	if err != nil {
		return 0, 0, "", err
	}

	return clientID, cleanerID, currentStatus, nil
}
