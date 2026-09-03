package bookings

import (
	"context"
	"strings"
	"time"

	bookingtimelinedomain "github.com/rodrigueghenda/jobira/internal/domain/bookingtimeline"
	emaildomain "github.com/rodrigueghenda/jobira/internal/domain/email"
	favoritesdomain "github.com/rodrigueghenda/jobira/internal/domain/favorites"
	notificationsdomain "github.com/rodrigueghenda/jobira/internal/domain/notifications"
	preferredcleanersdomain "github.com/rodrigueghenda/jobira/internal/domain/preferredcleaners"
	reviewsdomain "github.com/rodrigueghenda/jobira/internal/domain/reviews"
)

type AvailabilityChecker interface {
	ValidateBookingSlot(
		ctx context.Context,
		cleanerID uint,
		startAt time.Time,
		endAt time.Time,
	) error

	BookingBufferMinutes(
		ctx context.Context,
		cleanerID uint,
	) (int, error)
}

type Service struct {
	repo                     Repository
	emailService             *emaildomain.Service
	reviewsService           *reviewsdomain.Service
	favoritesService         *favoritesdomain.Service
	preferredCleanersService *preferredcleanersdomain.Service
	bookingTimeLineService   *bookingtimelinedomain.Service
	notificationsService     *notificationsdomain.Service
	availabilityChecker      AvailabilityChecker
}

func NewService(repo Repository, emailService *emaildomain.Service, reviewsService *reviewsdomain.Service, favoritesService *favoritesdomain.Service, preferredCleanersService *preferredcleanersdomain.Service, bookingTimeLineService *bookingtimelinedomain.Service, notificationsService *notificationsdomain.Service) *Service {
	return &Service{
		repo:                     repo,
		emailService:             emailService,
		reviewsService:           reviewsService,
		favoritesService:         favoritesService,
		preferredCleanersService: preferredCleanersService,
		bookingTimeLineService:   bookingTimeLineService,
		notificationsService:     notificationsService,
	}
}

func (s *Service) SetAvailabilityChecker(
	checker AvailabilityChecker,
) {
	s.availabilityChecker = checker
}

func (s *Service) Create(ctx context.Context, clientID uint, req CreateBookingRequest) (*Booking, error) {
	if clientID == 0 ||
		req.JobID == 0 ||
		req.CleanerID == 0 {
		return nil, ErrInvalidInput
	}

	rawStart := strings.TrimSpace(
		req.ScheduledAt,
	)

	rawEnd := strings.TrimSpace(
		req.ScheduledEndAt,
	)

	var scheduledAt *time.Time
	var scheduledEndAt *time.Time

	if (rawStart == "") != (rawEnd == "") {
		return nil, ErrInvalidInput
	}

	if rawStart != "" {
		parsedStart, err := time.Parse(
			time.RFC3339,
			rawStart,
		)
		if err != nil {
			return nil, ErrInvalidInput
		}

		parsedEnd, err := time.Parse(
			time.RFC3339,
			rawEnd,
		)
		if err != nil {
			return nil, ErrInvalidInput
		}

		if !parsedEnd.After(parsedStart) {
			return nil, ErrInvalidInput
		}

		if parsedStart.Before(time.Now()) {
			return nil, ErrInvalidInput
		}

		conflictStart := parsedStart
		conflictEnd := parsedEnd

		if s.availabilityChecker != nil {
			bufferMinutes, err := s.availabilityChecker.BookingBufferMinutes(
				ctx,
				req.CleanerID,
			)
			if err != nil {
				return nil, err
			}

			if bufferMinutes > 0 {
				bufferDuration := time.Duration(
					bufferMinutes,
				) * time.Minute

				conflictStart = conflictStart.Add(
					-bufferDuration,
				)

				conflictEnd = conflictEnd.Add(
					bufferDuration,
				)
			}
		}

		conflict, err := s.repo.HasScheduleConflict(
			ctx,
			req.CleanerID,
			conflictStart,
			conflictEnd,
			0,
		)
		if err != nil {
			return nil, err
		}

		if conflict {
			return nil, ErrBookingConflict
		}

		if s.availabilityChecker != nil {
			if err := s.availabilityChecker.ValidateBookingSlot(
				ctx,
				req.CleanerID,
				parsedStart,
				parsedEnd,
			); err != nil {
				return nil, err
			}
		}

		if err != nil {
			return nil, err
		}

		if conflict {
			return nil, ErrBookingConflict
		}

		if s.availabilityChecker != nil {
			if err := s.availabilityChecker.ValidateBookingSlot(
				ctx,
				req.CleanerID,
				parsedStart,
				parsedEnd,
			); err != nil {
				return nil, err
			}
		}

		scheduledAt = &parsedStart
		scheduledEndAt = &parsedEnd
	}

	booking := &Booking{
		JobID:          req.JobID,
		ApplicationID:  req.ApplicationID,
		ClientID:       clientID,
		CleanerID:      req.CleanerID,
		Status:         "pending",
		ScheduledAt:    scheduledAt,
		ScheduledEndAt: scheduledEndAt,
	}

	if err := s.repo.Create(ctx, booking); err != nil {
		return nil, err
	}

	if s.bookingTimeLineService != nil {
		if err := s.bookingTimeLineService.Record(
			ctx,
			booking.ID,
			clientID,
			"",
			"pending",
			"Booking created",
		); err != nil {
			return nil, err
		}
	}

	s.sendBookingNotification(
		ctx,
		booking.CleanerID,
		"New booking request",
		"A client has sent you a new booking request",
	)

	if req.ApplicationID != nil && *req.ApplicationID != 0 {
		if err := s.repo.MarkApplicationAccepted(ctx, *req.ApplicationID); err != nil {
			return nil, err
		}
	}
	return booking, nil
}

