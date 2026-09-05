package workhistory

import (
	"time"
)

type WorkProofSummary struct {
	BeforeCount    int        `json:"before_count"`
	AfterCount     int        `json:"after_count"`
	HasBeforeProof bool       `json:"has_before_proof"`
	HasAfterProof  bool       `json:"has_after_proof"`
	IsVerifiedWork bool       `json:"is_verified_work"`
	LastUploadedAt *time.Time `json:"last_uploaded_at,omitempty"`
}

type ReviewSummary struct {
	Rating  *int    `json:"rating,omitempty"`
	Comment *string `json:"comment,omitempty"`
}

type WorkHistoryEntry struct {
	BookingID uint `json:"booking_id"`
	JobID     uint `json:"job_id"`

	ClientID   uint   `json:"client_id"`
	ClientName string `json:"client_name"`

	CleanerID   uint   `json:"cleaner_id"`
	CleanerName string `json:"cleaner_name"`

	JobTitle string `json:"job_title"`
	JobType  string `json:"job_type"`
	Location string `json:"location"`
	Budget   int    `json:"budget"`

	Status string `json:"status"`

	ScheduledAt    *time.Time `json:"scheduled_at,omitempty"`
	ScheduledEndAt *time.Time `json:"scheduled_end_at,omitempty"`
	CompletedAt    *time.Time `json:"completed_at,omitempty"`
	CancelledAt    *time.Time `json:"cancelled_at,omitempty"`

	CancellationReason string `json:"cancellation_reason,omitempty"`
	CancelledBy        *uint  `json:"cancelled_by,omitempty"`

	ClosedAt       *time.Time `json:"closed_at,omitempty"`
	ClosureStatus  string     `json:"closure_status,omitempty"`
	ClosureComment string     `json:"closure_comment,omitempty"`

	ClientConfirmedCompletion bool  `json:"client_confirmed_completion"`
	ClientWouldHireAgain      *bool `json:"client_would_hire_again,omitempty"`

	Review ReviewSummary `json:"review"`

	WorkProof WorkProofSummary `json:"work_proof"`

	IsRepeatClient bool `json:"is_repeat_client"`
	CanBookAgain   bool `json:"can_book_again"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type WorkHistoryDetail struct {
	WorkHistoryEntry

	BeforePhotos []WorkProofPhoto `json:"before_photos"`
	AfterPhoto   []WorkProofPhoto `json:"after_photos"`
}

type WorkProofPhoto struct {
	ID        uint      `json:"id"`
	ProofType string    `json:"proof_type"`
	PhotoURL  string    `json:"photo_url"`
	Caption   string    `json:"caption"`
	CreatedAt time.Time `json:"created_at"`
}
