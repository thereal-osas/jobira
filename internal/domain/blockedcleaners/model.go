package blockedcleaners

import "time"

type BlockedCleaner struct {
	ID        uint      `json:"id"`
	ClientID  uint      `json:"client_id"`
	CleanerID uint      `json:"cleaner_id"`
	Reason    string    `json:"reason"`
	CreatedAt time.Time `json:"created_at"`
}