func (s *Service) GetByID(ctx context.Context, bookingID uint, userID uint, role string) (*Booking, error) {
	if bookingID == 0 || userID == 0 {
		return nil, ErrInvalidInput
	}

	booking, err := s.repo.GetByID(ctx, bookingID)
	if err != nil {
		return nil, err
	}

	if role != "admin" && booking.ClientID != userID && booking.CleanerID != userID {
		return nil, ErrForbidden
	}

	return booking, nil
}

func (s *Service) ListMine(ctx context.Context, userID uint) ([]Booking, error) {
	if userID == 0 {
		return nil, ErrInvalidInput
	}

	return s.repo.ListByUserID(ctx, userID)
}

func (s *Service) Confirm(ctx context.Context, bookingID uint, userID uint, role string) (*Booking, error) {
	booking, err := s.GetByID(ctx, bookingID, userID, role)
	if err != nil {
		return nil, err
	}

	if booking.Status != "pending" {
		return nil, ErrInvalidInput
	}

	if role != "admin" && booking.CleanerID != userID {
		return nil, ErrForbidden
	}

	if err := s.repo.UpdateStatus(ctx, bookingID, "confirmed"); err != nil {
		return nil, err
	}

	if s.bookingTimeLineService != nil {
		if err := s.bookingTimeLineService.Record(
			ctx,
			bookingID,
			userID,
			booking.Status,
			"confirmed",
			"Cleaner confirmed the booking",
		); err != nil {
			return nil, err
		}
	}

	s.sendBookingNotification(
		ctx,
		booking.ClientID,
		"Booking confirmed",
		"The cleaner confirmed your booking.",
	)

	updatedBooking, err := s.repo.GetByID(ctx, bookingID)
	if err != nil {
		return nil, err
	}

	if s.emailService != nil {
		clientEmail, err := s.repo.GetUserEmail(ctx, updatedBooking.ClientID)
		if err == nil && clientEmail != "" {
			_ = s.emailService.Send(
				ctx,
				emaildomain.BookingConfirmedEmail(clientEmail),
			)
		}
	}

	return updatedBooking, nil
}

func (s *Service) Start(ctx context.Context, bookingID uint, userID uint, role string) (*Booking, error) {
	booking, err := s.GetByID(ctx, bookingID, userID, role)
	if err != nil {
		return nil, err
	}

	if booking.Status != "confirmed" {
		return nil, ErrInvalidInput
	}

	if role != "admin" && booking.CleanerID != userID {
		return nil, ErrForbidden
	}

	if err := s.repo.UpdateStatus(ctx, bookingID, "in_progress"); err != nil {
		return nil, err
	}

	if s.bookingTimeLineService != nil {
		if err := s.bookingTimeLineService.Record(
			ctx,
			bookingID,
			userID,
			booking.Status,
			"in_progress",
			"Cleaner started the booking",
		); err != nil {
			return nil, err
		}
	}

	s.sendBookingNotification(
		ctx,
		booking.ClientID,
		"Cleaner started",
		"The cleaner has started working on your booking",
	)

	return s.repo.GetByID(ctx, bookingID)
}

func (s *Service) Complete(ctx context.Context, bookingID uint, userID uint, role string) (*Booking, error) {
	booking, err := s.GetByID(ctx, bookingID, userID, role)
	if err != nil {
		return nil, err
	}

	if booking.Status != "confirmed" && booking.Status != "in_progress" {
		return nil, ErrInvalidInput
	}

	if role != "admin" && booking.CleanerID != userID {
		return nil, ErrForbidden
	}

	if err := s.repo.Complete(ctx, bookingID, "completed"); err != nil {
		return nil, err
	}

	if err := s.repo.MarkJobCompleted(ctx, booking.JobID); err != nil {
		return nil, err
	}

	if s.bookingTimeLineService != nil {
		if err := s.bookingTimeLineService.Record(
			ctx,
			bookingID,
			userID,
			booking.Status,
			"completed",
			"Work marked as completed",
		); err != nil {
			return nil, err
		}
	}

	s.sendBookingNotification(
		ctx,
		booking.ClientID,
		"Work marked complete",
		"The cleaner marked the work as completed. Please confirm and leave your review.",
	)

	updatedBooking, err := s.repo.GetByID(ctx, bookingID)
	if err != nil {
		return nil, err
	}

	if s.emailService != nil {
		cleanerEmail, err := s.repo.GetUserEmail(ctx, updatedBooking.CleanerID)
		if err == nil && cleanerEmail != "" {
			_ = s.emailService.Send(
				ctx,
				emaildomain.BookingCompletedEmail(cleanerEmail),
			)
		}
	}

	return updatedBooking, nil
}

