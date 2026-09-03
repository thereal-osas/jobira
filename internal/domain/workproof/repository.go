package wokrproof

import "context"

type BookingSnapshot struct {
	ID uint

	JobID uint

	ClientID uint

	CleanerID uint

	Status string
}

type Repository interface {
	GetBooking(ctx context.Context, bookingID uint) (*BookingSnapshot, error)
	Create(ctx context.Context, proof *WorkProof) error
	GetByID(ctx context.Context, proofID uint) (*WorkProof, error)
	ListByBookingID(ctx context.Context, bookingID uint) ([]WorkProof, error)
	CountByBookingAndType(ctx context.Context, bookingID uint, proofType ProofType) (int, error)
	Delete(ctx context.Context, proofID uint, cleanerID uint) error
	GetVerifiedWorkSummary(ctx context.Context, cleanerID uint) (*VerifiedWorkSummary, error)
}
