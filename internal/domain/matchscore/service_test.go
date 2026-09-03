package matchscore

import (
	"context"
	"errors"
	"testing"
)

type mockRepository struct {
	listCandidatesFn     func(context.Context, uint, uint) ([]CandidateData, error)
	jobBelongsToClientFn func(context.Context, uint, uint) (bool, error)
}

func (m *mockRepository) ListCandidates(
	ctx context.Context,
	jobID uint,
	clientID uint,
) ([]CandidateData, error) {
	if m.listCandidatesFn != nil {
		return m.listCandidatesFn(
			ctx,
			jobID,
			clientID,
		)
	}

	return nil, nil
}

func (m *mockRepository) JobBelongsToClient(
	ctx context.Context,
	jobID uint,
	clientID uint,
) (bool, error) {
	if m.jobBelongsToClientFn != nil {
		return m.jobBelongsToClientFn(
			ctx,
			jobID,
			clientID,
		)
	}

	return false, nil
}

func TestService_RankJobCandidates_Success(t *testing.T) {
	repo := &mockRepository{
		jobBelongsToClientFn: func(
			_ context.Context,
			jobID uint,
			clientID uint,
		) (bool, error) {
			if jobID != 10 {
				t.Fatalf(
					"expected job ID 10, got %d",
					jobID,
				)
			}

			if clientID != 5 {
				t.Fatalf(
					"expected client ID 5, got %d",
					clientID,
				)
			}

			return true, nil
		},

		listCandidatesFn: func(
			_ context.Context,
			jobID uint,
			clientID uint,
		) ([]CandidateData, error) {
			return []CandidateData{
				{
					ApplicationID: 1,
					JobID:         jobID,
					ClientID:      clientID,
					CleanerID:     20,
					CleanerName:   "Strong Cleaner",

					JobType:     "end_of_tenancy",
					JobLocation: "East London",

					CleanerLocation:     "East London",
					CleanerPostcodeArea: "E14",
					ServicesOffered:     "domestic, end of tenancy, deep clean",

					AvailabilityStatus: "available",

					IsVerified: true,

					ReliabilityScore: 95,

					ResponseRate:           95,
					AverageResponseMinutes: 10,

					AverageRating: 4.9,
					TotalReviews:  30,

					Badge: "Elite Cleaner",
				},
				{
					ApplicationID: 2,
					JobID:         jobID,
					ClientID:      clientID,
					CleanerID:     21,
					CleanerName:   "Average Cleaner",

					JobType:     "end_of_tenancy",
					JobLocation: "East London",

					CleanerLocation: "London",
					ServicesOffered: "domestic cleaning",

					AvailabilityStatus: "limited",

					IsVerified: false,

					ReliabilityScore: 70,

					ResponseRate:           70,
					AverageResponseMinutes: 75,

					AverageRating: 4.5,
					TotalReviews:  8,

					Badge: "Active Cleaner",
				},
			}, nil
		},
	}

	service := NewService(repo)

	result, err := service.RankJobCandidates(
		context.Background(),
		10,
		5,
	)
	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}

	if result == nil {
		t.Fatal("expected result, got nil")
	}

	if result.JobID != 10 {
		t.Fatalf(
			"expected job ID 10, got %d",
			result.JobID,
		)
	}

	if result.TotalApplicants != 2 {
		t.Fatalf(
			"expected 2 applicants, got %d",
			result.TotalApplicants,
		)
	}

	if len(result.Matches) != 2 {
		t.Fatalf(
			"expected 2 matches, got %d",
			len(result.Matches),
		)
	}

	first := result.Matches[0]

	if first.CleanerID != 20 {
		t.Fatalf(
			"expected cleaner 20 ranked first, got %d",
			first.CleanerID,
		)
	}

	if first.Rank != 1 {
		t.Fatalf(
			"expected rank 1, got %d",
			first.Rank,
		)
	}

	if !first.IsTopOffer {
		t.Fatal(
			"expected first cleaner to be Top Offer",
		)
	}

	if !first.IsRecommended {
		t.Fatal(
			"expected first cleaner to be recommended",
		)
	}

	if result.TopOffer == nil {
		t.Fatal(
			"expected Top Offer in result",
		)
	}

	if result.TopOffer.CleanerID != 20 {
		t.Fatalf(
			"expected Top Offer cleaner 20, got %d",
			result.TopOffer.CleanerID,
		)
	}
}

