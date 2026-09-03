package favorites

import (
	"context"
	"errors"
	notificationsdomain "github.com/rodrigueghenda/jobira/internal/domain/notifications"
	"testing"
	"time"
)

type mockRepository struct {
	createFn         func(context.Context, *FavoriteCleaner) error
	deleteFn         func(context.Context, uint, uint) error
	listByClientIDFn func(context.Context, uint) ([]FavoriteCleaner, error)
	existsFn         func(context.Context, uint, uint) (bool, error)
}

func (m *mockRepository) Create(
	ctx context.Context,
	favorite *FavoriteCleaner,
) error {
	if m.createFn != nil {
		return m.createFn(ctx, favorite)
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
) ([]FavoriteCleaner, error) {
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

type mockNotificationsService struct {
	createFn func(
		context.Context,
		notificationsdomain.CreateNotificationsRequest,
	) (*notificationsdomain.Notification, error)
}

func (m *mockNotificationsService) Create(
	ctx context.Context,
	req notificationsdomain.CreateNotificationsRequest,
) (*notificationsdomain.Notification, error) {
	if m.createFn != nil {
		return m.createFn(ctx, req)
	}

	return nil, nil
}

func TestNewService(t *testing.T) {
	repo := &mockRepository{}
	blockChecker := &mockBlockChecker{}

	service := NewService(
		repo,
		nil,
		blockChecker,
	)

	if service == nil {
		t.Fatal("expected service")
	}

	if service.repo != repo {
		t.Fatal("expected repository assigned")
	}

	if service.notificationsService != nil {
		t.Fatal("expected nil notifications service")
	}

	if service.blockChecker != blockChecker {
		t.Fatal("expected block checker assigned")
	}
}

func TestService_Save_Success(t *testing.T) {
	existsCalled := false
	blockCalled := false
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
			favorite *FavoriteCleaner,
		) error {
			createCalled = true

			if favorite.ClientID != 5 {
				t.Fatalf(
					"expected client ID 5, got %d",
					favorite.ClientID,
				)
			}

			if favorite.CleanerID != 8 {
				t.Fatalf(
					"expected cleaner ID 8, got %d",
					favorite.CleanerID,
				)
			}

			favorite.ID = 12
			favorite.CreatedAt = time.Now()

			return nil
		},
	}

	blockChecker := &mockBlockChecker{
		isBlockedFn: func(
			ctx context.Context,
			clientID uint,
			cleanerID uint,
		) (bool, error) {
			blockCalled = true

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
	}

	service := NewService(
		repo,
		nil,
		blockChecker,
	)

	favorite, err := service.Save(
		context.Background(),
		5,
		CreateFavoriteRequest{
			CleanerID: 8,
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if favorite == nil {
		t.Fatal("expected favorite")
	}

	if favorite.ID != 12 {
		t.Fatalf(
			"expected ID 12, got %d",
			favorite.ID,
		)
	}

	if favorite.ClientID != 5 {
		t.Fatalf(
			"expected client ID 5, got %d",
			favorite.ClientID,
		)
	}

	if favorite.CleanerID != 8 {
		t.Fatalf(
			"expected cleaner ID 8, got %d",
			favorite.CleanerID,
		)
	}

	if favorite.CreatedAt.IsZero() {
		t.Fatal("expected created_at")
	}

	if !blockCalled {
		t.Fatal("expected block checker call")
	}

	if !existsCalled {
		t.Fatal("expected Exists repository call")
	}

	if !createCalled {
		t.Fatal("expected Create repository call")
	}
}

func TestService_Save_WithoutBlockChecker(t *testing.T) {
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
			*FavoriteCleaner,
		) error {
			return nil
		},
	}

	service := NewService(repo, nil, nil)

	favorite, err := service.Save(
		context.Background(),
		5,
		CreateFavoriteRequest{
			CleanerID: 8,
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if favorite == nil {
		t.Fatal("expected favorite")
	}
}

func TestService_Save_InvalidInput(t *testing.T) {
	service := NewService(
		&mockRepository{},
		nil,
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
		{
			name:      "same client and cleaner",
			clientID:  5,
			cleanerID: 5,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			favorite, err := service.Save(
				context.Background(),
				test.clientID,
				CreateFavoriteRequest{
					CleanerID: test.cleanerID,
				},
			)

			if favorite != nil {
				t.Fatalf(
					"expected nil favorite, got %+v",
					favorite,
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

func TestService_Save_BlockCheckerError(t *testing.T) {
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
		nil,
		blockChecker,
	)

	favorite, err := service.Save(
		context.Background(),
		5,
		CreateFavoriteRequest{
			CleanerID: 8,
		},
	)

	if favorite != nil {
		t.Fatalf(
			"expected nil favorite, got %+v",
			favorite,
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

func TestService_Save_CleanerBlocked(t *testing.T) {
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
		nil,
		blockChecker,
	)

	favorite, err := service.Save(
		context.Background(),
		5,
		CreateFavoriteRequest{
			CleanerID: 8,
		},
	)

	if favorite != nil {
		t.Fatalf(
			"expected nil favorite, got %+v",
			favorite,
		)
	}

	if !errors.Is(err, ErrCleanerBlocked) {
		t.Fatalf(
			"expected ErrCleanerBlocked, got %v",
			err,
		)
	}
}

func TestService_Save_ExistsError(t *testing.T) {
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

	service := NewService(repo, nil, nil)

	favorite, err := service.Save(
		context.Background(),
		5,
		CreateFavoriteRequest{
			CleanerID: 8,
		},
	)

	if favorite != nil {
		t.Fatalf(
			"expected nil favorite, got %+v",
			favorite,
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

func TestService_Save_AlreadySaved(t *testing.T) {
	repo := &mockRepository{
		existsFn: func(
			context.Context,
			uint,
			uint,
		) (bool, error) {
			return true, nil
		},
	}

	service := NewService(repo, nil, nil)

	favorite, err := service.Save(
		context.Background(),
		5,
		CreateFavoriteRequest{
			CleanerID: 8,
		},
	)

	if favorite != nil {
		t.Fatalf(
			"expected nil favorite, got %+v",
			favorite,
		)
	}

	if !errors.Is(err, ErrAlreadySaved) {
		t.Fatalf(
			"expected ErrAlreadySaved, got %v",
			err,
		)
	}
}

func TestService_Save_CreateRepositoryError(t *testing.T) {
	expectedErr := errors.New("create favorite failed")

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
			*FavoriteCleaner,
		) error {
			return expectedErr
		},
	}

	service := NewService(repo, nil, nil)

	favorite, err := service.Save(
		context.Background(),
		5,
		CreateFavoriteRequest{
			CleanerID: 8,
		},
	)

	if favorite != nil {
		t.Fatalf(
			"expected nil favorite, got %+v",
			favorite,
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
		listByClientIDFn: func(
			ctx context.Context,
			clientID uint,
		) ([]FavoriteCleaner, error) {
			if clientID != 5 {
				t.Fatalf(
					"expected client ID 5, got %d",
					clientID,
				)
			}

			return []FavoriteCleaner{
				{
					ID:        1,
					ClientID:  5,
					CleanerID: 8,
					CreatedAt: now,
				},
				{
					ID:        2,
					ClientID:  5,
					CleanerID: 9,
					CreatedAt: now.Add(-time.Hour),
				},
			}, nil
		},
	}

	service := NewService(repo, nil, nil)

	favorites, err := service.ListMine(
		context.Background(),
		5,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(favorites) != 2 {
		t.Fatalf(
			"expected 2 favorites, got %d",
			len(favorites),
		)
	}

	if favorites[0].CleanerID != 8 {
		t.Fatalf(
			"expected first cleaner ID 8, got %d",
			favorites[0].CleanerID,
		)
	}

	if favorites[1].CleanerID != 9 {
		t.Fatalf(
			"expected second cleaner ID 9, got %d",
			favorites[1].CleanerID,
		)
	}
}

func TestService_ListMine_Empty(t *testing.T) {
	repo := &mockRepository{
		listByClientIDFn: func(
			context.Context,
			uint,
		) ([]FavoriteCleaner, error) {
			return []FavoriteCleaner{}, nil
		},
	}

	service := NewService(repo, nil, nil)

	favorites, err := service.ListMine(
		context.Background(),
		5,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(favorites) != 0 {
		t.Fatalf(
			"expected no favorites, got %d",
			len(favorites),
		)
	}
}

func TestService_ListMine_InvalidInput(t *testing.T) {
	service := NewService(
		&mockRepository{},
		nil,
		nil,
	)

	favorites, err := service.ListMine(
		context.Background(),
		0,
	)

	if favorites != nil {
		t.Fatalf(
			"expected nil favorites, got %+v",
			favorites,
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
	expectedErr := errors.New("list favorites failed")

	repo := &mockRepository{
		listByClientIDFn: func(
			context.Context,
			uint,
		) ([]FavoriteCleaner, error) {
			return nil, expectedErr
		},
	}

	service := NewService(repo, nil, nil)

	favorites, err := service.ListMine(
		context.Background(),
		5,
	)

	if favorites != nil {
		t.Fatalf(
			"expected nil favorites, got %+v",
			favorites,
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

	service := NewService(repo, nil, nil)

	err := service.Remove(
		context.Background(),
		5,
		8,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !deleteCalled {
		t.Fatal("expected Delete repository call")
	}
}

func TestService_Remove_InvalidInput(t *testing.T) {
	service := NewService(
		&mockRepository{},
		nil,
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

func TestService_Remove_NotFound(t *testing.T) {
	repo := &mockRepository{
		deleteFn: func(
			context.Context,
			uint,
			uint,
		) error {
			return ErrFavoriteNotFound
		},
	}

	service := NewService(repo, nil, nil)

	err := service.Remove(
		context.Background(),
		5,
		8,
	)

	if !errors.Is(err, ErrFavoriteNotFound) {
		t.Fatalf(
			"expected ErrFavoriteNotFound, got %v",
			err,
		)
	}
}

func TestService_Remove_RepositoryError(t *testing.T) {
	expectedErr := errors.New("remove favorite failed")

	repo := &mockRepository{
		deleteFn: func(
			context.Context,
			uint,
			uint,
		) error {
			return expectedErr
		},
	}

	service := NewService(repo, nil, nil)

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

func TestService_Save_WithNotification(t *testing.T) {
	notificationCalled := false

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
			*FavoriteCleaner,
		) error {
			return nil
		},
	}

	notificationsService := &mockNotificationsService{
		createFn: func(
			ctx context.Context,
			req notificationsdomain.CreateNotificationsRequest,
		) (*notificationsdomain.Notification, error) {
			notificationCalled = true

			if req.UserID != 8 {
				t.Fatalf(
					"expected notification user ID 8, got %d",
					req.UserID,
				)
			}

			if req.Title != "New Favourite" {
				t.Fatalf(
					"expected title New Favourite, got %q",
					req.Title,
				)
			}

			if req.Message != "A Client added you to their favourite cleaners list" {
				t.Fatalf(
					"unexpected notification message: %q",
					req.Message,
				)
			}

			if req.Type != "favourite" {
				t.Fatalf(
					"expected type favourite, got %q",
					req.Type,
				)
			}

			return nil, nil
		},
	}

	service := NewService(
		repo,
		notificationsService,
		nil,
	)

	favorite, err := service.Save(
		context.Background(),
		5,
		CreateFavoriteRequest{
			CleanerID: 8,
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if favorite == nil {
		t.Fatal("expected favorite")
	}

	if !notificationCalled {
		t.Fatal("expected notification service to be called")
	}
}
