package availablenow

type SetAvailableNowRequest struct {
	DurationMinutes   int      `json:"duration_minutes"`
	Location          string   `json:"location"`
	TravelRadiusMiles int      `json:"travel_radius_miles"`
	JobTypes          []string `json:"job_types"`
}

type UpdateAvailableNowRequest struct {
	DurationMinutes   int      `json:"duration_minutes"`
	Location          string   `json:"location"`
	TravelRadiusMiles int      `json:"travel_radius_miles"`
	JobTypes          []string `json:"job_types"`
}

type AvailableNowSearchRequest struct {
	Location      string  `json:"location"`
	JobType       string  `json:"job_type"`
	MinimumRating float64 `json:"minimum_rating"`
	VerifiedOnly  bool    `json:"verified_only"`
	DBSRequired   bool    `json:"dbs_required"`
	Limit         int     `json:"limit"`
	Offset        int     `json:"offset"`
}

type AvailableNowStatusResponse struct {
	IsAvailable      bool   `json:"is_available"`
	AvailableUntil   string `json:"available_until,omitempty"`
	MinutesRemaining int    `json:"minutes_remaining"`
}
