package workhistory

import (
	"context"
	"errors"
	"testing"
)

type mockRepository struct {
	listByCleanerIDFn      func(context.Context, uint) ([]WorkHistoryEntry, error)
	listByClientIDFn       func(context.Context, uint) ([]WorkHistoryEntry, error)
	getByBookingIDFn       func(context.Context, uint, uint) (*WorkHistoryDetail, error)
	userCanAccessBookingFn func(context.Context, uint, uint) (bool, error)
}

func (m *mockRepository) ListByCleanerID(
	ctx context.Context,
	cleanerID uint,
) ([]WorkHistoryEntry, error) {
	if m.listByCleanerIDFn != nil {
		return m.listByCleanerIDFn(
			ctx,
			cleanerID,
		)
	}

	return []WorkHistoryEntry{}, nil
}

func (m *mockRepository) ListByClientID(
	ctx context.Context,
	clientID uint,
) ([]WorkHistoryEntry, error) {
	if m.listByClientIDFn != nil {
		return m.listByClientIDFn(
			ctx,
			clientID,
		)
	}

	return []WorkHistoryEntry{}, nil
}

func (m *mockRepository) GetByBookingID(
	ctx context.Context,
	bookingID uint,
	userID uint,
) (*WorkHistoryDetail, error) {
	if m.getByBookingIDFn != nil {
		return m.getByBookingIDFn(
			ctx,
			bookingID,
			userID,
		)
	}

	return nil, ErrHistoryNotFound
}

func (m *mockRepository) UserCanAccessBooking(
	ctx context.Context,
	bookingID uint,
	userID uint,
) (bool, error) {
	if m.userCanAccessBookingFn != nil {
		return m.userCanAccessBookingFn(
			ctx,
			bookingID,
			userID,
		)
	}

	return false, nil
}

func TestNewService(t *testing.T) {
	repo := &mockRepository{}

	service := NewService(repo)

	if service == nil {
		t.Fatal("expected service")
	}

	if service.repo != repo {
		t.Fatal("expected repository assigned")
	}
}

func TestService_ListMine_CleanerSuccess(
	t *testing.T,
) {
	repo := &mockRepository{
		listByCleanerIDFn: func(
			ctx context.Context,
			cleanerID uint,
		) ([]WorkHistoryEntry, error) {
			if cleanerID != 8 {
				t.Fatalf(
					"expected cleaner id 8, got %d",
					cleanerID,
				)
			}

			return []WorkHistoryEntry{
				{
					BookingID: 10,
					CleanerID: 8,
					ClientID:  5,
					Status:    "completed",
				},
			}, nil
		},
	}

	service := NewService(repo)

	history, err := service.ListMine(
		context.Background(),
		8,
		"cleaner",
	)
	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}

	if len(history) != 1 {
		t.Fatalf(
			"expected 1 history entry, got %d",
			len(history),
		)
	}

	if history[0].BookingID != 10 {
		t.Fatalf(
			"expected booking id 10, got %d",
			history[0].BookingID,
		)
	}
}

func TestService_ListMine_ClientSuccess(
	t *testing.T,
) {
	repo := &mockRepository{
		listByClientIDFn: func(
			ctx context.Context,
			clientID uint,
		) ([]WorkHistoryEntry, error) {
			if clientID != 5 {
				t.Fatalf(
					"expected client id 5, got %d",
					clientID,
				)
			}

			return []WorkHistoryEntry{
				{
					BookingID: 10,
					ClientID:  5,
					CleanerID: 8,
					Status:    "completed",
				},
			}, nil
		},
	}

	service := NewService(repo)

	history, err := service.ListMine(
		context.Background(),
		5,
		" client ",
	)
	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}

	if len(history) != 1 {
		t.Fatalf(
			"expected 1 history entry, got %d",
			len(history),
		)
	}
}

func TestService_ListMine_InvalidUserID(
	t *testing.T,
) {
	service := NewService(
		&mockRepository{},
	)

	_, err := service.ListMine(
		context.Background(),
		0,
		"cleaner",
	)

	if !errors.Is(
		err,
		ErrInvalidInput,
	) {
		t.Fatalf(
			"expected ErrInvalidInput, got %v",
			err,
		)
	}
}

func TestService_ListMine_ForbiddenRole(
	t *testing.T,
) {
	service := NewService(
		&mockRepository{},
	)

	_, err := service.ListMine(
		context.Background(),
		5,
		"user",
	)

	if !errors.Is(
		err,
		ErrForbidden,
	) {
		t.Fatalf(
			"expected ErrForbidden, got %v",
			err,
		)
	}
}

func TestService_ListMine_RepositoryError(
	t *testing.T,
) {
	expectedErr := errors.New(
		"database failed",
	)

	repo := &mockRepository{
		listByCleanerIDFn: func(
			context.Context,
			uint,
		) ([]WorkHistoryEntry, error) {
			return nil, expectedErr
		},
	}

	service := NewService(repo)

	_, err := service.ListMine(
		context.Background(),
		8,
		"cleaner",
	)

	if !errors.Is(
		err,
		expectedErr,
	) {
		t.Fatalf(
			"expected repository error, got %v",
			err,
		)
	}
}

