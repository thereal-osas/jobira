package savedjobs

import "errors"

var (
	ErrInvalidInput     = errors.New("invalid input")
	ErrSavedJobNotFound = errors.New("saved job not found")
	ErrAlreadySved      = errors.New("job already saved")
)
