package jobs

import "context"

type Repository interface {
	Create(ctx context.Context, job *Job) error
	GetByID(ctx context.Context, id uint) (*Job, error)
	List(ctx context.Context) ([]Job, error)
	ListByClientID(ctx context.Context, clientID uint) ([]Job, error)
	Update(ctx context.Context, job *Job) error
	Delete(ctx context.Context, id uint) error
	Search(ctx context.Context, req SearchJobRequest) ([]Job, error)
}
