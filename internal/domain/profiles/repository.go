package profiles

import "context"

type Repository interface {
	Create(ctx context.Context, profile *CleanerProfile) error
	GetByUserID(ctx context.Context, userID uint) (*CleanerProfile, error)
	Update(ctx context.Context, profile *CleanerProfile) error
	UpdateVerificationStatus(ctx context.Context, userID uint, status string, verified bool) error
	Search(ctx context.Context, req SearchProfilesRequest) ([]CleanerProfile, error)
	IncrementJobsCompleted(ctx context.Context, userID uint) error
	IncrementJobsCancelled(ctx context.Context, userID uint) error
	RecalculateReputation(ctx context.Context, userID uint) error
	GetCompletedJobs(ctx context.Context, userID uint) ([]JobHistoryItem, error)
	GetCancelledJobs(ctx context.Context, userID uint) ([]JobHistoryItem, error)
	GetFullHistory(ctx context.Context, userID uint) ([]JobHistoryItem, error)
}