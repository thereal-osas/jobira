package analytics

import "context"

type Repository interface {
	TotalUsers(ctx context.Context) (int, error)
	UsersByRole(ctx context.Context, role string) (int, error)
	TotalCompanies(ctx context.Context) (int, error)
	ActiveJobs(ctx context.Context) (int, error)
	CompletedBookings(ctx context.Context) (int, error)
	PendingVerifications(ctx context.Context) (int, error)
	OpenReports(ctx context.Context) (int, error)
	BookingsToday(ctx context.Context) (int, error)
}
