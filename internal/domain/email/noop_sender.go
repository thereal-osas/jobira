package email

import (
	"context"
	"log"
)

type NoopSender struct{}

func NewNoopSender() *NoopSender {
	return &NoopSender{}
}

func (s *NoopSender) Send(ctx context.Context, message EmailMessage) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	log.Printf(
		"email noop sender: to=%s subject=%s body=%s",
		message.To,
		message.Subject,
		message.Body,
	)

	return nil
}
