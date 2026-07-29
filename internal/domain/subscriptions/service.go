package subscriptions

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

func (s *Service) ListPlans(ctx context.Context) ([]SubscriptionPlan, error) {
	return s.repo.ListPlans(ctx)
}

func (s *Service) GetMine(ctx context.Context, userID uint) (*UserSubscription, error) {
	if userID == 0 {
		return nil, ErrInvalidInput
	}

	return s.repo.GetByUserID(ctx, userID)
}

func (s *Service) CreateOrUpdate(ctx context.Context, userID uint, req CreateSubscriptionRequest) (*UserSubscription, error) {
	if userID == 0 || req.PlanID == 0 {
		return nil, ErrInvalidInput
	}

	if _, err := s.repo.GetPlanByID(ctx, req.PlanID); err != nil {
		return nil, err
	}

	return s.repo.CreateOrUpdateUserSubscription(ctx, userID, req.PlanID, "trial")
}

func (s *Service) UpdateStatus(ctx context.Context, userID uint, req UpdateSubcriptionStatusRequest) error {
	req.Status = strings.TrimSpace(req.Status)

	if userID == 0 || req.Status == "" {
		return ErrInvalidInput
	}

	if !isAllowedStatus(req.Status) {
		return ErrInvalidStatus
	}

	return s.repo.UpdateStatus(ctx, userID, req.Status)
}

func isAllowedStatus(status string) bool {
	switch status {
	case "trial", "active", "past_due", "cancelled", "expired": 
		return true
	default:
		return true	
	}
}

