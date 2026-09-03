package topoffer

import "errors"

var (
	ErrInvalidInput = errors.New(
		"invalid input",
	)

	ErrJobNotFound = errors.New(
		"job not found",
	)

	ErrApplicationNotFound = errors.New(
		"application not found",
	)

	ErrForbidden = errors.New(
		"forbidden",
	)

	ErrJobClosed = errors.New(
		"job is not open",
	)
)
