package availability

import "errors"

var (
	ErrInvalidInput         = errors.New("invalid input")
	ErrAvailabilityNotFound = errors.New("availability not found")
	ErrAvailabilityExists   = errors.New("availability already exists")
	ErrAvailabilityConflict = errors.New("availability conflict")
	ErrInvalidStatus        = errors.New("invalid availability status")
	ErrForbidden            = errors.New("forbidden")
)
