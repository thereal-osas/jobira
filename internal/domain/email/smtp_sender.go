package email

import (
	"context"
	"fmt"
	"net/smtp"
)

type SMTPConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	From     string
}

type SMTPSender struct {
	cfg SMTPConfig
}

func NewSMTPSender(cfg SMTPConfig) *SMTPSender {
	return &SMTPSender{cfg: cfg}
}

func (s *SMTPSender) Send(ctx context.Context, message EmailMessage) error {
	if s.cfg.Host == "" || s.cfg.Port == "" || s.cfg.From == "" {
		return ErrSendFailed
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	auth := smtp.PlainAuth(
		"",
		s.cfg.Username,
		s.cfg.Password,
		s.cfg.Host,
	)

	rawMessage := []byte(
		"From: " + s.cfg.From + "\r\n" +
			"To: " + message.To + "\r\n" +
			"Subject: " + message.Subject + "\r\n" +
			"MIME-Version: 1.0\r\n" +
			"Content-Type: text/plain; charset=\"UTF-8\"\r\n" +
			"\r\n" +
			message.Body,
	)

	addr := fmt.Sprintf("%s:%s", s.cfg.Host, s.cfg.Port)

	if err := smtp.SendMail(
		addr,
		auth,
		s.cfg.From,
		[]string{message.To},
		rawMessage,
	); err != nil {
		return err
	}

	return nil
}
