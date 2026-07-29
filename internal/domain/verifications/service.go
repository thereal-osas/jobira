package verifications

import (
	"context"
	"strings"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, userID uint, req CreateVerificationRequest) (*VerificationRequest, error) {
	req.VerificationType = strings.TrimSpace(req.VerificationType)
	req.DocumentURL = strings.TrimSpace(req.DocumentURL)

	if userID == 0 || req.VerificationType == "" || req.DocumentURL == "" {
		return nil, ErrInvalidInput
	}

	if !isAllowedVerficationType(req.VerificationType) {
		return nil, ErrInvalidVerificationType
	}

	request := &VerificationRequest{
		UserID:           userID,
		VerificationType: req.VerificationType,
		DocumentURL:      req.DocumentURL,
		Status:           "pending",
	}

	if err := s.repo.Create(ctx, request); err != nil {
		return nil, err
	}

	return request, nil
}

func (s *Service) GetByID(ctx context.Context, requestID uint, userID uint, role string) (*VerificationRequest, error) {
	if requestID == 0 || userID == 0 {
		return nil, ErrInvalidInput
	}

	request, err := s.repo.GetByID(ctx, requestID)
	if err != nil {
		return nil, err
	}

	if role != "admin" && request.UserID != userID {
		return nil, ErrForbidden
	}

	return request, nil
}
func (s *Service) ListMine(ctx context.Context, userID uint) ([]VerificationRequest, error) {
	if userID == 0 {
		return nil, ErrInvalidInput
	}

	return s.repo.ListByUserID(ctx, userID)
}

func (s *Service) ListAll(ctx context.Context) ([]VerificationRequest, error) {
	return s.repo.ListAll(ctx)
}

func (s *Service) ListPending(ctx context.Context) ([]VerificationRequest, error) {
	return s.repo.ListPending(ctx)
}

func (s *Service) Review(ctx context.Context, requestID uint, adminID uint, req ReviewVerificationRequest) (*VerificationRequest, error) {
	req.Status = strings.TrimSpace(req.Status)
	req.AdminNotes = strings.TrimSpace(req.AdminNotes)

	if requestID == 0 || adminID == 0 {
		return nil, ErrInvalidInput
	}

	if !isAllowedVerificationStatus(req.Status) {
		return nil, ErrInValidVerificationStatus
	}

	if err := s.repo.Review(ctx, requestID, req.Status, req.AdminNotes, adminID); err != nil {
		return nil, err
	}

	return s.repo.GetByID(ctx, requestID)
}

func isAllowedVerficationType(verificationType string) bool {
	if verificationType == "id_check" {
		return true
	}

	if verificationType == "dbs_check" {
		return true
	}

	if verificationType == "insurance" {
		return true
	}

	return false
}

func isAllowedVerificationStatus(status string) bool {
	if status == "pending" {
		return true
	}

	if status == "pending" {
		return true
	}

	if status == "approved" {
		return true
	}

	if status == "rejected" {
		return true
	}

	return false
}
