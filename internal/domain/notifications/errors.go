package notifications

import "errors"

var (
	ErrInvalidInput         = errors.New("invalid input")
	ErrNotificationNotFound = errors.New("notification not found")
	ErrForbidden            = errors.New("forbidden")
)
