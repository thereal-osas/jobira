package reviews

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

func TestSQLRepository_Create_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)
	now := time.Now()

	mock.ExpectQuery(
		regexp.QuoteMeta("INSERT INTO reviews"),
	).
		WithArgs(
			uint(20),
			uint(12),
			uint(8),
			uint(5),
			5,
			"Excellent cleaner",
			sqlmock.AnyArg(),
		).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"id",
				"created_at",
				"updated_at",
			}).AddRow(
				30,
				now,
				now,
			),
		)

	review := &Review{
		BookingID: 20,
		JobID:     12,
		CleanerID: 8,
		ClientID:  5,
		Rating:    5,
		Comment:   "Excellent cleaner",
	}

	err = repo.Create(
		context.Background(),
		review,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if review.ID != 30 {
		t.Fatalf(
			"expected review ID 30, got %d",
			review.ID,
		)
	}

	if review.CreatedAt.IsZero() {
		t.Fatal("expected created_at")
	}

	if review.UpdatedAt.IsZero() {
		t.Fatal("expected updated_at")
	}
}

func TestSQLRepository_Create_Duplicate(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	mock.ExpectQuery(
		regexp.QuoteMeta("INSERT INTO reviews"),
	).
		WillReturnError(
			errors.New(
				"duplicate key value violates unique constraint",
			),
		)

	err = repo.Create(
		context.Background(),
		&Review{},
	)

	if !errors.Is(err, ErrReviewAlreadyExist) {
		t.Fatalf(
			"expected ErrReviewAlreadyExist, got %v",
			err,
		)
	}
}

func TestSQLRepository_Create_DatabaseError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)
	expectedErr := errors.New("create review failed")

	mock.ExpectQuery(
		regexp.QuoteMeta("INSERT INTO reviews"),
	).
		WillReturnError(expectedErr)

	err = repo.Create(
		context.Background(),
		&Review{},
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected %v, got %v",
			expectedErr,
			err,
		)
	}
}

func reviewRows() *sqlmock.Rows {
	now := time.Now()

	return sqlmock.NewRows([]string{
		"id",
		"booking_id",
		"cleaner_id",
		"client_id",
		"job_id",
		"rating",
		"comment",
		"created_at",
		"updated_at",
	}).AddRow(
		30,
		20,
		8,
		5,
		12,
		5,
		"Excellent cleaner",
		now,
		now,
	)
}

func TestSQLRepository_ListByCleanerID_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	mock.ExpectQuery(
		regexp.QuoteMeta("FROM reviews"),
	).
		WithArgs(uint(8)).
		WillReturnRows(reviewRows())

	reviews, err := repo.ListByCleanerID(
		context.Background(),
		8,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(reviews) != 1 {
		t.Fatalf(
			"expected 1 review, got %d",
			len(reviews),
		)
	}

	if reviews[0].BookingID != 20 {
		t.Fatalf(
			"expected booking ID 20, got %d",
			reviews[0].BookingID,
		)
	}

	if reviews[0].CleanerID != 8 {
		t.Fatalf(
			"expected cleaner ID 8, got %d",
			reviews[0].CleanerID,
		)
	}
}

func TestSQLRepository_ListByCleanerID_Empty(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	mock.ExpectQuery(
		regexp.QuoteMeta("FROM reviews"),
	).
		WithArgs(uint(8)).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"id",
				"booking_id",
				"cleaner_id",
				"client_id",
				"job_id",
				"rating",
				"comment",
				"created_at",
				"updated_at",
			}),
		)

	reviews, err := repo.ListByCleanerID(
		context.Background(),
		8,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(reviews) != 0 {
		t.Fatalf(
			"expected no reviews, got %d",
			len(reviews),
		)
	}
}

func TestSQLRepository_ListByCleanerID_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)
	expectedErr := errors.New("query cleaner reviews failed")

	mock.ExpectQuery(
		regexp.QuoteMeta("FROM reviews"),
	).
		WithArgs(uint(8)).
		WillReturnError(expectedErr)

	reviews, err := repo.ListByCleanerID(
		context.Background(),
		8,
	)

	if reviews != nil {
		t.Fatalf(
			"expected nil reviews, got %+v",
			reviews,
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

func TestSQLRepository_ListByCleanerID_ScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)
	now := time.Now()

	rows := sqlmock.NewRows([]string{
		"id",
		"booking_id",
		"cleaner_id",
		"client_id",
		"job_id",
		"rating",
		"comment",
		"created_at",
		"updated_at",
	}).AddRow(
		"invalid-id",
		20,
		8,
		5,
		12,
		5,
		"Excellent",
		now,
		now,
	)

	mock.ExpectQuery(
		regexp.QuoteMeta("FROM reviews"),
	).
		WithArgs(uint(8)).
		WillReturnRows(rows)

	reviews, err := repo.ListByCleanerID(
		context.Background(),
		8,
	)

	if reviews != nil {
		t.Fatalf(
			"expected nil reviews, got %+v",
			reviews,
		)
	}

	if err == nil {
		t.Fatal("expected scan error")
	}
}

