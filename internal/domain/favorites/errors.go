package favorites

import "errors"

var (
	ErrInvalidInput = errors.New("invalid input")
	ErrAlreadySaved = errors.New("cleaner already saved")
	ErrCleanerBlocked = errors.New("cannot favourite blocked cleaner")
	ErrAlreadyExists = errors.New("Already exists")
)
