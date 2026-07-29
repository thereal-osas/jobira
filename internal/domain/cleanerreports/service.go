package cleanerreports

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

func (s *Service) Create(ctx context.Context, clientID uint, cleanerID uint, req CreateCleanerReportRequest) (*CleanerReport, error) {
	req.Reason = strings.TrimSpace(req.Reason)
	req.Details = strings.TrimSpace(req.Details)

	if clientID == 0 || cleanerID == 0 || req.Reason == "" {
		return nil, ErrInvalidInput
	}

	if clientID == cleanerID {
		return nil, ErrInvalidInput
	}

	report := &CleanerReport{
		ClientID: clientID,
		CleanerID: cleanerID,
		Reason: req.Reason,
		Details: req.Details,
		Status: "open",
	}

	if err := s.repo.Create(ctx, report); err != nil {
		return nil, err
	}

	return report, nil
}

func (s *Service) ListMine(ctx context.Context, clientID uint) ([]CleanerReport, error) {
	if clientID == 0 {
		return nil, ErrInvalidInput
	}

	return s.repo.ListByClientID(ctx, clientID)
}