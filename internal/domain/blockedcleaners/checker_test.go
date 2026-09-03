package blockedcleaners

import (
	"context"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestNewChecker(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	checker := NewChecker(db)

	if checker == nil {
		t.Fatal("expected checker")
	}

	if checker.db != db {
		t.Fatal("expected database assigned")
	}
}

func TestChecker_IsBlocked_True(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	checker := NewChecker(db)

	mock.ExpectQuery(
		regexp.QuoteMeta(`
			SELECT EXISTS (
				SELECT 1
				FROM blocked_cleaners
				WHERE client_id = $1
				AND cleaner_id = $2
			)
		`),
	).
		WithArgs(
			uint(5),
			uint(8),
		).
		WillReturnRows(
			sqlmock.NewRows([]string{"exists"}).
				AddRow(true),
		)

	blocked, err := checker.IsBlocked(
		context.Background(),
		5,
		8,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !blocked {
		t.Fatal("expected cleaner to be blocked")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestChecker_IsBlocked_False(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	checker := NewChecker(db)

	mock.ExpectQuery(
		regexp.QuoteMeta(`
			SELECT EXISTS (
				SELECT 1
				FROM blocked_cleaners
				WHERE client_id = $1
				AND cleaner_id = $2
			)
		`),
	).
		WithArgs(
			uint(5),
			uint(8),
		).
		WillReturnRows(
			sqlmock.NewRows([]string{"exists"}).
				AddRow(false),
		)

	blocked, err := checker.IsBlocked(
		context.Background(),
		5,
		8,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if blocked {
		t.Fatal("expected cleaner not to be blocked")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestChecker_IsBlocked_DatabaseError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	checker := NewChecker(db)

	expectedErr := errors.New("block check failed")

	mock.ExpectQuery(
		regexp.QuoteMeta(`
			SELECT EXISTS (
				SELECT 1
				FROM blocked_cleaners
				WHERE client_id = $1
				AND cleaner_id = $2
			)
		`),
	).
		WithArgs(
			uint(5),
			uint(8),
		).
		WillReturnError(expectedErr)

	blocked, err := checker.IsBlocked(
		context.Background(),
		5,
		8,
	)

	if blocked {
		t.Fatal("expected blocked to be false")
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected %v, got %v",
			expectedErr,
			err,
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}
