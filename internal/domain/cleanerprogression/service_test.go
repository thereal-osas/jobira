package cleanerprogression

import (
	"context"
	"errors"
	"testing"

	reputationdomain "github.com/rodrigueghenda/jobira/internal/domain/reputation"
)

type mockRepository struct {
	getCleanerReputationFn func(
		context.Context,
		uint,
	) (*reputationdomain.CleanerReputation, error)
}

func (m *mockRepository) GetCleanerReputation(
	ctx context.Context,
	cleanerID uint,
) (*reputationdomain.CleanerReputation, error) {
	if m.getCleanerReputationFn != nil {
		return m.getCleanerReputationFn(
			ctx,
			cleanerID,
		)
	}

	return nil, nil
}

func TestService_GetMyProgression_Success(t *testing.T) {
	repo := &mockRepository{
		getCleanerReputationFn: func(
			ctx context.Context,
			cleanerID uint,
		) (*reputationdomain.CleanerReputation, error) {
			if cleanerID != 10 {
				t.Fatalf(
					"expected cleaner ID 10, got %d",
					cleanerID,
				)
			}

			return &reputationdomain.CleanerReputation{
				CleanerID:                10,
				AverageRating:            4.9,
				TotalReviews:             15,
				CompletedJobs:            35,
				RepeatClients:            5,
				RecommendationPercentage: 94,
				ReliabilityScore:         92,

				TotalBookings:        40,
				CleanerCancellations: 2,

				EligibleResponseMessages: 30,
				RespondedMessages:        27,

				Badge: "Top Rated",
			}, nil
		},
	}

	service := NewService(repo)

	result, err :=
		service.GetMyProgression(
			context.Background(),
			10,
		)
	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}

	if result == nil {
		t.Fatal("expected progression")
	}

	if result.CleanerID != 10 {
		t.Fatalf(
			"expected cleaner ID 10, got %d",
			result.CleanerID,
		)
	}

	if result.CurrentBadge != "Top Rated" {
		t.Fatalf(
			"expected Top Rated, got %q",
			result.CurrentBadge,
		)
	}

	if result.CompletionRate != 87 {
		t.Fatalf(
			"expected completion rate 87, got %d",
			result.CompletionRate,
		)
	}

	if result.CancellationRate != 5 {
		t.Fatalf(
			"expected cancellation rate 5, got %d",
			result.CancellationRate,
		)
	}

	if result.ResponseRate != 90 {
		t.Fatalf(
			"expected response rate 90, got %d",
			result.ResponseRate,
		)
	}

	if result.BadgeProgress.NextBadge !=
		"Elite Cleaner" {
		t.Fatalf(
			"expected Elite Cleaner next badge, got %q",
			result.BadgeProgress.NextBadge,
		)
	}

	if len(result.Milestones) == 0 {
		t.Fatal("expected milestones")
	}
}

func TestService_GetMyProgression_InvalidCleanerID(
	t *testing.T,
) {
	service :=
		NewService(
			&mockRepository{},
		)

	result, err :=
		service.GetMyProgression(
			context.Background(),
			0,
		)

	if result != nil {
		t.Fatalf(
			"expected nil result, got %+v",
			result,
		)
	}

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

func TestService_GetMyProgression_RepositoryError(
	t *testing.T,
) {
	expectedErr :=
		errors.New(
			"repository failed",
		)

	repo :=
		&mockRepository{
			getCleanerReputationFn: func(
				context.Context,
				uint,
			) (*reputationdomain.CleanerReputation, error) {
				return nil, expectedErr
			},
		}

	service := NewService(repo)

	result, err :=
		service.GetMyProgression(
			context.Background(),
			10,
		)

	if result != nil {
		t.Fatalf(
			"expected nil result, got %+v",
			result,
		)
	}

	if !errors.Is(
		err,
		expectedErr,
	) {
		t.Fatalf(
			"expected %v, got %v",
			expectedErr,
			err,
		)
	}
}

func TestService_GetMyProgression_NotFound(
	t *testing.T,
) {
	repo :=
		&mockRepository{
			getCleanerReputationFn: func(
				context.Context,
				uint,
			) (*reputationdomain.CleanerReputation, error) {
				return nil, nil
			},
		}

	service := NewService(repo)

	result, err :=
		service.GetMyProgression(
			context.Background(),
			10,
		)

	if result != nil {
		t.Fatalf(
			"expected nil result, got %+v",
			result,
		)
	}

	if !errors.Is(
		err,
		reputationdomain.ErrReputationNotFound,
	) {
		t.Fatalf(
			"expected ErrReputationNotFound, got %v",
			err,
		)
	}
}

func TestCalculateBadgeProgress_Elite(
	t *testing.T,
) {
	result :=
		calculateBadgeProgress(
			&reputationdomain.CleanerReputation{
				Badge: "Elite Cleaner",
			},
		)

	if result.Progress != 100 {
		t.Fatalf(
			"expected 100, got %d",
			result.Progress,
		)
	}

	if !result.IsHighestLevel {
		t.Fatal(
			"expected highest level",
		)
	}
}

func TestCalculateBadgeProgress_TopRated(
	t *testing.T,
) {
	result :=
		calculateBadgeProgress(
			&reputationdomain.CleanerReputation{
				Badge:                    "Top Rated",
				TotalReviews:             15,
				AverageRating:            4.8,
				RecommendationPercentage: 92,
			},
		)

	if result.NextBadge !=
		"Elite Cleaner" {
		t.Fatalf(
			"expected Elite Cleaner, got %q",
			result.NextBadge,
		)
	}

	if result.Progress <= 0 {
		t.Fatalf(
			"expected positive progress, got %d",
			result.Progress,
		)
	}
}

func TestPercentage(t *testing.T) {
	tests := []struct {
		name     string
		value    int
		target   int
		expected int
	}{
		{
			name:     "zero target",
			value:    5,
			target:   0,
			expected: 0,
		},
		{
			name:     "zero value",
			value:    0,
			target:   10,
			expected: 0,
		},
		{
			name:     "half",
			value:    5,
			target:   10,
			expected: 50,
		},
		{
			name:     "complete",
			value:    10,
			target:   10,
			expected: 100,
		},
		{
			name:     "above target",
			value:    15,
			target:   10,
			expected: 100,
		},
	}

	for _, test := range tests {
		t.Run(
			test.name,
			func(t *testing.T) {
				result :=
					percentage(
						test.value,
						test.target,
					)

				if result != test.expected {
					t.Fatalf(
						"expected %d, got %d",
						test.expected,
						result,
					)
				}
			},
		)
	}
}
