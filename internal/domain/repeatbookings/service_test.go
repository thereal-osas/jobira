package repeatbookings

import (
	"context"
	"errors"
	"testing"
	"time"
)

type mockRepository struct {
	createJobFn                  func(context.Context, uint, CreateRepeatBookingRequest) (uint, error)
	createInvitationFn           func(context.Context, uint, uint, uint, string) error
	createRepeatBookingFn        func(context.Context, *RepeatBooking) error
	listByClientIDFn             func(context.Context, uint) ([]RepeatBooking, error)
	getOriginBookingFn           func(context.Context, uint) (*BookingSnapshot, error)
	createBookAgainRequestFn     func(context.Context, *RepeatBookingRequest) error
	createRepeatBookingsFn       func(context.Context, *BookingSnapshot, time.Time, time.Time) (uint, error)
	createBookAgainTransactionFn func(context.Context, BookAgainTransaction) (*RepeatBookingRequest, error)
}

func (m *mockRepository) CreateJob(
	ctx context.Context,
	clientID uint,
	req CreateRepeatBookingRequest,
) (uint, error) {
	if m.createJobFn != nil {
		return m.createJobFn(ctx, clientID, req)
	}

	return 0, nil
}

func (m *mockRepository) CreateInvitation(
	ctx context.Context,
	jobID uint,
	clientID uint,
	cleanerID uint,
	message string,
) error {
	if m.createInvitationFn != nil {
		return m.createInvitationFn(
			ctx,
			jobID,
			clientID,
			cleanerID,
			message,
		)
	}

	return nil
}

func (m *mockRepository) CreateRepeatBooking(
	ctx context.Context,
	booking *RepeatBooking,
) error {
	if m.createRepeatBookingFn != nil {
		return m.createRepeatBookingFn(ctx, booking)
	}

	return nil
}

func (m *mockRepository) ListByClientID(
	ctx context.Context,
	clientID uint,
) ([]RepeatBooking, error) {
	if m.listByClientIDFn != nil {
		return m.listByClientIDFn(ctx, clientID)
	}

	return nil, nil
}

func (m *mockRepository) GetOriginBooking(
	ctx context.Context,
	bookingID uint,
) (*BookingSnapshot, error) {
	if m.getOriginBookingFn != nil {
		return m.getOriginBookingFn(ctx, bookingID)
	}

	return nil, nil
}

func (m *mockRepository) CreateBookAgainRequest(
	ctx context.Context,
	request *RepeatBookingRequest,
) error {
	if m.createBookAgainRequestFn != nil {
		return m.createBookAgainRequestFn(ctx, request)
	}

	return nil
}

func (m *mockRepository) CreateRepeatBookings(
	ctx context.Context,
	booking *BookingSnapshot,
	scheduledAt time.Time,
	scheduledEndAt time.Time,
) (uint, error) {
	if m.createRepeatBookingsFn != nil {
		return m.createRepeatBookingsFn(
			ctx,
			booking,
			scheduledAt,
			scheduledEndAt,
		)
	}

	return 0, nil
}

func (m *mockRepository) CreateBookAgainTransaction(
	ctx context.Context,
	input BookAgainTransaction,
) (*RepeatBookingRequest, error) {
	if m.createBookAgainTransactionFn != nil {
		return m.createBookAgainTransactionFn(ctx, input)
	}

	return nil, nil
}

func TestNewService(t *testing.T) {
	repo := &mockRepository{}

	service := NewService(
		repo,
		nil,
		nil,
		nil,
	)

	if service == nil {
		t.Fatal("expected service")
	}

	if service.repo != repo {
		t.Fatal("expected repository to be assigned")
	}

	if service.notificationsService != nil {
		t.Fatal("expected nil notifications service")
	}

	if service.blockChecker != nil {
		t.Fatal("expected nil block checker")
	}

	if service.availabilityService != nil {
		t.Fatal("expected nil availability service")
	}
}

