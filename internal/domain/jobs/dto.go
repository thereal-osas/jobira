package jobs

type CreateJobRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Location    string `json:"location"`
	JobType     string `json:"job_type"`
	ListingType string `json:"listing_type"`
	Budget      int    `json:"budget"`
}

type UpdateJobRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Location    string `json:"location"`
	JobType     string `json:"job_type"`
	ListingType string `json:"listing_type"`
	Budget      int    `json:"budget"`
	Status      string `json:"status"`
}
