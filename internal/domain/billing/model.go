package billing

import "time"

type BillingPlan struct {
	ID            uint   `json:"id"`
	Name          string `json:"name"`
	RoleType      string `json:"role_type"`
	PricePence    int    `json:"price_pence"`
	StripePriceID string `json:"stripe_price_id"`
}

type BillingCustomer struct {
	ID               uint      `json:"id"`
	UserID           uint      `json:"user_id"`
	StripeCustomerID string    `json:"stripe_customer_id"`
	Email            string    `json:"email"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type CheckoutSessionRecord struct {
	ID                   uint      `json:"id"`
	UserID               uint      `json:"user_id"`
	PlanID               uint      `json:"plan_id"`
	StripeSessionID      string    `json:"stripe_session_id"`
	StripeCustomerID     string    `json:"stripe_customer_id"`
	StripeSubscriptionID string    `json:"stripe_subscription_id"`
	Status               string    `json:"status"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

type PromoCode struct {
	ID               uint       `json:"id"`
	Code             string     `json:"code"`
	Description      string     `json:"description"`
	DiscountType     string     `json:"discount_type"`
	PercentageOff    int        `json:"percentage_off"`
	FixedAmountPence int        `json:"fixed_amount_pence"`
	FreeMonths       int        `json:"free_months"`
	MaxUses          int        `json:"max_uses"`
	TimesUsed        int        `json:"times_used"`
	UserType         string     `json:"user_type"`
	PlanID           *uint      `json:"plan_id"`
	StartsAt         *time.Time `json:"starts_at"`
	ExpiresAt        *time.Time `json:"expires_at"`
	IsActive         bool       `json:"is_active"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}
