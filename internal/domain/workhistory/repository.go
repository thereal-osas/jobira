package workhistory

import "context"

type Repository interface {
	ListByCleanerID(ctx context.Context, cleanerID uint) ([]WorkHistoryEntry, error)
	ListByClientID(ctx context.Context, clientID uint) ([]WorkHistoryEntry, error)
	GetByBookingID(ctx context.Context, bookingID uint, userID uint) (*WorkHistoryDetail, error)
	UserCanAccessBooking(ctx context.Context, bookingID uint, userID uint) (bool, error)
}
