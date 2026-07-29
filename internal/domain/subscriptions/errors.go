package subscriptions

import "errors"

var (
	ErrInvalidInput         = errors.New("invalid input")
	ErrPlanNotFound         = errors.New("subscription plan not found")
	ErrSubscriptionNotFound = errors.New("subscription not found")
	ErrInvalidStatus        = errors.New("invalid subscription status")
)
