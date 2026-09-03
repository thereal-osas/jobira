package reputation

import (
	"context"
	"errors"
	"testing"
)

type mockRepository struct {
	ensureCleanerFn  func(ctx context.Context, cleanerID uint) error
	refreshFn        func(ctx context.Context, cleanerID uint) error
	getByCleanerIDFn func(ctx context.Context, cleanerID uint) (*CleanerReputation, error)
	updateBadgeFn    func(ctx context.Context, cleanerID uint, badge string) error
}

func (m *mockRepository) EnsureCleaner(ctx context.Context, cleanerID uint) error {
	if m.ensureCleanerFn != nil {
		return m.ensureCleanerFn(ctx, cleanerID)
	}

	return nil
}

func (m *mockRepository) Refresh(ctx context.Context, cleanerID uint) error {
	if m.refreshFn != nil {
		return m.refreshFn(ctx, cleanerID)
	}

	return nil
}

func (m *mockRepository) GetByCleanerID(
	ctx context.Context,
	cleanerID uint,
) (*CleanerReputation, error) {
	if m.getByCleanerIDFn != nil {
		return m.getByCleanerIDFn(ctx, cleanerID)
	}

	return nil, nil
}

func (m *mockRepository) UpdateBadge(
	ctx context.Context,
	cleanerID uint,
	badge string,
) error {
	if m.updateBadgeFn != nil {
		return m.updateBadgeFn(ctx, cleanerID, badge)
	}

	return nil
}

func TestService_GetByCleanerID_Success(t *testing.T) {
	repository := &mockRepository{
		refreshFn: func(ctx context.Context, cleanerID uint) error {
			if cleanerID != 10 {
				t.Fatalf("expected cleaner ID 10, got %d", cleanerID)
			}

			return nil
		},
		getByCleanerIDFn: func(
			ctx context.Context,
			cleanerID uint,
		) (*CleanerReputation, error) {
			return &CleanerReputation{
				CleanerID:                cleanerID,
				AverageRating:            5.0,
				TotalReviews:             25,
				CompletedJobs:            50,
				RepeatClients:            10,
				WouldHireAgainCount:      25,
				RecommendationPercentage: 100,

				TotalBookings:        50,
				CleanerCancellations: 0,

				EligibleResponseMessages: 30,
				RespondedMessages:        30,
				AverageResponseMinutes:   8,

				Badge: "Old Badge",
			}, nil
		},
	}

	service := NewService(repository)

	result, err := service.GetByCleanerID(context.Background(), 10)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result == nil {
		t.Fatal("expected reputation, got nil")
	}

	if result.CleanerID != 10 {
		t.Fatalf("expected cleaner ID 10, got %d", result.CleanerID)
	}

	if result.Badge != "Elite Cleaner" {
		t.Fatalf("expected Elite Cleaner badge, got %q", result.Badge)
	}
}

func TestService_GetByCleanerID_InvalidCleanerID(t *testing.T) {
	refreshCalled := false

	repository := &mockRepository{
		refreshFn: func(ctx context.Context, cleanerID uint) error {
			refreshCalled = true
			return nil
		},
	}

	service := NewService(repository)

	result, err := service.GetByCleanerID(context.Background(), 0)

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}

	if result != nil {
		t.Fatalf("expected nil result, got %+v", result)
	}

	if refreshCalled {
		t.Fatal("expected repository Refresh not to be called")
	}
}

func TestService_GetByCleanerID_RefreshError(t *testing.T) {
	expectedErr := errors.New("refresh failed")
	getCalled := false

	repository := &mockRepository{
		refreshFn: func(ctx context.Context, cleanerID uint) error {
			return expectedErr
		},
		getByCleanerIDFn: func(
			ctx context.Context,
			cleanerID uint,
		) (*CleanerReputation, error) {
			getCalled = true
			return nil, nil
		},
	}

	service := NewService(repository)

	result, err := service.GetByCleanerID(context.Background(), 10)

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected refresh error, got %v", err)
	}

	if result != nil {
		t.Fatalf("expected nil result, got %+v", result)
	}

	if getCalled {
		t.Fatal("expected GetByCleanerID not to be called")
	}
}

