package cleanerreports

import "context"

type Repository interface {
	Create(ctx context.Context, report *CleanerReport) error
	ListByClientID(ctx context.Context, clientID uint) ([]CleanerReport, error)
}