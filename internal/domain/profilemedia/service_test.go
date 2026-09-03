package profilemedia

import (
	"context"
	"errors"
	"testing"
)

type mockRepository struct {
	createFn                func(context.Context, *ProfileMedia) error
	getByIDFn               func(context.Context, uint) (*ProfileMedia, error)
	listByUserIDFn          func(context.Context, uint) ([]ProfileMedia, error)
	listByCompanyIDFn       func(context.Context, uint) ([]ProfileMedia, error)
	countByUserAndTypeFn    func(context.Context, uint, MediaType) (int, error)
	countByCompanyAndTypeFn func(context.Context, uint, MediaType) (int, error)
	updateFn                func(context.Context, *ProfileMedia) error
	deleteFn                func(context.Context, uint) error
}

func (m *mockRepository) Create(
	ctx context.Context,
	media *ProfileMedia,
) error {
	if m.createFn != nil {
		return m.createFn(ctx, media)
	}
	return nil
}

func (m *mockRepository) GetByID(
	ctx context.Context,
	id uint,
) (*ProfileMedia, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *mockRepository) ListByUserID(
	ctx context.Context,
	userID uint,
) ([]ProfileMedia, error) {
	if m.listByUserIDFn != nil {
		return m.listByUserIDFn(ctx, userID)
	}
	return nil, nil
}

func (m *mockRepository) ListByCompanyID(
	ctx context.Context,
	companyID uint,
) ([]ProfileMedia, error) {
	if m.listByCompanyIDFn != nil {
		return m.listByCompanyIDFn(ctx, companyID)
	}
	return nil, nil
}

func (m *mockRepository) CountByUserAndType(
	ctx context.Context,
	userID uint,
	mediaType MediaType,
) (int, error) {
	if m.countByUserAndTypeFn != nil {
		return m.countByUserAndTypeFn(
			ctx,
			userID,
			mediaType,
		)
	}
	return 0, nil
}

func (m *mockRepository) CountByCompanyAndType(
	ctx context.Context,
	companyID uint,
	mediaType MediaType,
) (int, error) {
	if m.countByCompanyAndTypeFn != nil {
		return m.countByCompanyAndTypeFn(
			ctx,
			companyID,
			mediaType,
		)
	}
	return 0, nil
}

func (m *mockRepository) Update(
	ctx context.Context,
	media *ProfileMedia,
) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, media)
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

func (m *mockRepository) CanManageCompany(
	ctx context.Context,
	userID uint,
	companyID uint,
) (bool, error) {
	return true, nil
}

func TestService_CreateForUser_Success(t *testing.T) {
	repo := &mockRepository{
		countByUserAndTypeFn: func(
			ctx context.Context,
			userID uint,
			mediaType MediaType,
		) (int, error) {
			if userID != 10 {
				t.Fatalf(
					"expected userID 10 got %d",
					userID,
				)
			}

			if mediaType != MediaTypePortfolio {
				t.Fatalf(
					"expected portfolio got %s",
					mediaType,
				)
			}

			return 2, nil
		},
		createFn: func(
			ctx context.Context,
			media *ProfileMedia,
		) error {
			media.ID = 50
			return nil
		},
	}

	service := NewService(repo)

	result, err := service.CreateForUser(
		context.Background(),
		10,
		CreateMediaRequest{
			MediaType: MediaTypePortfolio,
			URL:       " https://example.com/work.jpg ",
			Caption:   " Kitchen clean ",
			SortOrder: 1,
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ID != 50 {
		t.Fatalf(
			"expected ID 50 got %d",
			result.ID,
		)
	}

	if result.OwnerUserID == nil ||
		*result.OwnerUserID != 10 {
		t.Fatal("expected owner user ID 10")
	}

	if result.CompanyID != nil {
		t.Fatal("expected company ID to be nil")
	}

	if result.URL != "https://example.com/work.jpg" {
		t.Fatalf(
			"unexpected URL: %s",
			result.URL,
		)
	}

	if result.Caption != "Kitchen clean" {
		t.Fatalf(
			"unexpected caption: %s",
			result.Caption,
		)
	}
}

func TestService_CreateForUser_InvalidUser(t *testing.T) {
	service := NewService(&mockRepository{})

	_, err := service.CreateForUser(
		context.Background(),
		0,
		CreateMediaRequest{
			MediaType: MediaTypePortfolio,
			URL:       "https://example.com/photo.jpg",
		},
	)

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf(
			"expected ErrInvalidInput got %v",
			err,
		)
	}
}

func TestService_CreateForUser_EmptyURL(t *testing.T) {
	service := NewService(&mockRepository{})

	_, err := service.CreateForUser(
		context.Background(),
		1,
		CreateMediaRequest{
			MediaType: MediaTypePortfolio,
			URL:       "   ",
		},
	)

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf(
			"expected ErrInvalidInput got %v",
			err,
		)
	}
}

