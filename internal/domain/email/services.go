package email

import (
	"context"
	"strings"
)

type Sender interface {
	Send(ctx context.Context, message EmailMessage) error
}

type Service struct {
	sender Sender
}

func NewService(sender Sender) *Service {
	return &Service{sender: sender}
}

func (s *Service) Send(ctx context.Context, message EmailMessage) error {
	message.To = strings.TrimSpace(message.To)
	message.Subject = strings.TrimSpace(message.Subject)
	message.Body = strings.TrimSpace(message.Body)

	if message.To == "" || message.Subject == "" || message.Body == "" {
		return ErrInvalidInput
	}

	if s.sender == nil {
		return ErrSendFailed
	}

	return s.sender.Send(ctx, message)
}
