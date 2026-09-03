package matchscore

type CandidateData struct {
	ApplicationID uint `json:"application_id"`
	JobID         uint `json:"job_id"`
	ClientID      uint `json:"client_id"`
	CleanerID     uint `json:"cleaner_id"`

	CleanerName string `json:"cleaner_name"`

	JobType     string `json:"job_type"`
	JobLocation string `json:"job_location"`

	CleanerLocation     string `json:"cleaner_location"`
	CleanerPostcodeArea string `json:"cleaner_postcode_area"`
	ServicesOffered     string `json:"services_offered"`

	AvailabilityStatus string `json:"availability_status"`

	IsVerified bool `json:"is_verified"`

	ReliabilityScore int `json:"reliability_score"`

	ResponseRate           int `json:"response_rate"`
	AverageResponseMinutes int `json:"average_response_minutes"`

	AverageRating float64 `json:"average_rating"`
	TotalReviews  int     `json:"total_reviews"`

	Badge string `json:"badge"`
}

type MatchScoreBreakdown struct {
	ServiceFit   int `json:"service_fit"`
	Availability int `json:"availabilty"`
	Reliability  int `json:"reliabilty"`
	Location     int `json:"location"`
	Verification int `json:"verification"`
	Response     int `json:"response"`
}

type MatchReason struct {
	Code    string `json:"code"`
	Label   string `json:"label"`
	Message string `json:"message"`
	Points  int    `json:"points"`
}

type CleanerMatch struct {
	ApplicationID uint `json:"application_id"`
	JobID         uint `json:"job_id"`
	CleanerID     uint `json:"cleaner_id"`

	CleanerName string `json:"cleaner_name"`

	MatchScore int    `json:"match_score"`
	MatchLabel string `json:"match_label"`

	Rank int `json:"rank"`

	IsTopOffer    bool `json:"is_top_offer"`
	IsRecommended bool `json:"is_recommended"`

	Breakdown MatchScoreBreakdown `json:"breakdown"`

	WhyYoureSeeingThis []MatchReason `json:"why_youre_seeing_this"`

	AverageRating float64 `json:"average_rating"`
	TotalReviews  int     `json:"total_reviews"`

	ReliabilityScore int `json:"reliability_score"`

	ResponseRate           int `json:"response_rate"`
	AverageResponseMinutes int `json:"average_response_minutes"`

	IsVerified         bool   `json:"is_verified"`
	AvailabilityStatus string `json:"availability_status"`
	Badge              string `json:"badge"`
}

type JobMatches struct {
	JobID uint `json:"job_id"`

	TotalApplicants int `json:"total_applicants"`

	TopOffer *CleanerMatch `json:"top_offer,omitempty"`

	Matches []CleanerMatch `json:"matches"`
}
