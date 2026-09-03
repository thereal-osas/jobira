package applicationtimeline

import (
	"context"
	"errors"
	"testing"
)

type mockRepository struct {
	createFn                  func(ctx context.Context, event *Event) error
	listByApplicationIDFn     func(ctx context.Context, applicationID uint) ([]Event, error)
	getCurrentStatusFn        func(ctx context.Context, applicationID uint) (string, error)
	getApplicationCleanerIDFn func(ctx context.Context, applicationID uint) (uint, error)
	getApplicationClientIDFn  func(ctx context.Context, applicationID uint) (uint, error)
}

func (m *mockRepository) Create(
	ctx context.Context,
	event *Event,
) error {
	if m.createFn != nil {
		return m.createFn(ctx, event)
	}

	return nil
}

func (m *mockRepository) ListByApplicationID(
	ctx context.Context,
	applicationID uint,
) ([]Event, error) {
	if m.listByApplicationIDFn != nil {
		return m.listByApplicationIDFn(ctx, applicationID)
	}

	return nil, nil
}

func (m *mockRepository) GetCurrentStatus(
	ctx context.Context,
	applicationID uint,
) (string, error) {
	if m.getCurrentStatusFn != nil {
		return m.getCurrentStatusFn(ctx, applicationID)
	}

	return "", nil
}

func (m *mockRepository) GetApplicationCleanerID(
	ctx context.Context,
	applicationID uint,
) (uint, error) {
	if m.getApplicationCleanerIDFn != nil {
		return m.getApplicationCleanerIDFn(ctx, applicationID)
	}

	return 0, nil
}

func (m *mockRepository) GetApplicationClientID(
	ctx context.Context,
	applicationID uint,
) (uint, error) {
	if m.getApplicationClientIDFn != nil {
		return m.getApplicationClientIDFn(ctx, applicationID)
	}

	return 0, nil
}

func TestService_Record_Success(t *testing.T) {
	repo := &mockRepository{
		createFn: func(
			ctx context.Context,
			event *Event,
		) error {
			if event.ApplicationID != 10 {
				t.Fatalf(
					"expected application id 10, got %d",
					event.ApplicationID,
				)
			}

			if event.Status != "shortlisted" {
				t.Fatalf(
					"expected shortlisted, got %s",
					event.Status,
				)
			}

			if event.ActorUserID == nil {
				t.Fatal("expected actor user id")
			}

			if *event.ActorUserID != 7 {
				t.Fatalf(
					"expected actor user id 7, got %d",
					*event.ActorUserID,
				)
			}

			if event.Note != "Application shortlisted" {
				t.Fatalf(
					"unexpected note: %s",
					event.Note,
				)
			}

			return nil
		},
	}

	service := NewService(repo)

	err := service.Record(
		context.Background(),
		10,
		"shortlisted",
		7,
		" Application shortlisted ",
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestService_Record_InvalidApplicationID(t *testing.T) {
	service := NewService(&mockRepository{})

	err := service.Record(
		context.Background(),
		0,
		"pending",
		7,
		"",
	)

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf(
			"expected ErrInvalidInput, got %v",
			err,
		)
	}
}

func TestService_Record_InvalidActorUserID(t *testing.T) {
	service := NewService(&mockRepository{})

	err := service.Record(
		context.Background(),
		10,
		"pending",
		0,
		"",
	)

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf(
			"expected ErrInvalidInput, got %v",
			err,
		)
	}
}

func TestService_Record_InvalidStatus(t *testing.T) {
	service := NewService(&mockRepository{})

	err := service.Record(
		context.Background(),
		10,
		"unknown",
		7,
		"",
	)

	if !errors.Is(err, ErrInvalidStatus) {
		t.Fatalf(
			"expected ErrInvalidStatus, got %v",
			err,
		)
	}
}

func TestService_Record_RepositoryError(t *testing.T) {
	expectedErr := errors.New("database error")

	repo := &mockRepository{
		createFn: func(
			ctx context.Context,
			event *Event,
		) error {
			return expectedErr
		},
	}

	service := NewService(repo)

	err := service.Record(
		context.Background(),
		10,
		"pending",
		7,
		"",
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected repository error, got %v",
			err,
		)
	}
}

func TestService_GetTimeline_CleanerSuccess(t *testing.T) {
	repo := successfulTimelineRepository()

	service := NewService(repo)

	result, err := service.GetTimeline(
		context.Background(),
		10,
		20,
		"cleaner",
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ApplicationID != 10 {
		t.Fatalf(
			"expected application id 10, got %d",
			result.ApplicationID,
		)
	}

	if result.CurrentStatus != "shortlisted" {
		t.Fatalf(
			"expected shortlisted, got %s",
			result.CurrentStatus,
		)
	}

	if len(result.Events) != 2 {
		t.Fatalf(
			"expected 2 events, got %d",
			len(result.Events),
		)
	}

	if result.Events[0].Status != "pending" {
		t.Fatalf(
			"expected first event pending, got %s",
			result.Events[0].Status,
		)
	}

	if result.Events[1].Status != "shortlisted" {
		t.Fatalf(
			"expected second event shortlisted, got %s",
			result.Events[1].Status,
		)
	}
}

func TestService_GetTimeline_ClientSuccess(t *testing.T) {
	repo := successfulTimelineRepository()

	service := NewService(repo)

	result, err := service.GetTimeline(
		context.Background(),
		10,
		30,
		"client",
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected timeline")
	}
}

func TestService_GetTimeline_AdminSuccess(t *testing.T) {
	repo := successfulTimelineRepository()

	service := NewService(repo)

	result, err := service.GetTimeline(
		context.Background(),
		10,
		999,
		"admin",
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected timeline")
	}
}

func TestService_GetTimeline_InvalidInput(t *testing.T) {
	service := NewService(&mockRepository{})

	_, err := service.GetTimeline(
		context.Background(),
		0,
		20,
		"cleaner",
	)

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf(
			"expected ErrInvalidInput, got %v",
			err,
		)
	}
}

