package jobalerts

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

func (s *Service) Create(ctx context.Context, userID uint, req CreateJobAlertRequest) (*JobAlert, error) {
	req.Location = strings.TrimSpace(req.Location)
	req.JobType = strings.TrimSpace(req.JobType)

	if userID == 0 || req.Location == "" || req.JobType == "" {
		return nil, ErrInvalidInput
	}

	if req.MinimumBudget < 0 {
		return nil, ErrInvalidInput
	}

	alert := &JobAlert{
		UserID:        userID,
		Location:      req.Location,
		JobType:       req.JobType,
		MinimumBudget: req.MinimumBudget,
		IsActive:      true,
	}

	if err := s.repo.Create(ctx, alert); err != nil {
		return nil, err
	}

	return alert, nil
}

func (s *Service) GetByID(ctx context.Context, alertID uint, userID uint, role string) (*JobAlert, error) {
	if alertID == 0 || userID == 0 {
		return nil, ErrInvalidInput
	}

	alert, err := s.repo.GetByID(ctx, alertID)
	if err != nil {
		return nil, err
	}

	if role != "admin" && alert.UserID != userID {
		return nil, ErrForbidden
	}

	return alert, nil
}

func (s *Service) ListMine(ctx context.Context, userID uint) ([]JobAlert, error) {
	if userID == 0 {
		return nil, ErrInvalidInput
	}

	return s.repo.ListByUserID(ctx, userID)
}

func (s *Service) Update(ctx context.Context, alertID uint, userID uint, role string, req UpdateJobAlertRequest) (*JobAlert, error) {
	if alertID == 0 || userID == 0 {
		return nil, ErrInvalidInput
	}

	alert, err := s.repo.GetByID(ctx, alertID)
	if err != nil {
		return nil, err
	}

	if role != "admin" && alert.UserID != userID {
		return nil, ErrForbidden
	}

	req.Location = strings.TrimSpace(req.Location)
	req.JobType = strings.TrimSpace(req.JobType)

	if req.Location == "" || req.JobType == "" {
		return nil, ErrInvalidInput
	}

	if req.MinimumBudget < 0 {
		return nil, ErrInvalidInput
	}

	alert.Location = req.Location
	alert.JobType = req.JobType
	alert.MinimumBudget = req.MinimumBudget
	alert.IsActive = req.IsActive

	if err := s.repo.Update(ctx, alert); err != nil {
		return nil, err
	}

	return s.repo.GetByID(ctx, alertID)
}

func (s *Service) Delete(ctx context.Context, alertID uint, userID uint, role string) error {
	if alertID == 0 || userID == 0 {
		return ErrInvalidInput
	}

	alert, err := s.repo.GetByID(ctx, alertID)
	if err != nil {
		return err
	}

	if role != "admin" && alert.UserID != userID {
		return ErrForbidden
	}

	return s.repo.Delete(ctx, alertID)
}
