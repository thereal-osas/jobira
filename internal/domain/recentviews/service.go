package recentviews

import "context"

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) RecordView(ctx context.Context, clientID uint, cleanerID uint) error {
	if clientID == 0 || cleanerID == 0 {
		return ErrInvalidInput
	}

	if clientID == cleanerID {
		return ErrInvalidInput
	}

	return s.repo.RecordView(ctx, clientID, cleanerID)
}

func (s *Service) ListMine(ctx context.Context, clientID uint) ([]RecentlyViewedCleaner, error) {
	if clientID == 0 {
		return nil, ErrInvalidInput
	}

	return s.repo.ListByClientID(ctx, clientID)
}