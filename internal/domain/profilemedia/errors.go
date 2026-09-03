package profilemedia

import "errors"

var (
	ErrInvalidInput = errors.New(
		"invalid input",
	)

	ErrMediaNotFound = errors.New(
		"profile media not found",
	)

	Errforbidden = errors.New(
		"forbidden",
	)

	ErrInvalidMediaType = errors.New(
		"invalid media type",
	)

	ErrProfilePhotoLimitReached = errors.New(
		"profile photo limit reached",
	)

	ErrPortfolioLimitReached = errors.New(
		"portfolio photo limit reached",
	)

	ErrPricingLimitReached = errors.New(
		"pricing image limit reached",
	)

	ErrCompanyLogoLimitReached = errors.New(
		"company logo limit reached",
	)

	ErrCompanyNotFound = errors.New(
		"company not found",
	)
)