func TestService_RankJobCandidates_InvalidInput(
	t *testing.T,
) {
	repo := &mockRepository{}

	service := NewService(repo)

	tests := []struct {
		jobID    uint
		clientID uint
	}{
		{0, 5},
		{10, 0},
		{0, 0},
	}

	for _, tt := range tests {
		result, err := service.RankJobCandidates(
			context.Background(),
			tt.jobID,
			tt.clientID,
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

		if result != nil {
			t.Fatalf(
				"expected nil result, got %+v",
				result,
			)
		}
	}
}

func TestService_RankJobCandidates_Forbidden(
	t *testing.T,
) {
	repo := &mockRepository{
		jobBelongsToClientFn: func(
			context.Context,
			uint,
			uint,
		) (bool, error) {
			return false, nil
		},
	}

	service := NewService(repo)

	result, err := service.RankJobCandidates(
		context.Background(),
		10,
		5,
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

	if result != nil {
		t.Fatalf(
			"expected nil result, got %+v",
			result,
		)
	}
}

func TestService_RankJobCandidates_OwnershipError(
	t *testing.T,
) {
	expectedErr := errors.New(
		"ownership lookup failed",
	)

	repo := &mockRepository{
		jobBelongsToClientFn: func(
			context.Context,
			uint,
			uint,
		) (bool, error) {
			return false, expectedErr
		},
	}

	service := NewService(repo)

	result, err := service.RankJobCandidates(
		context.Background(),
		10,
		5,
	)

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

	if result != nil {
		t.Fatalf(
			"expected nil result, got %+v",
			result,
		)
	}
}

func TestService_RankJobCandidates_ListError(
	t *testing.T,
) {
	expectedErr := errors.New(
		"candidate lookup failed",
	)

	repo := &mockRepository{
		jobBelongsToClientFn: func(
			context.Context,
			uint,
			uint,
		) (bool, error) {
			return true, nil
		},

		listCandidatesFn: func(
			context.Context,
			uint,
			uint,
		) ([]CandidateData, error) {
			return nil, expectedErr
		},
	}

	service := NewService(repo)

	result, err := service.RankJobCandidates(
		context.Background(),
		10,
		5,
	)

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

	if result != nil {
		t.Fatalf(
			"expected nil result, got %+v",
			result,
		)
	}
}

func TestService_RankJobCandidates_Empty(
	t *testing.T,
) {
	repo := &mockRepository{
		jobBelongsToClientFn: func(
			context.Context,
			uint,
			uint,
		) (bool, error) {
			return true, nil
		},

		listCandidatesFn: func(
			context.Context,
			uint,
			uint,
		) ([]CandidateData, error) {
			return []CandidateData{}, nil
		},
	}

	service := NewService(repo)

	result, err := service.RankJobCandidates(
		context.Background(),
		10,
		5,
	)
	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}

	if result.TotalApplicants != 0 {
		t.Fatalf(
			"expected zero applicants, got %d",
			result.TotalApplicants,
		)
	}

	if result.TopOffer != nil {
		t.Fatal(
			"expected no Top Offer",
		)
	}
}

func TestService_RankJobCandidates_NoTopOfferBelow70(
	t *testing.T,
) {
	repo := &mockRepository{
		jobBelongsToClientFn: func(
			context.Context,
			uint,
			uint,
		) (bool, error) {
			return true, nil
		},

		listCandidatesFn: func(
			context.Context,
			uint,
			uint,
		) ([]CandidateData, error) {
			return []CandidateData{
				{
					ApplicationID: 1,
					JobID:         10,
					ClientID:      5,
					CleanerID:     20,

					JobType:     "airbnb",
					JobLocation: "Canary Wharf",

					ServicesOffered: "domestic",

					AvailabilityStatus: "busy",

					ReliabilityScore: 50,

					ResponseRate: 40,
				},
			}, nil
		},
	}

	service := NewService(repo)

	result, err := service.RankJobCandidates(
		context.Background(),
		10,
		5,
	)
	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}

	if len(result.Matches) != 1 {
		t.Fatalf(
			"expected one match, got %d",
			len(result.Matches),
		)
	}

	if result.Matches[0].IsTopOffer {
		t.Fatal(
			"did not expect Top Offer below 70",
		)
	}

	if result.TopOffer != nil {
		t.Fatal(
			"expected nil TopOffer",
		)
	}
}

