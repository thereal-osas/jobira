package preferredcleaners

import (
	"context"
	"errors"
	"testing"
	"time"
)

type mockRepository struct {
	createFn         func(context.Context, *PreferredCleaners) error
	deleteFn         func(context.Context, uint, uint) error
	listByClientIDFn func(context.Context, uint) ([]PreferredCleaners, error)
	existsFn         func(context.Context, uint, uint) (bool, error)
}

func (m *mockRepository) Create(
	ctx context.Context,
	preferred *PreferredCleaners,
) error {
	if m.createFn != nil {
		return m.createFn(ctx, preferred)
	}

	return nil
}

func (m *mockRepository) Delete(
	ctx context.Context,
	clientID uint,
	cleanerID uint,
) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, clientID, cleanerID)
	}

	return nil
}

func (m *mockRepository) ListByClientID(
	ctx context.Context,
	clientID uint,
) ([]PreferredCleaners, error) {
	if m.listByClientIDFn != nil {
		return m.listByClientIDFn(ctx, clientID)
	}

	return nil, nil
}

func (m *mockRepository) Exists(
	ctx context.Context,
	clientID uint,
	cleanerID uint,
) (bool, error) {
	if m.existsFn != nil {
		return m.existsFn(ctx, clientID, cleanerID)
	}

	return false, nil
}

type mockBlockChecker struct {
	isBlockedFn func(
		context.Context,
		uint,
		uint,
	) (bool, error)
}

func (m *mockBlockChecker) IsBlocked(
	ctx context.Context,
	clientID uint,
	cleanerID uint,
) (bool, error) {
	if m.isBlockedFn != nil {
		return m.isBlockedFn(
			ctx,
			clientID,
			cleanerID,
		)
	}

	return false, nil
}

func TestNewService(t *testing.T) {
	repo := &mockRepository{}
	blockChecker := &mockBlockChecker{}

	service := NewService(
		repo,
		blockChecker,
	)

	if service == nil {
		t.Fatal("expected service")
	}

	if service.repo != repo {
		t.Fatal("expected repository assigned")
	}

	if service.blockChecker != blockChecker {
		t.Fatal("expected block checker assigned")
	}
}

