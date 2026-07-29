package jobs

import "time"

type Job struct {
	ID          uint      `json:"id"`
	ClientID    uint      `json:"client_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Location    string    `json:"location"`
	JobType     string    `json:"job_type"`
	ListingType string    `json:"listing_type"`
	Budget      int       `json:"budget"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
