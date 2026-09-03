package clientdashboard

import (
	"context"
	"errors"
	"testing"
	"time"

	subscriptionaccessdomain "github.com/rodrigueghenda/jobira/internal/domain/subscriptionaccess"
)

type mockRepository struct {
	listActiveJobsFn            func(context.Context, uint, int) ([]ActiveJob, error)
	listUpcomingBookingsFn      func(context.Context, uint, int) ([]ClientBooking, error)
	countApplicationsReceivedFn func(context.Context, uint) (int, error)
	countCompletedBookingsFn    func(context.Context, uint) (int, error)
	countFavouriteCleanersFn    func(context.Context, uint) (int, error)
	countPreferredCleanersFn    func(context.Context, uint) (int, error)
	countRepeatBookingsFn       func(context.Context, uint) (int, error)
}

func (m *mockRepository) ListActiveJobs(
	ctx context.Context,
	clientID uint,
	limit int,
) ([]ActiveJob, error) {
	if m.listActiveJobsFn != nil {
		return m.listActiveJobsFn(ctx, clientID, limit)
	}

	return nil, nil
}

func (m *mockRepository) ListUpcomingBookings(
	ctx context.Context,
	clientID uint,
	limit int,
) ([]ClientBooking, error) {
	if m.listUpcomingBookingsFn != nil {
		return m.listUpcomingBookingsFn(ctx, clientID, limit)
	}

	return nil, nil
}

func (m *mockRepository) CountApplicationsReceived(
	ctx context.Context,
	clientID uint,
) (int, error) {
	if m.countApplicationsReceivedFn != nil {
		return m.countApplicationsReceivedFn(ctx, clientID)
	}

	return 0, nil
}

func (m *mockRepository) CountCompletedBookings(
	ctx context.Context,
	clientID uint,
) (int, error) {
	if m.countCompletedBookingsFn != nil {
		return m.countCompletedBookingsFn(ctx, clientID)
	}

	return 0, nil
}

func (m *mockRepository) CountFavouriteCleaners(
	ctx context.Context,
	clientID uint,
) (int, error) {
	if m.countFavouriteCleanersFn != nil {
		return m.countFavouriteCleanersFn(ctx, clientID)
	}

	return 0, nil
}

func (m *mockRepository) CountPreferredCleaners(
	ctx context.Context,
	clientID uint,
) (int, error) {
	if m.countPreferredCleanersFn != nil {
		return m.countPreferredCleanersFn(ctx, clientID)
	}

	return 0, nil
}

func (m *mockRepository) CountRepeatBookings(
	ctx context.Context,
	clientID uint,
) (int, error) {
	if m.countRepeatBookingsFn != nil {
		return m.countRepeatBookingsFn(ctx, clientID)
	}

	return 0, nil
}

type mockSubscriptionAccessRepository struct {
	getCleanerAccessStatusFn     func(context.Context, uint) (*subscriptionaccessdomain.CleanerAccessStatus, error)
	getClientAccessStatusFn      func(context.Context, uint) (*subscriptionaccessdomain.ClientAccessStatus, error)
	ensureDailyAccessFn          func(context.Context, uint) error
	isLaunchGraceActiveFn        func(context.Context) (bool, error)
	incrementApplicationsTodayFn func(context.Context, uint) error
	incrementJobsPostTodayFn     func(context.Context, uint) error
}

func (m *mockSubscriptionAccessRepository) GetCleanerAccessStatus(
	ctx context.Context,
	userID uint,
) (*subscriptionaccessdomain.CleanerAccessStatus, error) {
	if m.getCleanerAccessStatusFn != nil {
		return m.getCleanerAccessStatusFn(ctx, userID)
	}

	return &subscriptionaccessdomain.CleanerAccessStatus{
		UserID: userID,
	}, nil
}

func (m *mockSubscriptionAccessRepository) GetClientAccessStatus(
	ctx context.Context,
	userID uint,
) (*subscriptionaccessdomain.ClientAccessStatus, error) {
	if m.getClientAccessStatusFn != nil {
		return m.getClientAccessStatusFn(ctx, userID)
	}

	return &subscriptionaccessdomain.ClientAccessStatus{
		UserID: userID,
	}, nil
}

