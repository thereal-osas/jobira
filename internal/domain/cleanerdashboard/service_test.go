package cleanerdashboard

import (
	"context"
	"errors"
	"testing"
	"time"

	reputationdomain "github.com/rodrigueghenda/jobira/internal/domain/reputation"
	subscriptionaccessdomain "github.com/rodrigueghenda/jobira/internal/domain/subscriptionaccess"
)

type mockRepository struct {
	countFavouritesFn       func(ctx context.Context, cleanerID uint) (int, error)
	countPreferredClientsFn func(ctx context.Context, cleanerID uint) (int, error)
	listUpcomingBookingsFn  func(ctx context.Context, cleanerID uint, limit int) ([]UpcomingBooking, error)
}

func (m *mockRepository) CountFavourites(
	ctx context.Context,
	cleanerID uint,
) (int, error) {
	if m.countFavouritesFn != nil {
		return m.countFavouritesFn(ctx, cleanerID)
	}
	return 0, nil
}

func (m *mockRepository) CountPreferredClients(
	ctx context.Context,
	cleanerID uint,
) (int, error) {
	if m.countPreferredClientsFn != nil {
		return m.countPreferredClientsFn(ctx, cleanerID)
	}
	return 0, nil
}

func (m *mockRepository) ListUpcomingBookings(
	ctx context.Context,
	cleanerID uint,
	limit int,
) ([]UpcomingBooking, error) {
	if m.listUpcomingBookingsFn != nil {
		return m.listUpcomingBookingsFn(ctx, cleanerID, limit)
	}
	return nil, nil
}

type mockSubscriptionAccessRepository struct {
	getCleanerAccessStatusFn func(
		ctx context.Context,
		userID uint,
	) (*subscriptionaccessdomain.CleanerAccessStatus, error)
}

func (m *mockSubscriptionAccessRepository) GetCleanerAccessStatus(
	ctx context.Context,
	userID uint,
) (*subscriptionaccessdomain.CleanerAccessStatus, error) {
	if m.getCleanerAccessStatusFn != nil {
		return m.getCleanerAccessStatusFn(ctx, userID)
	}
	return nil, nil
}

func (m *mockSubscriptionAccessRepository) GetClientAccessStatus(
	ctx context.Context,
	userID uint,
) (*subscriptionaccessdomain.ClientAccessStatus, error) {
	return nil, nil
}

func (m *mockSubscriptionAccessRepository) EnsureDailyAccess(
	ctx context.Context,
	userID uint,
) error {
	return nil
}

func (m *mockSubscriptionAccessRepository) IsLaunchGraceActive(
	ctx context.Context,
) (bool, error) {
	return false, nil
}

func (m *mockSubscriptionAccessRepository) IncrementApplicationsToday(
	ctx context.Context,
	userID uint,
) error {
	return nil
}

func (m *mockSubscriptionAccessRepository) IncrementJobsPostToday(
	ctx context.Context,
	userID uint,
) error {
	return nil
}

type mockReputationRepository struct {
	getByCleanerIDFn func(
		ctx context.Context,
		cleanerID uint,
	) (*reputationdomain.CleanerReputation, error)
}

func (m *mockReputationRepository) EnsureCleaner(
	ctx context.Context,
	cleanerID uint,
) error {
	return nil
}

func (m *mockReputationRepository) Refresh(
	ctx context.Context,
	cleanerID uint,
) error {
	return nil
}

func (m *mockReputationRepository) GetByCleanerID(
	ctx context.Context,
	cleanerID uint,
) (*reputationdomain.CleanerReputation, error) {
	if m.getByCleanerIDFn != nil {
		return m.getByCleanerIDFn(ctx, cleanerID)
	}
	return nil, nil
}

func (m *mockReputationRepository) UpdateBadge(
	ctx context.Context,
	cleanerID uint,
	badge string,
) error {
	return nil
}

func newDashboardService(
	repo Repository,
	accessRepo subscriptionaccessdomain.Repository,
	reputationRepo reputationdomain.Repository,
) *Service {
	accessService := subscriptionaccessdomain.NewService(accessRepo)
	reputationService := reputationdomain.NewService(reputationRepo)

	return NewService(
		repo,
		reputationService,
		accessService,
	)
}

