package cleaningteam

import (
	"context"
	"errors"
	"testing"
	"time"
)

type mockRepository struct {
	listTeamMembersFn func(context.Context, uint) ([]TeamMemberData, error)
}

func (m *mockRepository) ListTeamMembers(
	ctx context.Context,
	clientID uint,
) ([]TeamMemberData, error) {
	if m.listTeamMembersFn != nil {
		return m.listTeamMembersFn(
			ctx,
			clientID,
		)
	}

	return nil, nil
}

func TestService_GetMyCleaningTeam_Success(
	t *testing.T,
) {
	lastBookingID := uint(55)
	lastBookedAt := time.Now().UTC()

	repo := &mockRepository{
		listTeamMembersFn: func(
			_ context.Context,
			clientID uint,
		) ([]TeamMemberData, error) {
			if clientID != 5 {
				t.Fatalf(
					"expected client ID 5, got %d",
					clientID,
				)
			}

			return []TeamMemberData{
				{
					CleanerID:   20,
					CleanerName: "Maria",

					Location: "East London",

					ServicesOffered: "housekeeping, laundry, ironing",

					LastJobType: "housekeeping",

					AvailabilityStatus: "available",

					IsVerified: true,

					IsFavorite:  true,
					IsPreferred: true,

					PrivateNote: "Great housekeeper",

					CompletedJobsTogether: 6,

					LastBookingID: &lastBookingID,

					LastBookedAt: &lastBookedAt,

					AverageRating: 4.9,

					TotalReviews: 30,

					ReliabilityScore: 94,

					Badge: "Elite Cleaner",
				},
				{
					CleanerID:   21,
					CleanerName: "Sarah",

					ServicesOffered: "domestic cleaning",

					IsFavorite: true,

					CompletedJobsTogether: 0,

					ReliabilityScore: 80,
				},
			}, nil
		},
	}

	service := NewService(repo)

	result, err := service.GetMyCleaningTeam(
		context.Background(),
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

	if result.ClientID != 5 {
		t.Fatalf(
			"expected client ID 5, got %d",
			result.ClientID,
		)
	}

	if result.TotalMembers != 2 {
		t.Fatalf(
			"expected 2 members, got %d",
			result.TotalMembers,
		)
	}

	if result.PreferredCount != 1 {
		t.Fatalf(
			"expected preferred count 1, got %d",
			result.PreferredCount,
		)
	}

	if result.FavoriteCount != 2 {
		t.Fatalf(
			"expected favorite count 2, got %d",
			result.FavoriteCount,
		)
	}

	if result.HouseKeeperCount != 1 {
		t.Fatalf(
			"expected housekeeper count 1, got %d",
			result.HouseKeeperCount,
		)
	}

	first := result.Members[0]

	if first.CleanerID != 20 {
		t.Fatalf(
			"expected Maria first, got cleaner %d",
			first.CleanerID,
		)
	}

	if first.ServiceRole != "Housekeeper" {
		t.Fatalf(
			"expected Housekeeper role, got %q",
			first.ServiceRole,
		)
	}

	if first.RelationshipType != "preferred" {
		t.Fatalf(
			"expected preferred relationship, got %q",
			first.RelationshipType,
		)
	}

	if !first.CanRebook {
		t.Fatal(
			"expected preferred previously-booked cleaner to be rebookable",
		)
	}

	second := result.Members[1]

	if second.RelationshipType != "shortlisted" {
		t.Fatalf(
			"expected shortlisted relationship, got %q",
			second.RelationshipType,
		)
	}

	if second.CanRebook {
		t.Fatal(
			"did not expect never-booked cleaner to be rebookable",
		)
	}
}

func TestService_GetMyCleaningTeam_InvalidClientID(
	t *testing.T,
) {
	service := NewService(
		&mockRepository{},
	)

	result, err := service.GetMyCleaningTeam(
		context.Background(),
		0,
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

func TestService_GetMyCleaningTeam_RepositoryError(
	t *testing.T,
) {
	expectedErr := errors.New(
		"repository failed",
	)

	repo := &mockRepository{
		listTeamMembersFn: func(
			context.Context,
			uint,
		) ([]TeamMemberData, error) {
			return nil, expectedErr
		},
	}

	service := NewService(repo)

	result, err := service.GetMyCleaningTeam(
		context.Background(),
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

func TestService_GetMyCleaningTeam_Empty(
	t *testing.T,
) {
	repo := &mockRepository{
		listTeamMembersFn: func(
			context.Context,
			uint,
		) ([]TeamMemberData, error) {
			return []TeamMemberData{}, nil
		},
	}

	service := NewService(repo)

	result, err := service.GetMyCleaningTeam(
		context.Background(),
		5,
	)
	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}

	if result.TotalMembers != 0 {
		t.Fatalf(
			"expected zero members, got %d",
			result.TotalMembers,
		)
	}

	if len(result.Members) != 0 {
		t.Fatalf(
			"expected empty members, got %d",
			len(result.Members),
		)
	}
}

func TestDetermineServiceRole(
	t *testing.T,
) {
	tests := []struct {
		lastJobType string
		services    []string
		expected    string
	}{
		{
			"housekeeping",
			[]string{"domestic cleaning"},
			"Housekeeper",
		},
		{
			"",
			[]string{"housekeeping", "laundry"},
			"Housekeeper",
		},
		{
			"airbnb",
			[]string{"domestic cleaning"},
			"Airbnb Cleaner",
		},
		{
			"end_of_tenancy",
			nil,
			"End of Tenancy Cleaner",
		},
		{
			"deep_clean",
			nil,
			"Deep Clean Specialist",
		},
		{
			"commercial",
			nil,
			"Commercial Cleaner",
		},
		{
			"",
			[]string{"domestic cleaning"},
			"Cleaner",
		},
	}

	for _, tt := range tests {
		actual := determineServiceRole(
			tt.lastJobType,
			tt.services,
		)

		if actual != tt.expected {
			t.Fatalf(
				"expected %q, got %q",
				tt.expected,
				actual,
			)
		}
	}
}

func TestParseServices(
	t *testing.T,
) {
	result := parseServices(
		"Housekeeping, laundry; ironing | domestic_cleaning, housekeeping",
	)

	expected := []string{
		"housekeeping",
		"laundry",
		"ironing",
		"domestic cleaning",
	}

	if len(result) != len(expected) {
		t.Fatalf(
			"expected %d services, got %d",
			len(expected),
			len(result),
		)
	}

	for i := range expected {
		if result[i] != expected[i] {
			t.Fatalf(
				"expected service %q, got %q",
				expected[i],
				result[i],
			)
		}
	}
}

func TestDetermineRelationshipType(
	t *testing.T,
) {
	tests := []struct {
		data     TeamMemberData
		expected string
	}{
		{
			TeamMemberData{
				IsPreferred: true,
				IsFavorite:  true,
			},
			"preferred",
		},
		{
			TeamMemberData{
				CompletedJobsTogether: 2,
			},
			"previously_booked",
		},
		{
			TeamMemberData{
				IsFavorite: true,
			},
			"shortlisted",
		},
		{
			TeamMemberData{},
			"saved",
		},
	}

	for _, tt := range tests {
		actual := determineRelationshipType(
			tt.data,
		)

		if actual != tt.expected {
			t.Fatalf(
				"expected %q, got %q",
				tt.expected,
				actual,
			)
		}
	}
}
