package applications

import (
	"context"
	"strings"

	emaildomain "github.com/rodrigueghenda/jobira/internal/domain/email"
	notificationsdomain "github.com/rodrigueghenda/jobira/internal/domain/notifications"
	"github.com/rodrigueghenda/jobira/internal/domain/profiles"
	subscriptionaccess "github.com/rodrigueghenda/jobira/internal/domain/subscriptionaccess"
	usagedomain "github.com/rodrigueghenda/jobira/internal/domain/subscriptions/usage"
)

type Service struct {
	repo                 Repository
	notificationsService *notificationsdomain.Service
	profilesService      *profiles.Service
	usageService         *usagedomain.Service
	emailService         *emaildomain.Service
	subscriptionaccess   *subscriptionaccess.Service
}

func NewService(repo Repository, notificationsService *notificationsdomain.Service, profilesService *profiles.Service, usageService *usagedomain.Service, emailService *emaildomain.Service, subscriptionAccess *subscriptionaccess.Service) *Service {
	return &Service{
		repo:                 repo,
		notificationsService: notificationsService,
		profilesService:      profilesService,
		usageService:         usageService,
		emailService:         emailService,
		subscriptionaccess:   subscriptionAccess,
	}
}
func (s *Service) Apply(ctx context.Context, jobID uint, cleanerID uint, req CreateApplicationRequest) (*Application, error) {
	req.CoverMessage = strings.TrimSpace(req.CoverMessage)

	if jobID == 0 || cleanerID == 0 {
		return nil, ErrInvalidInput
	}

	if req.CoverMessage == "" {
		return nil, ErrInvalidInput
	}

	if req.ProposedRate < 0 {
		return nil, ErrInvalidInput
	}

	clientID, err := s.repo.GetJobClientID(ctx, jobID)
	if err != nil {
		return nil, err
	}

	if clientID == cleanerID {
		return nil, ErrCannotApplyToOwnJob
	}

	if s.subscriptionaccess != nil {
		if err := s.subscriptionaccess.EnsureCanApply(ctx, cleanerID); err != nil {
			return nil, err
		}
	}

	if s.usageService != nil {
		if err := s.usageService.CanApply(ctx, cleanerID); err != nil {
			return nil, err
		}
	}

	application := &Application{
		JobID:        jobID,
		CleanerID:    cleanerID,
		CoverMessage: req.CoverMessage,
		ProposedRate: req.ProposedRate,
		Status:       "pending",
	}

	if err := s.repo.Create(ctx, application); err != nil {
		return nil, err
	}

	if s.subscriptionaccess != nil {
		if err := s.subscriptionaccess.IncrementApplicationsToday(ctx, cleanerID); err != nil {
			return nil, err
		}
	}

	if s.usageService != nil {
		if err := s.usageService.IncrementApplicationCount(ctx, cleanerID); err != nil {
			return nil, err
		}
	}

	jobTitle, _ := s.repo.GetJobTitle(ctx, jobID)

	if s.notificationsService != nil {
		_, _ = s.notificationsService.Create(ctx, notificationsdomain.CreateNotificationsRequest{
			UserID:  clientID,
			Title:   "New application",
			Message: "A cleaner applied to your job: " + jobTitle,
			Type:    "application",
		})
	}

	if s.emailService != nil {
		clientEmail, err := s.repo.GetUserEmail(ctx, clientID)
		if err == nil && clientEmail != "" {
			_ = s.emailService.Send(ctx, emaildomain.NewApplicationEmail(clientEmail, jobTitle))
		}
	}

	return application, nil
}

func (s *Service) GetByID(ctx context.Context, id uint, userID uint, role string) (*Application, error) {
	if id == 0 || userID == 0 {
		return nil, ErrInvalidInput
	}

	foundApplication, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	clientID, err := s.repo.GetJobClientID(ctx, foundApplication.JobID)
	if err != nil {
		return nil, err
	}

	if role != "admin" && foundApplication.CleanerID != userID && clientID != userID {
		return nil, ErrForbidden
	}

	return foundApplication, nil
}

func (s *Service) ListByJobID(ctx context.Context, jobID uint, userID uint, role string) ([]Application, error) {
	if jobID == 0 || userID == 0 {
		return nil, ErrInvalidInput
	}

	clientID, err := s.repo.GetJobClientID(ctx, jobID)
	if err != nil {
		return nil, err
	}

	if role != "admin" && clientID != userID {
		return nil, ErrForbidden
	}

	return s.repo.ListByJobID(ctx, jobID)
}

func (s *Service) ListMine(ctx context.Context, cleanerID uint) ([]Application, error) {
	if cleanerID == 0 {
		return nil, ErrInvalidInput
	}

	return s.repo.ListByCleanerID(ctx, cleanerID)
}
func (s *Service) UpdateStatus(
	ctx context.Context,
	applicationID uint,
	userID uint,
	role string,
	req UpdateApplicationStatusRequest,
) (*Application, error) {
	req.Status = strings.TrimSpace(req.Status)

	if applicationID == 0 || userID == 0 {
		return nil, ErrInvalidInput
	}

	if !isAllowedStatus(req.Status) {
		return nil, ErrInvalidApplicationStatus
	}

	foundApplication, err := s.repo.GetByID(ctx, applicationID)
	if err != nil {
		return nil, err
	}

	clientID, err := s.repo.GetJobClientID(ctx, foundApplication.JobID)
	if err != nil {
		return nil, err
	}

	if role != "admin" && clientID != userID {
		return nil, ErrForbidden
	}

	if s.profilesService != nil {
		if req.Status == "completed" &&
			foundApplication.Status != "completed" {
			if err := s.profilesService.IncrementJobsCompleted(
				ctx,
				foundApplication.CleanerID,
			); err != nil {
				return nil, err
			}
		}

		if req.Status == "cancelled" &&
			foundApplication.Status != "cancelled" &&
			foundApplication.Status != "completed" {
			if err := s.profilesService.IncrementJobsCancelled(
				ctx,
				foundApplication.CleanerID,
			); err != nil {
				return nil, err
			}
		}
	}

	if err := s.repo.UpdateStatus(
		ctx,
		applicationID,
		req.Status,
	); err != nil {
		return nil, err
	}

	if s.notificationsService != nil {
		_, _ = s.notificationsService.Create(
			ctx,
			notificationsdomain.CreateNotificationsRequest{
				UserID:  foundApplication.CleanerID,
				Title:   "Application update",
				Message: "Your application status was updated to: " + req.Status,
				Type:    "application_status",
			},
		)
	}

	updatedApplication, err := s.repo.GetByID(ctx, applicationID)
	if err != nil {
		return nil, err
	}

	return updatedApplication, nil
}

func isAllowedStatus(status string) bool {
	if status == "pending" {
		return true
	}

	if status == "shortlisted" {
		return true
	}

	if status == "invited" {
		return true
	}

	if status == "accepted" {
		return true
	}

	if status == "rejected" {
		return true
	}

	if status == "completed" {
		return true
	}

	if status == "cancelled" {
		return true
	}

	return false
}
