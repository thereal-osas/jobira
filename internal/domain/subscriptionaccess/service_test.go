package subscriptionaccess

import (
	"context"
	"errors"
	"testing"
)

type mockRepository struct {
	getCleanerAccessStatusFn     func(ctx context.Context, userID uint) (*CleanerAccessStatus, error)
	getClientAccessStatusFn      func(ctx context.Context, userID uint) (*ClientAccessStatus, error)
	ensureDailyAccessFn          func(ctx context.Context, userID uint) error
	isLaunchGraceActiveFn        func(ctx context.Context) (bool, error)
	incrementApplicationsTodayFn func(ctx context.Context, userID uint) error
	incrementJobsPostTodayFn     func(ctx context.Context, userID uint) error
}

func (m *mockRepository) GetCleanerAccessStatus(
	ctx context.Context,
	userID uint,
) (*CleanerAccessStatus, error) {
	if m.getCleanerAccessStatusFn != nil {
		return m.getCleanerAccessStatusFn(ctx, userID)
	}

	return nil, nil
}

func (m *mockRepository) GetClientAccessStatus(
	ctx context.Context,
	userID uint,
) (*ClientAccessStatus, error) {
	if m.getClientAccessStatusFn != nil {
		return m.getClientAccessStatusFn(ctx, userID)
	}

	return nil, nil
}

func (m *mockRepository) EnsureDailyAccess(
	ctx context.Context,
	userID uint,
) error {
	if m.ensureDailyAccessFn != nil {
		return m.ensureDailyAccessFn(ctx, userID)
	}

	return nil
}

func (m *mockRepository) IsLaunchGraceActive(
	ctx context.Context,
) (bool, error) {
	if m.isLaunchGraceActiveFn != nil {
		return m.isLaunchGraceActiveFn(ctx)
	}

	return false, nil
}

func (m *mockRepository) IncrementApplicationsToday(
	ctx context.Context,
	userID uint,
) error {
	if m.incrementApplicationsTodayFn != nil {
		return m.incrementApplicationsTodayFn(ctx, userID)
	}

	return nil
}

func (m *mockRepository) IncrementJobsPostToday(
	ctx context.Context,
	userID uint,
) error {
	if m.incrementJobsPostTodayFn != nil {
		return m.incrementJobsPostTodayFn(ctx, userID)
	}

	return nil
}

func TestService_GetCleanerAccessStatus_Success(t *testing.T) {
	expected := &CleanerAccessStatus{
		UserID:                7,
		SubscriptionStatus:    "active",
		Premium:               true,
		CanViewJobs:           true,
		CanApply:              true,
		CanViewClientDetails:  true,
		HasInstantAlerts:      true,
		HasPriorityPlacement:  true,
		HasPremiumProfile:     true,
		HasEarlyJobAccess:     true,
		ApplicationCount:      2,
		FreeApplicationLimit:  5,
		ApplicationsToday:     1,
		DailyApplicationLimit: 5,
	}

	repo := &mockRepository{
		getCleanerAccessStatusFn: func(
			ctx context.Context,
			userID uint,
		) (*CleanerAccessStatus, error) {
			if userID != 7 {
				t.Fatalf("expected userID 7, got %d", userID)
			}

			return expected, nil
		},
	}

	service := NewService(repo)

	got, err := service.GetCleanerAccessStatus(
		context.Background(),
		7,
	)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if got != expected {
		t.Fatalf("expected returned status pointer to match")
	}
}

func TestService_GetCleanerAccessStatus_InvalidUserID(t *testing.T) {
	service := NewService(&mockRepository{})

	got, err := service.GetCleanerAccessStatus(
		context.Background(),
		0,
	)

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf(
			"expected ErrInvalidInput, got %v",
			err,
		)
	}

	if got != nil {
		t.Fatalf("expected nil status")
	}
}

func TestService_GetCleanerAccessStatus_RepositoryError(t *testing.T) {
	expectedErr := errors.New("repository failure")

	repo := &mockRepository{
		getCleanerAccessStatusFn: func(
			ctx context.Context,
			userID uint,
		) (*CleanerAccessStatus, error) {
			return nil, expectedErr
		},
	}

	service := NewService(repo)

	got, err := service.GetCleanerAccessStatus(
		context.Background(),
		7,
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected repository error, got %v",
			err,
		)
	}

	if got != nil {
		t.Fatalf("expected nil status")
	}
}

