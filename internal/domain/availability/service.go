package availability

import (
	"context"
	"strings"
	"time"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, cleanerID uint, req CleanerAvailabilityRequest) (*CleanerAvailability, error) {
	req.AvailableDate = strings.TrimSpace(req.AvailableDate)
	req.StartTime = strings.TrimSpace(req.StartTime)
	req.EndTime = strings.TrimSpace(req.EndTime)
	req.Status = strings.TrimSpace(req.Status)
	req.Notes = strings.TrimSpace(req.Notes)

	if cleanerID == 0 || req.AvailableDate == "" || req.StartTime == "" || req.EndTime == "" {
		return nil, ErrInvalidInput
	}

	if _, err := time.Parse("2006-01-02", req.AvailableDate); err != nil {
		return nil, ErrInvalidInput
	}

	if req.Status == "" {
		req.Status = "available"
	}

	if !isAllowedStatus(req.Status) {
		return nil, ErrInvalidInput
	}

	conflict, err := s.repo.HasConflict(
		ctx,
		cleanerID,
		req.AvailableDate,
		req.StartTime,
		req.EndTime,
		0,
	)
	if err != nil {
		return nil, err
	}

	if conflict {
		return nil, ErrAvailabilityConflict
	}

	availability := &CleanerAvailability{
		CleanerID:     cleanerID,
		AvailableDate: req.AvailableDate,
		StartTime:     req.StartTime,
		EndTime:       req.EndTime,
		Status:        req.Status,
		Notes:         req.Notes,
	}

	if err := s.repo.Create(ctx, availability); err != nil {
		return nil, err
	}

	return availability, nil
}

func (s *Service) GetByID(ctx context.Context, id uint, userID uint, role string) (*CleanerAvailability, error) {
	if id == 0 || userID == 0 {
		return nil, ErrInvalidInput
	}

	availability, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if role != "admin" && availability.CleanerID != userID {
		return nil, ErrForbidden
	}

	return availability, nil
}

func (s *Service) ListByCleanerID(ctx context.Context, cleanerID uint) ([]CleanerAvailability, error) {
	if cleanerID == 0 {
		return nil, ErrInvalidInput
	}

	return s.repo.ListByCleanerID(ctx, cleanerID)
}

func (s *Service) Update(ctx context.Context, id uint, userID uint, role string, req UpdateAvailabilityRequest) (*CleanerAvailability, error) {
	if id == 0 || userID == 0 {
		return nil, ErrInvalidInput
	}

	foundAvailability, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if role != "admin" && foundAvailability.CleanerID != userID {
		return nil, ErrForbidden
	}

	req.AvailableDate = strings.TrimSpace(req.AvailableDate)
	req.StartTime = strings.TrimSpace(req.StartTime)
	req.EndTime = strings.TrimSpace(req.EndTime)
	req.Status = strings.TrimSpace(req.Status)
	req.Notes = strings.TrimSpace(req.Notes)

	if req.AvailableDate == "" || req.StartTime == "" || req.EndTime == "" {
		return nil, ErrInvalidInput
	}

	if _, err := time.Parse("2006-01-02", req.AvailableDate); err != nil {
		return nil, ErrInvalidInput
	}

	if req.Status == "" {
		req.Status = "available"
	}

	if !isAllowedStatus(req.Status) {
		return nil, ErrInvalidStatus
	}

	conflict, err := s.repo.HasConflict(
		ctx,
		foundAvailability.CleanerID,
		req.AvailableDate,
		req.StartTime,
		req.EndTime,
		foundAvailability.ID,
	)
	if err != nil {
		return nil, err
	}

	if conflict {
		return nil, ErrAvailabilityConflict
	}

	foundAvailability.AvailableDate = req.AvailableDate
	foundAvailability.StartTime = req.StartTime
	foundAvailability.EndTime = req.EndTime
	foundAvailability.Status = req.Status
	foundAvailability.Notes = req.Notes

	if err := s.repo.Update(ctx, foundAvailability); err != nil {
		return nil, err
	}

	return s.repo.GetByID(ctx, id)
}

func (s *Service) Delete(ctx context.Context, id uint, userID uint, role string) error {
	if id == 0 || userID == 0 {
		return ErrInvalidInput
	}

	foundAvailability, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if role != "admin" && foundAvailability.CleanerID != userID {
		return ErrForbidden
	}

	return s.repo.Delete(ctx, id)
}

func isAllowedStatus(status string) bool {
	if status == "available" {
		return true
	}

	if status == "busy" {
		return true
	}

	if status == "unavailable" {
		return true
	}

	return false
}

func (s *Service) ListMine(ctx context.Context, cleanerID uint) ([]CleanerAvailability, error) {
	if cleanerID == 0 {
		return nil, ErrInvalidInput
	}

	return s.repo.ListByCleanerID(ctx, cleanerID)
}

func (s *Service) CreateBlock(ctx context.Context, cleanerID uint, req CreateAvailabilityBlockRequest) (*AvailabilityBlock, error) {
	req.StartAt = strings.TrimSpace(req.StartAt)
	req.EndAt = strings.TrimSpace(req.EndAt)
	req.Reason = strings.TrimSpace(req.Reason)

	if cleanerID == 0 || req.Reason == "" {
		return nil, ErrInvalidInput
	}

	if req.StartAt == "" || req.EndAt == "" {
		return nil, ErrInvalidInput
	}

	startAt, err := time.Parse(time.RFC3339, req.StartAt)
	if err != nil {
		return nil, ErrInvalidInput
	}

	if startAt.Before(time.Now()) {
		return nil, ErrInvalidInput
	}

	endAt, err := time.Parse(time.RFC3339, req.EndAt)
	if err != nil {
		return nil, ErrInvalidInput
	}

	if !endAt.After(startAt) {
		return nil, ErrInvalidInput
	}

	block := &AvailabilityBlock{
		CleanerID: cleanerID,
		StartAt:   startAt,
		EndAt:     endAt,
		Reason:    req.Reason,
	}

	if err := s.repo.CreateBlock(ctx, block); err != nil {
		return nil, err
	}

	return block, nil
}

func (s *Service) ListBlocks(ctx context.Context, cleanerID uint) ([]AvailabilityBlock, error) {
	if cleanerID == 0 {
		return nil, ErrInvalidInput
	}

	return s.repo.ListBlocksByCleanerID(ctx, cleanerID)
}

func (s *Service) DeleteBlock(ctx context.Context, blockID uint, cleanerID uint) error {
	if blockID == 0 || cleanerID == 0 {
		return ErrInvalidInput
	}

	return s.repo.DeleteBlock(ctx, blockID, cleanerID)
}
