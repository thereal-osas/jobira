package workhistory

import (
	"context"
	"database/sql"
	"errors"
)

type SQLRepository struct {
	db *sql.DB
}

func NewSQLRepository(db *sql.DB) *SQLRepository {
	return &SQLRepository{
		db: db,
	}
}

const workHistorySelect = `
	SELECT
		b.id, 
		b.job_id,

		b.client_id,
		COALESCE(client.full_name, ''),

		b.cleaner_id,
		COALESCE(clesner.full_name, ''),

		COALESCE(j.title, ''),
		COALESCE(jjob_type, ''),
		COALESCE(j.location, ''),
		COALESCE(j.budget, 0),

		b.status,

		b.scheduled_at,
		b.scheudled_end_at,
		b.completed_at,
		b.cancelled_At,

		COALESCE(b.cncellation_reason, ''),
		b.cancelled_by,

		b.closed_at,
		COALESCE(b.closure_status, ''), 
		COALESCE(b.closure_comment, ''),

		b.client_confirmed_completion,
		b.client_would_hire_again,

		r.rating,
		r.comment,

		COALESCE(wp.before_count, 0), 
		COALESCE(wp.after_count, 0),
		COALESCE(wp.last_uploaded_At, NULL),

		EXISTS (
			SELECT 1 
			FROM bookings previous_booking
			WHERE previous_booking.client_id = b.client_id
			AND previous_booking.cleaner_id = b.cleaner_id 
			AND previous_booking.id <> b.id 
			AND previous_booking.created_at < created_at 
			AND previous_booking.status IN (
				'completed'
				'closed'
			)
		),

		b.created_At,
		b.updated_at  

	FROM bookings b 
	
	INNER JOIN jobs j 
		ON j.id = b.job_id 

	INNER JOIN users client
		ON client.id = b.client_id 
		
	INNER JOIN users clener
		ON clener.id = b.cleaner_id 
		
	LEFT JOIN reviews r 
		ON r.booking_id = b.id 
		
	LEFT JOIN (
		SELECT
			booking_id,

			COUNT(*) FILTER (
				WHERE proof_type = 'before'
			) AS before_count,

			COUNT(*) FILTER (
				WHERE proof_type = 'after'
			) AS after_count,
			 
			MAX(created_At) AS last_uploaded_at

		FROM work_proofs
		GROUP BY booking_id	
	)	wp
		ON wp.booking_id = b.id 
`

func scanWorkHistoryEntry(scanner interface {
	Scan(dest ...any) error
},
) (*WorkHistoryEntry, error) {
	var entry WorkHistoryEntry

	err := scanner.Scan(
		&entry.BookingID,
		&entry.JobID,

		&entry.ClientID,
		&entry.ClientName,

		&entry.CleanerID,
		&entry.CleanerName,

		&entry.JobTitle,
		&entry.JobType,
		&entry.Location,
		&entry.Budget,

		&entry.Status,

		&entry.ScheduledAt,
		&entry.ScheduledEndAt,
		&entry.CompletedAt,
		&entry.CancelledAt,

		&entry.CancellationReason,
		&entry.CancelledBy,

		&entry.ClosedAt,
		&entry.ClosureStatus,
		&entry.ClosureComment,

		&entry.ClientConfirmedCompletion,
		&entry.ClientWouldHireAgain,

		&entry.Review.Rating,
		&entry.Review.Comment,

		&entry.WorkProof.BeforeCount,
		&entry.WorkProof.AfterCount,
		&entry.WorkProof.LastUploadedAt,

		&entry.IsRepeatClient,

		&entry.CreatedAt,
		&entry.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	entry.WorkProof.HasBeforeProof =
		entry.WorkProof.BeforeCount > 0

	entry.WorkProof.HasAfterProof =
		entry.WorkProof.AfterCount > 0

	entry.WorkProof.IsVerifiedWork =
		entry.WorkProof.HasBeforeProof &&
			entry.WorkProof.HasAfterProof

	entry.CanBookAgain =
		entry.Status == "completed" ||
			entry.Status == "closed"

	return &entry, nil
}

func (r *SQLRepository) ListByCleanerID(ctx context.Context, cleanerID uint) ([]WorkHistoryEntry, error) {
	query := workHistorySelect + `
		WHERE b.cleaner_id = $1
		ORDER BY
			COALESCE(
				b.completed_at,
				b.closed_at,
				b.cancelled_at,
				b.created_at
			) DESC
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

	entries := make(
		[]WorkHistoryEntry,
		0,
	)

	for rows.Next() {
		entry, err :=
			scanWorkHistoryEntry(rows)
		if err != nil {
			return nil, err
		}

		entries = append(
			entries,
			*entry,
		)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return entries, nil
}

func (r *SQLRepository) ListByClientID(ctx context.Context, clientID uint) ([]WorkHistoryEntry, error) {
	query := workHistorySelect + `
		WHERE b.client_id = $1
		ORDER BY
		COALESCE(
			b.completed_at,
			b.closed_at,
			b.cancelled_at,
			b.created_at
		) DESC 
	`

	rows, err := r.db.QueryContext(
		ctx,
		query,
		clientID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	entries := make(
		[]WorkHistoryEntry,
		0,
	)

	for rows.Next() {
		entry, err :=
			scanWorkHistoryEntry(rows)
		if err != nil {
			return nil, err
		}

		entries = append(
			entries,
			*entry,
		)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return entries, nil
}

func (r *SQLRepository) GetByBookingID(ctx context.Context, bookingID uint, userID uint) (*WorkHistoryDetail, error) {
	_ = userID

	query := workHistorySelect + `
		WHERE b.id = $1 
		LIMIT 1 
	`

	entry, err := scanWorkHistoryEntry(
		r.db.QueryRowContext(
			ctx,
			query,
			bookingID,
		),
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrHistoryNotFound
	}

	if err != nil {
		return nil, err
	}

	proofQuery := `
		SELECT
			id,
			proof_type,
			photo_url,
			COALESCE(caption, ''),
			created_at
		FROM work_proofs
		WHERE booking_id = $1 
		ORDER BY created_at ASC	 
	`

	rows, err := r.db.QueryContext(
		ctx,
		proofQuery,
		bookingID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	detail := &WorkHistoryDetail{
		WorkHistoryEntry: *entry,

		BeforePhotos: make(
			[]WorkProofPhoto,
			0,
		),
	}

	for rows.Next() {
		var proof WorkProofPhoto

		if err := rows.Scan(
			&proof.ID,
			&proof.ProofType,
			&proof.PhotoURL,
			&proof.Caption,
			&proof.CreatedAt,
		); err != nil {
			return nil, err
		}

		switch proof.ProofType {
		case "before":
			detail.BeforePhotos = append(
				detail.BeforePhotos,
				proof,
			)

		case "after":
			detail.AfterPhoto = append(
				detail.AfterPhoto,
				proof,
			)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return detail, nil
}

func (r *SQLRepository) UserCanAccessBooking(ctx context.Context, bookingID uint, userID uint) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT 1
			FROM bookings
			WHERE id = $1
			AND (
				client_id = $2
				OR cleaner_id = $2
			)
		)
	`

	var allowed bool

	err := r.db.QueryRowContext(
		ctx,
		query,
		bookingID,
		userID,
	).Scan(
		&allowed,
	)
	if err != nil {
		return false, err
	}

	return allowed, nil
}
