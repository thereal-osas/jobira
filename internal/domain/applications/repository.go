package applications

import "context"

type Repository interface {
	Create(ctx context.Context, application *Application) error
	GetByID(ctx context.Context, id uint) (*Application, error)
	ListByJobID(ctx context.Context, jobID uint) ([]Application, error)
	ListByCleanerID(ctx context.Context, cleanerID uint) ([]Application, error)
	UpdateStatus(ctx context.Context, id uint, status string) error
	GetJobClientID(ctx context.Context, jobID uint) (uint, error)
	GetJobTitle(ctx context.Context, jobID uint) (string, error)
	GetUserEmail(ctx context.Context, userID uint) (string, error)
	GetCleanerEmailByID(ctx context.Context, userID uint) (string, error)
}