func TestService_CreateForUser_InvalidMediaType(t *testing.T) {
	service := NewService(&mockRepository{})

	_, err := service.CreateForUser(
		context.Background(),
		1,
		CreateMediaRequest{
			MediaType: MediaTypeCompanyLogo,
			URL:       "https://example.com/logo.jpg",
		},
	)

	if !errors.Is(err, ErrInvalidMediaType) {
		t.Fatalf(
			"expected ErrInvalidMediaType got %v",
			err,
		)
	}
}

func TestService_CreateForUser_ProfilePhotoLimit(t *testing.T) {
	repo := &mockRepository{
		countByUserAndTypeFn: func(
			context.Context,
			uint,
			MediaType,
		) (int, error) {
			return 1, nil
		},
	}

	service := NewService(repo)

	_, err := service.CreateForUser(
		context.Background(),
		1,
		CreateMediaRequest{
			MediaType: MediaTypeProfilePhoto,
			URL:       "https://example.com/profile.jpg",
		},
	)

	if !errors.Is(
		err,
		ErrProfilePhotoLimitReached,
	) {
		t.Fatalf(
			"expected profile photo limit error got %v",
			err,
		)
	}
}

func TestService_CreateForUser_PortfolioLimit(t *testing.T) {
	repo := &mockRepository{
		countByUserAndTypeFn: func(
			context.Context,
			uint,
			MediaType,
		) (int, error) {
			return 6, nil
		},
	}

	service := NewService(repo)

	_, err := service.CreateForUser(
		context.Background(),
		1,
		CreateMediaRequest{
			MediaType: MediaTypePortfolio,
			URL:       "https://example.com/work.jpg",
		},
	)

	if !errors.Is(
		err,
		ErrPortfolioLimitReached,
	) {
		t.Fatalf(
			"expected portfolio limit error got %v",
			err,
		)
	}
}

func TestService_CreateForUser_PricingLimit(t *testing.T) {
	repo := &mockRepository{
		countByUserAndTypeFn: func(
			context.Context,
			uint,
			MediaType,
		) (int, error) {
			return 1, nil
		},
	}

	service := NewService(repo)

	_, err := service.CreateForUser(
		context.Background(),
		1,
		CreateMediaRequest{
			MediaType: MediaTypePricing,
			URL:       "https://example.com/pricing.jpg",
		},
	)

	if !errors.Is(
		err,
		ErrPricingLimitReached,
	) {
		t.Fatalf(
			"expected pricing limit error got %v",
			err,
		)
	}
}

func TestService_CreateForUser_CountError(t *testing.T) {
	expectedErr := errors.New("count failed")

	repo := &mockRepository{
		countByUserAndTypeFn: func(
			context.Context,
			uint,
			MediaType,
		) (int, error) {
			return 0, expectedErr
		},
	}

	service := NewService(repo)

	_, err := service.CreateForUser(
		context.Background(),
		1,
		CreateMediaRequest{
			MediaType: MediaTypePortfolio,
			URL:       "https://example.com/work.jpg",
		},
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected repository error got %v",
			err,
		)
	}
}

func TestService_CreateForUser_CreateError(t *testing.T) {
	expectedErr := errors.New("create failed")

	repo := &mockRepository{
		countByUserAndTypeFn: func(
			context.Context,
			uint,
			MediaType,
		) (int, error) {
			return 0, nil
		},
		createFn: func(
			context.Context,
			*ProfileMedia,
		) error {
			return expectedErr
		},
	}

	service := NewService(repo)

	_, err := service.CreateForUser(
		context.Background(),
		1,
		CreateMediaRequest{
			MediaType: MediaTypePortfolio,
			URL:       "https://example.com/work.jpg",
		},
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected repository error got %v",
			err,
		)
	}
}

