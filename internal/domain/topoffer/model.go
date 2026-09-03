package topoffer

import "time"

type ScoreBreakdown struct {
	PriceFitScore int `json:"price_fit_score"`

	RatingScore int `json:"rating_score"`

	ReliabilityScore int `json:"reliability_score"`

	ExperienceScore int `json:"experience_score"`

	RecommendationScore int `json:"recommendation_score"`

	ResponseScore int `json:"response_score"`

	VerificationScore int `json:"verificartion_score"`
}

type RankedOffer struct {
	ApplicationID uint `json:"application_id"`

	JobID uint `json:"job_id"`

	CleanerID uint `json:"cleaner_id"`

	CleanerName string `json:"cleaner_name"`

	CoverMessage string `json:"cover_message"`

	ProposedRate int `json:"proposed_rate"`

	ApplicationStatus string `json:"application_status"`

	AverageRating float64 `json:"average_rating"`

	TotalReviews int `json:"total_reviews"`

	CompletedJobs int `json:"completed_jobs"`

	ReliabilityScore int `json:"reliability_score"`

	RecommendationPercentage int `json:"recommendation_percentage"`

	AverageResponseMinutes int `json:"average_response_minutes"`

	Badge string `json:"badge"`

	IsVerified bool `json:"is_verified"`

	DBSVerified bool `json:"dbs_verified"`

	TopOfferScore int `json:"top_offer_score"`

	Rank int `json:"rank"`

	IsTopOffer bool `json:"is_top_offer"`

	ScoreBreakdown ScoreBreakdown `json:"score_breakdown"`

	Reasons []string `json:"reason"`

	AppliedAt time.Time `json:"applied_at"`
}

type JobOfferContext struct {
	JobID uint `json:"job_id"`

	ClientID uint `json:"client_id"`

	JobType string `json:"job_type"`

	Budget int `json:"budget"`

	Status string `json:"status"`

	Location string `json:"location"`
}

type OfferCandidate struct {
	ApplicationID uint

	JobID uint

	CleanerID uint

	CleanerName string

	CoverMessage string

	ProposedRate int

	ApplicationStatus string

	AverageRating float64

	TotalReviews int

	CompletedJobs int

	ReliabilityScore int

	RecommendationPercentage int

	AverageResponseMinutes int

	Badge string

	IsVerified bool

	DBSVerified bool

	AppliedAt time.Time
}

type TopOfferResult struct {
	JobID uint `json:"job_id"`

	TotalOffers int `json:"total_offers"`

	TopOfferApplicationID *uint `json:"top_offer_application_id,omitempty"`

	Offers []RankedOffer `json:"offers"`
}