func TestService_GetByCleanerID_RepositoryGetError(t *testing.T) {
	expectedErr := errors.New("get reputation failed")

	repository := &mockRepository{
		refreshFn: func(ctx context.Context, cleanerID uint) error {
			return nil
		},
		getByCleanerIDFn: func(
			ctx context.Context,
			cleanerID uint,
		) (*CleanerReputation, error) {
			return nil, expectedErr
		},
	}

	service := NewService(repository)

	result, err := service.GetByCleanerID(context.Background(), 10)

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected repository error, got %v", err)
	}

	if result != nil {
		t.Fatalf("expected nil result, got %+v", result)
	}
}

func TestService_Refresh_Success(t *testing.T) {
	updateBadgeCalled := false

	repository := &mockRepository{
		refreshFn: func(ctx context.Context, cleanerID uint) error {
			if cleanerID != 20 {
				t.Fatalf("expected cleaner ID 20, got %d", cleanerID)
			}

			return nil
		},
		getByCleanerIDFn: func(
			ctx context.Context,
			cleanerID uint,
		) (*CleanerReputation, error) {
			return &CleanerReputation{
				CleanerID:                10,
				AverageRating:            4.9,
				TotalReviews:             15,
				CompletedJobs:            20,
				RepeatClients:            4,
				WouldHireAgainCount:      18,
				RecommendationPercentage: 95,

				TotalBookings:        21,
				CleanerCancellations: 1,

				EligibleResponseMessages: 20,
				RespondedMessages:        18,
				AverageResponseMinutes:   20,
			}, nil
		},
		updateBadgeFn: func(
			ctx context.Context,
			cleanerID uint,
			badge string,
		) error {
			updateBadgeCalled = true

			if cleanerID != 20 {
				t.Fatalf("expected cleaner ID 20, got %d", cleanerID)
			}

			if badge != "Top Rated" {
				t.Fatalf("expected Top Rated badge, got %q", badge)
			}

			return nil
		},
	}

	service := NewService(repository)

	result, err := service.Refresh(context.Background(), 20)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result == nil {
		t.Fatal("expected reputation, got nil")
	}

	if result.Badge != "Top Rated" {
		t.Fatalf("expected Top Rated badge, got %q", result.Badge)
	}

	if !updateBadgeCalled {
		t.Fatal("expected UpdateBadge to be called")
	}
}

func TestService_Refresh_InvalidCleanerID(t *testing.T) {
	repository := &mockRepository{}

	service := NewService(repository)

	result, err := service.Refresh(context.Background(), 0)

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}

	if result != nil {
		t.Fatalf("expected nil result, got %+v", result)
	}
}

func TestService_Refresh_RepositoryRefreshError(t *testing.T) {
	expectedErr := errors.New("refresh failed")

	repository := &mockRepository{
		refreshFn: func(ctx context.Context, cleanerID uint) error {
			return expectedErr
		},
	}

	service := NewService(repository)

	result, err := service.Refresh(context.Background(), 10)

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected refresh error, got %v", err)
	}

	if result != nil {
		t.Fatalf("expected nil result, got %+v", result)
	}
}

func TestService_Refresh_GetByCleanerIDError(t *testing.T) {
	expectedErr := errors.New("get failed")

	repository := &mockRepository{
		refreshFn: func(ctx context.Context, cleanerID uint) error {
			return nil
		},
		getByCleanerIDFn: func(
			ctx context.Context,
			cleanerID uint,
		) (*CleanerReputation, error) {
			return nil, expectedErr
		},
	}

	service := NewService(repository)

	result, err := service.Refresh(context.Background(), 10)

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected get error, got %v", err)
	}

	if result != nil {
		t.Fatalf("expected nil result, got %+v", result)
	}
}

func TestService_Refresh_UpdateBadgeError(t *testing.T) {
	expectedErr := errors.New("update badge failed")

	repository := &mockRepository{
		refreshFn: func(ctx context.Context, cleanerID uint) error {
			return nil
		},
		getByCleanerIDFn: func(
			ctx context.Context,
			cleanerID uint,
		) (*CleanerReputation, error) {
			return &CleanerReputation{
				CleanerID:     cleanerID,
				CompletedJobs: 5,
			}, nil
		},
		updateBadgeFn: func(
			ctx context.Context,
			cleanerID uint,
			badge string,
		) error {
			return expectedErr
		},
	}

	service := NewService(repository)

	result, err := service.Refresh(context.Background(), 10)

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected update badge error, got %v", err)
	}

	if result != nil {
		t.Fatalf("expected nil result, got %+v", result)
	}
}

