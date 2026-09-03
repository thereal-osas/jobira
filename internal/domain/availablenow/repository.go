package availablenow

import (
	"context"
	"time"
)

type Repository interface {
	Upsert(ctx context.Context, availability *CleanerAvailableNow) error

	GetByCleanerID(ctx context.Context, cleanerID uint) (*CleanerAvailableNow, error)

	Disable(ctx context.Context, cleanerID uint) error

	ListAvailableCleaners(ctx context.Context, search AvailableNowSearchRequest, now time.Time) ([]AvailableCleaner, error)
}
