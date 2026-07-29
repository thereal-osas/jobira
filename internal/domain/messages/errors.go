package messages

import "errors"

var (
	ErrInvalidInput = errors.New("invalid input")
	ErrForbidden    = errors.New("forbidden")
	ErrConversation = errors.New("conversation not found")
)