func TestService_RankJobCandidates_TieUsesReliability(
	t *testing.T,
) {
	repo := &mockRepository{
		jobBelongsToClientFn: func(
			context.Context,
			uint,
			uint,
		) (bool, error) {
			return true, nil
		},

		listCandidatesFn: func(
			context.Context,
			uint,
			uint,
		) ([]CandidateData, error) {
			return []CandidateData{
				{
					ApplicationID: 1,
					JobID:         10,
					ClientID:      5,
					CleanerID:     20,

					JobType: "domestic",

					ServicesOffered: "domestic",

					AvailabilityStatus: "available",

					ReliabilityScore: 80,

					IsVerified: true,
				},
				{
					ApplicationID: 2,
					JobID:         10,
					ClientID:      5,
					CleanerID:     21,

					JobType: "domestic",

					ServicesOffered: "domestic",

					AvailabilityStatus: "available",

					ReliabilityScore: 90,

					IsVerified: true,
				},
			}, nil
		},
	}

	service := NewService(repo)

	result, err := service.RankJobCandidates(
		context.Background(),
		10,
		5,
	)
	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}

	if result.Matches[0].CleanerID != 21 {
		t.Fatalf(
			"expected higher-reliability cleaner first, got %d",
			result.Matches[0].CleanerID,
		)
	}
}

func TestCalculateCandidateMatch_PerfectScore(
	t *testing.T,
) {
	candidate := CandidateData{
		ApplicationID: 1,
		JobID:         10,
		ClientID:      5,
		CleanerID:     20,

		JobType:     "end_of_tenancy",
		JobLocation: "East London",

		CleanerLocation: "East London",

		ServicesOffered: "end of tenancy",

		AvailabilityStatus: "available now",

		IsVerified: true,

		ReliabilityScore: 100,

		ResponseRate: 100,

		AverageResponseMinutes: 5,
	}

	result := calculateCandidateMatch(
		candidate,
	)

	if result.MatchScore != 100 {
		t.Fatalf(
			"expected score 100, got %d",
			result.MatchScore,
		)
	}

	if result.MatchLabel !=
		"Excellent Match" {
		t.Fatalf(
			"expected Excellent Match, got %q",
			result.MatchLabel,
		)
	}

	if result.Breakdown.ServiceFit != 25 {
		t.Fatalf(
			"expected service score 25, got %d",
			result.Breakdown.ServiceFit,
		)
	}

	if result.Breakdown.Availability != 25 {
		t.Fatalf(
			"expected availability score 25, got %d",
			result.Breakdown.Availability,
		)
	}

	if result.Breakdown.Reliability != 20 {
		t.Fatalf(
			"expected reliability score 20, got %d",
			result.Breakdown.Reliability,
		)
	}

	if result.Breakdown.Location != 15 {
		t.Fatalf(
			"expected location score 15, got %d",
			result.Breakdown.Location,
		)
	}

	if result.Breakdown.Verification != 10 {
		t.Fatalf(
			"expected verification score 10, got %d",
			result.Breakdown.Verification,
		)
	}

	if result.Breakdown.Response != 5 {
		t.Fatalf(
			"expected response score 5, got %d",
			result.Breakdown.Response,
		)
	}
}

func TestMatchLabel(t *testing.T) {
	tests := []struct {
		score    int
		expected string
	}{
		{100, "Excellent Match"},
		{90, "Excellent Match"},
		{89, "Strong Match"},
		{80, "Strong Match"},
		{79, "Good Match"},
		{70, "Good Match"},
		{69, "Potential Match"},
		{50, "Potential Match"},
		{49, "Low Match"},
		{0, "Low Match"},
	}

	for _, tt := range tests {
		actual := MatchLabel(
			tt.score,
		)

		if actual != tt.expected {
			t.Fatalf(
				"score %d: expected %q, got %q",
				tt.score,
				tt.expected,
				actual,
			)
		}
	}
}

func TestCalculateResponseScore(t *testing.T) {
	tests := []struct {
		rate     int
		minutes  int
		expected int
	}{
		{100, 10, 5},
		{90, 15, 5},
		{90, 45, 4},
		{80, 30, 3},
		{60, 120, 1},
		{0, 0, 0},
	}

	for _, tt := range tests {
		actual := calculateResponseScore(
			tt.rate,
			tt.minutes,
		)

		if actual != tt.expected {
			t.Fatalf(
				"rate=%d minutes=%d: expected %d, got %d",
				tt.rate,
				tt.minutes,
				tt.expected,
				actual,
			)
		}
	}
}

