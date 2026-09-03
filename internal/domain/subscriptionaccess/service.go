package subscriptionaccess

import "context"

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetCleanerAccessStatus(ctx context.Context, userID uint) (*CleanerAccessStatus, error) {
	if userID == 0 {
		return nil, ErrInvalidInput
	}

	return s.repo.GetCleanerAccessStatus(ctx, userID)
}

func (s *Service) GetClientAccessStatus(ctx context.Context, userID uint) (*ClientAccessStatus, error) {
	if userID == 0 {
		return nil, ErrInvalidInput
	}

	return s.repo.GetClientAccessStatus(ctx, userID)
}

func (s *Service) EnsureCanPostJob(ctx context.Context, userID uint) error {
	status, err := s.GetClientAccessStatus(ctx, userID)
	if err != nil {
		return err
	}

	if !status.CanPostJob {
		return ErrDailyJobPostLimit
	}

	return nil
}

func (s *Service) IncrementApplicationsToday(ctx context.Context, userID uint) error {
	if userID == 0 {
		return ErrInvalidInput
	}

	return s.repo.IncrementApplicationsToday(ctx, userID)
}

func (s *Service) IncrementJobsPostToday(ctx context.Context, userID uint) error {
	if userID == 0 {
		return ErrInvalidInput
	}

	return s.repo.IncrementJobsPostToday(ctx, userID)
}

func (s *Service) EnsureCanApply(ctx context.Context, userID uint) error {
	status, err := s.GetCleanerAccessStatus(ctx, userID)
	if err != nil {
		return err
	}

	if !status.CanApply {
		return ErrDailyApplicationLimit
	}

	return nil
}
