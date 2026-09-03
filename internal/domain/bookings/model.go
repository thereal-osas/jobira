package bookings

import "time"

type Booking struct {
	ID                        uint       `json:"id"`
	JobID                     uint       `json:"job_id"`
	ApplicationID             *uint      `json:"application_id"`
	ClientID                  uint       `json:"client_id"`
	CleanerID                 uint       `json:"cleaner_id"`
	Status                    string     `json:"status"`
	ScheduledAt               *time.Time `json:"scheduled_at"`
	ScheduledEndAt            *time.Time `json:"scheduled_end_at"`
	CompletedAt               *time.Time `json:"completed_at"`
	CancelledAt               *time.Time `json:"cancelled_at"`
	CancellationReason        string     `json:"cancellation_reason"`
	CancelledBy               *uint      `json:"cancelled_by,omitempty"`
	CreatedAt                 time.Time  `json:"created_at"`
	UpdatedAt                 time.Time  `json:"updated_at"`
	ClosedAt                  *time.Time `json:"closed_at"`
	ClosureStatus             string     `json:"closure_status"`
	ClosureComment            string     `json:"closure_comment"`
	ClientConfirmedCompletion bool       `json:"client_confirmed_completion"`
	ClientWouldHireAgain      *bool      `json:"client_would_hire_again"`
	ClientRating              *int       `json:"client_rating"`
}

type CloseBookingTransaction struct {
	BookingID           uint
	JobID               uint
	ClientID            uint
	CleanerID           uint
	Rating              int
	WouldHireAgain      bool
	Comment             string
	AddToFavourites     bool
	SetAsPreferred      bool
	ChangedBy           uint
	PreviousStatus      string
	NotificationTitle   string
	NotificationMessage string
}
