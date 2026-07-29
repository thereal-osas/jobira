package savedjobs

import "time"

type SavedJob struct {
	ID        uint      `json:"id"`
	UserID    uint      `json:"user_id"`
	JobID     uint      `json:"job_id"`
	CreatedAt time.Time `json:"created_at"`
}
