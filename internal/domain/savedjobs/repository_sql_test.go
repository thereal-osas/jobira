package savedjobs

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
		t.Fatal("expected database assigned")
	}
}

func TestSQLRepository_Save_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)
	now := time.Now()

	mock.ExpectQuery(
		regexp.QuoteMeta(`
			INSERT INTO saved_jobs (
				user_id,
				job_id,
				created_at
			)
			VALUES ($1, $2, $3)
			RETURNING id, created_at
		`),
	).
		WithArgs(
			uint(5),
			uint(12),
			sqlmock.AnyArg(),
		).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"id",
				"created_at",
			}).AddRow(
				20,
				now,
			),
		)

	savedJob := &SavedJob{
		UserID: 5,
		JobID:  12,
	}

	err = repo.Save(
		context.Background(),
		savedJob,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if savedJob.ID != 20 {
		t.Fatalf(
			"expected ID 20, got %d",
			savedJob.ID,
		)
	}

	if savedJob.CreatedAt.IsZero() {
		t.Fatal("expected created_at")
	}
}

func TestSQLRepository_Save_Duplicate(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	mock.ExpectQuery(
		regexp.QuoteMeta(
			"INSERT INTO saved_jobs",
		),
	).
		WillReturnError(
			errors.New(
				"duplicate key value violates unique constraint",
			),
		)

	err = repo.Save(
		context.Background(),
		&SavedJob{},
	)

	if !errors.Is(err, ErrAlreadySved) {
		t.Fatalf(
			"expected ErrAlreadySved, got %v",
			err,
		)
	}
}

func TestSQLRepository_Save_DatabaseError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)
	expectedErr := errors.New("save failed")

	mock.ExpectQuery(
		regexp.QuoteMeta(
			"INSERT INTO saved_jobs",
		),
	).
		WillReturnError(expectedErr)

	err = repo.Save(
		context.Background(),
		&SavedJob{},
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected %v, got %v",
			expectedErr,
			err,
		)
	}
}

func TestSQLRepository_ListByUserID_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)
	now := time.Now()

	rows := sqlmock.NewRows([]string{
		"id",
		"user_id",
		"job_id",
		"created_at",
	}).
		AddRow(
			1,
			5,
			12,
			now,
		).
		AddRow(
			2,
			5,
			15,
			now.Add(-time.Hour),
		)

	mock.ExpectQuery(
		regexp.QuoteMeta(`
			SELECT
				id,
				user_id,
				job_id,
				created_at
			FROM saved_jobs
			WHERE user_id = $1
			ORDER BY created_at DESC
		`),
	).
		WithArgs(uint(5)).
		WillReturnRows(rows)

	savedJobs, err := repo.ListByUserID(
		context.Background(),
		5,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(savedJobs) != 2 {
		t.Fatalf(
			"expected 2 saved jobs, got %d",
			len(savedJobs),
		)
	}

	if savedJobs[0].JobID != 12 {
		t.Fatalf(
			"expected first job ID 12, got %d",
			savedJobs[0].JobID,
		)
	}

	if savedJobs[1].JobID != 15 {
		t.Fatalf(
			"expected second job ID 15, got %d",
			savedJobs[1].JobID,
		)
	}
}

func TestSQLRepository_ListByUserID_Empty(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	mock.ExpectQuery(
		regexp.QuoteMeta(
			"FROM saved_jobs",
		),
	).
		WithArgs(uint(5)).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"id",
				"user_id",
				"job_id",
				"created_at",
			}),
		)

	savedJobs, err := repo.ListByUserID(
		context.Background(),
		5,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(savedJobs) != 0 {
		t.Fatalf(
			"expected no saved jobs, got %d",
			len(savedJobs),
		)
	}
}

func TestSQLRepository_ListByUserID_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)
	expectedErr := errors.New("query failed")

	mock.ExpectQuery(
		regexp.QuoteMeta(
			"FROM saved_jobs",
		),
	).
		WithArgs(uint(5)).
		WillReturnError(expectedErr)

	savedJobs, err := repo.ListByUserID(
		context.Background(),
		5,
	)

	if savedJobs != nil {
		t.Fatalf(
			"expected nil saved jobs, got %+v",
			savedJobs,
		)
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected %v, got %v",
			expectedErr,
			err,
		)
	}
}

func TestSQLRepository_ListByUserID_ScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)
	now := time.Now()

	rows := sqlmock.NewRows([]string{
		"id",
		"user_id",
		"job_id",
		"created_at",
	}).AddRow(
		"invalid-id",
		5,
		12,
		now,
	)

	mock.ExpectQuery(
		regexp.QuoteMeta(
			"FROM saved_jobs",
		),
	).
		WithArgs(uint(5)).
		WillReturnRows(rows)

	savedJobs, err := repo.ListByUserID(
		context.Background(),
		5,
	)

	if savedJobs != nil {
		t.Fatalf(
			"expected nil saved jobs, got %+v",
			savedJobs,
		)
	}

	if err == nil {
		t.Fatal("expected scan error")
	}
}