func TestService_CreateForCompany_Success(t *testing.T) {
	repo := &mockRepository{
		countByCompanyAndTypeFn: func(
			ctx context.Context,
			companyID uint,
			mediaType MediaType,
		) (int, error) {
			if companyID != 7 {
				t.Fatalf(
					"expected companyID 7 got %d",
					companyID,
				)
			}

			return 0, nil
		},
		createFn: func(
			ctx context.Context,
			media *ProfileMedia,
		) error {
			media.ID = 99
			return nil
		},
	}

	service := NewService(repo)

	result, err := service.CreateForCompany(
		context.Background(),
		99,
		7,
		CreateMediaRequest{
			MediaType: MediaTypeCompanyLogo,
			URL:       "https://example.com/logo.jpg",
		},
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ID != 99 {
		t.Fatalf(
			"expected ID 99 got %d",
			result.ID,
		)
	}

	if result.CompanyID == nil ||
		*result.CompanyID != 7 {
		t.Fatal("expected company ID 7")
	}

	if result.OwnerUserID != nil {
		t.Fatal("expected owner user ID to be nil")
	}
}

func TestService_CreateForCompany_ProfilePhotoRejected(
	t *testing.T,
) {
	service := NewService(&mockRepository{})

	_, err := service.CreateForCompany(
		context.Background(),
		99,
		7,
		CreateMediaRequest{
			MediaType: MediaTypeProfilePhoto,
			URL:       "https://example.com/profile.jpg",
		},
	)

	if !errors.Is(err, ErrInvalidMediaType) {
		t.Fatalf(
			"expected ErrInvalidMediaType got %v",
			err,
		)
	}
}

func TestService_CreateForCompany_LogoLimit(t *testing.T) {
	repo := &mockRepository{
		countByCompanyAndTypeFn: func(
			context.Context,
			uint,
			MediaType,
		) (int, error) {
			return 1, nil
		},
	}

	service := NewService(repo)

	_, err := service.CreateForCompany(
		context.Background(),
		99,
		7,
		CreateMediaRequest{
			MediaType: MediaTypeCompanyLogo,
			URL:       "https://example.com/logo.jpg",
		},
	)

	if !errors.Is(
		err,
		ErrCompanyLogoLimitReached,
	) {
		t.Fatalf(
			"expected company logo limit error got %v",
			err,
		)
	}
}

func TestService_ListForUser_Success(t *testing.T) {
	repo := &mockRepository{
		listByUserIDFn: func(
			ctx context.Context,
			userID uint,
		) ([]ProfileMedia, error) {
			return []ProfileMedia{
				{
					ID:        1,
					MediaType: MediaTypePortfolio,
				},
			}, nil
		},
	}

	service := NewService(repo)

	result, err := service.ListForUser(
		context.Background(),
		10,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf(
			"expected 1 media item got %d",
			len(result),
		)
	}
}

func TestService_ListForCompany_Success(t *testing.T) {
	repo := &mockRepository{
		listByCompanyIDFn: func(
			ctx context.Context,
			companyID uint,
		) ([]ProfileMedia, error) {
			return []ProfileMedia{
				{
					ID:        1,
					MediaType: MediaTypeCompanyLogo,
				},
			}, nil
		},
	}

	service := NewService(repo)

	result, err := service.ListForCompany(
		context.Background(),
		5,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf(
			"expected 1 media item got %d",
			len(result),
		)
	}
}

func TestService_UpdateForUser_Success(t *testing.T) {
	userID := uint(10)

	repo := &mockRepository{
		getByIDFn: func(
			context.Context,
			uint,
		) (*ProfileMedia, error) {
			return &ProfileMedia{
				ID:          8,
				OwnerUserID: &userID,
				Caption:     "old",
			}, nil
		},
	}

	service := NewService(repo)

	result, err := service.UpdateForUser(
		context.Background(),
		10,
		8,
		UpdateMediaRequest{
			Caption:   " New caption ",
			SortOrder: 3,
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Caption != "New caption" {
		t.Fatalf(
			"unexpected caption %q",
			result.Caption,
		)
	}

	if result.SortOrder != 3 {
		t.Fatalf(
			"expected sort order 3 got %d",
			result.SortOrder,
		)
	}
}

func TestService_UpdateForUser_Forbidden(t *testing.T) {
	otherUserID := uint(20)

	repo := &mockRepository{
		getByIDFn: func(
			context.Context,
			uint,
		) (*ProfileMedia, error) {
			return &ProfileMedia{
				ID:          8,
				OwnerUserID: &otherUserID,
			}, nil
		},
	}

	service := NewService(repo)

	_, err := service.UpdateForUser(
		context.Background(),
		10,
		8,
		UpdateMediaRequest{},
	)

	if !errors.Is(err, Errforbidden) {
		t.Fatalf(
			"expected ErrForbidden got %v",
			err,
		)
	}
}

func TestService_UpdateForCompany_Forbidden(t *testing.T) {
	otherCompanyID := uint(20)

	repo := &mockRepository{
		getByIDFn: func(
			context.Context,
			uint,
		) (*ProfileMedia, error) {
			return &ProfileMedia{
				ID:        8,
				CompanyID: &otherCompanyID,
			}, nil
		},
	}

	service := NewService(repo)

	_, err := service.UpdateForCompany(
		context.Background(),
		99,
		10,
		8,
		UpdateMediaRequest{},
	)

	if !errors.Is(err, Errforbidden) {
		t.Fatalf(
			"expected ErrForbidden got %v",
			err,
		)
	}
}

func TestService_DeleteForUser_Success(t *testing.T) {
	userID := uint(10)
	deleteCalled := false

	repo := &mockRepository{
		getByIDFn: func(
			context.Context,
			uint,
		) (*ProfileMedia, error) {
			return &ProfileMedia{
				ID:          3,
				OwnerUserID: &userID,
			}, nil
		},
		deleteFn: func(
			ctx context.Context,
			id uint,
		) error {
			deleteCalled = true

			if id != 3 {
				t.Fatalf(
					"expected ID 3 got %d",
					id,
				)
			}

			return nil
		},
	}

	service := NewService(repo)

	err := service.DeleteForUser(
		context.Background(),
		10,
		3,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !deleteCalled {
		t.Fatal("expected delete to be called")
	}
}

func TestService_DeleteForUser_Forbidden(t *testing.T) {
	otherUserID := uint(11)

	repo := &mockRepository{
		getByIDFn: func(
			context.Context,
			uint,
		) (*ProfileMedia, error) {
			return &ProfileMedia{
				ID:          3,
				OwnerUserID: &otherUserID,
			}, nil
		},
	}

	service := NewService(repo)

	err := service.DeleteForUser(
		context.Background(),
		10,
		3,
	)

	if !errors.Is(err, Errforbidden) {
		t.Fatalf(
			"expected ErrForbidden got %v",
			err,
		)
	}
}

func TestService_DeleteForCompany_Success(t *testing.T) {
	companyID := uint(5)

	repo := &mockRepository{
		getByIDFn: func(
			context.Context,
			uint,
		) (*ProfileMedia, error) {
			return &ProfileMedia{
				ID:        12,
				CompanyID: &companyID,
			}, nil
		},
	}

	service := NewService(repo)

	err := service.DeleteForCompany(
		context.Background(),
		99,
		5,
		12,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateMediaLimit(t *testing.T) {
	tests := []struct {
		name      string
		mediaType MediaType
		count     int
		expected  error
	}{
		{
			name:      "profile photo allowed",
			mediaType: MediaTypeProfilePhoto,
			count:     0,
			expected:  nil,
		},
		{
			name:      "profile photo blocked",
			mediaType: MediaTypeProfilePhoto,
			count:     1,
			expected:  ErrProfilePhotoLimitReached,
		},
		{
			name:      "portfolio fifth allowed",
			mediaType: MediaTypePortfolio,
			count:     5,
			expected:  nil,
		},
		{
			name:      "portfolio seventh blocked",
			mediaType: MediaTypePortfolio,
			count:     6,
			expected:  ErrPortfolioLimitReached,
		},
		{
			name:      "pricing blocked",
			mediaType: MediaTypePricing,
			count:     1,
			expected:  ErrPricingLimitReached,
		},
		{
			name:      "company logo blocked",
			mediaType: MediaTypeCompanyLogo,
			count:     1,
			expected:  ErrCompanyLogoLimitReached,
		},
		{
			name:      "unknown media type",
			mediaType: MediaType("banana"),
			count:     0,
			expected:  ErrInvalidMediaType,
		},
	}

	for _, test := range tests {
		t.Run(
			test.name,
			func(t *testing.T) {
				err := validateMediaLimit(
					test.mediaType,
					test.count,
				)

				if !errors.Is(
					err,
					test.expected,
				) {
					t.Fatalf(
						"expected %v got %v",
						test.expected,
						err,
					)
				}
			},
		)
	}
}

var _ Repository = (*mockRepository)(nil)
