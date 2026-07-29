package adminmoderation

import "errors"

var (
	ErrInvalidInput = errors.New("invalid input")
	ErrInvalidStatus = errors.New("invalid report status")
	ErrReportNotFound = errors.New("report not found")
)