func TestCalculateBadge(t *testing.T) {
	tests := []struct {
		name       string
		reputation *CleanerReputation
		expected   string
	}{
		{
			name:       "nil reputation",
			reputation: nil,
			expected:   "New Cleaner",
		},
		{
			name: "elite cleaner",
			reputation: &CleanerReputation{
				TotalReviews:             20,
				AverageRating:            4.9,
				RecommendationPercentage: 95,
				CompletedJobs:            50,
				ReliabilityScore:         90,
			},
			expected: "Elite Cleaner",
		},
		{
			name: "top rated",
			reputation: &CleanerReputation{
				TotalReviews:             10,
				AverageRating:            4.8,
				RecommendationPercentage: 90,
				CompletedJobs:            10,
				ReliabilityScore:         80,
			},
			expected: "Top Rated",
		},
		{
			name: "experienced cleaner",
			reputation: &CleanerReputation{
				TotalReviews:             5,
				AverageRating:            4.5,
				RecommendationPercentage: 80,
				CompletedJobs:            50,
			},
			expected: "Experienced Cleaner",
		},
		{
			name: "active cleaner",
			reputation: &CleanerReputation{
				CompletedJobs: 1,
			},
			expected: "Active Cleaner",
		},
		{
			name: "new cleaner",
			reputation: &CleanerReputation{
				CompletedJobs: 0,
			},
			expected: "New Cleaner",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := calculateBadge(test.reputation)

			if result != test.expected {
				t.Fatalf(
					"expected badge %q, got %q",
					test.expected,
					result,
				)
			}
		})
	}
}

func TestMilestoneProgress(t *testing.T) {
	tests := []struct {
		completedJobs int
		current       int
		next          int
		remaining     int
		progress      int
	}{
		{0, 0, 1, 1, 0},
		{1, 1, 5, 4, 0},
		{5, 5, 10, 5, 0},
		{10, 10, 15, 5, 0},
		{12, 10, 15, 3, 40},
		{15, 15, 25, 10, 0},
		{20, 15, 25, 5, 50},
		{25, 25, 50, 25, 0},
		{50, 50, 100, 50, 0},
		{100, 100, 0, 0, 100},
	}

	for _, tt := range tests {
		reputation := &CleanerReputation{
			CompletedJobs: tt.completedJobs,
		}

		setMilestoneProgress(reputation)

		if reputation.CurrentMilestone != tt.current {
			t.Fatalf(
				"completed jobs %d: expected current milestone %d, got %d",
				tt.completedJobs,
				tt.current,
				reputation.CurrentMilestone,
			)
		}

		if reputation.NextMilestone != tt.next {
			t.Fatalf(
				"completed jobs %d: expected next milestone %d, got %d",
				tt.completedJobs,
				tt.next,
				reputation.NextMilestone,
			)
		}

		if reputation.JobsUntilNextMilestone != tt.remaining {
			t.Fatalf(
				"completed jobs %d: expected %d jobs remaining, got %d",
				tt.completedJobs,
				tt.remaining,
				reputation.JobsUntilNextMilestone,
			)
		}

		if reputation.MilestoneProgress != tt.progress {
			t.Fatalf(
				"completed jobs %d: expected progress %d, got %d",
				tt.completedJobs,
				tt.progress,
				reputation.MilestoneProgress,
			)
		}
	}
}

