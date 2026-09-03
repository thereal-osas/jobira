package wokrproof

type CreateWorkProofRequest struct {
	ProofType string `json:"proof_type"`

	PhotoURL string `json:"photo_url"`

	Caption string `json:"caption"`
}

type ListBookingProofRequest struct {
	ProofType string `json:"proof_type"`
}

type DeleteWorkProofRequest struct {
	ProofID uint `json:"proof_type"`
}