func TestService_Create_Success(t *testing.T) {
	repo := &mockRepository{
		createJobFn: func(
			ctx context.Context,
			clientID uint,
			req CreateRepeatBookingRequest,
		) (uint, error) {
			if clientID != 5 {
				t.Fatalf(
					"expected client ID 5, got %d",
					clientID,
				)
			}

			if req.ListingType != "shift" {
				t.Fatalf(
					"expected default listing type shift, got %q",
					req.ListingType,
				)
			}

			return 12, nil
		},
		createInvitationFn: func(
			ctx context.Context,
			jobID uint,
			clientID uint,
			cleanerID uint,
			message string,
		) error {
			if jobID != 12 {
				t.Fatalf("expected job ID 12, got %d", jobID)
			}

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

			return nil
		},
		createRepeatBookingFn: func(
			ctx context.Context,
			booking *RepeatBooking,
		) error {
			booking.ID = 20
			booking.CreatedAt = time.Now()
			booking.UpdatedAt = booking.CreatedAt

			return nil
		},
	}

	service := NewService(repo, nil, nil, nil)

	booking, err := service.Create(
		context.Background(),
		5,
		8,
		CreateRepeatBookingRequest{
			Title:       "Weekly domestic clean",
			Description: "Clean a two-bedroom flat",
			Location:    "East London",
			JobType:     "domestic",
			Budget:      70,
			Message:     "Would you like to clean for me again?",
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if booking == nil {
		t.Fatal("expected repeat booking")
	}

	if booking.ID != 20 {
		t.Fatalf("expected ID 20, got %d", booking.ID)
	}

	if booking.ClientID != 5 {
		t.Fatalf(
			"expected client ID 5, got %d",
			booking.ClientID,
		)
	}

	if booking.CleanerID != 8 {
		t.Fatalf(
			"expected cleaner ID 8, got %d",
			booking.CleanerID,
		)
	}

	if booking.JobID != 12 {
		t.Fatalf(
			"expected job ID 12, got %d",
			booking.JobID,
		)
	}

	if booking.Status != "sent" {
		t.Fatalf(
			"expected status sent, got %q",
			booking.Status,
		)
	}
}

func TestService_Create_InvalidInput(t *testing.T) {
	service := NewService(
		&mockRepository{},
		nil,
		nil,
		nil,
	)

	validRequest := CreateRepeatBookingRequest{
		Title:       "Weekly clean",
		Description: "Clean a flat",
		Location:    "London",
		JobType:     "domestic",
		Budget:      70,
	}

	tests := []struct {
		name      string
		clientID  uint
		cleanerID uint
		req       CreateRepeatBookingRequest
	}{
		{
			name:      "zero client ID",
			clientID:  0,
			cleanerID: 8,
			req:       validRequest,
		},
		{
			name:      "zero cleaner ID",
			clientID:  5,
			cleanerID: 0,
			req:       validRequest,
		},
		{
			name:      "same client and cleaner",
			clientID:  5,
			cleanerID: 5,
			req:       validRequest,
		},
		{
			name:      "missing title",
			clientID:  5,
			cleanerID: 8,
			req: CreateRepeatBookingRequest{
				Description: "Clean a flat",
				Location:    "London",
				JobType:     "domestic",
			},
		},
		{
			name:      "missing description",
			clientID:  5,
			cleanerID: 8,
			req: CreateRepeatBookingRequest{
				Title:    "Weekly clean",
				Location: "London",
				JobType:  "domestic",
			},
		},
		{
			name:      "missing location",
			clientID:  5,
			cleanerID: 8,
			req: CreateRepeatBookingRequest{
				Title:       "Weekly clean",
				Description: "Clean a flat",
				JobType:     "domestic",
			},
		},
		{
			name:      "missing job type",
			clientID:  5,
			cleanerID: 8,
			req: CreateRepeatBookingRequest{
				Title:       "Weekly clean",
				Description: "Clean a flat",
				Location:    "London",
			},
		},
		{
			name:      "negative budget",
			clientID:  5,
			cleanerID: 8,
			req: CreateRepeatBookingRequest{
				Title:       "Weekly clean",
				Description: "Clean a flat",
				Location:    "London",
				JobType:     "domestic",
				Budget:      -1,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			booking, err := service.Create(
				context.Background(),
				test.clientID,
				test.cleanerID,
				test.req,
			)

			if booking != nil {
				t.Fatalf(
					"expected nil booking, got %+v",
					booking,
				)
			}

			if !errors.Is(err, ErrInvalidInput) {
				t.Fatalf(
					"expected ErrInvalidInput, got %v",
					err,
				)
			}
		})
	}
}

func TestService_Create_CreateJobError(t *testing.T) {
	expectedErr := errors.New("create job failed")

	repo := &mockRepository{
		createJobFn: func(
			context.Context,
			uint,
			CreateRepeatBookingRequest,
		) (uint, error) {
			return 0, expectedErr
		},
	}

	service := NewService(repo, nil, nil, nil)

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
		t.Fatalf("expected nil booking, got %+v", booking)
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
}

func TestService_Create_InvitationError(t *testing.T) {
	expectedErr := errors.New("invitation failed")

	repo := &mockRepository{
		createJobFn: func(
			context.Context,
			uint,
			CreateRepeatBookingRequest,
		) (uint, error) {
			return 12, nil
		},
		createInvitationFn: func(
			context.Context,
			uint,
			uint,
			uint,
			string,
		) error {
			return expectedErr
		},
	}

	service := NewService(repo, nil, nil, nil)

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
		t.Fatalf("expected nil booking, got %+v", booking)
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
}

func TestService_Create_RepeatBookingError(t *testing.T) {
	expectedErr := errors.New("repeat booking failed")

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
			return expectedErr
		},
	}

	service := NewService(repo, nil, nil, nil)

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
		t.Fatalf("expected nil booking, got %+v", booking)
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
}

