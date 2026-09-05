package workhistory

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func newWorkHistoryMockDB(
	t *testing.T,
) (*sql.DB, sqlmock.Sqlmock) {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf(
			"failed creating sqlmock: %v",
			err,
		)
	}

	t.Cleanup(
		func() {
			_ = db.Close()
		},
	)

	return db, mock
}

func workHistoryColumns() []string {
	return []string{
		"id",
		"job_id",
		"client_id",
		"client_name",
		"cleaner_id",
		"cleaner_name",
		"job_title",
		"job_type",
		"location",
		"budget",
		"status",
		"scheduled_at",
		"scheduled_end_at",
		"completed_at",
		"cancelled_at",
		"cancellation_reason",
		"cancelled_by",
		"closed_at",
		"closure_status",
		"closure_comment",
		"client_confirmed_completion",
		"client_would_hire_again",
		"rating",
		"comment",
		"before_count",
		"after_count",
		"last_uploaded_at",
		"is_repeat_client",
		"created_at",
		"updated_at",
	}
}

func workHistoryRow(
	now time.Time,
) *sqlmock.Rows {
	rating := 5
	comment := "Great cleaner"
	wouldHireAgain := true
	cancelledBy := uint(5)

	return sqlmock.NewRows(
		workHistoryColumns(),
	).AddRow(
		10,
		20,
		5,
		"Client User",
		8,
		"Cleaner User",
		"End of tenancy clean",
		"end_of_tenancy",
		"London",
		150,
		"completed",
		now,
		now.Add(2*time.Hour),
		now.Add(2*time.Hour),
		nil,
		"",
		cancelledBy,
		now.Add(2*time.Hour),
		"completed",
		"Job complete",
		true,
		wouldHireAgain,
		rating,
		comment,
		2,
		3,
		now.Add(time.Hour),
		true,
		now.Add(-24*time.Hour),
		now,
	)
}

