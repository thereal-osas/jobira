package applications

import "time"

type Application struct {
	ID           uint      `json:"id"`
	JobID        uint      `json:"job_id"`
	CleanerID    uint      `json:"cleaner_id"`
	CoverMessage string    `json:"cover_message"`
	ProposedRate int       `json:"proposed_rate"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
