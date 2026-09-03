package clientnotes

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
			INSERT INTO client_cleaner_notes (
				client_id,
				cleaner_id,
				note,
				created_at,
				updated_at
			)
				VALUES ($1, $2, $3, $4, $5)
				RETURNING id, created_at, updated_at
		`),
	).
		WithArgs(
			uint(5),
			uint(8),
			"Reliable cleaner",
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
		).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"id",
				"created_at",
				"updated_at",
			}).AddRow(
				12,
				now,
				now,
			),
		)

	note := &ClientCleanerNote{
		ClientID:  5,
		CleanerID: 8,
		Note:      "Reliable cleaner",
	}

	err = repo.Create(
		context.Background(),
		note,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if note.ID != 12 {
		t.Fatalf(
			"expected ID 12, got %d",
			note.ID,
		)
	}

	if note.CreatedAt.IsZero() {
		t.Fatal("expected created_at")
	}

	if note.UpdatedAt.IsZero() {
		t.Fatal("expected updated_at")
	}
}

func TestSQLRepository_Create_DatabaseError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)
	expectedErr := errors.New("create note failed")

	mock.ExpectQuery(
		regexp.QuoteMeta(
			"INSERT INTO client_cleaner_notes",
		),
	).
		WillReturnError(expectedErr)

	err = repo.Create(
		context.Background(),
		&ClientCleanerNote{},
	)

	if !errors.Is(err, expectedErr) {
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
	now := time.Now()

	rows := sqlmock.NewRows([]string{
		"id",
		"client_id",
		"cleaner_id",
		"note",
		"created_at",
		"updated_at",
	}).
		AddRow(
			1,
			5,
			8,
			"Reliable",
			now,
			now,
		).
		AddRow(
			2,
			5,
			8,
			"Good communication",
			now.Add(-time.Hour),
			now.Add(-time.Hour),
		)

	mock.ExpectQuery(
		regexp.QuoteMeta(`
			SELECT
				id,
				client_id,
				cleaner_id,
				note,
				created_at,
				updated_at
			FROM client_cleaner_notes
			WHERE client_id = $1
			AND cleaner_id = $2
			ORDER BY created_at DESC
		`),
	).
		WithArgs(
			uint(5),
			uint(8),
		).
		WillReturnRows(rows)

	notes, err := repo.GetByCleanerID(
		context.Background(),
		5,
		8,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(notes) != 2 {
		t.Fatalf(
			"expected 2 notes, got %d",
			len(notes),
		)
	}

	if notes[0].Note != "Reliable" {
		t.Fatalf(
			"expected first note Reliable, got %q",
			notes[0].Note,
		)
	}

	if notes[1].Note != "Good communication" {
		t.Fatalf(
			"expected second note Good communication, got %q",
			notes[1].Note,
		)
	}
}

func TestSQLRepository_GetByCleanerID_Empty(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	mock.ExpectQuery(
		regexp.QuoteMeta(
			"FROM client_cleaner_notes",
		),
	).
		WithArgs(
			uint(5),
			uint(8),
		).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"id",
				"client_id",
				"cleaner_id",
				"note",
				"created_at",
				"updated_at",
			}),
		)

	notes, err := repo.GetByCleanerID(
		context.Background(),
		5,
		8,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(notes) != 0 {
		t.Fatalf(
			"expected no notes, got %d",
			len(notes),
		)
	}
}

func TestSQLRepository_GetByCleanerID_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)
	expectedErr := errors.New("query failed")

	mock.ExpectQuery(
		regexp.QuoteMeta(
			"FROM client_cleaner_notes",
		),
	).
		WithArgs(
			uint(5),
			uint(8),
		).
		WillReturnError(expectedErr)

	notes, err := repo.GetByCleanerID(
		context.Background(),
		5,
		8,
	)

	if notes != nil {
		t.Fatalf(
			"expected nil notes, got %+v",
			notes,
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

func TestSQLRepository_GetByCleanerID_ScanError(t *testing.T) {
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
		"note",
		"created_at",
		"updated_at",
	}).AddRow(
		"invalid-id",
		5,
		8,
		"Reliable",
		now,
		now,
	)

	mock.ExpectQuery(
		regexp.QuoteMeta(
			"FROM client_cleaner_notes",
		),
	).
		WithArgs(
			uint(5),
			uint(8),
		).
		WillReturnRows(rows)

	notes, err := repo.GetByCleanerID(
		context.Background(),
		5,
		8,
	)

	if notes != nil {
		t.Fatalf(
			"expected nil notes, got %+v",
			notes,
		)
	}

	if err == nil {
		t.Fatal("expected scan error")
	}
}

func TestSQLRepository_GetByCleanerID_RowsError(t *testing.T) {
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
		"note",
		"created_at",
		"updated_at",
	}).
		AddRow(
			1,
			5,
			8,
			"Reliable",
			now,
			now,
		).
		RowError(0, expectedErr)

	mock.ExpectQuery(
		regexp.QuoteMeta(
			"FROM client_cleaner_notes",
		),
	).
		WithArgs(
			uint(5),
			uint(8),
		).
		WillReturnRows(rows)

	notes, err := repo.GetByCleanerID(
		context.Background(),
		5,
		8,
	)

	if notes != nil {
		t.Fatalf(
			"expected nil notes, got %+v",
			notes,
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

func TestSQLRepository_Update_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	mock.ExpectExec(
		regexp.QuoteMeta(`
			UPDATE client_cleaner_notes
			SET
				note = $1,
				updated_at = $2
			WHERE id = $3
			AND client_id = $4
		`),
	).
		WithArgs(
			"Updated note",
			sqlmock.AnyArg(),
			uint(12),
			uint(5),
		).
		WillReturnResult(
			sqlmock.NewResult(0, 1),
		)

	err = repo.Update(
		context.Background(),
		12,
		5,
		"Updated note",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSQLRepository_Update_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	mock.ExpectExec(
		regexp.QuoteMeta(
			"UPDATE client_cleaner_notes",
		),
	).
		WithArgs(
			"Updated note",
			sqlmock.AnyArg(),
			uint(12),
			uint(5),
		).
		WillReturnResult(
			sqlmock.NewResult(0, 0),
		)

	err = repo.Update(
		context.Background(),
		12,
		5,
		"Updated note",
	)

	if !errors.Is(err, ErrNoteNotFound) {
		t.Fatalf(
			"expected ErrNoteNotFound, got %v",
			err,
		)
	}
}

func TestSQLRepository_Update_ExecError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)
	expectedErr := errors.New("update failed")

	mock.ExpectExec(
		regexp.QuoteMeta(
			"UPDATE client_cleaner_notes",
		),
	).
		WithArgs(
			"Updated note",
			sqlmock.AnyArg(),
			uint(12),
			uint(5),
		).
		WillReturnError(expectedErr)

	err = repo.Update(
		context.Background(),
		12,
		5,
		"Updated note",
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected %v, got %v",
			expectedErr,
			err,
		)
	}
}

func TestSQLRepository_Update_RowsAffectedError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)
	expectedErr := errors.New("rows affected failed")

	mock.ExpectExec(
		regexp.QuoteMeta(
			"UPDATE client_cleaner_notes",
		),
	).
		WithArgs(
			"Updated note",
			sqlmock.AnyArg(),
			uint(12),
			uint(5),
		).
		WillReturnResult(
			sqlmock.NewErrorResult(expectedErr),
		)

	err = repo.Update(
		context.Background(),
		12,
		5,
		"Updated note",
	)

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
			DELETE FROM client_cleaner_notes
			WHERE id = $1
			AND client_id = $2
		`),
	).
		WithArgs(
			uint(12),
			uint(5),
		).
		WillReturnResult(
			sqlmock.NewResult(0, 1),
		)

	err = repo.Delete(
		context.Background(),
		12,
		5,
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
			"DELETE FROM client_cleaner_notes",
		),
	).
		WithArgs(
			uint(12),
			uint(5),
		).
		WillReturnResult(
			sqlmock.NewResult(0, 0),
		)

	err = repo.Delete(
		context.Background(),
		12,
		5,
	)

	if !errors.Is(err, ErrNoteNotFound) {
		t.Fatalf(
			"expected ErrNoteNotFound, got %v",
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
			"DELETE FROM client_cleaner_notes",
		),
	).
		WithArgs(
			uint(12),
			uint(5),
		).
		WillReturnError(expectedErr)

	err = repo.Delete(
		context.Background(),
		12,
		5,
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
			"DELETE FROM client_cleaner_notes",
		),
	).
		WithArgs(
			uint(12),
			uint(5),
		).
		WillReturnResult(
			sqlmock.NewErrorResult(expectedErr),
		)

	err = repo.Delete(
		context.Background(),
		12,
		5,
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected %v, got %v",
			expectedErr,
			err,
		)
	}
}
