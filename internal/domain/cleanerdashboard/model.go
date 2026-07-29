package cleanerdashboard

import "time"

type UpcomingBooking struct {
	ID          uint      `json:"id"`
	JobID       uint      `json:"job_id"`
	ClientID    uint      `json:"client_id"`
	Status      string    `json:"status"`
	ScheduledAt time.Time `json:"scheduled_at"`
}

type CleanerDashboard struct {
	CleanerID uint `json:"cleaner_id"`

	SubscriptionStatus string `json:"subscription_status"`
	TrialActive        bool   `json:"trial_active"`
	LaunchGraceActive  bool   `json:"launch_grace_active"`
	Premium            bool   `json:"premium"`

	CanApply              bool `json:"can_apply"`
	ApplicationsToday     int  `json:"applications_today"`
	DailyApplicationLimit int  `json:"daily_application_limit"`

	AverageRating            float64 `json:"average_rating"`
	TotalReviews             int     `json:"total_reviews"`
	CompletedJobs            int     `json:"completed_jobs"`
	RepeatClients            int     `json:"repeat_clients"`
	RecommendationPercentage int     `json:"recommendation_percentage"`
	Badge                    string  `json:"badge"`

	FavouriteCount       int `json:"favourite_count"`
	PreferredClientCount int `json:"preferred_client_count"`
	UpcomingBookingCount int `json:"upcoming_booking_count"`

	UpcomingBookings []UpcomingBooking `json:"upcoming_bookings"`
	UpgradeMessage   string            `json:"upgrade_message"`
}
