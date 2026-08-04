package preferredcleaners

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
		regexp.QuoteMeta(`
			INSERT INTO preferred_cleaners (
				client_id,
				cleaner_id
			)
			VALUES ($1, $2)
			RETURNING id, created_at
		`),
	).
		WithArgs(
			uint(5),
			uint(8),
		).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"id",
				"created_at",
			}).AddRow(
				12,
				now,
			),
		)

	preferred := &PreferredCleaners{
		ClientID:  5,
		CleanerID: 8,
	}

	err = repo.Create(
		context.Background(),
		preferred,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if preferred.ID != 12 {
		t.Fatalf(
			"expected ID 12, got %d",
			preferred.ID,
		)
	}

	if preferred.CreatedAt.IsZero() {
		t.Fatal("expected created_at")
	}
}

func TestSQLRepository_Create_DatabaseError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)
	expectedErr := errors.New("create failed")

	mock.ExpectQuery(
		regexp.QuoteMeta(
			"INSERT INTO preferred_cleaners",
		),
	).
		WillReturnError(expectedErr)

	err = repo.Create(
		context.Background(),
		&PreferredCleaners{},
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected %v, got %v",
			expectedErr,
			err,
		)
	}
}

func TestSQLRepository_Exists_True(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	mock.ExpectQuery(
		regexp.QuoteMeta(`
			SELECT EXISTS (
				SELECT 1
				FROM preferred_cleaners
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

	exists, err := repo.Exists(
		context.Background(),
		5,
		8,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !exists {
		t.Fatal("expected exists to be true")
	}
}

func TestSQLRepository_Exists_False(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	mock.ExpectQuery(
		regexp.QuoteMeta(
			"SELECT EXISTS",
		),
	).
		WithArgs(
			uint(5),
			uint(8),
		).
		WillReturnRows(
			sqlmock.NewRows([]string{"exists"}).
				AddRow(false),
		)

	exists, err := repo.Exists(
		context.Background(),
		5,
		8,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if exists {
		t.Fatal("expected exists to be false")
	}
}

func TestSQLRepository_Exists_DatabaseError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)
	expectedErr := errors.New("exists check failed")

	mock.ExpectQuery(
		regexp.QuoteMeta(
			"SELECT EXISTS",
		),
	).
		WithArgs(
			uint(5),
			uint(8),
		).
		WillReturnError(expectedErr)

	exists, err := repo.Exists(
		context.Background(),
		5,
		8,
	)

	if exists {
		t.Fatal("expected exists to be false")
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
			DELETE FROM preferred_cleaners
			WHERE client_id = $1
			AND cleaner_id = $2
		`),
	).
		WithArgs(
			uint(5),
			uint(8),
		).
		WillReturnResult(
			sqlmock.NewResult(0, 1),
		)

	err = repo.Delete(
		context.Background(),
		5,
		8,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSQLRepository_Delete_DatabaseError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)
	expectedErr := errors.New("delete failed")

	mock.ExpectExec(
		regexp.QuoteMeta(
			"DELETE FROM preferred_cleaners",
		),
	).
		WithArgs(
			uint(5),
			uint(8),
		).
		WillReturnError(expectedErr)

	err = repo.Delete(
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
		"created_at",
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
				created_at
			FROM preferred_cleaners
			WHERE client_id = $1
			ORDER BY created_at DESC
		`),
	).
		WithArgs(uint(5)).
		WillReturnRows(rows)

	cleaners, err := repo.ListByClientID(
		context.Background(),
		5,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(cleaners) != 2 {
		t.Fatalf(
			"expected 2 cleaners, got %d",
			len(cleaners),
		)
	}

	if cleaners[0].ID != 1 {
		t.Fatalf(
			"expected first ID 1, got %d",
			cleaners[0].ID,
		)
	}

	if cleaners[1].CleanerID != 9 {
		t.Fatalf(
			"expected cleaner ID 9, got %d",
			cleaners[1].CleanerID,
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
			"FROM preferred_cleaners",
		),
	).
		WithArgs(uint(5)).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"id",
				"client_id",
				"cleaner_id",
				"created_at",
			}),
		)

	cleaners, err := repo.ListByClientID(
		context.Background(),
		5,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(cleaners) != 0 {
		t.Fatalf(
			"expected no cleaners, got %d",
			len(cleaners),
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
			"FROM preferred_cleaners",
		),
	).
		WithArgs(uint(5)).
		WillReturnError(expectedErr)

	cleaners, err := repo.ListByClientID(
		context.Background(),
		5,
	)

	if cleaners != nil {
		t.Fatalf(
			"expected nil cleaners, got %+v",
			cleaners,
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
		"created_at",
	}).AddRow(
		"invalid-id",
		5,
		8,
		now,
	)

	mock.ExpectQuery(
		regexp.QuoteMeta(
			"FROM preferred_cleaners",
		),
	).
		WithArgs(uint(5)).
		WillReturnRows(rows)

	cleaners, err := repo.ListByClientID(
		context.Background(),
		5,
	)

	if cleaners != nil {
		t.Fatalf(
			"expected nil cleaners, got %+v",
			cleaners,
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
		"created_at",
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
			"FROM preferred_cleaners",
		),
	).
		WithArgs(uint(5)).
		WillReturnRows(rows)

	cleaners, err := repo.ListByClientID(
		context.Background(),
		5,
	)

	if cleaners != nil {
		t.Fatalf(
			"expected nil cleaners, got %+v",
			cleaners,
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

