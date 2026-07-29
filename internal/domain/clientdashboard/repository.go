package clientdashboard

import "context"

type Repository interface {
	ListActiveJobs(ctx context.Context, clientID uint, limit int) ([]ActiveJob, error)
	ListUpcomingBookings(ctx context.Context, clientID uint, limit int) ([]ClientBooking, error)

	CountApplicationsReceived(ctx context.Context, clientID uint) (int, error)
	CountCompletedBookings(ctx context.Context, clientID uint) (int, error)
	CountFavouriteCleaners(ctx context.Context, clientID uint) (int, error)
	CountPreferredCleaners(ctx context.Context, clientID uint) (int, error)
	CountRepeatBookings(ctx context.Context, clientID uint) (int, error)
}
