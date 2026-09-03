package topoffer

type ListTopOffersRequest struct {
	Limit int `json:"limit"`

	Offset int `json:"offset"`
}

type TopOfferSummary struct {
	ApplicationID uint `json:"application_id"`

	Rank int `json:"rank"`

	Score int `json:"score"`

	IsTopOffer bool `json:"is_top_offer"`

	Reason []string `json:"reason"`
}
