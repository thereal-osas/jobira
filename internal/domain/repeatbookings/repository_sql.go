package repeatbookings

import (
	"context"
	"database/sql"
	"time"
)

type SQLRepository struct {
	db *sql.DB
}

func NewSQLRepository(db *sql.DB) *SQLRepository {
	return &SQLRepository{db: db}
}

func (r *SQLRepository) CreateJob(ctx context.Context, clientID uint, req CreateRepeatBookingRequest) (uint, error) {
	query := `
		INSERT INTO jobs (
			client_id,
			title,
			description,
			location,
			job_type,
			listing_type,
			budget, 
			status,
			created_at, 
			updated_at 
		)
	VALUES ($1, $2, $3, $4, $5, $6, $7, 'open', $8, $9)
	RETURNING id 
	`

	now := time.Now()

	var jobID uint

	err := r.db.QueryRowContext(
		ctx,
		query,
		clientID,
		req.Title,
		req.Description,
		req.Location,
		req.JobType,
		req.ListingType,
		req.Budget,
		now,
		now,
	).Scan(&jobID)

	return jobID, err
}

func (r *SQLRepository) CreateInvitation(ctx context.Context, jobID uint, clientID uint, cleanerID uint, message string) error {
	query := `
		INSERT INTO job_invitations (
			job_id,
			client_id,
			cleaner_id,
			status,
			message,
			created_at, 
			updated_at 
		)
	VALUES ($1, $2, $3, 'sent', $4, $5, $6) 
	`

	now := time.Now()

	_, err := r.db.ExecContext(
		ctx,
		query,
		jobID,
		clientID,
		cleanerID,
		message,
		now,
		now,
	)

	return err
}


func (r *SQLRepository) CreateRepeatBooking(ctx context.Context, booking *RepeatBooking) error {
	query := `
		INSERT INTO repeat_bookings (
			client_id,
			cleaner_id,
			job_id,
			message,
			status,
			created_at, 
			updated_at 
		)
	VALUES ($1, $2, $3, $4, $5, $6, $7)
	RETURNING id, created_at, updated_at 
	`

	now := time.Now()

	return r.db.QueryRowContext(
		ctx,
		query,
		booking.ClientID,
		booking.CleanerID,
		booking.JobID,
		booking.Message,
		booking.Status,
		now,
		now,
	).Scan(
		&booking.ID,
		&booking.CreatedAt,
		&booking.UpdatedAt,
	)

}


func (r *SQLRepository) ListByClientID(ctx context.Context, clientID uint) ([]RepeatBooking, error) {
	query := `
			SELECT 
			id,
			client_id,
			cleaner_id,
			job_id,
			message, 
			status,
			created_at, 
			updated_at 
		FROM repeat_bookings 
		WHERE client_id = $1 
		ORDER BY created_at DESC  
	`

	rows, err := r.db.QueryContext(ctx, query, clientID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var bookings []RepeatBooking

	for rows.Next() {
		var booking RepeatBooking

		err := rows.Scan(
			&booking.ID,
			&booking.ClientID,
			&booking.CleanerID,
			&booking.JobID,
			&booking.Message,
			&booking.Status,
			&booking.CreatedAt,
			&booking.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		bookings = append(bookings, booking)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return bookings, nil

}

func (r *SQLRepository) GetOriginBooking(ctx context.Context, bookingID uint) (*BookingSnapshot, error) {
	query := `
		SELECT 
			id, 
			job_id,
			application_id,
			client_id,
			cleaner_id,
			status
		FROM bookings
		WHERE id = $1
		LIMIT 1 	
	`

	var booking BookingSnapshot
	var applicationID sql.NullInt64

	err := r.db.QueryRowContext(ctx, query, bookingID).Scan(
		&booking.ID,
		&booking.JobID,
		&applicationID,
		&booking.ClientID,
		&booking.CleanerID,
		&booking.Status,
	)
	if err != nil {
		return nil, err
	}

	if applicationID.Valid {
		value := uint(applicationID.Int64)
		booking.ApplicationID = &value
	}

	return &booking, nil
}
func (r *SQLRepository) CreateRepeatBookings(
	ctx context.Context,
	booking *BookingSnapshot,
	scheduledAt time.Time,
	scheduledEndAt time.Time,
) (uint, error) {
	query := `
		INSERT INTO bookings (
			job_id,
			application_id,
			client_id,
			cleaner_id,
			status,
			scheduled_at,
			scheduled_end_at,
			created_at,
			updated_at
		)
		VALUES (
			$1,
			$2,
			$3,
			$4,
			'pending',
			$5,
			$6,
			$7,
			$7
		)
		RETURNING id
	`

	now := time.Now()

	var newBookingID uint

	err := r.db.QueryRowContext(
		ctx,
		query,
		booking.JobID,
		booking.ApplicationID,
		booking.ClientID,
		booking.CleanerID,
		scheduledAt,
		scheduledEndAt,
		now,
	).Scan(
		&newBookingID,
	)

	return newBookingID, err
}

func (r *SQLRepository) CreateBookAgainRequest(ctx context.Context, request *RepeatBookingRequest) error {
	query := `
		INSERT INTO repeat_booking_requests (
			original_booking_id,
			new_booking_id,
			client_id,
			cleaner_id,
			job_id,
			scheduled_at,
			scheduled_end_at,
			status, 
			message, 
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $10)
		RETURNING id, created_at, updated_at	
	`

	now := time.Now()

	err := r.db.QueryRowContext(
		ctx,
		query,
		request.OriginalBookingID,
		request.NewBookingID,
		request.ClientID,
		request.CleanerID,
		request.JobID,
		request.ScheduledAt,
		request.ScheduledEndAt,
		request.Status,
		request.Message,
		now,
	).Scan(
		&request.ID,
		&request.CreatedAt,
		&request.UpdatedAt,
	)

	return err
}

func (r *SQLRepository) CreateBookAgainTransaction(
	ctx context.Context,
	input BookAgainTransaction,
) (*RepeatBookingRequest, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	committed := false

	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	now := time.Now()

	createBookingQuery := `
		INSERT INTO bookings (
			job_id,
			application_id,
			client_id,
			cleaner_id,
			status,
			scheduled_at,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, 'pending', $5, $6, $6)
		RETURNING id
	`

	var newBookingID uint

	err = tx.QueryRowContext(
		ctx,
		createBookingQuery,
		input.JobID,
		input.ApplicationID,
		input.ClientID,
		input.CleanerID,
		input.ScheduledAt,
		now,
	).Scan(&newBookingID)
	if err != nil {
		return nil, err
	}

	createRequestQuery := `
		INSERT INTO repeat_booking_requests (
			original_booking_id,
			new_booking_id,
			client_id,
			cleaner_id,
			job_id,
			scheduled_at,
			status,
			message,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $9)
		RETURNING id, created_at, updated_at
	`

	request := &RepeatBookingRequest{
		OriginalBookingID: input.OriginalBookingID,
		NewBookingID:      &newBookingID,
		ClientID:          input.ClientID,
		CleanerID:         input.CleanerID,
		JobID:             input.JobID,
		ScheduledAt:       input.ScheduledAt,
		Status:            input.Status,
		Message:           input.Message,
	}

	err = tx.QueryRowContext(
		ctx,
		createRequestQuery,
		request.OriginalBookingID,
		request.NewBookingID,
		request.ClientID,
		request.CleanerID,
		request.JobID,
		request.ScheduledAt,
		request.Status,
		request.Message,
		now,
	).Scan(
		&request.ID,
		&request.CreatedAt,
		&request.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	committed = true

	return request, nil
}
