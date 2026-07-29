package reports

import "context"

type Repository interface {
	Create(ctx context.Context, report *Report) error
	GetByID(ctx context.Context, id uint) (*Report, error)
	ListByReporterID(ctx context.Context, reporterID uint) ([]Report, error)
	ListAll(ctx context.Context) ([]Report, error)
	ListOpen(ctx context.Context) ([]Report, error)
	Review(ctx context.Context, reportID uint, status string, adminNotes string, adminID uint) error
}
