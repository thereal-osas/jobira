package reviews

import (
	"context"
	"strings"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) Create(ctx context.Context, jobID uint, clientID uint, req CreateReviewRequest) (*Review, error) {
	req.Comment = strings.TrimSpace(req.Comment)

	if jobID == 0 || clientID == 0 {
		return nil, ErrInvalidInput
	}

	if req.Comment == "" {
		return nil, ErrInvalidInput
	}

	if req.Rating < 1 || req.Rating > 5 {
		return nil, ErrInvalidInput
	}

	jobClientID, err := s.repo.GetJobClientID(ctx, jobID)
	if err != nil {
		return nil, err
	}

	if jobClientID != clientID {
		return nil, ErrForbidden
	}

	jobStatus, err := s.repo.GetJobStatus(ctx, jobID)
	if err != nil {
		return nil, err
	}

	if jobStatus != "completed" {
		return nil, ErrJobNotCompleted
	}

	cleanerID, err := s.repo.GetAcceptedCleanerID(ctx, jobID)
	if err != nil {
		return nil, err
	}

	review := &Review{
		BookingID: req.BookingID,
		CleanerID: cleanerID,
		ClientID:  clientID,
		JobID:     jobID,
		Rating:    req.Rating,
		Comment:   req.Comment,
	}

	if jobID == 0 || clientID == 0 || req.BookingID == 0 {
		return nil, ErrInvalidInput
	}

	if err := s.repo.Create(ctx, review); err != nil {
		return nil, err
	}
	return review, nil
}

func (s *Service) ListByCleanerID(ctx context.Context, cleanerID uint) ([]Review, error) {
	if cleanerID == 0 {
		return nil, ErrInvalidInput
	}

	return s.repo.ListByCleanerID(ctx, cleanerID)
}

func (s *Service) ListMine(ctx context.Context, clientID uint) ([]Review, error) {
	if clientID == 0 {
		return nil, ErrInvalidInput
	}

	return s.repo.ListByClientID(ctx, clientID)
}
