package jobpulse

type GetJobPulseRequest struct {
	JobID uint `json:"job_id"`
}

type JobPulseSummary struct {
	JobID uint `json:"job_id"`

	PulseLevel string `json:"pulse_level"`

	PulseScore int `json:"pulse_score"`

	ApplicationCount int `json:"application_count"`

	Message string `json:"message"`
}
