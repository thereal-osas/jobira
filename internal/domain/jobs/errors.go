package jobs

import "errors"

var (
	ErrInvalidInput = errors.New("invalid input")
	ErrJobNotFound  = errors.New("job not found")
	ErrForbidden    = errors.New("forbidden")
)
