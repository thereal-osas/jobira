package bookings

import (
	"context"
	"errors"
	"testing"
	"time"
)

type mockRepository struct {
	createFn                  func(context.Context, *Booking) error
	getByIDFn                 func(context.Context, uint) (*Booking, error)
	listByUserIDFn            func(context.Context, uint) ([]Booking, error)
	updateStatusFn            func(context.Context, uint, string) error
	completeFn                func(context.Context, uint, string) error
	cancelFn                  func(context.Context, uint, uint, string) error
	getUserEmailFn            func(context.Context, uint) (string, error)
	closeFn                   func(context.Context, uint, int, bool, string) error
	markJobCompletedFn        func(context.Context, uint) error
	markApplicationAcceptedFn func(context.Context, uint) error
	closeWithTransactionFn    func(context.Context, CloseBookingTransaction) error
}

func (m *mockRepository) Create(ctx context.Context, booking *Booking) error {
	if m.createFn != nil {
		return m.createFn(ctx, booking)
	}
	return nil
}

func (m *mockRepository) GetByID(ctx context.Context, id uint) (*Booking, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *mockRepository) ListByUserID(ctx context.Context, userID uint) ([]Booking, error) {
	if m.listByUserIDFn != nil {
		return m.listByUserIDFn(ctx, userID)
	}
	return nil, nil
}

func (m *mockRepository) UpdateStatus(ctx context.Context, bookingID uint, status string) error {
	if m.updateStatusFn != nil {
		return m.updateStatusFn(ctx, bookingID, status)
	}
	return nil
}

func (m *mockRepository) Complete(ctx context.Context, bookingID uint, status string) error {
	if m.completeFn != nil {
		return m.completeFn(ctx, bookingID, status)
	}
	return nil
}
func (m *mockRepository) Cancel(
	ctx context.Context,
	bookingID uint,
	cancelledBy uint,
	reason string,
) error {
	if m.cancelFn != nil {
		return m.cancelFn(
			ctx,
			bookingID,
			cancelledBy,
			reason,
		)
	}

	return nil
}

func (m *mockRepository) GetUserEmail(ctx context.Context, userID uint) (string, error) {
	if m.getUserEmailFn != nil {
		return m.getUserEmailFn(ctx, userID)
	}
	return "", nil
}

func (m *mockRepository) Close(ctx context.Context, bookingID uint, rating int, wouldHireAgain bool, comment string) error {
	if m.closeFn != nil {
		return m.closeFn(ctx, bookingID, rating, wouldHireAgain, comment)
	}
	return nil
}

func (m *mockRepository) MarkJobCompleted(ctx context.Context, jobID uint) error {
	if m.markJobCompletedFn != nil {
		return m.markJobCompletedFn(ctx, jobID)
	}
	return nil
}

func (m *mockRepository) MarkApplicationAccepted(ctx context.Context, applicationID uint) error {
	if m.markApplicationAcceptedFn != nil {
		return m.markApplicationAcceptedFn(ctx, applicationID)
	}
	return nil
}

func (m *mockRepository) CloseWithTransaction(ctx context.Context, tx CloseBookingTransaction) error {
	if m.closeWithTransactionFn != nil {
		return m.closeWithTransactionFn(ctx, tx)
	}
	return nil
}

func (m *mockRepository) HasScheduleConflict(ctx context.Context, cleanerID uint, startAt time.Time, endAt time.Time, excludeBookingID uint,
) (bool, error) {
	return false, nil
}

func TestNewService(t *testing.T) {
	repo := &mockRepository{}

	service := NewService(repo, nil, nil, nil, nil, nil, nil)

	if service == nil {
		t.Fatal("expected service")
	}

	if service.repo != repo {
		t.Fatal("repository not assigned")
	}
}

func TestService_Create_InvalidInput(t *testing.T) {
	service := NewService(&mockRepository{}, nil, nil, nil, nil, nil, nil)

	_, err := service.Create(context.Background(), 0, CreateBookingRequest{})

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput got %v", err)
	}
}

func TestService_GetByID_InvalidInput(t *testing.T) {
	service := NewService(&mockRepository{}, nil, nil, nil, nil, nil, nil)

	_, err := service.GetByID(context.Background(), 0, 0, "client")

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput got %v", err)
	}
}

func TestService_ListMine_InvalidInput(t *testing.T) {
	service := NewService(&mockRepository{}, nil, nil, nil, nil, nil, nil)

	_, err := service.ListMine(context.Background(), 0)

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput got %v", err)
	}
}

