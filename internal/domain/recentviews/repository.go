package recentviews

import (
	"context"
	"time"
)

type Repository interface {
	RecordView(ctx context.Context, clientID uint, cleanerID uint) error
	ListByClientID(ctx context.Context, clientID uint) ([]RecentlyViewedCleaner, error)
	CountByCleanerID(ctx context.Context, cleanerID uint) (int, error)
	CountUniqueViewersByCleanerID(ctx context.Context, cleanerID uint) (int, error)
	CountByCleanerIDSince(ctx context.Context, cleanerID uint, since time.Time) (int, error)
	CountByCleanerIDBetween(ctx context.Context, cleanerID uint, from time.Time, to time.Time) (int, error)
}
