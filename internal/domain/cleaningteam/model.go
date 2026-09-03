package cleaningteam

import "time"

type TeamMemberData struct {
	CleanerID   uint   `json:"cleaner_id"`
	CleanerName string `json:"cleaner_name"`

	Location string `json:"location"`

	ServicesOffered string `json:"services_offered"`
	LastJobType     string `json:"last_job_type"`

	AvailabilityStatus string `json:"availability_status"`

	IsVerified bool `json:"is_verified"`

	IsFavorite  bool `json:"is_favorite"`
	IsPreferred bool `json:"is_preferred"`

	PrivateNote string `json:"private_note"`

	CompletedJobsTogether int `json:"completed_jobs_together"`

	LastBookingID *uint      `json:"last_booking_id"`
	LastBookedAt  *time.Time `json:"last_booked_at"`

	AverageRating    float64 `json:"average_rating"`
	TotalReviews     int     `json:"total_reviews"`
	ReliabilityScore int     `json:"reliability_score"`
	Badge            string  `json:"badge"`
}

type CleaningTeamMember struct {
	CleanerID   uint   `json:"cleaner_id"`
	CleanerName string `json:"cleaner_name"`

	ServiceRole string   `json:"serivce_role"`
	Services    []string `json:"services"`

	Location string `json:"location"`

	AvailabilityStatus string `json:"availability_status"`

	IsVerified bool `json:"is_verified"`

	IsFavorite  bool `json:"is_favorite"`
	IsPreferred bool `json:"is_preferred"`

	RelationshipType string `json:"relationship_type"`

	PrivateNote string `json:"private_note"`

	CompletedJobsTogether int `json:"completed_jobs_together"`

	LastBookingID *uint      `json:"last_booking_id"`
	LastBookedAt  *time.Time `json:"last_booked_at"`

	CanRebook bool `json:"can_rebook"`

	AverageRating    float64 `json:"average_rating"`
	TotalReviews     int     `json:"total_reviews"`
	ReliabilityScore int     `json:"reliability_score"`
	Badge            string  `json:"badge"`
}

type CleaningTeam struct {
	ClientID uint `json:"client_id"`

	TotalMembers int `json:"total_members"`

	PreferredCount   int `json:"preferred_count"`
	FavoriteCount    int `json:"favorite_count"`
	HouseKeeperCount int `json:"housekeeper_count"`

	Members []CleaningTeamMember `json:"members"`
}
