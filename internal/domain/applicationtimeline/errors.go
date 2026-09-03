package applicationtimeline

import "errors"

var (
	ErrInvalidInput        = errors.New("invalid input")
	ErrTimelineNotFound    = errors.New("application timeline not found")
	ErrApplicationNotFound = errors.New("application not found")
	ErrForbidden           = errors.New("forbidden")
	ErrInvalidStatus       = errors.New("invalid application timeline status")
)
