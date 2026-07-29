package bookingtimeline

import "time"

type StatusHistory struct {
	ID         uint      `json:"id"`
	BookingID  uint      `json:"booking_id"`
	ChangedBy  *uint     `json:"changed_by"`
	FromStatus string    `json:"from_status"`
	ToStatus   string    `json:"to_status"`
	Note       string    `json:"note"`
	CreatedAt  time.Time `json:"created_at"`
}

type BookingTimeline struct {
	BookingID     uint            `json:"booking_id"`
	CurrentStatus string          `json:"current_status"`
	History       []StatusHistory `json:"history"`
}
