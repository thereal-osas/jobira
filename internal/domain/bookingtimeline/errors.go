package bookingtimeline

import "errors"

var (
	ErrInvalidInput    = errors.New("invalid input")
	ErrBookingNotFound = errors.New("booking not found")
	ErrForbidden       = errors.New("forbidden")
)
