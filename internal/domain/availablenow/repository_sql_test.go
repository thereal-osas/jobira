package availablenow

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lib/pq"
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

func TestSQLRepository_Upsert_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	now := time.Now().UTC()

	availability := &CleanerAvailableNow{
		CleanerID:         10,
		IsAvailable:       true,
		AvailableFrom:     now,
		AvailableUntil:    now.Add(2 * time.Hour),
		Location:          "Stratford",
		TravelRadiusMiles: 8,
		JobTypes: []string{
			"domestic",
			"airbnb",
		},
	}

	mock.ExpectQuery(
		regexp.QuoteMeta(
			"INSERT INTO cleaner_available_now",
		),
	).
		WithArgs(
			availability.CleanerID,
			availability.IsAvailable,
			availability.AvailableFrom,
			availability.AvailableUntil,
			availability.Location,
			availability.TravelRadiusMiles,
			pq.Array(availability.JobTypes),
			sqlmock.AnyArg(),
		).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{
					"id",
					"created_at",
					"updated_at",
				},
			).AddRow(
				20,
				now,
				now,
			),
		)

	err = repo.Upsert(
		context.Background(),
		availability,
	)
	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}

	if availability.ID != 20 {
		t.Fatalf(
			"expected ID 20, got %d",
			availability.ID,
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf(
			"unmet expectations: %v",
			err,
		)
	}
}

func TestSQLRepository_Upsert_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	expectedErr := errors.New(
		"upsert failed",
	)

	now := time.Now().UTC()

	availability := &CleanerAvailableNow{
		CleanerID:         10,
		IsAvailable:       true,
		AvailableFrom:     now,
		AvailableUntil:    now.Add(time.Hour),
		Location:          "Bow",
		TravelRadiusMiles: 5,
		JobTypes:          []string{"domestic"},
	}

	mock.ExpectQuery(
		regexp.QuoteMeta(
			"INSERT INTO cleaner_available_now",
		),
	).
		WillReturnError(
			expectedErr,
		)

	err = repo.Upsert(
		context.Background(),
		availability,
	)

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

func TestSQLRepository_GetByCleanerID_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	now := time.Now().UTC()

	mock.ExpectQuery(
		regexp.QuoteMeta(
			"SELECT",
		),
	).
		WithArgs(uint(10)).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{
					"id",
					"cleaner_id",
					"is_available",
					"available_from",
					"available_until",
					"location",
					"travel_radius_miles",
					"job_types",
					"created_at",
					"updated_at",
				},
			).AddRow(
				20,
				10,
				true,
				now,
				now.Add(2*time.Hour),
				"Stratford",
				8,
				pq.Array([]string{
					"domestic",
					"airbnb",
				}),
				now,
				now,
			),
		)

	result, err :=
		repo.GetByCleanerID(
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
		t.Fatal("expected availability")
	}

	if result.CleanerID != 10 {
		t.Fatalf(
			"expected cleaner ID 10, got %d",
			result.CleanerID,
		)
	}

	if !result.IsAvailable {
		t.Fatal(
			"expected cleaner to be available",
		)
	}
}

func TestSQLRepository_GetByCleanerID_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	mock.ExpectQuery(
		regexp.QuoteMeta("SELECT"),
	).
		WithArgs(uint(10)).
		WillReturnError(
			sql.ErrNoRows,
		)

	result, err :=
		repo.GetByCleanerID(
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
		ErrAvailableNowNotFound,
	) {
		t.Fatalf(
			"expected ErrAvailableNowNotFound, got %v",
			err,
		)
	}
}

func TestSQLRepository_Disable_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	mock.ExpectExec(
		regexp.QuoteMeta(
			"UPDATE cleaner_available_now",
		),
	).
		WithArgs(
			sqlmock.AnyArg(),
			uint(10),
		).
		WillReturnResult(
			sqlmock.NewResult(
				0,
				1,
			),
		)

	err = repo.Disable(
		context.Background(),
		10,
	)
	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}
}

func TestSQLRepository_Disable_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	mock.ExpectExec(
		regexp.QuoteMeta(
			"UPDATE cleaner_available_now",
		),
	).
		WithArgs(
			sqlmock.AnyArg(),
			uint(10),
		).
		WillReturnResult(
			sqlmock.NewResult(
				0,
				0,
			),
		)

	err = repo.Disable(
		context.Background(),
		10,
	)

	if !errors.Is(
		err,
		ErrAvailableNowNotFound,
	) {
		t.Fatalf(
			"expected ErrAvailableNowNotFound, got %v",
			err,
		)
	}
}

func TestSQLRepository_ListAvailableCleaners_Success(
	t *testing.T,
) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	now := time.Now().UTC()

	search := AvailableNowSearchRequest{
		Location:      "Stratford",
		JobType:       "domestic",
		MinimumRating: 4.5,
		VerifiedOnly:  false,
		DBSRequired:   false,
		Limit:         20,
		Offset:        0,
	}

	mock.ExpectQuery(
		regexp.QuoteMeta(
			"SELECT",
		),
	).
		WithArgs(
			now,
			search.Location,
			search.JobType,
			search.MinimumRating,
			search.VerifiedOnly,
			search.DBSRequired,
			search.Limit,
			search.Offset,
		).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{
					"cleaner_id",
					"full_name",
					"location",
					"travel_radius_miles",
					"available_until",
					"job_types",
					"average_rating",
					"total_reviews",
					"reliability_score",
					"badge",
					"is_verified",
					"dbs_verified",
					"average_response_minutes",
				},
			).AddRow(
				10,
				"Sarah Cleaner",
				"Stratford",
				8,
				now.Add(2*time.Hour),
				pq.Array([]string{
					"domestic",
					"airbnb",
				}),
				4.9,
				25,
				94,
				"Top Rated",
				true,
				true,
				12,
			),
		)

	results, err :=
		repo.ListAvailableCleaners(
			context.Background(),
			search,
			now,
		)
	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}

	if len(results) != 1 {
		t.Fatalf(
			"expected 1 cleaner, got %d",
			len(results),
		)
	}

	if results[0].CleanerID != 10 {
		t.Fatalf(
			"expected cleaner ID 10, got %d",
			results[0].CleanerID,
		)
	}

	if !results[0].DBSVerified {
		t.Fatal(
			"expected DBS verified cleaner",
		)
	}
}
