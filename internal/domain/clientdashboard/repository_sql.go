package clientdashboard

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

func (r *SQLRepository) ListActiveJobs(ctx context.Context, clientID uint, limit int) ([]ActiveJob, error) {
	query := `
		SELECT 
			j.id,
			j.title,
			j.location,
			j.job_type,
			j.budget,
			j.status,
			COUNT(a.id) AS application_count,
			j.created_at 
		FROM jobs j 
		LEFT JOIN applications a ON a.job_id = j.id 
		WHERE j.client_id = $1
		AND j.status = 'open'
		GROUP BY 
			j.id,
			j.title, 
			j.location, 
			j.job_type, 
			j.budget, 
			j.status, 
			j.created_at 
		ORDER BY j.created_at DESC 
		LIMIT $2 	
	`

	rows, err := r.db.QueryContext(ctx, query, clientID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []ActiveJob

	for rows.Next() {
		var job ActiveJob

		if err := rows.Scan(
			&job.ID,
			&job.Title,
			&job.Location,
			&job.JobType,
			&job.Budget,
			&job.Status,
			&job.ApplicationCount,
			&job.CreatedAt,
		); err != nil {
			return nil, err
		}

		jobs = append(jobs, job)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return jobs, nil
}

func (r *SQLRepository) ListUpcomingBookings(ctx context.Context, clientID uint, limit int) ([]ClientBooking, error) {
	query := `
		SELECT 
			id, 
			job_id,
			cleaner_id, 
			status, 
			scheduled_at 
		FROM bookings 
		WHERE client_id = $1 
		AND status IN ('pending', 'confirmed', 'in_progress')
		AND scheduled_at >= NOW()
		ORDER BY scheduled_at ASC
		LIMIT $2	
	`

	rows, err := r.db.QueryContext(ctx, query, clientID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bookings []ClientBooking

	for rows.Next() {
		var booking ClientBooking

		if err := rows.Scan(
			&booking.ID,
			&booking.JobID,
			&booking.CleanerID,
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

func (r *SQLRepository) CountApplicationsReceived(ctx context.Context, clientID uint) (int, error) {
	query := `
		SELECT COUNT(a.id)
		FROM applications a 
		INNER JOIN jobs j ON j.id = a.job_id 
		WHERE j.client_id = $1
	`

	var count int

	err := r.db.QueryRowContext(ctx, query, clientID).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *SQLRepository) CountCompletedBookings(ctx context.Context, clientID uint) (int, error) {
	query := `
		SELECT COUNT (*)
		FROM bookings 
		WHERE client_id = $1
		AND status IN ('completed', 'closed')
	`

	var count int

	err := r.db.QueryRowContext(ctx, query, clientID).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *SQLRepository) CountFavouriteCleaners(ctx context.Context, clientID uint) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM favorite_cleaners
		WHERE client_id = $1
	`

	var count int

	err := r.db.QueryRowContext(ctx, query, clientID).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *SQLRepository) CountPreferredCleaners(ctx context.Context, clientID uint) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM preferred_cleaners
		WHERE client_id = $1
	`

	var count int

	err := r.db.QueryRowContext(ctx, query, clientID).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *SQLRepository) CountRepeatBookings(ctx context.Context, clientID uint) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM repeat_booking_requests
		WHERE client_id = $1
	`

	var count int

	err := r.db.QueryRowContext(ctx, query, clientID).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}
