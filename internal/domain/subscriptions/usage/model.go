package usage

import "time"

type UserUsage struct {
	ID                   uint       `json:"id"`
	UserID               uint       `json:"user_id"`
	ApplicationCount     int        `json:"application_count"`
	JobPostCount         int        `json:"job_post_count"`
	FreeApplicationLimit int        `json:"free_application_limit"`
	FreeJobPostLimit     int        `json:"free_job_post_limit"`
	MonetisationEnabled  bool       `json:"monetisation_enabled"`
	TrialStartedAt       time.Time  `json:"trial_started_at"`
	TrialEndsAt          *time.Time `json:"trial_ends_at"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

type SubscriptionAccess struct {
	Status           string
	ApplicationLimit int
	JobPostLimit     int
}
