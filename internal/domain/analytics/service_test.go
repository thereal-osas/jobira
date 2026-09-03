package analytics

import (
	"context"
	"errors"
	"testing"
)

type mockRepository struct {
	totalUsersFn           func(context.Context) (int, error)
	usersByRoleFn          func(context.Context, string) (int, error)
	totalCompaniesFn       func(context.Context) (int, error)
	activeJobsFn           func(context.Context) (int, error)
	completedBookingsFn    func(context.Context) (int, error)
	pendingVerificationsFn func(context.Context) (int, error)
	openReportsFn          func(context.Context) (int, error)
	bookingsTodayFn        func(context.Context) (int, error)
}

func (m *mockRepository) TotalUsers(
	ctx context.Context,
) (int, error) {
	if m.totalUsersFn != nil {
		return m.totalUsersFn(ctx)
	}
	return 0, nil
}

func (m *mockRepository) UsersByRole(
	ctx context.Context,
	role string,
) (int, error) {
	if m.usersByRoleFn != nil {
		return m.usersByRoleFn(ctx, role)
	}
	return 0, nil
}

func (m *mockRepository) TotalCompanies(
	ctx context.Context,
) (int, error) {
	if m.totalCompaniesFn != nil {
		return m.totalCompaniesFn(ctx)
	}
	return 0, nil
}

func (m *mockRepository) ActiveJobs(
	ctx context.Context,
) (int, error) {
	if m.activeJobsFn != nil {
		return m.activeJobsFn(ctx)
	}
	return 0, nil
}

func (m *mockRepository) CompletedBookings(
	ctx context.Context,
) (int, error) {
	if m.completedBookingsFn != nil {
		return m.completedBookingsFn(ctx)
	}
	return 0, nil
}

func (m *mockRepository) PendingVerifications(
	ctx context.Context,
) (int, error) {
	if m.pendingVerificationsFn != nil {
		return m.pendingVerificationsFn(ctx)
	}
	return 0, nil
}

func (m *mockRepository) OpenReports(
	ctx context.Context,
) (int, error) {
	if m.openReportsFn != nil {
		return m.openReportsFn(ctx)
	}
	return 0, nil
}

func (m *mockRepository) BookingsToday(
	ctx context.Context,
) (int, error) {
	if m.bookingsTodayFn != nil {
		return m.bookingsTodayFn(ctx)
	}
	return 0, nil
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

func TestService_Dashboard_Success(t *testing.T) {
	var roles []string

	repo := &mockRepository{
		totalUsersFn: func(context.Context) (int, error) {
			return 100, nil
		},
		usersByRoleFn: func(
			ctx context.Context,
			role string,
		) (int, error) {
			roles = append(roles, role)

			switch role {
			case "client":
				return 40, nil
			case "cleaner":
				return 50, nil
			default:
				return 0, nil
			}
		},
		totalCompaniesFn: func(context.Context) (int, error) {
			return 10, nil
		},
		activeJobsFn: func(context.Context) (int, error) {
			return 25, nil
		},
		completedBookingsFn: func(context.Context) (int, error) {
			return 80, nil
		},
		pendingVerificationsFn: func(context.Context) (int, error) {
			return 7, nil
		},
		openReportsFn: func(context.Context) (int, error) {
			return 3, nil
		},
		bookingsTodayFn: func(context.Context) (int, error) {
			return 12, nil
		},
	}

	service := NewService(repo)

	stats, err := service.Dashboard(
		context.Background(),
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if stats == nil {
		t.Fatal("expected dashboard stats")
	}

	if stats.TotalUsers != 100 {
		t.Fatalf(
			"expected 100 total users, got %d",
			stats.TotalUsers,
		)
	}

	if stats.Clients != 40 {
		t.Fatalf(
			"expected 40 clients, got %d",
			stats.Clients,
		)
	}

	if stats.Cleaners != 50 {
		t.Fatalf(
			"expected 50 cleaners, got %d",
			stats.Cleaners,
		)
	}

	if stats.Companies != 10 {
		t.Fatalf(
			"expected 10 companies, got %d",
			stats.Companies,
		)
	}

	if stats.ActiveJobs != 25 {
		t.Fatalf(
			"expected 25 active jobs, got %d",
			stats.ActiveJobs,
		)
	}

	if stats.CompletedBookings != 80 {
		t.Fatalf(
			"expected 80 completed bookings, got %d",
			stats.CompletedBookings,
		)
	}

	if stats.PendingVerifications != 7 {
		t.Fatalf(
			"expected 7 pending verifications, got %d",
			stats.PendingVerifications,
		)
	}

	if stats.OpenReports != 3 {
		t.Fatalf(
			"expected 3 open reports, got %d",
			stats.OpenReports,
		)
	}

	if stats.BookingsToday != 12 {
		t.Fatalf(
			"expected 12 bookings today, got %d",
			stats.BookingsToday,
		)
	}

	if len(roles) != 2 {
		t.Fatalf(
			"expected 2 role lookups, got %d",
			len(roles),
		)
	}

	if roles[0] != "client" {
		t.Fatalf(
			"expected first role client, got %q",
			roles[0],
		)
	}

	if roles[1] != "cleaner" {
		t.Fatalf(
			"expected second role cleaner, got %q",
			roles[1],
		)
	}
}

func TestService_Dashboard_TotalUsersError(
	t *testing.T,
) {
	expectedErr := errors.New("total users failed")

	repo := &mockRepository{
		totalUsersFn: func(context.Context) (int, error) {
			return 0, expectedErr
		},
	}

	service := NewService(repo)

	stats, err := service.Dashboard(
		context.Background(),
	)

	if stats != nil {
		t.Fatal("expected nil stats")
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected %v, got %v",
			expectedErr,
			err,
		)
	}
}

func TestService_Dashboard_ClientCountError(
	t *testing.T,
) {
	expectedErr := errors.New("client count failed")

	repo := &mockRepository{
		totalUsersFn: func(context.Context) (int, error) {
			return 100, nil
		},
		usersByRoleFn: func(
			context.Context,
			string,
		) (int, error) {
			return 0, expectedErr
		},
	}

	service := NewService(repo)

	stats, err := service.Dashboard(
		context.Background(),
	)

	if stats != nil {
		t.Fatal("expected nil stats")
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected %v, got %v",
			expectedErr,
			err,
		)
	}
}

