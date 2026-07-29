package blockedcleaners

import "errors"

var (
	ErrInvalidInput   = errors.New("invalid input")
	ErrAlreadyBlocked = errors.New("cleaner already blocked")
	ErrBlockedNotFound	= errors.New("cleaner was not blocked")
)
