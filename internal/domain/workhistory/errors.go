package workhistory

import "errors"

var (
	ErrInvalidInput = errors.New("invalid input")

	ErrForbidden = errors.New("forbidden")

	ErrHistoryNotFound = errors.New("work history not found")
)


