package bookings

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func closeBookingQuery() string {
	return regexp.QuoteMeta(`
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
		AND status = 'completed'
	`)
}

func createReviewQuery() string {
	return regexp.QuoteMeta(`
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
	`)
}

func addFavouriteQuery() string {
	return regexp.QuoteMeta(`
		INSERT INTO favorite_cleaners (
			client_id,
			cleaner_id,
			created_at
		)
			VALUES ($1, $2, $3)
			ON CONFLICT DO NOTHING
	`)
}

func addPreferredQuery() string {
	return regexp.QuoteMeta(`
		INSERT INTO preferred_cleaners (
			client_id,
			cleaner_id,
			created_at
		)
			VALUES ($1, $2, $3)
			ON CONFLICT DO NOTHING
	`)
}

func createTimelineQuery() string {
	return regexp.QuoteMeta(`
		INSERT INTO booking_status_history(
			booking_id,
			changed_by, 
			from_status, 
			to_status, 
			notes, 
			created_at
		)
		VALUES ($1, $2, $3, 'closed', $4, $5)
	`)
}

func createNotificationQuery() string {
	return regexp.QuoteMeta(`
		INSERT INTO notifications (
			user_id, 
			title,
			message,
			type,
			is_read,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, 'booking', false, $4, $4)
	`)
}

func bookingColumnNames() []string {
	return []string{
		"id",
		"job_id",
		"application_id",
		"client_id",
		"cleaner_id",
		"status",
		"scheduled_at",
		"completed_at",
		"cancelled_at",
		"cancellation_reason",
		"closed_at",
		"closure_status",
		"closure_comment",
		"client_confirmed_completion",
		"client_would_hire_again",
		"client_rating",
		"created_at",
		"updated_at",
	}
}

func newRepositoryMock(t *testing.T) (*SQLRepository, sqlmock.Sqlmock, *sql.DB) {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}

	t.Cleanup(func() {
		_ = db.Close()
	})

	return NewSQLRepository(db), mock, db
}

