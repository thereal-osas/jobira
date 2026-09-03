package wokrproof

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
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

func TestSQLRepository_GetBooking_Success(t *testing.T) {
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
					"job_id",
					"client_id",
					"cleaner_id",
					"status",
				},
			).AddRow(
				10,
				7,
				5,
				8,
				"in_progress",
			),
		)

	result, err := repo.GetBooking(
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
		t.Fatal("expected booking")
	}

	if result.ID != 10 {
		t.Fatalf(
			"expected booking ID 10, got %d",
			result.ID,
		)
	}

	if result.CleanerID != 8 {
		t.Fatalf(
			"expected cleaner ID 8, got %d",
			result.CleanerID,
		)
	}
}

func TestSQLRepository_GetBooking_NotFound(t *testing.T) {
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

	result, err := repo.GetBooking(
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
		ErrBookingNotFound,
	) {
		t.Fatalf(
			"expected ErrBookingNotFound, got %v",
			err,
		)
	}
}

func TestSQLRepository_Create_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	now := time.Now().UTC()

	proof := &WorkProof{
		BookingID: 10,
		CleanerID: 8,
		ProofType: ProofTypeBefore,
		PhotoURL:  "https://example.com/before.jpg",
		Caption:   "Kitchen before cleaning",
	}

	mock.ExpectQuery(
		regexp.QuoteMeta(
			"INSERT INTO work_proofs",
		),
	).
		WithArgs(
			uint(10),
			uint(8),
			ProofTypeBefore,
			"https://example.com/before.jpg",
			"Kitchen before cleaning",
		).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{
					"id",
					"created_at",
				},
			).AddRow(
				30,
				now,
			),
		)

	err = repo.Create(
		context.Background(),
		proof,
	)
	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}

	if proof.ID != 30 {
		t.Fatalf(
			"expected proof ID 30, got %d",
			proof.ID,
		)
	}
}

func TestSQLRepository_GetByID_Success(t *testing.T) {
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
		WithArgs(uint(30)).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{
					"id",
					"booking_id",
					"cleaner_id",
					"proof_type",
					"photo_url",
					"caption",
					"created_at",
				},
			).AddRow(
				30,
				10,
				8,
				"before",
				"https://example.com/before.jpg",
				"Kitchen before",
				now,
			),
		)

	result, err := repo.GetByID(
		context.Background(),
		30,
	)
	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}

	if result.ProofType != ProofTypeBefore {
		t.Fatalf(
			"expected before proof, got %q",
			result.ProofType,
		)
	}
}

func TestSQLRepository_GetByID_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	mock.ExpectQuery(
		regexp.QuoteMeta("SELECT"),
	).
		WithArgs(uint(30)).
		WillReturnError(
			sql.ErrNoRows,
		)

	result, err := repo.GetByID(
		context.Background(),
		30,
	)

	if result != nil {
		t.Fatalf(
			"expected nil result, got %+v",
			result,
		)
	}

	if !errors.Is(
		err,
		ErrProofNotFound,
	) {
		t.Fatalf(
			"expected ErrProofNotFound, got %v",
			err,
		)
	}
}

func TestSQLRepository_ListByBookingID_Success(
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
		WithArgs(uint(10)).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{
					"id",
					"booking_id",
					"cleaner_id",
					"proof_type",
					"photo_url",
					"caption",
					"created_at",
				},
			).
				AddRow(
					1,
					10,
					8,
					"before",
					"https://example.com/before.jpg",
					"Before",
					now,
				).
				AddRow(
					2,
					10,
					8,
					"after",
					"https://example.com/after.jpg",
					"After",
					now.Add(time.Hour),
				),
		)

	results, err := repo.ListByBookingID(
		context.Background(),
		10,
	)
	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}

	if len(results) != 2 {
		t.Fatalf(
			"expected 2 proofs, got %d",
			len(results),
		)
	}
}

func TestSQLRepository_CountByBookingAndType_Success(
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
		WithArgs(
			uint(10),
			ProofTypeBefore,
		).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{"count"},
			).AddRow(
				4,
			),
		)

	count, err :=
		repo.CountByBookingAndType(
			context.Background(),
			10,
			ProofTypeBefore,
		)
	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}

	if count != 4 {
		t.Fatalf(
			"expected count 4, got %d",
			count,
		)
	}
}

func TestSQLRepository_Delete_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	mock.ExpectExec(
		regexp.QuoteMeta(
			"DELETE FROM work_proofs",
		),
	).
		WithArgs(
			uint(30),
			uint(8),
		).
		WillReturnResult(
			sqlmock.NewResult(
				0,
				1,
			),
		)

	err = repo.Delete(
		context.Background(),
		30,
		8,
	)
	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}
}

func TestSQLRepository_Delete_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	mock.ExpectExec(
		regexp.QuoteMeta(
			"DELETE FROM work_proofs",
		),
	).
		WithArgs(
			uint(30),
			uint(8),
		).
		WillReturnResult(
			sqlmock.NewResult(
				0,
				0,
			),
		)

	err = repo.Delete(
		context.Background(),
		30,
		8,
	)

	if !errors.Is(
		err,
		ErrProofNotFound,
	) {
		t.Fatalf(
			"expected ErrProofNotFound, got %v",
			err,
		)
	}
}

func TestSQLRepository_GetVerifiedWorkSummary_Success(
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
		WithArgs(uint(8)).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{
					"cleaner_id",
					"verified_jobs",
					"before_photos",
					"after_photos",
					"total_photos",
				},
			).AddRow(
				8,
				5,
				8,
				9,
				17,
			),
		)

	result, err :=
		repo.GetVerifiedWorkSummary(
			context.Background(),
			8,
		)
	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}

	if result.VerifiedJobs != 5 {
		t.Fatalf(
			"expected 5 verified jobs, got %d",
			result.VerifiedJobs,
		)
	}

	if result.TotalPhotos != 17 {
		t.Fatalf(
			"expected 17 photos, got %d",
			result.TotalPhotos,
		)
	}
}

