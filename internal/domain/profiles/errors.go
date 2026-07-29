package profiles

import "errors"

var (
	ErrInvalidInput             = errors.New("invalid input")
	ErrProfileNotFound          = errors.New("profile not found")
	ErrProfileAlreadyExists     = errors.New("profile already exists")
	ErrForbidden                = errors.New("forbidden")
	ErrInvalidVerificationState = errors.New("invalid verification status ")
	ErrInvalidAvailabilityStatus = errors.New("invalid availability status")
)
