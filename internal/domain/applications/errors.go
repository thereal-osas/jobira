package applications

import "errors"

var (
	ErrInvalidInput             = errors.New("invalid input")
	ErrApplicationNotFound      = errors.New("applicatoin not found")
	ErrAlreadyApplied           = errors.New("you have already applied to this job")
	ErrCannotApplyToOwnJob      = errors.New("you cannot apply to your own job")
	ErrForbidden                = errors.New("forbidden")
	ErrInvalidApplicationStatus = errors.New("invalid application status")
)
