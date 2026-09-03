package jobalerts

import (
	"context"
	"errors"
	"testing"
	"time"
)

type mockRepository struct {
	createFn       func(context.Context, *JobAlert) error
	getByIDFn      func(context.Context, uint) (*JobAlert, error)
	listByUserIDFn func(context.Context, uint) ([]JobAlert, error)
	updateFn       func(context.Context, *JobAlert) error
	deleteFn       func(context.Context, uint) error
}

func (m *mockRepository) Create(
	ctx context.Context,
	alert *JobAlert,
) error {
	if m.createFn != nil {
		return m.createFn(ctx, alert)
	}

	return nil
}

func (m *mockRepository) GetByID(
	ctx context.Context,
	id uint,
) (*JobAlert, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}

	return nil, nil
}

func (m *mockRepository) ListByUserID(
	ctx context.Context,
	userID uint,
) ([]JobAlert, error) {
	if m.listByUserIDFn != nil {
		return m.listByUserIDFn(ctx, userID)
	}

	return nil, nil
}

func (m *mockRepository) Update(
	ctx context.Context,
	alert *JobAlert,
) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, alert)
	}

	return nil
}

func (m *mockRepository) Delete(
	ctx context.Context,
	id uint,
) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}

	return nil
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

func TestService_Create_Success(t *testing.T) {
	createCalled := false

	repo := &mockRepository{
		createFn: func(
			ctx context.Context,
			alert *JobAlert,
		) error {
			createCalled = true

			if alert.UserID != 5 {
				t.Fatalf(
					"expected user ID 5, got %d",
					alert.UserID,
				)
			}

			if alert.Location != "East London" {
				t.Fatalf(
					"expected trimmed location East London, got %q",
					alert.Location,
				)
			}

			if alert.JobType != "domestic" {
				t.Fatalf(
					"expected trimmed job type domestic, got %q",
					alert.JobType,
				)
			}

			if alert.MinimumBudget != 80 {
				t.Fatalf(
					"expected minimum budget 80, got %d",
					alert.MinimumBudget,
				)
			}

			if !alert.IsActive {
				t.Fatal("expected alert to be active")
			}

			now := time.Now()

			alert.ID = 12
			alert.CreatedAt = now
			alert.UpdatedAt = now

			return nil
		},
	}

	service := NewService(repo)

	alert, err := service.Create(
		context.Background(),
		5,
		CreateJobAlertRequest{
			Location:      "  East London  ",
			JobType:       "  domestic  ",
			MinimumBudget: 80,
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if alert == nil {
		t.Fatal("expected job alert")
	}

	if alert.ID != 12 {
		t.Fatalf(
			"expected alert ID 12, got %d",
			alert.ID,
		)
	}

	if alert.Location != "East London" {
		t.Fatalf(
			"expected East London, got %q",
			alert.Location,
		)
	}

	if alert.JobType != "domestic" {
		t.Fatalf(
			"expected domestic, got %q",
			alert.JobType,
		)
	}

	if !createCalled {
		t.Fatal("expected Create repository call")
	}
}

func TestService_Create_InvalidInput(t *testing.T) {
	service := NewService(&mockRepository{})

	tests := []struct {
		name    string
		userID  uint
		request CreateJobAlertRequest
	}{
		{
			name:   "zero user ID",
			userID: 0,
			request: CreateJobAlertRequest{
				Location: "London",
				JobType:  "domestic",
			},
		},
		{
			name:   "blank location",
			userID: 5,
			request: CreateJobAlertRequest{
				Location: "   ",
				JobType:  "domestic",
			},
		},
		{
			name:   "blank job type",
			userID: 5,
			request: CreateJobAlertRequest{
				Location: "London",
				JobType:  "   ",
			},
		},
		{
			name:   "negative minimum budget",
			userID: 5,
			request: CreateJobAlertRequest{
				Location:      "London",
				JobType:       "domestic",
				MinimumBudget: -1,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			alert, err := service.Create(
				context.Background(),
				test.userID,
				test.request,
			)

			if alert != nil {
				t.Fatalf(
					"expected nil alert, got %+v",
					alert,
				)
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

func TestService_Create_RepositoryError(t *testing.T) {
	expectedErr := errors.New("create job alert failed")

	repo := &mockRepository{
		createFn: func(
			context.Context,
			*JobAlert,
		) error {
			return expectedErr
		},
	}

	service := NewService(repo)

	alert, err := service.Create(
		context.Background(),
		5,
		CreateJobAlertRequest{
			Location: "London",
			JobType:  "domestic",
		},
	)

	if alert != nil {
		t.Fatalf(
			"expected nil alert, got %+v",
			alert,
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

func TestService_GetByID_Success(t *testing.T) {
	now := time.Now()

	repo := &mockRepository{
		getByIDFn: func(
			ctx context.Context,
			id uint,
		) (*JobAlert, error) {
			if id != 12 {
				t.Fatalf(
					"expected alert ID 12, got %d",
					id,
				)
			}

			return &JobAlert{
				ID:            12,
				UserID:        5,
				Location:      "London",
				JobType:       "domestic",
				MinimumBudget: 80,
				IsActive:      true,
				CreatedAt:     now,
				UpdatedAt:     now,
			}, nil
		},
	}

	service := NewService(repo)

	alert, err := service.GetByID(
		context.Background(),
		12,
		5,
		"user",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if alert == nil {
		t.Fatal("expected job alert")
	}

	if alert.ID != 12 {
		t.Fatalf(
			"expected alert ID 12, got %d",
			alert.ID,
		)
	}
}

func TestService_GetByID_AdminSuccess(t *testing.T) {
	repo := &mockRepository{
		getByIDFn: func(
			context.Context,
			uint,
		) (*JobAlert, error) {
			return &JobAlert{
				ID:     12,
				UserID: 99,
			}, nil
		},
	}

	service := NewService(repo)

	alert, err := service.GetByID(
		context.Background(),
		12,
		5,
		"admin",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if alert == nil {
		t.Fatal("expected job alert")
	}

	if alert.UserID != 99 {
		t.Fatalf(
			"expected owner ID 99, got %d",
			alert.UserID,
		)
	}
}

func TestService_GetByID_InvalidInput(t *testing.T) {
	service := NewService(&mockRepository{})

	tests := []struct {
		name    string
		alertID uint
		userID  uint
	}{
		{
			name:    "zero alert ID",
			alertID: 0,
			userID:  5,
		},
		{
			name:    "zero user ID",
			alertID: 12,
			userID:  0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			alert, err := service.GetByID(
				context.Background(),
				test.alertID,
				test.userID,
				"user",
			)

			if alert != nil {
				t.Fatalf(
					"expected nil alert, got %+v",
					alert,
				)
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

func TestService_GetByID_NotFound(t *testing.T) {
	repo := &mockRepository{
		getByIDFn: func(
			context.Context,
			uint,
		) (*JobAlert, error) {
			return nil, ErrAlertNotFound
		},
	}

	service := NewService(repo)

	alert, err := service.GetByID(
		context.Background(),
		12,
		5,
		"user",
	)

	if alert != nil {
		t.Fatalf(
			"expected nil alert, got %+v",
			alert,
		)
	}

	if !errors.Is(err, ErrAlertNotFound) {
		t.Fatalf(
			"expected ErrAlertNotFound, got %v",
			err,
		)
	}
}

func TestService_GetByID_Forbidden(t *testing.T) {
	repo := &mockRepository{
		getByIDFn: func(
			context.Context,
			uint,
		) (*JobAlert, error) {
			return &JobAlert{
				ID:     12,
				UserID: 99,
			}, nil
		},
	}

	service := NewService(repo)

	alert, err := service.GetByID(
		context.Background(),
		12,
		5,
		"user",
	)

	if alert != nil {
		t.Fatalf(
			"expected nil alert, got %+v",
			alert,
		)
	}

	if !errors.Is(err, ErrForbidden) {
		t.Fatalf(
			"expected ErrForbidden, got %v",
			err,
		)
	}
}

func TestService_GetByID_RepositoryError(t *testing.T) {
	expectedErr := errors.New("get alert failed")

	repo := &mockRepository{
		getByIDFn: func(
			context.Context,
			uint,
		) (*JobAlert, error) {
			return nil, expectedErr
		},
	}

	service := NewService(repo)

	alert, err := service.GetByID(
		context.Background(),
		12,
		5,
		"user",
	)

	if alert != nil {
		t.Fatalf(
			"expected nil alert, got %+v",
			alert,
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

func TestService_ListMine_Success(t *testing.T) {
	now := time.Now()

	repo := &mockRepository{
		listByUserIDFn: func(
			ctx context.Context,
			userID uint,
		) ([]JobAlert, error) {
			if userID != 5 {
				t.Fatalf(
					"expected user ID 5, got %d",
					userID,
				)
			}

			return []JobAlert{
				{
					ID:            1,
					UserID:        5,
					Location:      "London",
					JobType:       "domestic",
					MinimumBudget: 80,
					IsActive:      true,
					CreatedAt:     now,
					UpdatedAt:     now,
				},
				{
					ID:            2,
					UserID:        5,
					Location:      "Essex",
					JobType:       "commercial",
					MinimumBudget: 120,
					IsActive:      false,
					CreatedAt:     now.Add(-time.Hour),
					UpdatedAt:     now.Add(-time.Hour),
				},
			}, nil
		},
	}

	service := NewService(repo)

	alerts, err := service.ListMine(
		context.Background(),
		5,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(alerts) != 2 {
		t.Fatalf(
			"expected 2 alerts, got %d",
			len(alerts),
		)
	}

	if alerts[0].Location != "London" {
		t.Fatalf(
			"expected London, got %q",
			alerts[0].Location,
		)
	}
}

func TestService_ListMine_Empty(t *testing.T) {
	repo := &mockRepository{
		listByUserIDFn: func(
			context.Context,
			uint,
		) ([]JobAlert, error) {
			return []JobAlert{}, nil
		},
	}

	service := NewService(repo)

	alerts, err := service.ListMine(
		context.Background(),
		5,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(alerts) != 0 {
		t.Fatalf(
			"expected no alerts, got %d",
			len(alerts),
		)
	}
}

func TestService_ListMine_InvalidInput(t *testing.T) {
	service := NewService(&mockRepository{})

	alerts, err := service.ListMine(
		context.Background(),
		0,
	)

	if alerts != nil {
		t.Fatalf(
			"expected nil alerts, got %+v",
			alerts,
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
	expectedErr := errors.New("list alerts failed")

	repo := &mockRepository{
		listByUserIDFn: func(
			context.Context,
			uint,
		) ([]JobAlert, error) {
			return nil, expectedErr
		},
	}

	service := NewService(repo)

	alerts, err := service.ListMine(
		context.Background(),
		5,
	)

	if alerts != nil {
		t.Fatalf(
			"expected nil alerts, got %+v",
			alerts,
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

func TestService_Update_Success(t *testing.T) {
	getCalls := 0
	updateCalled := false

	repo := &mockRepository{
		getByIDFn: func(
			ctx context.Context,
			id uint,
		) (*JobAlert, error) {
			getCalls++

			if getCalls == 1 {
				return &JobAlert{
					ID:            12,
					UserID:        5,
					Location:      "London",
					JobType:       "domestic",
					MinimumBudget: 80,
					IsActive:      true,
				}, nil
			}

			return &JobAlert{
				ID:            12,
				UserID:        5,
				Location:      "East London",
				JobType:       "airbnb",
				MinimumBudget: 120,
				IsActive:      false,
			}, nil
		},
		updateFn: func(
			ctx context.Context,
			alert *JobAlert,
		) error {
			updateCalled = true

			if alert.ID != 12 {
				t.Fatalf(
					"expected alert ID 12, got %d",
					alert.ID,
				)
			}

			if alert.Location != "East London" {
				t.Fatalf(
					"expected trimmed location East London, got %q",
					alert.Location,
				)
			}

			if alert.JobType != "airbnb" {
				t.Fatalf(
					"expected trimmed job type airbnb, got %q",
					alert.JobType,
				)
			}

			if alert.MinimumBudget != 120 {
				t.Fatalf(
					"expected minimum budget 120, got %d",
					alert.MinimumBudget,
				)
			}

			if alert.IsActive {
				t.Fatal("expected alert to be inactive")
			}

			return nil
		},
	}

	service := NewService(repo)

	alert, err := service.Update(
		context.Background(),
		12,
		5,
		"user",
		UpdateJobAlertRequest{
			Location:      "  East London  ",
			JobType:       "  airbnb  ",
			MinimumBudget: 120,
			IsActive:      false,
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if alert == nil {
		t.Fatal("expected updated alert")
	}

	if alert.Location != "East London" {
		t.Fatalf(
			"expected East London, got %q",
			alert.Location,
		)
	}

	if getCalls != 2 {
		t.Fatalf(
			"expected GetByID called twice, got %d",
			getCalls,
		)
	}

	if !updateCalled {
		t.Fatal("expected Update repository call")
	}
}

func TestService_Update_AdminSuccess(t *testing.T) {
	getCalls := 0

	repo := &mockRepository{
		getByIDFn: func(
			context.Context,
			uint,
		) (*JobAlert, error) {
			getCalls++

			return &JobAlert{
				ID:     12,
				UserID: 99,
			}, nil
		},
	}

	service := NewService(repo)

	alert, err := service.Update(
		context.Background(),
		12,
		5,
		"admin",
		UpdateJobAlertRequest{
			Location: "London",
			JobType:  "domestic",
			IsActive: true,
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if alert == nil {
		t.Fatal("expected updated alert")
	}

	if getCalls != 2 {
		t.Fatalf(
			"expected GetByID called twice, got %d",
			getCalls,
		)
	}
}

func TestService_Update_InvalidInput(t *testing.T) {
	tests := []struct {
		name    string
		alertID uint
		userID  uint
		request UpdateJobAlertRequest
	}{
		{
			name:    "zero alert ID",
			alertID: 0,
			userID:  5,
			request: UpdateJobAlertRequest{
				Location: "London",
				JobType:  "domestic",
			},
		},
		{
			name:    "zero user ID",
			alertID: 12,
			userID:  0,
			request: UpdateJobAlertRequest{
				Location: "London",
				JobType:  "domestic",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service := NewService(&mockRepository{})

			alert, err := service.Update(
				context.Background(),
				test.alertID,
				test.userID,
				"user",
				test.request,
			)

			if alert != nil {
				t.Fatalf(
					"expected nil alert, got %+v",
					alert,
				)
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

func TestService_Update_InvalidRequest(t *testing.T) {
	tests := []struct {
		name    string
		request UpdateJobAlertRequest
	}{
		{
			name: "blank location",
			request: UpdateJobAlertRequest{
				Location: "   ",
				JobType:  "domestic",
			},
		},
		{
			name: "blank job type",
			request: UpdateJobAlertRequest{
				Location: "London",
				JobType:  "   ",
			},
		},
		{
			name: "negative minimum budget",
			request: UpdateJobAlertRequest{
				Location:      "London",
				JobType:       "domestic",
				MinimumBudget: -1,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := &mockRepository{
				getByIDFn: func(
					context.Context,
					uint,
				) (*JobAlert, error) {
					return &JobAlert{
						ID:     12,
						UserID: 5,
					}, nil
				},
			}

			service := NewService(repo)

			alert, err := service.Update(
				context.Background(),
				12,
				5,
				"user",
				test.request,
			)

			if alert != nil {
				t.Fatalf(
					"expected nil alert, got %+v",
					alert,
				)
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

func TestService_Update_NotFound(t *testing.T) {
	repo := &mockRepository{
		getByIDFn: func(
			context.Context,
			uint,
		) (*JobAlert, error) {
			return nil, ErrAlertNotFound
		},
	}

	service := NewService(repo)

	alert, err := service.Update(
		context.Background(),
		12,
		5,
		"user",
		UpdateJobAlertRequest{
			Location: "London",
			JobType:  "domestic",
		},
	)

	if alert != nil {
		t.Fatalf(
			"expected nil alert, got %+v",
			alert,
		)
	}

	if !errors.Is(err, ErrAlertNotFound) {
		t.Fatalf(
			"expected ErrAlertNotFound, got %v",
			err,
		)
	}
}

func TestService_Update_Forbidden(t *testing.T) {
	repo := &mockRepository{
		getByIDFn: func(
			context.Context,
			uint,
		) (*JobAlert, error) {
			return &JobAlert{
				ID:     12,
				UserID: 99,
			}, nil
		},
	}

	service := NewService(repo)

	alert, err := service.Update(
		context.Background(),
		12,
		5,
		"user",
		UpdateJobAlertRequest{
			Location: "London",
			JobType:  "domestic",
		},
	)

	if alert != nil {
		t.Fatalf(
			"expected nil alert, got %+v",
			alert,
		)
	}

	if !errors.Is(err, ErrForbidden) {
		t.Fatalf(
			"expected ErrForbidden, got %v",
			err,
		)
	}
}

func TestService_Update_GetRepositoryError(t *testing.T) {
	expectedErr := errors.New("get alert failed")

	repo := &mockRepository{
		getByIDFn: func(
			context.Context,
			uint,
		) (*JobAlert, error) {
			return nil, expectedErr
		},
	}

	service := NewService(repo)

	alert, err := service.Update(
		context.Background(),
		12,
		5,
		"user",
		UpdateJobAlertRequest{
			Location: "London",
			JobType:  "domestic",
		},
	)

	if alert != nil {
		t.Fatalf(
			"expected nil alert, got %+v",
			alert,
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

func TestService_Update_UpdateRepositoryError(t *testing.T) {
	expectedErr := errors.New("update alert failed")

	repo := &mockRepository{
		getByIDFn: func(
			context.Context,
			uint,
		) (*JobAlert, error) {
			return &JobAlert{
				ID:     12,
				UserID: 5,
			}, nil
		},
		updateFn: func(
			context.Context,
			*JobAlert,
		) error {
			return expectedErr
		},
	}

	service := NewService(repo)

	alert, err := service.Update(
		context.Background(),
		12,
		5,
		"user",
		UpdateJobAlertRequest{
			Location: "London",
			JobType:  "domestic",
		},
	)

	if alert != nil {
		t.Fatalf(
			"expected nil alert, got %+v",
			alert,
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

func TestService_Update_FinalGetRepositoryError(t *testing.T) {
	expectedErr := errors.New("reload alert failed")
	getCalls := 0

	repo := &mockRepository{
		getByIDFn: func(
			context.Context,
			uint,
		) (*JobAlert, error) {
			getCalls++

			if getCalls == 1 {
				return &JobAlert{
					ID:     12,
					UserID: 5,
				}, nil
			}

			return nil, expectedErr
		},
	}

	service := NewService(repo)

	alert, err := service.Update(
		context.Background(),
		12,
		5,
		"user",
		UpdateJobAlertRequest{
			Location: "London",
			JobType:  "domestic",
		},
	)

	if alert != nil {
		t.Fatalf(
			"expected nil alert, got %+v",
			alert,
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

func TestService_Delete_Success(t *testing.T) {
	deleteCalled := false

	repo := &mockRepository{
		getByIDFn: func(
			context.Context,
			uint,
		) (*JobAlert, error) {
			return &JobAlert{
				ID:     12,
				UserID: 5,
			}, nil
		},
		deleteFn: func(
			ctx context.Context,
			id uint,
		) error {
			deleteCalled = true

			if id != 12 {
				t.Fatalf(
					"expected alert ID 12, got %d",
					id,
				)
			}

			return nil
		},
	}

	service := NewService(repo)

	err := service.Delete(
		context.Background(),
		12,
		5,
		"user",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !deleteCalled {
		t.Fatal("expected Delete repository call")
	}
}

func TestService_Delete_AdminSuccess(t *testing.T) {
	repo := &mockRepository{
		getByIDFn: func(
			context.Context,
			uint,
		) (*JobAlert, error) {
			return &JobAlert{
				ID:     12,
				UserID: 99,
			}, nil
		},
	}

	service := NewService(repo)

	err := service.Delete(
		context.Background(),
		12,
		5,
		"admin",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestService_Delete_InvalidInput(t *testing.T) {
	service := NewService(&mockRepository{})

	tests := []struct {
		name    string
		alertID uint
		userID  uint
	}{
		{
			name:    "zero alert ID",
			alertID: 0,
			userID:  5,
		},
		{
			name:    "zero user ID",
			alertID: 12,
			userID:  0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := service.Delete(
				context.Background(),
				test.alertID,
				test.userID,
				"user",
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

func TestService_Delete_NotFound(t *testing.T) {
	repo := &mockRepository{
		getByIDFn: func(
			context.Context,
			uint,
		) (*JobAlert, error) {
			return nil, ErrAlertNotFound
		},
	}

	service := NewService(repo)

	err := service.Delete(
		context.Background(),
		12,
		5,
		"user",
	)

	if !errors.Is(err, ErrAlertNotFound) {
		t.Fatalf(
			"expected ErrAlertNotFound, got %v",
			err,
		)
	}
}

func TestService_Delete_Forbidden(t *testing.T) {
	repo := &mockRepository{
		getByIDFn: func(
			context.Context,
			uint,
		) (*JobAlert, error) {
			return &JobAlert{
				ID:     12,
				UserID: 99,
			}, nil
		},
	}

	service := NewService(repo)

	err := service.Delete(
		context.Background(),
		12,
		5,
		"user",
	)

	if !errors.Is(err, ErrForbidden) {
		t.Fatalf(
			"expected ErrForbidden, got %v",
			err,
		)
	}
}

func TestService_Delete_GetRepositoryError(t *testing.T) {
	expectedErr := errors.New("get alert failed")

	repo := &mockRepository{
		getByIDFn: func(
			context.Context,
			uint,
		) (*JobAlert, error) {
			return nil, expectedErr
		},
	}

	service := NewService(repo)

	err := service.Delete(
		context.Background(),
		12,
		5,
		"user",
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected %v, got %v",
			expectedErr,
			err,
		)
	}
}

func TestService_Delete_DeleteRepositoryError(t *testing.T) {
	expectedErr := errors.New("delete alert failed")

	repo := &mockRepository{
		getByIDFn: func(
			context.Context,
			uint,
		) (*JobAlert, error) {
			return &JobAlert{
				ID:     12,
				UserID: 5,
			}, nil
		},
		deleteFn: func(
			context.Context,
			uint,
		) error {
			return expectedErr
		},
	}

	service := NewService(repo)

	err := service.Delete(
		context.Background(),
		12,
		5,
		"user",
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected %v, got %v",
			expectedErr,
			err,
		)
	}
}