func TestService_GetMine_Success(t *testing.T) {
	now := time.Now().Add(24 * time.Hour)

	repo := &mockRepository{
		countFavouritesFn: func(
			ctx context.Context,
			cleanerID uint,
		) (int, error) {
			if cleanerID != 7 {
				t.Fatalf("expected cleanerID 7, got %d", cleanerID)
			}
			return 8, nil
		},
		countPreferredClientsFn: func(
			ctx context.Context,
			cleanerID uint,
		) (int, error) {
			return 3, nil
		},
		listUpcomingBookingsFn: func(
			ctx context.Context,
			cleanerID uint,
			limit int,
		) ([]UpcomingBooking, error) {
			if limit != 5 {
				t.Fatalf("expected limit 5, got %d", limit)
			}

			return []UpcomingBooking{
				{
					ID:          1,
					JobID:       20,
					ClientID:    30,
					Status:      "confirmed",
					ScheduledAt: now,
				},
				{
					ID:          2,
					JobID:       21,
					ClientID:    31,
					Status:      "pending",
					ScheduledAt: now.Add(time.Hour),
				},
			}, nil
		},
	}

	accessRepo := &mockSubscriptionAccessRepository{
		getCleanerAccessStatusFn: func(
			ctx context.Context,
			userID uint,
		) (*subscriptionaccessdomain.CleanerAccessStatus, error) {
			return &subscriptionaccessdomain.CleanerAccessStatus{
				UserID:                userID,
				SubscriptionStatus:    "active",
				TrialActive:           false,
				LaunchGraceActive:     false,
				Premium:               true,
				CanApply:              true,
				ApplicationsToday:     2,
				DailyApplicationLimit: 5,
				UpgradeMessage:        "",
			}, nil
		},
	}

	reputationRepo := &mockReputationRepository{
		getByCleanerIDFn: func(
			ctx context.Context,
			cleanerID uint,
		) (*reputationdomain.CleanerReputation, error) {
			return &reputationdomain.CleanerReputation{
				CleanerID:                cleanerID,
				AverageRating:            4.8,
				TotalReviews:             21,
				CompletedJobs:            40,
				RepeatClients:            12,
				RecommendationPercentage: 95,
				Badge:                    "trusted",
			}, nil
		},
	}

	service := newDashboardService(
		repo,
		accessRepo,
		reputationRepo,
	)

	got, err := service.GetMine(
		context.Background(),
		7,
	)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if got.CleanerID != 7 {
		t.Fatalf("expected cleanerID 7, got %d", got.CleanerID)
	}

	if got.SubscriptionStatus != "active" {
		t.Fatalf(
			"expected active subscription, got %s",
			got.SubscriptionStatus,
		)
	}

	if !got.Premium {
		t.Fatal("expected premium true")
	}

	if !got.CanApply {
		t.Fatal("expected CanApply true")
	}

	if got.ApplicationsToday != 2 {
		t.Fatalf(
			"expected 2 applications today, got %d",
			got.ApplicationsToday,
		)
	}

	if got.AverageRating != 4.8 {
		t.Fatalf(
			"expected rating 4.8, got %f",
			got.AverageRating,
		)
	}

	if got.TotalReviews != 21 {
		t.Fatalf(
			"expected 21 reviews, got %d",
			got.TotalReviews,
		)
	}

	if got.CompletedJobs != 40 {
		t.Fatalf(
			"expected 40 completed jobs, got %d",
			got.CompletedJobs,
		)
	}

	if got.RepeatClients != 12 {
		t.Fatalf(
			"expected 12 repeat clients, got %d",
			got.RepeatClients,
		)
	}

	if got.RecommendationPercentage != 95 {
		t.Fatalf(
			"expected recommendation 95, got %d",
			got.RecommendationPercentage,
		)
	}

	if got.Badge != "Top Rated" {
		t.Fatalf(
			"expected Top Rated badge, got %s",
			got.Badge,
		)
	}

	if got.FavouriteCount != 8 {
		t.Fatalf(
			"expected 8 favourites, got %d",
			got.FavouriteCount,
		)
	}

	if got.PreferredClientCount != 3 {
		t.Fatalf(
			"expected 3 preferred clients, got %d",
			got.PreferredClientCount,
		)
	}

	if got.UpcomingBookingCount != 2 {
		t.Fatalf(
			"expected 2 upcoming bookings, got %d",
			got.UpcomingBookingCount,
		)
	}

	if len(got.UpcomingBookings) != 2 {
		t.Fatalf(
			"expected 2 bookings, got %d",
			len(got.UpcomingBookings),
		)
	}
}

func TestService_GetMine_InvalidCleanerID(t *testing.T) {
	service := newDashboardService(
		&mockRepository{},
		&mockSubscriptionAccessRepository{},
		&mockReputationRepository{},
	)

	got, err := service.GetMine(
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
		t.Fatal("expected nil dashboard")
	}
}

func TestService_GetMine_SubscriptionAccessError(t *testing.T) {
	expectedErr := errors.New("subscription access failure")

	accessRepo := &mockSubscriptionAccessRepository{
		getCleanerAccessStatusFn: func(
			ctx context.Context,
			userID uint,
		) (*subscriptionaccessdomain.CleanerAccessStatus, error) {
			return nil, expectedErr
		},
	}

	service := newDashboardService(
		&mockRepository{},
		accessRepo,
		&mockReputationRepository{},
	)

	got, err := service.GetMine(
		context.Background(),
		7,
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected access error, got %v",
			err,
		)
	}

	if got != nil {
		t.Fatal("expected nil dashboard")
	}
}