func TestService_Create_Success(t *testing.T) {
	repo := &mockRepository{
		createFn: func(ctx context.Context, booking *Booking) error {
			booking.ID = 12
			return nil
		},
	}

	service := NewService(repo, nil, nil, nil, nil, nil, nil)

	startAt := time.Now().Add(48 * time.Hour)
	endAt := startAt.Add(2 * time.Hour)

	req := CreateBookingRequest{
		JobID:          10,
		CleanerID:      30,
		ScheduledAt:    startAt.Format(time.RFC3339),
		ScheduledEndAt: endAt.Format(time.RFC3339),
	}

	booking, err := service.Create(context.Background(), 5, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if booking == nil {
		t.Fatal("expected booking")
	}

	if booking.ID != 12 {
		t.Fatalf("expected booking ID 12, got %d", booking.ID)
	}

	if booking.JobID != 10 {
		t.Fatalf("expected job ID 10, got %d", booking.JobID)
	}

	if booking.ClientID != 5 {
		t.Fatalf("expected client ID 5, got %d", booking.ClientID)
	}

	if booking.CleanerID != 30 {
		t.Fatalf("expected cleaner ID 8, got %d", booking.CleanerID)
	}

	if booking.Status != "pending" {
		t.Fatalf("expected pending status, got %q", booking.Status)
	}

	if booking.ScheduledAt == nil {
		t.Fatal("expected scheduled time")
	}
}

func TestService_Create_WithoutScheduledAt(t *testing.T) {
	repo := &mockRepository{
		createFn: func(ctx context.Context, booking *Booking) error {
			booking.ID = 13
			return nil
		},
	}

	service := NewService(repo, nil, nil, nil, nil, nil, nil)

	req := CreateBookingRequest{
		JobID:       3,
		CleanerID:   8,
		ScheduledAt: "   ",
	}

	booking, err := service.Create(context.Background(), 5, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if booking.ScheduledAt != nil {
		t.Fatalf("expected nil scheduled time, got %v", booking.ScheduledAt)
	}
}

func TestService_Create_InvalidScheduledAt(t *testing.T) {
	service := NewService(&mockRepository{}, nil, nil, nil, nil, nil, nil)

	req := CreateBookingRequest{
		JobID:       3,
		CleanerID:   8,
		ScheduledAt: "tomorrow morning",
	}

	booking, err := service.Create(context.Background(), 5, req)

	if booking != nil {
		t.Fatalf("expected nil booking, got %+v", booking)
	}

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestService_Create_RepositoryError(t *testing.T) {
	expectedErr := errors.New("create failed")

	repo := &mockRepository{
		createFn: func(ctx context.Context, booking *Booking) error {
			return expectedErr
		},
	}

	service := NewService(repo, nil, nil, nil, nil, nil, nil)

	req := CreateBookingRequest{
		JobID:     3,
		CleanerID: 8,
	}

	booking, err := service.Create(context.Background(), 5, req)

	if booking != nil {
		t.Fatalf("expected nil booking, got %+v", booking)
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
}

func TestService_Create_MarksApplicationAccepted(t *testing.T) {
	applicationID := uint(21)
	markCalled := false

	repo := &mockRepository{
		createFn: func(ctx context.Context, booking *Booking) error {
			booking.ID = 14
			return nil
		},
		markApplicationAcceptedFn: func(ctx context.Context, id uint) error {
			markCalled = true

			if id != applicationID {
				t.Fatalf("expected application ID %d, got %d", applicationID, id)
			}

			return nil
		},
	}

	service := NewService(repo, nil, nil, nil, nil, nil, nil)

	req := CreateBookingRequest{
		JobID:         3,
		ApplicationID: &applicationID,
		CleanerID:     8,
	}

	booking, err := service.Create(context.Background(), 5, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if booking == nil {
		t.Fatal("expected booking")
	}

	if !markCalled {
		t.Fatal("expected MarkApplicationAccepted to be called")
	}
}

func TestService_Create_MarkApplicationAcceptedError(t *testing.T) {
	applicationID := uint(21)
	expectedErr := errors.New("application update failed")

	repo := &mockRepository{
		createFn: func(ctx context.Context, booking *Booking) error {
			booking.ID = 14
			return nil
		},
		markApplicationAcceptedFn: func(ctx context.Context, id uint) error {
			return expectedErr
		},
	}

	service := NewService(repo, nil, nil, nil, nil, nil, nil)

	req := CreateBookingRequest{
		JobID:         3,
		ApplicationID: &applicationID,
		CleanerID:     8,
	}

	booking, err := service.Create(context.Background(), 5, req)

	if booking != nil {
		t.Fatalf("expected nil booking, got %+v", booking)
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
}

func TestService_GetByID_ClientSuccess(t *testing.T) {
	expectedBooking := &Booking{
		ID:        10,
		ClientID:  5,
		CleanerID: 8,
		Status:    "pending",
	}

	repo := &mockRepository{
		getByIDFn: func(ctx context.Context, id uint) (*Booking, error) {
			if id != 10 {
				t.Fatalf("expected booking ID 10, got %d", id)
			}

			return expectedBooking, nil
		},
	}

	service := NewService(repo, nil, nil, nil, nil, nil, nil)

	booking, err := service.GetByID(context.Background(), 10, 5, "client")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if booking != expectedBooking {
		t.Fatal("expected returned booking")
	}
}

func TestService_GetByID_CleanerSuccess(t *testing.T) {
	expectedBooking := &Booking{
		ID:        10,
		ClientID:  5,
		CleanerID: 8,
	}

	repo := &mockRepository{
		getByIDFn: func(ctx context.Context, id uint) (*Booking, error) {
			return expectedBooking, nil
		},
	}

	service := NewService(repo, nil, nil, nil, nil, nil, nil)

	booking, err := service.GetByID(context.Background(), 10, 8, "cleaner")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if booking != expectedBooking {
		t.Fatal("expected returned booking")
	}
}

func TestService_GetByID_AdminSuccess(t *testing.T) {
	expectedBooking := &Booking{
		ID:        10,
		ClientID:  5,
		CleanerID: 8,
	}

	repo := &mockRepository{
		getByIDFn: func(ctx context.Context, id uint) (*Booking, error) {
			return expectedBooking, nil
		},
	}

	service := NewService(repo, nil, nil, nil, nil, nil, nil)

	booking, err := service.GetByID(context.Background(), 10, 99, "admin")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if booking != expectedBooking {
		t.Fatal("expected returned booking")
	}
}

func TestService_GetByID_Forbidden(t *testing.T) {
	repo := &mockRepository{
		getByIDFn: func(ctx context.Context, id uint) (*Booking, error) {
			return &Booking{
				ID:        10,
				ClientID:  5,
				CleanerID: 8,
			}, nil
		},
	}

	service := NewService(repo, nil, nil, nil, nil, nil, nil)

	booking, err := service.GetByID(context.Background(), 10, 20, "client")

	if booking != nil {
		t.Fatalf("expected nil booking, got %+v", booking)
	}

	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestService_GetByID_RepositoryError(t *testing.T) {
	expectedErr := errors.New("database failed")

	repo := &mockRepository{
		getByIDFn: func(ctx context.Context, id uint) (*Booking, error) {
			return nil, expectedErr
		},
	}

	service := NewService(repo, nil, nil, nil, nil, nil, nil)

	booking, err := service.GetByID(context.Background(), 10, 5, "client")

	if booking != nil {
		t.Fatalf("expected nil booking, got %+v", booking)
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
}

func TestService_ListMine_Success(t *testing.T) {
	expectedBookings := []Booking{
		{
			ID:        1,
			ClientID:  5,
			CleanerID: 8,
		},
		{
			ID:        2,
			ClientID:  9,
			CleanerID: 5,
		},
	}

	repo := &mockRepository{
		listByUserIDFn: func(ctx context.Context, userID uint) ([]Booking, error) {
			if userID != 5 {
				t.Fatalf("expected user ID 5, got %d", userID)
			}

			return expectedBookings, nil
		},
	}

	service := NewService(repo, nil, nil, nil, nil, nil, nil)

	bookings, err := service.ListMine(context.Background(), 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(bookings) != 2 {
		t.Fatalf("expected 2 bookings, got %d", len(bookings))
	}

	if bookings[0].ID != 1 {
		t.Fatalf("expected first booking ID 1, got %d", bookings[0].ID)
	}
}

func TestService_ListMine_RepositoryError(t *testing.T) {
	expectedErr := errors.New("list failed")

	repo := &mockRepository{
		listByUserIDFn: func(ctx context.Context, userID uint) ([]Booking, error) {
			return nil, expectedErr
		},
	}

	service := NewService(repo, nil, nil, nil, nil, nil, nil)

	bookings, err := service.ListMine(context.Background(), 5)

	if bookings != nil {
		t.Fatalf("expected nil bookings, got %+v", bookings)
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
}

func TestService_Confirm_Success(t *testing.T) {
	repo := &mockRepository{
		getByIDFn: func(ctx context.Context, id uint) (*Booking, error) {
			return &Booking{
				ID:        1,
				ClientID:  5,
				CleanerID: 8,
				Status:    "pending",
			}, nil
		},
		updateStatusFn: func(ctx context.Context, bookingID uint, status string) error {
			if bookingID != 1 {
				t.Fatalf("expected bookingID 1 got %d", bookingID)
			}

			if status != "confirmed" {
				t.Fatalf("expected confirmed got %s", status)
			}

			return nil
		},
	}

	service := NewService(repo, nil, nil, nil, nil, nil, nil)

	booking, err := service.Confirm(context.Background(), 1, 8, "cleaner")
	if err != nil {
		t.Fatal(err)
	}

	if booking.Status != "pending" && booking.ID != 1 {
		t.Fatal("unexpected booking returned")
	}
}

func TestService_Confirm_InvalidStatus(t *testing.T) {
	repo := &mockRepository{
		getByIDFn: func(ctx context.Context, id uint) (*Booking, error) {
			return &Booking{
				ID:        1,
				ClientID:  5,
				CleanerID: 8,
				Status:    "completed",
			}, nil
		},
	}

	service := NewService(repo, nil, nil, nil, nil, nil, nil)

	booking, err := service.Confirm(context.Background(), 1, 8, "cleaner")

	if booking != nil {
		t.Fatal("expected nil booking")
	}

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatal(err)
	}
}

func TestService_Confirm_Forbidden(t *testing.T) {
	repo := &mockRepository{
		getByIDFn: func(ctx context.Context, id uint) (*Booking, error) {
			return &Booking{
				ID:        1,
				ClientID:  5,
				CleanerID: 8,
				Status:    "pending",
			}, nil
		},
	}

	service := NewService(repo, nil, nil, nil, nil, nil, nil)

	booking, err := service.Confirm(context.Background(), 1, 22, "cleaner")

	if booking != nil {
		t.Fatal("expected nil booking")
	}

	if !errors.Is(err, ErrForbidden) {
		t.Fatal(err)
	}
}

func TestService_Confirm_UpdateStatusError(t *testing.T) {
	expected := errors.New("update failed")

	repo := &mockRepository{
		getByIDFn: func(ctx context.Context, id uint) (*Booking, error) {
			return &Booking{
				ID:        1,
				ClientID:  5,
				CleanerID: 8,
				Status:    "pending",
			}, nil
		},
		updateStatusFn: func(ctx context.Context, bookingID uint, status string) error {
			return expected
		},
	}

	service := NewService(repo, nil, nil, nil, nil, nil, nil)

	booking, err := service.Confirm(context.Background(), 1, 8, "cleaner")

	if booking != nil {
		t.Fatal("expected nil booking")
	}

	if !errors.Is(err, expected) {
		t.Fatal(err)
	}
}

func TestService_Start_Success(t *testing.T) {
	getCalls := 0

	repo := &mockRepository{
		getByIDFn: func(ctx context.Context, id uint) (*Booking, error) {
			getCalls++

			if getCalls == 1 {
				return &Booking{
					ID:        1,
					ClientID:  5,
					CleanerID: 8,
					Status:    "confirmed",
				}, nil
			}

			return &Booking{
				ID:        1,
				ClientID:  5,
				CleanerID: 8,
				Status:    "in_progress",
			}, nil
		},
		updateStatusFn: func(ctx context.Context, bookingID uint, status string) error {
			if bookingID != 1 {
				t.Fatalf("expected booking ID 1, got %d", bookingID)
			}

			if status != "in_progress" {
				t.Fatalf("expected in_progress, got %q", status)
			}

			return nil
		},
	}

	service := NewService(repo, nil, nil, nil, nil, nil, nil)

	booking, err := service.Start(context.Background(), 1, 8, "cleaner")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if booking == nil {
		t.Fatal("expected booking")
	}

	if booking.Status != "in_progress" {
		t.Fatalf("expected in_progress, got %q", booking.Status)
	}

	if getCalls != 2 {
		t.Fatalf("expected GetByID to be called twice, got %d", getCalls)
	}
}

func TestService_Start_InvalidStatus(t *testing.T) {
	repo := &mockRepository{
		getByIDFn: func(ctx context.Context, id uint) (*Booking, error) {
			return &Booking{
				ID:        1,
				ClientID:  5,
				CleanerID: 8,
				Status:    "pending",
			}, nil
		},
	}

	service := NewService(repo, nil, nil, nil, nil, nil, nil)

	booking, err := service.Start(context.Background(), 1, 8, "cleaner")

	if booking != nil {
		t.Fatalf("expected nil booking, got %+v", booking)
	}

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestService_Start_Forbidden(t *testing.T) {
	repo := &mockRepository{
		getByIDFn: func(ctx context.Context, id uint) (*Booking, error) {
			return &Booking{
				ID:        1,
				ClientID:  5,
				CleanerID: 8,
				Status:    "confirmed",
			}, nil
		},
	}

	service := NewService(repo, nil, nil, nil, nil, nil, nil)

	booking, err := service.Start(context.Background(), 1, 5, "client")

	if booking != nil {
		t.Fatalf("expected nil booking, got %+v", booking)
	}

	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestService_Start_UpdateStatusError(t *testing.T) {
	expectedErr := errors.New("update failed")

	repo := &mockRepository{
		getByIDFn: func(ctx context.Context, id uint) (*Booking, error) {
			return &Booking{
				ID:        1,
				ClientID:  5,
				CleanerID: 8,
				Status:    "confirmed",
			}, nil
		},
		updateStatusFn: func(ctx context.Context, bookingID uint, status string) error {
			return expectedErr
		},
	}

	service := NewService(repo, nil, nil, nil, nil, nil, nil)

	booking, err := service.Start(context.Background(), 1, 8, "cleaner")

	if booking != nil {
		t.Fatalf("expected nil booking, got %+v", booking)
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
}

func TestService_Start_FinalGetByIDError(t *testing.T) {
	expectedErr := errors.New("reload failed")
	getCalls := 0

	repo := &mockRepository{
		getByIDFn: func(ctx context.Context, id uint) (*Booking, error) {
			getCalls++

			if getCalls == 1 {
				return &Booking{
					ID:        1,
					ClientID:  5,
					CleanerID: 8,
					Status:    "confirmed",
				}, nil
			}

			return nil, expectedErr
		},
		updateStatusFn: func(ctx context.Context, bookingID uint, status string) error {
			return nil
		},
	}

	service := NewService(repo, nil, nil, nil, nil, nil, nil)

	booking, err := service.Start(context.Background(), 1, 8, "cleaner")

	if booking != nil {
		t.Fatalf("expected nil booking, got %+v", booking)
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
}

func TestService_Complete_Success(t *testing.T) {
	getCalls := 0
	completeCalled := false
	jobCompletedCalled := false

	repo := &mockRepository{
		getByIDFn: func(ctx context.Context, id uint) (*Booking, error) {
			getCalls++

			if getCalls == 1 {
				return &Booking{
					ID:        1,
					JobID:     7,
					ClientID:  5,
					CleanerID: 8,
					Status:    "in_progress",
				}, nil
			}

			return &Booking{
				ID:        1,
				JobID:     7,
				ClientID:  5,
				CleanerID: 8,
				Status:    "completed",
			}, nil
		},
		completeFn: func(ctx context.Context, bookingID uint, status string) error {
			completeCalled = true

			if bookingID != 1 {
				t.Fatalf("expected booking ID 1, got %d", bookingID)
			}

			if status != "completed" {
				t.Fatalf("expected completed, got %q", status)
			}

			return nil
		},
		markJobCompletedFn: func(ctx context.Context, jobID uint) error {
			jobCompletedCalled = true

			if jobID != 7 {
				t.Fatalf("expected job ID 7, got %d", jobID)
			}

			return nil
		},
	}

	service := NewService(repo, nil, nil, nil, nil, nil, nil)

	booking, err := service.Complete(context.Background(), 1, 8, "cleaner")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if booking == nil {
		t.Fatal("expected booking")
	}

	if booking.Status != "completed" {
		t.Fatalf("expected completed, got %q", booking.Status)
	}

	if !completeCalled {
		t.Fatal("expected Complete to be called")
	}

	if !jobCompletedCalled {
		t.Fatal("expected MarkJobCompleted to be called")
	}
}

func TestService_Complete_FromConfirmedStatus(t *testing.T) {
	getCalls := 0

	repo := &mockRepository{
		getByIDFn: func(ctx context.Context, id uint) (*Booking, error) {
			getCalls++

			if getCalls == 1 {
				return &Booking{
					ID:        1,
					JobID:     7,
					ClientID:  5,
					CleanerID: 8,
					Status:    "confirmed",
				}, nil
			}

			return &Booking{
				ID:        1,
				JobID:     7,
				ClientID:  5,
				CleanerID: 8,
				Status:    "completed",
			}, nil
		},
		completeFn: func(ctx context.Context, bookingID uint, status string) error {
			return nil
		},
		markJobCompletedFn: func(ctx context.Context, jobID uint) error {
			return nil
		},
	}

	service := NewService(repo, nil, nil, nil, nil, nil, nil)

	booking, err := service.Complete(context.Background(), 1, 8, "cleaner")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if booking.Status != "completed" {
		t.Fatalf("expected completed, got %q", booking.Status)
	}
}

func TestService_Complete_InvalidStatus(t *testing.T) {
	repo := &mockRepository{
		getByIDFn: func(ctx context.Context, id uint) (*Booking, error) {
			return &Booking{
				ID:        1,
				JobID:     7,
				ClientID:  5,
				CleanerID: 8,
				Status:    "pending",
			}, nil
		},
	}

	service := NewService(repo, nil, nil, nil, nil, nil, nil)

	booking, err := service.Complete(context.Background(), 1, 8, "cleaner")

	if booking != nil {
		t.Fatalf("expected nil booking, got %+v", booking)
	}

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestService_Complete_Forbidden(t *testing.T) {
	repo := &mockRepository{
		getByIDFn: func(ctx context.Context, id uint) (*Booking, error) {
			return &Booking{
				ID:        1,
				JobID:     7,
				ClientID:  5,
				CleanerID: 8,
				Status:    "in_progress",
			}, nil
		},
	}

	service := NewService(repo, nil, nil, nil, nil, nil, nil)

	booking, err := service.Complete(context.Background(), 1, 5, "client")

	if booking != nil {
		t.Fatalf("expected nil booking, got %+v", booking)
	}

	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestService_Complete_RepositoryError(t *testing.T) {
	expectedErr := errors.New("complete failed")

	repo := &mockRepository{
		getByIDFn: func(ctx context.Context, id uint) (*Booking, error) {
			return &Booking{
				ID:        1,
				JobID:     7,
				ClientID:  5,
				CleanerID: 8,
				Status:    "in_progress",
			}, nil
		},
		completeFn: func(ctx context.Context, bookingID uint, status string) error {
			return expectedErr
		},
	}

	service := NewService(repo, nil, nil, nil, nil, nil, nil)

	booking, err := service.Complete(context.Background(), 1, 8, "cleaner")

	if booking != nil {
		t.Fatalf("expected nil booking, got %+v", booking)
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
}

func TestService_Complete_MarkJobCompletedError(t *testing.T) {
	expectedErr := errors.New("job update failed")

	repo := &mockRepository{
		getByIDFn: func(ctx context.Context, id uint) (*Booking, error) {
			return &Booking{
				ID:        1,
				JobID:     7,
				ClientID:  5,
				CleanerID: 8,
				Status:    "in_progress",
			}, nil
		},
		completeFn: func(ctx context.Context, bookingID uint, status string) error {
			return nil
		},
		markJobCompletedFn: func(ctx context.Context, jobID uint) error {
			return expectedErr
		},
	}

	service := NewService(repo, nil, nil, nil, nil, nil, nil)

	booking, err := service.Complete(context.Background(), 1, 8, "cleaner")

	if booking != nil {
		t.Fatalf("expected nil booking, got %+v", booking)
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
}

func TestService_Complete_FinalGetByIDError(t *testing.T) {
	expectedErr := errors.New("reload failed")
	getCalls := 0

	repo := &mockRepository{
		getByIDFn: func(ctx context.Context, id uint) (*Booking, error) {
			getCalls++

			if getCalls == 1 {
				return &Booking{
					ID:        1,
					JobID:     7,
					ClientID:  5,
					CleanerID: 8,
					Status:    "in_progress",
				}, nil
			}

			return nil, expectedErr
		},
		completeFn: func(ctx context.Context, bookingID uint, status string) error {
			return nil
		},
		markJobCompletedFn: func(ctx context.Context, jobID uint) error {
			return nil
		},
	}

	service := NewService(repo, nil, nil, nil, nil, nil, nil)

	booking, err := service.Complete(context.Background(), 1, 8, "cleaner")

	if booking != nil {
		t.Fatalf("expected nil booking, got %+v", booking)
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
}

func TestService_Cancel_InvalidReason(t *testing.T) {
	service := NewService(&mockRepository{}, nil, nil, nil, nil, nil, nil)

	booking, err := service.Cancel(
		context.Background(),
		1,
		5,
		"client",
		CancelBookingRequest{},
	)

	if booking != nil {
		t.Fatal("expected nil booking")
	}

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestService_Cancel_Success(t *testing.T) {
	getCalls := 0
	cancelCalled := false

	repo := &mockRepository{
		getByIDFn: func(ctx context.Context, id uint) (*Booking, error) {
			getCalls++

			if getCalls == 1 {
				return &Booking{
					ID:        1,
					ClientID:  5,
					CleanerID: 8,
					Status:    "confirmed",
				}, nil
			}

			return &Booking{
				ID:                 1,
				ClientID:           5,
				CleanerID:          8,
				Status:             "cancelled",
				CancellationReason: "Cleaner unavailable",
			}, nil
		},
		cancelFn: func(
			_ context.Context,
			bookingID uint,
			cancelledBy uint,
			reason string,
		) error {
			cancelCalled = true

			if bookingID != 1 {
				t.Fatalf("expected booking id 1, got %d", bookingID)
			}

			if cancelledBy != 5 {
				t.Fatalf("expected cancelled by 5, got %d", cancelledBy)
			}

			if reason != "Cleaner unavailable" {
				t.Fatalf("unexpected reason %q", reason)
			}

			return nil
		},
	}
	service := NewService(repo, nil, nil, nil, nil, nil, nil)

	booking, err := service.Cancel(
		context.Background(),
		1,
		5,
		"client",
		CancelBookingRequest{
			Reason: "Cleaner unavailable",
		},
	)

	if err != nil {
		t.Fatal(err)
	}

	if booking == nil {
		t.Fatal("expected booking")
	}

	if booking.Status != "cancelled" {
		t.Fatalf("expected cancelled, got %q", booking.Status)
	}

	if !cancelCalled {
		t.Fatal("expected Cancel() to be called")
	}
}

func TestService_Cancel_CompletedBooking(t *testing.T) {
	repo := &mockRepository{
		getByIDFn: func(ctx context.Context, id uint) (*Booking, error) {
			return &Booking{
				ID:        1,
				ClientID:  5,
				CleanerID: 8,
				Status:    "completed",
			}, nil
		},
	}

	service := NewService(repo, nil, nil, nil, nil, nil, nil)

	booking, err := service.Cancel(
		context.Background(),
		1,
		5,
		"client",
		CancelBookingRequest{
			Reason: "Reason",
		},
	)

	if booking != nil {
		t.Fatal("expected nil booking")
	}

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatal(err)
	}
}

func TestService_Cancel_ClosedBooking(t *testing.T) {
	repo := &mockRepository{
		getByIDFn: func(ctx context.Context, id uint) (*Booking, error) {
			return &Booking{
				ID:        1,
				ClientID:  5,
				CleanerID: 8,
				Status:    "closed",
			}, nil
		},
	}

	service := NewService(repo, nil, nil, nil, nil, nil, nil)

	booking, err := service.Cancel(
		context.Background(),
		1,
		5,
		"client",
		CancelBookingRequest{
			Reason: "Reason",
		},
	)

	if booking != nil {
		t.Fatal("expected nil booking")
	}

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatal(err)
	}
}

func TestService_Cancel_Forbidden(t *testing.T) {
	repo := &mockRepository{
		getByIDFn: func(ctx context.Context, id uint) (*Booking, error) {
			return &Booking{
				ID:        1,
				ClientID:  5,
				CleanerID: 8,
				Status:    "confirmed",
			}, nil
		},
	}

	service := NewService(repo, nil, nil, nil, nil, nil, nil)

	booking, err := service.Cancel(
		context.Background(),
		1,
		99,
		"client",
		CancelBookingRequest{
			Reason: "Reason",
		},
	)

	if booking != nil {
		t.Fatal("expected nil booking")
	}

	if !errors.Is(err, ErrForbidden) {
		t.Fatal(err)
	}
}

func TestService_Cancel_GetByIDError(t *testing.T) {
	expectedErr := errors.New("booking lookup failed")

	repo := &mockRepository{
		getByIDFn: func(ctx context.Context, id uint) (*Booking, error) {
			return nil, expectedErr
		},
	}

	service := NewService(repo, nil, nil, nil, nil, nil, nil)

	booking, err := service.Cancel(
		context.Background(),
		1,
		5,
		"client",
		CancelBookingRequest{
			Reason: "Plans changed",
		},
	)

	if booking != nil {
		t.Fatalf("expected nil booking, got %+v", booking)
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
}

func TestService_Cancel_CleanerSuccess(t *testing.T) {
	getCalls := 0

	repo := &mockRepository{
		getByIDFn: func(ctx context.Context, id uint) (*Booking, error) {
			getCalls++

			if getCalls == 1 {
				return &Booking{
					ID:        1,
					ClientID:  5,
					CleanerID: 8,
					Status:    "confirmed",
				}, nil
			}

			return &Booking{
				ID:                 1,
				ClientID:           5,
				CleanerID:          8,
				Status:             "cancelled",
				CancellationReason: "Unable to attend",
			}, nil
		},
		cancelFn: func(
			_ context.Context,
			bookingID uint,
			cancelledBy uint,
			reason string,
		) error {
			if bookingID != 1 {
				t.Fatalf("expected booking id 1, got %d", bookingID)
			}

			if cancelledBy != 8 {
				t.Fatalf("expected cancelled by 8, got %d", cancelledBy)
			}

			if reason != "Unable to attend" {
				t.Fatalf("unexpected reason %q", reason)
			}

			return nil
		},
	}

	service := NewService(repo, nil, nil, nil, nil, nil, nil)

	booking, err := service.Cancel(
		context.Background(),
		1,
		8,
		"cleaner",
		CancelBookingRequest{
			Reason: "Unable to attend",
		},
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if booking == nil {
		t.Fatal("expected booking")
	}

	if booking.Status != "cancelled" {
		t.Fatalf("expected cancelled, got %q", booking.Status)
	}
}

func TestService_Cancel_AdminSuccess(t *testing.T) {
	getCalls := 0

	repo := &mockRepository{
		getByIDFn: func(ctx context.Context, id uint) (*Booking, error) {
			getCalls++

			if getCalls == 1 {
				return &Booking{
					ID:        1,
					ClientID:  5,
					CleanerID: 8,
					Status:    "confirmed",
				}, nil
			}

			return &Booking{
				ID:                 1,
				ClientID:           5,
				CleanerID:          8,
				Status:             "cancelled",
				CancellationReason: "Administrative cancellation",
			}, nil
		},
		cancelFn: func(
			_ context.Context,
			bookingID uint,
			cancelledBy uint,
			reason string,
		) error {
			if bookingID != 1 {
				t.Fatalf("expected booking id 1, got %d", bookingID)
			}

			if cancelledBy != 99 {
				t.Fatalf("expected cancelled by 99, got %d", cancelledBy)
			}

			// around line 1449
			if reason != "Administrative cancellation" {
				t.Fatalf("unexpected reason %q", reason)
			}

			return nil
		},
	}

	service := NewService(repo, nil, nil, nil, nil, nil, nil)

	booking, err := service.Cancel(
		context.Background(),
		1,
		99,
		"admin",
		CancelBookingRequest{
			Reason: "Administrative cancellation",
		},
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if booking == nil {
		t.Fatal("expected booking")
	}

	if booking.Status != "cancelled" {
		t.Fatalf("expected cancelled, got %q", booking.Status)
	}
}

func TestService_Cancel_RepositoryError(t *testing.T) {
	expectedErr := errors.New("cancel update failed")

	repo := &mockRepository{
		getByIDFn: func(ctx context.Context, id uint) (*Booking, error) {
			return &Booking{
				ID:        1,
				ClientID:  5,
				CleanerID: 8,
				Status:    "confirmed",
			}, nil
		},
		cancelFn: func(
			_ context.Context,
			bookingID uint,
			cancelledBy uint,
			reason string,
		) error {
			if bookingID != 1 {
				t.Fatalf("expected booking id 1, got %d", bookingID)
			}

			if cancelledBy != 5 {
				t.Fatalf("expected cancelled by 5, got %d", cancelledBy)
			}

			// around line 1508
			if reason != "Plans changed" {
				t.Fatalf("unexpected reason %q", reason)
			}

			return expectedErr
		},
	}

	service := NewService(repo, nil, nil, nil, nil, nil, nil)

	booking, err := service.Cancel(
		context.Background(),
		1,
		5,
		"client",
		CancelBookingRequest{
			Reason: "Plans changed",
		},
	)

	if booking != nil {
		t.Fatalf("expected nil booking, got %+v", booking)
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
}

func TestService_Cancel_FinalGetByIDError(t *testing.T) {
	expectedErr := errors.New("booking reload failed")
	getCalls := 0

	repo := &mockRepository{
		getByIDFn: func(ctx context.Context, id uint) (*Booking, error) {
			getCalls++

			if getCalls == 1 {
				return &Booking{
					ID:        1,
					ClientID:  5,
					CleanerID: 8,
					Status:    "confirmed",
				}, nil
			}

			return nil, expectedErr
		},
		cancelFn: func(
			_ context.Context,
			bookingID uint,
			cancelledBy uint,
			reason string,
		) error {
			if bookingID != 1 {
				t.Fatalf("expected booking id 1, got %d", bookingID)
			}

			if cancelledBy != 5 {
				t.Fatalf("expected cancelled by 5, got %d", cancelledBy)
			}

			// around line 1570
			if reason != "Plans changed" {
				t.Fatalf("unexpected reason %q", reason)
			}

			return nil
		},
	}
	service := NewService(repo, nil, nil, nil, nil, nil, nil)

	booking, err := service.Cancel(
		context.Background(),
		1,
		5,
		"client",
		CancelBookingRequest{
			Reason: "Plans changed",
		},
	)

	if booking != nil {
		t.Fatalf("expected nil booking, got %+v", booking)
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}

	if getCalls != 2 {
		t.Fatalf("expected GetByID to be called twice, got %d", getCalls)
	}
}

func TestService_Close_InvalidBookingID(t *testing.T) {
	service := NewService(&mockRepository{}, nil, nil, nil, nil, nil, nil)

	booking, err := service.Close(
		context.Background(),
		0,
		5,
		"client",
		CloseBookingRequest{
			ClientRating:   5,
			ClosureComment: "Everything went well",
		},
	)

	if booking != nil {
		t.Fatalf("expected nil booking, got %+v", booking)
	}

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestService_Close_InvalidUserID(t *testing.T) {
	service := NewService(&mockRepository{}, nil, nil, nil, nil, nil, nil)

	booking, err := service.Close(
		context.Background(),
		1,
		0,
		"client",
		CloseBookingRequest{
			ClientRating:   5,
			ClosureComment: "Everything went well",
		},
	)

	if booking != nil {
		t.Fatalf("expected nil booking, got %+v", booking)
	}

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestService_Close_RatingTooLow(t *testing.T) {
	service := NewService(&mockRepository{}, nil, nil, nil, nil, nil, nil)

	booking, err := service.Close(
		context.Background(),
		1,
		5,
		"client",
		CloseBookingRequest{
			ClientRating:   0,
			ClosureComment: "Everything went well",
		},
	)

	if booking != nil {
		t.Fatalf("expected nil booking, got %+v", booking)
	}

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestService_Close_RatingTooHigh(t *testing.T) {
	service := NewService(&mockRepository{}, nil, nil, nil, nil, nil, nil)

	booking, err := service.Close(
		context.Background(),
		1,
		5,
		"client",
		CloseBookingRequest{
			ClientRating:   6,
			ClosureComment: "Everything went well",
		},
	)

	if booking != nil {
		t.Fatalf("expected nil booking, got %+v", booking)
	}

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestService_Close_BlankComment(t *testing.T) {
	service := NewService(&mockRepository{}, nil, nil, nil, nil, nil, nil)

	booking, err := service.Close(
		context.Background(),
		1,
		5,
		"client",
		CloseBookingRequest{
			ClientRating:   5,
			ClosureComment: "   ",
		},
	)

	if booking != nil {
		t.Fatalf("expected nil booking, got %+v", booking)
	}

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestService_Close_GetByIDError(t *testing.T) {
	expectedErr := errors.New("booking lookup failed")

	repo := &mockRepository{
		getByIDFn: func(ctx context.Context, id uint) (*Booking, error) {
			return nil, expectedErr
		},
	}

	service := NewService(repo, nil, nil, nil, nil, nil, nil)

	booking, err := service.Close(
		context.Background(),
		1,
		5,
		"client",
		CloseBookingRequest{
			ClientRating:   5,
			ClosureComment: "Everything went well",
		},
	)

	if booking != nil {
		t.Fatalf("expected nil booking, got %+v", booking)
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
}

func TestService_Close_CleanerForbidden(t *testing.T) {
	repo := &mockRepository{
		getByIDFn: func(ctx context.Context, id uint) (*Booking, error) {
			return &Booking{
				ID:        1,
				JobID:     7,
				ClientID:  5,
				CleanerID: 8,
				Status:    "completed",
			}, nil
		},
	}

	service := NewService(repo, nil, nil, nil, nil, nil, nil)

	booking, err := service.Close(
		context.Background(),
		1,
		8,
		"cleaner",
		CloseBookingRequest{
			ClientRating:   5,
			ClosureComment: "Everything went well",
		},
	)

	if booking != nil {
		t.Fatalf("expected nil booking, got %+v", booking)
	}

	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestService_Close_InvalidStatus(t *testing.T) {
	repo := &mockRepository{
		getByIDFn: func(ctx context.Context, id uint) (*Booking, error) {
			return &Booking{
				ID:        1,
				JobID:     7,
				ClientID:  5,
				CleanerID: 8,
				Status:    "in_progress",
			}, nil
		},
	}

	service := NewService(repo, nil, nil, nil, nil, nil, nil)

	booking, err := service.Close(
		context.Background(),
		1,
		5,
		"client",
		CloseBookingRequest{
			ClientRating:   5,
			ClosureComment: "Everything went well",
		},
	)

	if booking != nil {
		t.Fatalf("expected nil booking, got %+v", booking)
	}

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestService_Close_Success(t *testing.T) {
	getCalls := 0
	transactionCalled := false

	repo := &mockRepository{
		getByIDFn: func(ctx context.Context, id uint) (*Booking, error) {
			getCalls++

			if id != 1 {
				t.Fatalf("expected booking ID 1, got %d", id)
			}

			if getCalls == 1 {
				return &Booking{
					ID:        1,
					JobID:     7,
					ClientID:  5,
					CleanerID: 8,
					Status:    "completed",
				}, nil
			}

			rating := 5
			hireAgain := true

			return &Booking{
				ID:                        1,
				JobID:                     7,
				ClientID:                  5,
				CleanerID:                 8,
				Status:                    "closed",
				ClosureStatus:             "closed",
				ClosureComment:            "Excellent cleaner",
				ClientConfirmedCompletion: true,
				ClientWouldHireAgain:      &hireAgain,
				ClientRating:              &rating,
			}, nil
		},
		closeWithTransactionFn: func(ctx context.Context, transaction CloseBookingTransaction) error {
			transactionCalled = true

			if transaction.BookingID != 1 {
				t.Fatalf("expected booking ID 1, got %d", transaction.BookingID)
			}

			if transaction.JobID != 7 {
				t.Fatalf("expected job ID 7, got %d", transaction.JobID)
			}

			if transaction.ClientID != 5 {
				t.Fatalf("expected client ID 5, got %d", transaction.ClientID)
			}

			if transaction.CleanerID != 8 {
				t.Fatalf("expected cleaner ID 8, got %d", transaction.CleanerID)
			}

			if transaction.Rating != 5 {
				t.Fatalf("expected rating 5, got %d", transaction.Rating)
			}

			if !transaction.WouldHireAgain {
				t.Fatal("expected WouldHireAgain to be true")
			}

			if transaction.Comment != "Excellent cleaner" {
				t.Fatalf("expected trimmed comment, got %q", transaction.Comment)
			}

			if !transaction.AddToFavourites {
				t.Fatal("expected AddToFavourites to be true")
			}

			if !transaction.SetAsPreferred {
				t.Fatal("expected SetAsPreferred to be true")
			}

			if transaction.ChangedBy != 5 {
				t.Fatalf("expected ChangedBy 5, got %d", transaction.ChangedBy)
			}

			if transaction.PreviousStatus != "completed" {
				t.Fatalf(
					"expected previous status completed, got %q",
					transaction.PreviousStatus,
				)
			}

			if transaction.NotificationTitle != "Booking closed" {
				t.Fatalf(
					"expected notification title Booking closed, got %q",
					transaction.NotificationTitle,
				)
			}

			expectedMessage := "The client confirmed completion and closed the booking."
			if transaction.NotificationMessage != expectedMessage {
				t.Fatalf(
					"expected notification message %q, got %q",
					expectedMessage,
					transaction.NotificationMessage,
				)
			}

			return nil
		},
	}

	service := NewService(repo, nil, nil, nil, nil, nil, nil)

	booking, err := service.Close(
		context.Background(),
		1,
		5,
		"client",
		CloseBookingRequest{
			ClientConfirmedCompletion: true,
			ClientRating:              5,
			ClientWouldHireAgain:      true,
			ClosureComment:            "  Excellent cleaner  ",
			AddFavourites:             true,
			SetAsPreferred:            true,
		},
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if booking == nil {
		t.Fatal("expected booking")
	}

	if booking.Status != "closed" {
		t.Fatalf("expected closed status, got %q", booking.Status)
	}

	if booking.ClosureComment != "Excellent cleaner" {
		t.Fatalf(
			"expected closure comment Excellent cleaner, got %q",
			booking.ClosureComment,
		)
	}

	if !transactionCalled {
		t.Fatal("expected CloseWithTransaction to be called")
	}

	if getCalls != 2 {
		t.Fatalf("expected GetByID to be called twice, got %d", getCalls)
	}
}

func TestService_Close_AdminSuccess(t *testing.T) {
	getCalls := 0

	repo := &mockRepository{
		getByIDFn: func(ctx context.Context, id uint) (*Booking, error) {
			getCalls++

			if getCalls == 1 {
				return &Booking{
					ID:        1,
					JobID:     7,
					ClientID:  5,
					CleanerID: 8,
					Status:    "completed",
				}, nil
			}

			return &Booking{
				ID:        1,
				JobID:     7,
				ClientID:  5,
				CleanerID: 8,
				Status:    "closed",
			}, nil
		},
		closeWithTransactionFn: func(ctx context.Context, transaction CloseBookingTransaction) error {
			if transaction.ChangedBy != 99 {
				t.Fatalf("expected ChangedBy 99, got %d", transaction.ChangedBy)
			}

			return nil
		},
	}

	service := NewService(repo, nil, nil, nil, nil, nil, nil)

	booking, err := service.Close(
		context.Background(),
		1,
		99,
		"admin",
		CloseBookingRequest{
			ClientRating:   4,
			ClosureComment: "Closed by administrator",
		},
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if booking == nil {
		t.Fatal("expected booking")
	}

	if booking.Status != "closed" {
		t.Fatalf("expected closed status, got %q", booking.Status)
	}
}

func TestService_Close_TransactionError(t *testing.T) {
	expectedErr := errors.New("transaction failed")

	repo := &mockRepository{
		getByIDFn: func(ctx context.Context, id uint) (*Booking, error) {
			return &Booking{
				ID:        1,
				JobID:     7,
				ClientID:  5,
				CleanerID: 8,
				Status:    "completed",
			}, nil
		},
		closeWithTransactionFn: func(ctx context.Context, transaction CloseBookingTransaction) error {
			return expectedErr
		},
	}

	service := NewService(repo, nil, nil, nil, nil, nil, nil)

	booking, err := service.Close(
		context.Background(),
		1,
		5,
		"client",
		CloseBookingRequest{
			ClientRating:   5,
			ClosureComment: "Everything went well",
		},
	)

	if booking != nil {
		t.Fatalf("expected nil booking, got %+v", booking)
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
}

func TestService_Close_FinalGetByIDError(t *testing.T) {
	expectedErr := errors.New("booking reload failed")
	getCalls := 0

	repo := &mockRepository{
		getByIDFn: func(ctx context.Context, id uint) (*Booking, error) {
			getCalls++

			if getCalls == 1 {
				return &Booking{
					ID:        1,
					JobID:     7,
					ClientID:  5,
					CleanerID: 8,
					Status:    "completed",
				}, nil
			}

			return nil, expectedErr
		},
		closeWithTransactionFn: func(ctx context.Context, transaction CloseBookingTransaction) error {
			return nil
		},
	}

	service := NewService(repo, nil, nil, nil, nil, nil, nil)

	booking, err := service.Close(
		context.Background(),
		1,
		5,
		"client",
		CloseBookingRequest{
			ClientRating:   5,
			ClosureComment: "Everything went well",
		},
	)

	if booking != nil {
		t.Fatalf("expected nil booking, got %+v", booking)
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}

	if getCalls != 2 {
		t.Fatalf("expected GetByID to be called twice, got %d", getCalls)
	}
}
