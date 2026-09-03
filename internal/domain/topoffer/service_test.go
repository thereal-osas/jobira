package topoffer

import (
	"context"
	"errors"
	"testing"
	"time"
)

type mockRepository struct {
	getJobContextFn func(
		context.Context,
		uint,
	) (*JobOfferContext, error)

	listOfferCandidatesFn func(
		context.Context,
		uint,
		int,
		int,
	) ([]OfferCandidate, error)

	getApplicationCandidateFn func(
		context.Context,
		uint,
	) (*OfferCandidate, error)

	countOffersFn func(
		context.Context,
		uint,
	) (int, error)
}

func (m *mockRepository) GetJobContext(
	ctx context.Context,
	jobID uint,
) (*JobOfferContext, error) {
	if m.getJobContextFn != nil {
		return m.getJobContextFn(
			ctx,
			jobID,
		)
	}

	return nil, ErrJobNotFound
}

func (m *mockRepository) ListOfferCandidates(
	ctx context.Context,
	jobID uint,
	limit int,
	offset int,
) ([]OfferCandidate, error) {
	if m.listOfferCandidatesFn != nil {
		return m.listOfferCandidatesFn(
			ctx,
			jobID,
			limit,
			offset,
		)
	}

	return []OfferCandidate{}, nil
}

func (m *mockRepository) GetApplicationCandidate(
	ctx context.Context,
	applicationID uint,
) (*OfferCandidate, error) {
	if m.getApplicationCandidateFn != nil {
		return m.getApplicationCandidateFn(
			ctx,
			applicationID,
		)
	}

	return nil, ErrApplicationNotFound
}

func (m *mockRepository) CountOffers(
	ctx context.Context,
	jobID uint,
) (int, error) {
	if m.countOffersFn != nil {
		return m.countOffersFn(
			ctx,
			jobID,
		)
	}

	return 0, nil
}

func TestService_RankJobOffers_Success(t *testing.T) {
	now := time.Now().UTC()

	repo := &mockRepository{
		getJobContextFn: func(
			context.Context,
			uint,
		) (*JobOfferContext, error) {
			return &JobOfferContext{
				JobID:    10,
				ClientID: 5,
				JobType:  "domestic",
				Budget:   100,
				Status:   "open",
				Location: "London",
			}, nil
		},

		listOfferCandidatesFn: func(
			context.Context,
			uint,
			int,
			int,
		) ([]OfferCandidate, error) {
			return []OfferCandidate{
				{
					ApplicationID:            1,
					JobID:                    10,
					CleanerID:                20,
					CleanerName:              "Sarah",
					ProposedRate:             95,
					ApplicationStatus:        "pending",
					AverageRating:            4.9,
					TotalReviews:             30,
					CompletedJobs:            50,
					ReliabilityScore:         95,
					RecommendationPercentage: 96,
					AverageResponseMinutes:   12,
					IsVerified:               true,
					AppliedAt:                now,
				},
				{
					ApplicationID:            2,
					JobID:                    10,
					CleanerID:                21,
					CleanerName:              "Maria",
					ProposedRate:             70,
					ApplicationStatus:        "pending",
					AverageRating:            4.2,
					TotalReviews:             8,
					CompletedJobs:            10,
					ReliabilityScore:         70,
					RecommendationPercentage: 75,
					AverageResponseMinutes:   90,
					IsVerified:               false,
					AppliedAt:                now.Add(time.Minute),
				},
			}, nil
		},

		countOffersFn: func(
			context.Context,
			uint,
		) (int, error) {
			return 2, nil
		},
	}

	service := NewService(repo)

	result, err := service.RankJobOffers(
		context.Background(),
		10,
		5,
		ListTopOffersRequest{},
	)
	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}

	if result == nil {
		t.Fatal("expected result")
	}

	if result.TotalOffers != 2 {
		t.Fatalf(
			"expected 2 offers, got %d",
			result.TotalOffers,
		)
	}

	if len(result.Offers) != 2 {
		t.Fatalf(
			"expected 2 ranked offers, got %d",
			len(result.Offers),
		)
	}

	if result.Offers[0].ApplicationID != 1 {
		t.Fatalf(
			"expected application 1 ranked first, got %d",
			result.Offers[0].ApplicationID,
		)
	}

	if result.Offers[0].Rank != 1 {
		t.Fatalf(
			"expected rank 1, got %d",
			result.Offers[0].Rank,
		)
	}

	if !result.Offers[0].IsTopOffer {
		t.Fatal(
			"expected first offer to be Top Offer",
		)
	}

	if result.TopOfferApplicationID == nil ||
		*result.TopOfferApplicationID != 1 {
		t.Fatal(
			"expected application 1 as Top Offer",
		)
	}
}

