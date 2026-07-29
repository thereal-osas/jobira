package chat

import "errors"

var (
	ErrInvalidInput         = errors.New("invalid input")
	ErrConversationNotFound = errors.New("conversation not found")
	ErrMessageNotFound      = errors.New("message not found")
	ErrForbidden            = errors.New("forbidden")
	ErrBookingNotFound      = errors.New("booking not found")
	ErrConversationExists   = errors.New("conversation already exists")
	ErrInvalidMessage       = errors.New("message content is required")
	ErrMessageTooLong       = errors.New("message exceeds 4000 characters")
)
