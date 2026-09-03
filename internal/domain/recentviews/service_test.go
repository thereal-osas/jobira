package recentviews

import (
	"context"
	"errors"
	"testing"
	"time"
)

type mockRepository struct {
	recordViewFn func(
		context.Context,
		uint,
		uint,
	) error

	listByClientIDFn func(
		context.Context,
		uint,
	) ([]RecentlyViewedCleaner, error)

	countByCleanerIDFn func(
		context.Context,
		uint,
	) (int, error)

	countUniqueViewersByCleanerIDFn func(
		context.Context,
		uint,
	) (int, error)

	countByCleanerIDSinceFn func(
		context.Context,
		uint,
		time.Time,
	) (int, error)

	countByCleanerIDBetweenFn func(
		context.Context,
		uint,
		time.Time,
		time.Time,
	) (int, error)
}

func (m *mockRepository) CountByCleanerID(
	ctx context.Context,
	cleanerID uint,
) (int, error) {
	if m.countByCleanerIDFn != nil {
		return m.countByCleanerIDFn(
			ctx,
			cleanerID,
		)
	}

	return 0, nil
}

func (m *mockRepository) CountUniqueViewersByCleanerID(
	ctx context.Context,
	cleanerID uint,
) (int, error) {
	if m.countUniqueViewersByCleanerIDFn != nil {
		return m.countUniqueViewersByCleanerIDFn(
			ctx,
			cleanerID,
		)
	}

	return 0, nil
}

func (m *mockRepository) CountByCleanerIDSince(
	ctx context.Context,
	cleanerID uint,
	since time.Time,
) (int, error) {
	if m.countByCleanerIDSinceFn != nil {
		return m.countByCleanerIDSinceFn(
			ctx,
			cleanerID,
			since,
		)
	}

	return 0, nil
}

func (m *mockRepository) CountByCleanerIDBetween(
	ctx context.Context,
	cleanerID uint,
	from time.Time,
	to time.Time,
) (int, error) {
	if m.countByCleanerIDBetweenFn != nil {
		return m.countByCleanerIDBetweenFn(
			ctx,
			cleanerID,
			from,
			to,
		)
	}

	return 0, nil
}

func (m *mockRepository) RecordView(
	ctx context.Context,
	clientID uint,
	cleanerID uint,
) error {
	if m.recordViewFn != nil {
		return m.recordViewFn(
			ctx,
			clientID,
			cleanerID,
		)
	}

	return nil
}

