package bookings

import "errors"

var (
	ErrInvalidInput    = errors.New("invalid input")
	ErrBookingNotFound = errors.New("booking not found")
	ErrBookingExists   = errors.New("booking already exists")
	ErrForbidden       = errors.New("forbidden")
	ErrInvalidStatus   = errors.New("invalid booking status")
	ErrBookingConflict = errors.New("cleaner already has a booking this time")
)
