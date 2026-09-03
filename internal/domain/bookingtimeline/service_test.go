package bookingtimeline

import (
	"context"
	"errors"
	"testing"
	"time"
)

type mockRepository struct {
	createHistoryFn func(
		context.Context,
		uint,
		uint,
		string,
		string,
		string,
	) error

	listByBookingIDFn func(
		context.Context,
		uint,
	) ([]StatusHistory, error)

	getBookingAccessFn func(
		context.Context,
		uint,
	) (uint, uint, string, error)
}

func (m *mockRepository) CreateHistory(
	ctx context.Context,
	bookingID uint,
	changedBy uint,
	fromStatus string,
	toStatus string,
	note string,
) error {
	if m.createHistoryFn != nil {
		return m.createHistoryFn(
			ctx,
			bookingID,
			changedBy,
			fromStatus,
			toStatus,
			note,
		)
	}

	return nil
}

func (m *mockRepository) ListByBookingID(
	ctx context.Context,
	bookingID uint,
) ([]StatusHistory, error) {
	if m.listByBookingIDFn != nil {
		return m.listByBookingIDFn(
			ctx,
			bookingID,
		)
	}

	return []StatusHistory{}, nil
}

func (m *mockRepository) GetBookingAccess(
	ctx context.Context,
	bookingID uint,
) (
	uint,
	uint,
	string,
	error,
) {
	if m.getBookingAccessFn != nil {
		return m.getBookingAccessFn(
			ctx,
			bookingID,
		)
	}

	return 0, 0, "", nil
}

func TestService_Record_Success(t *testing.T) {
	called := false

	repo := &mockRepository{
		createHistoryFn: func(
			ctx context.Context,
			bookingID uint,
			changedBy uint,
			fromStatus string,
			toStatus string,
			note string,
		) error {
			called = true

			if bookingID != 12 {
				t.Fatalf(
					"expected booking ID 12, got %d",
					bookingID,
				)
			}

			if changedBy != 8 {
				t.Fatalf(
					"expected changedBy 8, got %d",
					changedBy,
				)
			}

			if fromStatus != "accepted" {
				t.Fatalf(
					"expected accepted, got %q",
					fromStatus,
				)
			}

			if toStatus != "completed" {
				t.Fatalf(
					"expected completed, got %q",
					toStatus,
				)
			}

			if note != "Job completed successfully" {
				t.Fatalf(
					"unexpected note %q",
					note,
				)
			}

			return nil
		},
	}

	service := NewService(repo)

	err := service.Record(
		context.Background(),
		12,
		8,
		"  accepted  ",
		"  completed  ",
		"  Job completed successfully  ",
	)
	if err != nil {
		t.Fatalf(
			"expected nil error, got %v",
			err,
		)
	}

	if !called {
		t.Fatal(
			"expected CreateHistory to be called",
		)
	}
}

func TestService_Record_InvalidBookingID(
	t *testing.T,
) {
	service := NewService(
		&mockRepository{},
	)

	err := service.Record(
		context.Background(),
		0,
		8,
		"accepted",
		"completed",
		"",
	)

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf(
			"expected ErrInvalidInput, got %v",
			err,
		)
	}
}

func TestService_Record_EmptyToStatus(
	t *testing.T,
) {
	service := NewService(
		&mockRepository{},
	)

	err := service.Record(
		context.Background(),
		12,
		8,
		"accepted",
		"   ",
		"",
	)

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf(
			"expected ErrInvalidInput, got %v",
			err,
		)
	}
}

func TestService_Record_RepositoryError(
	t *testing.T,
) {
	expectedErr := errors.New(
		"create history failed",
	)

	repo := &mockRepository{
		createHistoryFn: func(
			context.Context,
			uint,
			uint,
			string,
			string,
			string,
		) error {
			return expectedErr
		},
	}

	service := NewService(repo)

	err := service.Record(
		context.Background(),
		12,
		8,
		"accepted",
		"completed",
		"",
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected repository error, got %v",
			err,
		)
	}
}

func TestService_GetTimeline_ClientAccess(
	t *testing.T,
) {
	now := time.Now()

	repo := &mockRepository{
		getBookingAccessFn: func(
			context.Context,
			uint,
		) (uint, uint, string, error) {
			return 5, 8, "accepted", nil
		},

		listByBookingIDFn: func(
			context.Context,
			uint,
		) ([]StatusHistory, error) {
			return []StatusHistory{
				{
					ID:         1,
					BookingID:  12,
					FromStatus: "pending",
					ToStatus:   "accepted",
					CreatedAt:  now,
				},
			}, nil
		},
	}

	service := NewService(repo)

	timeline, err := service.GetTimeline(
		context.Background(),
		12,
		5,
		"client",
	)
	if err != nil {
		t.Fatalf(
			"expected nil error, got %v",
			err,
		)
	}

	if timeline.BookingID != 12 {
		t.Fatalf(
			"expected booking ID 12, got %d",
			timeline.BookingID,
		)
	}

	if timeline.CurrentStatus != "accepted" {
		t.Fatalf(
			"expected accepted, got %q",
			timeline.CurrentStatus,
		)
	}

	if len(timeline.History) != 1 {
		t.Fatalf(
			"expected 1 history item, got %d",
			len(timeline.History),
		)
	}
}

