package availablenow

import (
	"context"
	"errors"
	"testing"
	"time"
)

type mockRepository struct {
	upsertFn func(
		context.Context,
		*CleanerAvailableNow,
	) error

	getByCleanerIDFn func(
		context.Context,
		uint,
	) (*CleanerAvailableNow, error)

	disableFn func(
		context.Context,
		uint,
	) error

	listAvailableCleanersFn func(
		context.Context,
		AvailableNowSearchRequest,
		time.Time,
	) ([]AvailableCleaner, error)
}

func (m *mockRepository) Upsert(
	ctx context.Context,
	availability *CleanerAvailableNow,
) error {
	if m.upsertFn != nil {
		return m.upsertFn(
			ctx,
			availability,
		)
	}

	return nil
}

func (m *mockRepository) GetByCleanerID(
	ctx context.Context,
	cleanerID uint,
) (*CleanerAvailableNow, error) {
	if m.getByCleanerIDFn != nil {
		return m.getByCleanerIDFn(
			ctx,
			cleanerID,
		)
	}

	return nil,
		ErrAvailableNowNotFound
}

func (m *mockRepository) Disable(
	ctx context.Context,
	cleanerID uint,
) error {
	if m.disableFn != nil {
		return m.disableFn(
			ctx,
			cleanerID,
		)
	}

	return nil
}

func (m *mockRepository) ListAvailableCleaners(
	ctx context.Context,
	search AvailableNowSearchRequest,
	now time.Time,
) ([]AvailableCleaner, error) {
	if m.listAvailableCleanersFn != nil {
		return m.listAvailableCleanersFn(
			ctx,
			search,
			now,
		)
	}

	return []AvailableCleaner{},
		nil
}

func TestService_SetAvailableNow_Success(t *testing.T) {
	repo := &mockRepository{
		upsertFn: func(
			ctx context.Context,
			availability *CleanerAvailableNow,
		) error {
			if availability.CleanerID != 10 {
				t.Fatalf(
					"expected cleaner ID 10, got %d",
					availability.CleanerID,
				)
			}

			if !availability.IsAvailable {
				t.Fatal(
					"expected availability enabled",
				)
			}

			if availability.Location !=
				"Stratford" {
				t.Fatalf(
					"expected Stratford, got %q",
					availability.Location,
				)
			}

			availability.ID = 20

			return nil
		},
	}

	service := NewService(repo)

	result, err :=
		service.SetAvailableNow(
			context.Background(),
			10,
			SetAvailableNowRequest{
				DurationMinutes:   180,
				Location:          " Stratford ",
				TravelRadiusMiles: 8,
				JobTypes: []string{
					" Domestic ",
					"airbnb",
					"domestic",
				},
			},
		)
	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}

	if result == nil {
		t.Fatal(
			"expected availability",
		)
	}

	if result.ID != 20 {
		t.Fatalf(
			"expected ID 20, got %d",
			result.ID,
		)
	}

	if len(result.JobTypes) != 2 {
		t.Fatalf(
			"expected 2 unique job types, got %d",
			len(result.JobTypes),
		)
	}

	if result.JobTypes[0] != "domestic" {
		t.Fatalf(
			"expected normalized domestic job type, got %q",
			result.JobTypes[0],
		)
	}
}

func TestService_SetAvailableNow_DefaultDuration(
	t *testing.T,
) {
	repo := &mockRepository{
		upsertFn: func(
			context.Context,
			*CleanerAvailableNow,
		) error {
			return nil
		},
	}

	service := NewService(repo)

	result, err :=
		service.SetAvailableNow(
			context.Background(),
			10,
			SetAvailableNowRequest{
				Location:          "Bow",
				TravelRadiusMiles: 5,
				JobTypes: []string{
					"domestic",
				},
			},
		)
	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}

	duration :=
		result.AvailableUntil.Sub(
			result.AvailableFrom,
		)

	if duration != 2*time.Hour {
		t.Fatalf(
			"expected 2 hour duration, got %v",
			duration,
		)
	}
}

