package email

import (
	"context"
	"errors"
	"testing"
)

func TestNewSMTPSender(t *testing.T) {
	cfg := SMTPConfig{
		Host:     "smtp.example.com",
		Port:     "587",
		Username: "user",
		Password: "password",
		From:     "noreply@example.com",
	}

	sender := NewSMTPSender(cfg)

	if sender == nil {
		t.Fatal("expected sender")
	}

	if sender.cfg != cfg {
		t.Fatalf(
			"expected config %+v, got %+v",
			cfg,
			sender.cfg,
		)
	}
}

func TestSMTPSender_Send_MissingHost(t *testing.T) {
	sender := NewSMTPSender(
		SMTPConfig{
			Port: "587",
			From: "noreply@example.com",
		},
	)

	err := sender.Send(
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

func TestSMTPSender_Send_MissingPort(t *testing.T) {
	sender := NewSMTPSender(
		SMTPConfig{
			Host: "smtp.example.com",
			From: "noreply@example.com",
		},
	)

	err := sender.Send(
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

func TestSMTPSender_Send_MissingFrom(t *testing.T) {
	sender := NewSMTPSender(
		SMTPConfig{
			Host: "smtp.example.com",
			Port: "587",
		},
	)

	err := sender.Send(
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

func TestSMTPSender_Send_ContextCancelled(
	t *testing.T,
) {
	sender := NewSMTPSender(
		SMTPConfig{
			Host:     "smtp.example.com",
			Port:     "587",
			Username: "user",
			Password: "password",
			From:     "noreply@example.com",
		},
	)

	ctx, cancel := context.WithCancel(
		context.Background(),
	)
	cancel()

	err := sender.Send(
		ctx,
		EmailMessage{
			To:      "user@example.com",
			Subject: "Test",
			Body:    "Body",
		},
	)

	if !errors.Is(err, context.Canceled) {
		t.Fatalf(
			"expected context.Canceled, got %v",
			err,
		)
	}
}

func TestSMTPSender_Send_InvalidServer(
	t *testing.T,
) {
	sender := NewSMTPSender(
		SMTPConfig{
			Host:     "127.0.0.1",
			Port:     "1",
			Username: "user",
			Password: "password",
			From:     "noreply@example.com",
		},
	)

	err := sender.Send(
		context.Background(),
		EmailMessage{
			To:      "user@example.com",
			Subject: "Test",
			Body:    "Body",
		},
	)

	if err == nil {
		t.Fatal(
			"expected SMTP connection error",
		)
	}
}