func TestService_GetTimeline_CleanerLookupError(t *testing.T) {
	expectedErr := errors.New("cleaner lookup error")

	repo := &mockRepository{
		getApplicationCleanerIDFn: func(
			ctx context.Context,
			applicationID uint,
		) (uint, error) {
			return 0, expectedErr
		},
	}

	service := NewService(repo)

	_, err := service.GetTimeline(
		context.Background(),
		10,
		20,
		"cleaner",
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected cleaner lookup error, got %v",
			err,
		)
	}
}

func TestService_GetTimeline_ClientLookupError(t *testing.T) {
	expectedErr := errors.New("client lookup error")

	repo := &mockRepository{
		getApplicationCleanerIDFn: func(
			ctx context.Context,
			applicationID uint,
		) (uint, error) {
			return 20, nil
		},
		getApplicationClientIDFn: func(
			ctx context.Context,
			applicationID uint,
		) (uint, error) {
			return 0, expectedErr
		},
	}

	service := NewService(repo)

	_, err := service.GetTimeline(
		context.Background(),
		10,
		20,
		"cleaner",
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected client lookup error, got %v",
			err,
		)
	}
}

func TestService_GetTimeline_Forbidden(t *testing.T) {
	repo := &mockRepository{
		getApplicationCleanerIDFn: func(
			ctx context.Context,
			applicationID uint,
		) (uint, error) {
			return 20, nil
		},
		getApplicationClientIDFn: func(
			ctx context.Context,
			applicationID uint,
		) (uint, error) {
			return 30, nil
		},
	}

	service := NewService(repo)

	_, err := service.GetTimeline(
		context.Background(),
		10,
		99,
		"user",
	)

	if !errors.Is(err, ErrForbidden) {
		t.Fatalf(
			"expected ErrForbidden, got %v",
			err,
		)
	}
}

func TestService_GetTimeline_StatusLookupError(t *testing.T) {
	expectedErr := errors.New("status lookup error")

	repo := &mockRepository{
		getApplicationCleanerIDFn: func(
			ctx context.Context,
			applicationID uint,
		) (uint, error) {
			return 20, nil
		},
		getApplicationClientIDFn: func(
			ctx context.Context,
			applicationID uint,
		) (uint, error) {
			return 30, nil
		},
		getCurrentStatusFn: func(
			ctx context.Context,
			applicationID uint,
		) (string, error) {
			return "", expectedErr
		},
	}

	service := NewService(repo)

	_, err := service.GetTimeline(
		context.Background(),
		10,
		20,
		"cleaner",
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected status lookup error, got %v",
			err,
		)
	}
}

func TestService_GetTimeline_EventLookupError(t *testing.T) {
	expectedErr := errors.New("event lookup error")

	repo := &mockRepository{
		getApplicationCleanerIDFn: func(
			ctx context.Context,
			applicationID uint,
		) (uint, error) {
			return 20, nil
		},
		getApplicationClientIDFn: func(
			ctx context.Context,
			applicationID uint,
		) (uint, error) {
			return 30, nil
		},
		getCurrentStatusFn: func(
			ctx context.Context,
			applicationID uint,
		) (string, error) {
			return "pending", nil
		},
		listByApplicationIDFn: func(
			ctx context.Context,
			applicationID uint,
		) ([]Event, error) {
			return nil, expectedErr
		},
	}

	service := NewService(repo)

	_, err := service.GetTimeline(
		context.Background(),
		10,
		20,
		"cleaner",
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected event lookup error, got %v",
			err,
		)
	}
}

func TestIsAllowedStatus(t *testing.T) {
	validStatuses := []string{
		"pending",
		"shortlisted",
		"invited",
		"accepted",
		"rejected",
		"completed",
		"cancelled",
	}

	for _, status := range validStatuses {
		if !isAllowedStatus(status) {
			t.Fatalf(
				"expected status %s to be allowed",
				status,
			)
		}
	}

	if isAllowedStatus("unknown") {
		t.Fatal("expected unknown status to be rejected")
	}
}

func successfulTimelineRepository() *mockRepository {
	return &mockRepository{
		getApplicationCleanerIDFn: func(
			ctx context.Context,
			applicationID uint,
		) (uint, error) {
			return 20, nil
		},
		getApplicationClientIDFn: func(
			ctx context.Context,
			applicationID uint,
		) (uint, error) {
			return 30, nil
		},
		getCurrentStatusFn: func(
			ctx context.Context,
			applicationID uint,
		) (string, error) {
			return "shortlisted", nil
		},
		listByApplicationIDFn: func(
			ctx context.Context,
			applicationID uint,
		) ([]Event, error) {
			return []Event{
				{
					ID:            1,
					ApplicationID: applicationID,
					Status:        "pending",
				},
				{
					ID:            2,
					ApplicationID: applicationID,
					Status:        "shortlisted",
				},
			}, nil
		},
	}
}
