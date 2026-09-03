package companyaccounts

import (
	"context"
	"errors"
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

func (s *Service) AddMember(ctx context.Context, companyID uint, requesterID uint, req AddMemberRequest) error {
	if companyID == 0 || requesterID == 0 || req.UserID == 0 {
		return ErrInvalidInput
	}

	req.Role = strings.ToLower(
		strings.TrimSpace(
			req.Role,
		),
	)

	if req.Role == "" {
		req.Role = "cleaner"
	}

	if !validCompnayRole(
		req.Role,
	) {
		return ErrInvalidInput
	}

	if req.Role == "owner" {
		return ErrForbidden
	}

	if err := s.requireCompanyManager(
		ctx,
		companyID,
		requesterID,
	); err != nil {
		return err
	}

	if req.Role == "cleaner" {
		seatLimit, err :=
			s.repo.GetCleanerSeatLimit(
				ctx,
				companyID,
			)
		if err != nil {
			return err
		}

		seatsUsed, err :=
			s.repo.CountActiveCleanerSeats(
				ctx,
				companyID,
			)
		if err != nil {
			return err
		}

		if seatsUsed >=
			seatLimit {
			return ErrSeatLimitReached
		}
	}

	return s.repo.AddMember(ctx, companyID, req.UserID, req.Role)
}

func (s *Service) RemoveMember(ctx context.Context, companyID uint, requesterID uint, userID uint) error {
	if companyID == 0 || requesterID == 0 || userID == 0 {
		return ErrInvalidInput
	}

	if err := s.requireCompanyManager(
		ctx,
		companyID,
		requesterID,
	); err != nil {
		return err
	}

	company, err := s.repo.GetByID(
		ctx,
		companyID,
	)
	if err != nil {
		return err
	}

	if company.OwnerID == userID {
		return ErrForbidden
	}

	member, err :=
		s.repo.GetMember(
			ctx,
			companyID,
			userID,
		)
	if err != nil {
		return err
	}

	if member.Role == "owner" {
		return ErrForbidden
	}

	return s.repo.RemoveMember(ctx, companyID, userID)
}

func (s *Service) ListMembers(ctx context.Context, companyID uint, requesterID uint) ([]CompanyMember, error) {
	if companyID == 0 ||
		requesterID == 0 {
		return nil,
			ErrInvalidInput
	}

	if err :=
		s.requireCompanyMember(
			ctx,
			companyID,
			requesterID,
		); err != nil {
		return nil, err
	}

	return s.repo.ListMembers(
		ctx,
		companyID,
	)
}
func (s *Service) UpdateMemberStatus(ctx context.Context, companyID uint, requesterID uint, userID uint, req UpdateMemberStatusRequest) error {
	if companyID == 0 || requesterID == 0 || userID == 0 {
		return ErrInvalidInput
	}

	req.Status =
		strings.ToLower(
			strings.TrimSpace(
				req.Status,
			),
		)

	if req.Status != "active" &&
		req.Status != "inactive" {
		return ErrInvalidInput
	}

	if err :=
		s.requireCompanyManager(
			ctx,
			companyID,
			requesterID,
		); err != nil {
		return err
	}

	company, err :=
		s.repo.GetByID(
			ctx,
			companyID,
		)
	if err != nil {
		return err
	}

	if company.OwnerID == userID {
		return ErrForbidden
	}

	member, err :=
		s.repo.GetMember(
			ctx,
			companyID,
			userID,
		)
	if err != nil {
		return err
	}

	if member.Role == "owner" {
		return ErrForbidden
	}

	if member.Role == "cleaner" &&
		req.Status == "active" &&
		member.Status != "active" {
		seatLimit, err :=
			s.repo.GetCleanerSeatLimit(
				ctx,
				companyID,
			)
		if err != nil {
			return err
		}

		seatUsed, err :=
			s.repo.CountActiveCleanerSeats(
				ctx,
				companyID,
			)
		if err != nil {
			return err
		}

		if seatUsed >=
			seatLimit {
			return ErrSeatLimitReached
		}
	}

	return s.repo.UpdateMemberStatus(
		ctx,
		companyID,
		userID,
		req.Status,
	)
}

func (s *Service) UpdateMemberRole(ctx context.Context, companyID uint, requesterID uint, userID uint, req UpdateMemberRolesRequest) error {
	if companyID == 0 || requesterID == 0 || userID == 0 {
		return ErrInvalidInput
	}

	req.Role =
		strings.ToLower(
			strings.TrimSpace(
				req.Role,
			),
		)

	if req.Role != "admin" &&
		req.Role != "cleaner" {
		return ErrInvalidInput
	}

	if err :=
		s.requireCompanyManager(
			ctx,
			companyID,
			requesterID,
		); err != nil {
		return err
	}

	company, err :=
		s.repo.GetByID(
			ctx,
			companyID,
		)
	if err != nil {
		return err
	}

	if company.OwnerID == userID {
		return ErrForbidden
	}

	member, err := s.repo.GetMember(
		ctx,
		companyID,
		userID,
	)

	if err != nil {
		return err
	}

	if member.Role == "owner" {
		return ErrForbidden
	}

	if req.Role == "clenaer" &&
		member.Role != "cleaner" &&
		member.Status == "active" {
		seatLimit, err :=
			s.repo.GetCleanerSeatLimit(
				ctx,
				companyID,
			)
		if err != nil {
			return err
		}

		seatsUsed, err :=
			s.repo.CountActiveCleanerSeats(
				ctx,
				companyID,
			)
		if err != nil {
			return err
		}

		if seatsUsed >=
			seatLimit {
			return ErrSeatLimitReached
		}
	}

	return s.repo.UpdateMemberRole(
		ctx,
		companyID,
		userID,
		req.Role,
	)
}

func (s *Service) GetDashboard(ctx context.Context, companyID uint, requesterID uint) (*CompanyDashboard, error) {
	if companyID == 0 || requesterID == 0 {
		return nil, ErrInvalidInput
	}

	if err := s.requireCompanyManager(
		ctx,
		companyID,
		requesterID,
	); err != nil {
		return nil, err
	}

	company, err :=
		s.repo.GetByID(
			ctx,
			companyID,
		)
	if err != nil {
		return nil, err
	}

	members, err :=
		s.repo.ListMembers(
			ctx,
			companyID,
		)
	if err != nil {
		return nil, err
	}

	seatLimit, err :=
		s.repo.GetCleanerSeatLimit(
			ctx,
			companyID,
		)
	if err != nil {
		return nil, err
	}

	activeCleaners, err :=
		s.repo.CountMembersByRolesAndStatus(
			ctx,
			companyID,
			"cleaner",
			"active",
		)
	if err != nil {
		return nil, err
	}

	inactiveCleaners, err :=
		s.repo.CountMembersByRolesAndStatus(
			ctx,
			companyID,
			"cleaner",
			"inactive",
		)
	if err != nil {
		return nil, err
	}

	adminCount, err :=
		s.repo.CountMembersByRolesAndStatus(
			ctx,
			companyID,
			"admin",
			"",
		)
	if err != nil {
		return nil, err
	}

	seatsRemaining :=
		seatLimit -
			activeCleaners

	if seatsRemaining < 0 {
		seatsRemaining = 0
	}

	return &CompanyDashboard{
		Company: *company,

		TotalMembers: len(members),

		ActiveCleaners: activeCleaners,

		InactiveCleaners: inactiveCleaners,

		AdminCount: adminCount,

		SeatUsage: CompanySeatUsage{
			CompanyID: companyID,

			SeatLimit: seatLimit,

			SeatsUsed: activeCleaners,

			SeatsRemaining: seatsRemaining,
		},
	}, nil
}

func (s *Service) requireCompanyMember(ctx context.Context, companyID uint, userID uint,
) error {
	if companyID == 0 ||
		userID == 0 {
		return ErrInvalidInput
	}

	_, err := s.repo.GetMember(
		ctx,
		companyID,
		userID,
	)

	if errors.Is(
		err,
		ErrMemberNotFound,
	) {
		return ErrForbidden
	}

	if err != nil {
		return err
	}

	return nil
}

func (s *Service) requireCompanyManager(ctx context.Context, companyID uint, userID uint) error {
	member, err :=
		s.repo.GetMember(
			ctx,
			companyID,
			userID,
		)

	if err != nil {
		if err == ErrMemberNotFound {
			return ErrForbidden
		}

		return err
	}

	if member.Role != "owner" &&
		member.Role != "admin" {
		return ErrForbidden
	}

	if member.Status != "active" {
		return ErrForbidden
	}

	return nil
}

func validCompnayRole(
	role string,
) bool {
	switch role {
	case "owner",
		"admin",
		"cleaner":
		return true

	default:
		return false
	}
}
