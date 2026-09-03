package repeatbookings

import (
	"context"
	"errors"
	"testing"
	"time"

	availabilitydomain "github.com/rodrigueghenda/jobira/internal/domain/availability"
)

type mockBlockChecker struct {
	isBlockedFn func(
		context.Context,
		uint,
		uint,
	) (bool, error)
}

func (m *mockBlockChecker) IsBlocked(
	ctx context.Context,
	clientID uint,
	cleanerID uint,
) (bool, error) {
	if m.isBlockedFn != nil {
		return m.isBlockedFn(
			ctx,
			clientID,
			cleanerID,
		)
	}

	return false, nil
}

type mockAvailabilityChecker struct {
	listByCleanerIDFn func(
		context.Context,
		uint,
	) ([]availabilitydomain.CleanerAvailability, error)
}

func (m *mockAvailabilityChecker) ListByCleanerID(
	ctx context.Context,
	cleanerID uint,
) ([]availabilitydomain.CleanerAvailability, error) {
	if m.listByCleanerIDFn != nil {
		return m.listByCleanerIDFn(ctx, cleanerID)
	}

	return nil, nil
}

func TestService_Create_BlockCheckerError_Edge(t *testing.T) {
	expectedErr := errors.New("block check failed")

	blockChecker := &mockBlockChecker{
		isBlockedFn: func(
			ctx context.Context,
			clientID uint,
			cleanerID uint,
		) (bool, error) {
			if clientID != 5 {
				t.Fatalf(
					"expected client ID 5, got %d",
					clientID,
				)
			}

			if cleanerID != 8 {
				t.Fatalf(
					"expected cleaner ID 8, got %d",
					cleanerID,
				)
			}

			return false, expectedErr
		},
	}

	service := NewService(
		&mockRepository{},
		nil,
		blockChecker,
		nil,
	)

	booking, err := service.Create(
		context.Background(),
		5,
		8,
		CreateRepeatBookingRequest{
			Title:       "Weekly clean",
			Description: "Clean a flat",
			Location:    "London",
			JobType:     "domestic",
		},
	)

	if booking != nil {
		t.Fatalf(
			"expected nil booking, got %+v",
			booking,
		)
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected %v, got %v",
			expectedErr,
			err,
		)
	}
}

func TestService_Create_Blocked_Edge(t *testing.T) {
	blockChecker := &mockBlockChecker{
		isBlockedFn: func(
			context.Context,
			uint,
			uint,
		) (bool, error) {
			return true, nil
		},
	}

	service := NewService(
		&mockRepository{},
		nil,
		blockChecker,
		nil,
	)

	booking, err := service.Create(
		context.Background(),
		5,
		8,
		CreateRepeatBookingRequest{
			Title:       "Weekly clean",
			Description: "Clean a flat",
			Location:    "London",
			JobType:     "domestic",
		},
	)

	if booking != nil {
		t.Fatalf(
			"expected nil booking, got %+v",
			booking,
		)
	}

	if !errors.Is(err, ErrForbidden) {
		t.Fatalf(
			"expected ErrForbidden, got %v",
			err,
		)
	}
}

