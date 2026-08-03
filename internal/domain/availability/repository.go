package availability

import "context"

type Repository interface {
	Create(ctx context.Context, availability *CleanerAvailability) error
	GetByID(ctx context.Context, id uint) (*CleanerAvailability, error)
	ListByCleanerID(ctx context.Context, cleanerID uint) ([]CleanerAvailability, error)
	Update(ctx context.Context, availability *CleanerAvailability) error
	Delete(ctx context.Context, id uint) error
	HasConflict(ctx context.Context, cleanrID uint, availableDate string, startTime string, endTime string, excludeID uint) (bool, error)
	CreateBlock(ctx context.Context, block *AvailabilityBlock) error
	ListBlocksByCleanerID(ctx context.Context, cleanerID uint) ([]AvailabilityBlock, error)
	DeleteBlock(ctx context.Context, blockID uint, cleanerID uint) error
}
