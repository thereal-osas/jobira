package notifications

import (
	"context"
	"strings"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) Create(ctx context.Context, req CreateNotificationsRequest) (*Notification, error) {
	req.Title = strings.TrimSpace(req.Title)
	req.Message = strings.TrimSpace(req.Message)
	req.Type = strings.TrimSpace(req.Type)

	if req.UserID == 0 {
		return nil, ErrInvalidInput
	}

	if req.Title == "" || req.Message == "" || req.Type == "" {
		return nil, ErrInvalidInput
	}

	notifcation := &Notification{
		UserID:  req.UserID,
		Title:   req.Title,
		Message: req.Message,
		Type:    req.Type,
		IsRead:  false,
	}

	if err := s.repo.Create(ctx, notifcation); err != nil {
		return nil, err
	}

	return notifcation, nil
}

func (s *Service) ListMine(ctx context.Context, userID uint) ([]Notification, error) {
	if userID == 0 {
		return nil, ErrInvalidInput
	}

	return s.repo.ListByUserID(ctx, userID)
}

func (s *Service) ListUnreadMine(ctx context.Context, userID uint) ([]Notification, error) {
	if userID == 0 {
		return nil, ErrInvalidInput
	}

	return s.repo.ListUnreadByUserID(ctx, userID)
}

func (s *Service) MarkAsRead(ctx context.Context, notificationID uint, userID uint) error {
	if notificationID == 0 || userID == 0 {
		return ErrInvalidInput
	}

	return s.repo.MarkAsRead(ctx, notificationID, userID)
}

func (s *Service) CountUnreadMine(ctx context.Context, userID uint) (int, error) {
	if userID == 0 {
		return 0, ErrInvalidInput
	}

	return s.repo.CountUnreadByUserID(ctx, userID)
}

func (s *Service) MarkAllAsRead(ctx context.Context, userID uint) error {
	if userID == 0 {
		return ErrInvalidInput
	}

	return s.repo.MarkAllAsRead(ctx, userID)
}
