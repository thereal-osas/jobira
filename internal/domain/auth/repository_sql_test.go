package auth

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
		t.Fatalf("failed to create sql mock: %v", err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	if repo == nil {
		t.Fatal("expected repository")
	}

	if repo.db != db {
		t.Fatal("expected database to be assigned")
	}
}

func TestSQLRepository_CreateUser_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sql mock: %v", err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	user := &User{
		FullName:     "John Smith",
		Email:        "john@test.com",
		PasswordHash: "hashed-password",
		Role:         "user",
	}

	query := regexp.QuoteMeta(`
		INSERT INTO users (full_name, email, password_hash, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
		`)

	mock.ExpectQuery(query).
		WithArgs(
			user.FullName,
			user.Email,
			user.PasswordHash,
			user.Role,
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
		).
		WillReturnRows(
			sqlmock.NewRows([]string{"id"}).
				AddRow(uint(7)),
		)

	err = repo.CreateUser(context.Background(), user)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if user.ID != 7 {
		t.Fatalf("expected user ID 7, got %d", user.ID)
	}

	if user.CreatedAt.IsZero() {
		t.Fatal("expected CreatedAt to be assigned")
	}

	if user.UpdatedAt.IsZero() {
		t.Fatal("expected UpdatedAt to be assigned")
	}

	if !user.CreatedAt.Equal(user.UpdatedAt) {
		t.Fatal("expected CreatedAt and UpdatedAt to match")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestSQLRepository_CreateUser_DuplicateEmail(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sql mock: %v", err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	user := &User{
		FullName:     "John Smith",
		Email:        "john@test.com",
		PasswordHash: "hashed-password",
		Role:         "user",
	}

	query := regexp.QuoteMeta(`
		INSERT INTO users (full_name, email, password_hash, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
		`)

	mock.ExpectQuery(query).
		WithArgs(
			user.FullName,
			user.Email,
			user.PasswordHash,
			user.Role,
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
		).
		WillReturnError(
			errors.New(`pq: duplicate key value violates unique constraint "users_email_key"`),
		)

	err = repo.CreateUser(context.Background(), user)

	if !errors.Is(err, ErrEmailAlreadyInUse) {
		t.Fatalf("expected ErrEmailAlreadyInUse, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestSQLRepository_CreateUser_DatabaseError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sql mock: %v", err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	expected := errors.New("database unavailable")

	user := &User{
		FullName:     "John Smith",
		Email:        "john@test.com",
		PasswordHash: "hashed-password",
		Role:         "user",
	}

	query := regexp.QuoteMeta(`
		INSERT INTO users (full_name, email, password_hash, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
		`)

	mock.ExpectQuery(query).
		WithArgs(
			user.FullName,
			user.Email,
			user.PasswordHash,
			user.Role,
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
		).
		WillReturnError(expected)

	err = repo.CreateUser(context.Background(), user)

	if !errors.Is(err, expected) {
		t.Fatalf("expected %v, got %v", expected, err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestSQLRepository_GetUserByEmail_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sql mock: %v", err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	createdAt := time.Now().Add(-time.Hour)
	updatedAt := time.Now()

	query := regexp.QuoteMeta(`
	    SELECT id, full_name, email, password_hash, role, created_at, updated_at
	    FROM users
	    WHERE email = $1
	    LIMIT 1
	`)

	mock.ExpectQuery(query).
		WithArgs("john@test.com").
		WillReturnRows(
			sqlmock.NewRows([]string{
				"id",
				"full_name",
				"email",
				"password_hash",
				"role",
				"created_at",
				"updated_at",
			}).AddRow(
				uint(1),
				"John Smith",
				"john@test.com",
				"hashed-password",
				"user",
				createdAt,
				updatedAt,
			),
		)

	user, err := repo.GetUserByEmail(
		context.Background(),
		"john@test.com",
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if user == nil {
		t.Fatal("expected user")
	}

	if user.ID != 1 {
		t.Fatalf("expected user ID 1, got %d", user.ID)
	}

	if user.FullName != "John Smith" {
		t.Fatalf("expected John Smith, got %q", user.FullName)
	}

	if user.Email != "john@test.com" {
		t.Fatalf("expected john@test.com, got %q", user.Email)
	}

	if user.PasswordHash != "hashed-password" {
		t.Fatalf("unexpected password hash %q", user.PasswordHash)
	}

	if user.Role != "user" {
		t.Fatalf("expected role user, got %q", user.Role)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestSQLRepository_GetUserByEmail_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sql mock: %v", err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	query := regexp.QuoteMeta(`
	    SELECT id, full_name, email, password_hash, role, created_at, updated_at
	    FROM users
	    WHERE email = $1
	    LIMIT 1
	`)

	mock.ExpectQuery(query).
		WithArgs("missing@test.com").
		WillReturnError(sql.ErrNoRows)

	user, err := repo.GetUserByEmail(
		context.Background(),
		"missing@test.com",
	)

	if user != nil {
		t.Fatalf("expected nil user, got %+v", user)
	}

	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestSQLRepository_GetUserByEmail_ScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sql mock: %v", err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	query := regexp.QuoteMeta(`
	    SELECT id, full_name, email, password_hash, role, created_at, updated_at
	    FROM users
	    WHERE email = $1
	    LIMIT 1
	`)

	mock.ExpectQuery(query).
		WithArgs("john@test.com").
		WillReturnRows(
			sqlmock.NewRows([]string{
				"id",
				"full_name",
				"email",
				"password_hash",
				"role",
				"created_at",
				"updated_at",
			}).AddRow(
				"invalid-id",
				"John Smith",
				"john@test.com",
				"hashed-password",
				"user",
				time.Now(),
				time.Now(),
			),
		)

	user, err := repo.GetUserByEmail(
		context.Background(),
		"john@test.com",
	)

	if user != nil {
		t.Fatalf("expected nil user, got %+v", user)
	}

	if err == nil {
		t.Fatal("expected scan error")
	}

	if errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected scan error, got ErrInvalidCredentials")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestSQLRepository_GetUserByEmail_DatabaseError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sql mock: %v", err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	expected := errors.New("database unavailable")

	query := regexp.QuoteMeta(`
	    SELECT id, full_name, email, password_hash, role, created_at, updated_at
	    FROM users
	    WHERE email = $1
	    LIMIT 1
	`)

	mock.ExpectQuery(query).
		WithArgs("john@test.com").
		WillReturnError(expected)

	user, err := repo.GetUserByEmail(
		context.Background(),
		"john@test.com",
	)

	if user != nil {
		t.Fatalf("expected nil user, got %+v", user)
	}

	if !errors.Is(err, expected) {
		t.Fatalf("expected %v, got %v", expected, err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}
