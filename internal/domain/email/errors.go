package email

import "errors"

var (
	ErrInvalidInput = errors.New("invalid input")
	ErrSendFailed   = errors.New("email send failed")
)