func TestAchievements_JobMilestones(t *testing.T) {
	reputation := &CleanerReputation{
		CompletedJobs: 15,
	}

	achievements := calculateAchievements(reputation)

	tests := []struct {
		code   string
		earned bool
	}{
		{"first_job", true},
		{"getting_started", true},
		{"established_cleaner", true},
		{"rising_pro", true},
		{"experienced_cleaner", false},
		{"proven_professional", false},
		{"century_cleaner", false},
	}

	for _, tt := range tests {
		achievement := findAchievementForTest(
			achievements,
			tt.code,
		)

		if achievement == nil {
			t.Fatalf(
				"expected achievement %q",
				tt.code,
			)
		}

		if achievement.Earned != tt.earned {
			t.Fatalf(
				"achievement %q: expected earned=%v, got %v",
				tt.code,
				tt.earned,
				achievement.Earned,
			)
		}
	}
}

func TestAchievements_CenturyCleaner(t *testing.T) {
	reputation := &CleanerReputation{
		CompletedJobs: 100,
	}

	achievements := calculateAchievements(reputation)

	achievement := findAchievementForTest(
		achievements,
		"century_cleaner",
	)

	if achievement == nil {
		t.Fatal("expected century cleaner achievement")
	}

	if !achievement.Earned {
		t.Fatal("expected century cleaner achievement to be earned")
	}
}

func TestReliabilityRates_IgnoreClientCancellations(t *testing.T) {
	reputation := &CleanerReputation{
		CompletedJobs:        8,
		TotalBookings:        12,
		CleanerCancellations: 1,
	}

	completionRate := calculateCompletionRate(
		reputation,
	)

	cancellationRate := calculateCancellationRate(
		reputation,
	)

	if completionRate != 89 {
		t.Fatalf(
			"expected completion rate 89, got %d",
			completionRate,
		)
	}

	if cancellationRate != 11 {
		t.Fatalf(
			"expected cancellation rate 11, got %d",
			cancellationRate,
		)
	}
}

func TestReliabilityScore_ClientCancellationsDoNotReduceScore(
	t *testing.T,
) {
	base := &CleanerReputation{
		AverageRating:            4.8,
		TotalReviews:             12,
		CompletedJobs:            8,
		RepeatClients:            2,
		RecommendationPercentage: 90,
		CleanerCancellations:     1,
		EligibleResponseMessages: 20,
		RespondedMessages:        18,
		AverageResponseMinutes:   20,
		TotalBookings:            9,
	}

	withClientCancellations := &CleanerReputation{
		AverageRating:            4.8,
		TotalReviews:             12,
		CompletedJobs:            8,
		RepeatClients:            2,
		RecommendationPercentage: 90,
		CleanerCancellations:     1,
		EligibleResponseMessages: 20,
		RespondedMessages:        18,
		AverageResponseMinutes:   20,

		// Three extra bookings were cancelled by clients/admins.
		TotalBookings: 12,
	}

	enrichReputation(base)
	enrichReputation(withClientCancellations)

	if base.ReliabilityScore !=
		withClientCancellations.ReliabilityScore {
		t.Fatalf(
			"client cancellations changed reliability score: %d vs %d",
			base.ReliabilityScore,
			withClientCancellations.ReliabilityScore,
		)
	}

	if base.CompletionRate !=
		withClientCancellations.CompletionRate {
		t.Fatalf(
			"client cancellations changed completion rate: %d vs %d",
			base.CompletionRate,
			withClientCancellations.CompletionRate,
		)
	}
}

func TestReliabilityStatus_HighlyReliable(t *testing.T) {
	reputation := &CleanerReputation{
		CompletedJobs:    15,
		ReliabilityScore: 92,
	}

	status := calculateReliabilityStatus(
		reputation,
	)

	if status != "Highly Reliable" {
		t.Fatalf(
			"expected Highly Reliable, got %q",
			status,
		)
	}
}

func TestAchievements_HighlyReliable(t *testing.T) {
	reputation := &CleanerReputation{
		CompletedJobs:    10,
		ReliabilityScore: 90,
	}

	achievement := calculateHighlyReliableAchievement(
		reputation,
	)

	if !achievement.Earned {
		t.Fatal(
			"expected highly reliable achievement to be earned",
		)
	}
}

func TestAchievements_HighlyRecommended(t *testing.T) {
	reputation := &CleanerReputation{
		TotalReviews:             10,
		RecommendationPercentage: 92,
	}

	achievement := calculateHighlyRecommendedAchievement(
		reputation,
	)

	if !achievement.Earned {
		t.Fatal(
			"expected highly recommended achievement to be earned",
		)
	}
}

