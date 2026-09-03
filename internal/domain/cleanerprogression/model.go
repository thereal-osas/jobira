package cleanerprogression

type ProgressMetric struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Current     int    `json:"current"`
	Target      int    `json:"target"`
	Percentage  int    `json:"percentage"`
	Completed   bool   `json:"completed"`
	Description string `json:"description"`
}

type BadgeProgress struct {
	CurrentBadge         string  `json:"current_badge"`
	NextBadge            string  `json:"next_badge,omitempty"`
	Progress             int     `json:"progress"`
	JobsRemaining        int     `json:"jobs_remaining"`
	ReviewsNeeded        int     `json:"reviews_needed"`
	RatingNeeded         float64 `json:"rating_needed"`
	RecommendationNeeded int     `json:"recommendation_needed"`
	IsHighestLevel       bool    `json:"is_highest_level"`
}

type ProgressSnapshot struct {
	CleanerID uint `json:"cleaner_id"`

	CurrentBadge string `json:"current_badge"`

	ReliabilityScore int `json:"reliability_score"`
	CompletionRate   int `json:"completion_rate"`
	CancellationRate int `json:"cancellation_rate"`
	ResponseRate     int `json:"response_rate"`

	AverageRating float64 `json:"average_rating"`
	TotalReviews  int     `json:"total_reviews"`

	CompletedJobs int `json:"completed_jobs"`
	RepeatClients int `json:"repeat_clients"`

	RecommendationPercentage int `json:"recommendation_percentage"`

	BadgeProgress BadgeProgress `json:"badge_progress"`

	Milestones []ProgressMetric `json:"milestones"`

	OverallProgress int `json:"overall_progress"`

	Encouragement string `json:"encouragement"`
}