func TestService_Create_Success(t *testing.T) {
	blockChecked := false
	existsChecked := false
	createCalled := false

	repo := &mockRepository{
		existsFn: func(
			ctx context.Context,
			clientID uint,
			cleanerID uint,
		) (bool, error) {
			existsChecked = true

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
			preferred *PreferredCleaners,
		) error {
			createCalled = true

			if preferred.ClientID != 5 {
				t.Fatalf(
					"expected client ID 5, got %d",
					preferred.ClientID,
				)
			}

			if preferred.CleanerID != 8 {
				t.Fatalf(
					"expected cleaner ID 8, got %d",
					preferred.CleanerID,
				)
			}

			preferred.ID = 12
			preferred.CreatedAt = time.Now()

			return nil
		},
	}

	blockChecker := &mockBlockChecker{
		isBlockedFn: func(
			ctx context.Context,
			clientID uint,
			cleanerID uint,
		) (bool, error) {
			blockChecked = true
			return false, nil
		},
	}

	service := NewService(
		repo,
		blockChecker,
	)

	preferred, err := service.Create(
		context.Background(),
		5,
		CreatePreferredCleanerRequest{
			CleanerID: 8,
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if preferred == nil {
		t.Fatal("expected preferred cleaner")
	}

	if preferred.ID != 12 {
		t.Fatalf(
			"expected ID 12, got %d",
			preferred.ID,
		)
	}

	if preferred.ClientID != 5 {
		t.Fatalf(
			"expected client ID 5, got %d",
			preferred.ClientID,
		)
	}

	if preferred.CleanerID != 8 {
		t.Fatalf(
			"expected cleaner ID 8, got %d",
			preferred.CleanerID,
		)
	}

	if !blockChecked {
		t.Fatal("expected block checker called")
	}

	if !existsChecked {
		t.Fatal("expected exists check called")
	}

	if !createCalled {
		t.Fatal("expected create called")
	}
}

func TestService_Create_WithoutBlockChecker(t *testing.T) {
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
			*PreferredCleaners,
		) error {
			return nil
		},
	}

	service := NewService(repo, nil)

	preferred, err := service.Create(
		context.Background(),
		5,
		CreatePreferredCleanerRequest{
			CleanerID: 8,
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if preferred == nil {
		t.Fatal("expected preferred cleaner")
	}
}

func TestService_Create_InvalidInput(t *testing.T) {
	service := NewService(
		&mockRepository{},
		nil,
	)

	tests := []struct {
		name     string
		clientID uint
		cleaner  uint
	}{
		{
			name:     "zero client ID",
			clientID: 0,
			cleaner:  8,
		},
		{
			name:     "zero cleaner ID",
			clientID: 5,
			cleaner:  0,
		},
		{
			name:     "same client and cleaner",
			clientID: 5,
			cleaner:  5,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			preferred, err := service.Create(
				context.Background(),
				test.clientID,
				CreatePreferredCleanerRequest{
					CleanerID: test.cleaner,
				},
			)

			if preferred != nil {
				t.Fatalf(
					"expected nil preferred cleaner, got %+v",
					preferred,
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

func TestService_Create_BlockCheckerError(t *testing.T) {
	expectedErr := errors.New("block check failed")

	blockChecker := &mockBlockChecker{
		isBlockedFn: func(
			context.Context,
			uint,
			uint,
		) (bool, error) {
			return false, expectedErr
		},
	}

	service := NewService(
		&mockRepository{},
		blockChecker,
	)

	preferred, err := service.Create(
		context.Background(),
		5,
		CreatePreferredCleanerRequest{
			CleanerID: 8,
		},
	)

	if preferred != nil {
		t.Fatalf(
			"expected nil preferred cleaner, got %+v",
			preferred,
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

func TestService_Create_CleanerBlocked(t *testing.T) {
	blockChecker := &mockBlockChecker{
		isBlockedFn: func(
			context.Context,
			uint,
			uint,
		) (bool, error) {
			return true, nil
		},
	}

	service := NewService(
		&mockRepository{},
		blockChecker,
	)

	preferred, err := service.Create(
		context.Background(),
		5,
		CreatePreferredCleanerRequest{
			CleanerID: 8,
		},
	)

	if preferred != nil {
		t.Fatalf(
			"expected nil preferred cleaner, got %+v",
			preferred,
		)
	}

	if !errors.Is(err, ErrCleanerBlocked) {
		t.Fatalf(
			"expected ErrCleanerBlocked, got %v",
			err,
		)
	}
}

func TestService_Create_ExistsError(t *testing.T) {
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

	service := NewService(repo, nil)

	preferred, err := service.Create(
		context.Background(),
		5,
		CreatePreferredCleanerRequest{
			CleanerID: 8,
		},
	)

	if preferred != nil {
		t.Fatalf(
			"expected nil preferred cleaner, got %+v",
			preferred,
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

func TestService_Create_AlreadyPreferred(t *testing.T) {
	repo := &mockRepository{
		existsFn: func(
			context.Context,
			uint,
			uint,
		) (bool, error) {
			return true, nil
		},
	}

	service := NewService(repo, nil)

	preferred, err := service.Create(
		context.Background(),
		5,
		CreatePreferredCleanerRequest{
			CleanerID: 8,
		},
	)

	if preferred != nil {
		t.Fatalf(
			"expected nil preferred cleaner, got %+v",
			preferred,
		)
	}

	if !errors.Is(err, ErrAlreadyPreferred) {
		t.Fatalf(
			"expected ErrAlreadyPreferred, got %v",
			err,
		)
	}
}

func TestService_Create_RepositoryError(t *testing.T) {
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
			*PreferredCleaners,
		) error {
			return expectedErr
		},
	}

	service := NewService(repo, nil)

	preferred, err := service.Create(
		context.Background(),
		5,
		CreatePreferredCleanerRequest{
			CleanerID: 8,
		},
	)

	if preferred != nil {
		t.Fatalf(
			"expected nil preferred cleaner, got %+v",
			preferred,
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
	repo := &mockRepository{
		listByClientIDFn: func(
			ctx context.Context,
			clientID uint,
		) ([]PreferredCleaners, error) {
			if clientID != 5 {
				t.Fatalf(
					"expected client ID 5, got %d",
					clientID,
				)
			}

			return []PreferredCleaners{
				{
					ID:        1,
					ClientID:  5,
					CleanerID: 8,
				},
				{
					ID:        2,
					ClientID:  5,
					CleanerID: 9,
				},
			}, nil
		},
	}

	service := NewService(repo, nil)

	cleaners, err := service.ListMine(
		context.Background(),
		5,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(cleaners) != 2 {
		t.Fatalf(
			"expected 2 cleaners, got %d",
			len(cleaners),
		)
	}
}

func TestService_ListMine_InvalidInput(t *testing.T) {
	service := NewService(
		&mockRepository{},
		nil,
	)

	cleaners, err := service.ListMine(
		context.Background(),
		0,
	)

	if cleaners != nil {
		t.Fatalf(
			"expected nil cleaners, got %+v",
			cleaners,
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
		) ([]PreferredCleaners, error) {
			return nil, expectedErr
		},
	}

	service := NewService(repo, nil)

	cleaners, err := service.ListMine(
		context.Background(),
		5,
	)

	if cleaners != nil {
		t.Fatalf(
			"expected nil cleaners, got %+v",
			cleaners,
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

func TestService_Remove_Success(t *testing.T) {
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

	service := NewService(repo, nil)

	err := service.Remove(
		context.Background(),
		5,
		8,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !deleteCalled {
		t.Fatal("expected delete called")
	}
}

func TestService_Remove_InvalidInput(t *testing.T) {
	service := NewService(
		&mockRepository{},
		nil,
	)

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
			err := service.Remove(
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

func TestService_Remove_RepositoryError(t *testing.T) {
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

	service := NewService(repo, nil)

	err := service.Remove(
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
