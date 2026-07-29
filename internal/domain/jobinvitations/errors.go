package jobinvitations

import "errors"

var (
	ErrInvalidInput     = errors.New("invalid input")
	ErrInvitationExists = errors.New("cleaner already invited to this job")
	ErrForbidden        = errors.New("forbidden")
)