func TestSQLRepository_ListByCleanerID_Success(
	t *testing.T,
) {
	db, mock := newWorkHistoryMockDB(t)

	repo := NewSQLRepository(db)

	now := time.Now()

	mock.ExpectQuery(
		regexp.QuoteMeta(
			workHistorySelect + `
		WHERE b.cleaner_id = $1
		ORDER BY
			COALESCE(
				b.completed_at,
				b.closed_at,
				b.cancelled_at,
				b.created_at
			) DESC
	`,
		),
	).
		WithArgs(uint(8)).
		WillReturnRows(
			workHistoryRow(now),
		)

	history, err := repo.ListByCleanerID(
		context.Background(),
		8,
	)
	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}

	if len(history) != 1 {
		t.Fatalf(
			"expected 1 history entry, got %d",
			len(history),
		)
	}

	entry := history[0]

	if entry.BookingID != 10 {
		t.Fatalf(
			"expected booking id 10, got %d",
			entry.BookingID,
		)
	}

	if !entry.WorkProof.HasBeforeProof {
		t.Fatal(
			"expected before proof",
		)
	}

	if !entry.WorkProof.HasAfterProof {
		t.Fatal(
			"expected after proof",
		)
	}

	if !entry.WorkProof.IsVerifiedWork {
		t.Fatal(
			"expected verified work",
		)
	}

	if !entry.CanBookAgain {
		t.Fatal(
			"expected booking to be rebookable",
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLRepository_ListByCleanerID_Empty(
	t *testing.T,
) {
	db, mock := newWorkHistoryMockDB(t)

	repo := NewSQLRepository(db)

	mock.ExpectQuery(
		regexp.QuoteMeta(
			workHistorySelect + `
		WHERE b.cleaner_id = $1
		ORDER BY
			COALESCE(
				b.completed_at,
				b.closed_at,
				b.cancelled_at,
				b.created_at
			) DESC
	`,
		),
	).
		WithArgs(uint(8)).
		WillReturnRows(
			sqlmock.NewRows(
				workHistoryColumns(),
			),
		)

	history, err := repo.ListByCleanerID(
		context.Background(),
		8,
	)
	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}

	if len(history) != 0 {
		t.Fatalf(
			"expected empty history, got %d",
			len(history),
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLRepository_ListByCleanerID_QueryError(
	t *testing.T,
) {
	db, mock := newWorkHistoryMockDB(t)

	repo := NewSQLRepository(db)

	expectedErr := errors.New(
		"query failed",
	)

	mock.ExpectQuery(
		regexp.QuoteMeta(
			workHistorySelect + `
		WHERE b.cleaner_id = $1
		ORDER BY
			COALESCE(
				b.completed_at,
				b.closed_at,
				b.cancelled_at,
				b.created_at
			) DESC
	`,
		),
	).
		WithArgs(uint(8)).
		WillReturnError(expectedErr)

	_, err := repo.ListByCleanerID(
		context.Background(),
		8,
	)

	if !errors.Is(
		err,
		expectedErr,
	) {
		t.Fatalf(
			"expected query error, got %v",
			err,
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLRepository_ListByClientID_Success(
	t *testing.T,
) {
	db, mock := newWorkHistoryMockDB(t)

	repo := NewSQLRepository(db)

	now := time.Now()

	mock.ExpectQuery(
		regexp.QuoteMeta(
			workHistorySelect + `
		WHERE b.client_id = $1
		ORDER BY
			COALESCE(
				b.completed_at,
				b.closed_at,
				b.cancelled_at,
				b.created_at
			) DESC
	`,
		),
	).
		WithArgs(uint(5)).
		WillReturnRows(
			workHistoryRow(now),
		)

	history, err := repo.ListByClientID(
		context.Background(),
		5,
	)
	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}

	if len(history) != 1 {
		t.Fatalf(
			"expected 1 history entry, got %d",
			len(history),
		)
	}

	if history[0].ClientID != 5 {
		t.Fatalf(
			"expected client id 5, got %d",
			history[0].ClientID,
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLRepository_ListByClientID_QueryError(
	t *testing.T,
) {
	db, mock := newWorkHistoryMockDB(t)

	repo := NewSQLRepository(db)

	expectedErr := errors.New(
		"query failed",
	)

	mock.ExpectQuery(
		regexp.QuoteMeta(
			workHistorySelect + `
		WHERE b.client_id = $1
		ORDER BY
			COALESCE(
				b.completed_at,
				b.closed_at,
				b.cancelled_at,
				b.created_at
			) DESC
	`,
		),
	).
		WithArgs(uint(5)).
		WillReturnError(expectedErr)

	_, err := repo.ListByClientID(
		context.Background(),
		5,
	)

	if !errors.Is(
		err,
		expectedErr,
	) {
		t.Fatalf(
			"expected query error, got %v",
			err,
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLRepository_GetByBookingID_Success(
	t *testing.T,
) {
	db, mock := newWorkHistoryMockDB(t)

	repo := NewSQLRepository(db)

	now := time.Now()

	mock.ExpectQuery(
		regexp.QuoteMeta(
			workHistorySelect + `
		WHERE b.id = $1
		LIMIT 1
	`,
		),
	).
		WithArgs(uint(10)).
		WillReturnRows(
			workHistoryRow(now),
		)

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

	mock.ExpectQuery(
		regexp.QuoteMeta(proofQuery),
	).
		WithArgs(uint(10)).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{
					"id",
					"proof_type",
					"photo_url",
					"caption",
					"created_at",
				},
			).
				AddRow(
					1,
					"before",
					"https://example.com/before.jpg",
					"Before cleaning",
					now,
				).
				AddRow(
					2,
					"after",
					"https://example.com/after.jpg",
					"After cleaning",
					now.Add(time.Hour),
				),
		)

	history, err := repo.GetByBookingID(
		context.Background(),
		10,
		5,
	)
	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}

	if history.BookingID != 10 {
		t.Fatalf(
			"expected booking id 10, got %d",
			history.BookingID,
		)
	}

	if len(history.BeforePhotos) != 1 {
		t.Fatalf(
			"expected 1 before photo, got %d",
			len(history.BeforePhotos),
		)
	}

	if len(history.AfterPhoto) != 1 {
		t.Fatalf(
			"expected 1 after photo, got %d",
			len(history.AfterPhoto),
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLRepository_GetByBookingID_NotFound(
	t *testing.T,
) {
	db, mock := newWorkHistoryMockDB(t)

	repo := NewSQLRepository(db)

	mock.ExpectQuery(
		regexp.QuoteMeta(
			workHistorySelect + `
		WHERE b.id = $1
		LIMIT 1
	`,
		),
	).
		WithArgs(uint(999)).
		WillReturnError(
			sql.ErrNoRows,
		)

	_, err := repo.GetByBookingID(
		context.Background(),
		999,
		5,
	)

	if !errors.Is(
		err,
		ErrHistoryNotFound,
	) {
		t.Fatalf(
			"expected ErrHistoryNotFound, got %v",
			err,
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLRepository_GetByBookingID_QueryError(
	t *testing.T,
) {
	db, mock := newWorkHistoryMockDB(t)

	repo := NewSQLRepository(db)

	expectedErr := errors.New(
		"database failed",
	)

	mock.ExpectQuery(
		regexp.QuoteMeta(
			workHistorySelect + `
		WHERE b.id = $1
		LIMIT 1
	`,
		),
	).
		WithArgs(uint(10)).
		WillReturnError(expectedErr)

	_, err := repo.GetByBookingID(
		context.Background(),
		10,
		5,
	)

	if !errors.Is(
		err,
		expectedErr,
	) {
		t.Fatalf(
			"expected database error, got %v",
			err,
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLRepository_UserCanAccessBooking_True(
	t *testing.T,
) {
	db, mock := newWorkHistoryMockDB(t)

	repo := NewSQLRepository(db)

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

	mock.ExpectQuery(
		regexp.QuoteMeta(query),
	).
		WithArgs(
			uint(10),
			uint(5),
		).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{
					"exists",
				},
			).AddRow(true),
		)

	allowed, err :=
		repo.UserCanAccessBooking(
			context.Background(),
			10,
			5,
		)
	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}

	if !allowed {
		t.Fatal(
			"expected access allowed",
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLRepository_UserCanAccessBooking_False(
	t *testing.T,
) {
	db, mock := newWorkHistoryMockDB(t)

	repo := NewSQLRepository(db)

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

	mock.ExpectQuery(
		regexp.QuoteMeta(query),
	).
		WithArgs(
			uint(10),
			uint(50),
		).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{
					"exists",
				},
			).AddRow(false),
		)

	allowed, err :=
		repo.UserCanAccessBooking(
			context.Background(),
			10,
			50,
		)
	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}

	if allowed {
		t.Fatal(
			"expected access denied",
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLRepository_UserCanAccessBooking_QueryError(
	t *testing.T,
) {
	db, mock := newWorkHistoryMockDB(t)

	repo := NewSQLRepository(db)

	expectedErr := errors.New(
		"access query failed",
	)

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

	mock.ExpectQuery(
		regexp.QuoteMeta(query),
	).
		WithArgs(
			uint(10),
			uint(5),
		).
		WillReturnError(expectedErr)

	_, err :=
		repo.UserCanAccessBooking(
			context.Background(),
			10,
			5,
		)

	if !errors.Is(
		err,
		expectedErr,
	) {
		t.Fatalf(
			"expected access query error, got %v",
			err,
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
