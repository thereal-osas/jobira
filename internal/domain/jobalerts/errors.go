package jobalerts

import "errors"

var (
	ErrInvalidInput  = errors.New("invalid input")
	ErrAlertNotFound = errors.New("job alert not found")
	ErrForbidden     = errors.New("forbidden")
)
