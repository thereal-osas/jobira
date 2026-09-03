package email

import (
	"strings"
	"testing"
)

func TestNewApplicationEmail(t *testing.T) {
	message := NewApplicationEmail(
		"client@example.com",
		"End of tenancy clean",
	)

	if message.To != "client@example.com" {
		t.Fatalf(
			"expected recipient client@example.com, got %q",
			message.To,
		)
	}

	if message.Subject != "New application received" {
		t.Fatalf(
			"unexpected subject: %q",
			message.Subject,
		)
	}

	if !strings.Contains(
		message.Body,
		"End of tenancy clean",
	) {
		t.Fatalf(
			"expected job title in body, got %q",
			message.Body,
		)
	}
}

func TestApplicationStatusEmail(t *testing.T) {
	message := ApplicationStatusEmail(
		"cleaner@example.com",
		"accepted",
	)

	if message.To != "cleaner@example.com" {
		t.Fatalf(
			"unexpected recipient: %q",
			message.To,
		)
	}

	if message.Subject != "Application status updated" {
		t.Fatalf(
			"unexpected subject: %q",
			message.Subject,
		)
	}

	if !strings.Contains(
		message.Body,
		"accepted",
	) {
		t.Fatalf(
			"expected status in body, got %q",
			message.Body,
		)
	}
}

func TestBookingConfirmedEmail(t *testing.T) {
	message := BookingConfirmedEmail(
		"user@example.com",
	)

	if message.To != "user@example.com" {
		t.Fatalf(
			"unexpected recipient: %q",
			message.To,
		)
	}

	if message.Subject != "Booking confirmed" {
		t.Fatalf(
			"unexpected subject: %q",
			message.Subject,
		)
	}

	if !strings.Contains(
		message.Body,
		"booking has been confirmed",
	) {
		t.Fatalf(
			"unexpected body: %q",
			message.Body,
		)
	}
}

func TestBookingCompletedEmail(t *testing.T) {
	message := BookingCompletedEmail(
		"user@example.com",
	)

	if message.To != "user@example.com" {
		t.Fatalf(
			"unexpected recipient: %q",
			message.To,
		)
	}

	if message.Subject != "Booking completed" {
		t.Fatalf(
			"unexpected subject: %q",
			message.Subject,
		)
	}

	if !strings.Contains(
		message.Body,
		"marked as completed",
	) {
		t.Fatalf(
			"unexpected body: %q",
			message.Body,
		)
	}
}

func TestJobInvitationEmail(t *testing.T) {
	message := JobInvitationEmail(
		"cleaner@example.com",
		"Can you cover this shift?",
	)

	if message.To != "cleaner@example.com" {
		t.Fatalf(
			"unexpected recipient: %q",
			message.To,
		)
	}

	if message.Subject != "You have received a job invitation" {
		t.Fatalf(
			"unexpected subject: %q",
			message.Subject,
		)
	}

	if !strings.Contains(
		message.Body,
		"Can you cover this shift?",
	) {
		t.Fatalf(
			"expected invitation message in body, got %q",
			message.Body,
		)
	}
}

func TestSubscriptionLimitReachedEmail(t *testing.T) {
	message := SubscriptionLimitReachedEmail(
		"user@example.com",
	)

	if message.To != "user@example.com" {
		t.Fatalf(
			"unexpected recipient: %q",
			message.To,
		)
	}

	if message.Subject != "Jobira usage limit reached" {
		t.Fatalf(
			"unexpected subject: %q",
			message.Subject,
		)
	}

	if !strings.Contains(
		message.Body,
		"continue using Jobira",
	) {
		t.Fatalf(
			"unexpected body: %q",
			message.Body,
		)
	}
}

func TestBookingCancelledEmail(t *testing.T) {
	message := BookingCancelledEmail(
		"user@example.com",
		"Cleaner unavailable",
	)

	if message.To != "user@example.com" {
		t.Fatalf(
			"unexpected recipient: %q",
			message.To,
		)
	}

	if message.Subject != "Booking cancelled" {
		t.Fatalf(
			"unexpected subject: %q",
			message.Subject,
		)
	}

	if !strings.Contains(
		message.Body,
		"Cleaner unavailable",
	) {
		t.Fatalf(
			"expected cancellation reason in body, got %q",
			message.Body,
		)
	}

	if !strings.Contains(
		message.Body,
		"Log in to Jobira to view the booking details.",
	) {
		t.Fatalf(
			"unexpected booking link text: %q",
			message.Body,
		)
	}
}
