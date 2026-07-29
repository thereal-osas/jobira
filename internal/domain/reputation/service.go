package reputation

import (
	"context"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetByCleanerID(ctx context.Context, cleanerID uint) (*CleanerReputation, error) {
	if cleanerID == 0 {
		return nil, ErrInvalidInput
	}

	if err := s.repo.Refresh(ctx, cleanerID); err != nil {
		return nil, err
	}

	reputation, err := s.repo.GetByCleanerID(ctx, cleanerID)
	if err != nil {
		return nil, err
	}

	reputation.Badge = calculateBadge(reputation)

	return reputation, nil
}

func (s *Service) Refresh(ctx context.Context, cleanerID uint) (*CleanerReputation, error) {
	if cleanerID == 0 {
		return nil, ErrInvalidInput
	}

	if err := s.repo.Refresh(ctx, cleanerID); err != nil {
		return nil, err
	}

	reputation, err := s.repo.GetByCleanerID(ctx, cleanerID)
	if err != nil {
		return nil, err
	}

	badge := calculateBadge(reputation)

	if err := s.repo.UpdateBadge(ctx, cleanerID, badge); err != nil {
		return nil, err
	}

	reputation.Badge = badge

	return reputation, nil
}

func calculateBadge(reputation *CleanerReputation) string {
	if reputation == nil {
		return "New Cleaner"
	}

	if reputation.TotalReviews >= 20 &&
		reputation.AverageRating >= 4.9 &&
		reputation.RecommendationPercentage >= 95 {
		return "Elite Cleaner"
	}

	if reputation.TotalReviews >= 10 &&
		reputation.AverageRating >= 4.8 &&
		reputation.RecommendationPercentage >= 90 {
		return "Top Rated"
	}

	if reputation.CompletedJobs >= 50 {
		return "Experienced Cleaner"
	}

	if reputation.CompletedJobs >= 1 {
		return "Active Cleaner"
	}

	return "New Cleaner"

}
