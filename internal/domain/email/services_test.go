package email

import (
	"context"
	"errors"
	"testing"
)

type mockSender struct {
	sendFn func(context.Context, EmailMessage) error
}

func (m *mockSender) Send(
	ctx context.Context,
	message EmailMessage,
) error {
	if m.sendFn != nil {
		return m.sendFn(ctx, message)
	}

	return nil
}

func TestNewService(t *testing.T) {
	sender := &mockSender{}

	service := NewService(sender)

	if service == nil {
		t.Fatal("expected service")
	}

	if service.sender != sender {
		t.Fatal("expected sender assigned")
	}
}

func TestService_Send_Success(t *testing.T) {
	var received EmailMessage

	sender := &mockSender{
		sendFn: func(
			ctx context.Context,
			message EmailMessage,
		) error {
			received = message
			return nil
		},
	}

	service := NewService(sender)

	err := service.Send(
		context.Background(),
		EmailMessage{
			To:      "  user@example.com  ",
			Subject: "  Booking confirmed  ",
			Body:    "  Your booking is confirmed.  ",
		},
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if received.To != "user@example.com" {
		t.Fatalf(
			"expected trimmed recipient, got %q",
			received.To,
		)
	}

	if received.Subject != "Booking confirmed" {
		t.Fatalf(
			"expected trimmed subject, got %q",
			received.Subject,
		)
	}

	if received.Body != "Your booking is confirmed." {
		t.Fatalf(
			"expected trimmed body, got %q",
			received.Body,
		)
	}
}

func TestService_Send_EmptyRecipient(t *testing.T) {
	service := NewService(&mockSender{})

	err := service.Send(
		context.Background(),
		EmailMessage{
			Subject: "Test",
			Body:    "Body",
		},
	)

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf(
			"expected ErrInvalidInput, got %v",
			err,
		)
	}
}

func TestService_Send_WhitespaceRecipient(t *testing.T) {
	service := NewService(&mockSender{})

	err := service.Send(
		context.Background(),
		EmailMessage{
			To:      "   ",
			Subject: "Test",
			Body:    "Body",
		},
	)

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf(
			"expected ErrInvalidInput, got %v",
			err,
		)
	}
}

func TestService_Send_EmptySubject(t *testing.T) {
	service := NewService(&mockSender{})

	err := service.Send(
		context.Background(),
		EmailMessage{
			To:   "user@example.com",
			Body: "Body",
		},
	)

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf(
			"expected ErrInvalidInput, got %v",
			err,
		)
	}
}

func TestService_Send_WhitespaceSubject(t *testing.T) {
	service := NewService(&mockSender{})

	err := service.Send(
		context.Background(),
		EmailMessage{
			To:      "user@example.com",
			Subject: "   ",
			Body:    "Body",
		},
	)

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf(
			"expected ErrInvalidInput, got %v",
			err,
		)
	}
}

func TestService_Send_EmptyBody(t *testing.T) {
	service := NewService(&mockSender{})

	err := service.Send(
		context.Background(),
		EmailMessage{
			To:      "user@example.com",
			Subject: "Test",
		},
	)

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf(
			"expected ErrInvalidInput, got %v",
			err,
		)
	}
}

func TestService_Send_WhitespaceBody(t *testing.T) {
	service := NewService(&mockSender{})

	err := service.Send(
		context.Background(),
		EmailMessage{
			To:      "user@example.com",
			Subject: "Test",
			Body:    "   ",
		},
	)

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf(
			"expected ErrInvalidInput, got %v",
			err,
		)
	}
}

func TestService_Send_NilSender(t *testing.T) {
	service := NewService(nil)

	err := service.Send(
		context.Background(),
		EmailMessage{
			To:      "user@example.com",
			Subject: "Test",
			Body:    "Body",
		},
	)

	if !errors.Is(err, ErrSendFailed) {
		t.Fatalf(
			"expected ErrSendFailed, got %v",
			err,
		)
	}
}

func TestService_Send_SenderError(t *testing.T) {
	expectedErr := errors.New("smtp failure")

	sender := &mockSender{
		sendFn: func(
			context.Context,
			EmailMessage,
		) error {
			return expectedErr
		},
	}

	service := NewService(sender)

	err := service.Send(
		context.Background(),
		EmailMessage{
			To:      "user@example.com",
			Subject: "Test",
			Body:    "Body",
		},
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected sender error %v, got %v",
			expectedErr,
			err,
		)
	}
}

