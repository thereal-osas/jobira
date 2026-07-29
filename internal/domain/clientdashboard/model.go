package clientdashboard

import "time"

type ActiveJob struct {
	ID               uint      `json:"id"`
	Title            string    `json:"title"`
	Location         string    `json:"location"`
	JobType          string    `json:"job_type"`
	Budget           float64   `json:"budget"`
	Status           string    `json:"status"`
	ApplicationCount int       `json:"application_count"`
	CreatedAt        time.Time `json:"created_at"`
}

type ClientBooking struct {
	ID          uint      `json:"id"`
	JobID       uint      `json:"job_id"`
	CleanerID   uint      `json:"cleaner_id"`
	Status      string    `json:"status"`
	ScheduledAt time.Time `json:"scheduled_at"`
}

type ClientDashboard struct {
	ClientID uint `json:"client_id"`

	SubscriptionStatus string `json:"subscription_status"`
	LaunchGraceActive  bool   `json:"launch_grace_active"`
	Premium            bool   `json:"premium"`

	CanPostJob        bool `json:"can_post_job"`
	JobPostCount      int  `json:"job_post_count"`
	FreeJobPostLimit  int  `json:"free_job_post_limit"`
	JobsPostedToday   int  `json:"jobs_posted_today"`
	DailyJobPostLimit int  `json:"daily_job_post_limit"`

	ActiveJobCount        int `json:"active_job_count"`
	ApplicationsReceived  int `json:"applications_received"`
	ActiveBookingCount    int `json:"active_booking_count"`
	CompletedBookingCount int `json:"completed_booking_count"`
	FavouriteCleanerCount int `json:"favourite_cleaner_count"`
	PreferredCleanerCount int `json:"preferred_cleaner_count"`
	RepeatBookingCount    int `json:"repeat_booking_count"`

	ActiveJobs       []ActiveJob     `json:"active_jobs"`
	UpcomingBookings []ClientBooking `json:"upcoming_bookings"`

	UpgradeMessage string `json:"upgrade_message"`
}
