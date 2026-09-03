package availability

import (
	"context"
	"errors"
	"testing"
	"time"
)

type mockRepository struct {
	createFn                func(context.Context, *CleanerAvailability) error
	getByIDFn               func(context.Context, uint) (*CleanerAvailability, error)
	listByCleanerIDFn       func(context.Context, uint) ([]CleanerAvailability, error)
	updateFn                func(context.Context, *CleanerAvailability) error
	deleteFn                func(context.Context, uint) error
	hasConflictFn           func(context.Context, uint, string, string, string, uint) (bool, error)
	createBlockFn           func(context.Context, *AvailabilityBlock) error
	listBlocksByCleanerIDFn func(context.Context, uint) ([]AvailabilityBlock, error)
	deleteBlockFn           func(context.Context, uint, uint) error
}

func (m *mockRepository) Create(
	ctx context.Context,
	availability *CleanerAvailability,
) error {
	if m.createFn != nil {
		return m.createFn(ctx, availability)
	}

	return nil
}

func (m *mockRepository) GetByID(
	ctx context.Context,
	id uint,
) (*CleanerAvailability, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}

	return nil, nil
}

func (m *mockRepository) ListByCleanerID(
	ctx context.Context,
	cleanerID uint,
) ([]CleanerAvailability, error) {
	if m.listByCleanerIDFn != nil {
		return m.listByCleanerIDFn(ctx, cleanerID)
	}

	return nil, nil
}

func (m *mockRepository) Update(
	ctx context.Context,
	availability *CleanerAvailability,
) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, availability)
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

func (m *mockRepository) HasConflict(
	ctx context.Context,
	cleanerID uint,
	availableDate string,
	startTime string,
	endTime string,
	excludeID uint,
) (bool, error) {
	if m.hasConflictFn != nil {
		return m.hasConflictFn(
			ctx,
			cleanerID,
			availableDate,
			startTime,
			endTime,
			excludeID,
		)
	}

	return false, nil
}

func (m *mockRepository) CreateBlock(
	ctx context.Context,
	block *AvailabilityBlock,
) error {
	if m.createBlockFn != nil {
		return m.createBlockFn(ctx, block)
	}

	return nil
}

func (m *mockRepository) ListBlocksByCleanerID(
	ctx context.Context,
	cleanerID uint,
) ([]AvailabilityBlock, error) {
	if m.listBlocksByCleanerIDFn != nil {
		return m.listBlocksByCleanerIDFn(ctx, cleanerID)
	}

	return nil, nil
}

func (m *mockRepository) DeleteBlock(
	ctx context.Context,
	blockID uint,
	cleanerID uint,
) error {
	if m.deleteBlockFn != nil {
		return m.deleteBlockFn(ctx, blockID, cleanerID)
	}

	return nil
}

func (m *mockRepository) ListByCleanerIDRange(
	ctx context.Context,
	cleanerID uint,
	fromDate string,
	toDate string,
) ([]CleanerAvailability, error) {
	return nil, nil
}

func (m *mockRepository) ListBlocksByCleanerIDRange(
	ctx context.Context,
	cleanerID uint,
	startAt time.Time,
	endAt time.Time,
) ([]AvailabilityBlock, error) {
	return nil, nil
}

func (m *mockRepository) CreateRecurring(
	ctx context.Context,
	recurring *RecurringAvailability,
) error {
	return nil
}

func (m *mockRepository) ListRecurringByCleanerID(
	ctx context.Context,
	cleanerID uint,
) ([]RecurringAvailability, error) {
	return nil, nil
}

func (m *mockRepository) DeleteRecurring(
	ctx context.Context,
	recurringID uint,
	cleanerID uint,
) error {
	return nil
}

func (m *mockRepository) UpsertSettings(
	ctx context.Context,
	settings *AvailabilitySettings,
) error {
	return nil
}

func (m *mockRepository) GetSettings(
	ctx context.Context,
	cleanerID uint,
) (*AvailabilitySettings, error) {
	return nil, ErrAvailabilitySettingsNotFound
}

func (m *mockRepository) CreateOverride(
	ctx context.Context,
	override *AvailabilityOverride,
) error {
	return nil
}

func (m *mockRepository) ListOverridesByCleanerIDRange(
	ctx context.Context,
	cleanerID uint,
	fromDate string,
	toDate string,
) ([]AvailabilityOverride, error) {
	return nil, nil
}

func (m *mockRepository) DeleteOverride(
	ctx context.Context,
	overrideID uint,
	cleanerID uint,
) error {
	return nil
}

func TestNewService(t *testing.T) {
	repo := &mockRepository{}

	service := NewService(repo)

	if service == nil {
		t.Fatal("expected service")
	}

	if service.repo != repo {
		t.Fatal("expected repository to be assigned")
	}
}