func TestService_Dashboard_CleanerCountError(
	t *testing.T,
) {
	expectedErr := errors.New("cleaner count failed")

	callCount := 0

	repo := &mockRepository{
		totalUsersFn: func(context.Context) (int, error) {
			return 100, nil
		},
		usersByRoleFn: func(
			ctx context.Context,
			role string,
		) (int, error) {
			callCount++

			if role == "cleaner" {
				return 0, expectedErr
			}

			return 40, nil
		},
	}

	service := NewService(repo)

	stats, err := service.Dashboard(
		context.Background(),
	)

	if stats != nil {
		t.Fatal("expected nil stats")
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected %v, got %v",
			expectedErr,
			err,
		)
	}

	if callCount != 2 {
		t.Fatalf(
			"expected 2 role calls, got %d",
			callCount,
		)
	}
}

func TestService_Dashboard_TotalCompaniesError(
	t *testing.T,
) {
	expectedErr := errors.New("companies failed")

	repo := dashboardBaseMock()

	repo.totalCompaniesFn = func(
		context.Context,
	) (int, error) {
		return 0, expectedErr
	}

	assertDashboardError(
		t,
		repo,
		expectedErr,
	)
}

func TestService_Dashboard_ActiveJobsError(
	t *testing.T,
) {
	expectedErr := errors.New("jobs failed")

	repo := dashboardBaseMock()

	repo.activeJobsFn = func(
		context.Context,
	) (int, error) {
		return 0, expectedErr
	}

	assertDashboardError(
		t,
		repo,
		expectedErr,
	)
}

func TestService_Dashboard_CompletedBookingsError(
	t *testing.T,
) {
	expectedErr := errors.New("bookings failed")

	repo := dashboardBaseMock()

	repo.completedBookingsFn = func(
		context.Context,
	) (int, error) {
		return 0, expectedErr
	}

	assertDashboardError(
		t,
		repo,
		expectedErr,
	)
}

func TestService_Dashboard_PendingVerificationsError(
	t *testing.T,
) {
	expectedErr := errors.New("verifications failed")

	repo := dashboardBaseMock()

	repo.pendingVerificationsFn = func(
		context.Context,
	) (int, error) {
		return 0, expectedErr
	}

	assertDashboardError(
		t,
		repo,
		expectedErr,
	)
}

func TestService_Dashboard_OpenReportsError(
	t *testing.T,
) {
	expectedErr := errors.New("reports failed")

	repo := dashboardBaseMock()

	repo.openReportsFn = func(
		context.Context,
	) (int, error) {
		return 0, expectedErr
	}

	assertDashboardError(
		t,
		repo,
		expectedErr,
	)
}

func TestService_Dashboard_BookingsTodayError(
	t *testing.T,
) {
	expectedErr := errors.New("today failed")

	repo := dashboardBaseMock()

	repo.bookingsTodayFn = func(
		context.Context,
	) (int, error) {
		return 0, expectedErr
	}

	assertDashboardError(
		t,
		repo,
		expectedErr,
	)
}

func dashboardBaseMock() *mockRepository {
	return &mockRepository{
		totalUsersFn: func(context.Context) (int, error) {
			return 100, nil
		},
		usersByRoleFn: func(
			context.Context,
			string,
		) (int, error) {
			return 10, nil
		},
		totalCompaniesFn: func(context.Context) (int, error) {
			return 5, nil
		},
		activeJobsFn: func(context.Context) (int, error) {
			return 5, nil
		},
		completedBookingsFn: func(context.Context) (int, error) {
			return 5, nil
		},
		pendingVerificationsFn: func(context.Context) (int, error) {
			return 5, nil
		},
		openReportsFn: func(context.Context) (int, error) {
			return 5, nil
		},
		bookingsTodayFn: func(context.Context) (int, error) {
			return 5, nil
		},
	}
}

func assertDashboardError(
	t *testing.T,
	repo *mockRepository,
	expectedErr error,
) {
	t.Helper()

	service := NewService(repo)

	stats, err := service.Dashboard(
		context.Background(),
	)

	if stats != nil {
		t.Fatal("expected nil stats")
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected %v, got %v",
			expectedErr,
			err,
		)
	}
}
