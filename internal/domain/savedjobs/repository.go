package savedjobs

import "context"

type Repository interface {
	Save(ctx context.Context, savedJob *SavedJob) error
	ListByUserID(ctx context.Context, userID uint) ([]SavedJob, error)
	Delete(ctx context.Context, userID uint, jobID uint) error
}
