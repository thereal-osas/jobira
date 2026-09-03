package bookings

type CreateBookingRequest struct {
	JobID          uint   `json:"job_id"`
	ApplicationID  *uint  `json:"application_id"`
	CleanerID      uint   `json:"cleaner_id"`
	ScheduledAt    string `json:"scheduled_at"`
	ScheduledEndAt string `json:"scheduled_end_at"`
}

type CancelBookingRequest struct {
	Reason string `json:"reason"`
}

type CloseBookingRequest struct {
	ClientConfirmedCompletion bool   `json:"client_confirmed_completion"`
	ClientRating              int    `json:"client_rating"`
	ClientWouldHireAgain      bool   `json:"client_would_hire_again"`
	ClosureComment            string `json:"closure_comment"`
	AddFavourites             bool   `json:"add_to_favourites"`
	SetAsPreferred            bool   `json:"set_as_preffered"`
}
