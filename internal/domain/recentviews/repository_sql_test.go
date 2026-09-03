package recentviews

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestNewSQLRepository(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf(
			"failed to create sqlmock database: %v",
			err,
		)
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

func TestSQLRepository_RecordView_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	mock.ExpectExec(
		regexp.QuoteMeta(`
			INSERT INTO recently_viewed_cleaners (
				client_id,
				cleaner_id,
				viewed_at
			)
				VALUES ($1, $2, $3)
		`),
	).
		WithArgs(
			uint(5),
			uint(8),
			sqlmock.AnyArg(),
		).
		WillReturnResult(
			sqlmock.NewResult(1, 1),
		)

	err = repo.RecordView(
		context.Background(),
		5,
		8,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf(
			"unmet SQL expectations: %v",
			err,
		)
	}
}

func TestSQLRepository_RecordView_DatabaseError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)
	expectedErr := errors.New("record view failed")

	mock.ExpectExec(
		regexp.QuoteMeta(
			"INSERT INTO recently_viewed_cleaners",
		),
	).
		WithArgs(
			uint(5),
			uint(8),
			sqlmock.AnyArg(),
		).
		WillReturnError(expectedErr)

	err = repo.RecordView(
		context.Background(),
		5,
		8,
	)

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
	now := time.Now()

	rows := sqlmock.NewRows([]string{
		"id",
		"client_id",
		"cleaner_id",
		"viewed_at",
	}).
		AddRow(
			1,
			5,
			8,
			now,
		).
		AddRow(
			2,
			5,
			9,
			now.Add(-time.Hour),
		)

	mock.ExpectQuery(
		regexp.QuoteMeta(`
			SELECT
				id,
				client_id,
				cleaner_id,
				viewed_at
			FROM recently_viewed_cleaners
			WHERE client_id = $1
			ORDER BY viewed_at DESC
			LIMIT 20
		`),
	).
		WithArgs(uint(5)).
		WillReturnRows(rows)

	views, err := repo.ListByClientID(
		context.Background(),
		5,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(views) != 2 {
		t.Fatalf(
			"expected 2 recent views, got %d",
			len(views),
		)
	}

	if views[0].ID != 1 {
		t.Fatalf(
			"expected first ID 1, got %d",
			views[0].ID,
		)
	}

	if views[0].ClientID != 5 {
		t.Fatalf(
			"expected first client ID 5, got %d",
			views[0].ClientID,
		)
	}

	if views[0].CleanerID != 8 {
		t.Fatalf(
			"expected first cleaner ID 8, got %d",
			views[0].CleanerID,
		)
	}

	if views[1].CleanerID != 9 {
		t.Fatalf(
			"expected second cleaner ID 9, got %d",
			views[1].CleanerID,
		)
	}

	if views[0].ViewedAt.IsZero() {
		t.Fatal("expected first viewed_at value")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf(
			"unmet SQL expectations: %v",
			err,
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
		regexp.QuoteMeta(
			"FROM recently_viewed_cleaners",
		),
	).
		WithArgs(uint(5)).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"id",
				"client_id",
				"cleaner_id",
				"viewed_at",
			}),
		)

	views, err := repo.ListByClientID(
		context.Background(),
		5,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(views) != 0 {
		t.Fatalf(
			"expected no recent views, got %d",
			len(views),
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
	expectedErr := errors.New("query failed")

	mock.ExpectQuery(
		regexp.QuoteMeta(
			"FROM recently_viewed_cleaners",
		),
	).
		WithArgs(uint(5)).
		WillReturnError(expectedErr)

	views, err := repo.ListByClientID(
		context.Background(),
		5,
	)

	if views != nil {
		t.Fatalf(
			"expected nil views, got %+v",
			views,
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
		"client_id",
		"cleaner_id",
		"viewed_at",
	}).AddRow(
		"invalid-id",
		5,
		8,
		now,
	)

	mock.ExpectQuery(
		regexp.QuoteMeta(
			"FROM recently_viewed_cleaners",
		),
	).
		WithArgs(uint(5)).
		WillReturnRows(rows)

	views, err := repo.ListByClientID(
		context.Background(),
		5,
	)

	if views != nil {
		t.Fatalf(
			"expected nil views, got %+v",
			views,
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
		"client_id",
		"cleaner_id",
		"viewed_at",
	}).
		AddRow(
			1,
			5,
			8,
			now,
		).
		RowError(0, expectedErr)

	mock.ExpectQuery(
		regexp.QuoteMeta(
			"FROM recently_viewed_cleaners",
		),
	).
		WithArgs(uint(5)).
		WillReturnRows(rows)

	views, err := repo.ListByClientID(
		context.Background(),
		5,
	)

	if views != nil {
		t.Fatalf(
			"expected nil views, got %+v",
			views,
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


func TestSQLRepository_CountByCleanerID_Success(
	t *testing.T,
) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	mock.ExpectQuery(
		regexp.QuoteMeta(`
			SELECT COUNT(*)
			FROM recently_viewed_cleaners
			WHERE cleaner_id = $1
		`),
	).
		WithArgs(uint(8)).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{"count"},
			).AddRow(25),
		)

	count, err := repo.CountByCleanerID(
		context.Background(),
		8,
	)
	if err != nil {
		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}

	if count != 25 {
		t.Fatalf(
			"expected 25 views, got %d",
			count,
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf(
			"unmet SQL expectations: %v",
			err,
		)
	}
}

func TestSQLRepository_CountByCleanerID_Error(
	t *testing.T,
) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	expectedErr := errors.New(
		"count views failed",
	)

	mock.ExpectQuery(
		regexp.QuoteMeta(`
			SELECT COUNT(*)
			FROM recently_viewed_cleaners
			WHERE cleaner_id = $1
		`),
	).
		WithArgs(uint(8)).
		WillReturnError(expectedErr)

	count, err := repo.CountByCleanerID(
		context.Background(),
		8,
	)

	if count != 0 {
		t.Fatalf(
			"expected 0, got %d",
			count,
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

func TestSQLRepository_CountUniqueViewersByCleanerID_Success(
	t *testing.T,
) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	mock.ExpectQuery(
		regexp.QuoteMeta(`
			SELECT COUNT(DISTINCT client_id)
			FROM recently_viewed_cleaners
			WHERE cleaner_id = $1
		`),
	).
		WithArgs(uint(8)).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{"count"},
			).AddRow(14),
		)

	count, err :=
		repo.CountUniqueViewersByCleanerID(
			context.Background(),
			8,
		)

	if err != nil {
		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}

	if count != 14 {
		t.Fatalf(
			"expected 14 unique viewers, got %d",
			count,
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf(
			"unmet SQL expectations: %v",
			err,
		)
	}
}

func TestSQLRepository_CountUniqueViewersByCleanerID_Error(
	t *testing.T,
) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	expectedErr := errors.New(
		"unique viewers failed",
	)

	mock.ExpectQuery(
		regexp.QuoteMeta(`
			SELECT COUNT(DISTINCT client_id)
			FROM recently_viewed_cleaners
			WHERE cleaner_id = $1
		`),
	).
		WithArgs(uint(8)).
		WillReturnError(expectedErr)

	count, err :=
		repo.CountUniqueViewersByCleanerID(
			context.Background(),
			8,
		)

	if count != 0 {
		t.Fatalf(
			"expected 0, got %d",
			count,
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

func TestSQLRepository_CountByCleanerIDSince_Success(
	t *testing.T,
) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	since := time.Date(
		2026,
		time.August,
		24,
		0,
		0,
		0,
		0,
		time.UTC,
	)

	mock.ExpectQuery(
		regexp.QuoteMeta(`
			SELECT COUNT(*)
			FROM recently_viewed_cleaners
			WHERE cleaner_id = $1
			  AND viewed_at >= $2
		`),
	).
		WithArgs(
			uint(8),
			since,
		).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{"count"},
			).AddRow(18),
		)

	count, err := repo.CountByCleanerIDSince(
		context.Background(),
		8,
		since,
	)
	if err != nil {
		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}

	if count != 18 {
		t.Fatalf(
			"expected 18 views, got %d",
			count,
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf(
			"unmet SQL expectations: %v",
			err,
		)
	}
}

func TestSQLRepository_CountByCleanerIDSince_Error(
	t *testing.T,
) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	since := time.Now()

	expectedErr := errors.New(
		"count since failed",
	)

	mock.ExpectQuery(
		regexp.QuoteMeta(`
			SELECT COUNT(*)
			FROM recently_viewed_cleaners
			WHERE cleaner_id = $1
			  AND viewed_at >= $2
		`),
	).
		WithArgs(
			uint(8),
			since,
		).
		WillReturnError(expectedErr)

	count, err := repo.CountByCleanerIDSince(
		context.Background(),
		8,
		since,
	)

	if count != 0 {
		t.Fatalf(
			"expected 0, got %d",
			count,
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

func TestSQLRepository_CountByCleanerIDBetween_Success(
	t *testing.T,
) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	from := time.Date(
		2026,
		time.August,
		17,
		0,
		0,
		0,
		0,
		time.UTC,
	)

	to := time.Date(
		2026,
		time.August,
		24,
		0,
		0,
		0,
		0,
		time.UTC,
	)

	mock.ExpectQuery(
		regexp.QuoteMeta(`
			SELECT COUNT(*)
			FROM recently_viewed_cleaners
			WHERE cleaner_id = $1
			  AND viewed_at >= $2
			  AND viewed_at < $3
		`),
	).
		WithArgs(
			uint(8),
			from,
			to,
		).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{"count"},
			).AddRow(14),
		)

	count, err := repo.CountByCleanerIDBetween(
		context.Background(),
		8,
		from,
		to,
	)
	if err != nil {
		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}

	if count != 14 {
		t.Fatalf(
			"expected 14 views, got %d",
			count,
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf(
			"unmet SQL expectations: %v",
			err,
		)
	}
}

func TestSQLRepository_CountByCleanerIDBetween_Error(
	t *testing.T,
) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	from := time.Now().AddDate(
		0,
		0,
		-7,
	)
	to := time.Now()

	expectedErr := errors.New(
		"count between failed",
	)

	mock.ExpectQuery(
		regexp.QuoteMeta(`
			SELECT COUNT(*)
			FROM recently_viewed_cleaners
			WHERE cleaner_id = $1
			  AND viewed_at >= $2
			  AND viewed_at < $3
		`),
	).
		WithArgs(
			uint(8),
			from,
			to,
		).
		WillReturnError(expectedErr)

	count, err := repo.CountByCleanerIDBetween(
		context.Background(),
		8,
		from,
		to,
	)

	if count != 0 {
		t.Fatalf(
			"expected 0, got %d",
			count,
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