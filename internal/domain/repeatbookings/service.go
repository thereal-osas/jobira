package repeatbookings

import (
	"context"
	"strings"
	"time"

	availabilitydomain "github.com/rodrigueghenda/jobira/internal/domain/availability"
	notificationsdomain "github.com/rodrigueghenda/jobira/internal/domain/notifications"
)

type BlockChecker interface {
	IsBlocked(
		ctx context.Context,
		clientID uint,
		cleanerID uint,
	) (bool, error)
}

type AvailabilityChecker interface {
	ListByCleanerID(
		ctx context.Context,
		cleanerID uint,
	) ([]availabilitydomain.CleanerAvailability, error)
}

type Service struct {
	repo                 Repository
	notificationsService *notificationsdomain.Service
	blockChecker         BlockChecker
	availabilityService  AvailabilityChecker
}

func NewService(
	repo Repository,
	notificationsService *notificationsdomain.Service,
	blockChecker BlockChecker,
	availabilityService AvailabilityChecker,
) *Service {
	return &Service{
		repo:                 repo,
		notificationsService: notificationsService,
		blockChecker:         blockChecker,
		availabilityService:  availabilityService,
	}
}

func (s *Service) Create(ctx context.Context, clientID uint, cleanerID uint, req CreateRepeatBookingRequest) (*RepeatBooking, error) {
	if s.blockChecker != nil {
		blocked, err := s.blockChecker.IsBlocked(ctx, clientID, cleanerID)
		if err != nil {
			return nil, err
		}

		if blocked {
			return nil, ErrForbidden
		}
	}

	req.Title = strings.TrimSpace(req.Title)
	req.Description = strings.TrimSpace(req.Description)
	req.Location = strings.TrimSpace(req.Location)
	req.JobType = strings.TrimSpace(req.JobType)
	req.ListingType = strings.TrimSpace(req.ListingType)
	req.Message = strings.TrimSpace(req.Message)

	if clientID == 0 || cleanerID == 0 {
		return nil, ErrInvalidInput
	}

	if clientID == cleanerID {
		return nil, ErrInvalidInput
	}

	if req.Title == "" || req.Description == "" || req.Location == "" || req.JobType == "" {
		return nil, ErrInvalidInput
	}

	if req.ListingType == "" {
		req.ListingType = "shift"
	}

	if req.Budget < 0 {
		return nil, ErrInvalidInput
	}

	jobID, err := s.repo.CreateJob(ctx, clientID, req)
	if err != nil {
		return nil, err
	}

	if err := s.repo.CreateInvitation(ctx, jobID, clientID, cleanerID, req.Message); err != nil {
		return nil, err
	}

	booking := &RepeatBooking{
		ClientID:  clientID,
		CleanerID: cleanerID,
		JobID:     jobID,
		Message:   req.Message,
		Status:    "sent",
	}

	if err := s.repo.CreateRepeatBooking(ctx, booking); err != nil {
		return nil, err
	}

	if s.notificationsService != nil {
		_, _ = s.notificationsService.Create(ctx, notificationsdomain.CreateNotificationsRequest{
			UserID:  cleanerID,
			Title:   "Repeat booking request",
			Message: "A client wants to book you again",
			Type:    "repeat_bookings",
		})
	}

	return booking, nil
}

func (s *Service) ListMine(ctx context.Context, clientID uint) ([]RepeatBooking, error) {
	if clientID == 0 {
		return nil, ErrInvalidInput
	}

	return s.repo.ListByClientID(ctx, clientID)
}
func (s *Service) BookingAgain(
	ctx context.Context,
	originalBookingID uint,
	clientID uint,
	req BookingAgainRequest,
) (*RepeatBookingRequest, error) {
	req.ScheduledAt = strings.TrimSpace(
		req.ScheduledAt,
	)

	req.ScheduledEndAt = strings.TrimSpace(
		req.ScheduledEndAt,
	)

	req.Message = strings.TrimSpace(
		req.Message,
	)

	if originalBookingID == 0 ||
		clientID == 0 {
		return nil, ErrInvalidInput
	}

	if req.ScheduledAt == "" ||
		req.ScheduledEndAt == "" {
		return nil, ErrInvalidInput
	}

	scheduledAt, err := time.Parse(
		time.RFC3339,
		req.ScheduledAt,
	)
	if err != nil {
		return nil, ErrInvalidInput
	}

	scheduledEndAt, err := time.Parse(
		time.RFC3339,
		req.ScheduledEndAt,
	)
	if err != nil {
		return nil, ErrInvalidInput
	}

	if !scheduledEndAt.After(
		scheduledAt,
	) {
		return nil, ErrInvalidInput
	}

	if !scheduledAt.After(
		time.Now(),
	) {
		return nil, ErrInvalidInput
	}

	originalBooking, err :=
		s.repo.GetOriginBooking(
			ctx,
			originalBookingID,
		)
	if err != nil {
		return nil, err
	}

	if originalBooking.ClientID !=
		clientID {
		return nil, ErrForbidden
	}

	if originalBooking.Status != "closed" &&
		originalBooking.Status != "completed" {
		return nil, ErrInvalidInput
	}

	if s.blockChecker != nil {
		blocked, err :=
			s.blockChecker.IsBlocked(
				ctx,
				clientID,
				originalBooking.CleanerID,
			)
		if err != nil {
			return nil, err
		}

		if blocked {
			return nil, ErrForbidden
		}
	}

	if s.availabilityService != nil {
		availabilityRecords, err :=
			s.availabilityService.ListByCleanerID(
				ctx,
				originalBooking.CleanerID,
			)
		if err != nil {
			return nil, err
		}

		isAvailable := false

		requestedDate :=
			scheduledAt.Format(
				"2006-01-02",
			)

		requestedStart :=
			scheduledAt.Format(
				"15:04",
			)

		requestedEnd :=
			scheduledEndAt.Format(
				"15:04",
			)

		for _, slot := range availabilityRecords {
			if slot.AvailableDate !=
				requestedDate {
				continue
			}

			if slot.Status != "available" {
				continue
			}

			if requestedStart >=
				slot.StartTime &&
				requestedEnd <=
					slot.EndTime {
				isAvailable = true
				break
			}
		}

		if !isAvailable {
			return nil, ErrCleanerUnavailable
		}
	}

	newBookingID, err :=
		s.repo.CreateRepeatBookings(
			ctx,
			originalBooking,
			scheduledAt,
			scheduledEndAt,
		)
	if err != nil {
		return nil, err
	}

	request := &RepeatBookingRequest{
		OriginalBookingID: originalBooking.ID,

		NewBookingID: &newBookingID,

		ClientID: originalBooking.ClientID,

		CleanerID: originalBooking.CleanerID,

		JobID: originalBooking.JobID,

		ScheduledAt: scheduledAt,

		ScheduledEndAt: &scheduledEndAt,

		Status: "created",

		Message: req.Message,
	}

	if err :=
		s.repo.CreateBookAgainRequest(
			ctx,
			request,
		); err != nil {
		return nil, err
	}

	if s.notificationsService != nil {
		_, _ =
			s.notificationsService.Create(
				ctx,
				notificationsdomain.CreateNotificationsRequest{
					UserID: originalBooking.CleanerID,

					Title: "New repeat booking",

					Message: "A previous client has booked you again for " +
						scheduledAt.Format(
							"2 January 2006 at 15:04",
						),

					Type: "repeat_bookings",
				},
			)
	}

	return request, nil
}
