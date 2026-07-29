package cleanerdashboard

import (
	"context"
	"database/sql"
)

type SQLRepository struct {
	db *sql.DB
}

func NewSQLRepository(db *sql.DB) *SQLRepository {
	return &SQLRepository{db: db}
}

func (r *SQLRepository) CountFavourites(ctx context.Context, cleanerID uint) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM favourite_cleaners
		WHERE cleaner_id = $1
	`

	var count int

	err := r.db.QueryRowContext(ctx, query, cleanerID).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *SQLRepository) CountPreferredClients(ctx context.Context, cleanerID uint) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM preferred_cleaners
		WHERE cleaner_id = $1
	`

	var count int

	err := r.db.QueryRowContext(ctx, query, cleanerID).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *SQLRepository) ListUpcomingBookings(ctx context.Context, cleanerID uint, limit int) ([]UpcomingBooking, error) {
	query := `
		SELECT 
			id,
			job_id,
			client_id,
			status, 
			scheduled_at
		FROM bookings
		WHERE cleaner_id = $1
		AND status IN ('pending', 'confirmed', 'in_progress')
		AND scheduled_at >= NOW()
		ORDER BY scheduled_at ASC
		LIMIT $2	
	`

	rows, err := r.db.QueryContext(ctx, query, cleanerID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bookings []UpcomingBooking

	for rows.Next() {
		var booking UpcomingBooking

		if err := rows.Scan(
			&booking.ID,
			&booking.JobID,
			&booking.ClientID,
			&booking.Status,
			&booking.ScheduledAt,
		); err != nil {
			return nil, err
		}

		bookings = append(bookings, booking)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return bookings, nil
}
