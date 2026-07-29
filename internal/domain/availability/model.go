package availability

import "time"

type CleanerAvailability struct {
	ID            uint      `json:"id"`
	CleanerID     uint      `json:"cleaner_id"`
	AvailableDate string    `json:"available_date"`
	StartTime     string    `json:"start_time"`
	EndTime       string    `json:"end_time"`
	Status        string    `json:"status"`
	Notes         string    `json:"notes"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
