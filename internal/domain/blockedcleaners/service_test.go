package blockedcleaners

import (
	"context"
	"errors"
	"testing"
	"time"
)

type mockRepository struct {
	createFn         func(context.Context, *BlockedCleaner) error
	deleteFn         func(context.Context, uint, uint) error
	listByClientIDFn func(context.Context, uint) ([]BlockedCleaner, error)
	existsFn         func(context.Context, uint, uint) (bool, error)
}

func (m *mockRepository) Create(
	ctx context.Context,
	block *BlockedCleaner,
) error {
	if m.createFn != nil {
		return m.createFn(ctx, block)
	}

	return nil
}

func (m *mockRepository) Delete(
	ctx context.Context,
	clientID uint,
	cleanerID uint,
) error {
	if m.deleteFn != nil {
		return m.deleteFn(
			ctx,
			clientID,
			cleanerID,
		)
	}

	return nil
}

func (m *mockRepository) ListByClientID(
	ctx context.Context,
	clientID uint,
) ([]BlockedCleaner, error) {
	if m.listByClientIDFn != nil {
		return m.listByClientIDFn(
			ctx,
			clientID,
		)
	}

	return nil, nil
}

func (m *mockRepository) Exists(
	ctx context.Context,
	clientID uint,
	cleanerID uint,
) (bool, error) {
	if m.existsFn != nil {
		return m.existsFn(
			ctx,
			clientID,
			cleanerID,
		)
	}

	return false, nil
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

func TestService_Block_Success(t *testing.T) {
	existsCalled := false
	createCalled := false

	repo := &mockRepository{
		existsFn: func(
			ctx context.Context,
			clientID uint,
			cleanerID uint,
		) (bool, error) {
			existsCalled = true

			if clientID != 5 {
				t.Fatalf(
					"expected client ID 5, got %d",
					clientID,
				)
			}

			if cleanerID != 8 {
				t.Fatalf(
					"expected cleaner ID 8, got %d",
					cleanerID,
				)
			}

			return false, nil
		},
		createFn: func(
			ctx context.Context,
			block *BlockedCleaner,
		) error {
			createCalled = true

			if block.ClientID != 5 {
				t.Fatalf(
					"expected client ID 5, got %d",
					block.ClientID,
				)
			}

			if block.CleanerID != 8 {
				t.Fatalf(
					"expected cleaner ID 8, got %d",
					block.CleanerID,
				)
			}

			if block.Reason != "Repeated no-shows" {
				t.Fatalf(
					"expected trimmed reason, got %q",
					block.Reason,
				)
			}

			block.ID = 12
			block.CreatedAt = time.Now()

			return nil
		},
	}

	service := NewService(repo)

	block, err := service.Block(
		context.Background(),
		5,
		8,
		BlockedCleanerRequest{
			Reason: "  Repeated no-shows  ",
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if block == nil {
		t.Fatal("expected blocked cleaner")
	}

	if block.ID != 12 {
		t.Fatalf(
			"expected ID 12, got %d",
			block.ID,
		)
	}

	if block.ClientID != 5 {
		t.Fatalf(
			"expected client ID 5, got %d",
			block.ClientID,
		)
	}

	if block.CleanerID != 8 {
		t.Fatalf(
			"expected cleaner ID 8, got %d",
			block.CleanerID,
		)
	}

	if block.Reason != "Repeated no-shows" {
		t.Fatalf(
			"expected trimmed reason, got %q",
			block.Reason,
		)
	}

	if block.CreatedAt.IsZero() {
		t.Fatal("expected created_at")
	}

	if !existsCalled {
		t.Fatal("expected Exists to be called")
	}

	if !createCalled {
		t.Fatal("expected Create to be called")
	}
}

func TestService_Block_InvalidInput(t *testing.T) {
	service := NewService(&mockRepository{})

	tests := []struct {
		name      string
		clientID  uint
		cleanerID uint
	}{
		{
			name:      "zero client ID",
			clientID:  0,
			cleanerID: 8,
		},
		{
			name:      "zero cleaner ID",
			clientID:  5,
			cleanerID: 0,
		},
		{
			name:      "same client and cleaner",
			clientID:  5,
			cleanerID: 5,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			block, err := service.Block(
				context.Background(),
				test.clientID,
				test.cleanerID,
				BlockedCleanerRequest{
					Reason: "test",
				},
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

func TestService_Block_ExistsRepositoryError(t *testing.T) {
	expectedErr := errors.New("exists check failed")

	repo := &mockRepository{
		existsFn: func(
			context.Context,
			uint,
			uint,
		) (bool, error) {
			return false, expectedErr
		},
	}

	service := NewService(repo)

	block, err := service.Block(
		context.Background(),
		5,
		8,
		BlockedCleanerRequest{
			Reason: "Repeated no-shows",
		},
	)

	if block != nil {
		t.Fatalf(
			"expected nil block, got %+v",
			block,
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

func TestService_Block_AlreadyBlocked(t *testing.T) {
	repo := &mockRepository{
		existsFn: func(
			context.Context,
			uint,
			uint,
		) (bool, error) {
			return true, nil
		},
	}

	service := NewService(repo)

	block, err := service.Block(
		context.Background(),
		5,
		8,
		BlockedCleanerRequest{
			Reason: "Repeated no-shows",
		},
	)

	if block != nil {
		t.Fatalf(
			"expected nil block, got %+v",
			block,
		)
	}

	if !errors.Is(err, ErrAlreadyBlocked) {
		t.Fatalf(
			"expected ErrAlreadyBlocked, got %v",
			err,
		)
	}
}

func TestService_Block_CreateRepositoryError(t *testing.T) {
	expectedErr := errors.New("create failed")

	repo := &mockRepository{
		existsFn: func(
			context.Context,
			uint,
			uint,
		) (bool, error) {
			return false, nil
		},
		createFn: func(
			context.Context,
			*BlockedCleaner,
		) error {
			return expectedErr
		},
	}

	service := NewService(repo)

	block, err := service.Block(
		context.Background(),
		5,
		8,
		BlockedCleanerRequest{
			Reason: "Repeated no-shows",
		},
	)

	if block != nil {
		t.Fatalf(
			"expected nil block, got %+v",
			block,
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

func TestService_Unblock_Success(t *testing.T) {
	deleteCalled := false

	repo := &mockRepository{
		deleteFn: func(
			ctx context.Context,
			clientID uint,
			cleanerID uint,
		) error {
			deleteCalled = true

			if clientID != 5 {
				t.Fatalf(
					"expected client ID 5, got %d",
					clientID,
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

	err := service.Unblock(
		context.Background(),
		5,
		8,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !deleteCalled {
		t.Fatal("expected Delete to be called")
	}
}

func TestService_Unblock_InvalidInput(t *testing.T) {
	service := NewService(&mockRepository{})

	tests := []struct {
		name      string
		clientID  uint
		cleanerID uint
	}{
		{
			name:      "zero client ID",
			clientID:  0,
			cleanerID: 8,
		},
		{
			name:      "zero cleaner ID",
			clientID:  5,
			cleanerID: 0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := service.Unblock(
				context.Background(),
				test.clientID,
				test.cleanerID,
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

func TestService_Unblock_NotFound(t *testing.T) {
	repo := &mockRepository{
		deleteFn: func(
			context.Context,
			uint,
			uint,
		) error {
			return ErrBlockedNotFound
		},
	}

	service := NewService(repo)

	err := service.Unblock(
		context.Background(),
		5,
		8,
	)

	if !errors.Is(err, ErrBlockedNotFound) {
		t.Fatalf(
			"expected ErrBlockedNotFound, got %v",
			err,
		)
	}
}

func TestService_Unblock_DeleteRepositoryError(t *testing.T) {
	expectedErr := errors.New("delete failed")

	repo := &mockRepository{
		deleteFn: func(
			context.Context,
			uint,
			uint,
		) error {
			return expectedErr
		},
	}

	service := NewService(repo)

	err := service.Unblock(
		context.Background(),
		5,
		8,
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected %v, got %v",
			expectedErr,
			err,
		)
	}
}

func TestService_ListMine_Success(t *testing.T) {
	repo := &mockRepository{
		listByClientIDFn: func(
			ctx context.Context,
			clientID uint,
		) ([]BlockedCleaner, error) {
			if clientID != 5 {
				t.Fatalf(
					"expected client ID 5, got %d",
					clientID,
				)
			}

			return []BlockedCleaner{
				{
					ID:        1,
					ClientID:  5,
					CleanerID: 8,
					Reason:    "Repeated no-shows",
				},
				{
					ID:        2,
					ClientID:  5,
					CleanerID: 9,
					Reason:    "Abusive messages",
				},
			}, nil
		},
	}

	service := NewService(repo)

	blocks, err := service.ListMine(
		context.Background(),
		5,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(blocks) != 2 {
		t.Fatalf(
			"expected 2 blocked cleaners, got %d",
			len(blocks),
		)
	}

	if blocks[0].CleanerID != 8 {
		t.Fatalf(
			"expected first cleaner ID 8, got %d",
			blocks[0].CleanerID,
		)
	}

	if blocks[1].CleanerID != 9 {
		t.Fatalf(
			"expected second cleaner ID 9, got %d",
			blocks[1].CleanerID,
		)
	}
}

func TestService_ListMine_InvalidInput(t *testing.T) {
	service := NewService(&mockRepository{})

	blocks, err := service.ListMine(
		context.Background(),
		0,
	)

	if blocks != nil {
		t.Fatalf(
			"expected nil blocks, got %+v",
			blocks,
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
	expectedErr := errors.New("list failed")

	repo := &mockRepository{
		listByClientIDFn: func(
			context.Context,
			uint,
		) ([]BlockedCleaner, error) {
			return nil, expectedErr
		},
	}

	service := NewService(repo)

	blocks, err := service.ListMine(
		context.Background(),
		5,
	)

	if blocks != nil {
		t.Fatalf(
			"expected nil blocks, got %+v",
			blocks,
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

func TestService_Exists_True(t *testing.T) {
	repo := &mockRepository{
		existsFn: func(
			ctx context.Context,
			clientID uint,
			cleanerID uint,
		) (bool, error) {
			if clientID != 5 {
				t.Fatalf(
					"expected client ID 5, got %d",
					clientID,
				)
			}

			if cleanerID != 8 {
				t.Fatalf(
					"expected cleaner ID 8, got %d",
					cleanerID,
				)
			}

			return true, nil
		},
	}

	service := NewService(repo)

	exists, err := service.Exists(
		context.Background(),
		5,
		8,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !exists {
		t.Fatal("expected cleaner to be blocked")
	}
}

func TestService_Exists_False(t *testing.T) {
	repo := &mockRepository{
		existsFn: func(
			context.Context,
			uint,
			uint,
		) (bool, error) {
			return false, nil
		},
	}

	service := NewService(repo)

	exists, err := service.Exists(
		context.Background(),
		5,
		8,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if exists {
		t.Fatal("expected cleaner not to be blocked")
	}
}

func TestService_Exists_InvalidInput(t *testing.T) {
	service := NewService(&mockRepository{})

	tests := []struct {
		name      string
		clientID  uint
		cleanerID uint
	}{
		{
			name:      "zero client ID",
			clientID:  0,
			cleanerID: 8,
		},
		{
			name:      "zero cleaner ID",
			clientID:  5,
			cleanerID: 0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			exists, err := service.Exists(
				context.Background(),
				test.clientID,
				test.cleanerID,
			)

			if exists {
				t.Fatal("expected exists to be false")
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

func TestService_Exists_RepositoryError(t *testing.T) {
	expectedErr := errors.New("exists failed")

	repo := &mockRepository{
		existsFn: func(
			context.Context,
			uint,
			uint,
		) (bool, error) {
			return false, expectedErr
		},
	}

	service := NewService(repo)

	exists, err := service.Exists(
		context.Background(),
		5,
		8,
	)

	if exists {
		t.Fatal("expected exists to be false")
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected %v, got %v",
			expectedErr,
			err,
		)
	}
}
