package profiles

type SearchProfilesRequest struct {
	ClientID           uint
	Country            string
	City               string
	Region             string
	PostcodeArea       string
	AvailabilityStatus string
	ServicesOffered    string
	MinExperience      int
	MaxHourlyRate      int
	MaxTravelRadius    int
	IsVerified         *bool
}

type SearchProfilesResult struct {
	CleanerProfile
	MatchScore int `json:"match_score"`
}
