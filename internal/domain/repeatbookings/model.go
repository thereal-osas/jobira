package repeatbookings

import "time"

type RepeatBooking struct {
	ID        uint      `json:"id"`
	ClientID  uint      `json:"client_id"`
	CleanerID uint      `json:"cleaner_id"`
	JobID     uint      `json:"job_id"`
	Message   string    `json:"message"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type BookingSnapshot struct {
	ID            uint
	JobID         uint
	ApplicationID *uint
	ClientID      uint
	CleanerID     uint
	Status        string
}

type RepeatBookingRequest struct {
	ID                uint       `json:"id"`
	OriginalBookingID uint       `json:"original_booking_id"`
	NewBookingID      *uint      `json:"new_booking_id"`
	ClientID          uint       `json:"client_id"`
	CleanerID         uint       `json:"cleaner_id"`
	JobID             uint       `json:"job_id"`
	ScheduledAt       time.Time  `json:"scheduled_at"`
	ScheduledEndAt    *time.Time `json:"scheduled_end_at,omitempty"`
	Status            string     `json:"status"`
	Message           string     `json:"message"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

type BookAgainTransaction struct {
	OriginalBookingID uint
	ClientID          uint
	CleanerID         uint
	JobID             uint
	ApplicationID     *uint
	ScheduledAt       time.Time
	Status            string
	Message           string
}
