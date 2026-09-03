package wokrproof

import "time"

type ProofType string

const (
	ProofTypeBefore ProofType = "before"
	ProofTypeAfter  ProofType = "after"
)

type WorkProof struct {
	ID uint `json:"id"`

	BookingID uint `json:"booking_id"`

	CleanerID uint `json:"cleaner_id"`

	ProofType ProofType `json:"proof_type"`

	PhotoURL string `json:"photo_url"`

	Caption string `json:"caption"`

	CreatedAt time.Time `json:"created_at"`
}

type BookingProofSummary struct {
	BookingID uint `json:"booking_id"`

	CleanerID uint `json:"cleaner_id"`

	BeforePhotos []WorkProof `json:"before_photos"`

	AfterPhotos []WorkProof `json:"after_photos"`

	BeforeCount int `json:"before_count"`

	AfterCount int `json:"after_count"`

	HasBeforeProof bool `json:"has_before_proof"`

	HasAfterProof bool `json:"has_after_proof"`

	IsVerifiedWork bool `json:"is_verified_work"`

	LastUploadedAt *time.Time `json:"last_uploaded_at,omitempty"`
}

type VerifiedWorkSummary struct {
	CleanerID uint `json:"cleaner_id"`

	VerifiedJobs int `json:"verified_jobs"`

	BeforePhotos int `json:"before_photos"`

	AfterPhotos int `json:"after_photos"`

	TotalPhotos int `json:"total_photos"`
}
