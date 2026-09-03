package adminsubscriptions

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

func (s *Service) ListPlans(ctx context.Context) ([]AdminSubscriptionPlan, error) {
	return s.repo.ListPlans(ctx)
}
func (s *Service) CreatePlan(ctx context.Context, req CreatePlanRequest) (*AdminSubscriptionPlan, error) {
	req.Name = strings.TrimSpace(req.Name)
	req.RoleType = strings.TrimSpace(req.RoleType)
	req.BillingInterval = strings.TrimSpace(req.BillingInterval)

	if req.Name == "" || req.RoleType == "" || req.BillingInterval == "" {
		return nil, ErrInvalidInput
	}

	if req.PricePence < 0 || req.ApplicationLimit < 0 || req.JobPostLimit < 0 || req.CleanerSeatLimit < 1 {
		return nil, ErrInvalidInput
	}

	plan := &AdminSubscriptionPlan{
		Name:             req.Name,
		RoleType:         req.RoleType,
		PricePence:       req.PricePence,
		BillingInterval:  req.BillingInterval,
		ApplicationLimit: req.ApplicationLimit,
		JobPostLimit:     req.JobPostLimit,
		CleanerSeatLimit: req.CleanerSeatLimit,
		IsActive:         true,
	}

	if err := s.repo.CreatePlan(ctx, plan); err != nil {
		return nil, err
	}

	return plan, nil
}

func (s *Service) UpdatePlan(ctx context.Context, planID uint, req UpdatePlanRequest) error {
	req.Name = strings.TrimSpace(req.Name)
	req.RoleType = strings.TrimSpace(req.RoleType)
	req.BillingInterval = strings.TrimSpace(req.BillingInterval)

	if planID == 0 || req.Name == "" || req.RoleType == "" || req.BillingInterval == "" {
		return ErrInvalidInput
	}

	if req.PricePence < 0 || req.ApplicationLimit < 0 || req.JobPostLimit < 0 || req.CleanerSeatLimit < 1 {
		return ErrInvalidInput
	}

	plan := &AdminSubscriptionPlan{
		ID:               planID,
		Name:             req.Name,
		RoleType:         req.RoleType,
		PricePence:       req.PricePence,
		BillingInterval:  req.BillingInterval,
		ApplicationLimit: req.ApplicationLimit,
		JobPostLimit:     req.JobPostLimit,
		CleanerSeatLimit: req.CleanerSeatLimit,
		IsActive:         req.IsActive,
	}

	return s.repo.UpdatePlan(ctx, plan)

}
func (s *Service) DisablePlan(ctx context.Context, planID uint) error {
	if planID == 0 {
		return ErrInvalidInput
	}

	return s.repo.DisablePlan(ctx, planID)
}

func (s *Service) ListUserSubscriptions(ctx context.Context) ([]AdminUserSubscription, error) {
	return s.repo.ListUserSubscriptions(ctx)
}

func (s *Service) GetUserSubscriptions(ctx context.Context, userID uint) (*AdminUserSubscription, error) {
	if userID == 0 {
		return nil, ErrInvalidInput
	}

	return s.repo.GetUserSubscriptions(ctx, userID)
}

func (s *Service) CreateOrUpdateUserSubscriptions(ctx context.Context, userID uint, req UpdateUserSubscriptionRequest) error {
	req.Status = strings.TrimSpace(req.Status)

	if userID == 0 || req.PlanID == 0 || req.Status == "" {
		return ErrInvalidInput
	}

	if !isAllowedStatus(req.Status) {
		return ErrInvalidStatus
	}

	return s.repo.CreateOrUpdateUserSubscription(ctx, userID, req.PlanID, req.Status)
}

func isAllowedStatus(status string) bool {
	switch status {
	case "trial", "active", "past_due", "cancelled", "expired":
		return true
	default:
		return false
	}
}