func TestService_GetTimeline_CleanerAccess(
	t *testing.T,
) {
	repo := &mockRepository{
		getBookingAccessFn: func(
			context.Context,
			uint,
		) (uint, uint, string, error) {
			return 5, 8, "confirmed", nil
		},
	}

	service := NewService(repo)

	timeline, err := service.GetTimeline(
		context.Background(),
		12,
		8,
		"cleaner",
	)
	if err != nil {
		t.Fatalf(
			"expected nil error, got %v",
			err,
		)
	}

	if timeline.CurrentStatus != "confirmed" {
		t.Fatalf(
			"expected confirmed, got %q",
			timeline.CurrentStatus,
		)
	}
}

func TestService_GetTimeline_AdminAccess(
	t *testing.T,
) {
	repo := &mockRepository{
		getBookingAccessFn: func(
			context.Context,
			uint,
		) (uint, uint, string, error) {
			return 5, 8, "completed", nil
		},
	}

	service := NewService(repo)

	timeline, err := service.GetTimeline(
		context.Background(),
		12,
		99,
		"admin",
	)
	if err != nil {
		t.Fatalf(
			"expected nil error, got %v",
			err,
		)
	}

	if timeline.CurrentStatus != "completed" {
		t.Fatalf(
			"expected completed, got %q",
			timeline.CurrentStatus,
		)
	}
}

func TestService_GetTimeline_Forbidden(
	t *testing.T,
) {
	listCalled := false

	repo := &mockRepository{
		getBookingAccessFn: func(
			context.Context,
			uint,
		) (uint, uint, string, error) {
			return 5, 8, "accepted", nil
		},

		listByBookingIDFn: func(
			context.Context,
			uint,
		) ([]StatusHistory, error) {
			listCalled = true
			return nil, nil
		},
	}

	service := NewService(repo)

	timeline, err := service.GetTimeline(
		context.Background(),
		12,
		25,
		"client",
	)

	if timeline != nil {
		t.Fatal(
			"expected nil timeline",
		)
	}

	if !errors.Is(err, ErrForbidden) {
		t.Fatalf(
			"expected ErrForbidden, got %v",
			err,
		)
	}

	if listCalled {
		t.Fatal(
			"history should not be loaded for forbidden user",
		)
	}
}

func TestService_GetTimeline_InvalidBookingID(
	t *testing.T,
) {
	service := NewService(
		&mockRepository{},
	)

	timeline, err := service.GetTimeline(
		context.Background(),
		0,
		5,
		"client",
	)

	if timeline != nil {
		t.Fatal(
			"expected nil timeline",
		)
	}

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf(
			"expected ErrInvalidInput, got %v",
			err,
		)
	}
}

func TestService_GetTimeline_InvalidUserID(
	t *testing.T,
) {
	service := NewService(
		&mockRepository{},
	)

	timeline, err := service.GetTimeline(
		context.Background(),
		12,
		0,
		"client",
	)

	if timeline != nil {
		t.Fatal(
			"expected nil timeline",
		)
	}

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf(
			"expected ErrInvalidInput, got %v",
			err,
		)
	}
}

func TestService_GetTimeline_BookingNotFound(
	t *testing.T,
) {
	repo := &mockRepository{
		getBookingAccessFn: func(
			context.Context,
			uint,
		) (uint, uint, string, error) {
			return 0, 0, "", ErrBookingNotFound
		},
	}

	service := NewService(repo)

	timeline, err := service.GetTimeline(
		context.Background(),
		999,
		5,
		"client",
	)

	if timeline != nil {
		t.Fatal(
			"expected nil timeline",
		)
	}

	if !errors.Is(err, ErrBookingNotFound) {
		t.Fatalf(
			"expected ErrBookingNotFound, got %v",
			err,
		)
	}
}

func TestService_GetTimeline_AccessRepositoryError(
	t *testing.T,
) {
	expectedErr := errors.New(
		"booking access failed",
	)

	repo := &mockRepository{
		getBookingAccessFn: func(
			context.Context,
			uint,
		) (uint, uint, string, error) {
			return 0, 0, "", expectedErr
		},
	}

	service := NewService(repo)

	timeline, err := service.GetTimeline(
		context.Background(),
		12,
		5,
		"client",
	)

	if timeline != nil {
		t.Fatal(
			"expected nil timeline",
		)
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected repository error, got %v",
			err,
		)
	}
}

func TestService_GetTimeline_HistoryRepositoryError(
	t *testing.T,
) {
	expectedErr := errors.New(
		"history failed",
	)

	repo := &mockRepository{
		getBookingAccessFn: func(
			context.Context,
			uint,
		) (uint, uint, string, error) {
			return 5, 8, "accepted", nil
		},

		listByBookingIDFn: func(
			context.Context,
			uint,
		) ([]StatusHistory, error) {
			return nil, expectedErr
		},
	}

	service := NewService(repo)

	timeline, err := service.GetTimeline(
		context.Background(),
		12,
		5,
		"client",
	)

	if timeline != nil {
		t.Fatal(
			"expected nil timeline",
		)
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected history error, got %v",
			err,
		)
	}
}
