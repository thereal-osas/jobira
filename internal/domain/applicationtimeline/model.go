package applicationtimeline

import "time"

type Event struct {
	ID            uint      `json:"id"`
	ApplicationID uint      `json:"application_id"`
	Status        string    `json:"status"`
	ActorUserID   *uint     `json:"actor_user_id,omitempty"`
	Note          string    `json:"note"`
	CreatedAt     time.Time `json:"created_at"`
}

type TimeLine struct {
	ApplicationID uint    `json:"application_id"`
	CurrentStatus string  `json:"current_status"`
	Events        []Event `json:"events"`
}
