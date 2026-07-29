package favorites

import (
	"context"

	blockedcleanersdomain "github.com/rodrigueghenda/jobira/internal/domain/blockedcleaners"
	notificationsdomain "github.com/rodrigueghenda/jobira/internal/domain/notifications"
)

type Service struct {
	repo Repository
	notificationsService *notificationsdomain.Service
	blockChecker 		 *blockedcleanersdomain.Checker
}

func NewService(repo Repository, notificationsService *notificationsdomain.Service, blockChecker *blockedcleanersdomain.Checker) *Service {
	return &Service{
		repo: repo,
		notificationsService: notificationsService,
		blockChecker: blockChecker,
	}
}

func (s *Service) Save(ctx context.Context, clientID uint, req CreateFavoriteRequest) (*FavoriteCleaner, error) {
	if s.blockChecker != nil {
		blocked, err := s.blockChecker.IsBlocked(ctx, clientID, req.CleanerID)
		if err != nil {
			return nil, err
		}

		if blocked {
			return nil, ErrCleanerBlocked
		}
	}

	if clientID == 0 || req.CleanerID == 0 {
		return nil, ErrInvalidInput
	}

	if clientID == req.CleanerID {
		return nil, ErrInvalidInput
	}

	exists, err := s.repo.Exists(ctx, clientID, req.CleanerID)
	if err != nil {
		return nil, err
	}

	if exists {
		return nil, ErrAlreadySaved
	}

	favorite := &FavoriteCleaner{
		ClientID: clientID,
		CleanerID: req.CleanerID,
	}

	if err := s.repo.Create(ctx, favorite); err != nil {
		return nil, err
	}

	if s.notificationsService != nil {
		_, _ = s.notificationsService.Create(ctx, notificationsdomain.CreateNotificationsRequest{
			UserID: req.CleanerID,
			Title: "New Favourite",
			Message: "A Client added you to their favourite cleaners list",
			Type: "favourite",
		})
	}

	return favorite, nil
}

func (s *Service) ListMine(ctx context.Context, clientID uint) ([]FavoriteCleaner, error) {
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