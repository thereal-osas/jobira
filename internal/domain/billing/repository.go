package billing

import (
	"context"
	"time"
)

type Repository interface {
	GetPlanByID(ctx context.Context, planID uint) (*BillingPlan, error)
	GetUserEmail(ctx context.Context, userID uint) (string, error)
	GetStripeCustomerID(ctx context.Context, userID uint) (string, error)
	UpsertBillingCustomer(ctx context.Context, userID uint, email string, stripeCustomerID string) error
	SaveCheckoutSession(ctx context.Context, record *CheckoutSessionRecord) error
	MarkCheckoutSessionComplete(ctx context.Context, stripeSessionID string, stripeCustomerID string, stripeSubscriptionID string) (bool, error)
	MarkCheckoutSessionExpired(ctx context.Context, stripeSessionID string) error
	ActivateUserSubscription(ctx context.Context, userID uint, planID uint, stripeSubscriptionID string, periodStart time.Time, periodEnd time.Time) error

	UpdateUserSubscriptionStatusByCustomer(ctx context.Context, stripeCustomerID string, status string) error
	UpdateUserSubscriptionRenewalByCustomer(ctx context.Context, stripeCustomerID string, stripeSubscriptionID string, periodStart time.Time, periodEnd time.Time) error
	GetPromoCode(ctx context.Context, code string) (*PromoCode, error)
	IncrementPromoUse(ctx context.Context, code string) error
}
