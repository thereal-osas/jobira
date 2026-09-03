package repeatbookings

import (
	"context"
	"time"
)

type Repository interface {
	CreateJob(ctx context.Context, clientID uint, req CreateRepeatBookingRequest) (uint, error)
	CreateInvitation(ctx context.Context, jobID uint, clientID uint, cleanerID uint, message string) error
	CreateRepeatBooking(ctx context.Context, booking *RepeatBooking) error
	ListByClientID(ctx context.Context, clientID uint) ([]RepeatBooking, error)
	GetOriginBooking(ctx context.Context, bookingID uint) (*BookingSnapshot, error)
	CreateBookAgainRequest(ctx context.Context, request *RepeatBookingRequest) error
	CreateRepeatBookings(ctx context.Context, booking *BookingSnapshot, scheduledAt time.Time, scheduledEndAt time.Time) (uint, error)
	CreateBookAgainTransaction(ctx context.Context, input BookAgainTransaction) (*RepeatBookingRequest, error)
}
