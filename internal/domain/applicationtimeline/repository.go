package applicationtimeline

import "context"

type Repository interface {
	Create(ctx context.Context, event *Event) error
	ListByApplicationID(ctx context.Context, applicationID uint) ([]Event, error)
	GetCurrentStatus(ctx context.Context, applicationID uint) (string, error)
	GetApplicationClientID(ctx context.Context, applicationID uint) (uint, error)
	GetApplicationCleanerID(ctx context.Context, applicationID uint) (uint, error)
}
