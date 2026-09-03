package reputation

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func newMockRepository(
	t *testing.T,
) (*SQLRepository, sqlmock.Sqlmock, func()) {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create SQL mock: %v", err)
	}

	cleanup := func() {
		_ = db.Close()
	}

	return NewSQLRepository(db), mock, cleanup
}

func TestSQLRepository_EnsureCleaner_Success(t *testing.T) {
	repository, mock, cleanup := newMockRepository(t)
	defer cleanup()

	query := `
		INSERT INTO cleaner_reputation (
			cleaner_id,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $2)
		ON CONFLICT (cleaner_id)
		DO NOTHING
	`

	mock.ExpectExec(regexp.QuoteMeta(query)).
		WithArgs(uint(10), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := repository.EnsureCleaner(context.Background(), 10)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestSQLRepository_EnsureCleaner_DatabaseError(t *testing.T) {
	repository, mock, cleanup := newMockRepository(t)
	defer cleanup()

	expectedErr := errors.New("database insert failed")

	query := `
		INSERT INTO cleaner_reputation (
			cleaner_id,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $2)
		ON CONFLICT (cleaner_id)
		DO NOTHING
	`

	mock.ExpectExec(regexp.QuoteMeta(query)).
		WithArgs(uint(10), sqlmock.AnyArg()).
		WillReturnError(expectedErr)

	err := repository.EnsureCleaner(context.Background(), 10)

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected database error, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestSQLRepository_Refresh_Success(t *testing.T) {
	repository, mock, cleanup := newMockRepository(t)
	defer cleanup()

	ensureQuery := `
		INSERT INTO cleaner_reputation (
			cleaner_id,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $2)
		ON CONFLICT (cleaner_id)
		DO NOTHING
	`

	mock.ExpectExec(regexp.QuoteMeta(ensureQuery)).
		WithArgs(uint(10), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec(`WITH review_stats AS`).
		WithArgs(
			uint(10),
			sqlmock.AnyArg(),
		).
		WillReturnResult(
			sqlmock.NewResult(0, 1),
		)

	err := repository.Refresh(context.Background(), 10)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestSQLRepository_Refresh_EnsureCleanerError(t *testing.T) {
	repository, mock, cleanup := newMockRepository(t)
	defer cleanup()

	expectedErr := errors.New("ensure cleaner failed")

	ensureQuery := `
		INSERT INTO cleaner_reputation (
			cleaner_id,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $2)
		ON CONFLICT (cleaner_id)
		DO NOTHING
	`

	mock.ExpectExec(regexp.QuoteMeta(ensureQuery)).
		WithArgs(uint(10), sqlmock.AnyArg()).
		WillReturnError(expectedErr)

	err := repository.Refresh(context.Background(), 10)

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected ensure cleaner error, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestSQLRepository_GetByCleanerID_Success(t *testing.T) {
	repository, mock, cleanup := newMockRepository(t)
	defer cleanup()

	now := time.Now()

	query := `
		SELECT
			cleaner_id,
			average_rating,
			total_reviews,
			completed_jobs,
			repeat_clients,
			would_hire_again_count,
			recommendation_percentage,
			total_bookings,
			cleaner_cancellations,
			eligible_response_messages,
			responded_messages,
			average_response_minutes,
			badge,
			created_at,
			updated_at
		FROM cleaner_reputation
		WHERE cleaner_id = $1
		LIMIT 1
		`

	rows := sqlmock.NewRows([]string{
		"cleaner_id",
		"average_rating",
		"total_reviews",
		"completed_jobs",
		"repeat_clients",
		"would_hire_again_count",
		"recommendation_percentage",
		"total_bookings",
		"cleaner_cancellations",
		"eligible_response_messages",
		"responded_messages",
		"average_response_minutes",
		"badge",
		"created_at",
		"updated_at",
	}).AddRow(
		10,
		4.85,
		20,
		25,
		4,
		18,
		90,
		28,
		2,
		30,
		27,
		12,
		"Top Rated",
		now,
		now,
	)

	mock.ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs(uint(10)).
		WillReturnRows(rows)

	result, err := repository.GetByCleanerID(context.Background(), 10)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result == nil {
		t.Fatal("expected reputation, got nil")
	}

	if result.CleanerID != 10 {
		t.Fatalf("expected cleaner ID 10, got %d", result.CleanerID)
	}

	if result.AverageRating != 4.85 {
		t.Fatalf(
			"expected average rating 4.85, got %.2f",
			result.AverageRating,
		)
	}

	if result.Badge != "Top Rated" {
		t.Fatalf("expected Top Rated badge, got %q", result.Badge)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestSQLRepository_GetByCleanerID_NotFound(t *testing.T) {
	repository, mock, cleanup := newMockRepository(t)
	defer cleanup()

	query := `
		SELECT
			cleaner_id,
			average_rating,
			total_reviews,
			completed_jobs,
			repeat_clients,
			would_hire_again_count,
			recommendation_percentage,
			total_bookings,
			cleaner_cancellations,
			eligible_response_messages,
			responded_messages,
			average_response_minutes,
			badge,
			created_at,
			updated_at
		FROM cleaner_reputation
		WHERE cleaner_id = $1
		LIMIT 1
			`

	mock.ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs(uint(99)).
		WillReturnError(sql.ErrNoRows)

	result, err := repository.GetByCleanerID(
		context.Background(),
		99,
	)

	if !errors.Is(err, ErrReputationNotFound) {
		t.Fatalf("expected ErrReputationNotFound, got %v", err)
	}

	if result != nil {
		t.Fatalf("expected nil result, got %+v", result)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestSQLRepository_GetByCleanerID_DatabaseError(t *testing.T) {
	repository, mock, cleanup := newMockRepository(t)
	defer cleanup()

	expectedErr := errors.New("database query failed")

	query := `
	SELECT
			cleaner_id,
			average_rating,
			total_reviews,
			completed_jobs,
			repeat_clients,
			would_hire_again_count,
			recommendation_percentage,
			total_bookings,
			cleaner_cancellations,
			eligible_response_messages,
			responded_messages,
			average_response_minutes,
			badge,
			created_at,
			updated_at
		FROM cleaner_reputation
		WHERE cleaner_id = $1
		LIMIT 1
	`

	mock.ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs(uint(10)).
		WillReturnError(expectedErr)

	result, err := repository.GetByCleanerID(
		context.Background(),
		10,
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected database error, got %v", err)
	}

	if result != nil {
		t.Fatalf("expected nil result, got %+v", result)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestSQLRepository_UpdateBadge_Success(t *testing.T) {
	repository, mock, cleanup := newMockRepository(t)
	defer cleanup()

	query := `
		UPDATE cleaner_reputation
		SET
			badge = $1,
			updated_at = $2
		WHERE cleaner_id = $3
	`

	mock.ExpectExec(regexp.QuoteMeta(query)).
		WithArgs("Elite Cleaner", sqlmock.AnyArg(), uint(10)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repository.UpdateBadge(
		context.Background(),
		10,
		"Elite Cleaner",
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestSQLRepository_UpdateBadge_DatabaseError(t *testing.T) {
	repository, mock, cleanup := newMockRepository(t)
	defer cleanup()

	expectedErr := errors.New("badge update failed")

	query := `
		UPDATE cleaner_reputation
		SET
			badge = $1,
			updated_at = $2
		WHERE cleaner_id = $3
	`

	mock.ExpectExec(regexp.QuoteMeta(query)).
		WithArgs("Top Rated", sqlmock.AnyArg(), uint(10)).
		WillReturnError(expectedErr)

	err := repository.UpdateBadge(
		context.Background(),
		10,
		"Top Rated",
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected update error, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}
