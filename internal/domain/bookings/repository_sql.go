package bookings

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

func (r *SQLRepository) Create(ctx context.Context, booking *Booking) error {
	query := `
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
		VALUES ($1, $2, $3, $4, $5, $6, $7, $7)
		RETURNING id, created_at, updated_at
	`

	now := time.Now()

	err := r.db.QueryRowContext(
		ctx,
		query,
		booking.JobID,
		booking.ApplicationID,
		booking.ClientID,
		booking.CleanerID,
		booking.Status,
		booking.ScheduledAt,
		now,
	).Scan(
		&booking.ID,
		&booking.CreatedAt,
		&booking.UpdatedAt,
	)

	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "duplicate key") {
			return ErrBookingExists
		}
		return err
	}

	return nil
}

func (r *SQLRepository) GetByID(ctx context.Context, id uint) (*Booking, error) {
	query := `
		SELECT 
			id,
			job_id,
			application_id, 
			client_id,
			cleaner_id,
			status,
			scheduled_at,
			completed_at,
			cancelled_at,
			COALESCE(cancellation_reason, ''),

			closed_at,
			COALESCE(closure_status, ''), 
			COALESCE(closure_comment, ''), 
			client_confirmed_completion,
			client_would_hire_again,
			client_rating,


			created_at,
			updated_at
		FROM bookings
		WHERE id = $1 
		LIMIT 1 	
	`

	var booking Booking
	var applicationID sql.NullInt64
	var ScheduledAt sql.NullTime
	var CompletedAt sql.NullTime
	var CancelledAt sql.NullTime
	var closedAt sql.NullTime
	var wouldHireAgain sql.NullBool
	var clientRating sql.NullInt64

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&booking.ID,
		&booking.JobID,
		&applicationID,
		&booking.ClientID,
		&booking.CleanerID,
		&booking.Status,
		&ScheduledAt,
		&CompletedAt,
		&CancelledAt,
		&booking.CancellationReason,

		&closedAt,
		&booking.ClosureStatus,
		&booking.ClosureComment,
		&booking.ClientConfirmedCompletion,
		&wouldHireAgain,
		&clientRating,

		&booking.CreatedAt,
		&booking.UpdatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrBookingNotFound
	}

	if err != nil {
		return nil, err
	}
	if closedAt.Valid {
		booking.ClosedAt = &closedAt.Time
	}

	if wouldHireAgain.Valid {
		value := wouldHireAgain.Bool
		booking.ClientWouldHireAgain = &value
	}

	if clientRating.Valid {
		value := int(clientRating.Int64)
		booking.ClientRating = &value
	}

	if applicationID.Valid {
		value := uint(applicationID.Int64)
		booking.ApplicationID = &value
	}

	if ScheduledAt.Valid {
		booking.ScheduledAt = &ScheduledAt.Time
	}

	if CompletedAt.Valid {
		booking.CompletedAt = &CompletedAt.Time
	}

	if CancelledAt.Valid {
		booking.CancelledAt = &CancelledAt.Time
	}

	return &booking, nil
}

