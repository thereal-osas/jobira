package preferredcleaners

import "errors"

var (
	ErrInvalidInput = errors.New("invalid input")
	ErrAlreadyPreferred = errors.New("cleaner already preferred")
	ErrCleanerBlocked = errors.New("cannot prefer blocked cleaner")
	ErrAlreadyExists = errors.New("Already exists")
)