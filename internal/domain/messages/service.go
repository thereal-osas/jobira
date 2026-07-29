package messages

import (
	"context"
	"strings"

notificationsdomain "github.com/rodrigueghenda/jobira/internal/domain/notifications"
	
)

type Service struct {
	repo Repository
	notificationsService  *notificationsdomain.Service
}

func NewService(repo Repository, notificationsService *notificationsdomain.Service) *Service {
	return &Service{
		repo: repo,
		notificationsService: notificationsService,
	}
}

func (s *Service) Send(
	ctx context.Context, 
	jobID uint,
	senderID uint,
	req SendMessageRequest,  
) (*Message, error) {
	req.Content = strings.TrimSpace(req.Content)

	if jobID == 0 || senderID == 0 || req.ReceiverID == 0 {
		return nil, ErrInvalidInput
	}

	if req.Content == "" {
		return nil, ErrInvalidInput
	}

	isParticipant, err := s.repo.IsJobParticipant(ctx, jobID, senderID)
	if err != nil {
		return nil, err
	}

	if !isParticipant {
		return nil, ErrForbidden
	}

	message := &Message{
		JobID: jobID,
		SenderID: senderID,
		ReceiverID: req.ReceiverID,
		Content: req.Content,
	}

	if err := s.repo.Create(ctx, message); err != nil {
		return nil, err
	}

	if s.notificationsService != nil {
		_, _ = s.notificationsService.Create(ctx, notificationsdomain.CreateNotificationsRequest{
			UserID: req.ReceiverID,
			Title: "New Message",
			Message: "You received a new message.",
			Type: "message",
		})
	}

	return message, nil


}

func (s *Service) ListConversation(
	ctx context.Context,
	jobID uint, 
	userID uint,
) ([]Message, error) {

	if jobID == 0 || userID == 0 {
		return nil, ErrInvalidInput
	}

	isParticipant, err := s.repo.IsJobParticipant(ctx, jobID, userID)
	if err != nil {
		return nil, err
	}

	if !isParticipant {
		return nil, ErrForbidden
	}

	return s.repo.ListByJobID(ctx, jobID)
}