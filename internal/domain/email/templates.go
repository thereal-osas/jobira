package email

import (
	"fmt"
)

func NewApplicationEmail(to string, jobTitle string) EmailMessage {
	return EmailMessage{
		To:      to,
		Subject: "New application received",
		Body: fmt.Sprintf(
			"You have received a new application for your job: %s.\n\nLog in to Jobira to review the application.",
			jobTitle,
		),
	}
}

func ApplicationStatusEmail(to string, status string) EmailMessage {
	return EmailMessage{
		To:      to,
		Subject: "Application status updated",
		Body: fmt.Sprintf(
			"Your application status has been updated to: %s.\n\nLog in to Jobira to view the details.",
			status,
		),
	}
}

func BookingConfirmedEmail(to string) EmailMessage {
	return EmailMessage{
		To:      to,
		Subject: "Booking confirmed",
		Body:    "Your booking has been confirmed.\n\nLog in to Jobira to view the booking details.",
	}
}

func BookingCompletedEmail(to string) EmailMessage {
	return EmailMessage{
		To:      to,
		Subject: "Booking completed",
		Body:    "Your booking has been marked as completed.\n\nThank you for using Jobira",
	}
}

func JobInvitationEmail(to string, message string) EmailMessage {
	return EmailMessage{
		To:      to,
		Subject: "You have received a job invitation",
		Body: fmt.Sprintf(
			"You have received a job invitation.\n\nMessage:\n%s\n\nLog into Jobira to respond.",
			message,
		),
	}
}

func SubscriptionLimitReachedEmail(to string) EmailMessage {
	return EmailMessage{
		To:      to,
		Subject: "Jobira usage limit reached",
		Body:    "You have reached your free usage limit.\n\nPlease upgrade your subscription to continue using Jobira.",
	}
}

func BookingCancelledEmail(to string, reason string) EmailMessage {
	return EmailMessage{
		To:      to,
		Subject: "Booking cancelled",
		Body: fmt.Sprintf(
			"Your booking has been cancelled.\n\nReason:\n%s\n\nLog in to Jobira to view the booking details.",
			reason,
		),
	}
}