func (m *mockSubscriptionAccessRepository) EnsureDailyAccess(
	ctx context.Context,
	userID uint,
) error {
	if m.ensureDailyAccessFn != nil {
		return m.ensureDailyAccessFn(ctx, userID)
	}

	return nil
}

func (m *mockSubscriptionAccessRepository) IsLaunchGraceActive(
	ctx context.Context,
) (bool, error) {
	if m.isLaunchGraceActiveFn != nil {
		return m.isLaunchGraceActiveFn(ctx)
	}

	return false, nil
}

func (m *mockSubscriptionAccessRepository) IncrementApplicationsToday(
	ctx context.Context,
	userID uint,
) error {
	if m.incrementApplicationsTodayFn != nil {
		return m.incrementApplicationsTodayFn(ctx, userID)
	}

	return nil
}

func (m *mockSubscriptionAccessRepository) IncrementJobsPostToday(
	ctx context.Context,
	userID uint,
) error {
	if m.incrementJobsPostTodayFn != nil {
		return m.incrementJobsPostTodayFn(ctx, userID)
	}

	return nil
}

func newClientDashboardService(
	repo Repository,
	accessRepo subscriptionaccessdomain.Repository,
) *Service {
	accessService := subscriptionaccessdomain.NewService(accessRepo)

	return NewService(repo, accessService)
}