func TestService_SetAvailableNow_InvalidCleanerID(
	t *testing.T,
) {
	service := NewService(
		&mockRepository{},
	)

	result, err :=
		service.SetAvailableNow(
			context.Background(),
			0,
			SetAvailableNowRequest{},
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

func TestService_SetAvailableNow_InvalidDuration(
	t *testing.T,
) {
	service := NewService(
		&mockRepository{},
	)

	result, err :=
		service.SetAvailableNow(
			context.Background(),
			10,
			SetAvailableNowRequest{
				DurationMinutes:   10,
				Location:          "Bow",
				TravelRadiusMiles: 5,
				JobTypes: []string{
					"domestic",
				},
			},
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

func TestService_SetAvailableNow_EmptyLocation(
	t *testing.T,
) {
	service := NewService(
		&mockRepository{},
	)

	_, err :=
		service.SetAvailableNow(
			context.Background(),
			10,
			SetAvailableNowRequest{
				DurationMinutes:   120,
				Location:          " ",
				TravelRadiusMiles: 5,
				JobTypes: []string{
					"domestic",
				},
			},
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
}

func TestService_GetStatus_Active(t *testing.T) {
	now := time.Now().UTC()

	repo := &mockRepository{
		getByCleanerIDFn: func(
			context.Context,
			uint,
		) (*CleanerAvailableNow, error) {
			return &CleanerAvailableNow{
				CleanerID:      10,
				IsAvailable:    true,
				AvailableFrom:  now.Add(-time.Hour),
				AvailableUntil: now.Add(time.Hour),
			}, nil
		},
	}

	service := NewService(repo)

	result, err :=
		service.GetStatus(
			context.Background(),
			10,
		)
	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}

	if !result.IsAvailable {
		t.Fatal(
			"expected active availability",
		)
	}

	if result.MinutesRemaining <= 0 {
		t.Fatalf(
			"expected remaining minutes, got %d",
			result.MinutesRemaining,
		)
	}
}

func TestService_GetStatus_NotFoundReturnsInactive(
	t *testing.T,
) {
	service := NewService(
		&mockRepository{},
	)

	result, err :=
		service.GetStatus(
			context.Background(),
			10,
		)
	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}

	if result.IsAvailable {
		t.Fatal(
			"expected unavailable status",
		)
	}
}

func TestService_Disable_Success(t *testing.T) {
	called := false

	repo := &mockRepository{
		disableFn: func(
			ctx context.Context,
			cleanerID uint,
		) error {
			called = true

			if cleanerID != 10 {
				t.Fatalf(
					"expected cleaner ID 10, got %d",
					cleanerID,
				)
			}

			return nil
		},
	}

	service := NewService(repo)

	err :=
		service.Disable(
			context.Background(),
			10,
		)
	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}

	if !called {
		t.Fatal(
			"expected repository Disable call",
		)
	}
}

func TestService_Search_Success(t *testing.T) {
	repo := &mockRepository{
		listAvailableCleanersFn: func(
			ctx context.Context,
			search AvailableNowSearchRequest,
			now time.Time,
		) ([]AvailableCleaner, error) {
			if search.JobType != "domestic" {
				t.Fatalf(
					"expected domestic, got %q",
					search.JobType,
				)
			}

			if search.Limit != 20 {
				t.Fatalf(
					"expected default limit 20, got %d",
					search.Limit,
				)
			}

			return []AvailableCleaner{
				{
					CleanerID: 10,
					FullName:  "Sarah Cleaner",
				},
			}, nil
		},
	}

	service := NewService(repo)

	results, err :=
		service.Search(
			context.Background(),
			AvailableNowSearchRequest{
				Location: " Stratford ",
				JobType:  " Domestic ",
			},
		)
	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}

	if len(results) != 1 {
		t.Fatalf(
			"expected 1 cleaner, got %d",
			len(results),
		)
	}
}

func TestService_Search_InvalidRating(
	t *testing.T,
) {
	service := NewService(
		&mockRepository{},
	)

	_, err :=
		service.Search(
			context.Background(),
			AvailableNowSearchRequest{
				MinimumRating: 6,
			},
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
}