func TestSQLRepository_ListByCleanerID_RowsError(t *testing.T) {
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
		"booking_id",
		"cleaner_id",
		"client_id",
		"job_id",
		"rating",
		"comment",
		"created_at",
		"updated_at",
	}).
		AddRow(
			30,
			20,
			8,
			5,
			12,
			5,
			"Excellent",
			now,
			now,
		).
		RowError(0, expectedErr)

	mock.ExpectQuery(
		regexp.QuoteMeta("FROM reviews"),
	).
		WithArgs(uint(8)).
		WillReturnRows(rows)

	reviews, err := repo.ListByCleanerID(
		context.Background(),
		8,
	)

	if reviews != nil {
		t.Fatalf(
			"expected nil reviews, got %+v",
			reviews,
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

func TestSQLRepository_ListByClientID_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	mock.ExpectQuery(
		regexp.QuoteMeta("FROM reviews"),
	).
		WithArgs(uint(5)).
		WillReturnRows(reviewRows())

	reviews, err := repo.ListByClientID(
		context.Background(),
		5,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(reviews) != 1 {
		t.Fatalf(
			"expected 1 review, got %d",
			len(reviews),
		)
	}

	if reviews[0].BookingID != 20 {
		t.Fatalf(
			"expected booking ID 20, got %d",
			reviews[0].BookingID,
		)
	}

	if reviews[0].ClientID != 5 {
		t.Fatalf(
			"expected client ID 5, got %d",
			reviews[0].ClientID,
		)
	}
}

func TestSQLRepository_ListByClientID_Empty(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	mock.ExpectQuery(
		regexp.QuoteMeta("FROM reviews"),
	).
		WithArgs(uint(5)).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"id",
				"booking_id",
				"cleaner_id",
				"client_id",
				"job_id",
				"rating",
				"comment",
				"created_at",
				"updated_at",
			}),
		)

	reviews, err := repo.ListByClientID(
		context.Background(),
		5,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(reviews) != 0 {
		t.Fatalf(
			"expected no reviews, got %d",
			len(reviews),
		)
	}
}

func TestSQLRepository_ListByClientID_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)
	expectedErr := errors.New("query client reviews failed")

	mock.ExpectQuery(
		regexp.QuoteMeta("FROM reviews"),
	).
		WithArgs(uint(5)).
		WillReturnError(expectedErr)

	reviews, err := repo.ListByClientID(
		context.Background(),
		5,
	)

	if reviews != nil {
		t.Fatalf(
			"expected nil reviews, got %+v",
			reviews,
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

func TestSQLRepository_ListByClientID_ScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)
	now := time.Now()

	rows := sqlmock.NewRows([]string{
		"id",
		"booking_id",
		"cleaner_id",
		"client_id",
		"job_id",
		"rating",
		"comment",
		"created_at",
		"updated_at",
	}).AddRow(
		"invalid-id",
		20,
		8,
		5,
		12,
		5,
		"Excellent",
		now,
		now,
	)

	mock.ExpectQuery(
		regexp.QuoteMeta("FROM reviews"),
	).
		WithArgs(uint(5)).
		WillReturnRows(rows)

	reviews, err := repo.ListByClientID(
		context.Background(),
		5,
	)

	if reviews != nil {
		t.Fatalf(
			"expected nil reviews, got %+v",
			reviews,
		)
	}

	if err == nil {
		t.Fatal("expected scan error")
	}
}

func TestSQLRepository_ListByClientID_RowsError(t *testing.T) {
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
		"booking_id",
		"cleaner_id",
		"client_id",
		"job_id",
		"rating",
		"comment",
		"created_at",
		"updated_at",
	}).
		AddRow(
			30,
			20,
			8,
			5,
			12,
			5,
			"Excellent",
			now,
			now,
		).
		RowError(0, expectedErr)

	mock.ExpectQuery(
		regexp.QuoteMeta("FROM reviews"),
	).
		WithArgs(uint(5)).
		WillReturnRows(rows)

	reviews, err := repo.ListByClientID(
		context.Background(),
		5,
	)

	if reviews != nil {
		t.Fatalf(
			"expected nil reviews, got %+v",
			reviews,
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

func TestSQLRepository_GetJobClientID_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	mock.ExpectQuery(
		regexp.QuoteMeta("SELECT client_id"),
	).
		WithArgs(uint(12)).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"client_id",
			}).AddRow(5),
		)

	clientID, err := repo.GetJobClientID(
		context.Background(),
		12,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if clientID != 5 {
		t.Fatalf(
			"expected client ID 5, got %d",
			clientID,
		)
	}
}

