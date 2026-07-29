package subscriptionaccess

type CleanerAccessStatus struct {
	UserID             uint   `json:"user_id"`
	SubscriptionStatus string `json:"subscription_status"`
	TrialActive        bool   `json:"trial_active"`
	LaunchGraceActive  bool   `json:"launch_grace_active"`
	Premium            bool   `json:"premium"`

	CanViewJobs          bool `json:"can_view_jobs"`
	CanApply             bool `json:"can_apply"`
	CanViewClientDetails bool `json:"can_view_client_details"`

	HasInstantAlerts     bool `json:"has_instant_alerts"`
	HasPriorityPlacement bool `json:"has_priority_placement"`
	HasPremiumProfile    bool `json:"has_premium_profile"`
	HasEarlyJobAccess    bool `json:"has_early_job_access"`

	ApplicationCount              int `json:"application_count"`
	FreeApplicationLimit          int `json:"free_application_limit"`
	ApplicationsToday             int `json:"applications_today"`
	DailyApplicationLimit         int `json:"daily_application_limit"`
	FreeJobVisibilityDelayMinutes int `json:"free_job_visibility_delay_minutes"`

	UpgradeMessage string `json:"upgrade_message"`
}

type ClientAccessStatus struct {
	UserID             uint   `json:"user_id"`
	SubscriptionStatus string `json:"subscription_status"`
	LaunchGraceActive  bool   `json:"launch_grace_active"`
	Premium            bool   `json:"premium"`

	CanPostJob          bool `json:"can_post_job"`
	CanViewApplicants   bool `json:"can_view_applicants"`
	CanContactCleaners  bool `json:"can_contact_cleaners"`
	CanUseRepeatBooking bool `json:"can_use_repeat_booking"`

	JobsPostedToday   int `json:"jobs_posted_today"`
	DailyJobPostLimit int `json:"daily_job_post_limit"`
	FreeJobPostLimit  int `json:"free_job_post_limit"`
	JobPostCount      int `json:"job_post_count"`

	UpgradeMessage string `json:"upgrade_message"`
}
