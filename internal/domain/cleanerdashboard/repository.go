package cleanerdashboard

import "context"

type Repository interface {
	CountFavourites(ctx context.Context, cleanerID uint) (int, error)
	CountPreferredClients(ctx context.Context, cleanerID uint) (int, error)
	ListUpcomingBookings(ctx context.Context, cleanerID uint, limit int) ([]UpcomingBooking, error)
}
