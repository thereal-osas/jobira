package applicationtimeline

import (
	"context"
	"strings"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) Record(ctx context.Context, applicationID uint, status string, actorUserID uint, note string,) error {
	status = strings.TrimSpace(status)
	note = strings.TrimSpace(note)

	if applicationID == 0 || actorUserID == 0 {
		return ErrInvalidInput
	}

	if !isAllowedStatus(status) {
		return ErrInvalidStatus
	}

	actorID := actorUserID

	event := &Event{
		ApplicationID: applicationID,
		Status: status,
		ActorUserID: &actorID,
		Note: note,
	}

	return s.repo.Create(ctx, event)
}
func (s *Service) GetTimeline(
	ctx context.Context,
	applicationID uint,
	userID uint,
	role string,
) (*TimeLine, error) {
	if applicationID == 0 || userID == 0 {
		return nil, ErrInvalidInput
	}

	cleanerID, err := s.repo.GetApplicationCleanerID(
		ctx,
		applicationID,
	)
	if err != nil {
		return nil, err
	}

	clientID, err := s.repo.GetApplicationClientID(
		ctx,
		applicationID,
	)
	if err != nil {
		return nil, err
	}

	isAdmin := role == "admin"
	isCleaner := userID == cleanerID
	isClient := userID == clientID

	if !isAdmin && !isCleaner && !isClient {
		return nil, ErrForbidden
	}

	status, err := s.repo.GetCurrentStatus(
		ctx,
		applicationID,
	)
	if err != nil {
		return nil, err
	}

	events, err := s.repo.ListByApplicationID(
		ctx,
		applicationID,
	)
	if err != nil {
		return nil, err
	}

	return &TimeLine{
		ApplicationID: applicationID,
		CurrentStatus: status,
		Events:        events,
	}, nil
}

func isAllowedStatus(status string) bool {
	switch status {
	case "pending", 
		"shortlisted",
		"invited",
		"accepted",
		"rejected",
		"completed",
		"cancelled":
		return true
	default:
		return false		
	}
}