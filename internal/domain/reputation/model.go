package reputation

import "time"

type CleanerReputation struct {
	CleanerID                uint      `json:"cleaner_id"`
	AverageRating            float64   `json:"average_rating"`
	TotalReviews             int       `json:"total_reviews"`
	CompletedJobs            int       `json:"completed_jobs"`
	RepeatClients            int       `json:"repeat_clients"`
	WouldHireAgainCount      int       `json:"would_hire_again_count"`
	RecommendationPercentage int       `json:"recommendation_percentage"`
	Badge                    string    `json:"badge"`
	CreatedAt                time.Time `json:"created_at"`
	UpdatedAt                time.Time `json:"updated_at"`
}