func (m *mockRepository) ListByClientID(
	ctx context.Context,
	clientID uint,
) ([]RecentlyViewedCleaner, error) {
	if m.listByClientIDFn != nil {
		return m.listByClientIDFn(
			ctx,
			clientID,
		)
	}

	return nil, nil
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

func TestService_RecordView_Success(t *testing.T) {
	recordCalled := false

	repo := &mockRepository{
		recordViewFn: func(
			ctx context.Context,
			clientID uint,
			cleanerID uint,
		) error {
			recordCalled = true

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
	}

	service := NewService(repo)

	err := service.RecordView(
		context.Background(),
		5,
		8,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !recordCalled {
		t.Fatal("expected RecordView repository method to be called")
	}
}

func TestService_RecordView_InvalidInput(t *testing.T) {
	service := NewService(&mockRepository{})

	tests := []struct {
		name      string
		clientID  uint
		cleanerID uint
	}{
		{
			name:      "zero client ID",
			clientID:  0,
			cleanerID: 8,
		},
		{
			name:      "zero cleaner ID",
			clientID:  5,
			cleanerID: 0,
		},
		{
			name:      "same client and cleaner",
			clientID:  5,
			cleanerID: 5,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := service.RecordView(
				context.Background(),
				test.clientID,
				test.cleanerID,
			)

			if !errors.Is(err, ErrInvalidInput) {
				t.Fatalf(
					"expected ErrInvalidInput, got %v",
					err,
				)
			}
		})
	}
}

func TestService_RecordView_RepositoryError(t *testing.T) {
	expectedErr := errors.New("record view failed")

	repo := &mockRepository{
		recordViewFn: func(
			context.Context,
			uint,
			uint,
		) error {
			return expectedErr
		},
	}

	service := NewService(repo)

	err := service.RecordView(
		context.Background(),
		5,
		8,
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected %v, got %v",
			expectedErr,
			err,
		)
	}
}

func TestService_ListMine_Success(t *testing.T) {
	now := time.Now()

	repo := &mockRepository{
		listByClientIDFn: func(
			ctx context.Context,
			clientID uint,
		) ([]RecentlyViewedCleaner, error) {
			if clientID != 5 {
				t.Fatalf(
					"expected client ID 5, got %d",
					clientID,
				)
			}

			return []RecentlyViewedCleaner{
				{
					ID:        1,
					ClientID:  5,
					CleanerID: 8,
					ViewedAt:  now,
				},
				{
					ID:        2,
					ClientID:  5,
					CleanerID: 9,
					ViewedAt:  now.Add(-time.Hour),
				},
			}, nil
		},
	}

	service := NewService(repo)

	views, err := service.ListMine(
		context.Background(),
		5,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(views) != 2 {
		t.Fatalf(
			"expected 2 recent views, got %d",
			len(views),
		)
	}

	if views[0].CleanerID != 8 {
		t.Fatalf(
			"expected first cleaner ID 8, got %d",
			views[0].CleanerID,
		)
	}

	if views[1].CleanerID != 9 {
		t.Fatalf(
			"expected second cleaner ID 9, got %d",
			views[1].CleanerID,
		)
	}
}

func TestService_ListMine_Empty(t *testing.T) {
	repo := &mockRepository{
		listByClientIDFn: func(
			context.Context,
			uint,
		) ([]RecentlyViewedCleaner, error) {
			return []RecentlyViewedCleaner{}, nil
		},
	}

	service := NewService(repo)

	views, err := service.ListMine(
		context.Background(),
		5,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(views) != 0 {
		t.Fatalf(
			"expected no recent views, got %d",
			len(views),
		)
	}
}

func TestService_ListMine_InvalidInput(t *testing.T) {
	service := NewService(&mockRepository{})

	views, err := service.ListMine(
		context.Background(),
		0,
	)

	if views != nil {
		t.Fatalf(
			"expected nil views, got %+v",
			views,
		)
	}

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf(
			"expected ErrInvalidInput, got %v",
			err,
		)
	}
}

func TestService_ListMine_RepositoryError(t *testing.T) {
	expectedErr := errors.New("list views failed")

	repo := &mockRepository{
		listByClientIDFn: func(
			context.Context,
			uint,
		) ([]RecentlyViewedCleaner, error) {
			return nil, expectedErr
		},
	}

	service := NewService(repo)

	views, err := service.ListMine(
		context.Background(),
		5,
	)

	if views != nil {
		t.Fatalf(
			"expected nil views, got %+v",
			views,
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
func TestService_GetProfileViewAnalytics_Success(
	t *testing.T,
) {
	now := time.Now()

	startOfToday := time.Date(
		now.Year(),
		now.Month(),
		now.Day(),
		0,
		0,
		0,
		0,
		now.Location(),
	)

	startOfThisWeek := startOfWeek(now)

	expectedViewsThisWeek := 18

	// On Monday, "today" and "this week" begin at the same timestamp,
	// so both repository queries correctly return the same count.
	if startOfToday.Equal(startOfThisWeek) {
		expectedViewsThisWeek = 5
	}

	repo := &mockRepository{
		countByCleanerIDFn: func(
			ctx context.Context,
			cleanerID uint,
		) (int, error) {
			if cleanerID != 8 {
				t.Fatalf(
					"expected cleaner ID 8, got %d",
					cleanerID,
				)
			}

			return 127, nil
		},

		countUniqueViewersByCleanerIDFn: func(
			ctx context.Context,
			cleanerID uint,
		) (int, error) {
			return 64, nil
		},

		countByCleanerIDSinceFn: func(
			ctx context.Context,
			cleanerID uint,
			since time.Time,
		) (int, error) {
			if since.Hour() != 0 ||
				since.Minute() != 0 ||
				since.Second() != 0 {
				t.Fatalf(
					"expected boundary at midnight, got %v",
					since,
				)
			}

			if since.Equal(startOfToday) {
				return 5, nil
			}

			if since.Equal(startOfThisWeek) {
				return 18, nil
			}

			t.Fatalf(
				"unexpected since boundary: %v",
				since,
			)

			return 0, nil
		},

		countByCleanerIDBetweenFn: func(
			ctx context.Context,
			cleanerID uint,
			from time.Time,
			to time.Time,
		) (int, error) {
			return 14, nil
		},
	}

	service := NewService(repo)

	result, err :=
		service.GetProfileViewAnalytics(
			context.Background(),
			8,
		)

	if err != nil {
		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}

	if result == nil {
		t.Fatal("expected analytics result")
	}

	if result.CleanerID != 8 {
		t.Fatalf(
			"expected cleaner ID 8, got %d",
			result.CleanerID,
		)
	}

	if result.TotalViews != 127 {
		t.Fatalf(
			"expected 127 total views, got %d",
			result.TotalViews,
		)
	}

	if result.UniqueViewers != 64 {
		t.Fatalf(
			"expected 64 unique viewers, got %d",
			result.UniqueViewers,
		)
	}

	if result.ViewsToday != 5 {
		t.Fatalf(
			"expected 5 views today, got %d",
			result.ViewsToday,
		)
	}

	if result.ViewsThisWeek != expectedViewsThisWeek {
		t.Fatalf(
			"expected %d views this week, got %d",
			expectedViewsThisWeek,
			result.ViewsThisWeek,
		)
	}

	if result.ViewsLastWeek != 14 {
		t.Fatalf(
			"expected 14 views last week, got %d",
			result.ViewsLastWeek,
		)
	}

	expectedChange :=
		(float64(expectedViewsThisWeek-14) / 14.0) * 100

	if result.WeeklyChange < expectedChange-0.001 ||
		result.WeeklyChange > expectedChange+0.001 {
		t.Fatalf(
			"expected weekly change %.2f, got %.2f",
			expectedChange,
			result.WeeklyChange,
		)
	}

	expectedGrowth := expectedViewsThisWeek > 14

	if result.HasWeeklyGrowth != expectedGrowth {
		t.Fatalf(
			"expected HasWeeklyGrowth %v, got %v",
			expectedGrowth,
			result.HasWeeklyGrowth,
		)
	}
}

func TestService_GetProfileViewAnalytics_InvalidCleanerID(
	t *testing.T,
) {
	service := NewService(
		&mockRepository{},
	)

	result, err :=
		service.GetProfileViewAnalytics(
			context.Background(),
			0,
		)

	if result != nil {
		t.Fatal(
			"expected nil result",
		)
	}

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf(
			"expected ErrInvalidInput, got %v",
			err,
		)
	}
}

func TestCalculateWeeklyChange_Growth(
	t *testing.T,
) {
	result := calculateWeeklyChange(
		18,
		14,
	)

	expected := (4.0 / 14.0) * 100

	if result < expected-0.001 ||
		result > expected+0.001 {
		t.Fatalf(
			"expected %.2f, got %.2f",
			expected,
			result,
		)
	}
}

func TestCalculateWeeklyChange_Decline(
	t *testing.T,
) {
	result := calculateWeeklyChange(
		10,
		20,
	)

	if result != -50 {
		t.Fatalf(
			"expected -50, got %.2f",
			result,
		)
	}
}

func TestCalculateWeeklyChange_NoPreviousViews(
	t *testing.T,
) {
	result := calculateWeeklyChange(
		5,
		0,
	)

	if result != 100 {
		t.Fatalf(
			"expected 100, got %.2f",
			result,
		)
	}
}

func TestCalculateWeeklyChange_NoViews(
	t *testing.T,
) {
	result := calculateWeeklyChange(
		0,
		0,
	)

	if result != 0 {
		t.Fatalf(
			"expected 0, got %.2f",
			result,
		)
	}
}

func TestStartOfWeek_Monday(
	t *testing.T,
) {
	input := time.Date(
		2026,
		time.August,
		29,
		14,
		30,
		0,
		0,
		time.UTC,
	)

	result := startOfWeek(input)

	expected := time.Date(
		2026,
		time.August,
		24,
		0,
		0,
		0,
		0,
		time.UTC,
	)

	if !result.Equal(expected) {
		t.Fatalf(
			"expected %v, got %v",
			expected,
			result,
		)
	}
}
