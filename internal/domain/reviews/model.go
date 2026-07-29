package reviews

import "time"

type Review struct {
	ID        uint      `json:"id"`
	BookingID uint      `json:"booking_id"`
	CleanerID uint      `json:"cleaner_id"`
	ClientID  uint      `json:"client_id"`
	JobID     uint      `json:"job_id"`
	Rating    int       `json:"rating"`
	Comment   string    `json:"comment"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