func TestNewSQLRepository(t *testing.T) {
	db := &sql.DB{}

	repo := NewSQLRepository(db)

	if repo == nil {
		t.Fatal("expected repository")
	}

	if repo.db != db {
		t.Fatal("expected database to be assigned")
	}
}
func TestSQLRepository_Create_Success(t *testing.T) {
	repo, mock, _ := newRepositoryMock(t)

	ctx := context.Background()

	createdAt := time.Date(
		2026,
		time.July,
		29,
		18,
		0,
		0,
		0,
		time.UTC,
	)

	updatedAt := createdAt.Add(time.Minute)

	booking := &Booking{
		JobID:     1,
		ClientID:  2,
		CleanerID: 3,
		Status:    "pending",
	}

	rows := sqlmock.NewRows([]string{
		"id",
		"created_at",
		"updated_at",
	}).AddRow(
		10,
		createdAt,
		updatedAt,
	)

	mock.
		ExpectQuery(regexp.QuoteMeta(`
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
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $8)
			RETURNING id, created_at, updated_at
		`)).
		WithArgs(
			booking.JobID,
			booking.ApplicationID,
			booking.ClientID,
			booking.CleanerID,
			booking.Status,
			booking.ScheduledAt,
			booking.ScheduledEndAt,
			sqlmock.AnyArg(),
		).
		WillReturnRows(rows)

	err := repo.Create(ctx, booking)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if booking.ID != 10 {
		t.Fatalf("expected booking id 10, got %d", booking.ID)
	}

	if !booking.CreatedAt.Equal(createdAt) {
		t.Fatalf(
			"expected created_at %v, got %v",
			createdAt,
			booking.CreatedAt,
		)
	}

	if !booking.UpdatedAt.Equal(updatedAt) {
		t.Fatalf(
			"expected updated_at %v, got %v",
			updatedAt,
			booking.UpdatedAt,
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLRepository_Create_Duplicate(t *testing.T) {
	repo, mock, _ := newRepositoryMock(t)

	ctx := context.Background()

	booking := &Booking{
		JobID:     1,
		ClientID:  2,
		CleanerID: 3,
		Status:    "pending",
	}

	mock.
		ExpectQuery(regexp.QuoteMeta(`
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
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $8)
		RETURNING id, created_at, updated_at
	`)).
		WillReturnError(
			sql.ErrNoRows,
		)

	err := repo.Create(ctx, booking)

	if err == nil {
		t.Fatal("expected error")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLRepository_Create_DatabaseError(t *testing.T) {
	repo, mock, _ := newRepositoryMock(t)

	ctx := context.Background()

	booking := &Booking{
		JobID:     1,
		ClientID:  2,
		CleanerID: 3,
		Status:    "pending",
	}

	expected := errors.New("database error")

	mock.
		ExpectQuery(regexp.QuoteMeta(`
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
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $8)
		RETURNING id, created_at, updated_at
	`)).
		WillReturnError(expected)

	err := repo.Create(ctx, booking)

	if !errors.Is(err, expected) {
		t.Fatalf("expected %v got %v", expected, err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLRepository_GetByID_Success(t *testing.T) {
	repo, mock, _ := newRepositoryMock(t)

	ctx := context.Background()

	now := time.Now()

	rows := sqlmock.NewRows([]string{
		"id",
		"job_id",
		"application_id",
		"client_id",
		"cleaner_id",
		"status",
		"scheduled_at",
		"scheduled_end_at",
		"completed_at",
		"cancelled_at",
		"cancellation_reason",
		"closed_at",
		"closure_status",
		"closure_comment",
		"client_confirmed_completion",
		"client_would_hire_again",
		"client_rating",
		"created_at",
		"updated_at",
	}).AddRow(
		1,                    // id
		10,                   // job_id
		11,                   // application_id
		20,                   // client_id
		30,                   // cleaner_id
		"completed",          // status
		now,                  // scheduled_at
		now.Add(2*time.Hour), // scheduled_end_at
		now,                  // completed_at
		nil,                  // cancelled_at
		"",                   // cancellation_reason
		now,                  // closed_at
		"client_confirmed",   // closure_status
		"Great work",         // closure_comment
		true,                 // client_confirmed_completion
		true,                 // client_would_hire_again
		5,                    // client_rating
		now,                  // created_at
		now,                  // updated_at
	)
	mock.
		ExpectQuery(regexp.QuoteMeta(`
		SELECT 
			id,
			job_id,
			application_id, 
			client_id,
			cleaner_id,
			status,
			scheduled_at,
			scheduled_end_at,
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
	`)).
		WithArgs(uint(1)).
		WillReturnRows(rows)

	booking, err := repo.GetByID(ctx, 1)
	if err != nil {
		t.Fatalf("expected no error got %v", err)
	}

	if booking == nil {
		t.Fatal("expected booking")
	}

	if booking.ID != 1 {
		t.Fatal("unexpected booking id")
	}

	if booking.ClientRating == nil || *booking.ClientRating != 5 {
		t.Fatal("expected rating")
	}

	if booking.ClientWouldHireAgain == nil || !*booking.ClientWouldHireAgain {
		t.Fatal("expected would hire again")
	}

	if booking.ApplicationID == nil || *booking.ApplicationID != 11 {
		t.Fatal("expected application id")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLRepository_GetByID_NotFound(t *testing.T) {
	repo, mock, _ := newRepositoryMock(t)

	ctx := context.Background()

	mock.
		ExpectQuery(regexp.QuoteMeta(`
		SELECT 
			id,
			job_id,
			application_id, 
			client_id,
			cleaner_id,
			status,
			scheduled_at,
			scheduled_end_at,
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
	`)).
		WithArgs(uint(99)).
		WillReturnError(sql.ErrNoRows)

	booking, err := repo.GetByID(ctx, 99)

	if booking != nil {
		t.Fatal("expected nil booking")
	}

	if !errors.Is(err, ErrBookingNotFound) {
		t.Fatalf("expected ErrBookingNotFound got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLRepository_GetByID_DatabaseError(t *testing.T) {
	repo, mock, _ := newRepositoryMock(t)

	ctx := context.Background()

	expected := errors.New("database error")

	mock.
		ExpectQuery(regexp.QuoteMeta(`
		SELECT 
			id,
			job_id,
			application_id, 
			client_id,
			cleaner_id,
			status,
			scheduled_at,
			scheduled_end_at,
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
	`)).
		WithArgs(uint(1)).
		WillReturnError(expected)

	booking, err := repo.GetByID(ctx, 1)

	if booking != nil {
		t.Fatal("expected nil booking")
	}

	if !errors.Is(err, expected) {
		t.Fatalf("expected %v got %v", expected, err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func listBookingColumnNames() []string {
	return []string{
		"id",
		"job_id",
		"application_id",
		"client_id",
		"cleaner_id",
		"status",
		"scheduled_at",
		"scheduled_end_at",
		"completed_at",
		"cancelled_at",
		"cancellation_reason",
		"created_at",
		"updated_at",
	}
}

func TestSQLRepository_ListByUserID_Success(t *testing.T) {
	repo, mock, _ := newRepositoryMock(t)

	ctx := context.Background()
	now := time.Now()

	rows := sqlmock.NewRows(
		listBookingColumnNames(),
	).
		AddRow(
			1,
			10,
			11,
			20,
			30,
			"confirmed",
			now,                  // scheduled_at
			now.Add(2*time.Hour), // scheduled_end_at
			nil,                  // completed_at
			nil,                  // cancelled_at
			"",
			now,
			now,
		).
		AddRow(
			2,
			12,
			nil,
			20,
			31,
			"completed",
			nil, // scheduled_at
			nil, // scheduled_end_at
			now, // completed_at
			nil, // cancelled_at
			"",
			now,
			now,
		)

	mock.
		ExpectQuery(regexp.QuoteMeta(`
			SELECT
				id,
				job_id,
				application_id,
				client_id,
				cleaner_id,
				status,
				scheduled_at,
				scheduled_end_at,
				completed_at,
				cancelled_at,
				COALESCE(cancellation_reason, ''),
				created_at,
				updated_at
			FROM bookings
			WHERE client_id = $1 
			OR cleaner_id = $1 
			ORDER BY created_at DESC
		`)).
		WithArgs(uint(20)).
		WillReturnRows(rows)

	bookings, err := repo.ListByUserID(ctx, 20)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(bookings) != 2 {
		t.Fatalf("expected 2 bookings, got %d", len(bookings))
	}

	if bookings[0].ApplicationID == nil || *bookings[0].ApplicationID != 11 {
		t.Fatal("expected first booking application id 11")
	}

	if bookings[0].ScheduledAt == nil {
		t.Fatal("expected first booking scheduled time")
	}

	if bookings[1].CompletedAt == nil {
		t.Fatal("expected second booking completed time")
	}

	if bookings[1].ApplicationID != nil {
		t.Fatal("expected second booking application id to be nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLRepository_ListByUserID_Empty(t *testing.T) {
	repo, mock, _ := newRepositoryMock(t)

	mock.
		ExpectQuery(regexp.QuoteMeta(`
			SELECT
				id,
				job_id,
				application_id,
				client_id,
				cleaner_id,
				status,
				scheduled_at,
				scheduled_end_at,
				completed_at,
				cancelled_at,
				COALESCE(cancellation_reason, ''),
				created_at,
				updated_at
			FROM bookings
			WHERE client_id = $1 
			OR cleaner_id = $1 
			ORDER BY created_at DESC
		`)).
		WithArgs(uint(99)).
		WillReturnRows(
			sqlmock.NewRows(listBookingColumnNames()),
		)

	bookings, err := repo.ListByUserID(context.Background(), 99)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(bookings) != 0 {
		t.Fatalf("expected no bookings, got %d", len(bookings))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLRepository_ListByUserID_QueryError(t *testing.T) {
	repo, mock, _ := newRepositoryMock(t)

	expected := errors.New("query failed")

	mock.
		ExpectQuery(regexp.QuoteMeta(`
			SELECT
				id,
				job_id,
				application_id,
				client_id,
				cleaner_id,
				status,
				scheduled_at,
				scheduled_end_at,
				completed_at,
				cancelled_at,
				COALESCE(cancellation_reason, ''),
				created_at,
				updated_at
			FROM bookings
			WHERE client_id = $1 
			OR cleaner_id = $1 
			ORDER BY created_at DESC
		`)).
		WithArgs(uint(20)).
		WillReturnError(expected)

	bookings, err := repo.ListByUserID(context.Background(), 20)

	if bookings != nil {
		t.Fatal("expected nil bookings")
	}

	if !errors.Is(err, expected) {
		t.Fatalf("expected %v, got %v", expected, err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLRepository_ListByUserID_ScanError(t *testing.T) {
	repo, mock, _ := newRepositoryMock(t)

	rows := sqlmock.NewRows(
		listBookingColumnNames(),
	).AddRow(
		"invalid-id",
		10,
		nil,
		20,
		30,
		"confirmed",
		nil, // scheduled_at
		nil, // scheduled_end_at
		nil, // completed_at
		nil, // cancelled_at
		"",
		time.Now(),
		time.Now(),
	)

	mock.
		ExpectQuery(regexp.QuoteMeta(`
			SELECT
				id,
				job_id,
				application_id,
				client_id,
				cleaner_id,
				status,
				scheduled_at,
				scheduled_end_at,
				completed_at,
				cancelled_at,
				COALESCE(cancellation_reason, ''),
				created_at,
				updated_at
			FROM bookings
			WHERE client_id = $1 
			OR cleaner_id = $1 
			ORDER BY created_at DESC
		`)).
		WithArgs(uint(20)).
		WillReturnRows(rows)

	bookings, err := repo.ListByUserID(context.Background(), 20)

	if bookings != nil {
		t.Fatal("expected nil bookings")
	}

	if err == nil {
		t.Fatal("expected scan error")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLRepository_ListByUserID_RowsError(t *testing.T) {
	repo, mock, _ := newRepositoryMock(t)

	expected := errors.New("rows iteration failed")

	rows := sqlmock.NewRows(
		listBookingColumnNames(),
	).
		AddRow(
			1,
			10,
			nil,
			20,
			30,
			"confirmed",
			nil, // scheduled_at
			nil, // scheduled_end_at
			nil, // completed_at
			nil, // cancelled_at
			"",
			time.Now(),
			time.Now(),
		).
		RowError(0, expected)

	mock.
		ExpectQuery(regexp.QuoteMeta(`
			SELECT
				id,
				job_id,
				application_id,
				client_id,
				cleaner_id,
				status,
				scheduled_at,
				scheduled_end_at,
				completed_at,
				cancelled_at,
				COALESCE(cancellation_reason, ''),
				created_at,
				updated_at
			FROM bookings
			WHERE client_id = $1 
			OR cleaner_id = $1 
			ORDER BY created_at DESC
		`)).
		WithArgs(uint(20)).
		WillReturnRows(rows)

	bookings, err := repo.ListByUserID(context.Background(), 20)

	if bookings != nil {
		t.Fatal("expected nil bookings")
	}

	if !errors.Is(err, expected) {
		t.Fatalf("expected %v, got %v", expected, err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLRepository_UpdateStatus_Success(t *testing.T) {
	repo, mock, _ := newRepositoryMock(t)

	mock.
		ExpectExec(regexp.QuoteMeta(`
			UPDATE bookings
			SET status = $1, 
				updated_at = $2 
			WHERE id = $3
		`)).
		WithArgs("confirmed", sqlmock.AnyArg(), uint(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.UpdateStatus(context.Background(), 1, "confirmed")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLRepository_UpdateStatus_NotFound(t *testing.T) {
	repo, mock, _ := newRepositoryMock(t)

	mock.
		ExpectExec(regexp.QuoteMeta(`
			UPDATE bookings
			SET status = $1, 
				updated_at = $2 
			WHERE id = $3
		`)).
		WithArgs("confirmed", sqlmock.AnyArg(), uint(99)).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := repo.UpdateStatus(context.Background(), 99, "confirmed")

	if !errors.Is(err, ErrBookingNotFound) {
		t.Fatalf("expected ErrBookingNotFound, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLRepository_UpdateStatus_ExecError(t *testing.T) {
	repo, mock, _ := newRepositoryMock(t)

	expected := errors.New("update failed")

	mock.
		ExpectExec(regexp.QuoteMeta(`
			UPDATE bookings
			SET status = $1, 
				updated_at = $2 
			WHERE id = $3
		`)).
		WithArgs("confirmed", sqlmock.AnyArg(), uint(1)).
		WillReturnError(expected)

	err := repo.UpdateStatus(context.Background(), 1, "confirmed")

	if !errors.Is(err, expected) {
		t.Fatalf("expected %v, got %v", expected, err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLRepository_UpdateStatus_RowsAffectedError(t *testing.T) {
	repo, mock, _ := newRepositoryMock(t)

	expected := errors.New("rows affected failed")

	mock.
		ExpectExec(regexp.QuoteMeta(`
			UPDATE bookings
			SET status = $1, 
				updated_at = $2 
			WHERE id = $3
		`)).
		WithArgs("confirmed", sqlmock.AnyArg(), uint(1)).
		WillReturnResult(sqlmock.NewErrorResult(expected))

	err := repo.UpdateStatus(context.Background(), 1, "confirmed")

	if !errors.Is(err, expected) {
		t.Fatalf("expected %v, got %v", expected, err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLRepository_Complete_Success(t *testing.T) {
	repo, mock, _ := newRepositoryMock(t)

	mock.
		ExpectExec(regexp.QuoteMeta(`
			UPDATE bookings
			SET status = $1,
				completed_at = $2,
				updated_at = $2
			WHERE id = $3
		`)).
		WithArgs("completed", sqlmock.AnyArg(), uint(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Complete(context.Background(), 1, "completed")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLRepository_Complete_NotFound(t *testing.T) {
	repo, mock, _ := newRepositoryMock(t)

	mock.
		ExpectExec(regexp.QuoteMeta(`
			UPDATE bookings
			SET status = $1,
				completed_at = $2,
				updated_at = $2
			WHERE id = $3
		`)).
		WithArgs("completed", sqlmock.AnyArg(), uint(99)).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := repo.Complete(context.Background(), 99, "completed")

	if !errors.Is(err, ErrBookingNotFound) {
		t.Fatalf("expected ErrBookingNotFound, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLRepository_Complete_ExecError(t *testing.T) {
	repo, mock, _ := newRepositoryMock(t)

	expected := errors.New("complete failed")

	mock.
		ExpectExec(regexp.QuoteMeta(`
			UPDATE bookings
			SET status = $1,
				completed_at = $2,
				updated_at = $2
			WHERE id = $3
		`)).
		WithArgs("completed", sqlmock.AnyArg(), uint(1)).
		WillReturnError(expected)

	err := repo.Complete(context.Background(), 1, "completed")

	if !errors.Is(err, expected) {
		t.Fatalf("expected %v, got %v", expected, err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLRepository_Complete_RowsAffectedError(t *testing.T) {
	repo, mock, _ := newRepositoryMock(t)

	expected := errors.New("rows affected failed")

	mock.
		ExpectExec(regexp.QuoteMeta(`
			UPDATE bookings
			SET status = $1,
				completed_at = $2,
				updated_at = $2
			WHERE id = $3
		`)).
		WithArgs("completed", sqlmock.AnyArg(), uint(1)).
		WillReturnResult(sqlmock.NewErrorResult(expected))

	err := repo.Complete(context.Background(), 1, "completed")

	if !errors.Is(err, expected) {
		t.Fatalf("expected %v, got %v", expected, err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLRepository_Cancel_Success(t *testing.T) {
	repo, mock, _ := newRepositoryMock(t)

	mock.
		ExpectExec(regexp.QuoteMeta(`
		UPDATE bookings
		SET
			status = 'cancelled',
			cancelled_at = $1,
			cancellation_reason = $2,
			cancelled_by = $3,
			updated_at = $1
		WHERE id = $4
	`)).
		WithArgs(
			sqlmock.AnyArg(),
			"client unavailable",
			uint(5),
			uint(1),
		).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Cancel(context.Background(), 1, 5, "client unavailable")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLRepository_Cancel_NotFound(t *testing.T) {
	repo, mock, _ := newRepositoryMock(t)

	mock.
		ExpectExec(regexp.QuoteMeta(`
		UPDATE bookings
		SET
			status = 'cancelled',
			cancelled_at = $1,
			cancellation_reason = $2,
			cancelled_by = $3,
			updated_at = $1
		WHERE id = $4
	`)).
		WithArgs(
			sqlmock.AnyArg(),
			"client unavailable",
			uint(5),
			uint(99),
		).
		WillReturnResult(
			sqlmock.NewResult(0, 0),
		)

	err := repo.Cancel(
		context.Background(),
		99,
		5,
		"client unavailable",
	)

	if !errors.Is(err, ErrBookingNotFound) {
		t.Fatalf("expected ErrBookingNotFound, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLRepository_Cancel_ExecError(t *testing.T) {
	repo, mock, _ := newRepositoryMock(t)

	expected := errors.New("cancel failed")

	mock.
		ExpectExec(regexp.QuoteMeta(`
		UPDATE bookings
		SET
			status = 'cancelled',
			cancelled_at = $1,
			cancellation_reason = $2,
			cancelled_by = $3,
			updated_at = $1
		WHERE id = $4
	`)).
		WithArgs(
			sqlmock.AnyArg(),
			"client unavailable",
			uint(5),
			uint(1),
		).
		WillReturnError(expected)

	err := repo.Cancel(
		context.Background(),
		1,
		5,
		"client unavailable",
	)

	if !errors.Is(err, expected) {
		t.Fatalf("expected %v, got %v", expected, err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLRepository_Cancel_RowsAffectedError(t *testing.T) {
	repo, mock, _ := newRepositoryMock(t)

	expected := errors.New("rows affected failed")

	mock.
		ExpectExec(regexp.QuoteMeta(`
		UPDATE bookings
		SET
			status = 'cancelled',
			cancelled_at = $1,
			cancellation_reason = $2,
			cancelled_by = $3,
			updated_at = $1
		WHERE id = $4
	`)).
		WithArgs(
			sqlmock.AnyArg(),
			"client unavailable",
			uint(5),
			uint(1),
		).
		WillReturnResult(
			sqlmock.NewErrorResult(expected),
		)

	err := repo.Cancel(
		context.Background(),
		1,
		5,
		"client unavailable",
	)

	if !errors.Is(err, expected) {
		t.Fatalf("expected %v, got %v", expected, err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLRepository_GetUserEmail_Success(t *testing.T) {
	repo, mock, _ := newRepositoryMock(t)

	rows := sqlmock.NewRows([]string{"email"}).
		AddRow("cleaner@test.com")

	mock.
		ExpectQuery(regexp.QuoteMeta(`
			SELECT email
			FROM users
			WHERE id = $1
		`)).
		WithArgs(uint(5)).
		WillReturnRows(rows)

	email, err := repo.GetUserEmail(context.Background(), 5)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if email != "cleaner@test.com" {
		t.Fatalf("unexpected email %s", email)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLRepository_GetUserEmail_NotFound(t *testing.T) {
	repo, mock, _ := newRepositoryMock(t)

	mock.
		ExpectQuery(regexp.QuoteMeta(`
			SELECT email
			FROM users
			WHERE id = $1
		`)).
		WithArgs(uint(999)).
		WillReturnError(sql.ErrNoRows)

	_, err := repo.GetUserEmail(context.Background(), 999)

	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected sql.ErrNoRows, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLRepository_GetUserEmail_DatabaseError(t *testing.T) {
	repo, mock, _ := newRepositoryMock(t)

	expected := errors.New("database failed")

	mock.
		ExpectQuery(regexp.QuoteMeta(`
			SELECT email
			FROM users
			WHERE id = $1
		`)).
		WithArgs(uint(1)).
		WillReturnError(expected)

	_, err := repo.GetUserEmail(context.Background(), 1)

	if !errors.Is(err, expected) {
		t.Fatalf("expected %v got %v", expected, err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
func TestSQLRepository_Close_Success(t *testing.T) {
	repo, mock, _ := newRepositoryMock(t)

	mock.
		ExpectExec(regexp.QuoteMeta(`
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
		`)).
		WithArgs(
			5,
			true,
			"Excellent work",
			sqlmock.AnyArg(),
			uint(1),
		).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Close(
		context.Background(),
		1,
		5,
		true,
		"Excellent work",
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLRepository_Close_NotFound(t *testing.T) {
	repo, mock, _ := newRepositoryMock(t)

	mock.
		ExpectExec(regexp.QuoteMeta(`
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
		`)).
		WithArgs(
			5,
			true,
			"Excellent work",
			sqlmock.AnyArg(),
			uint(99),
		).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := repo.Close(
		context.Background(),
		99,
		5,
		true,
		"Excellent work",
	)

	if !errors.Is(err, ErrBookingNotFound) {
		t.Fatalf("expected ErrBookingNotFound got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLRepository_Close_ExecError(t *testing.T) {
	repo, mock, _ := newRepositoryMock(t)

	expected := errors.New("database failed")

	mock.
		ExpectExec(regexp.QuoteMeta(`
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
		`)).
		WithArgs(
			5,
			true,
			"Excellent work",
			sqlmock.AnyArg(),
			uint(1),
		).
		WillReturnError(expected)

	err := repo.Close(
		context.Background(),
		1,
		5,
		true,
		"Excellent work",
	)

	if !errors.Is(err, expected) {
		t.Fatalf("expected %v got %v", expected, err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLRepository_Close_RowsAffectedError(t *testing.T) {
	repo, mock, _ := newRepositoryMock(t)

	expected := errors.New("rows affected failed")

	mock.
		ExpectExec(regexp.QuoteMeta(`
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
		`)).
		WithArgs(
			5,
			true,
			"Excellent work",
			sqlmock.AnyArg(),
			uint(1),
		).
		WillReturnResult(sqlmock.NewErrorResult(expected))

	err := repo.Close(
		context.Background(),
		1,
		5,
		true,
		"Excellent work",
	)

	if !errors.Is(err, expected) {
		t.Fatalf("expected %v got %v", expected, err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLRepository_MarkJobCompleted_Success(t *testing.T) {
	repo, mock, _ := newRepositoryMock(t)

	mock.
		ExpectExec(regexp.QuoteMeta(`
			UPDATE jobs
			SET
				status = 'completed',
				updated_at = $1
			WHERE id = $2
		`)).
		WithArgs(sqlmock.AnyArg(), uint(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.MarkJobCompleted(context.Background(), 1)
	if err != nil {
		t.Fatalf("expected nil got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLRepository_MarkJobCompleted_ExecError(t *testing.T) {
	repo, mock, _ := newRepositoryMock(t)

	expected := errors.New("update failed")

	mock.
		ExpectExec(regexp.QuoteMeta(`
			UPDATE jobs
			SET
				status = 'completed',
				updated_at = $1
			WHERE id = $2
		`)).
		WithArgs(sqlmock.AnyArg(), uint(1)).
		WillReturnError(expected)

	err := repo.MarkJobCompleted(context.Background(), 1)

	if !errors.Is(err, expected) {
		t.Fatalf("expected %v got %v", expected, err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLRepository_MarkApplicationAccepted_Success(t *testing.T) {
	repo, mock, _ := newRepositoryMock(t)

	mock.
		ExpectExec(regexp.QuoteMeta(`
			UPDATE applications
			SET status = 'accepted',
				updated_at = $1
			WHERE id = $2
		`)).
		WithArgs(sqlmock.AnyArg(), uint(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.MarkApplicationAccepted(context.Background(), 1)
	if err != nil {
		t.Fatalf("expected nil got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLRepository_MarkApplicationAccepted_ExecError(t *testing.T) {
	repo, mock, _ := newRepositoryMock(t)

	expected := errors.New("update failed")

	mock.
		ExpectExec(regexp.QuoteMeta(`
			UPDATE applications
			SET status = 'accepted',
				updated_at = $1
			WHERE id = $2
		`)).
		WithArgs(sqlmock.AnyArg(), uint(1)).
		WillReturnError(expected)

	err := repo.MarkApplicationAccepted(context.Background(), 1)

	if !errors.Is(err, expected) {
		t.Fatalf("expected %v got %v", expected, err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func validCloseBookingTransaction() CloseBookingTransaction {
	return CloseBookingTransaction{
		BookingID:           1,
		JobID:               10,
		ClientID:            20,
		CleanerID:           30,
		Rating:              5,
		WouldHireAgain:      true,
		Comment:             "Excellent work",
		AddToFavourites:     true,
		SetAsPreferred:      true,
		ChangedBy:           20,
		PreviousStatus:      "completed",
		NotificationTitle:   "Booking closed",
		NotificationMessage: "The client closed the booking",
	}
}

func TestSQLRepository_CloseWithTransaction_Success(t *testing.T) {
	repo, mock, _ := newRepositoryMock(t)
	input := validCloseBookingTransaction()

	mock.ExpectBegin()

	mock.
		ExpectExec(closeBookingQuery()).
		WithArgs(
			input.Rating,
			input.WouldHireAgain,
			input.Comment,
			sqlmock.AnyArg(),
			input.BookingID,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.
		ExpectExec(createReviewQuery()).
		WithArgs(
			input.BookingID,
			input.JobID,
			input.CleanerID,
			input.ClientID,
			input.Rating,
			input.Comment,
			sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.
		ExpectExec(addFavouriteQuery()).
		WithArgs(
			input.ClientID,
			input.CleanerID,
			sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.
		ExpectExec(addPreferredQuery()).
		WithArgs(
			input.ClientID,
			input.CleanerID,
			sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.
		ExpectExec(createTimelineQuery()).
		WithArgs(
			input.BookingID,
			input.ChangedBy,
			input.PreviousStatus,
			"Client confirmed completion and closed the booking",
			sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.
		ExpectExec(createNotificationQuery()).
		WithArgs(
			input.CleanerID,
			input.NotificationTitle,
			input.NotificationMessage,
			sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	err := repo.CloseWithTransaction(context.Background(), input)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLRepository_CloseWithTransaction_BeginError(t *testing.T) {
	repo, mock, _ := newRepositoryMock(t)
	input := validCloseBookingTransaction()

	expected := errors.New("begin transaction failed")

	mock.ExpectBegin().WillReturnError(expected)

	err := repo.CloseWithTransaction(context.Background(), input)

	if !errors.Is(err, expected) {
		t.Fatalf("expected %v, got %v", expected, err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLRepository_CloseWithTransaction_BookingUpdateError(t *testing.T) {
	repo, mock, _ := newRepositoryMock(t)
	input := validCloseBookingTransaction()

	expected := errors.New("booking update failed")

	mock.ExpectBegin()

	mock.
		ExpectExec(closeBookingQuery()).
		WithArgs(
			input.Rating,
			input.WouldHireAgain,
			input.Comment,
			sqlmock.AnyArg(),
			input.BookingID,
		).
		WillReturnError(expected)

	mock.ExpectRollback()

	err := repo.CloseWithTransaction(context.Background(), input)

	if !errors.Is(err, expected) {
		t.Fatalf("expected %v, got %v", expected, err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLRepository_CloseWithTransaction_RowsAffectedError(t *testing.T) {
	repo, mock, _ := newRepositoryMock(t)
	input := validCloseBookingTransaction()

	expected := errors.New("rows affected failed")

	mock.ExpectBegin()

	mock.
		ExpectExec(closeBookingQuery()).
		WithArgs(
			input.Rating,
			input.WouldHireAgain,
			input.Comment,
			sqlmock.AnyArg(),
			input.BookingID,
		).
		WillReturnResult(sqlmock.NewErrorResult(expected))

	mock.ExpectRollback()

	err := repo.CloseWithTransaction(context.Background(), input)

	if !errors.Is(err, expected) {
		t.Fatalf("expected %v, got %v", expected, err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLRepository_CloseWithTransaction_NoRowsUpdated(t *testing.T) {
	repo, mock, _ := newRepositoryMock(t)
	input := validCloseBookingTransaction()

	mock.ExpectBegin()

	mock.
		ExpectExec(closeBookingQuery()).
		WithArgs(
			input.Rating,
			input.WouldHireAgain,
			input.Comment,
			sqlmock.AnyArg(),
			input.BookingID,
		).
		WillReturnResult(sqlmock.NewResult(0, 0))

	mock.ExpectRollback()

	err := repo.CloseWithTransaction(context.Background(), input)

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLRepository_CloseWithTransaction_ReviewError(t *testing.T) {
	repo, mock, _ := newRepositoryMock(t)
	input := validCloseBookingTransaction()

	expected := errors.New("review insert failed")

	mock.ExpectBegin()

	mock.
		ExpectExec(closeBookingQuery()).
		WithArgs(
			input.Rating,
			input.WouldHireAgain,
			input.Comment,
			sqlmock.AnyArg(),
			input.BookingID,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.
		ExpectExec(createReviewQuery()).
		WithArgs(
			input.BookingID,
			input.JobID,
			input.CleanerID,
			input.ClientID,
			input.Rating,
			input.Comment,
			sqlmock.AnyArg(),
		).
		WillReturnError(expected)

	mock.ExpectRollback()

	err := repo.CloseWithTransaction(context.Background(), input)

	if !errors.Is(err, expected) {
		t.Fatalf("expected %v, got %v", expected, err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLRepository_CloseWithTransaction_FavouriteError(t *testing.T) {
	repo, mock, _ := newRepositoryMock(t)
	input := validCloseBookingTransaction()

	expected := errors.New("favourite insert failed")

	mock.ExpectBegin()

	mock.
		ExpectExec(closeBookingQuery()).
		WithArgs(
			input.Rating,
			input.WouldHireAgain,
			input.Comment,
			sqlmock.AnyArg(),
			input.BookingID,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.
		ExpectExec(createReviewQuery()).
		WithArgs(
			input.BookingID,
			input.JobID,
			input.CleanerID,
			input.ClientID,
			input.Rating,
			input.Comment,
			sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.
		ExpectExec(addFavouriteQuery()).
		WithArgs(
			input.ClientID,
			input.CleanerID,
			sqlmock.AnyArg(),
		).
		WillReturnError(expected)

	mock.ExpectRollback()

	err := repo.CloseWithTransaction(context.Background(), input)

	if !errors.Is(err, expected) {
		t.Fatalf("expected %v, got %v", expected, err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLRepository_CloseWithTransaction_PreferredError(t *testing.T) {
	repo, mock, _ := newRepositoryMock(t)
	input := validCloseBookingTransaction()

	input.AddToFavourites = false

	expected := errors.New("preferred insert failed")

	mock.ExpectBegin()

	mock.
		ExpectExec(closeBookingQuery()).
		WithArgs(
			input.Rating,
			input.WouldHireAgain,
			input.Comment,
			sqlmock.AnyArg(),
			input.BookingID,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.
		ExpectExec(createReviewQuery()).
		WithArgs(
			input.BookingID,
			input.JobID,
			input.CleanerID,
			input.ClientID,
			input.Rating,
			input.Comment,
			sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.
		ExpectExec(addPreferredQuery()).
		WithArgs(
			input.ClientID,
			input.CleanerID,
			sqlmock.AnyArg(),
		).
		WillReturnError(expected)

	mock.ExpectRollback()

	err := repo.CloseWithTransaction(context.Background(), input)

	if !errors.Is(err, expected) {
		t.Fatalf("expected %v, got %v", expected, err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLRepository_CloseWithTransaction_TimelineError(t *testing.T) {
	repo, mock, _ := newRepositoryMock(t)
	input := validCloseBookingTransaction()

	input.AddToFavourites = false
	input.SetAsPreferred = false

	expected := errors.New("timeline insert failed")

	mock.ExpectBegin()

	mock.
		ExpectExec(closeBookingQuery()).
		WithArgs(
			input.Rating,
			input.WouldHireAgain,
			input.Comment,
			sqlmock.AnyArg(),
			input.BookingID,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.
		ExpectExec(createReviewQuery()).
		WithArgs(
			input.BookingID,
			input.JobID,
			input.CleanerID,
			input.ClientID,
			input.Rating,
			input.Comment,
			sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.
		ExpectExec(createTimelineQuery()).
		WithArgs(
			input.BookingID,
			input.ChangedBy,
			input.PreviousStatus,
			"Client confirmed completion and closed the booking",
			sqlmock.AnyArg(),
		).
		WillReturnError(expected)

	mock.ExpectRollback()

	err := repo.CloseWithTransaction(context.Background(), input)

	if !errors.Is(err, expected) {
		t.Fatalf("expected %v, got %v", expected, err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLRepository_CloseWithTransaction_NotificationError(t *testing.T) {
	repo, mock, _ := newRepositoryMock(t)
	input := validCloseBookingTransaction()

	input.AddToFavourites = false
	input.SetAsPreferred = false

	expected := errors.New("notification insert failed")

	mock.ExpectBegin()

	mock.
		ExpectExec(closeBookingQuery()).
		WithArgs(
			input.Rating,
			input.WouldHireAgain,
			input.Comment,
			sqlmock.AnyArg(),
			input.BookingID,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.
		ExpectExec(createReviewQuery()).
		WithArgs(
			input.BookingID,
			input.JobID,
			input.CleanerID,
			input.ClientID,
			input.Rating,
			input.Comment,
			sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.
		ExpectExec(createTimelineQuery()).
		WithArgs(
			input.BookingID,
			input.ChangedBy,
			input.PreviousStatus,
			"Client confirmed completion and closed the booking",
			sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.
		ExpectExec(createNotificationQuery()).
		WithArgs(
			input.CleanerID,
			input.NotificationTitle,
			input.NotificationMessage,
			sqlmock.AnyArg(),
		).
		WillReturnError(expected)

	mock.ExpectRollback()

	err := repo.CloseWithTransaction(context.Background(), input)

	if !errors.Is(err, expected) {
		t.Fatalf("expected %v, got %v", expected, err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLRepository_CloseWithTransaction_CommitError(t *testing.T) {
	repo, mock, _ := newRepositoryMock(t)
	input := validCloseBookingTransaction()

	input.AddToFavourites = false
	input.SetAsPreferred = false

	expected := errors.New("commit failed")

	mock.ExpectBegin()

	mock.
		ExpectExec(closeBookingQuery()).
		WithArgs(
			input.Rating,
			input.WouldHireAgain,
			input.Comment,
			sqlmock.AnyArg(),
			input.BookingID,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.
		ExpectExec(createReviewQuery()).
		WithArgs(
			input.BookingID,
			input.JobID,
			input.CleanerID,
			input.ClientID,
			input.Rating,
			input.Comment,
			sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.
		ExpectExec(createTimelineQuery()).
		WithArgs(
			input.BookingID,
			input.ChangedBy,
			input.PreviousStatus,
			"Client confirmed completion and closed the booking",
			sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.
		ExpectExec(createNotificationQuery()).
		WithArgs(
			input.CleanerID,
			input.NotificationTitle,
			input.NotificationMessage,
			sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit().WillReturnError(expected)

	err := repo.CloseWithTransaction(context.Background(), input)

	if !errors.Is(err, expected) {
		t.Fatalf("expected %v, got %v", expected, err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLRepository_HasScheduleConflict_True(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	startAt := time.Now().Add(24 * time.Hour)
	endAt := startAt.Add(2 * time.Hour)

	mock.ExpectQuery(
		`(?s)SELECT EXISTS.*FROM bookings.*cleaner_id = \$1`,
	).
		WithArgs(
			uint(30),
			uint(0),
			startAt,
			endAt,
		).
		WillReturnRows(
			sqlmock.NewRows([]string{"exists"}).
				AddRow(true),
		)

	conflict, err := repo.HasScheduleConflict(
		context.Background(),
		30,
		startAt,
		endAt,
		0,
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !conflict {
		t.Fatal("expected conflict")
	}
}

func TestSQLRepository_HasScheduleConflict_False(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	startAt := time.Now().Add(24 * time.Hour)
	endAt := startAt.Add(2 * time.Hour)

	mock.ExpectQuery(
		`(?s)SELECT EXISTS.*FROM bookings.*cleaner_id = \$1`,
	).
		WithArgs(
			uint(30),
			uint(0),
			startAt,
			endAt,
		).
		WillReturnRows(
			sqlmock.NewRows([]string{"exists"}).
				AddRow(false),
		)

	conflict, err := repo.HasScheduleConflict(
		context.Background(),
		30,
		startAt,
		endAt,
		0,
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if conflict {
		t.Fatal("expected no conflict")
	}
}

func TestSQLRepository_HasScheduleConflict_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	startAt := time.Now().Add(24 * time.Hour)
	endAt := startAt.Add(2 * time.Hour)

	expectedErr := errors.New("conflict query failed")

	mock.ExpectQuery(
		`(?s)SELECT EXISTS.*FROM bookings.*cleaner_id = \$1`,
	).
		WithArgs(
			uint(30),
			uint(0),
			startAt,
			endAt,
		).
		WillReturnError(expectedErr)

	_, err = repo.HasScheduleConflict(
		context.Background(),
		30,
		startAt,
		endAt,
		0,
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected %v got %v",
			expectedErr,
			err,
		)
	}
}
