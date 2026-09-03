package companyaccounts

import "errors"

var (
	ErrInvalidInput     = errors.New("invalid input")
	ErrCompanyNotFound  = errors.New("company not found")
	ErrMemberExists     = errors.New("user already belongs to company")
	ErrForbidden        = errors.New("forbidden")
	ErrMemberNotFound   = errors.New("company member not found")
	ErrSeatLimitReached = errors.New("company seat limit reached")
	ErrSeatPlanNotFound = errors.New("company cleaner seat plan not found")
)