func TestService_GetBookingHistory_ClientSuccess(
	t *testing.T,
) {
	repo := &mockRepository{
		userCanAccessBookingFn: func(
			ctx context.Context,
			bookingID uint,
			userID uint,
		) (bool, error) {
			if bookingID != 10 {
				t.Fatalf(
					"expected booking id 10, got %d",
					bookingID,
				)
			}

			if userID != 5 {
				t.Fatalf(
					"expected user id 5, got %d",
					userID,
				)
			}

			return true, nil
		},

		getByBookingIDFn: func(
			ctx context.Context,
			bookingID uint,
			userID uint,
		) (*WorkHistoryDetail, error) {
			return &WorkHistoryDetail{
				WorkHistoryEntry: WorkHistoryEntry{
					BookingID: bookingID,
					ClientID:  5,
					CleanerID: 8,
					Status:    "completed",
				},
			}, nil
		},
	}

	service := NewService(repo)

	history, err := service.GetBookingHistory(
		context.Background(),
		10,
		5,
		"client",
	)
	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}

	if history.BookingID != 10 {
		t.Fatalf(
			"expected booking id 10, got %d",
			history.BookingID,
		)
	}
}

func TestService_GetBookingHistory_CleanerSuccess(
	t *testing.T,
) {
	repo := &mockRepository{
		userCanAccessBookingFn: func(
			context.Context,
			uint,
			uint,
		) (bool, error) {
			return true, nil
		},

		getByBookingIDFn: func(
			ctx context.Context,
			bookingID uint,
			userID uint,
		) (*WorkHistoryDetail, error) {
			return &WorkHistoryDetail{
				WorkHistoryEntry: WorkHistoryEntry{
					BookingID: bookingID,
				},
			}, nil
		},
	}

	service := NewService(repo)

	history, err := service.GetBookingHistory(
		context.Background(),
		20,
		8,
		"cleaner",
	)
	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}

	if history.BookingID != 20 {
		t.Fatalf(
			"expected booking id 20, got %d",
			history.BookingID,
		)
	}
}

func TestService_GetBookingHistory_AdminSuccess(
	t *testing.T,
) {
	accessChecked := false

	repo := &mockRepository{
		userCanAccessBookingFn: func(
			context.Context,
			uint,
			uint,
		) (bool, error) {
			accessChecked = true
			return false, nil
		},

		getByBookingIDFn: func(
			ctx context.Context,
			bookingID uint,
			userID uint,
		) (*WorkHistoryDetail, error) {
			return &WorkHistoryDetail{
				WorkHistoryEntry: WorkHistoryEntry{
					BookingID: bookingID,
				},
			}, nil
		},
	}

	service := NewService(repo)

	history, err := service.GetBookingHistory(
		context.Background(),
		10,
		9,
		"admin",
	)
	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}

	if accessChecked {
		t.Fatal(
			"admin should not require booking access check",
		)
	}

	if history.BookingID != 10 {
		t.Fatalf(
			"expected booking id 10, got %d",
			history.BookingID,
		)
	}
}

func TestService_GetBookingHistory_InvalidBookingID(
	t *testing.T,
) {
	service := NewService(
		&mockRepository{},
	)

	_, err := service.GetBookingHistory(
		context.Background(),
		0,
		5,
		"client",
	)

	if !errors.Is(
		err,
		ErrInvalidInput,
	) {
		t.Fatalf(
			"expected ErrInvalidInput, got %v",
			err,
		)
	}
}

func TestService_GetBookingHistory_InvalidUserID(
	t *testing.T,
) {
	service := NewService(
		&mockRepository{},
	)

	_, err := service.GetBookingHistory(
		context.Background(),
		10,
		0,
		"client",
	)

	if !errors.Is(
		err,
		ErrInvalidInput,
	) {
		t.Fatalf(
			"expected ErrInvalidInput, got %v",
			err,
		)
	}
}

func TestService_GetBookingHistory_ForbiddenRole(
	t *testing.T,
) {
	service := NewService(
		&mockRepository{},
	)

	_, err := service.GetBookingHistory(
		context.Background(),
		10,
		5,
		"user",
	)

	if !errors.Is(
		err,
		ErrForbidden,
	) {
		t.Fatalf(
			"expected ErrForbidden, got %v",
			err,
		)
	}
}

func TestService_GetBookingHistory_NoAccess(
	t *testing.T,
) {
	repo := &mockRepository{
		userCanAccessBookingFn: func(
			context.Context,
			uint,
			uint,
		) (bool, error) {
			return false, nil
		},
	}

	service := NewService(repo)

	_, err := service.GetBookingHistory(
		context.Background(),
		10,
		5,
		"client",
	)

	if !errors.Is(
		err,
		ErrForbidden,
	) {
		t.Fatalf(
			"expected ErrForbidden, got %v",
			err,
		)
	}
}

func TestService_GetBookingHistory_AccessCheckError(
	t *testing.T,
) {
	expectedErr := errors.New(
		"access query failed",
	)

	repo := &mockRepository{
		userCanAccessBookingFn: func(
			context.Context,
			uint,
			uint,
		) (bool, error) {
			return false, expectedErr
		},
	}

	service := NewService(repo)

	_, err := service.GetBookingHistory(
		context.Background(),
		10,
		5,
		"client",
	)

	if !errors.Is(
		err,
		expectedErr,
	) {
		t.Fatalf(
			"expected access error, got %v",
			err,
		)
	}
}

func TestService_GetBookingHistory_NotFound(
	t *testing.T,
) {
	repo := &mockRepository{
		userCanAccessBookingFn: func(
			context.Context,
			uint,
			uint,
		) (bool, error) {
			return true, nil
		},

		getByBookingIDFn: func(
			context.Context,
			uint,
			uint,
		) (*WorkHistoryDetail, error) {
			return nil, ErrHistoryNotFound
		},
	}

	service := NewService(repo)

	_, err := service.GetBookingHistory(
		context.Background(),
		10,
		5,
		"client",
	)

	if !errors.Is(
		err,
		ErrHistoryNotFound,
	) {
		t.Fatalf(
			"expected ErrHistoryNotFound, got %v",
			err,
		)
	}
}
