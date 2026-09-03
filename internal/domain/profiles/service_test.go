package profiles

import (
	"context"
	"errors"
	"testing"
	"time"
)

type MockRepository struct {
	profile  *CleanerProfile
	profiles []CleanerProfile
	history  []JobHistoryItem

	createErr                   error
	getByUserIDErr              error
	updateErr                   error
	updateVerificationStatusErr error
	searchErr                   error
	incrementCompletedErr       error
	incrementCancelledErr       error
	recalculateErr              error
	completedJobsErr            error
	cancelledJobsErr            error
	fullHistoryErr              error

	createdProfile *CleanerProfile
	updatedProfile *CleanerProfile

	lastVerificationUserID   uint
	lastVerificationStatus   string
	lastVerificationVerified bool

	lastSearchRequest SearchProfilesRequest

	incrementCompletedCalled bool
	incrementCancelledCalled bool
	recalculateCalled        bool
}

func (m *MockRepository) Create(
	ctx context.Context,
	profile *CleanerProfile,
) error {
	m.createdProfile = profile

	if m.createErr != nil {
		return m.createErr
	}

	return nil
}

func (m *MockRepository) GetByUserID(
	ctx context.Context,
	userID uint,
) (*CleanerProfile, error) {
	if m.getByUserIDErr != nil {
		return nil, m.getByUserIDErr
	}

	return m.profile, nil
}

func (m *MockRepository) Update(
	ctx context.Context,
	profile *CleanerProfile,
) error {
	m.updatedProfile = profile

	if m.updateErr != nil {
		return m.updateErr
	}

	return nil
}

func (m *MockRepository) UpdateVerificationStatus(
	ctx context.Context,
	userID uint,
	status string,
	verified bool,
) error {
	m.lastVerificationUserID = userID
	m.lastVerificationStatus = status
	m.lastVerificationVerified = verified

	return m.updateVerificationStatusErr
}

func (m *MockRepository) Search(
	ctx context.Context,
	req SearchProfilesRequest,
) ([]CleanerProfile, error) {
	m.lastSearchRequest = req

	if m.searchErr != nil {
		return nil, m.searchErr
	}

	return m.profiles, nil
}

func (m *MockRepository) IncrementJobsCompleted(
	ctx context.Context,
	userID uint,
) error {
	m.incrementCompletedCalled = true
	return m.incrementCompletedErr
}

func (m *MockRepository) IncrementJobsCancelled(
	ctx context.Context,
	userID uint,
) error {
	m.incrementCancelledCalled = true
	return m.incrementCancelledErr
}

func (m *MockRepository) RecalculateReputation(
	ctx context.Context,
	userID uint,
) error {
	m.recalculateCalled = true
	return m.recalculateErr
}

func (m *MockRepository) GetCompletedJobs(
	ctx context.Context,
	userID uint,
) ([]JobHistoryItem, error) {
	if m.completedJobsErr != nil {
		return nil, m.completedJobsErr
	}

	return m.history, nil
}

func (m *MockRepository) GetCancelledJobs(
	ctx context.Context,
	userID uint,
) ([]JobHistoryItem, error) {
	if m.cancelledJobsErr != nil {
		return nil, m.cancelledJobsErr
	}

	return m.history, nil
}

func (m *MockRepository) GetFullHistory(
	ctx context.Context,
	userID uint,
) ([]JobHistoryItem, error) {
	if m.fullHistoryErr != nil {
		return nil, m.fullHistoryErr
	}

	return m.history, nil
}

type mockPortfolioCounter struct {
	count int
	err   error
}

func (m *mockPortfolioCounter) CountPortfolioPhotos(
	ctx context.Context,
	userID uint,
) (int, error) {
	return m.count, m.err
}

func TestNewService(t *testing.T) {
	repo := &MockRepository{}

	service := NewService(repo)

	if service == nil {
		t.Fatal("expected service")
	}

	if service.repo != repo {
		t.Fatal("expected repository to be assigned")
	}
}

