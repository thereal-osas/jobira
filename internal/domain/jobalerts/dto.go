package jobalerts

type CreateJobAlertRequest struct {
	Location      string `json:"location"`
	JobType       string `json:"job_type"`
	MinimumBudget int    `json:"minimum_budget"`
}

type UpdateJobAlertRequest struct {
	Location      string `json:"location"`
	JobType       string `json:"job_type"`
	MinimumBudget int    `json:"minimum_budget"`
	IsActive      bool   `json:"is_active"`
}
