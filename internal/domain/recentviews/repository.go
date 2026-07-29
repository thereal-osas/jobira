package recentviews

import "context"

type Repository interface {
	RecordView(ctx context.Context, clientID uint, cleanerID uint) error
	ListByClientID(ctx context.Context, clientID uint) ([]RecentlyViewedCleaner, error)
}