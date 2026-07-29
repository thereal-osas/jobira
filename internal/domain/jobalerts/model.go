package jobalerts

import "time"

type JobAlert struct {
	ID            uint      `json:"id"`
	UserID        uint      `json:"user_id"`
	Location      string    `json:"location"`
	JobType       string    `json:"job_type"`
	MinimumBudget int       `json:"minimum_budget"`
	IsActive      bool      `json:"is_active"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
