package availablenow

import "time"

type CleanerAvailableNow struct {
	ID               uint      `json:"id"`
	CleanerID        uint      `json:"cleaner_id"`
	IsAvailable      bool      `json:"is_available"`
	AvailableFrom    time.Time `json:"available_from"`
	AvailableUntil   time.Time `json:"available_until"`
	Location         string    `json:"location"`
	TravelRadiusMiles int       `json:"travel_radius_miles"`
	JobTypes         []string  `json:"job_types"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type AvailableCleaner struct {
	CleanerID              uint      `json:"cleaner_id"`
	FullName               string    `json:"full_name"`
	Location               string    `json:"location"`
	TravelRadiusMiles      int       `json:"travel_radius_miles"`
	AvailableUntil         time.Time `json:"available_until"`
	JobTypes               []string  `json:"job_types"`
	AverageRating          float64   `json:"average_rating"`
	TotalReviews           int       `json:"total_reviews"`
	ReliabilityScore       int       `json:"reliability_score"`
	Badge                  string    `json:"badge"`
	IsVerified             bool      `json:"is_verified"`
	DBSVerified            bool      `json:"dbs_verified"`
	UsuallyRespondsMinutes int       `json:"usually_responds_minutes"`
}
