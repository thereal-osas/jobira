package usage

import "errors"

var (
	ErrInvalidInput		= errors.New("invalid input")
	ErrUsageNotFound 	= errors.New("usage record not found")
	ErrApplicationLimitReached	= errors.New("free application limit reached")
	ErrJobPostLimitReached  = errors.New("free job post limit reached")
)