func TestService_Create_Success(t *testing.T) {
	repo := &mockRepository{
		hasConflictFn: func(
			ctx context.Context,
			cleanerID uint,
			availableDate string,
			startTime string,
			endTime string,
			excludeID uint,
		) (bool, error) {
			if cleanerID != 8 {
				t.Fatalf("expected cleaner ID 8, got %d", cleanerID)
			}

			if excludeID != 0 {
				t.Fatalf("expected exclude ID 0, got %d", excludeID)
			}

			return false, nil
		},
		createFn: func(
			ctx context.Context,
			availability *CleanerAvailability,
		) error {
			availability.ID = 12
			return nil
		},
	}

	service := NewService(repo)

	availability, err := service.Create(
		context.Background(),
		8,
		CleanerAvailabilityRequest{
			AvailableDate: "2026-08-10",
			StartTime:     "09:00",
			EndTime:       "17:00",
			Notes:         "Available all day",
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if availability == nil {
		t.Fatal("expected availability")
	}

	if availability.ID != 12 {
		t.Fatalf("expected ID 12, got %d", availability.ID)
	}

	if availability.Status != "available" {
		t.Fatalf("expected default status available, got %q", availability.Status)
	}
}

func TestService_Create_InvalidInput(t *testing.T) {
	service := NewService(&mockRepository{})

	tests := []struct {
		name      string
		cleanerID uint
		req       CleanerAvailabilityRequest
	}{
		{
			name:      "zero cleaner ID",
			cleanerID: 0,
			req: CleanerAvailabilityRequest{
				AvailableDate: "2026-08-10",
				StartTime:     "09:00",
				EndTime:       "17:00",
			},
		},
		{
			name:      "missing date",
			cleanerID: 8,
			req: CleanerAvailabilityRequest{
				StartTime: "09:00",
				EndTime:   "17:00",
			},
		},
		{
			name:      "invalid date",
			cleanerID: 8,
			req: CleanerAvailabilityRequest{
				AvailableDate: "tomorrow",
				StartTime:     "09:00",
				EndTime:       "17:00",
			},
		},
		{
			name:      "invalid status",
			cleanerID: 8,
			req: CleanerAvailabilityRequest{
				AvailableDate: "2026-08-10",
				StartTime:     "09:00",
				EndTime:       "17:00",
				Status:        "sleeping",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			availability, err := service.Create(
				context.Background(),
				test.cleanerID,
				test.req,
			)

			if availability != nil {
				t.Fatalf(
					"expected nil availability, got %+v",
					availability,
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

func TestService_Create_Conflict(t *testing.T) {
	repo := &mockRepository{
		hasConflictFn: func(
			context.Context,
			uint,
			string,
			string,
			string,
			uint,
		) (bool, error) {
			return true, nil
		},
	}

	service := NewService(repo)

	availability, err := service.Create(
		context.Background(),
		8,
		CleanerAvailabilityRequest{
			AvailableDate: "2026-08-10",
			StartTime:     "09:00",
			EndTime:       "17:00",
		},
	)

	if availability != nil {
		t.Fatalf(
			"expected nil availability, got %+v",
			availability,
		)
	}

	if !errors.Is(err, ErrAvailabilityConflict) {
		t.Fatalf(
			"expected ErrAvailabilityConflict, got %v",
			err,
		)
	}
}

func TestService_Create_ConflictCheckError(t *testing.T) {
	expectedErr := errors.New("conflict check failed")

	repo := &mockRepository{
		hasConflictFn: func(
			context.Context,
			uint,
			string,
			string,
			string,
			uint,
		) (bool, error) {
			return false, expectedErr
		},
	}

	service := NewService(repo)

	availability, err := service.Create(
		context.Background(),
		8,
		CleanerAvailabilityRequest{
			AvailableDate: "2026-08-10",
			StartTime:     "09:00",
			EndTime:       "17:00",
		},
	)

	if availability != nil {
		t.Fatalf(
			"expected nil availability, got %+v",
			availability,
		)
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
}

func TestService_Create_RepositoryError(t *testing.T) {
	expectedErr := errors.New("create failed")

	repo := &mockRepository{
		createFn: func(
			context.Context,
			*CleanerAvailability,
		) error {
			return expectedErr
		},
	}

	service := NewService(repo)

	availability, err := service.Create(
		context.Background(),
		8,
		CleanerAvailabilityRequest{
			AvailableDate: "2026-08-10",
			StartTime:     "09:00",
			EndTime:       "17:00",
		},
	)

	if availability != nil {
		t.Fatalf(
			"expected nil availability, got %+v",
			availability,
		)
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
}

func TestService_GetByID_Success(t *testing.T) {
	expected := &CleanerAvailability{
		ID:        1,
		CleanerID: 8,
	}

	repo := &mockRepository{
		getByIDFn: func(
			context.Context,
			uint,
		) (*CleanerAvailability, error) {
			return expected, nil
		},
	}

	service := NewService(repo)

	availability, err := service.GetByID(
		context.Background(),
		1,
		8,
		"cleaner",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if availability != expected {
		t.Fatal("expected returned availability")
	}
}

func TestService_GetByID_AdminSuccess(t *testing.T) {
	expected := &CleanerAvailability{
		ID:        1,
		CleanerID: 8,
	}

	repo := &mockRepository{
		getByIDFn: func(
			context.Context,
			uint,
		) (*CleanerAvailability, error) {
			return expected, nil
		},
	}

	service := NewService(repo)

	availability, err := service.GetByID(
		context.Background(),
		1,
		99,
		"admin",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if availability != expected {
		t.Fatal("expected returned availability")
	}
}

func TestService_GetByID_Forbidden(t *testing.T) {
	repo := &mockRepository{
		getByIDFn: func(
			context.Context,
			uint,
		) (*CleanerAvailability, error) {
			return &CleanerAvailability{
				ID:        1,
				CleanerID: 8,
			}, nil
		},
	}

	service := NewService(repo)

	availability, err := service.GetByID(
		context.Background(),
		1,
		20,
		"cleaner",
	)

	if availability != nil {
		t.Fatalf(
			"expected nil availability, got %+v",
			availability,
		)
	}

	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestService_ListMine_Success(t *testing.T) {
	repo := &mockRepository{
		listByCleanerIDFn: func(
			ctx context.Context,
			cleanerID uint,
		) ([]CleanerAvailability, error) {
			return []CleanerAvailability{
				{ID: 1, CleanerID: cleanerID},
				{ID: 2, CleanerID: cleanerID},
			}, nil
		},
	}

	service := NewService(repo)

	records, err := service.ListMine(
		context.Background(),
		8,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(records) != 2 {
		t.Fatalf("expected 2 records, got %d", len(records))
	}
}

func TestService_Update_Success(t *testing.T) {
	getCalls := 0

	repo := &mockRepository{
		getByIDFn: func(
			context.Context,
			uint,
		) (*CleanerAvailability, error) {
			getCalls++

			return &CleanerAvailability{
				ID:            1,
				CleanerID:     8,
				AvailableDate: "2026-08-10",
				StartTime:     "09:00",
				EndTime:       "17:00",
				Status:        "available",
			}, nil
		},
		hasConflictFn: func(
			ctx context.Context,
			cleanerID uint,
			availableDate string,
			startTime string,
			endTime string,
			excludeID uint,
		) (bool, error) {
			if excludeID != 1 {
				t.Fatalf(
					"expected exclude ID 1, got %d",
					excludeID,
				)
			}

			return false, nil
		},
		updateFn: func(
			context.Context,
			*CleanerAvailability,
		) error {
			return nil
		},
	}

	service := NewService(repo)

	availability, err := service.Update(
		context.Background(),
		1,
		8,
		"cleaner",
		UpdateAvailabilityRequest{
			AvailableDate: "2026-08-11",
			StartTime:     "10:00",
			EndTime:       "18:00",
			Status:        "busy",
			Notes:         "Updated",
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if availability == nil {
		t.Fatal("expected availability")
	}

	if getCalls != 2 {
		t.Fatalf(
			"expected GetByID twice, got %d",
			getCalls,
		)
	}
}

func TestService_Update_Conflict(t *testing.T) {
	repo := &mockRepository{
		getByIDFn: func(
			context.Context,
			uint,
		) (*CleanerAvailability, error) {
			return &CleanerAvailability{
				ID:        1,
				CleanerID: 8,
			}, nil
		},
		hasConflictFn: func(
			context.Context,
			uint,
			string,
			string,
			string,
			uint,
		) (bool, error) {
			return true, nil
		},
	}

	service := NewService(repo)

	availability, err := service.Update(
		context.Background(),
		1,
		8,
		"cleaner",
		UpdateAvailabilityRequest{
			AvailableDate: "2026-08-11",
			StartTime:     "10:00",
			EndTime:       "18:00",
			Status:        "available",
		},
	)

	if availability != nil {
		t.Fatalf(
			"expected nil availability, got %+v",
			availability,
		)
	}

	if !errors.Is(err, ErrAvailabilityConflict) {
		t.Fatalf(
			"expected ErrAvailabilityConflict, got %v",
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
		) (*CleanerAvailability, error) {
			return &CleanerAvailability{
				ID:        1,
				CleanerID: 8,
			}, nil
		},
		deleteFn: func(
			context.Context,
			uint,
		) error {
			deleteCalled = true
			return nil
		},
	}

	service := NewService(repo)

	err := service.Delete(
		context.Background(),
		1,
		8,
		"cleaner",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !deleteCalled {
		t.Fatal("expected Delete to be called")
	}
}

func TestService_CreateBlock_Success(t *testing.T) {
	repo := &mockRepository{
		createBlockFn: func(
			ctx context.Context,
			block *AvailabilityBlock,
		) error {
			block.ID = 20
			return nil
		},
	}

	service := NewService(repo)
	startAt := time.Now().Add(24 * time.Hour)
	endAt := startAt.Add(2 * time.Hour)

	req := CreateAvailabilityBlockRequest{
		StartAt: startAt.Format(time.RFC3339),
		EndAt:   endAt.Format(time.RFC3339),
		Reason:  "personal appointment",
	}

	block, err := service.CreateBlock(
		context.Background(),
		8,
		req,
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if block == nil {
		t.Fatal("expected block")
	}

	if block.ID != 20 {
		t.Fatalf("expected block ID 20, got %d", block.ID)
	}

	if block.CleanerID != 8 {
		t.Fatalf(
			"expected cleaner ID 8, got %d",
			block.CleanerID,
		)
	}
}

func TestService_CreateBlock_InvalidInput(t *testing.T) {
	service := NewService(&mockRepository{})

	tests := []struct {
		name      string
		cleanerID uint
		req       CreateAvailabilityBlockRequest
	}{
		{
			name:      "zero cleaner ID",
			cleanerID: 0,
			req: CreateAvailabilityBlockRequest{
				StartAt: "2026-08-10T09:00:00Z",
				EndAt:   "2026-08-10T17:00:00Z",
			},
		},
		{
			name:      "invalid start",
			cleanerID: 8,
			req: CreateAvailabilityBlockRequest{
				StartAt: "tomorrow",
				EndAt:   "2026-08-10T17:00:00Z",
			},
		},
		{
			name:      "end before start",
			cleanerID: 8,
			req: CreateAvailabilityBlockRequest{
				StartAt: "2026-08-10T17:00:00Z",
				EndAt:   "2026-08-10T09:00:00Z",
			},
		},
		{
			name:      "block too long",
			cleanerID: 8,
			req: CreateAvailabilityBlockRequest{
				StartAt: "2026-08-10T09:00:00Z",
				EndAt:   "2026-10-10T09:00:00Z",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			block, err := service.CreateBlock(
				context.Background(),
				test.cleanerID,
				test.req,
			)

			if block != nil {
				t.Fatalf(
					"expected nil block, got %+v",
					block,
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

func TestService_CreateBlock_RepositoryError(t *testing.T) {
	expectedErr := errors.New("create block failed")

	repo := &mockRepository{
		createBlockFn: func(
			context.Context,
			*AvailabilityBlock,
		) error {
			return expectedErr
		},
	}

	service := NewService(repo)

	startAt := time.Now().Add(24 * time.Hour)
	endAt := startAt.Add(2 * time.Hour)

	req := CreateAvailabilityBlockRequest{
		StartAt: startAt.Format(time.RFC3339),
		EndAt:   endAt.Format(time.RFC3339),
		Reason:  "personal appointment",
	}

	block, err := service.CreateBlock(
		context.Background(),
		8,
		req,
	)
	
	if block != nil {
		t.Fatalf("expected nil block, got %+v", block)
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
}

func TestService_ListBlocks_Success(t *testing.T) {
	repo := &mockRepository{
		listBlocksByCleanerIDFn: func(
			context.Context,
			uint,
		) ([]AvailabilityBlock, error) {
			return []AvailabilityBlock{
				{
					ID:        1,
					CleanerID: 8,
					StartAt:   time.Now(),
				},
			}, nil
		},
	}

	service := NewService(repo)

	blocks, err := service.ListBlocks(
		context.Background(),
		8,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(blocks) != 1 {
		t.Fatalf("expected 1 block, got %d", len(blocks))
	}
}

func TestService_DeleteBlock_Success(t *testing.T) {
	deleteCalled := false

	repo := &mockRepository{
		deleteBlockFn: func(
			ctx context.Context,
			blockID uint,
			cleanerID uint,
		) error {
			deleteCalled = true

			if blockID != 1 {
				t.Fatalf(
					"expected block ID 1, got %d",
					blockID,
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

	err := service.DeleteBlock(
		context.Background(),
		1,
		8,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !deleteCalled {
		t.Fatal("expected DeleteBlock to be called")
	}
}

func TestService_DeleteBlock_InvalidInput(t *testing.T) {
	service := NewService(&mockRepository{})

	err := service.DeleteBlock(
		context.Background(),
		0,
		8,
	)

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf(
			"expected ErrInvalidInput, got %v",
			err,
		)
	}
}
