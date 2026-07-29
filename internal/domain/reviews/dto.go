package reviews

type CreateReviewRequest struct {
	Rating    int    `json:"rating"`
	Comment   string `json:"comment"`
	CleanerID uint   `json:"cleaner_id"`
	BookingID uint   `json:"booking_id"`
}