func (r *SQLRepository) ListByUserID(ctx context.Context, userID uint) ([]Booking, error) {
	query := `
		SELECT
			id,
			job_id,
			application_id,
			client_id,
			cleaner_id,
			status,
			scheduled_at,
			completed_at,
			cancelled_at,
			COALESCE(cancellation_reason, ''),
			created_at,
			updated_at
		FROM bookings
		WHERE client_id = $1 
		OR cleaner_id = $1 
		ORDER BY created_at DESC	
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bookings []Booking

	for rows.Next() {
		var booking Booking
		var applicationID sql.NullInt64
		var ScheduledAt sql.NullTime
		var CompletedAt sql.NullTime
		var CancelledAt sql.NullTime

		err := rows.Scan(
			&booking.ID,
			&booking.JobID,
			&applicationID,
			&booking.ClientID,
			&booking.CleanerID,
			&booking.Status,
			&ScheduledAt,
			&CompletedAt,
			&CancelledAt,
			&booking.CancellationReason,
			&booking.CreatedAt,
			&booking.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if applicationID.Valid {
			value := uint(applicationID.Int64)
			booking.ApplicationID = &value
		}

		if ScheduledAt.Valid {
			booking.ScheduledAt = &ScheduledAt.Time
		}

		if CompletedAt.Valid {
			booking.CompletedAt = &CompletedAt.Time
		}

		if CancelledAt.Valid {
			booking.CancelledAt = &CancelledAt.Time
		}

		bookings = append(bookings, booking)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return bookings, nil
}

func (r *SQLRepository) UpdateStatus(ctx context.Context, bookingID uint, status string) error {
	query := `
		UPDATE bookings
		SET status = $1, 
			updated_at = $2 
		WHERE id = $3 	
	`

	result, err := r.db.ExecContext(ctx, query, status, time.Now(), bookingID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrBookingNotFound
	}

	return nil
}

func (r *SQLRepository) Complete(ctx context.Context, bookingID uint, status string) error {
	query := `
		UPDATE bookings
		SET status = $1,
			completed_at = $2,
			updated_at = $2
		WHERE id = $3	
	`

	now := time.Now()

	result, err := r.db.ExecContext(ctx, query, status, now, bookingID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrBookingNotFound
	}

	return nil
}

func (r *SQLRepository) Cancel(ctx context.Context, bookingID uint, reason string) error {
	query := `
		UPDATE bookings
		SET status = 'cancelled',
			cancelled_at = $1,
			cancellation_reason = $2,
			updated_at = $1
		WHERE id = $3	
	`

	now := time.Now()

	result, err := r.db.ExecContext(ctx, query, now, reason, bookingID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrBookingNotFound
	}

	return nil
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

func (r *SQLRepository) Close(ctx context.Context, bookingID uint, rating int, wouldHireAgain bool, comment string) error {
	query := `
		UPDATE bookings
		SET
			status = 'closed',
			closure_status = 'client_confirmed',
			client_confirmed_completion = true,
			client_rating = $1,
			client_would_hire_again = $2,
			closure_comment = $3,
			closed_at = $4,
			updated_at = $4
		WHERE id = $5
	`

	now := time.Now()

	result, err := r.db.ExecContext(
		ctx,
		query,
		rating,
		wouldHireAgain,
		comment,
		now,
		bookingID,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrBookingNotFound
	}

	return nil
}

func (r *SQLRepository) MarkJobCompleted(ctx context.Context, jobID uint) error {
	query := `
		UPDATE jobs
		SET 	
			status = 'completed',
			updated_at = $1
		WHERE id = $2	
	`

	_, err := r.db.ExecContext(ctx, query, time.Now(), jobID)
	return err
}

func (r *SQLRepository) MarkApplicationAccepted(ctx context.Context, applicationID uint) error {
	query := `
		UPDATE applications
		SET status = 'accepted',
			updated_at = $1
		WHERE id = $2	
	`

	_, err := r.db.ExecContext(ctx, query, time.Now(), applicationID)
	return err
}

func (r *SQLRepository) CloseWithTransaction(ctx context.Context, input CloseBookingTransaction) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	commited := false

	defer func() {
		if !commited {
			_ = tx.Rollback()
		}
	}()

	now := time.Now()

	closedBookingQuery := `
		UPDATE bookings
		SET 
			status = 'closed',
			closure_status = 'client_confirmed',
			client_confirmed_completion = true,
			client_rating = $1,
			client_would_hire_again = $2,
			closure_comment = $3,
			closed_at = $4,
			updated_at $4
		WHERE id = $5
		AND status = 'completed'	
	`

	result, err := tx.ExecContext(
		ctx,
		closedBookingQuery,
		input.Rating,
		input.WouldHireAgain,
		input.Comment,
		now,
		input.BookingID,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrInvalidInput
	}

	createReviewQuery := `
		INSERT INTO reviews (
			booking_id,
			job_id,
			cleaner_id, 
			client_id,
			rating, 
			comment, 
			created_at,
			updated_at
		)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $7)
	`

	_, err = tx.ExecContext(
		ctx,
		createReviewQuery,
		input.BookingID,
		input.JobID,
		input.CleanerID,
		input.ClientID,
		input.Rating,
		input.Comment,
		now,
	)
	if input.AddToFavourites {
		addFavouriteQuery := `
			INSERT INTO favorite_cleaners (
				client_id,
				cleaner_id,
				created_at
			)
				VALUES ($1, $2, $3)
				ON CONFLICT DO NOTHING
		`

		_, err = tx.ExecContext(
			ctx,
			addFavouriteQuery,
			input.ClientID,
			input.CleanerID,
			now,
		)
		if err != nil {
			return err
		}
	}

	if input.SetAsPreferred {
		addPreferredQuery := `
			INSERT INTO preferred_cleaners (
				client_id,
				cleaner_id,
				created_at
			)
				VALUES ($1, $2, $3)
				ON CONFLICT DO NOTHING
		`

		_, err = tx.ExecContext(
			ctx,
			addPreferredQuery,
			input.ClientID,
			input.CleanerID,
			now,
		)
		if err != nil {
			return err
		}
	}

	createTimelineQuery := `
		INSERT INTO booking_status_history(
			booking_id,
			changed_by, 
			from_status, 
			to_status, 
			notes, 
			created_at
		)
		VALUES ($1, $2, $3, 'closed', $4, $5)	
	`

	_, err = tx.ExecContext(
		ctx,
		createTimelineQuery,
		input.BookingID,
		input.ChangedBy,
		input.PreviousStatus,
		"Client confirmed completion and closed the booking",
		now,
	)
	if err != nil {
		return err
	}

	createNotificationQuery := `
		INSERT INTO notifications (
			user_id, 
			title,
			message,
			type,
			is_read,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, 'booking' $4, $4)
	`

	_, err = tx.ExecContext(
		ctx,
		createNotificationQuery,
		input.CleanerID,
		input.NotificationTitle,
		input.NotificationMessage,
		now,
	)
	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	commited = true

	return nil
}
