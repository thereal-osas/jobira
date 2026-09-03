package cleanerprogression

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	reputationdomain "github.com/rodrigueghenda/jobira/internal/domain/reputation"
)

func TestNewSQLRepository(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	if repo == nil {
		t.Fatal("expected repository")
	}

	if repo.db != db {
		t.Fatal("expected repository database to match")
	}
}

func TestSQLRepository_GetCleanerReputation_Success(
	t *testing.T,
) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	now := time.Now().UTC()

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

	mock.ExpectQuery(
		regexp.QuoteMeta(query),
	).
		WithArgs(uint(10)).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{
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
				},
			).AddRow(
				10,
				4.9,
				25,
				40,
				8,
				23,
				96,
				42,
				2,
				30,
				28,
				12,
				"Top Rated",
				now,
				now,
			),
		)

	result, err :=
		repo.GetCleanerReputation(
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
		t.Fatal("expected reputation")
	}

	if result.CleanerID != 10 {
		t.Fatalf(
			"expected cleaner ID 10, got %d",
			result.CleanerID,
		)
	}

	if result.AverageRating != 4.9 {
		t.Fatalf(
			"expected rating 4.9, got %.2f",
			result.AverageRating,
		)
	}

	if result.TotalBookings != 42 {
		t.Fatalf(
			"expected 42 bookings, got %d",
			result.TotalBookings,
		)
	}

	if result.Badge != "Top Rated" {
		t.Fatalf(
			"expected Top Rated, got %q",
			result.Badge,
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf(
			"unmet expectations: %v",
			err,
		)
	}
}

func TestSQLRepository_GetCleanerReputation_NotFound(
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
		WithArgs(uint(10)).
		WillReturnError(
			sql.ErrNoRows,
		)

	result, err :=
		repo.GetCleanerReputation(
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
		reputationdomain.ErrReputationNotFound,
	) {
		t.Fatalf(
			"expected ErrReputationNotFound, got %v",
			err,
		)
	}
}

func TestSQLRepository_GetCleanerReputation_DatabaseError(
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
		WithArgs(uint(10)).
		WillReturnError(
			expectedErr,
		)

	result, err :=
		repo.GetCleanerReputation(
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
