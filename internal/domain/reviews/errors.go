package reviews

import "errors"

var (
	ErrInvalidInput       = errors.New("invalid input")
	ErrReviewNotFound     = errors.New("review not found")
	ErrReviewAlreadyExist = errors.New("review already exists for this job")
	ErrForbidden          = errors.New("forbidden")
	ErrJobNotCompleted    = errors.New("job is not completed")
)