func TestSQLRepository_GetJobClientID_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	mock.ExpectQuery(
		regexp.QuoteMeta("SELECT client_id"),
	).
		WithArgs(uint(12)).
		WillReturnError(sql.ErrNoRows)

	clientID, err := repo.GetJobClientID(
		context.Background(),
		12,
	)

	if clientID != 0 {
		t.Fatalf(
			"expected client ID 0, got %d",
			clientID,
		)
	}

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf(
			"expected ErrInvalidInput, got %v",
			err,
		)
	}
}

func TestSQLRepository_GetJobClientID_DatabaseError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)
	expectedErr := errors.New("get job client failed")

	mock.ExpectQuery(
		regexp.QuoteMeta("SELECT client_id"),
	).
		WithArgs(uint(12)).
		WillReturnError(expectedErr)

	clientID, err := repo.GetJobClientID(
		context.Background(),
		12,
	)

	if clientID != 0 {
		t.Fatalf(
			"expected client ID 0, got %d",
			clientID,
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

func TestSQLRepository_GetAcceptedCleanerID_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	mock.ExpectQuery(
		regexp.QuoteMeta("SELECT cleaner_id"),
	).
		WithArgs(uint(12)).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"cleaner_id",
			}).AddRow(8),
		)

	cleanerID, err := repo.GetAcceptedCleanerID(
		context.Background(),
		12,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cleanerID != 8 {
		t.Fatalf(
			"expected cleaner ID 8, got %d",
			cleanerID,
		)
	}
}

func TestSQLRepository_GetAcceptedCleanerID_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	mock.ExpectQuery(
		regexp.QuoteMeta("SELECT cleaner_id"),
	).
		WithArgs(uint(12)).
		WillReturnError(sql.ErrNoRows)

	cleanerID, err := repo.GetAcceptedCleanerID(
		context.Background(),
		12,
	)

	if cleanerID != 0 {
		t.Fatalf(
			"expected cleaner ID 0, got %d",
			cleanerID,
		)
	}

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf(
			"expected ErrInvalidInput, got %v",
			err,
		)
	}
}

func TestSQLRepository_GetAcceptedCleanerID_DatabaseError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)
	expectedErr := errors.New("get accepted cleaner failed")

	mock.ExpectQuery(
		regexp.QuoteMeta("SELECT cleaner_id"),
	).
		WithArgs(uint(12)).
		WillReturnError(expectedErr)

	cleanerID, err := repo.GetAcceptedCleanerID(
		context.Background(),
		12,
	)

	if cleanerID != 0 {
		t.Fatalf(
			"expected cleaner ID 0, got %d",
			cleanerID,
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

func TestSQLRepository_GetJobStatus_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	mock.ExpectQuery(
		regexp.QuoteMeta("SELECT status"),
	).
		WithArgs(uint(12)).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"status",
			}).AddRow("completed"),
		)

	status, err := repo.GetJobStatus(
		context.Background(),
		12,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if status != "completed" {
		t.Fatalf(
			"expected completed, got %q",
			status,
		)
	}
}

func TestSQLRepository_GetJobStatus_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	mock.ExpectQuery(
		regexp.QuoteMeta("SELECT status"),
	).
		WithArgs(uint(12)).
		WillReturnError(sql.ErrNoRows)

	status, err := repo.GetJobStatus(
		context.Background(),
		12,
	)

	if status != "" {
		t.Fatalf(
			"expected empty status, got %q",
			status,
		)
	}

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf(
			"expected ErrInvalidInput, got %v",
			err,
		)
	}
}

func TestSQLRepository_GetJobStatus_DatabaseError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)
	expectedErr := errors.New("get job status failed")

	mock.ExpectQuery(
		regexp.QuoteMeta("SELECT status"),
	).
		WithArgs(uint(12)).
		WillReturnError(expectedErr)

	status, err := repo.GetJobStatus(
		context.Background(),
		12,
	)

	if status != "" {
		t.Fatalf(
			"expected empty status, got %q",
			status,
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
