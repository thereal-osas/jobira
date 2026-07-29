package jobalerts

import "context"

type Repository interface {
	Create(ctx context.Context, alert *JobAlert) error
	GetByID(ctx context.Context, id uint) (*JobAlert, error)
	ListByUserID(ctx context.Context, userID uint) ([]JobAlert, error)
	Update(ctx context.Context, alert *JobAlert) error
	Delete(ctx context.Context, id uint) error
}
