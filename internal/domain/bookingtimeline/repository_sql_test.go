package bookingtimeline

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func newRepositoryTest(
	t *testing.T,
) (
	*SQLRepository,
	sqlmock.Sqlmock,
	func(),
) {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf(
			"failed to create sqlmock: %v",
			err,
		)
	}

	cleanup := func() {
		_ = db.Close()
	}

	return NewSQLRepository(db), mock, cleanup
}

func TestSQLRepository_CreateHistory_Success(
	t *testing.T,
) {
	repo, mock, cleanup := newRepositoryTest(t)
	defer cleanup()

	query := regexp.QuoteMeta(`
		INSERT INTO booking_status_history (
			booking_id,
			changed_by,
			from_status,
			to_status,
			note,
			created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6)
	`)

	mock.ExpectExec(query).
		WithArgs(
			uint(12),
			uint(8),
			"accepted",
			"completed",
			"completed successfully",
			sqlmock.AnyArg(),
		).
		WillReturnResult(
			sqlmock.NewResult(1, 1),
		)

	err := repo.CreateHistory(
		context.Background(),
		12,
		8,
		"accepted",
		"completed",
		"completed successfully",
	)
	if err != nil {
		t.Fatalf(
			"expected nil error, got %v",
			err,
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf(
			"unmet SQL expectations: %v",
			err,
		)
	}
}

func TestSQLRepository_CreateHistory_NullChangedBy(
	t *testing.T,
) {
	repo, mock, cleanup := newRepositoryTest(t)
	defer cleanup()

	query := regexp.QuoteMeta(`
		INSERT INTO booking_status_history (
			booking_id,
			changed_by,
			from_status,
			to_status,
			note,
			created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6)
	`)

	mock.ExpectExec(query).
		WithArgs(
			uint(12),
			nil,
			"",
			"confirmed",
			"",
			sqlmock.AnyArg(),
		).
		WillReturnResult(
			sqlmock.NewResult(1, 1),
		)

	err := repo.CreateHistory(
		context.Background(),
		12,
		0,
		"",
		"confirmed",
		"",
	)
	if err != nil {
		t.Fatalf(
			"expected nil error, got %v",
			err,
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf(
			"unmet SQL expectations: %v",
			err,
		)
	}
}

func TestSQLRepository_CreateHistory_Error(
	t *testing.T,
) {
	repo, mock, cleanup := newRepositoryTest(t)
	defer cleanup()

	expectedErr := errors.New(
		"insert failed",
	)

	mock.ExpectExec(
		regexp.QuoteMeta(
			"INSERT INTO booking_status_history",
		),
	).
		WillReturnError(expectedErr)

	err := repo.CreateHistory(
		context.Background(),
		12,
		8,
		"accepted",
		"completed",
		"",
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected insert error, got %v",
			err,
		)
	}
}

func TestSQLRepository_ListByBookingID_Success(
	t *testing.T,
) {
	repo, mock, cleanup := newRepositoryTest(t)
	defer cleanup()

	now := time.Now()

	query := regexp.QuoteMeta(`
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
	`)

	rows := sqlmock.NewRows(
		[]string{
			"id",
			"booking_id",
			"changed_by",
			"from_status",
			"to_status",
			"note",
			"created_at",
		},
	).
		AddRow(
			1,
			12,
			5,
			"pending",
			"accepted",
			"booking accepted",
			now,
		).
		AddRow(
			2,
			12,
			8,
			"accepted",
			"completed",
			"job completed",
			now.Add(time.Hour),
		)

	mock.ExpectQuery(query).
		WithArgs(uint(12)).
		WillReturnRows(rows)

	history, err := repo.ListByBookingID(
		context.Background(),
		12,
	)
	if err != nil {
		t.Fatalf(
			"expected nil error, got %v",
			err,
		)
	}

	if len(history) != 2 {
		t.Fatalf(
			"expected 2 history items, got %d",
			len(history),
		)
	}

	if history[0].ID != 1 {
		t.Fatalf(
			"expected first history ID 1, got %d",
			history[0].ID,
		)
	}

	if history[0].FromStatus != "pending" {
		t.Fatalf(
			"expected pending, got %q",
			history[0].FromStatus,
		)
	}

	if history[1].ToStatus != "completed" {
		t.Fatalf(
			"expected completed, got %q",
			history[1].ToStatus,
		)
	}

	if history[0].ChangedBy == nil ||
		*history[0].ChangedBy != 5 {
		t.Fatal(
			"expected first changed_by to be 5",
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf(
			"unmet SQL expectations: %v",
			err,
		)
	}
}

func TestSQLRepository_ListByBookingID_NullChangedBy(
	t *testing.T,
) {
	repo, mock, cleanup := newRepositoryTest(t)
	defer cleanup()

	now := time.Now()

	rows := sqlmock.NewRows(
		[]string{
			"id",
			"booking_id",
			"changed_by",
			"from_status",
			"to_status",
			"note",
			"created_at",
		},
	).
		AddRow(
			1,
			12,
			nil,
			"",
			"confirmed",
			"",
			now,
		)

	mock.ExpectQuery(
		regexp.QuoteMeta(`
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
		`),
	).
		WithArgs(uint(12)).
		WillReturnRows(rows)

	history, err := repo.ListByBookingID(
		context.Background(),
		12,
	)
	if err != nil {
		t.Fatalf(
			"expected nil error, got %v",
			err,
		)
	}

	if len(history) != 1 {
		t.Fatalf(
			"expected 1 history item, got %d",
			len(history),
		)
	}

	if history[0].ChangedBy != nil {
		t.Fatalf(
			"expected nil changed_by, got %v",
			history[0].ChangedBy,
		)
	}
}

func TestSQLRepository_ListByBookingID_QueryError(
	t *testing.T,
) {
	repo, mock, cleanup := newRepositoryTest(t)
	defer cleanup()

	expectedErr := errors.New(
		"query failed",
	)

	mock.ExpectQuery(
		regexp.QuoteMeta(`
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
		`),
	).
		WithArgs(uint(12)).
		WillReturnError(expectedErr)

	history, err := repo.ListByBookingID(
		context.Background(),
		12,
	)

	if history != nil {
		t.Fatal(
			"expected nil history",
		)
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected query error, got %v",
			err,
		)
	}
}

func TestSQLRepository_ListByBookingID_ScanError(
	t *testing.T,
) {
	repo, mock, cleanup := newRepositoryTest(t)
	defer cleanup()

	rows := sqlmock.NewRows(
		[]string{
			"id",
			"booking_id",
			"changed_by",
			"from_status",
			"to_status",
			"note",
			"created_at",
		},
	).
		AddRow(
			"invalid-id",
			12,
			5,
			"pending",
			"accepted",
			"",
			time.Now(),
		)

	mock.ExpectQuery(
		regexp.QuoteMeta(`
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
		`),
	).
		WithArgs(uint(12)).
		WillReturnRows(rows)

	history, err := repo.ListByBookingID(
		context.Background(),
		12,
	)

	if history != nil {
		t.Fatal(
			"expected nil history",
		)
	}

	if err == nil {
		t.Fatal(
			"expected scan error",
		)
	}
}

func TestSQLRepository_ListByBookingID_RowsError(
	t *testing.T,
) {
	repo, mock, cleanup := newRepositoryTest(t)
	defer cleanup()

	expectedErr := errors.New(
		"rows failed",
	)

	rows := sqlmock.NewRows(
		[]string{
			"id",
			"booking_id",
			"changed_by",
			"from_status",
			"to_status",
			"note",
			"created_at",
		},
	).
		AddRow(
			1,
			12,
			5,
			"pending",
			"accepted",
			"",
			time.Now(),
		).
		RowError(
			0,
			expectedErr,
		)

	mock.ExpectQuery(
		regexp.QuoteMeta(`
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
		`),
	).
		WithArgs(uint(12)).
		WillReturnRows(rows)

	_, err := repo.ListByBookingID(
		context.Background(),
		12,
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected rows error, got %v",
			err,
		)
	}
}