func TestSQLRepository_ListByUserID_RowsError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)
	now := time.Now()
	expectedErr := errors.New("rows failed")

	rows := sqlmock.NewRows([]string{
		"id",
		"user_id",
		"job_id",
		"created_at",
	}).
		AddRow(
			1,
			5,
			12,
			now,
		).
		RowError(0, expectedErr)

	mock.ExpectQuery(
		regexp.QuoteMeta(
			"FROM saved_jobs",
		),
	).
		WithArgs(uint(5)).
		WillReturnRows(rows)

	savedJobs, err := repo.ListByUserID(
		context.Background(),
		5,
	)

	if savedJobs != nil {
		t.Fatalf(
			"expected nil saved jobs, got %+v",
			savedJobs,
		)
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected %v, got %v",
			expectedErr,
			err,
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
		regexp.QuoteMeta(`
			DELETE FROM saved_jobs
			WHERE user_id = $1
			AND job_id = $2
		`),
	).
		WithArgs(
			uint(5),
			uint(12),
		).
		WillReturnResult(
			sqlmock.NewResult(0, 1),
		)

	err = repo.Delete(
		context.Background(),
		5,
		12,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
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
			"DELETE FROM saved_jobs",
		),
	).
		WithArgs(
			uint(5),
			uint(12),
		).
		WillReturnResult(
			sqlmock.NewResult(0, 0),
		)

	err = repo.Delete(
		context.Background(),
		5,
		12,
	)

	if !errors.Is(err, ErrSavedJobNotFound) {
		t.Fatalf(
			"expected ErrSavedJobNotFound, got %v",
			err,
		)
	}
}

func TestSQLRepository_Delete_ExecError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)
	expectedErr := errors.New("delete failed")

	mock.ExpectExec(
		regexp.QuoteMeta(
			"DELETE FROM saved_jobs",
		),
	).
		WithArgs(
			uint(5),
			uint(12),
		).
		WillReturnError(expectedErr)

	err = repo.Delete(
		context.Background(),
		5,
		12,
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected %v, got %v",
			expectedErr,
			err,
		)
	}
}

func TestSQLRepository_Delete_RowsAffectedError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)
	expectedErr := errors.New("rows affected failed")

	mock.ExpectExec(
		regexp.QuoteMeta(
			"DELETE FROM saved_jobs",
		),
	).
		WithArgs(
			uint(5),
			uint(12),
		).
		WillReturnResult(
			sqlmock.NewErrorResult(expectedErr),
		)

	err = repo.Delete(
		context.Background(),
		5,
		12,
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected %v, got %v",
			expectedErr,
			err,
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
	now := time.Now()

	mock.ExpectQuery(
		regexp.QuoteMeta(`
			SELECT
				id,
				user_id,
				job_id,
				created_at
			FROM saved_jobs
			WHERE id = $1
			LIMIT 1
		`),
	).
		WithArgs(uint(20)).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"id",
				"user_id",
				"job_id",
				"created_at",
			}).AddRow(
				20,
				5,
				12,
				now,
			),
		)

	savedJob, err := repo.GetByID(
		context.Background(),
		20,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if savedJob == nil {
		t.Fatal("expected saved job")
	}

	if savedJob.ID != 20 {
		t.Fatalf(
			"expected ID 20, got %d",
			savedJob.ID,
		)
	}

	if savedJob.UserID != 5 {
		t.Fatalf(
			"expected user ID 5, got %d",
			savedJob.UserID,
		)
	}

	if savedJob.JobID != 12 {
		t.Fatalf(
			"expected job ID 12, got %d",
			savedJob.JobID,
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
		regexp.QuoteMeta(
			"FROM saved_jobs",
		),
	).
		WithArgs(uint(20)).
		WillReturnError(sql.ErrNoRows)

	savedJob, err := repo.GetByID(
		context.Background(),
		20,
	)

	if savedJob != nil {
		t.Fatalf(
			"expected nil saved job, got %+v",
			savedJob,
		)
	}

	if !errors.Is(err, ErrSavedJobNotFound) {
		t.Fatalf(
			"expected ErrSavedJobNotFound, got %v",
			err,
		)
	}
}

func TestSQLRepository_GetByID_DatabaseError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)
	expectedErr := errors.New("get saved job failed")

	mock.ExpectQuery(
		regexp.QuoteMeta(
			"FROM saved_jobs",
		),
	).
		WithArgs(uint(20)).
		WillReturnError(expectedErr)

	savedJob, err := repo.GetByID(
		context.Background(),
		20,
	)

	if savedJob != nil {
		t.Fatalf(
			"expected nil saved job, got %+v",
			savedJob,
		)
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected %v, got %v",
			expectedErr,
			err,
		)
	}
}
