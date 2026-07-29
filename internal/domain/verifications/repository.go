package verifications

import "context"

type Repository interface {
	Create(ctx context.Context, request *VerificationRequest) error
	GetByID(ctx context.Context, id uint) (*VerificationRequest, error)
	ListByUserID(ctx context.Context, userID uint) ([]VerificationRequest, error)
	ListAll(ctx context.Context) ([]VerificationRequest, error)
	ListPending(ctx context.Context) ([]VerificationRequest, error)
	Review(ctx context.Context, requestID uint, status string, adminNotes string, adminID uint) error
}
