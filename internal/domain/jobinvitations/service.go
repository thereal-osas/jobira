package jobinvitations

import (
	"context"
	"strings"

	notificationsdomain "github.com/rodrigueghenda/jobira/internal/domain/notifications"
	blockedcleanersdomain "github.com/rodrigueghenda/jobira/internal/domain/blockedcleaners"
)

type Service struct {
	repo 		Repository
	notificationsService *notificationsdomain.Service
	blockChecker 		 *blockedcleanersdomain.Checker	
}
func NewService(
	repo Repository, notificationsService *notificationsdomain.Service,
	blockChecker *blockedcleanersdomain.Checker,
) *Service {
	return &Service{
		repo: repo,
		notificationsService: notificationsService,
		blockChecker: blockChecker,
	}
}

func (s *Service) Create(ctx context.Context, jobID uint, clientID uint, req CreateJobInvitationRequest) (*JobInvitation, error) {
	req.Message = strings.TrimSpace(req.Message)

	if jobID == 0 || clientID == 0 || req.CleanerID == 0 {
		return nil, ErrInvalidInput
	}

	if clientID == req.CleanerID {
		return nil, ErrInvalidInput
	}

	jobClientID, err := s.repo.GetJobClientID(ctx, jobID)
	if err != nil {
		return nil, err
	}

	if jobClientID != clientID {
		return nil, ErrForbidden
	}

	exists, err := s.repo.Exists(ctx, jobID, req.CleanerID)
	if err != nil {
		return nil, err
	}

	if exists {
		return nil, ErrInvitationExists
	}

	if s.blockChecker != nil {
		blocked, err := s.blockChecker.IsBlocked(ctx, clientID, req.CleanerID)
		if err != nil {
			return nil, err
		}

		if blocked {
			return nil, ErrForbidden
		}
	}

	invitation := &JobInvitation{
		JobID: jobID,
		ClientID: clientID,
		CleanerID: req.CleanerID,
		Status: "sent",
		Message: req.Message,
	}

	if err := s.repo.Create(ctx, invitation); err != nil {
		return nil, err 
	}

	if s.notificationsService != nil {
		_, _ = s.notificationsService.Create(ctx, notificationsdomain.CreateNotificationsRequest{
			UserID: req.CleanerID,
			Title: "New job invitation",
			Message: "A client invited you to apply for a job",
			Type: "job_invitations",
		})
	}

	return invitation, nil
}

func (s *Service) ListSent(ctx context.Context, clientID uint) ([]JobInvitation, error) {
	if clientID == 0 {
		return nil, ErrInvalidInput
	}

	return s.repo.ListSentByClientID(ctx, clientID)
}

func (s *Service) ListReceived(ctx context.Context, cleanerID uint) ([]JobInvitation, error) {
	if cleanerID == 0 {
		return nil, ErrInvalidInput
	}

	return s.repo.ListReceivedByCleanerID(ctx, cleanerID)
}