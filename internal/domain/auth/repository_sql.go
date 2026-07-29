package auth

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"
)

type SQLRepository struct {
	db *sql.DB
}

func NewSQLRepository(db *sql.DB) *SQLRepository {
	return &SQLRepository{db: db}
}

func (r *SQLRepository) CreateUser(ctx context.Context, user *User) error {
	query := `
		INSERT INTO users (full_name, email, password_hash, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
		`

	now := time.Now()

	user.CreatedAt = now
	user.UpdatedAt = now
	err := r.db.QueryRowContext(
		ctx,
		query,
		user.FullName,
		user.Email,
		user.PasswordHash,
		user.Role,
		user.CreatedAt,
		user.UpdatedAt,
	).Scan(&user.ID)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "duplicate key") {
			return ErrEmailAlreadyInUse
		}
		return err
	}

	return nil
}

func (r *SQLRepository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	query := `
	    SELECT id, full_name, email, password_hash, role, created_at, updated_at
	    FROM users
	    WHERE email = $1
	    LIMIT 1
	`

	var user User

	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.FullName,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	return &user, nil
}
