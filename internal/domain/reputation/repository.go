package reputation

import "context"

type Repository interface {
	EnsureCleaner(ctx context.Context, cleanerID uint) error
	Refresh(ctx context.Context, cleanerID uint) error
	GetByCleanerID(ctx context.Context, cleanerID uint) (*CleanerReputation, error)
	UpdateBadge(ctx context.Context, cleanerID uint, badge string) error
}
