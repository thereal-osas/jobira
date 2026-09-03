package profiles

import "time"

type CleanerProfile struct {
	ID                 uint      `json:"id"`
	UserID             uint      `json:"user_id"`
	Bio                string    `json:"bio"`
	Location           string    `json:"location"`
	Country            string    `json:"country"`
	City               string    `json:"city"`
	Region             string    `json:"region"`
	PostcodeArea       string    `json:"postcode_area"`
	AvailabilityStatus string    `json:"availability_status"`
	TravelRadiusMiles  int       `json:"travel_radius_miles"`
	YearsExperience    int       `json:"years_experience"`
	HourlyRate         int       `json:"hourly_rate"`
	ServicesOffered    string    `json:"services_offered"`
	JobsCompleted      int       `json:"jobs_completed"`
	JobsCancelled      int       `json:"jobs_cancelled"`
	ResponseRate       int       `json:"response_rate"`
	ReliabilityScore   int       `json:"reliability_score"`
	Badge              string    `json:"badge"`
	IsVerified         bool      `json:"is_verified"`
	VerificationStatus string    `json:"verification_status"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type ProfileStrengthItem struct {
	Code      string `json:"code"`
	Label     string `json:"label"`
	Completed bool   `json:"completed"`
	Points    int    `json:"points"`
}

type ProfileStrength struct {
	Percentage int `json:"percentage"`

	CompletedItems int `json:"completed_items"`
	TotalItems     int `json:"total_items"`

	Items []ProfileStrengthItem `json:"items"`

	NextAction string `json:"next_action"`

	IsComplete bool `json:"is_complete"`
}

type NewOnJobiraStatus struct {
	IsNew         bool      `json:"is_new"`
	Label         string    `json:"label"`
	JoinedAt      time.Time `json:"joined_at"`
	DaysOnJobira  int       `json:"days_on_jobira"`
	JobsCompleted int       `json:"jobs_completed"`
}
