package reputation

import "errors"

var (
	ErrInvalidInput       = errors.New("invalid input")
	ErrReputationNotFound = errors.New("cleaner reputation not found")
)
