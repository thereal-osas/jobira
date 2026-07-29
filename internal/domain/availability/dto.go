package availability

type CleanerAvailabilityRequest struct {
	AvailableDate string `json:"available_date"`
	StartTime     string `json:"start_time"`
	EndTime       string `json:"end_time"`
	Status        string `json:"status"`
	Notes         string `json:"notes"`
}

type UpdateAvailabilityRequest struct {
	AvailableDate string `json:"available_date"`
	StartTime     string `json:"start_time"`
	EndTime       string `json:"end_time"`
	Status        string `json:"status"`
	Notes         string `json:"notes"`
}