func TestService_GetClientAccessStatus_Success(t *testing.T) {
	expected := &ClientAccessStatus{
		UserID:              11,
		SubscriptionStatus:  "active",
		Premium:             true,
		CanPostJob:          true,
		CanViewApplicants:   true,
		CanContactCleaners:  true,
		CanUseRepeatBooking: true,
		JobsPostedToday:     2,
		DailyJobPostLimit:   5,
		FreeJobPostLimit:    5,
		JobPostCount:        3,
	}

	repo := &mockRepository{
		getClientAccessStatusFn: func(
			ctx context.Context,
			userID uint,
		) (*ClientAccessStatus, error) {
			if userID != 11 {
				t.Fatalf(
					"expected userID 11, got %d",
					userID,
				)
			}

			return expected, nil
		},
	}

	service := NewService(repo)

	got, err := service.GetClientAccessStatus(
		context.Background(),
		11,
	)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if got != expected {
		t.Fatalf("expected returned status pointer to match")
	}
}

func TestService_GetClientAccessStatus_InvalidUserID(t *testing.T) {
	service := NewService(&mockRepository{})

	got, err := service.GetClientAccessStatus(
		context.Background(),
		0,
	)

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf(
			"expected ErrInvalidInput, got %v",
			err,
		)
	}

	if got != nil {
		t.Fatalf("expected nil status")
	}
}

func TestService_GetClientAccessStatus_RepositoryError(t *testing.T) {
	expectedErr := errors.New("repository failure")

	repo := &mockRepository{
		getClientAccessStatusFn: func(
			ctx context.Context,
			userID uint,
		) (*ClientAccessStatus, error) {
			return nil, expectedErr
		},
	}

	service := NewService(repo)

	got, err := service.GetClientAccessStatus(
		context.Background(),
		9,
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected repository error, got %v",
			err,
		)
	}

	if got != nil {
		t.Fatalf("expected nil status")
	}
}

func TestService_EnsureCanApply_Success(t *testing.T) {
	repo := &mockRepository{
		getCleanerAccessStatusFn: func(
			ctx context.Context,
			userID uint,
		) (*CleanerAccessStatus, error) {
			return &CleanerAccessStatus{
				UserID:   userID,
				CanApply: true,
			}, nil
		},
	}

	service := NewService(repo)

	err := service.EnsureCanApply(
		context.Background(),
		7,
	)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestService_EnsureCanApply_DailyLimitReached(t *testing.T) {
	repo := &mockRepository{
		getCleanerAccessStatusFn: func(
			ctx context.Context,
			userID uint,
		) (*CleanerAccessStatus, error) {
			return &CleanerAccessStatus{
				UserID:   userID,
				CanApply: false,
			}, nil
		},
	}

	service := NewService(repo)

	err := service.EnsureCanApply(
		context.Background(),
		7,
	)

	if !errors.Is(err, ErrDailyApplicationLimit) {
		t.Fatalf(
			"expected ErrDailyApplicationLimit, got %v",
			err,
		)
	}
}

func TestService_EnsureCanApply_InvalidUserID(t *testing.T) {
	service := NewService(&mockRepository{})

	err := service.EnsureCanApply(
		context.Background(),
		0,
	)

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf(
			"expected ErrInvalidInput, got %v",
			err,
		)
	}
}

func TestService_EnsureCanApply_RepositoryError(t *testing.T) {
	expectedErr := errors.New("repository failure")

	repo := &mockRepository{
		getCleanerAccessStatusFn: func(
			ctx context.Context,
			userID uint,
		) (*CleanerAccessStatus, error) {
			return nil, expectedErr
		},
	}

	service := NewService(repo)

	err := service.EnsureCanApply(
		context.Background(),
		7,
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected repository error, got %v",
			err,
		)
	}
}