func TestService_GetMine_ReputationError(t *testing.T) {
	expectedErr := errors.New("reputation failure")

	accessRepo := &mockSubscriptionAccessRepository{
		getCleanerAccessStatusFn: func(
			ctx context.Context,
			userID uint,
		) (*subscriptionaccessdomain.CleanerAccessStatus, error) {
			return &subscriptionaccessdomain.CleanerAccessStatus{
				UserID: userID,
			}, nil
		},
	}

	reputationRepo := &mockReputationRepository{
		getByCleanerIDFn: func(
			ctx context.Context,
			cleanerID uint,
		) (*reputationdomain.CleanerReputation, error) {
			return nil, expectedErr
		},
	}

	service := newDashboardService(
		&mockRepository{},
		accessRepo,
		reputationRepo,
	)

	got, err := service.GetMine(
		context.Background(),
		7,
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected reputation error, got %v",
			err,
		)
	}

	if got != nil {
		t.Fatal("expected nil dashboard")
	}
}

func TestService_GetMine_CountFavouritesError(t *testing.T) {
	expectedErr := errors.New("favourite count failure")

	repo := &mockRepository{
		countFavouritesFn: func(
			ctx context.Context,
			cleanerID uint,
		) (int, error) {
			return 0, expectedErr
		},
	}

	accessRepo := &mockSubscriptionAccessRepository{
		getCleanerAccessStatusFn: func(
			ctx context.Context,
			userID uint,
		) (*subscriptionaccessdomain.CleanerAccessStatus, error) {
			return &subscriptionaccessdomain.CleanerAccessStatus{
				UserID: userID,
			}, nil
		},
	}

	reputationRepo := &mockReputationRepository{
		getByCleanerIDFn: func(
			ctx context.Context,
			cleanerID uint,
		) (*reputationdomain.CleanerReputation, error) {
			return &reputationdomain.CleanerReputation{
				CleanerID: cleanerID,
			}, nil
		},
	}

	service := newDashboardService(
		repo,
		accessRepo,
		reputationRepo,
	)

	got, err := service.GetMine(
		context.Background(),
		7,
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected favourite error, got %v",
			err,
		)
	}

	if got != nil {
		t.Fatal("expected nil dashboard")
	}
}

func TestService_GetMine_CountPreferredClientsError(t *testing.T) {
	expectedErr := errors.New("preferred clients failure")

	repo := &mockRepository{
		countFavouritesFn: func(
			ctx context.Context,
			cleanerID uint,
		) (int, error) {
			return 2, nil
		},
		countPreferredClientsFn: func(
			ctx context.Context,
			cleanerID uint,
		) (int, error) {
			return 0, expectedErr
		},
	}

	accessRepo := &mockSubscriptionAccessRepository{
		getCleanerAccessStatusFn: func(
			ctx context.Context,
			userID uint,
		) (*subscriptionaccessdomain.CleanerAccessStatus, error) {
			return &subscriptionaccessdomain.CleanerAccessStatus{
				UserID: userID,
			}, nil
		},
	}

	reputationRepo := &mockReputationRepository{
		getByCleanerIDFn: func(
			ctx context.Context,
			cleanerID uint,
		) (*reputationdomain.CleanerReputation, error) {
			return &reputationdomain.CleanerReputation{
				CleanerID: cleanerID,
			}, nil
		},
	}

	service := newDashboardService(
		repo,
		accessRepo,
		reputationRepo,
	)

	got, err := service.GetMine(
		context.Background(),
		7,
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected preferred clients error, got %v",
			err,
		)
	}

	if got != nil {
		t.Fatal("expected nil dashboard")
	}
}

func TestService_GetMine_ListUpcomingBookingsError(t *testing.T) {
	expectedErr := errors.New("booking query failure")

	repo := &mockRepository{
		countFavouritesFn: func(
			ctx context.Context,
			cleanerID uint,
		) (int, error) {
			return 2, nil
		},
		countPreferredClientsFn: func(
			ctx context.Context,
			cleanerID uint,
		) (int, error) {
			return 3, nil
		},
		listUpcomingBookingsFn: func(
			ctx context.Context,
			cleanerID uint,
			limit int,
		) ([]UpcomingBooking, error) {
			return nil, expectedErr
		},
	}

	accessRepo := &mockSubscriptionAccessRepository{
		getCleanerAccessStatusFn: func(
			ctx context.Context,
			userID uint,
		) (*subscriptionaccessdomain.CleanerAccessStatus, error) {
			return &subscriptionaccessdomain.CleanerAccessStatus{
				UserID: userID,
			}, nil
		},
	}

	reputationRepo := &mockReputationRepository{
		getByCleanerIDFn: func(
			ctx context.Context,
			cleanerID uint,
		) (*reputationdomain.CleanerReputation, error) {
			return &reputationdomain.CleanerReputation{
				CleanerID: cleanerID,
			}, nil
		},
	}

	service := newDashboardService(
		repo,
		accessRepo,
		reputationRepo,
	)

	got, err := service.GetMine(
		context.Background(),
		7,
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected booking error, got %v",
			err,
		)
	}

	if got != nil {
		t.Fatal("expected nil dashboard")
	}
}
