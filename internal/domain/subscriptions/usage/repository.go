package usage

import "context"

type Repository interface {
	CreateIfNotExists(ctx context.Context, userID uint) error
	GetByUserID(ctx context.Context, userID uint) (*UserUsage, error)
	GetUserSubscriptionAccess(ctx context.Context, userID uint) (*SubscriptionAccess, error)
	IncrementApplicationCount(ctx context.Context, userID uint) error
	IncrementJobPostCount(ctx context.Context, userID uint) error
	SetMonetisationEnabled(ctx context.Context, userID uint, enabled bool) error
}