func TestService_RankJobOffers_Forbidden(t *testing.T) {
	repo := &mockRepository{
		getJobContextFn: func(
			context.Context,
			uint,
		) (*JobOfferContext, error) {
			return &JobOfferContext{
				JobID:    10,
				ClientID: 5,
				Status:   "open",
			}, nil
		},
	}

	service := NewService(repo)

	result, err := service.RankJobOffers(
		context.Background(),
		10,
		99,
		ListTopOffersRequest{},
	)

	if result != nil {
		t.Fatalf(
			"expected nil result, got %+v",
			result,
		)
	}

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

func TestService_RankJobOffers_ClosedJob(t *testing.T) {
	repo := &mockRepository{
		getJobContextFn: func(
			context.Context,
			uint,
		) (*JobOfferContext, error) {
			return &JobOfferContext{
				JobID:    10,
				ClientID: 5,
				Status:   "closed",
			}, nil
		},
	}

	service := NewService(repo)

	result, err := service.RankJobOffers(
		context.Background(),
		10,
		5,
		ListTopOffersRequest{},
	)

	if result != nil {
		t.Fatalf(
			"expected nil result, got %+v",
			result,
		)
	}

	if !errors.Is(
		err,
		ErrJobClosed,
	) {
		t.Fatalf(
			"expected ErrJobClosed, got %v",
			err,
		)
	}
}

func TestService_GetApplicationRanking_Success(
	t *testing.T,
) {
	now := time.Now().UTC()

	repo := &mockRepository{
		getApplicationCandidateFn: func(
			context.Context,
			uint,
		) (*OfferCandidate, error) {
			return &OfferCandidate{
				ApplicationID: 1,
				JobID:         10,
				CleanerID:     20,
			}, nil
		},

		getJobContextFn: func(
			context.Context,
			uint,
		) (*JobOfferContext, error) {
			return &JobOfferContext{
				JobID:    10,
				ClientID: 5,
				Budget:   100,
				Status:   "open",
			}, nil
		},

		listOfferCandidatesFn: func(
			context.Context,
			uint,
			int,
			int,
		) ([]OfferCandidate, error) {
			return []OfferCandidate{
				{
					ApplicationID:            1,
					JobID:                    10,
					CleanerID:                20,
					ProposedRate:             95,
					AverageRating:            4.9,
					CompletedJobs:            50,
					ReliabilityScore:         95,
					RecommendationPercentage: 96,
					AverageResponseMinutes:   10,
					IsVerified:               true,
					AppliedAt:                now,
				},
				{
					ApplicationID:            2,
					JobID:                    10,
					CleanerID:                21,
					ProposedRate:             110,
					AverageRating:            4.2,
					CompletedJobs:            5,
					ReliabilityScore:         70,
					RecommendationPercentage: 70,
					AverageResponseMinutes:   120,
					AppliedAt:                now.Add(time.Minute),
				},
			}, nil
		},
	}

	service := NewService(repo)

	result, err :=
		service.GetApplicationRanking(
			context.Background(),
			1,
			20,
		)
	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}

	if result.Rank != 1 {
		t.Fatalf(
			"expected rank 1, got %d",
			result.Rank,
		)
	}

	if !result.IsTopOffer {
		t.Fatal(
			"expected Top Offer",
		)
	}

	if result.Score <= 0 {
		t.Fatalf(
			"expected positive score, got %d",
			result.Score,
		)
	}
}

func TestService_GetApplicationRanking_Forbidden(
	t *testing.T,
) {
	repo := &mockRepository{
		getApplicationCandidateFn: func(
			context.Context,
			uint,
		) (*OfferCandidate, error) {
			return &OfferCandidate{
				ApplicationID: 1,
				JobID:         10,
				CleanerID:     20,
			}, nil
		},
	}

	service := NewService(repo)

	result, err :=
		service.GetApplicationRanking(
			context.Background(),
			1,
			99,
		)

	if result != nil {
		t.Fatalf(
			"expected nil result, got %+v",
			result,
		)
	}

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

func TestCalculatePriceFitScore(t *testing.T) {
	tests := []struct {
		budget   int
		rate     int
		expected int
	}{
		{100, 100, 100},
		{100, 95, 100},
		{100, 80, 90},
		{100, 60, 75},
		{100, 105, 85},
		{100, 115, 70},
		{100, 130, 50},
		{100, 200, 25},
	}

	for _, test := range tests {
		result :=
			calculatePriceFitScore(
				test.budget,
				test.rate,
			)

		if result != test.expected {
			t.Fatalf(
				"budget %d rate %d: expected %d, got %d",
				test.budget,
				test.rate,
				test.expected,
				result,
			)
		}
	}
}
