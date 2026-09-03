package jobpulse

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestSQLRepository_GetSnapshot_Success(
	t *testing.T,
) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	createdAt :=
		time.Now().
			UTC().
			Add(-12 * time.Hour)

	mock.ExpectQuery(
		regexp.QuoteMeta(
			"SELECT",
		),
	).
		WithArgs(
			uint(10),
			sqlmock.AnyArg(),
		).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{
					"id",
					"status",
					"created_at",
					"application_count",
					"recent_application_count",
					"booking_count",
				},
			).AddRow(
				10,
				"open",
				createdAt,
				7,
				4,
				0,
			),
		)

	result, err :=
		repo.GetSnapshot(
			context.Background(),
			10,
		)
	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}

	if result == nil {
		t.Fatal(
			"expected snapshot",
		)
	}

	if result.JobID != 10 {
		t.Fatalf(
			"expected job ID 10, got %d",
			result.JobID,
		)
	}

	if result.JobStatus != "open" {
		t.Fatalf(
			"expected open status, got %q",
			result.JobStatus,
		)
	}

	if result.ApplicationCount != 7 {
		t.Fatalf(
			"expected 7 applications, got %d",
			result.ApplicationCount,
		)
	}

	if result.RecentApplicationCount != 4 {
		t.Fatalf(
			"expected 4 recent applications, got %d",
			result.RecentApplicationCount,
		)
	}

	if result.BookingCount != 0 {
		t.Fatalf(
			"expected zero bookings, got %d",
			result.BookingCount,
		)
	}

	if err :=
		mock.ExpectationsWereMet(); err != nil {
		t.Fatalf(
			"unmet expectations: %v",
			err,
		)
	}
}

func TestSQLRepository_GetSnapshot_WithBooking(
	t *testing.T,
) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	mock.ExpectQuery(
		regexp.QuoteMeta(
			"SELECT",
		),
	).
		WithArgs(
			uint(10),
			sqlmock.AnyArg(),
		).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{
					"id",
					"status",
					"created_at",
					"application_count",
					"recent_application_count",
					"booking_count",
				},
			).AddRow(
				10,
				"open",
				time.Now().
					UTC().
					Add(-24*time.Hour),
				5,
				1,
				1,
			),
		)

	result, err :=
		repo.GetSnapshot(
			context.Background(),
			10,
		)
	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}

	if result.BookingCount != 1 {
		t.Fatalf(
			"expected one booking, got %d",
			result.BookingCount,
		)
	}
}

func TestSQLRepository_GetSnapshot_NotFound(
	t *testing.T,
) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	mock.ExpectQuery(
		regexp.QuoteMeta(
			"SELECT",
		),
	).
		WithArgs(
			uint(999),
			sqlmock.AnyArg(),
		).
		WillReturnError(
			sql.ErrNoRows,
		)

	result, err :=
		repo.GetSnapshot(
			context.Background(),
			999,
		)

	if result != nil {
		t.Fatalf(
			"expected nil result, got %+v",
			result,
		)
	}

	if !errors.Is(
		err,
		ErrJobNotFound,
	) {
		t.Fatalf(
			"expected ErrJobNotFound, got %v",
			err,
		)
	}
}

func TestSQLRepository_GetSnapshot_DatabaseError(
	t *testing.T,
) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	expectedErr :=
		errors.New(
			"database failed",
		)

	mock.ExpectQuery(
		regexp.QuoteMeta(
			"SELECT",
		),
	).
		WithArgs(
			uint(10),
			sqlmock.AnyArg(),
		).
		WillReturnError(
			expectedErr,
		)

	result, err :=
		repo.GetSnapshot(
			context.Background(),
			10,
		)

	if result != nil {
		t.Fatalf(
			"expected nil result, got %+v",
			result,
		)
	}

	if !errors.Is(
		err,
		expectedErr,
	) {
		t.Fatalf(
			"expected %v, got %v",
			expectedErr,
			err,
		)
	}
}
