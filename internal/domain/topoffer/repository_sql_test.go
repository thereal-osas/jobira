package topoffer

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestSQLRepository_GetJobContext_Success(
	t *testing.T,
) {
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
		WillReturnRows(
			sqlmock.NewRows(
				[]string{
					"id",
					"client_id",
					"job_type",
					"budget",
					"status",
					"location",
				},
			).AddRow(
				10,
				5,
				"domestic",
				100,
				"open",
				"London",
			),
		)

	result, err :=
		repo.GetJobContext(
			context.Background(),
			10,
		)
	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}

	if result.JobID != 10 {
		t.Fatalf(
			"expected job ID 10, got %d",
			result.JobID,
		)
	}

	if result.ClientID != 5 {
		t.Fatalf(
			"expected client ID 5, got %d",
			result.ClientID,
		)
	}
}

func TestSQLRepository_GetJobContext_NotFound(
	t *testing.T,
) {
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
		repo.GetJobContext(
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
		ErrJobNotFound,
	) {
		t.Fatalf(
			"expected ErrJobNotFound, got %v",
			err,
		)
	}
}

func TestSQLRepository_ListOfferCandidates_Success(
	t *testing.T,
) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	now := time.Now().UTC()

	mock.ExpectQuery(
		regexp.QuoteMeta("SELECT"),
	).
		WithArgs(
			uint(10),
			50,
			0,
		).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{
					"id",
					"job_id",
					"cleaner_id",
					"full_name",
					"cover_message",
					"proposed_rate",
					"status",
					"average_rating",
					"total_reviews",
					"completed_jobs",
					"reliability_score",
					"recommendation_percentage",
					"average_response_minutes",
					"badge",
					"is_verified",
					"dbs_verified",
					"created_at",
				},
			).AddRow(
				1,
				10,
				20,
				"Sarah Cleaner",
				"I can do this job.",
				95,
				"pending",
				4.9,
				30,
				50,
				95,
				96,
				12,
				"Top Rated",
				true,
				true,
				now,
			),
		)

	results, err :=
		repo.ListOfferCandidates(
			context.Background(),
			10,
			50,
			0,
		)
	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}

	if len(results) != 1 {
		t.Fatalf(
			"expected 1 candidate, got %d",
			len(results),
		)
	}

	if results[0].ApplicationID != 1 {
		t.Fatalf(
			"expected application ID 1, got %d",
			results[0].ApplicationID,
		)
	}
}

func TestSQLRepository_GetApplicationCandidate_NotFound(
	t *testing.T,
) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	mock.ExpectQuery(
		regexp.QuoteMeta("SELECT"),
	).
		WithArgs(uint(1)).
		WillReturnError(
			sql.ErrNoRows,
		)

	result, err :=
		repo.GetApplicationCandidate(
			context.Background(),
			1,
		)

	if result != nil {
		t.Fatalf(
			"expected nil result, got %+v",
			result,
		)
	}

	if !errors.Is(
		err,
		ErrApplicationNotFound,
	) {
		t.Fatalf(
			"expected ErrApplicationNotFound, got %v",
			err,
		)
	}
}

func TestSQLRepository_CountOffers_Success(
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
			"SELECT COUNT(*)",
		),
	).
		WithArgs(uint(10)).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{"count"},
			).AddRow(
				8,
			),
		)

	count, err :=
		repo.CountOffers(
			context.Background(),
			10,
		)
	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}

	if count != 8 {
		t.Fatalf(
			"expected 8, got %d",
			count,
		)
	}
}
