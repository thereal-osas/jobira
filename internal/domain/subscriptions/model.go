package subscriptions

import "time"

type SubscriptionPlan struct {
	ID               uint      `json:"id"`
	Name             string    `json:"name"`
	RoleType         string    `json:"role_type"`
	PricePence       int       `json:"price_pence"`
	BillingInterval  string    `json:"billing_interval"`
	ApplicationLimit int       `json:"application_limit"`
	JobPostLimit     int       `json:"job_post_limit"`
	CleanerSeatLimit int       `json:"cleaner_seat_limit"`
	IsActive         bool `json:"is_active"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type UserSubscription struct {
	ID                 uint       `json:"id"`
	UserID             uint       `json:"user_id"`
	PlanID             *uint      `json:"plan_id"`
	Status             string     `json:"status"`
	TrialStartedAt     time.Time  `json:"trial_started_at"`
	TrialEndsAt        *time.Time `json:"trial_ends_at"`
	CurrentPeriodStart *time.Time `json:"current_period_start"`
	CurrentPeriodEnd   *time.Time `json:"current_period_end"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}