func (s *Service) Cancel(ctx context.Context, bookingID uint, userID uint, role string, req CancelBookingRequest) (*Booking, error) {
	req.Reason = strings.TrimSpace(req.Reason)

	if req.Reason == "" {
		return nil, ErrInvalidInput
	}

	booking, err := s.GetByID(ctx, bookingID, userID, role)
	if err != nil {
		return nil, err
	}

	if booking.Status == "completed" || booking.Status == "closed" {
		return nil, ErrInvalidInput
	}

	if role != "admin" && booking.ClientID != userID && booking.CleanerID != userID {
		return nil, ErrForbidden
	}

	if err := s.repo.Cancel(ctx, bookingID, userID, req.Reason); err != nil {
		return nil, err
	}

	if s.bookingTimeLineService != nil {
		if err := s.bookingTimeLineService.Record(
			ctx,
			bookingID,
			userID,
			booking.Status,
			"cancelled",
			req.Reason,
		); err != nil {
			return nil, err
		}
	}

	cancelMessage := "The booking was cancelled. Reason: " + req.Reason

	if booking.ClientID != userID {
		s.sendBookingNotification(
			ctx,
			booking.ClientID,
			"Booking cancelled",
			cancelMessage,
		)
	}

	if booking.CleanerID != userID {
		s.sendBookingNotification(
			ctx,
			booking.CleanerID,
			"Booking cancelled",
			cancelMessage,
		)
	}

	updatedBooking, err := s.repo.GetByID(ctx, bookingID)
	if err != nil {
		return nil, err
	}

	if s.emailService != nil {
		clientEmail, err := s.repo.GetUserEmail(ctx, updatedBooking.ClientID)
		if err == nil && clientEmail != "" {
			_ = s.emailService.Send(ctx, emaildomain.BookingCancelledEmail(clientEmail, req.Reason))
		}
	}
	if s.emailService != nil {
		cleanerEmail, err := s.repo.GetUserEmail(ctx, updatedBooking.CleanerID)
		if err == nil && cleanerEmail != "" {
			_ = s.emailService.Send(ctx, emaildomain.BookingCancelledEmail(cleanerEmail, req.Reason))
		}
	}

	return updatedBooking, nil
}

func (s *Service) Close(ctx context.Context, bookingID uint, userID uint, role string, req CloseBookingRequest) (*Booking, error) {
	req.ClosureComment = strings.TrimSpace(req.ClosureComment)

	if bookingID == 0 || userID == 0 {
		return nil, ErrInvalidInput
	}

	if req.ClientRating < 1 || req.ClientRating > 5 {
		return nil, ErrInvalidInput
	}

	if req.ClosureComment == "" {
		return nil, ErrInvalidInput
	}

	booking, err := s.GetByID(ctx, bookingID, userID, role)
	if err != nil {
		return nil, err
	}

	if role != "admin" && booking.ClientID != userID {
		return nil, ErrForbidden
	}

	if booking.Status != "completed" {
		return nil, ErrInvalidInput
	}

	transcation := CloseBookingTransaction{
		BookingID:           bookingID,
		JobID:               booking.JobID,
		ClientID:            booking.ClientID,
		CleanerID:           booking.CleanerID,
		Rating:              req.ClientRating,
		WouldHireAgain:      req.ClientWouldHireAgain,
		Comment:             req.ClosureComment,
		AddToFavourites:     req.AddFavourites,
		SetAsPreferred:      req.SetAsPreferred,
		ChangedBy:           userID,
		PreviousStatus:      booking.Status,
		NotificationTitle:   "Booking closed",
		NotificationMessage: "The client confirmed completion and closed the booking.",
	}

	if err := s.repo.CloseWithTransaction(ctx, transcation); err != nil {
		return nil, err
	}

	return s.repo.GetByID(ctx, bookingID)
}

func (s *Service) sendBookingNotification(ctx context.Context, userID uint, title string, message string) {
	if s.notificationsService == nil || userID == 0 {
		return
	}

	_, _ = s.notificationsService.Create(ctx, notificationsdomain.CreateNotificationsRequest{
		UserID:  userID,
		Title:   title,
		Message: message,
		Type:    "booking",
	},
	)
}
