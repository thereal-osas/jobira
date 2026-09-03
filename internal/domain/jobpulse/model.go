package jobpulse

import "time"

type PulseLevel string

const (
	PulseQuiet       PulseLevel = "quiet"
	PulseActive      PulseLevel = "active"
	PulseHeatingUp   PulseLevel = "heating_up"
	PulseCompetitive PulseLevel = "high_competition"
)

type JobPulse struct {
	JobID uint `json:"job_id"`

	ApplicationCount int `json:"application_count"`

	RecentApplicationCount int `json:"recent_application_count"`

	JobAgeHours int `json:"job_age_hours"`

	HasBooking bool `json:"has_booking"`

	IsFilled bool `json:"is_filled"`

	PulseLevel PulseLevel `json:"pulse_level"`

	PulseScore int `json:"pulse_score"`

	Message string `json:"message"`

	UpdatedAt time.Time `json:"updated_at"`
}

type JobPulseSnapshot struct {
	JobID uint

	JobStatus string

	CreatedAt time.Time

	ApplicationCount int

	RecentApplicationCount int

	BookingCount int
}
