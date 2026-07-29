package verifications

type CreateVerificationRequest struct {
	VerificationType string `json:"verification_type"`
	DocumentURL      string `json:"document_url"`
}

type ReviewVerificationRequest struct {
	Status     string `json:"status"`
	AdminNotes string `json:"admin_notes"`
}
