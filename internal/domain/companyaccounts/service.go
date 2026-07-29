package companyaccounts

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

func (s *Service) CreateCompany(ctx context.Context, ownerID uint, req CreateCompanyRequest) (*Company, error) {
	req.Name = strings.TrimSpace(req.Name)
	req.Description = strings.TrimSpace(req.Description)

	if ownerID == 0 || req.Name == "" {
		return nil, ErrInvalidInput
	}

	company := &Company{
		OwnerID:     ownerID,
		Name:        req.Name,
		Description: req.Description,
	}

	if err := s.repo.CreateCompany(ctx, company); err != nil {
		return nil, err
	}

	if err := s.repo.AddMember(ctx, company.ID, ownerID, "owner"); err != nil {
		return nil, err
	}

	return company, nil
}

func (s *Service) AddMember(ctx context.Context, companyID uint, ownerID uint, role string, req AddMemberRequest) error {
	if companyID == 0 || ownerID == 0 || req.UserID == 0 {
		return ErrInvalidInput
	}

	foundCompany, err := s.repo.GetByID(ctx, companyID)
	if err != nil {
		return err
	}

	if role != "admin" && foundCompany.OwnerID != ownerID {
		return ErrForbidden
	}

	seatLimit, err := s.repo.GetCleanerSeatLimit(ctx, companyID)
	if err != nil {
		return err
	}

	memberCount, err := s.repo.CountMembers(ctx, companyID)
	if err != nil {
		return err
	}

	if memberCount >= seatLimit {
		return ErrSeatLimitReached
	}

	return s.repo.AddMember(ctx, companyID, req.UserID, role)
}

func (s *Service) RemoveMember(ctx context.Context, companyID uint, userID uint) error {
	if companyID == 0 || userID == 0 {
		return ErrInvalidInput
	}

	return s.repo.RemoveMember(ctx, companyID, userID)
}

func (s *Service) ListMembers(ctx context.Context, companyID uint) ([]CompanyMember, error) {
	if companyID == 0 {
		return nil, ErrInvalidInput
	}

	return s.repo.ListMembers(ctx, companyID)
}
