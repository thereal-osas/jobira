package availablenow

import "errors"

var (
	ErrInvalidInput = errors.New("invalid input")

	ErrAvailableNowNotFound = errors.New("available now status not found")

	Errforbidden = errors.New("forbidden")
)
