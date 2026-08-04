package preferredcleaners

import (
	"context"
)

type BlockChecker interface {
	IsBlocked(
		ctx context.Context,
		clientID uint,
		cleanerID uint,
	) (bool, error)
}

type Service struct {
	repo         Repository
	blockChecker BlockChecker
}

func NewService(
	repo Repository,
	blockChecker BlockChecker,
) *Service {
	return &Service{
		repo:         repo,
		blockChecker: blockChecker,
	}
}

func (s *Service) Create(ctx context.Context, clientID uint, req CreatePreferredCleanerRequest) (*PreferredCleaners, error) {
	if clientID == 0 || req.CleanerID == 0 {
		return nil, ErrInvalidInput
	}

	if clientID == req.CleanerID {
		return nil, ErrInvalidInput
	}

	if s.blockChecker != nil {
		blocked, err := s.blockChecker.IsBlocked(ctx, clientID, req.CleanerID)
		if err != nil {
			return nil, err
		}

		if blocked {
			return nil, ErrCleanerBlocked
		}
	}

	exists, err := s.repo.Exists(ctx, clientID, req.CleanerID)
	if err != nil {
		return nil, err
	}

	if exists {
		return nil, ErrAlreadyPreferred
	}

	preferred := &PreferredCleaners{
		ClientID:  clientID,
		CleanerID: req.CleanerID,
	}

	if err := s.repo.Create(ctx, preferred); err != nil {
		return nil, err
	}

	return preferred, nil
}

func (s *Service) ListMine(ctx context.Context, clientID uint) ([]PreferredCleaners, error) {
	if clientID == 0 {
		return nil, ErrInvalidInput
	}

	return s.repo.ListByClientID(ctx, clientID)
}

func (s *Service) Remove(ctx context.Context, clientID uint, cleanerID uint) error {
	if clientID == 0 || cleanerID == 0 {
		return ErrInvalidInput
	}

	return s.repo.Delete(ctx, clientID, cleanerID)
}
