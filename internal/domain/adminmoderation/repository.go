package adminmoderation

import "context"

type Repository interface {
	ListReports(ctx context.Context) ([]CleanerReportAdminView, error)
	ListOpenReports(ctx context.Context) ([]CleanerReportAdminView, error)
	UpdateReportStatus(ctx context.Context, reportID uint, adminID uint, status string, adminNote string) error
	ListBlockCleaners(ctx context.Context) ([]BlockedCleanerAdminView, error)
}