func TestService_GetMine_Success(t *testing.T) {
	ctx := context.Background()
	clientID := uint(7)

	now := time.Now().UTC()

	activeJobs := []ActiveJob{
		{
			ID:               11,
			Title:            "End of tenancy clean",
			Location:         "East London",
			JobType:          "end_of_tenancy",
			Budget:           150,
			Status:           "open",
			ApplicationCount: 6,
			CreatedAt:        now,
		},
		{
			ID:               12,
			Title:            "Airbnb turnover",
			Location:         "London",
			JobType:          "airbnb",
			Budget:           80,
			Status:           "open",
			ApplicationCount: 3,
			CreatedAt:        now,
		},
	}

	upcomingBookings := []ClientBooking{
		{
			ID:          21,
			JobID:       11,
			CleanerID:   20,
			Status:      "confirmed",
			ScheduledAt: now.Add(24 * time.Hour),
		},
		{
			ID:          22,
			JobID:       12,
			CleanerID:   21,
			Status:      "pending",
			ScheduledAt: now.Add(48 * time.Hour),
		},
	}

	repo := &mockRepository{
		listActiveJobsFn: func(
			ctx context.Context,
			gotClientID uint,
			limit int,
		) ([]ActiveJob, error) {
			if gotClientID != clientID {
				t.Fatalf(
					"expected clientID %d, got %d",
					clientID,
					gotClientID,
				)
			}

			if limit != 5 {
				t.Fatalf("expected limit 5, got %d", limit)
			}

			return activeJobs, nil
		},

		listUpcomingBookingsFn: func(
			ctx context.Context,
			gotClientID uint,
			limit int,
		) ([]ClientBooking, error) {
			if gotClientID != clientID {
				t.Fatalf(
					"expected clientID %d, got %d",
					clientID,
					gotClientID,
				)
			}

			if limit != 5 {
				t.Fatalf("expected limit 5, got %d", limit)
			}

			return upcomingBookings, nil
		},

		countApplicationsReceivedFn: func(
			ctx context.Context,
			gotClientID uint,
		) (int, error) {
			return 19, nil
		},

		countCompletedBookingsFn: func(
			ctx context.Context,
			gotClientID uint,
		) (int, error) {
			return 14, nil
		},

		countFavouriteCleanersFn: func(
			ctx context.Context,
			gotClientID uint,
		) (int, error) {
			return 8, nil
		},

		countPreferredCleanersFn: func(
			ctx context.Context,
			gotClientID uint,
		) (int, error) {
			return 4, nil
		},

		countRepeatBookingsFn: func(
			ctx context.Context,
			gotClientID uint,
		) (int, error) {
			return 6, nil
		},
	}

	accessRepo := &mockSubscriptionAccessRepository{
		getClientAccessStatusFn: func(
			ctx context.Context,
			userID uint,
		) (*subscriptionaccessdomain.ClientAccessStatus, error) {
			if userID != clientID {
				t.Fatalf(
					"expected userID %d, got %d",
					clientID,
					userID,
				)
			}

			return &subscriptionaccessdomain.ClientAccessStatus{
				UserID:             clientID,
				SubscriptionStatus: "active",
				LaunchGraceActive:  false,
				Premium:            true,
				CanPostJob:         true,
				JobPostCount:       8,
				FreeJobPostLimit:   5,
				JobsPostedToday:    2,
				DailyJobPostLimit:  5,
				UpgradeMessage:     "",
			}, nil
		},
	}

	service := newClientDashboardService(repo, accessRepo)

	got, err := service.GetMine(ctx, clientID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got == nil {
		t.Fatal("expected dashboard, got nil")
	}

	if got.ClientID != clientID {
		t.Fatalf(
			"expected clientID %d, got %d",
			clientID,
			got.ClientID,
		)
	}

	if got.SubscriptionStatus != "active" {
		t.Fatalf(
			"expected active subscription, got %s",
			got.SubscriptionStatus,
		)
	}

	if got.LaunchGraceActive {
		t.Fatal("expected launch grace to be false")
	}

	if !got.Premium {
		t.Fatal("expected premium to be true")
	}

	if !got.CanPostJob {
		t.Fatal("expected CanPostJob to be true")
	}

	if got.JobPostCount != 8 {
		t.Fatalf(
			"expected JobPostCount 8, got %d",
			got.JobPostCount,
		)
	}

	if got.FreeJobPostLimit != 5 {
		t.Fatalf(
			"expected FreeJobPostLimit 5, got %d",
			got.FreeJobPostLimit,
		)
	}

	if got.JobsPostedToday != 2 {
		t.Fatalf(
			"expected JobsPostedToday 2, got %d",
			got.JobsPostedToday,
		)
	}

	if got.DailyJobPostLimit != 5 {
		t.Fatalf(
			"expected DailyJobPostLimit 5, got %d",
			got.DailyJobPostLimit,
		)
	}

	if got.ActiveJobCount != 2 {
		t.Fatalf(
			"expected ActiveJobCount 2, got %d",
			got.ActiveJobCount,
		)
	}

	if got.ApplicationsReceived != 19 {
		t.Fatalf(
			"expected ApplicationsReceived 19, got %d",
			got.ApplicationsReceived,
		)
	}

	if got.ActiveBookingCount != 2 {
		t.Fatalf(
			"expected ActiveBookingCount 2, got %d",
			got.ActiveBookingCount,
		)
	}

	if got.CompletedBookingCount != 14 {
		t.Fatalf(
			"expected CompletedBookingCount 14, got %d",
			got.CompletedBookingCount,
		)
	}

	if got.FavouriteCleanerCount != 8 {
		t.Fatalf(
			"expected FavouriteCleanerCount 8, got %d",
			got.FavouriteCleanerCount,
		)
	}

	if got.PreferredCleanerCount != 4 {
		t.Fatalf(
			"expected PreferredCleanerCount 4, got %d",
			got.PreferredCleanerCount,
		)
	}

	if got.RepeatBookingCount != 6 {
		t.Fatalf(
			"expected RepeatBookingCount 6, got %d",
			got.RepeatBookingCount,
		)
	}

	if len(got.ActiveJobs) != 2 {
		t.Fatalf(
			"expected 2 active jobs, got %d",
			len(got.ActiveJobs),
		)
	}

	if len(got.UpcomingBookings) != 2 {
		t.Fatalf(
			"expected 2 upcoming bookings, got %d",
			len(got.UpcomingBookings),
		)
	}

	if got.UpgradeMessage != "" {
		t.Fatalf(
			"expected empty upgrade message, got %q",
			got.UpgradeMessage,
		)
	}
}

func TestService_GetMine_InvalidClientID(t *testing.T) {
	service := newClientDashboardService(
		&mockRepository{},
		&mockSubscriptionAccessRepository{},
	)

	got, err := service.GetMine(context.Background(), 0)

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf(
			"expected ErrInvalidInput, got %v",
			err,
		)
	}

	if got != nil {
		t.Fatal("expected nil dashboard")
	}
}

func TestService_GetMine_SubscriptionAccessError(t *testing.T) {
	expectedErr := errors.New("subscription access failure")

	accessRepo := &mockSubscriptionAccessRepository{
		getClientAccessStatusFn: func(
			ctx context.Context,
			userID uint,
		) (*subscriptionaccessdomain.ClientAccessStatus, error) {
			return nil, expectedErr
		},
	}

	service := newClientDashboardService(
		&mockRepository{},
		accessRepo,
	)

	got, err := service.GetMine(
		context.Background(),
		7,
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}

	if got != nil {
		t.Fatal("expected nil dashboard")
	}
}

