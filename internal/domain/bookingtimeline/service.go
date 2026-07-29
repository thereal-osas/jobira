package bookingtimeline

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

func (s *Service) Record(ctx context.Context, bookingID uint, changedBy uint, fromStatus string, toStatus string, note string,) error {
	fromStatus = strings.TrimSpace(fromStatus)
	toStatus = strings.TrimSpace(toStatus)
	note = strings.TrimSpace(note)

	if bookingID == 0 || toStatus == "" {
		return ErrInvalidInput
	}

	return s.repo.CreateHistory(
		ctx,
		bookingID,
		changedBy,
		fromStatus,
		toStatus,
		note,
	)
}

func (s *Service) GetTimeline(ctx context.Context, bookingID uint, userID uint, role string,) (*BookingTimeline, error) {
	if bookingID == 0 || userID == 0 {
		return nil, ErrInvalidInput
	}

	clientID, cleanerID, currentStatus, err := s.repo.GetBookingAccess(ctx, bookingID)
	if err != nil {
		return nil, err
	}

	if role != "admin" && clientID != userID && cleanerID != userID {
		return nil, ErrForbidden
	}

	history, err := s.repo.ListByBookingID(ctx, bookingID)
	if err != nil {
		return nil, err
	}

	return &BookingTimeline{
		BookingID: bookingID,
		CurrentStatus: currentStatus,
		History: history,
	},nil
}