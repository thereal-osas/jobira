package jobinvitations

import "time"

type JobInvitation struct {
	ID        uint      `json:"id"`
	JobID     uint      `json:"job_id"`
	ClientID  uint      `json:"client_id"`
	CleanerID uint      `json:"cleaner_id"`
	Status    string    `json:"status"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
