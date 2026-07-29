package verifications

import "errors"

var (
	ErrInvalidInput              = errors.New("invalid input")
	ErrVerificationNotFound      = errors.New("verification request not found")
	ErrInvalidVerificationType   = errors.New("invalid verification type")
	ErrForbidden                 = errors.New("forbidden")
	ErrInValidVerificationStatus = errors.New("invalid verification status")
)
