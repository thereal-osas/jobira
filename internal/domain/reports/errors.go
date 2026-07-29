package reports

import "errors"

var (
	ErrInvalidInput        = errors.New("invalid input")
	ErrReportNotFound      = errors.New("report not found")
	ErrInvalidReportType   = errors.New("invalid report type")
	ErrInvalidReportStatus = errors.New("invalid report status")
	ErrForbidden           = errors.New("forbidden")
)
