package reports

import (
	"context"
	"strings"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, reporterID uint, req CreateReportRequest) (*Report, error) {
	req.ReportType = strings.TrimSpace(req.ReportType)
	req.Reason = strings.TrimSpace(req.Reason)
	req.Details = strings.TrimSpace(req.Details)

	if reporterID == 0 || req.ReportType == "" || req.Reason == "" {
		return nil, ErrInvalidInput
	}

	if !isAllowedReportType(req.ReportType) {
		return nil, ErrInvalidReportType
	}

	report := &Report{
		ReporterID:     reporterID,
		ReportedUserID: req.ReportedUserID,
		JobID:          req.JobID,
		BookingID:      req.BookingID,
		ReportType:     req.ReportType,
		Reason:         req.Reason,
		Details:        req.Details,
		Status:         "open",
	}

	if err := s.repo.Create(ctx, report); err != nil {
		return nil, err
	}

	return report, nil
}

func (s *Service) GetByID(ctx context.Context, reportID uint, userID uint, role string) (*Report, error) {
	if reportID == 0 || userID == 0 {
		return nil, ErrInvalidInput
	}

	report, err := s.repo.GetByID(ctx, reportID)
	if err != nil {
		return nil, err
	}

	if role != "admin" && report.ReporterID != userID {
		return nil, ErrForbidden
	}

	return report, nil
}

func (s *Service) ListMine(ctx context.Context, reporterID uint) ([]Report, error) {
	if reporterID == 0 {
		return nil, ErrInvalidInput
	}

	return s.repo.ListByReporterID(ctx, reporterID)
}
func (s *Service) ListAll(ctx context.Context) ([]Report, error) {
	return s.repo.ListAll(ctx)
}

func (s *Service) ListOpen(ctx context.Context) ([]Report, error) {
	return s.repo.ListOpen(ctx)
}
func (s *Service) Review(ctx context.Context, reportID uint, adminID uint, req ReviewReportRequest) (*Report, error) {
	req.Status = strings.TrimSpace(req.Status)
	req.AdminNotes = strings.TrimSpace(req.AdminNotes)

	if reportID == 0 || adminID == 0 {
		return nil, ErrInvalidInput
	}

	if !isAllowedReportStatus(req.Status) {
		return nil, ErrInvalidInput
	}

	if err := s.repo.Review(ctx, reportID, req.Status, req.AdminNotes, adminID); err != nil {
		return nil, err
	}

	return s.repo.GetByID(ctx, reportID)
}

func isAllowedReportType(reportType string) bool {
	if reportType == "user" {
		return true
	}

	if reportType == "job" {
		return true
	}

	if reportType == "booking" {
		return true
	}

	if reportType == "message" {
		return true
	}

	if reportType == "abuse" {
		return true
	}

	if reportType == "no_show" {
		return true
	}

	if reportType == "spam" {
		return true
	}

	return false
}

func isAllowedReportStatus(status string) bool {
	if status == "open" {
		return true
	}

	if status == "reviewing" {
		return true
	}

	if status == "resolved" {
		return true
	}

	if status == "dismissed" {
		return true
	}

	return false
}
