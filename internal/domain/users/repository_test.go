package users

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestNewSQLRepository(t *testing.T) {
	db, _, err := sqlmock.New()

	if err != nil {
		t.Fatalf(
			"failed creating mock db: %v",
			err,
		)
	}

	defer db.Close()

	repo := NewSQLRepository(db)

	if repo == nil {
		t.Fatal(
			"expected repository",
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

	rows := sqlmock.NewRows(
		[]string{
			"id",
			"full_name",
			"email",
			"role",
			"created_at",
			"updated_at",
		},
	).
		AddRow(
			1,
			"John Cleaner",
			"john@test.com",
			"cleaner",
			"2026-01-01",
			"2026-01-02",
		)

	mock.ExpectQuery(
		"SELECT id, full_name, email, role, created_at, updated_at",
	).
		WithArgs(1).
		WillReturnRows(rows)

	user, err := repo.GetByID(
		context.Background(),
		1,
	)

	if err != nil {
		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}

	if user.ID != 1 {
		t.Fatalf(
			"expected id 1 got %d",
			user.ID,
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
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
		"SELECT id, full_name, email, role, created_at, updated_at",
	).
		WithArgs(1).
		WillReturnError(sql.ErrNoRows)

	_, err = repo.GetByID(
		context.Background(),
		1,
	)

	if err != ErrUserNotFound {
		t.Fatalf(
			"expected ErrUserNotFound got %v",
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

	mock.ExpectQuery(
		"SELECT id, full_name, email, role, created_at, updated_at",
	).
		WithArgs(1).
		WillReturnError(
			errors.New("database failed"),
		)

	_, err = repo.GetByID(
		context.Background(),
		1,
	)

	if err == nil {
		t.Fatal(
			"expected error",
		)
	}
}
