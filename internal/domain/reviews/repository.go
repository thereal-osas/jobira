package reviews

import "context"

type Repository interface {
	Create(ctx context.Context, review *Review) error
	ListByCleanerID(ctx context.Context, cleanerID uint) ([]Review, error)
	ListByClientID(ctx context.Context, clientID uint) ([]Review, error)
	GetJobClientID(ctx context.Context, jobID uint) (uint, error)
	GetAcceptedCleanerID(ctx context.Context, jobID uint) (uint, error)
	GetJobStatus(ctx context.Context, jobID uint) (string, error)
}
