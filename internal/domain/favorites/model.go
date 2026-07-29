package favorites

import "time"

type FavoriteCleaner struct {
	ID        uint      `json:"id"`
	ClientID  uint      `json:"client_id"`
	CleanerID uint      `json:"cleaner_id"`
	CreatedAt time.Time `json:"created_at"`
}
