package verifications

import "time"

type VerificationRequest struct {
	ID               uint       `json:"id"`
	UserID           uint       `json:"user_id"`
	VerificationType string     `json:"verification_type"`
	DocumentURL      string     `json:"document_url"`
	Status           string     `json:"status"`
	AdminNotes       string     `json:"admin_notes"`
	ReviewedBy       *uint      `json:"reviewed_by"`
	ReviewedAt       *time.Time `json:"reviewed_at"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}
