package adminmoderation

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

func (s *Service) ListReports(ctx context.Context) ([]CleanerReportAdminView, error) {
	return s.repo.ListReports(ctx)
}
func (s *Service) ListOpenReports(ctx context.Context) ([]CleanerReportAdminView, error) {
	return s.repo.ListOpenReports(ctx)
}

func (s *Service) ListBlockCleaners(ctx context.Context) ([]BlockedCleanerAdminView, error) {
	return s.repo.ListBlockCleaners(ctx)
}

func (s *Service) UpdateReportStatus(ctx context.Context, reportID uint, adminID uint, req UpdateReportStatusRequest,) error {
	req.Status = strings.TrimSpace(req.Status)
	req.AdminNote = strings.TrimSpace(req.AdminNote)

	if reportID == 0 || adminID == 0 {
		return ErrInvalidInput
	}

	if !isAllowedReportStatus(req.Status) {
		return ErrInvalidInput
	}

	if !isAllowedReportStatus(req.Status) {
		return ErrInvalidInput
	}

	return s.repo.UpdateReportStatus(
		ctx, 
		reportID,
		adminID,
		req.Status,
		req.AdminNote,
	)
}

func isAllowedReportStatus(status string) bool {
	if status == "open" {
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