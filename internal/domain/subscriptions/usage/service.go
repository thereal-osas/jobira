package usage

import "context"

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) EnsureUsage(ctx context.Context, userID uint) error {
	if userID == 0 {
		return ErrInvalidInput
	}

	return s.repo.CreateIfNotExists(ctx, userID)
}

func (s *Service) GetByUserID(ctx context.Context, userID uint) (*UserUsage, error) {
	if userID == 0 {
		return nil, ErrInvalidInput
	}

	if err := s.repo.CreateIfNotExists(ctx, userID); err != nil {
		return nil, err
	}

	return s.repo.GetByUserID(ctx, userID)
}
func (s *Service) CanApply(ctx context.Context, userID uint) error {
	if userID == 0 {
		return ErrInvalidInput
	}

	if err := s.repo.CreateIfNotExists(ctx, userID); err != nil {
		return err
	}

	foundUsage, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return err
	}

	if !foundUsage.MonetisationEnabled {
		return nil
	}

	access, err := s.repo.GetUserSubscriptionAccess(ctx, userID)
	if err != nil {
		return err
	}

	if access != nil && isSubscriptionAllowed(access.Status) {
		if access.ApplicationLimit <= 0 {
			return ErrApplicationLimitReached
		}

		if foundUsage.ApplicationCount >= access.ApplicationLimit {
			return ErrApplicationLimitReached
		}

		return nil
	}

	if foundUsage.ApplicationCount >= foundUsage.FreeApplicationLimit {
		return ErrApplicationLimitReached
	}

	return nil
}

func (s *Service) CanPostJob(ctx context.Context, userID uint) error {
	if userID == 0 {
		return ErrInvalidInput
	}

	if err := s.repo.CreateIfNotExists(ctx, userID); err != nil {
		return err
	}

	foundUsage, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return err
	}

	if !foundUsage.MonetisationEnabled {
		return nil
	}

	access, err := s.repo.GetUserSubscriptionAccess(ctx, userID)
	if err != nil {
		return err
	}

	if access != nil && isSubscriptionAllowed(access.Status) {
		if access.JobPostLimit <= 0 {
			return ErrJobPostLimitReached
		}

		if foundUsage.JobPostCount >= access.JobPostLimit {
			return ErrJobPostLimitReached
		}

		return nil
	}

	if foundUsage.JobPostCount >= foundUsage.FreeJobPostLimit {
		return ErrJobPostLimitReached
	}

	return nil
}

func (s *Service) IncrementApplicationCount(ctx context.Context, userID uint) error {
	return s.repo.IncrementApplicationCount(ctx, userID)
}
func (s *Service) IncrementJobPostCount(ctx context.Context, userID uint) error {
	return s.repo.IncrementJobPostCount(ctx, userID)
}
func (s *Service) SetMonetisationEnabled(ctx context.Context, userID uint, enabled bool) error {
	return s.repo.SetMonetisationEnabled(ctx, userID, enabled)
}

func isSubscriptionAllowed(status string) bool {
	if status == "active" {
		return true
	}

	if status == "trial" {
		return true
	}

	return false
}
