package clientnotes

import "errors"

var (
	ErrInvalidInput = errors.New("invalid input")
	ErrNoteNotFound = errors.New("note not found")
)

