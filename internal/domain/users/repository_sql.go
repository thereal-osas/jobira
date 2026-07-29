package users

import (
	"context"
	"database/sql"
	"errors"
)

var ErrUserNotFound = errors.New("user not found")

type SQLRepository struct {
	db *sql.DB
}

func NewSQLRepository(db *sql.DB) *SQLRepository {
	return &SQLRepository{
		db: db,
	}
}

func (r *SQLRepository) GetByID(ctx context.Context, id uint) (*User, error) {
	query := `
		SELECT id, full_name, email, role, created_at, updated_at
		FROM users
		WHERE id =$1
		LIMIT 1
	`
	var user User
	var dbID int64

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&dbID,
		&user.FullName,
		&user.Email,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,	
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}

		return nil, err
	}

	user.ID = uint(dbID)

	return &user, nil
}