func TestService_Create_BlockCheckerAllows_Edge(t *testing.T) {
	blockChecked := false

	blockChecker := &mockBlockChecker{
		isBlockedFn: func(
			context.Context,
			uint,
			uint,
		) (bool, error) {
			blockChecked = true
			return false, nil
		},
	}

	repo := &mockRepository{
		createJobFn: func(
			context.Context,
			uint,
			CreateRepeatBookingRequest,
		) (uint, error) {
			return 12, nil
		},
		createRepeatBookingFn: func(
			context.Context,
			*RepeatBooking,
		) error {
			return nil
		},
	}

	service := NewService(
		repo,
		nil,
		blockChecker,
		nil,
	)

	booking, err := service.Create(
		context.Background(),
		5,
		8,
		CreateRepeatBookingRequest{
			Title:       "Weekly clean",
			Description: "Clean a flat",
			Location:    "London",
			JobType:     "domestic",
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if booking == nil {
		t.Fatal("expected repeat booking")
	}

	if !blockChecked {
		t.Fatal("expected block checker to be called")
	}
}

func TestService_BookingAgain_BlockCheckerError_Edge(
	t *testing.T,
) {
	expectedErr := errors.New("block check failed")

	repo := &mockRepository{
		getOriginBookingFn: func(
			context.Context,
			uint,
		) (*BookingSnapshot, error) {
			return &BookingSnapshot{
				ID:        10,
				JobID:     7,
				ClientID:  5,
				CleanerID: 8,
				Status:    "closed",
			}, nil
		},
	}

	blockChecker := &mockBlockChecker{
		isBlockedFn: func(
			context.Context,
			uint,
			uint,
		) (bool, error) {
			return false, expectedErr
		},
	}

	service := NewService(
		repo,
		nil,
		blockChecker,
		nil,
	)

	request, err := service.BookingAgain(
		context.Background(),
		10,
		5,
		BookingAgainRequest{
			ScheduledAt:    "2026-09-10T10:00:00Z",
			ScheduledEndAt: "2026-09-10T12:00:00Z",
		},
	)

	if request != nil {
		t.Fatalf(
			"expected nil request, got %+v",
			request,
		)
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected %v, got %v",
			expectedErr,
			err,
		)
	}
}

func TestService_BookingAgain_Blocked_Edge(t *testing.T) {
	repo := &mockRepository{
		getOriginBookingFn: func(
			context.Context,
			uint,
		) (*BookingSnapshot, error) {
			return &BookingSnapshot{
				ID:        10,
				JobID:     7,
				ClientID:  5,
				CleanerID: 8,
				Status:    "completed",
			}, nil
		},
	}

	blockChecker := &mockBlockChecker{
		isBlockedFn: func(
			context.Context,
			uint,
			uint,
		) (bool, error) {
			return true, nil
		},
	}

	service := NewService(
		repo,
		nil,
		blockChecker,
		nil,
	)

	request, err := service.BookingAgain(
		context.Background(),
		10,
		5,
		BookingAgainRequest{
			ScheduledAt:    "2026-09-10T10:00:00Z",
			ScheduledEndAt: "2026-09-10T12:00:00Z",
		},
	)

	if request != nil {
		t.Fatalf(
			"expected nil request, got %+v",
			request,
		)
	}

	if !errors.Is(err, ErrForbidden) {
		t.Fatalf(
			"expected ErrForbidden, got %v",
			err,
		)
	}
}

func TestService_BookingAgain_AvailabilityError_Edge(
	t *testing.T,
) {
	expectedErr := errors.New("availability lookup failed")

	repo := &mockRepository{
		getOriginBookingFn: func(
			context.Context,
			uint,
		) (*BookingSnapshot, error) {
			return &BookingSnapshot{
				ID:        10,
				JobID:     7,
				ClientID:  5,
				CleanerID: 8,
				Status:    "closed",
			}, nil
		},
	}

	availabilityChecker := &mockAvailabilityChecker{
		listByCleanerIDFn: func(
			context.Context,
			uint,
		) ([]availabilitydomain.CleanerAvailability, error) {
			return nil, expectedErr
		},
	}

	service := NewService(
		repo,
		nil,
		nil,
		availabilityChecker,
	)

	request, err := service.BookingAgain(
		context.Background(),
		10,
		5,
		BookingAgainRequest{
			ScheduledAt:    "2026-09-10T10:00:00Z",
			ScheduledEndAt: "2026-09-10T12:00:00Z",
		},
	)

	if request != nil {
		t.Fatalf(
			"expected nil request, got %+v",
			request,
		)
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected %v, got %v",
			expectedErr,
			err,
		)
	}
}

func TestService_BookingAgain_NoAvailability_Edge(
	t *testing.T,
) {
	repo := &mockRepository{
		getOriginBookingFn: func(
			context.Context,
			uint,
		) (*BookingSnapshot, error) {
			return &BookingSnapshot{
				ID:        10,
				JobID:     7,
				ClientID:  5,
				CleanerID: 8,
				Status:    "closed",
			}, nil
		},
	}

	availabilityChecker := &mockAvailabilityChecker{
		listByCleanerIDFn: func(
			context.Context,
			uint,
		) ([]availabilitydomain.CleanerAvailability, error) {
			return []availabilitydomain.CleanerAvailability{}, nil
		},
	}

	service := NewService(
		repo,
		nil,
		nil,
		availabilityChecker,
	)

	request, err := service.BookingAgain(
		context.Background(),
		10,
		5,
		BookingAgainRequest{
			ScheduledAt:    "2026-09-10T10:00:00Z",
			ScheduledEndAt: "2026-09-10T12:00:00Z",
		},
	)

	if request != nil {
		t.Fatalf(
			"expected nil request, got %+v",
			request,
		)
	}

	if !errors.Is(err, ErrCleanerUnavailable) {
		t.Fatalf(
			"expected ErrForbidden, got %v",
			err,
		)
	}
}

