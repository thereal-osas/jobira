package users

import "context"

type Repository interface {
	GetByID(ctx context.Context, id uint) (*User, error)
}