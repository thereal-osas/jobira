package repeatbookings

type CreateRepeatBookingRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Location    string `json:"location"`
	JobType     string `json:"job_type"`
	ListingType string `json:"listing_type"`
	Budget      int    `json:"budget"`
	Message     string `json:"message"`
}

type BookingAgainRequest struct {
	ScheduledAt string `json:"scheduled_at"`
	Message     string `json:"message"`
}
