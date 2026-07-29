package subscriptionaccess

import "context"

type Repository interface {
	GetCleanerAccessStatus(ctx context.Context, userID uint) (*CleanerAccessStatus, error)
	GetClientAccessStatus(ctx context.Context, userID uint) (*ClientAccessStatus, error)

	EnsureDailyAccess(ctx context.Context, userID uint) error
	IsLaunchGraceActive(ctx context.Context) (bool, error)

	IncrementApplicationsToday(ctx context.Context, userID uint) error
	IncrementJobsPostToday(ctx context.Context, userID uint) error
}
