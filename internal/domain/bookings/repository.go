package bookings

import "context"

type Repository interface {
	Create(ctx context.Context, booking *Booking) error
	GetByID(ctx context.Context, id uint) (*Booking, error)
	ListByUserID(ctx context.Context, userID uint) ([]Booking, error)
	UpdateStatus(ctx context.Context, bookingID uint, status string) error
	Complete(ctx context.Context, bookingID uint, status string) error
	Cancel(ctx context.Context, bookingID uint, reason string) error
	GetUserEmail(ctx context.Context, userID uint) (string, error)
	Close(ctx context.Context, bookingID uint, rating int, wouldHireAgain bool, comment string) error
	MarkJobCompleted(ctx context.Context, jobID uint) error
	MarkApplicationAccepted(ctx context.Context, applicationID uint) error
	CloseWithTransaction(ctx context.Context, transcation CloseBookingTransaction) error
}