func TestAchievements_RepeatClientFavourite(t *testing.T) {
	reputation := &CleanerReputation{
		CompletedJobs: 15,
		RepeatClients: 3,
	}

	achievement := calculateRepeatClientAchievement(
		reputation,
	)

	if !achievement.Earned {
		t.Fatal(
			"expected repeat client favourite achievement to be earned",
		)
	}
}

func TestAchievements_QuickResponder(t *testing.T) {
	reputation := &CleanerReputation{
		EligibleResponseMessages: 10,
		ResponseRate:             90,
		AverageResponseMinutes:   45,
	}

	quick := calculateQuickResponderAchievement(
		reputation,
	)

	if !quick.Earned {
		t.Fatal(
			"expected quick responder achievement to be earned",
		)
	}

	fast := calculateFastResponderAchievement(
		reputation,
	)

	if fast.Earned {
		t.Fatal(
			"did not expect fast responder achievement yet",
		)
	}

	rapid := calculateRapidResponderAchievement(
		reputation,
	)

	if rapid.Earned {
		t.Fatal(
			"did not expect rapid responder achievement yet",
		)
	}
}

func TestAchievements_AllResponderLevels(t *testing.T) {
	reputation := &CleanerReputation{
		EligibleResponseMessages: 30,
		ResponseRate:             95,
		AverageResponseMinutes:   8,
	}

	quick := calculateQuickResponderAchievement(
		reputation,
	)

	fast := calculateFastResponderAchievement(
		reputation,
	)

	rapid := calculateRapidResponderAchievement(
		reputation,
	)

	if !quick.Earned {
		t.Fatal("expected Quick Responder")
	}

	if !fast.Earned {
		t.Fatal("expected Fast Responder")
	}

	if !rapid.Earned {
		t.Fatal("expected Rapid Responder")
	}
}

func TestResponseRate(t *testing.T) {
	reputation := &CleanerReputation{
		EligibleResponseMessages: 20,
		RespondedMessages:        18,
	}

	rate := calculateResponseRate(
		reputation,
	)

	if rate != 90 {
		t.Fatalf(
			"expected response rate 90, got %d",
			rate,
		)
	}
}

func TestResponseRate_NoEligibleMessages(t *testing.T) {
	reputation := &CleanerReputation{}

	rate := calculateResponseRate(
		reputation,
	)

	if rate != 0 {
		t.Fatalf(
			"expected response rate 0, got %d",
			rate,
		)
	}
}

func TestCalculateBadge_TopRatedRequiresReliability(
	t *testing.T,
) {
	reputation := &CleanerReputation{
		TotalReviews:             15,
		AverageRating:            4.9,
		RecommendationPercentage: 95,
		CompletedJobs:            20,
		ReliabilityScore:         79,
	}

	badge := calculateBadge(
		reputation,
	)

	if badge == "Top Rated" {
		t.Fatal(
			"did not expect Top Rated below reliability threshold",
		)
	}

	reputation.ReliabilityScore = 85

	badge = calculateBadge(
		reputation,
	)

	if badge != "Top Rated" {
		t.Fatalf(
			"expected Top Rated, got %q",
			badge,
		)
	}
}

func TestCalculateBadge_EliteRequiresReliability(
	t *testing.T,
) {
	reputation := &CleanerReputation{
		TotalReviews:             25,
		AverageRating:            4.95,
		RecommendationPercentage: 97,
		CompletedJobs:            50,
		ReliabilityScore:         89,
	}

	badge := calculateBadge(
		reputation,
	)

	if badge == "Elite Cleaner" {
		t.Fatal(
			"did not expect Elite Cleaner below reliability threshold",
		)
	}

	reputation.ReliabilityScore = 92

	badge = calculateBadge(
		reputation,
	)

	if badge != "Elite Cleaner" {
		t.Fatalf(
			"expected Elite Cleaner, got %q",
			badge,
		)
	}
}

func findAchievementForTest(
	achievements []Achievement,
	code string,
) *Achievement {
	for i := range achievements {
		if achievements[i].Code == code {
			return &achievements[i]
		}
	}

	return nil
}
