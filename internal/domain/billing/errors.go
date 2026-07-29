package billing

import "errors"

var (
	ErrInvalidInput         = errors.New("invalid input")
	ErrPlanNotFound         = errors.New("subscription plan not found")
	ErrStripePriceMissing   = errors.New("stripe price id missing")
	ErrCustomerNotFound     = errors.New("billing customer not found")
	ErrCheckoutSessionError = errors.New("checkout session error")
	ErrPortalSessionError   = errors.New("billing portal session error")
	ErrWebhookInvalid       = errors.New("invalid stripe webhook")
	ErrPromoCodeNotFound    = errors.New("promo code not found")
	ErrPromoCodeInvalid     = errors.New("promo code invalid")
)
