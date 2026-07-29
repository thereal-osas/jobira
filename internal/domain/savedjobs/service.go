package savedjobs

import "context"

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Save(ctx context.Context, userID uint, req SavedJbRequest) (*SavedJob, error) {
	if userID == 0 || req.JobID == 0 {
		return nil, ErrInvalidInput
	}

	savedJob := &SavedJob{
		UserID: userID,
		JobID:  req.JobID,
	}

	if err := s.repo.Save(ctx, savedJob); err != nil {
		return nil, err
	}

	return savedJob, nil
}

func (s *Service) ListMine(ctx context.Context, userID uint) ([]SavedJob, error) {
	if userID == 0 {
		return nil, ErrInvalidInput
	}

	return s.repo.ListByUserID(ctx, userID)
}

func (s *Service) Delete(ctx context.Context, userID uint, jobID uint) error {
	if userID == 0 || jobID == 0 {
		return ErrInvalidInput
	}

	return s.repo.Delete(ctx, userID, jobID)
}