func TestSQLRepository_GetBookingAccess_Success(
	t *testing.T,
) {
	repo, mock, cleanup := newRepositoryTest(t)
	defer cleanup()

	query := regexp.QuoteMeta(`
		SELECT
			client_id,
			cleaner_id,
			status
		FROM bookings
		WHERE id = $1
		LIMIT 1
	`)

	mock.ExpectQuery(query).
		WithArgs(uint(12)).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{
					"client_id",
					"cleaner_id",
					"status",
				},
			).
				AddRow(
					5,
					8,
					"accepted",
				),
		)

	clientID,
		cleanerID,
		status,
		err := repo.GetBookingAccess(
		context.Background(),
		12,
	)
	if err != nil {
		t.Fatalf(
			"expected nil error, got %v",
			err,
		)
	}

	if clientID != 5 {
		t.Fatalf(
			"expected client ID 5, got %d",
			clientID,
		)
	}

	if cleanerID != 8 {
		t.Fatalf(
			"expected cleaner ID 8, got %d",
			cleanerID,
		)
	}

	if status != "accepted" {
		t.Fatalf(
			"expected accepted, got %q",
			status,
		)
	}
}

func TestSQLRepository_GetBookingAccess_NotFound(
	t *testing.T,
) {
	repo, mock, cleanup := newRepositoryTest(t)
	defer cleanup()

	mock.ExpectQuery(
		regexp.QuoteMeta(`
			SELECT
				client_id,
				cleaner_id,
				status
			FROM bookings
			WHERE id = $1
			LIMIT 1
		`),
	).
		WithArgs(uint(999)).
		WillReturnError(sql.ErrNoRows)

	clientID,
		cleanerID,
		status,
		err := repo.GetBookingAccess(
		context.Background(),
		999,
	)

	if clientID != 0 ||
		cleanerID != 0 ||
		status != "" {
		t.Fatal(
			"expected zero values for missing booking",
		)
	}

	if !errors.Is(err, ErrBookingNotFound) {
		t.Fatalf(
			"expected ErrBookingNotFound, got %v",
			err,
		)
	}
}

func TestSQLRepository_GetBookingAccess_Error(
	t *testing.T,
) {
	repo, mock, cleanup := newRepositoryTest(t)
	defer cleanup()

	expectedErr := errors.New(
		"database failed",
	)

	mock.ExpectQuery(
		regexp.QuoteMeta(`
			SELECT
				client_id,
				cleaner_id,
				status
			FROM bookings
			WHERE id = $1
			LIMIT 1
		`),
	).
		WithArgs(uint(12)).
		WillReturnError(expectedErr)

	clientID,
		cleanerID,
		status,
		err := repo.GetBookingAccess(
		context.Background(),
		12,
	)

	if clientID != 0 ||
		cleanerID != 0 ||
		status != "" {
		t.Fatal(
			"expected zero values on repository error",
		)
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected database error, got %v",
			err,
		)
	}
}