func TestService_GetMine_ListActiveJobsError(t *testing.T) {
	expectedErr := errors.New("list active jobs failure")

	repo := &mockRepository{
		listActiveJobsFn: func(
			ctx context.Context,
			clientID uint,
			limit int,
		) ([]ActiveJob, error) {
			return nil, expectedErr
		},
	}

	service := newClientDashboardService(
		repo,
		&mockSubscriptionAccessRepository{},
	)

	got, err := service.GetMine(
		context.Background(),
		7,
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}

	if got != nil {
		t.Fatal("expected nil dashboard")
	}
}

func TestService_GetMine_ListUpcomingBookingsError(t *testing.T) {
	expectedErr := errors.New("list upcoming bookings failure")

	repo := &mockRepository{
		listUpcomingBookingsFn: func(
			ctx context.Context,
			clientID uint,
			limit int,
		) ([]ClientBooking, error) {
			return nil, expectedErr
		},
	}

	service := newClientDashboardService(
		repo,
		&mockSubscriptionAccessRepository{},
	)

	got, err := service.GetMine(
		context.Background(),
		7,
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}

	if got != nil {
		t.Fatal("expected nil dashboard")
	}
}

func TestService_GetMine_CountApplicationsReceivedError(t *testing.T) {
	expectedErr := errors.New("count applications failure")

	repo := &mockRepository{
		countApplicationsReceivedFn: func(
			ctx context.Context,
			clientID uint,
		) (int, error) {
			return 0, expectedErr
		},
	}

	service := newClientDashboardService(
		repo,
		&mockSubscriptionAccessRepository{},
	)

	got, err := service.GetMine(
		context.Background(),
		7,
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}

	if got != nil {
		t.Fatal("expected nil dashboard")
	}
}

func TestService_GetMine_CountCompletedBookingsError(t *testing.T) {
	expectedErr := errors.New("count completed bookings failure")

	repo := &mockRepository{
		countCompletedBookingsFn: func(
			ctx context.Context,
			clientID uint,
		) (int, error) {
			return 0, expectedErr
		},
	}

	service := newClientDashboardService(
		repo,
		&mockSubscriptionAccessRepository{},
	)

	got, err := service.GetMine(
		context.Background(),
		7,
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}

	if got != nil {
		t.Fatal("expected nil dashboard")
	}
}

func TestService_GetMine_CountFavouriteCleanersError(t *testing.T) {
	expectedErr := errors.New("count favourites failure")

	repo := &mockRepository{
		countFavouriteCleanersFn: func(
			ctx context.Context,
			clientID uint,
		) (int, error) {
			return 0, expectedErr
		},
	}

	service := newClientDashboardService(
		repo,
		&mockSubscriptionAccessRepository{},
	)

	got, err := service.GetMine(
		context.Background(),
		7,
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}

	if got != nil {
		t.Fatal("expected nil dashboard")
	}
}

func TestService_GetMine_CountPreferredCleanersError(t *testing.T) {
	expectedErr := errors.New("count preferred cleaners failure")

	repo := &mockRepository{
		countPreferredCleanersFn: func(
			ctx context.Context,
			clientID uint,
		) (int, error) {
			return 0, expectedErr
		},
	}

	service := newClientDashboardService(
		repo,
		&mockSubscriptionAccessRepository{},
	)

	got, err := service.GetMine(
		context.Background(),
		7,
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}

	if got != nil {
		t.Fatal("expected nil dashboard")
	}
}

func TestService_GetMine_CountRepeatBookingsError(t *testing.T) {
	expectedErr := errors.New("count repeat bookings failure")

	repo := &mockRepository{
		countRepeatBookingsFn: func(
			ctx context.Context,
			clientID uint,
		) (int, error) {
			return 0, expectedErr
		},
	}

	service := newClientDashboardService(
		repo,
		&mockSubscriptionAccessRepository{},
	)

	got, err := service.GetMine(
		context.Background(),
		7,
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}

	if got != nil {
		t.Fatal("expected nil dashboard")
	}
}
