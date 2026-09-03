package reputation

import "time"

type CleanerReputation struct {
	CleanerID                uint    `json:"cleaner_id"`
	AverageRating            float64 `json:"average_rating"`
	TotalReviews             int     `json:"total_reviews"`
	CompletedJobs            int     `json:"completed_jobs"`
	RepeatClients            int     `json:"repeat_clients"`
	WouldHireAgainCount      int     `json:"would_hire_again_count"`
	RecommendationPercentage int     `json:"recommendation_percentage"`

	// Reliability Metrics.
	TotalBookings        int `json:"total_bookings"`
	CleanerCancellations int `json:"cleaner_cancellations"`
	CompletionRate       int `json:"completion_rate"`
	CancellationRate     int `json:"cancellation_rate"`
	ReliabilityScore     int `json:"reliability_score"`

	// Response behaviour.
	EligibleResponseMessages int `json:"eligible_response_messages"`
	RespondedMessages        int `json:"responded_messages"`
	AverageResponseMinutes   int `json:"average_response_minutes"`
	ResponseRate             int `json:"response_rate"`

	// Main reputation tier.
	Badge string `json:"badge"`

	// Achievement/progression system.
	Achievements []Achievement `json:"achievements"`

	CurrentMilestone       int `json:"current_milestone"`
	NextMilestone          int `json:"next_milestone"`
	JobsUntilNextMilestone int `json:"jobs_until_next_milestone"`
	MilestoneProgress      int `json:"milestone_progress"`

	// Human-friendly reliability state.
	ReliabilityStatus string `json:"reliability_status"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Achievement struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Category    string `json:"category"`

	Earned bool `json:"earned"`

	Current int `json:"current"`
	Target  int `json:"target"`

	EarnedAt *time.Time `json:"earned_at"`
}
