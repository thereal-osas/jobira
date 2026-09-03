package wokrproof

import "errors"

var (
	ErrInvalidInput = errors.New(
		"invalid input",
	)

	ErrBookingNotFound = errors.New(
		"booking not found",
	)

	ErrProofNotFound = errors.New(
		"work proof not found",
	)

	ErrForbidden = errors.New(
		"forbidden",
	)

	ErrInvalidProofType = errors.New(
		"invalid proof type",
	)

	ErrProofLimitReached = errors.New(
		"work proof photo limit reached",
	)

	ErrBookingNotEligible = errors.New(
		"booking is not eligible for work proof",
	)
)
