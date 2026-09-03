package email

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestNewNoopSender(t *testing.T) {
	sender := NewNoopSender()

	if sender == nil {
		t.Fatal("expected sender")
	}
}

func TestNoopSender_Send_Success(t *testing.T) {
	sender := NewNoopSender()

	err := sender.Send(
		context.Background(),
		EmailMessage{
			To:      "user@example.com",
			Subject: "Test email",
			Body:    "Test body",
		},
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNoopSender_Send_ContextCancelled(
	t *testing.T,
) {
	sender := NewNoopSender()

	ctx, cancel := context.WithCancel(
		context.Background(),
	)
	cancel()

	err := sender.Send(
		ctx,
		EmailMessage{
			To:      "user@example.com",
			Subject: "Test email",
			Body:    "Test body",
		},
	)

	if !errors.Is(err, context.Canceled) {
		t.Fatalf(
			"expected context.Canceled, got %v",
			err,
		)
	}
}

func TestNoopSender_Send_ContextDeadlineExceeded(
	t *testing.T,
) {
	sender := NewNoopSender()

	ctx, cancel := context.WithDeadline(
		context.Background(),
		time.Now().Add(-time.Second),
	)
	defer cancel()

	err := sender.Send(
		ctx,
		EmailMessage{
			To:      "user@example.com",
			Subject: "Test email",
			Body:    "Test body",
		},
	)

	if !errors.Is(
		err,
		context.DeadlineExceeded,
	) {
		t.Fatalf(
			"expected context.DeadlineExceeded, got %v",
			err,
		)
	}
}
