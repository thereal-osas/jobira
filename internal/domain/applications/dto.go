package applications

type CreateApplicationRequest struct {
	CoverMessage string `json:"cover_message"`
	ProposedRate int `json:"proposed_rate"`
}

type UpdateApplicationStatusRequest struct {
	Status string `json:"status"`
}