func TestService_BookingAgain_WrongDateAvailability_Edge(
	t *testing.T,
) {
	repo := &mockRepository{
		getOriginBookingFn: func(
			context.Context,
			uint,
		) (*BookingSnapshot, error) {
			return &BookingSnapshot{
				ID:        10,
				JobID:     7,
				ClientID:  5,
				CleanerID: 8,
				Status:    "completed",
			}, nil
		},
	}

	availabilityChecker := &mockAvailabilityChecker{
		listByCleanerIDFn: func(
			context.Context,
			uint,
		) ([]availabilitydomain.CleanerAvailability, error) {
			return []availabilitydomain.CleanerAvailability{
				{
					CleanerID:     8,
					AvailableDate: "2026-08-11",
					StartTime:     "09:00",
					EndTime:       "17:00",
					Status:        "available",
				},
			}, nil
		},
	}

	service := NewService(
		repo,
		nil,
		nil,
		availabilityChecker,
	)

	request, err := service.BookingAgain(
		context.Background(),
		10,
		5,
		BookingAgainRequest{
			ScheduledAt:    "2026-09-10T10:00:00Z",
			ScheduledEndAt: "2026-09-10T12:00:00Z",
		},
	)

	if request != nil {
		t.Fatalf(
			"expected nil request, got %+v",
			request,
		)
	}

	if !errors.Is(err, ErrCleanerUnavailable) {
		t.Fatalf(
			"expected ErrForbidden, got %v",
			err,
		)
	}
}

func TestService_BookingAgain_UnavailableStatus_Edge(
	t *testing.T,
) {
	repo := &mockRepository{
		getOriginBookingFn: func(
			context.Context,
			uint,
		) (*BookingSnapshot, error) {
			return &BookingSnapshot{
				ID:        10,
				JobID:     7,
				ClientID:  5,
				CleanerID: 8,
				Status:    "closed",
			}, nil
		},
	}

	availabilityChecker := &mockAvailabilityChecker{
		listByCleanerIDFn: func(
			context.Context,
			uint,
		) ([]availabilitydomain.CleanerAvailability, error) {
			return []availabilitydomain.CleanerAvailability{
				{
					CleanerID:     8,
					AvailableDate: "2026-08-10",
					StartTime:     "09:00",
					EndTime:       "17:00",
					Status:        "unavailable",
				},
			}, nil
		},
	}

	service := NewService(
		repo,
		nil,
		nil,
		availabilityChecker,
	)

	request, err := service.BookingAgain(
		context.Background(),
		10,
		5,
		BookingAgainRequest{
			ScheduledAt:    "2026-09-10T10:00:00Z",
			ScheduledEndAt: "2026-09-10T12:00:00Z",
		},
	)

	if request != nil {
		t.Fatalf(
			"expected nil request, got %+v",
			request,
		)
	}

	if !errors.Is(err, ErrCleanerUnavailable) {
		t.Fatalf(
			"expected ErrForbidden, got %v",
			err,
		)
	}
}

func TestService_BookingAgain_AvailableSuccess_Edge(
	t *testing.T,
) {
	transactionCalled := false

	repo := &mockRepository{
		getOriginBookingFn: func(
			context.Context,
			uint,
		) (*BookingSnapshot, error) {
			return &BookingSnapshot{
				ID:        10,
				JobID:     7,
				ClientID:  5,
				CleanerID: 8,
				Status:    "closed",
			}, nil
		},

		createRepeatBookingsFn: func(
			ctx context.Context,
			booking *BookingSnapshot,
			scheduledAt time.Time,
			scheduledEndAt time.Time,
		) (uint, error) {
			transactionCalled = true

			return 40, nil
		},

		createBookAgainRequestFn: func(
			ctx context.Context,
			request *RepeatBookingRequest,
		) error {
			request.ID = 50

			return nil
		},
	}

	blockChecker := &mockBlockChecker{
		isBlockedFn: func(
			context.Context,
			uint,
			uint,
		) (bool, error) {
			return false, nil
		},
	}

	availabilityChecker := &mockAvailabilityChecker{
		listByCleanerIDFn: func(
			ctx context.Context,
			cleanerID uint,
		) ([]availabilitydomain.CleanerAvailability, error) {
			if cleanerID != 8 {
				t.Fatalf(
					"expected cleaner ID 8, got %d",
					cleanerID,
				)
			}

			return []availabilitydomain.CleanerAvailability{
				{
					CleanerID:     8,
					AvailableDate: "2026-09-10",
					StartTime:     "09:00",
					EndTime:       "17:00",
					Status:        "available",
				},
			}, nil
		},
	}

	service := NewService(
		repo,
		nil,
		blockChecker,
		availabilityChecker,
	)

	request, err := service.BookingAgain(
		context.Background(),
		10,
		5,
		BookingAgainRequest{
			ScheduledAt:    "2026-09-10T10:00:00Z",
			ScheduledEndAt: "2026-09-10T12:00:00Z",
			Message:        "  Please return again.  ",
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if request == nil {
		t.Fatal("expected repeat-booking request")
	}

	if request.ID != 50 {
		t.Fatalf(
			"expected request ID 50, got %d",
			request.ID,
		)
	}

	if request.Message != "Please return again." {
		t.Fatalf(
			"expected trimmed message, got %q",
			request.Message,
		)
	}

	if !transactionCalled {
		t.Fatal(
			"expected CreateBookAgainTransaction to be called",
		)
	}
}
