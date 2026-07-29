package adminsubscriptions

type CreatePlanRequest struct {
	Name             string `json:"name"`
	RoleType         string `json:"role_type"`
	PricePence       int    `json:"price_pence"`
	BillingInterval  string `json:"billing_interval"`
	ApplicationLimit int    `json:"application_limit"`
	JobPostLimit     int    `json:"job_post_limit"`
	CleanerSeatLimit int    `json:"cleaner_seat_limit"`
}

type UpdatePlanRequest struct {
	Name             string `json:"name"`
	RoleType         string `json:"role_type"`
	PricePence       int    `json:"price_pence"`
	BillingInterval  string `json:"billing_interval"`
	ApplicationLimit int    `json:"application_limit"`
	JobPostLimit     int    `json:"job_post_limit"`
	CleanerSeatLimit int    `json:"cleaner_seat_limit"`
	IsActive         bool   `json:"is_active"`
}

type UpdateUserSubscriptionRequest struct {
	PlanID uint   `json:"plan_id"`
	Status string `json:"status"`
}