func TestService_ListMine_Success(t *testing.T) {
	repo := &mockRepository{
		listByClientIDFn: func(
			ctx context.Context,
			clientID uint,
		) ([]RepeatBooking, error) {
			return []RepeatBooking{
				{
					ID:        1,
					ClientID:  clientID,
					CleanerID: 8,
				},
				{
					ID:        2,
					ClientID:  clientID,
					CleanerID: 9,
				},
			}, nil
		},
	}

	service := NewService(repo, nil, nil, nil)

	bookings, err := service.ListMine(
		context.Background(),
		5,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(bookings) != 2 {
		t.Fatalf(
			"expected 2 bookings, got %d",
			len(bookings),
		)
	}
}

func TestService_ListMine_InvalidInput(t *testing.T) {
	service := NewService(
		&mockRepository{},
		nil,
		nil,
		nil,
	)

	bookings, err := service.ListMine(
		context.Background(),
		0,
	)

	if bookings != nil {
		t.Fatalf(
			"expected nil bookings, got %+v",
			bookings,
		)
	}

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestService_ListMine_RepositoryError(t *testing.T) {
	expectedErr := errors.New("list failed")

	repo := &mockRepository{
		listByClientIDFn: func(
			context.Context,
			uint,
		) ([]RepeatBooking, error) {
			return nil, expectedErr
		},
	}

	service := NewService(repo, nil, nil, nil)

	bookings, err := service.ListMine(
		context.Background(),
		5,
	)

	if bookings != nil {
		t.Fatalf(
			"expected nil bookings, got %+v",
			bookings,
		)
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
}
func TestService_BookingAgain_Success(t *testing.T) {
	applicationID := uint(33)

	repo := &mockRepository{
		getOriginBookingFn: func(
			ctx context.Context,
			bookingID uint,
		) (*BookingSnapshot, error) {
			if bookingID != 10 {
				t.Fatalf(
					"expected booking ID 10, got %d",
					bookingID,
				)
			}

			return &BookingSnapshot{
				ID:            10,
				JobID:         7,
				ApplicationID: &applicationID,
				ClientID:      5,
				CleanerID:     8,
				Status:        "completed",
			}, nil
		},

		createRepeatBookingsFn: func(
			ctx context.Context,
			booking *BookingSnapshot,
			scheduledAt time.Time,
			scheduledEndAt time.Time,
		) (uint, error) {
			if booking.ID != 10 {
				t.Fatalf(
					"expected original booking ID 10, got %d",
					booking.ID,
				)
			}

			return 40, nil
		},

		createBookAgainRequestFn: func(
			ctx context.Context,
			request *RepeatBookingRequest,
		) error {
			if request.OriginalBookingID != 10 {
				t.Fatalf(
					"expected original booking ID 10, got %d",
					request.OriginalBookingID,
				)
			}

			if request.ClientID != 5 {
				t.Fatalf(
					"expected client ID 5, got %d",
					request.ClientID,
				)
			}

			if request.CleanerID != 8 {
				t.Fatalf(
					"expected cleaner ID 8, got %d",
					request.CleanerID,
				)
			}

			if request.Status != "created" {
				t.Fatalf(
					"expected status created, got %q",
					request.Status,
				)
			}

			request.ID = 50

			return nil
		},
	}

	service := NewService(repo, nil, nil, nil)

	request, err := service.BookingAgain(
		context.Background(),
		10,
		5,
		BookingAgainRequest{
			ScheduledAt:    "2026-09-10T10:00:00Z",
			ScheduledEndAt: "2026-09-10T12:00:00Z",
			Message:        "Please bring the same equipment.",
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if request == nil {
		t.Fatal("expected repeat-booking request")
	}

	if request.ID != 50 {
		t.Fatalf("expected request ID 50, got %d", request.ID)
	}

	if request.NewBookingID == nil {
		t.Fatal("expected new booking ID")
	}

	if *request.NewBookingID != 40 {
		t.Fatalf(
			"expected new booking ID 40, got %d",
			*request.NewBookingID,
		)
	}
}

func TestService_BookingAgain_InvalidInput(t *testing.T) {
	service := NewService(
		&mockRepository{},
		nil,
		nil,
		nil,
	)

	tests := []struct {
		name              string
		originalBookingID uint
		clientID          uint
		req               BookingAgainRequest
	}{
		{
			name:              "zero original booking ID",
			originalBookingID: 0,
			clientID:          5,
			req: BookingAgainRequest{
				ScheduledAt:    "2026-09-10T10:00:00Z",
				ScheduledEndAt: "2026-09-10T12:00:00Z",
			},
		},
		{
			name:              "zero client ID",
			originalBookingID: 10,
			clientID:          0,
			req: BookingAgainRequest{
				ScheduledAt:    "2026-09-10T10:00:00Z",
				ScheduledEndAt: "2026-09-10T12:00:00Z",
			},
		},
		{
			name:              "missing scheduled time",
			originalBookingID: 10,
			clientID:          5,
			req:               BookingAgainRequest{},
		},
		{
			name:              "invalid scheduled time",
			originalBookingID: 10,
			clientID:          5,
			req: BookingAgainRequest{
				ScheduledAt:    "tomorrow morning",
				ScheduledEndAt: "2026-09-10T12:00:00Z",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request, err := service.BookingAgain(
				context.Background(),
				test.originalBookingID,
				test.clientID,
				test.req,
			)

			if request != nil {
				t.Fatalf(
					"expected nil request, got %+v",
					request,
				)
			}

			if !errors.Is(err, ErrInvalidInput) {
				t.Fatalf(
					"expected ErrInvalidInput, got %v",
					err,
				)
			}
		})
	}
}

func TestService_BookingAgain_GetOriginBookingError(t *testing.T) {
	expectedErr := errors.New("booking lookup failed")

	repo := &mockRepository{
		getOriginBookingFn: func(
			context.Context,
			uint,
		) (*BookingSnapshot, error) {
			return nil, expectedErr
		},
	}

	service := NewService(repo, nil, nil, nil)

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
		t.Fatalf("expected nil request, got %+v", request)
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
}

func TestService_BookingAgain_ForbiddenClient(t *testing.T) {
	repo := &mockRepository{
		getOriginBookingFn: func(
			context.Context,
			uint,
		) (*BookingSnapshot, error) {
			return &BookingSnapshot{
				ID:        10,
				ClientID:  5,
				CleanerID: 8,
				Status:    "closed",
			}, nil
		},
	}

	service := NewService(repo, nil, nil, nil)

	request, err := service.BookingAgain(
		context.Background(),
		10,
		99,
		BookingAgainRequest{
			ScheduledAt:    "2026-09-10T10:00:00Z",
			ScheduledEndAt: "2026-09-10T12:00:00Z",
		},
	)

	if request != nil {
		t.Fatalf("expected nil request, got %+v", request)
	}

	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestService_BookingAgain_InvalidOriginalStatus(t *testing.T) {
	repo := &mockRepository{
		getOriginBookingFn: func(
			context.Context,
			uint,
		) (*BookingSnapshot, error) {
			return &BookingSnapshot{
				ID:        10,
				ClientID:  5,
				CleanerID: 8,
				Status:    "pending",
			}, nil
		},
	}

	service := NewService(repo, nil, nil, nil)

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
		t.Fatalf("expected nil request, got %+v", request)
	}

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestService_BookingAgain_TransactionError(t *testing.T) {
	expectedErr := errors.New("transaction failed")

repo := &mockRepository{
	getOriginBookingFn: func(
		ctx context.Context,
		bookingID uint,
	) (*BookingSnapshot, error) {
		return &BookingSnapshot{
			ID:        10,
			JobID:     7,
			ClientID:  5,
			CleanerID: 8,
			Status:    "completed",
		}, nil
	},

	createRepeatBookingsFn: func(
		ctx context.Context,
		booking *BookingSnapshot,
		scheduledAt time.Time,
		scheduledEndAt time.Time,
	) (uint, error) {
		return 40, nil
	},

	createBookAgainRequestFn: func(
		ctx context.Context,
		request *RepeatBookingRequest,
	) error {
		return expectedErr
	},
}

	service := NewService(repo, nil, nil, nil)

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
		t.Fatalf("expected nil request, got %+v", request)
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
}
