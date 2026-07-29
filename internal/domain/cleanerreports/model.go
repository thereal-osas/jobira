package cleanerreports

import "time"

type CleanerReport struct {
	ID        uint      `json:"id"`
	ClientID  uint      `json:"client_id"`
	CleanerID uint      `json:"cleaner_id"`
	Reason    string    `json:"reason"`
	Details   string    `json:"details"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}