func TestCreate_InvalidUserID(t *testing.T) {
	service := NewService(&MockRepository{})

	profile, err := service.Create(
		context.Background(),
		0,
		CreateProfileRequest{
			Bio:             "Cleaner",
			Location:        "London",
			ServicesOffered: "Domestic",
		},
	)

	if profile != nil {
		t.Fatal("expected nil profile")
	}

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestCreate_MissingRequiredFields(t *testing.T) {
	tests := []struct {
		name string
		req  CreateProfileRequest
	}{
		{
			name: "missing bio",
			req: CreateProfileRequest{
				Location:        "London",
				ServicesOffered: "Domestic",
			},
		},
		{
			name: "missing location",
			req: CreateProfileRequest{
				Bio:             "Cleaner",
				ServicesOffered: "Domestic",
			},
		},
		{
			name: "missing services",
			req: CreateProfileRequest{
				Bio:      "Cleaner",
				Location: "London",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service := NewService(&MockRepository{})

			profile, err := service.Create(
				context.Background(),
				1,
				test.req,
			)

			if profile != nil {
				t.Fatal("expected nil profile")
			}

			if !errors.Is(err, ErrInvalidInput) {
				t.Fatalf(
					"expected ErrInvalidInput, got %v",
					err,
				)
			}
		})
	}
}

func TestCreate_InvalidAvailabilityStatus(t *testing.T) {
	service := NewService(&MockRepository{})

	_, err := service.Create(
		context.Background(),
		1,
		CreateProfileRequest{
			Bio:                "Cleaner",
			Location:           "London",
			ServicesOffered:    "Domestic",
			AvailabilityStatus: "sometimes_maybe",
		},
	)

	if !errors.Is(err, ErrInvalidAvailabilityStatus) {
		t.Fatalf(
			"expected ErrInvalidAvailabilityStatus, got %v",
			err,
		)
	}
}

func TestCreate_Success(t *testing.T) {
	repo := &MockRepository{}
	service := NewService(repo)

	profile, err := service.Create(
		context.Background(),
		7,
		CreateProfileRequest{
			Bio:                "  Experienced cleaner  ",
			Location:           "  East London  ",
			Country:            "UK",
			City:               "London",
			Region:             "East London",
			PostcodeArea:       "E14",
			AvailabilityStatus: "weekends_only",
			TravelRadiusMiles:  20,
			YearsExperience:    5,
			HourlyRate:         22,
			ServicesOffered:    "  domestic, airbnb  ",
		},
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if profile == nil {
		t.Fatal("expected profile")
	}

	if profile.UserID != 7 {
		t.Fatalf("expected user ID 7, got %d", profile.UserID)
	}

	if profile.Bio != "Experienced cleaner" {
		t.Fatalf("expected trimmed bio, got %q", profile.Bio)
	}

	if profile.Location != "East London" {
		t.Fatalf(
			"expected trimmed location, got %q",
			profile.Location,
		)
	}

	if profile.ServicesOffered != "domestic, airbnb" {
		t.Fatalf(
			"expected trimmed services, got %q",
			profile.ServicesOffered,
		)
	}

	if profile.AvailabilityStatus != "weekends_only" {
		t.Fatalf(
			"expected weekends_only, got %q",
			profile.AvailabilityStatus,
		)
	}

	if profile.TravelRadiusMiles != 20 {
		t.Fatalf(
			"expected travel radius 20, got %d",
			profile.TravelRadiusMiles,
		)
	}

	if profile.IsVerified {
		t.Fatal("expected new profile to be unverified")
	}

	if profile.VerificationStatus != "pending" {
		t.Fatalf(
			"expected pending verification, got %q",
			profile.VerificationStatus,
		)
	}

	if repo.createdProfile != profile {
		t.Fatal("expected repository to receive created profile")
	}
}

func TestCreate_DefaultAvailabilityAndTravelRadius(t *testing.T) {
	repo := &MockRepository{}
	service := NewService(repo)

	profile, err := service.Create(
		context.Background(),
		1,
		CreateProfileRequest{
			Bio:             "Cleaner",
			Location:        "London",
			ServicesOffered: "Domestic",
		},
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if profile.AvailabilityStatus != "available_immediately" {
		t.Fatalf(
			"expected default availability, got %q",
			profile.AvailabilityStatus,
		)
	}

	if profile.TravelRadiusMiles != 10 {
		t.Fatalf(
			"expected default travel radius 10, got %d",
			profile.TravelRadiusMiles,
		)
	}
}

func TestCreate_RepositoryError(t *testing.T) {
	createErr := errors.New("create failed")

	service := NewService(
		&MockRepository{
			createErr: createErr,
		},
	)

	profile, err := service.Create(
		context.Background(),
		1,
		CreateProfileRequest{
			Bio:             "Cleaner",
			Location:        "London",
			ServicesOffered: "Domestic",
		},
	)

	if profile != nil {
		t.Fatal("expected nil profile")
	}

	if !errors.Is(err, createErr) {
		t.Fatalf("expected create error, got %v", err)
	}
}

func TestGetMine_InvalidUser(t *testing.T) {
	service := NewService(&MockRepository{})

	_, err := service.GetMine(
		context.Background(),
		0,
	)

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestGetMine_Success(t *testing.T) {
	expected := &CleanerProfile{
		ID:     5,
		UserID: 10,
	}

	service := NewService(
		&MockRepository{
			profile: expected,
		},
	)

	result, err := service.GetMine(
		context.Background(),
		10,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != expected {
		t.Fatal("expected repository profile")
	}
}

func TestGetMine_RepositoryError(t *testing.T) {
	repositoryErr := errors.New("lookup failed")

	service := NewService(
		&MockRepository{
			getByUserIDErr: repositoryErr,
		},
	)

	_, err := service.GetMine(
		context.Background(),
		10,
	)

	if !errors.Is(err, repositoryErr) {
		t.Fatalf(
			"expected repository error, got %v",
			err,
		)
	}
}

func TestUpdate_InvalidAvailabilityStatus(t *testing.T) {
	service := NewService(&MockRepository{})

	_, err := service.Update(
		context.Background(),
		1,
		UpdateProfileRequest{
			AvailabilityStatus: "invalid-status",
		},
	)

	if !errors.Is(err, ErrInvalidAvailabilityStatus) {
		t.Fatalf(
			"expected ErrInvalidAvailabilityStatus, got %v",
			err,
		)
	}
}

func TestUpdate_UsesDefaultsAndUpdatesProfile(t *testing.T) {
	existing := &CleanerProfile{
		ID:              2,
		UserID:          7,
		IsVerified:      true,
		JobsCompleted:   12,
		JobsCancelled:   0,
		ResponseRate:    90,
		ServicesOffered: "Domestic",
	}

	repo := &MockRepository{
		profile: existing,
	}

	service := NewService(repo)

	result, err := service.Update(
		context.Background(),
		7,
		UpdateProfileRequest{
			Bio:             " Updated bio ",
			Location:        " Canary Wharf ",
			YearsExperience: 8,
			HourlyRate:      30,
			ServicesOffered: " airbnb ",
		},
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if repo.updatedProfile == nil {
		t.Fatal("expected profile update")
	}

	if repo.updatedProfile.Country != "UK" {
		t.Fatalf(
			"expected default country UK, got %q",
			repo.updatedProfile.Country,
		)
	}

	if repo.updatedProfile.City != "London" {
		t.Fatalf(
			"expected default city London, got %q",
			repo.updatedProfile.City,
		)
	}

	if repo.updatedProfile.AvailabilityStatus != "available_immediately" {
		t.Fatalf(
			"expected default availability, got %q",
			repo.updatedProfile.AvailabilityStatus,
		)
	}

	if repo.updatedProfile.TravelRadiusMiles != 10 {
		t.Fatalf(
			"expected default radius 10, got %d",
			repo.updatedProfile.TravelRadiusMiles,
		)
	}

	if repo.updatedProfile.ReliabilityScore == 0 {
		t.Fatal("expected reliability score to be calculated")
	}

	if repo.updatedProfile.Badge == "" {
		t.Fatal("expected badge to be calculated")
	}

	if result != existing {
		t.Fatal("expected refreshed profile")
	}
}

func TestUpdate_ProfileLookupError(t *testing.T) {
	lookupErr := errors.New("lookup failed")

	service := NewService(
		&MockRepository{
			getByUserIDErr: lookupErr,
		},
	)

	_, err := service.Update(
		context.Background(),
		1,
		UpdateProfileRequest{},
	)

	if !errors.Is(err, lookupErr) {
		t.Fatalf("expected lookup error, got %v", err)
	}
}

func TestUpdate_RepositoryUpdateError(t *testing.T) {
	updateErr := errors.New("update failed")

	service := NewService(
		&MockRepository{
			profile: &CleanerProfile{
				UserID: 1,
			},
			updateErr: updateErr,
		},
	)

	_, err := service.Update(
		context.Background(),
		1,
		UpdateProfileRequest{},
	)

	if !errors.Is(err, updateErr) {
		t.Fatalf("expected update error, got %v", err)
	}
}

func TestUpdateVerificationStatus_InvalidStatus(t *testing.T) {
	service := NewService(&MockRepository{})

	_, err := service.UpdateVerificationStatus(
		context.Background(),
		1,
		"invalid",
	)

	if !errors.Is(err, ErrInvalidVerificationState) {
		t.Fatalf(
			"expected ErrInvalidVerificationState, got %v",
			err,
		)
	}
}

func TestUpdateVerificationStatus_Verified(t *testing.T) {
	expected := &CleanerProfile{
		UserID:             5,
		IsVerified:         true,
		VerificationStatus: "verified",
	}

	repo := &MockRepository{
		profile: expected,
	}

	service := NewService(repo)

	result, err := service.UpdateVerificationStatus(
		context.Background(),
		5,
		" verified ",
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if repo.lastVerificationUserID != 5 {
		t.Fatalf(
			"expected user 5, got %d",
			repo.lastVerificationUserID,
		)
	}

	if repo.lastVerificationStatus != "verified" {
		t.Fatalf(
			"expected verified status, got %q",
			repo.lastVerificationStatus,
		)
	}

	if !repo.lastVerificationVerified {
		t.Fatal("expected verified=true")
	}

	if result != expected {
		t.Fatal("expected refreshed profile")
	}
}

func TestUpdateVerificationStatus_PendingIsNotVerified(t *testing.T) {
	repo := &MockRepository{
		profile: &CleanerProfile{
			UserID: 3,
		},
	}

	service := NewService(repo)

	_, err := service.UpdateVerificationStatus(
		context.Background(),
		3,
		"pending",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if repo.lastVerificationVerified {
		t.Fatal("expected pending status to set verified=false")
	}
}

func TestUpdateVerificationStatus_RepositoryError(t *testing.T) {
	updateErr := errors.New("verification update failed")

	service := NewService(
		&MockRepository{
			updateVerificationStatusErr: updateErr,
		},
	)

	_, err := service.UpdateVerificationStatus(
		context.Background(),
		1,
		"verified",
	)

	if !errors.Is(err, updateErr) {
		t.Fatalf(
			"expected verification update error, got %v",
			err,
		)
	}
}

func TestSearch_SortsByMatchScore(t *testing.T) {
	repo := &MockRepository{
		profiles: []CleanerProfile{
			{
				ID:                 1,
				City:               "London",
				Region:             "East",
				PostcodeArea:       "E14",
				AvailabilityStatus: "weekends_only",
				ServicesOffered:    "Domestic cleaning",
				YearsExperience:    1,
				HourlyRate:         25,
			},
			{
				ID:                 2,
				City:               "London",
				Region:             "East",
				PostcodeArea:       "E14",
				AvailabilityStatus: "weekends_only",
				ServicesOffered:    "Domestic and Airbnb",
				YearsExperience:    6,
				HourlyRate:         20,
				IsVerified:         true,
			},
		},
	}

	service := NewService(repo)

	results, err := service.Search(
		context.Background(),
		SearchProfilesRequest{
			City:               " London ",
			Region:             " East ",
			PostcodeArea:       " E14 ",
			AvailabilityStatus: " weekends_only ",
			ServicesOffered:    " airbnb ",
			MaxHourlyRate:      25,
		},
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 2 {
		t.Fatalf(
			"expected 2 results, got %d",
			len(results),
		)
	}

	if results[0].ID != 2 {
		t.Fatalf(
			"expected best matching profile first, got ID %d",
			results[0].ID,
		)
	}

	if results[0].MatchScore <= results[1].MatchScore {
		t.Fatal("expected first result to have higher score")
	}
}

func TestSearch_RepositoryError(t *testing.T) {
	searchErr := errors.New("search failed")

	service := NewService(
		&MockRepository{
			searchErr: searchErr,
		},
	)

	_, err := service.Search(
		context.Background(),
		SearchProfilesRequest{},
	)

	if !errors.Is(err, searchErr) {
		t.Fatalf(
			"expected search error, got %v",
			err,
		)
	}
}

func TestCalculateMatchScore(t *testing.T) {
	verified := true

	profile := CleanerProfile{
		PostcodeArea:       "E14",
		Region:             "East London",
		City:               "London",
		AvailabilityStatus: "weekends_only",
		ServicesOffered:    "Domestic and Airbnb cleaning",
		IsVerified:         true,
		YearsExperience:    5,
		HourlyRate:         20,
	}

	req := SearchProfilesRequest{
		PostcodeArea:       "e14",
		Region:             "east london",
		City:               "london",
		AvailabilityStatus: "weekends_only",
		ServicesOffered:    "airbnb",
		MaxHourlyRate:      25,
		IsVerified:         &verified,
	}

	score := calculateMatchScore(profile, req)

	if score != 120 {
		t.Fatalf(
			"expected match score 120, got %d",
			score,
		)
	}
}

func TestCalculateReliabilityScore_CappedAt100(t *testing.T) {
	profile := &CleanerProfile{
		IsVerified:       true,
		JobsCompleted:    20,
		JobsCancelled:    0,
		ResponseRate:     95,
		ReliabilityScore: 0,
	}

	score := calculateReliabilityScore(profile)

	if score != 100 {
		t.Fatalf("expected score 100, got %d", score)
	}
}

func TestCalculateCleanerBadge(t *testing.T) {
	tests := []struct {
		name     string
		profile  CleanerProfile
		expected string
	}{
		{
			name: "elite",
			profile: CleanerProfile{
				JobsCompleted:    100,
				ReliabilityScore: 90,
			},
			expected: "Elite Cleaner",
		},
		{
			name: "trusted",
			profile: CleanerProfile{
				JobsCompleted:    25,
				ReliabilityScore: 80,
			},
			expected: "Trusted Cleaner",
		},
		{
			name: "fast responder",
			profile: CleanerProfile{
				ResponseRate: 95,
			},
			expected: "Fast Responder",
		},
		{
			name: "airbnb specialist",
			profile: CleanerProfile{
				ServicesOffered: "Domestic, Airbnb",
			},
			expected: "Airbnb Specialist",
		},
		{
			name: "verified",
			profile: CleanerProfile{
				IsVerified: true,
			},
			expected: "Verified Cleaner",
		},
		{
			name:     "new cleaner",
			profile:  CleanerProfile{},
			expected: "New Cleaner",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := calculateCleanerBadge(&test.profile)

			if result != test.expected {
				t.Fatalf(
					"expected %q, got %q",
					test.expected,
					result,
				)
			}
		})
	}
}

func TestIncrementJobsCompleted_InvalidUser(t *testing.T) {
	service := NewService(&MockRepository{})

	err := service.IncrementJobsCompleted(
		context.Background(),
		0,
	)

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestIncrementJobsCompleted_Success(t *testing.T) {
	repo := &MockRepository{}
	service := NewService(repo)

	err := service.IncrementJobsCompleted(
		context.Background(),
		1,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !repo.incrementCompletedCalled {
		t.Fatal("expected completed counter update")
	}

	if !repo.recalculateCalled {
		t.Fatal("expected reputation recalculation")
	}
}

func TestIncrementJobsCompleted_IncrementError(t *testing.T) {
	incrementErr := errors.New("increment failed")

	repo := &MockRepository{
		incrementCompletedErr: incrementErr,
	}

	service := NewService(repo)

	err := service.IncrementJobsCompleted(
		context.Background(),
		1,
	)

	if !errors.Is(err, incrementErr) {
		t.Fatalf("expected increment error, got %v", err)
	}

	if repo.recalculateCalled {
		t.Fatal("recalculate should not run after increment failure")
	}
}

func TestIncrementJobsCompleted_RecalculateError(t *testing.T) {
	recalculateErr := errors.New("recalculate failed")

	service := NewService(
		&MockRepository{
			recalculateErr: recalculateErr,
		},
	)

	err := service.IncrementJobsCompleted(
		context.Background(),
		1,
	)

	if !errors.Is(err, recalculateErr) {
		t.Fatalf(
			"expected recalculate error, got %v",
			err,
		)
	}
}

func TestIncrementJobsCancelled_InvalidUser(t *testing.T) {
	service := NewService(&MockRepository{})

	err := service.IncrementJobsCancelled(
		context.Background(),
		0,
	)

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestIncrementJobsCancelled_Success(t *testing.T) {
	repo := &MockRepository{}
	service := NewService(repo)

	err := service.IncrementJobsCancelled(
		context.Background(),
		1,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !repo.incrementCancelledCalled {
		t.Fatal("expected cancelled counter update")
	}

	if !repo.recalculateCalled {
		t.Fatal("expected reputation recalculation")
	}
}

func TestIncrementJobsCancelled_IncrementError(t *testing.T) {
	incrementErr := errors.New("increment failed")

	repo := &MockRepository{
		incrementCancelledErr: incrementErr,
	}

	service := NewService(repo)

	err := service.IncrementJobsCancelled(
		context.Background(),
		1,
	)

	if !errors.Is(err, incrementErr) {
		t.Fatalf("expected increment error, got %v", err)
	}

	if repo.recalculateCalled {
		t.Fatal("recalculate should not run")
	}
}

func TestIncrementJobsCancelled_RecalculateError(t *testing.T) {
	recalculateErr := errors.New("recalculate failed")

	service := NewService(
		&MockRepository{
			recalculateErr: recalculateErr,
		},
	)

	err := service.IncrementJobsCancelled(
		context.Background(),
		1,
	)

	if !errors.Is(err, recalculateErr) {
		t.Fatalf(
			"expected recalculate error, got %v",
			err,
		)
	}
}

func TestGetCompletedJobs(t *testing.T) {
	now := time.Now()

	repo := &MockRepository{
		history: []JobHistoryItem{
			{
				ApplicationID: 1,
				JobID:         20,
				Status:        "completed",
				CreatedAt:     now,
			},
		},
	}

	service := NewService(repo)

	history, err := service.GetCompletedJobs(
		context.Background(),
		10,
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(history) != 1 {
		t.Fatalf(
			"expected one history item, got %d",
			len(history),
		)
	}

	_, err = service.GetCompletedJobs(
		context.Background(),
		0,
	)

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestGetCancelledJobs(t *testing.T) {
	repo := &MockRepository{
		history: []JobHistoryItem{
			{
				ApplicationID: 1,
				Status:        "cancelled",
			},
		},
	}

	service := NewService(repo)

	history, err := service.GetCancelledJobs(
		context.Background(),
		10,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(history) != 1 {
		t.Fatalf(
			"expected one item, got %d",
			len(history),
		)
	}

	_, err = service.GetCancelledJobs(
		context.Background(),
		0,
	)

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestGetFullHistory(t *testing.T) {
	repo := &MockRepository{
		history: []JobHistoryItem{
			{
				ApplicationID: 1,
			},
		},
	}

	service := NewService(repo)

	history, err := service.GetFullHistory(
		context.Background(),
		10,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(history) != 1 {
		t.Fatalf(
			"expected one item, got %d",
			len(history),
		)
	}

	_, err = service.GetFullHistory(
		context.Background(),
		0,
	)

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

var _ Repository = (*MockRepository)(nil)

func TestCalculateProfileStrength_EmptyProfile(t *testing.T) {
	profile := &CleanerProfile{}

	result := calculateProfilesStrength(
		profile,
		0,
	)

	if result.Percentage != 0 {
		t.Fatalf(
			"expected percentage 0, got %d",
			result.Percentage,
		)
	}

	if result.CompletedItems != 0 {
		t.Fatalf(
			"expected 0 completed items, got %d",
			result.CompletedItems,
		)
	}

	if result.TotalItems != 9 {
		t.Fatalf(
			"expected 9 total items, got %d",
			result.TotalItems,
		)
	}

	if result.IsComplete {
		t.Fatal("expected profile to be incomplete")
	}

	if result.NextAction != "Add your bio" {
		t.Fatalf(
			"expected next action %q, got %q",
			"Add your bio",
			result.NextAction,
		)
	}
}

func TestCalculateProfileStrength_PartialProfile(t *testing.T) {
	profile := &CleanerProfile{
		Bio:                "Experienced cleaner",
		Location:           "East London",
		ServicesOffered:    "Domestic and Airbnb",
		AvailabilityStatus: "available_immediately",
		HourlyRate:         25,
	}

	result := calculateProfilesStrength(
		profile,
		0,
	)

	if result.Percentage != 60 {
		t.Fatalf(
			"expected percentage 60, got %d",
			result.Percentage,
		)
	}

	if result.CompletedItems != 5 {
		t.Fatalf(
			"expected 5 completed items, got %d",
			result.CompletedItems,
		)
	}

	if result.TotalItems != 9 {
		t.Fatalf(
			"expected 9 total items, got %d",
			result.TotalItems,
		)
	}

	if result.NextAction != "Add your experience" {
		t.Fatalf(
			"expected next action %q, got %q",
			"Add your experience",
			result.NextAction,
		)
	}

	if result.IsComplete {
		t.Fatal("expected profile to be incomplete")
	}
}

func TestCalculateProfileStrength_CompleteProfile(t *testing.T) {
	profile := &CleanerProfile{
		Bio:                "Experienced cleaner",
		Location:           "East London",
		ServicesOffered:    "Domestic, Airbnb and end of tenancy",
		AvailabilityStatus: "available_immediately",
		HourlyRate:         25,
		YearsExperience:    5,
		TravelRadiusMiles:  20,
		IsVerified:         true,
		VerificationStatus: "verified",
	}

	result := calculateProfilesStrength(
		profile,
		3,
	)

	if result.Percentage != 100 {
		t.Fatalf(
			"expected percentage 100, got %d",
			result.Percentage,
		)
	}

	if result.CompletedItems != 9 {
		t.Fatalf(
			"expected 9 completed items, got %d",
			result.CompletedItems,
		)
	}

	if result.TotalItems != 9 {
		t.Fatalf(
			"expected 9 total items, got %d",
			result.TotalItems,
		)
	}

	if !result.IsComplete {
		t.Fatal("expected profile to be complete")
	}

	if result.NextAction != "Your profile is complete" {
		t.Fatalf(
			"unexpected next action %q",
			result.NextAction,
		)
	}
}

func TestCalculateProfileStrength_VerificationRequiresVerifiedState(
	t *testing.T,
) {
	profile := &CleanerProfile{
		IsVerified:         true,
		VerificationStatus: "pending",
	}

	result := calculateProfilesStrength(
		profile,
		0,
	)

	var verificationItem *ProfileStrengthItem

	for i := range result.Items {
		if result.Items[i].Code == "verification" {
			verificationItem = &result.Items[i]
			break
		}
	}

	if verificationItem == nil {
		t.Fatal("expected verification item")
	}

	if verificationItem.Completed {
		t.Fatal(
			"expected pending verification to be incomplete",
		)
	}
}

func TestGetMyProfileStrength_InvalidUserID(t *testing.T) {
	service := NewService(
		&MockRepository{},
	)

	result, err := service.GetMyProfileStrength(
		context.Background(),
		0,
	)

	if result != nil {
		t.Fatal("expected nil result")
	}

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf(
			"expected ErrInvalidInput, got %v",
			err,
		)
	}
}

func TestGetMyProfileStrength_Success(t *testing.T) {
	repo := &MockRepository{
		profile: &CleanerProfile{
			UserID:             10,
			Bio:                "Experienced cleaner",
			Location:           "East London",
			ServicesOffered:    "Domestic cleaning",
			AvailabilityStatus: "available_immediately",
			HourlyRate:         25,
			YearsExperience:    5,
			TravelRadiusMiles:  20,
			IsVerified:         true,
			VerificationStatus: "verified",
		},
	}

	service := NewService(repo)

	service.SetPortfolioCounter(
		&mockPortfolioCounter{
			count: 3,
		},
	)

	result, err := service.GetMyProfileStrength(
		context.Background(),
		10,
	)
	if err != nil {
		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}

	if result == nil {
		t.Fatal("expected profile strength")
	}

	if result.Percentage != 100 {
		t.Fatalf(
			"expected percentage 100, got %d",
			result.Percentage,
		)
	}

	if !result.IsComplete {
		t.Fatal(
			"expected profile to be complete",
		)
	}
}

func TestGetMyProfileStrength_RepositoryError(t *testing.T) {
	repositoryErr := errors.New(
		"profile lookup failed",
	)

	service := NewService(
		&MockRepository{
			getByUserIDErr: repositoryErr,
		},
	)

	result, err := service.GetMyProfileStrength(
		context.Background(),
		10,
	)

	if result != nil {
		t.Fatal("expected nil result")
	}

	if !errors.Is(err, repositoryErr) {
		t.Fatalf(
			"expected repository error, got %v",
			err,
		)
	}
}

func TestCalculateProfileStrength_PortfolioRequirement(t *testing.T) {
	profile := &CleanerProfile{
		Bio:                "Experienced cleaner",
		Location:           "London",
		ServicesOffered:    "Domestic cleaning",
		AvailabilityStatus: "available_immediately",
		HourlyRate:         20,
		YearsExperience:    3,
		TravelRadiusMiles:  10,
		IsVerified:         true,
		VerificationStatus: "verified",
	}

	tests := []struct {
		name             string
		portfolioPhotos  int
		expectedPercent  int
		expectedComplete bool
	}{
		{
			name:             "zero photos",
			portfolioPhotos:  0,
			expectedPercent:  90,
			expectedComplete: false,
		},
		{
			name:             "one photo",
			portfolioPhotos:  1,
			expectedPercent:  90,
			expectedComplete: false,
		},
		{
			name:             "two photos",
			portfolioPhotos:  2,
			expectedPercent:  90,
			expectedComplete: false,
		},
		{
			name:             "three photos",
			portfolioPhotos:  3,
			expectedPercent:  100,
			expectedComplete: true,
		},
		{
			name:             "six photos",
			portfolioPhotos:  6,
			expectedPercent:  100,
			expectedComplete: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			strength := calculateProfilesStrength(
				profile,
				tt.portfolioPhotos,
			)

			if strength.Percentage != tt.expectedPercent {
				t.Fatalf(
					"expected percentage %d, got %d",
					tt.expectedPercent,
					strength.Percentage,
				)
			}

			if strength.IsComplete != tt.expectedComplete {
				t.Fatalf(
					"expected complete=%v, got %v",
					tt.expectedComplete,
					strength.IsComplete,
				)
			}

			var portfolioItem *ProfileStrengthItem

			for i := range strength.Items {
				if strength.Items[i].Code == "portfolio" {
					portfolioItem = &strength.Items[i]
					break
				}
			}

			if portfolioItem == nil {
				t.Fatal("expected portfolio item")
			}

			expectedPortfolioComplete :=
				tt.portfolioPhotos >= 3

			if portfolioItem.Completed != expectedPortfolioComplete {
				t.Fatalf(
					"expected portfolio completed=%v, got %v",
					expectedPortfolioComplete,
					portfolioItem.Completed,
				)
			}
		})
	}
}


func TestGetNewOnJobiraStatus_NewCleaner(t *testing.T) {
	repo := &MockRepository{
		profile: &CleanerProfile{
			UserID:        10,
			JobsCompleted: 0,
			CreatedAt:     time.Now().AddDate(0, 0, -5),
		},
	}

	service := NewService(repo)

	result, err := service.GetNewOnJobiraStatus(
		context.Background(),
		10,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected result")
	}

	if !result.IsNew {
		t.Fatal("expected cleaner to be new on Jobira")
	}

	if result.Label != "New on Jobira" {
		t.Fatalf(
			"expected label %q, got %q",
			"New on Jobira",
			result.Label,
		)
	}

	if result.JobsCompleted != 0 {
		t.Fatalf(
			"expected 0 completed jobs, got %d",
			result.JobsCompleted,
		)
	}

	if result.DaysOnJobira < 4 ||
		result.DaysOnJobira > 5 {
		t.Fatalf(
			"expected about 5 days on Jobira, got %d",
			result.DaysOnJobira,
		)
	}
}

func TestGetNewOnJobiraStatus_TwoCompletedJobsStillNew(
	t *testing.T,
) {
	repo := &MockRepository{
		profile: &CleanerProfile{
			UserID:        10,
			JobsCompleted: 2,
			CreatedAt:     time.Now().AddDate(0, 0, -10),
		},
	}

	service := NewService(repo)

	result, err := service.GetNewOnJobiraStatus(
		context.Background(),
		10,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.IsNew {
		t.Fatal(
			"expected cleaner with 2 completed jobs to remain new",
		)
	}
}

func TestGetNewOnJobiraStatus_ThreeCompletedJobsNotNew(
	t *testing.T,
) {
	repo := &MockRepository{
		profile: &CleanerProfile{
			UserID:        10,
			JobsCompleted: 3,
			CreatedAt:     time.Now().AddDate(0, 0, -5),
		},
	}

	service := NewService(repo)

	result, err := service.GetNewOnJobiraStatus(
		context.Background(),
		10,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.IsNew {
		t.Fatal(
			"expected cleaner with 3 completed jobs not to be new",
		)
	}

	if result.Label != "" {
		t.Fatalf(
			"expected empty label, got %q",
			result.Label,
		)
	}
}

func TestGetNewOnJobiraStatus_OverThirtyDaysNotNew(
	t *testing.T,
) {
	repo := &MockRepository{
		profile: &CleanerProfile{
			UserID:        10,
			JobsCompleted: 0,
			CreatedAt:     time.Now().AddDate(0, 0, -31),
		},
	}

	service := NewService(repo)

	result, err := service.GetNewOnJobiraStatus(
		context.Background(),
		10,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.IsNew {
		t.Fatal(
			"expected established cleaner not to be new",
		)
	}

	if result.Label != "" {
		t.Fatalf(
			"expected empty label, got %q",
			result.Label,
		)
	}
}

func TestGetNewOnJobiraStatus_InvalidUserID(t *testing.T) {
	service := NewService(
		&MockRepository{},
	)

	result, err := service.GetNewOnJobiraStatus(
		context.Background(),
		0,
	)

	if result != nil {
		t.Fatal("expected nil result")
	}

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf(
			"expected ErrInvalidInput, got %v",
			err,
		)
	}
}

func TestGetNewOnJobiraStatus_ProfileNotFound(t *testing.T) {
	service := NewService(
		&MockRepository{
			getByUserIDErr: ErrProfileNotFound,
		},
	)

	result, err := service.GetNewOnJobiraStatus(
		context.Background(),
		10,
	)

	if result != nil {
		t.Fatal("expected nil result")
	}

	if !errors.Is(err, ErrProfileNotFound) {
		t.Fatalf(
			"expected ErrProfileNotFound, got %v",
			err,
		)
	}
}

func TestGetNewOnJobiraStatus_RepositoryError(t *testing.T) {
	repositoryErr := errors.New(
		"profile lookup failed",
	)

	service := NewService(
		&MockRepository{
			getByUserIDErr: repositoryErr,
		},
	)

	result, err := service.GetNewOnJobiraStatus(
		context.Background(),
		10,
	)

	if result != nil {
		t.Fatal("expected nil result")
	}

	if !errors.Is(err, repositoryErr) {
		t.Fatalf(
			"expected repository error, got %v",
			err,
		)
	}
}

func TestCalculateNewOnJobiraStatus_ExactlyThirtyDays(
	t *testing.T,
) {
	now := time.Date(
		2026,
		time.August,
		29,
		12,
		0,
		0,
		0,
		time.UTC,
	)

	profile := &CleanerProfile{
		UserID:        10,
		JobsCompleted: 0,
		CreatedAt:     now.AddDate(0, 0, -30),
	}

	result := calculateNewOnJobiraStatus(
		profile,
		now,
	)

	if result.IsNew {
		t.Fatal(
			"expected exactly 30 days not to be new",
		)
	}

	if result.DaysOnJobira != 30 {
		t.Fatalf(
			"expected 30 days, got %d",
			result.DaysOnJobira,
		)
	}
}

func TestCalculateNewOnJobiraStatus_FutureCreatedAt(
	t *testing.T,
) {
	now := time.Date(
		2026,
		time.August,
		29,
		12,
		0,
		0,
		0,
		time.UTC,
	)

	profile := &CleanerProfile{
		UserID:        10,
		JobsCompleted: 0,
		CreatedAt:     now.Add(24 * time.Hour),
	}

	result := calculateNewOnJobiraStatus(
		profile,
		now,
	)

	if result.DaysOnJobira != 0 {
		t.Fatalf(
			"expected future timestamp to clamp to 0 days, got %d",
			result.DaysOnJobira,
		)
	}

	if !result.IsNew {
		t.Fatal(
			"expected future timestamp protection to remain new",
		)
	}
}

func TestCalculateNewOnJobiraStatus_NilProfile(
	t *testing.T,
) {
	result := calculateNewOnJobiraStatus(
		nil,
		time.Now(),
	)

	if result.IsNew {
		t.Fatal("expected nil profile not to be new")
	}

	if result.Label != "" {
		t.Fatalf(
			"expected empty label, got %q",
			result.Label,
		)
	}
}

func TestCalculateNewOnJobiraStatus_ZeroCreatedAt(
	t *testing.T,
) {
	result := calculateNewOnJobiraStatus(
		&CleanerProfile{
			UserID: 10,
		},
		time.Now(),
	)

	if result.IsNew {
		t.Fatal(
			"expected zero created_at not to be new",
		)
	}

	if result.DaysOnJobira != 0 {
		t.Fatalf(
			"expected 0 days, got %d",
			result.DaysOnJobira,
		)
	}
}

