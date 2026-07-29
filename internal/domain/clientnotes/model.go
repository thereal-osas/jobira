package clientnotes

import "time"

type ClientCleanerNote struct {
	ID        uint      `json:"id"`
	ClientID  uint      `json:"client_id"`
	CleanerID uint      `json:"cleaner_id"`
	Note      string    `json:"note"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}


