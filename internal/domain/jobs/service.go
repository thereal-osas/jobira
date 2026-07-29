package jobs

import (
	"context"
	"strings"

	subscriptionaccess "github.com/rodrigueghenda/jobira/internal/domain/subscriptionaccess"
	usagedomain "github.com/rodrigueghenda/jobira/internal/domain/subscriptions/usage"
)

type Service struct {
	repo               Repository
	usageService       *usagedomain.Service
	subscriptionAccess *subscriptionaccess.Service
}

func NewService(repo Repository, usageService *usagedomain.Service, subscriptionAccess *subscriptionaccess.Service) *Service {
	return &Service{
		repo:               repo,
		usageService:       usageService,
		subscriptionAccess: subscriptionAccess,
	}
}

func (s *Service) Create(ctx context.Context, clientID uint, req CreateJobRequest) (*Job, error) {
	req.Title = strings.TrimSpace(req.Title)
	req.Description = strings.TrimSpace(req.Description)
	req.Location = strings.TrimSpace(req.Location)
	req.JobType = strings.TrimSpace(req.JobType)
	req.ListingType = strings.TrimSpace(req.ListingType)

	if clientID == 0 {
		return nil, ErrInvalidInput
	}

	if req.Title == "" || req.Description == "" || req.Location == "" || req.JobType == "" {
		return nil, ErrInvalidInput
	}

	if req.Budget < 0 {
		return nil, ErrInvalidInput
	}

	if req.ListingType == "" {
		req.ListingType = "shift"
	}

	if !isAllowedListingType(req.ListingType) {
		return nil, ErrInvalidInput
	}

	if s.usageService != nil {
		if err := s.usageService.CanPostJob(ctx, clientID); err != nil {
			return nil, err
		}
	}

	job := &Job{
		ClientID:    clientID,
		Title:       req.Title,
		Description: req.Description,
		Location:    req.Location,
		JobType:     req.JobType,
		ListingType: req.ListingType,
		Budget:      req.Budget,
		Status:      "open",
	}

	if s.subscriptionAccess != nil {
		if err := s.subscriptionAccess.EnsureCanPostJob(ctx, clientID); err != nil {
			return nil, err
		}
	}

	if err := s.repo.Create(ctx, job); err != nil {
		return nil, err
	}

	if s.subscriptionAccess != nil {
		if err := s.subscriptionAccess.IncrementJobsPostToday(ctx, clientID); err != nil {
			return nil, err
		}
	}

	if s.usageService != nil {
		if err := s.usageService.IncrementJobPostCount(ctx, clientID); err != nil {
			return nil, err
		}
	}

	return job, nil
}

func isAllowedListingType(value string) bool {
	if value == "job" {
		return true
	}

	if value == "shift" {
		return true
	}

	return false
}
func (s *Service) GetByID(ctx context.Context, id uint) (*Job, error) {
	if id == 0 {
		return nil, ErrInvalidInput
	}

	return s.repo.GetByID(ctx, id)
}

func (s *Service) List(ctx context.Context) ([]Job, error) {
	return s.repo.List(ctx)
}

func (s *Service) ListByClientID(ctx context.Context, clientID uint) ([]Job, error) {
	if clientID == 0 {
		return nil, ErrInvalidInput
	}

	return s.repo.ListByClientID(ctx, clientID)
}

func (s *Service) Update(ctx context.Context, id uint, clientID uint, role string, req UpdateJobRequest) (*Job, error) {
	if id == 0 || clientID == 0 {
		return nil, ErrInvalidInput
	}

	existingJob, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if role != "admin" && existingJob.ClientID != clientID {
		return nil, ErrForbidden
	}

	req.Title = strings.TrimSpace(req.Title)
	req.Description = strings.TrimSpace(req.Description)
	req.Location = strings.TrimSpace(req.Location)
	req.JobType = strings.TrimSpace(req.JobType)
	req.ListingType = strings.TrimSpace(req.ListingType)
	req.Status = strings.TrimSpace(req.Status)

	if req.Title == "" || req.Description == "" || req.Location == "" || req.JobType == "" || req.Status == "" {
		return nil, ErrInvalidInput
	}

	if req.Budget < 0 {
		return nil, ErrInvalidInput
	}

	existingJob.Title = req.Title
	existingJob.Description = req.Description
	existingJob.Location = req.Location
	existingJob.JobType = req.JobType
	existingJob.ListingType = req.ListingType
	existingJob.Budget = req.Budget
	existingJob.Status = req.Status

	if err := s.repo.Update(ctx, existingJob); err != nil {
		return nil, err
	}

	return existingJob, nil
}

func (s *Service) Delete(ctx context.Context, id uint, clientID uint, role string) error {
	if id == 0 || clientID == 0 {
		return ErrInvalidInput
	}

	existingJob, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if role != "admin" && existingJob.ClientID != clientID {
		return ErrForbidden
	}

	return s.repo.Delete(ctx, id)
}

func (s *Service) Search(ctx context.Context, req SearchJobRequest) ([]Job, error) {
	req.Location = strings.TrimSpace(req.Location)
	req.JobType = strings.TrimSpace(req.JobType)
	req.ListingType = strings.TrimSpace(req.ListingType)

	return s.repo.Search(ctx, req)
}
