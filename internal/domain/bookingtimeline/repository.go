package bookingtimeline

import "context"

type Repository interface {
	CreateHistory(ctx context.Context, bookingID uint, changedBy uint, fromStatus string, toStatus string, note string) error
	ListByBookingID(ctx context.Context, bookingID uint) ([]StatusHistory, error)
	GetBookingAccess(ctx context.Context, bookingID uint,) (clientID uint, cleanerID uint, currentStatus string, err error)
}