func TestBuildMatchReasons_AllPositiveSignals(t *testing.T) {
	breakdown := MatchScoreBreakdown{
		ServiceFit:   25,
		Availability: 25,
		Reliability:  18,
		Location:     15,
		Verification: 10,
		Response:     5,
	}

	reasons := buildMatchReasons(
		breakdown,
	)

	if len(reasons) != 6 {
		t.Fatalf(
			"expected 6 reasons, got %d",
			len(reasons),
		)
	}

	expectedCodes := []string{
		"service_fit",
		"availability",
		"location",
		"reliability",
		"verification",
		"response",
	}

	for i, expectedCode := range expectedCodes {
		if reasons[i].Code != expectedCode {
			t.Fatalf(
				"reason %d expected code %q, got %q",
				i,
				expectedCode,
				reasons[i].Code,
			)
		}
	}

	if reasons[0].Points != 25 {
		t.Fatalf(
			"expected service fit points 25, got %d",
			reasons[0].Points,
		)
	}

	if reasons[2].Points != 15 {
		t.Fatalf(
			"expected location points 15, got %d",
			reasons[2].Points,
		)
	}

	if reasons[4].Points != 10 {
		t.Fatalf(
			"expected verification points 10, got %d",
			reasons[4].Points,
		)
	}
}

func TestBuildMatchReasons_OnlyPositiveSignalsIncluded(t *testing.T) {
	breakdown := MatchScoreBreakdown{
		ServiceFit:   25,
		Availability: 0,
		Reliability:  0,
		Location:     15,
		Verification: 0,
		Response:     0,
	}

	reasons := buildMatchReasons(
		breakdown,
	)

	if len(reasons) != 2 {
		t.Fatalf(
			"expected 2 reasons, got %d",
			len(reasons),
		)
	}

	if reasons[0].Code != "service_fit" {
		t.Fatalf(
			"expected service_fit, got %q",
			reasons[0].Code,
		)
	}

	if reasons[1].Code != "location" {
		t.Fatalf(
			"expected location, got %q",
			reasons[1].Code,
		)
	}
}

func TestBuildMatchReasons_NoSignals(t *testing.T) {
	breakdown := MatchScoreBreakdown{}

	reasons := buildMatchReasons(
		breakdown,
	)

	if len(reasons) != 0 {
		t.Fatalf(
			"expected no reasons, got %d",
			len(reasons),
		)
	}
}

func TestBuildMatchReasons_PartialServiceFit(t *testing.T) {
	breakdown := MatchScoreBreakdown{
		ServiceFit: 15,
	}

	reasons := buildMatchReasons(
		breakdown,
	)

	if len(reasons) != 1 {
		t.Fatalf(
			"expected 1 reason, got %d",
			len(reasons),
		)
	}

	if reasons[0].Message !=
		"This cleaner offers services relevant to this job." {
		t.Fatalf(
			"unexpected message %q",
			reasons[0].Message,
		)
	}
}

func TestBuildMatchReasons_StrongServiceFit(t *testing.T) {
	breakdown := MatchScoreBreakdown{
		ServiceFit: 25,
	}

	reasons := buildMatchReasons(
		breakdown,
	)

	if len(reasons) != 1 {
		t.Fatalf(
			"expected 1 reason, got %d",
			len(reasons),
		)
	}

	if reasons[0].Message !=
		"This cleaner offers the service required for this job." {
		t.Fatalf(
			"unexpected message %q",
			reasons[0].Message,
		)
	}
}

func TestCalculateCandidateMatch_IncludesWhyYoureSeeingThis(t *testing.T) {
	candidate := CandidateData{
		ApplicationID: 10,
		JobID:         20,
		ClientID:      30,
		CleanerID:     40,

		CleanerName: "Test Cleaner",

		JobType:     "domestic",
		JobLocation: "London",

		CleanerLocation: "London",
		ServicesOffered: "domestic cleaning",

		AvailabilityStatus: "available",

		IsVerified: true,

		ReliabilityScore: 90,

		ResponseRate:           95,
		AverageResponseMinutes: 10,
	}

	result := calculateCandidateMatch(
		candidate,
	)

	if len(result.WhyYoureSeeingThis) == 0 {
		t.Fatal(
			"expected why you're seeing this reasons",
		)
	}

	if result.MatchScore <= 0 {
		t.Fatalf(
			"expected positive match score, got %d",
			result.MatchScore,
		)
	}

	for _, reason := range result.WhyYoureSeeingThis {
		if reason.Points <= 0 {
			t.Fatalf(
				"expected positive reason points for %q, got %d",
				reason.Code,
				reason.Points,
			)
		}
	}
}
