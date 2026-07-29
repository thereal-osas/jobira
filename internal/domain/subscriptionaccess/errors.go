package subscriptionaccess

import "errors"

var (
	ErrInvalidInput            = errors.New("invalid input")
	ErrApplicationLimitReached = errors.New("free application limit reached")
	ErrJobPostLimitReached     = errors.New("free job post limit reached")
	ErrSubscriptionRequired    = errors.New("subsctipion required")
	ErrDailyApplicationLimit   = errors.New("daily application limit reached")
	ErrAccessNotFound          = errors.New("access record not found")
	ErrDailyJobPostLimit       = errors.New("daily job post limit reached")
	
)
