package availability

import "context"

type Repository interface {
	Create(ctx context.Context, availability *CleanerAvailability) error
	GetByID(ctx context.Context, id uint) (*CleanerAvailability, error)
	ListByCleanerID(ctx context.Context, cleanerID uint) ([]CleanerAvailability, error)
	Update(ctx context.Context, availability *CleanerAvailability) error
	Delete(ctx context.Context, id uint) error
}