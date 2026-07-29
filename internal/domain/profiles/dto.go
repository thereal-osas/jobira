package profiles

import "time"

type CreateProfileRequest struct {
	Bio                string `json:"bio"`
	Location           string `json:"location"`
	Country            string `json:"country"`
	City               string `json:"city"`
	Region             string `json:"region"`
	PostcodeArea       string `json:"postcode_area"`
	AvailabilityStatus string `json:"availability_status"`
	TravelRadiusMiles  int    `json:"travel_radius_miles"`
	YearsExperience    int    `json:"years_experience"`
	HourlyRate         int    `json:"hourly_rate"`
	ServicesOffered    string `json:"services_offered"`
}

type UpdateProfileRequest struct {
	Bio                string `json:"bio"`
	Location           string `json:"location"`
	Country            string `json:"country"`
	City               string `json:"city"`
	Region             string `json:"region"`
	PostcodeArea       string `json:"postcode_area"`
	AvailabilityStatus string `json:"availability_status"`
	TravelRadiusMiles  int    `json:"travel_radius_miles"`
	YearsExperience    int    `json:"years_experience"`
	HourlyRate         int    `json:"hourly_rate"`
	ServicesOffered    string `json:"services_offered"`
}

type UpdateVerificationStatusRequest struct {
	VerificationStatus string `json:"verification_status"`
}

type JobHistoryItem struct {
	ApplicationID int       `json:"application_id"`
	JobID         int       `json:"job_id"`
	JobTitle      string    `json:"job_title"`
	Status        string    `json:"status"`
	Location      string    `json:"location"`
	CreatedAt     time.Time `json:"created_at"`
}
