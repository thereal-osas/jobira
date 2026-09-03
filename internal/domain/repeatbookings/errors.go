package repeatbookings

import "errors"

var (
	ErrInvalidInput = errors.New("invalid input")
	ErrForbidden    = errors.New("forbidden")
	ErrCleanerUnavailable = errors.New("cleaner unavailable for request time")
)