func TestService_EnsureCanPostJob_Success(t *testing.T) {
	repo := &mockRepository{
		getClientAccessStatusFn: func(
			ctx context.Context,
			userID uint,
		) (*ClientAccessStatus, error) {
			return &ClientAccessStatus{
				UserID:     userID,
				CanPostJob: true,
			}, nil
		},
	}

	service := NewService(repo)

	err := service.EnsureCanPostJob(
		context.Background(),
		5,
	)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestService_EnsureCanPostJob_DailyLimitReached(t *testing.T) {
	repo := &mockRepository{
		getClientAccessStatusFn: func(
			ctx context.Context,
			userID uint,
		) (*ClientAccessStatus, error) {
			return &ClientAccessStatus{
				UserID:     userID,
				CanPostJob: false,
			}, nil
		},
	}

	service := NewService(repo)

	err := service.EnsureCanPostJob(
		context.Background(),
		5,
	)

	if !errors.Is(err, ErrDailyJobPostLimit) {
		t.Fatalf(
			"expected ErrDailyJobPostLimit, got %v",
			err,
		)
	}
}

func TestService_EnsureCanPostJob_InvalidUserID(t *testing.T) {
	service := NewService(&mockRepository{})

	err := service.EnsureCanPostJob(
		context.Background(),
		0,
	)

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf(
			"expected ErrInvalidInput, got %v",
			err,
		)
	}
}

func TestService_EnsureCanPostJob_RepositoryError(t *testing.T) {
	expectedErr := errors.New("repository failure")

	repo := &mockRepository{
		getClientAccessStatusFn: func(
			ctx context.Context,
			userID uint,
		) (*ClientAccessStatus, error) {
			return nil, expectedErr
		},
	}

	service := NewService(repo)

	err := service.EnsureCanPostJob(
		context.Background(),
		5,
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected repository error, got %v",
			err,
		)
	}
}

func TestService_IncrementApplicationsToday_Success(t *testing.T) {
	called := false

	repo := &mockRepository{
		incrementApplicationsTodayFn: func(
			ctx context.Context,
			userID uint,
		) error {
			called = true

			if userID != 7 {
				t.Fatalf(
					"expected userID 7, got %d",
					userID,
				)
			}

			return nil
		},
	}

	service := NewService(repo)

	err := service.IncrementApplicationsToday(
		context.Background(),
		7,
	)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if !called {
		t.Fatalf("expected repository to be called")
	}
}

func TestService_IncrementApplicationsToday_InvalidUserID(t *testing.T) {
	service := NewService(&mockRepository{})

	err := service.IncrementApplicationsToday(
		context.Background(),
		0,
	)

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf(
			"expected ErrInvalidInput, got %v",
			err,
		)
	}
}

func TestService_IncrementApplicationsToday_RepositoryError(t *testing.T) {
	expectedErr := errors.New("repository failure")

	repo := &mockRepository{
		incrementApplicationsTodayFn: func(
			ctx context.Context,
			userID uint,
		) error {
			return expectedErr
		},
	}

	service := NewService(repo)

	err := service.IncrementApplicationsToday(
		context.Background(),
		7,
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected repository error, got %v",
			err,
		)
	}
}

func TestService_IncrementJobsPostToday_Success(t *testing.T) {
	called := false

	repo := &mockRepository{
		incrementJobsPostTodayFn: func(
			ctx context.Context,
			userID uint,
		) error {
			called = true

			if userID != 9 {
				t.Fatalf(
					"expected userID 9, got %d",
					userID,
				)
			}

			return nil
		},
	}

	service := NewService(repo)

	err := service.IncrementJobsPostToday(
		context.Background(),
		9,
	)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if !called {
		t.Fatalf("expected repository to be called")
	}
}

func TestService_IncrementJobsPostToday_InvalidUserID(t *testing.T) {
	service := NewService(&mockRepository{})

	err := service.IncrementJobsPostToday(
		context.Background(),
		0,
	)

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf(
			"expected ErrInvalidInput, got %v",
			err,
		)
	}
}

func TestService_IncrementJobsPostToday_RepositoryError(t *testing.T) {
	expectedErr := errors.New("repository failure")

	repo := &mockRepository{
		incrementJobsPostTodayFn: func(
			ctx context.Context,
			userID uint,
		) error {
			return expectedErr
		},
	}

	service := NewService(repo)

	err := service.IncrementJobsPostToday(
		context.Background(),
		9,
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected repository error, got %v",
			err,
		)
	}
}
