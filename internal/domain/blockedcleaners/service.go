package blockedcleaners

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

func (s *Service) Block(ctx context.Context, clientID uint, cleanerID uint, req BlockedCleanerRequest) (*BlockedCleaner, error) {
	req.Reason = strings.TrimSpace(req.Reason)

	if clientID == 0 || cleanerID == 0 {
		return nil, ErrInvalidInput
	}

	if clientID == cleanerID {
		return nil, ErrInvalidInput
	}

	exists, err := s.repo.Exists(ctx, clientID, cleanerID)
	if err != nil {
		return nil, err
	}

	if exists {
		return nil, ErrAlreadyBlocked
	}

	block := &BlockedCleaner{
		ClientID: clientID,
		CleanerID: cleanerID,
		Reason: req.Reason,
	}

	if err := s.repo.Create(ctx, block); err != nil {
		return nil, err
	}

	return block, nil
}

func (s *Service) Unblock(ctx context.Context, clientID uint, cleanerID uint) error {
	if clientID == 0 || cleanerID == 0 {
		return ErrInvalidInput
	}

	return s.repo.Delete(ctx, clientID, cleanerID)
}

func (s *Service) ListMine(ctx context.Context, clientID uint) ([]BlockedCleaner, error) {
	if clientID == 0 {
		return nil, ErrInvalidInput
	}

	return s.repo.ListByClientID(ctx, clientID)
}

func (s *Service) Exists(ctx context.Context, clientID uint, cleanerID uint) (bool, error) {
	if clientID == 0 || cleanerID == 0 {
		return false, ErrInvalidInput
	}

	return s.repo.Exists(ctx, clientID, cleanerID)
}