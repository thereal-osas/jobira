package profilemedia

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/rodrigueghenda/jobira/internal/security/identity"
)

type profileMediaHandlerTestRepository struct {
	createFn func(
		ctx context.Context,
		media *ProfileMedia,
	) error

	getByIDFn func(
		ctx context.Context,
		id uint,
	) (*ProfileMedia, error)

	listByUserIDFn func(
		ctx context.Context,
		userID uint,
	) ([]ProfileMedia, error)

	listByCompanyIDFn func(
		ctx context.Context,
		companyID uint,
	) ([]ProfileMedia, error)

	countByUserAndTypeFn func(
		ctx context.Context,
		userID uint,
		mediaType MediaType,
	) (int, error)

	countByCompanyAndTypeFn func(
		ctx context.Context,
		companyID uint,
		mediaType MediaType,
	) (int, error)

	canManageCompanyFn func(
		ctx context.Context,
		userID uint,
		companyID uint,
	) (bool, error)

	updateFn func(
		ctx context.Context,
		media *ProfileMedia,
	) error

	deleteFn func(
		ctx context.Context,
		id uint,
	) error
}

func (m *profileMediaHandlerTestRepository) Create(
	ctx context.Context,
	media *ProfileMedia,
) error {
	if m.createFn != nil {
		return m.createFn(ctx, media)
	}

	return nil
}

func (m *profileMediaHandlerTestRepository) GetByID(
	ctx context.Context,
	id uint,
) (*ProfileMedia, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}

	return nil, ErrMediaNotFound
}

func (m *profileMediaHandlerTestRepository) ListByUserID(
	ctx context.Context,
	userID uint,
) ([]ProfileMedia, error) {
	if m.listByUserIDFn != nil {
		return m.listByUserIDFn(ctx, userID)
	}

	return []ProfileMedia{}, nil
}

func (m *profileMediaHandlerTestRepository) ListByCompanyID(
	ctx context.Context,
	companyID uint,
) ([]ProfileMedia, error) {
	if m.listByCompanyIDFn != nil {
		return m.listByCompanyIDFn(ctx, companyID)
	}

	return []ProfileMedia{}, nil
}

func (m *profileMediaHandlerTestRepository) CountByUserAndType(
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

func (m *profileMediaHandlerTestRepository) CountByCompanyAndType(
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

func (m *profileMediaHandlerTestRepository) CanManageCompany(
	ctx context.Context,
	userID uint,
	companyID uint,
) (bool, error) {
	if m.canManageCompanyFn != nil {
		return m.canManageCompanyFn(
			ctx,
			userID,
			companyID,
		)
	}

	return true, nil
}

func (m *profileMediaHandlerTestRepository) Update(
	ctx context.Context,
	media *ProfileMedia,
) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, media)
	}

	return nil
}

func (m *profileMediaHandlerTestRepository) Delete(
	ctx context.Context,
	id uint,
) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}

	return nil
}

func newProfileMediaHandlerForTest(
	repo Repository,
) *Handler {
	service := NewService(repo)

	return NewHandler(service)
}

func requestWithProfileMediaUser(
	req *http.Request,
	userID uint,
	role string,
) *http.Request {
	ctx := identity.WithUser(
		req.Context(),
		identity.UserIdentity{
			UserID: userID,
			Role:   role,
		},
	)

	return req.WithContext(ctx)
}
func requestWithChiURLParam(
	req *http.Request,
	key string,
	value string,
) *http.Request {
	routeCtx := chi.RouteContext(
		req.Context(),
	)

	if routeCtx == nil {
		routeCtx = chi.NewRouteContext()

		ctx := context.WithValue(
			req.Context(),
			chi.RouteCtxKey,
			routeCtx,
		)

		req = req.WithContext(ctx)
	}

	routeCtx.URLParams.Add(
		key,
		value,
	)

	return req
}

func TestHandler_CreateMine_Success(
	t *testing.T,
) {
	repo := &profileMediaHandlerTestRepository{
		countByUserAndTypeFn: func(
			_ context.Context,
			userID uint,
			mediaType MediaType,
		) (int, error) {
			if userID != 5 {
				t.Fatalf(
					"expected user ID 5, got %d",
					userID,
				)
			}

			if mediaType != MediaTypePortfolio {
				t.Fatalf(
					"expected portfolio media type, got %s",
					mediaType,
				)
			}

			return 0, nil
		},
		createFn: func(
			_ context.Context,
			media *ProfileMedia,
		) error {
			if media.OwnerUserID == nil {
				t.Fatal(
					"expected owner user ID",
				)
			}

			if *media.OwnerUserID != 5 {
				t.Fatalf(
					"expected owner user ID 5, got %d",
					*media.OwnerUserID,
				)
			}

			if media.CompanyID != nil {
				t.Fatal(
					"expected company ID to be nil",
				)
			}

			if media.MediaType !=
				MediaTypePortfolio {
				t.Fatalf(
					"expected portfolio media type, got %s",
					media.MediaType,
				)
			}

			media.ID = 10

			return nil
		},
	}

	handler := newProfileMediaHandlerForTest(
		repo,
	)

	body := `{
		"media_type":"portfolio",
		"url":"https://example.com/work.jpg",
		"caption":"Kitchen deep clean",
		"sort_order":1
	}`

	req := httptest.NewRequest(
		http.MethodPost,
		"/profile-media/me",
		bytes.NewBufferString(body),
	)

	req = requestWithProfileMediaUser(
		req,
		5,
		"cleaner",
	)

	recorder := httptest.NewRecorder()

	handler.CreateMine(
		recorder,
		req,
	)

	if recorder.Code !=
		http.StatusCreated {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusCreated,
			recorder.Code,
			recorder.Body.String(),
		)
	}

	if !strings.Contains(
		recorder.Body.String(),
		`"id":10`,
	) {
		t.Fatalf(
			"expected created media ID in response: %s",
			recorder.Body.String(),
		)
	}
}

func TestHandler_CreateMine_Unauthorized(
	t *testing.T,
) {
	handler := newProfileMediaHandlerForTest(
		&profileMediaHandlerTestRepository{},
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/profile-media/me",
		bytes.NewBufferString(`{
			"media_type":"portfolio",
			"url":"https://example.com/work.jpg"
		}`),
	)

	recorder := httptest.NewRecorder()

	handler.CreateMine(
		recorder,
		req,
	)

	if recorder.Code !=
		http.StatusUnauthorized {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusUnauthorized,
			recorder.Code,
		)
	}
}

func TestHandler_CreateMine_InvalidBody(
	t *testing.T,
) {
	handler := newProfileMediaHandlerForTest(
		&profileMediaHandlerTestRepository{},
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/profile-media/me",
		bytes.NewBufferString(`{`),
	)

	req = requestWithProfileMediaUser(
		req,
		5,
		"cleaner",
	)

	recorder := httptest.NewRecorder()

	handler.CreateMine(
		recorder,
		req,
	)

	if recorder.Code !=
		http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			recorder.Code,
		)
	}
}

func TestHandler_CreateMine_PortfolioLimitReached(
	t *testing.T,
) {
	repo := &profileMediaHandlerTestRepository{
		countByUserAndTypeFn: func(
			context.Context,
			uint,
			MediaType,
		) (int, error) {
			return 6, nil
		},
	}

	handler := newProfileMediaHandlerForTest(
		repo,
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/profile-media/me",
		bytes.NewBufferString(`{
			"media_type":"portfolio",
			"url":"https://example.com/work.jpg"
		}`),
	)

	req = requestWithProfileMediaUser(
		req,
		5,
		"cleaner",
	)

	recorder := httptest.NewRecorder()

	handler.CreateMine(
		recorder,
		req,
	)

	if recorder.Code !=
		http.StatusConflict {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusConflict,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_ListMine_Success(
	t *testing.T,
) {
	userID := uint(5)

	repo := &profileMediaHandlerTestRepository{
		listByUserIDFn: func(
			_ context.Context,
			gotUserID uint,
		) ([]ProfileMedia, error) {
			if gotUserID != userID {
				t.Fatalf(
					"expected user ID %d, got %d",
					userID,
					gotUserID,
				)
			}

			return []ProfileMedia{
				{
					ID:          10,
					OwnerUserID: &userID,
					MediaType:   MediaTypePortfolio,
					URL:         "https://example.com/work.jpg",
				},
			}, nil
		},
	}

	handler := newProfileMediaHandlerForTest(
		repo,
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/profile-media/me",
		nil,
	)

	req = requestWithProfileMediaUser(
		req,
		userID,
		"cleaner",
	)

	recorder := httptest.NewRecorder()

	handler.ListMine(
		recorder,
		req,
	)

	if recorder.Code !=
		http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusOK,
			recorder.Code,
			recorder.Body.String(),
		)
	}

	if !strings.Contains(
		recorder.Body.String(),
		`"portfolio"`,
	) {
		t.Fatalf(
			"expected portfolio media in body: %s",
			recorder.Body.String(),
		)
	}
}

func TestHandler_ListByUser_Success(
	t *testing.T,
) {
	userID := uint(12)

	repo := &profileMediaHandlerTestRepository{
		listByUserIDFn: func(
			_ context.Context,
			gotUserID uint,
		) ([]ProfileMedia, error) {
			if gotUserID != userID {
				t.Fatalf(
					"expected user ID %d, got %d",
					userID,
					gotUserID,
				)
			}

			return []ProfileMedia{
				{
					ID:          3,
					OwnerUserID: &userID,
					MediaType:   MediaTypeProfilePhoto,
					URL:         "https://example.com/profile.jpg",
				},
			}, nil
		},
	}

	handler := newProfileMediaHandlerForTest(
		repo,
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/profile-media/users/12",
		nil,
	)

	req = requestWithChiURLParam(
		req,
		"userID",
		"12",
	)

	recorder := httptest.NewRecorder()

	handler.ListByUser(
		recorder,
		req,
	)

	if recorder.Code !=
		http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusOK,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_ListByUser_InvalidID(
	t *testing.T,
) {
	handler := newProfileMediaHandlerForTest(
		&profileMediaHandlerTestRepository{},
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/profile-media/users/nope",
		nil,
	)

	req = requestWithChiURLParam(
		req,
		"userID",
		"nope",
	)

	recorder := httptest.NewRecorder()

	handler.ListByUser(
		recorder,
		req,
	)

	if recorder.Code !=
		http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			recorder.Code,
		)
	}
}

func TestHandler_UpdateMine_Success(
	t *testing.T,
) {
	userID := uint(5)

	repo := &profileMediaHandlerTestRepository{
		getByIDFn: func(
			context.Context,
			uint,
		) (*ProfileMedia, error) {
			return &ProfileMedia{
				ID:          10,
				OwnerUserID: &userID,
				MediaType:   MediaTypePortfolio,
				URL:         "https://example.com/work.jpg",
			}, nil
		},

		updateFn: func(
			_ context.Context,
			media *ProfileMedia,
		) error {
			if media.Caption !=
				"Updated work" {
				t.Fatalf(
					"unexpected caption %q",
					media.Caption,
				)
			}

			if media.SortOrder != 2 {
				t.Fatalf(
					"expected sort order 2, got %d",
					media.SortOrder,
				)
			}

			return nil
		},
	}

	handler := newProfileMediaHandlerForTest(
		repo,
	)

	req := httptest.NewRequest(
		http.MethodPut,
		"/profile-media/me/10",
		bytes.NewBufferString(`{
			"caption":"Updated work",
			"sort_order":2
		}`),
	)

	req = requestWithProfileMediaUser(
		req,
		userID,
		"cleaner",
	)

	req = requestWithChiURLParam(
		req,
		"mediaID",
		"10",
	)

	recorder := httptest.NewRecorder()

	handler.UpdateMine(
		recorder,
		req,
	)

	if recorder.Code !=
		http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusOK,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_UpdateMine_Forbidden(
	t *testing.T,
) {
	ownerID := uint(88)

	repo := &profileMediaHandlerTestRepository{
		getByIDFn: func(
			context.Context,
			uint,
		) (*ProfileMedia, error) {
			return &ProfileMedia{
				ID:          10,
				OwnerUserID: &ownerID,
				MediaType:   MediaTypePortfolio,
			}, nil
		},
	}

	handler := newProfileMediaHandlerForTest(
		repo,
	)

	req := httptest.NewRequest(
		http.MethodPut,
		"/profile-media/me/10",
		bytes.NewBufferString(`{
			"caption":"Attempted edit",
			"sort_order":1
		}`),
	)

	req = requestWithProfileMediaUser(
		req,
		5,
		"cleaner",
	)

	req = requestWithChiURLParam(
		req,
		"mediaID",
		"10",
	)

	recorder := httptest.NewRecorder()

	handler.UpdateMine(
		recorder,
		req,
	)

	if recorder.Code !=
		http.StatusForbidden {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusForbidden,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_DeleteMine_Success(
	t *testing.T,
) {
	userID := uint(5)
	deleteCalled := false

	repo := &profileMediaHandlerTestRepository{
		getByIDFn: func(
			context.Context,
			uint,
		) (*ProfileMedia, error) {
			return &ProfileMedia{
				ID:          10,
				OwnerUserID: &userID,
				MediaType:   MediaTypePortfolio,
			}, nil
		},

		deleteFn: func(
			_ context.Context,
			id uint,
		) error {
			if id != 10 {
				t.Fatalf(
					"expected media ID 10, got %d",
					id,
				)
			}

			deleteCalled = true

			return nil
		},
	}

	handler := newProfileMediaHandlerForTest(
		repo,
	)

	req := httptest.NewRequest(
		http.MethodDelete,
		"/profile-media/me/10",
		nil,
	)

	req = requestWithProfileMediaUser(
		req,
		userID,
		"cleaner",
	)

	req = requestWithChiURLParam(
		req,
		"mediaID",
		"10",
	)

	recorder := httptest.NewRecorder()

	handler.DeleteMine(
		recorder,
		req,
	)

	if recorder.Code !=
		http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusOK,
			recorder.Code,
			recorder.Body.String(),
		)
	}

	if !deleteCalled {
		t.Fatal(
			"expected delete repository method to be called",
		)
	}
}

func TestHandler_CreateForCompany_Success(
	t *testing.T,
) {
	companyID := uint(7)

	repo := &profileMediaHandlerTestRepository{
		canManageCompanyFn: func(
			_ context.Context,
			userID uint,
			gotCompanyID uint,
		) (bool, error) {
			if userID != 99 {
				t.Fatalf(
					"expected user ID 99, got %d",
					userID,
				)
			}

			if gotCompanyID != companyID {
				t.Fatalf(
					"expected company ID %d, got %d",
					companyID,
					gotCompanyID,
				)
			}

			return true, nil
		},

		countByCompanyAndTypeFn: func(
			context.Context,
			uint,
			MediaType,
		) (int, error) {
			return 0, nil
		},

		createFn: func(
			_ context.Context,
			media *ProfileMedia,
		) error {
			if media.CompanyID == nil {
				t.Fatal(
					"expected company ID",
				)
			}

			if *media.CompanyID !=
				companyID {
				t.Fatalf(
					"expected company ID %d, got %d",
					companyID,
					*media.CompanyID,
				)
			}

			media.ID = 50

			return nil
		},
	}

	handler := newProfileMediaHandlerForTest(
		repo,
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/profile-media/companies/7",
		bytes.NewBufferString(`{
			"media_type":"company_logo",
			"url":"https://example.com/logo.jpg"
		}`),
	)

	req = requestWithProfileMediaUser(
		req,
		99,
		"client",
	)

	req = requestWithChiURLParam(
		req,
		"companyID",
		"7",
	)

	recorder := httptest.NewRecorder()

	handler.CreateForCompany(
		recorder,
		req,
	)

	if recorder.Code !=
		http.StatusCreated {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusCreated,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_CreateForCompany_Forbidden(
	t *testing.T,
) {
	repo := &profileMediaHandlerTestRepository{
		canManageCompanyFn: func(
			context.Context,
			uint,
			uint,
		) (bool, error) {
			return false, nil
		},
	}

	handler := newProfileMediaHandlerForTest(
		repo,
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/profile-media/companies/7",
		bytes.NewBufferString(`{
			"media_type":"company_logo",
			"url":"https://example.com/logo.jpg"
		}`),
	)

	req = requestWithProfileMediaUser(
		req,
		99,
		"client",
	)

	req = requestWithChiURLParam(
		req,
		"companyID",
		"7",
	)

	recorder := httptest.NewRecorder()

	handler.CreateForCompany(
		recorder,
		req,
	)

	if recorder.Code !=
		http.StatusForbidden {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusForbidden,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_ListByCompany_Success(
	t *testing.T,
) {
	companyID := uint(7)

	repo := &profileMediaHandlerTestRepository{
		listByCompanyIDFn: func(
			_ context.Context,
			gotCompanyID uint,
		) ([]ProfileMedia, error) {
			if gotCompanyID != companyID {
				t.Fatalf(
					"expected company ID %d, got %d",
					companyID,
					gotCompanyID,
				)
			}

			return []ProfileMedia{
				{
					ID:        22,
					CompanyID: &companyID,
					MediaType: MediaTypeCompanyLogo,
					URL:       "https://example.com/logo.jpg",
				},
			}, nil
		},
	}

	handler := newProfileMediaHandlerForTest(
		repo,
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/profile-media/companies/7",
		nil,
	)

	req = requestWithChiURLParam(
		req,
		"companyID",
		"7",
	)

	recorder := httptest.NewRecorder()

	handler.ListByCompany(
		recorder,
		req,
	)

	if recorder.Code !=
		http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusOK,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_UpdateForCompany_Success(
	t *testing.T,
) {
	companyID := uint(7)

	repo := &profileMediaHandlerTestRepository{
		canManageCompanyFn: func(
			context.Context,
			uint,
			uint,
		) (bool, error) {
			return true, nil
		},

		getByIDFn: func(
			context.Context,
			uint,
		) (*ProfileMedia, error) {
			return &ProfileMedia{
				ID:        30,
				CompanyID: &companyID,
				MediaType: MediaTypePortfolio,
				URL:       "https://example.com/company-work.jpg",
			}, nil
		},

		updateFn: func(
			_ context.Context,
			media *ProfileMedia,
		) error {
			if media.Caption !=
				"Updated company work" {
				t.Fatalf(
					"unexpected caption %q",
					media.Caption,
				)
			}

			return nil
		},
	}

	handler := newProfileMediaHandlerForTest(
		repo,
	)

	req := httptest.NewRequest(
		http.MethodPut,
		"/profile-media/companies/7/30",
		bytes.NewBufferString(`{
			"caption":"Updated company work",
			"sort_order":2
		}`),
	)

	req = requestWithProfileMediaUser(
		req,
		99,
		"client",
	)

	req = requestWithChiURLParam(
		req,
		"companyID",
		"7",
	)

	req = requestWithChiURLParam(
		req,
		"mediaID",
		"30",
	)

	recorder := httptest.NewRecorder()

	handler.UpdateForCompany(
		recorder,
		req,
	)

	if recorder.Code !=
		http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusOK,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_DeleteForCompany_Success(
	t *testing.T,
) {
	companyID := uint(7)
	deleteCalled := false

	repo := &profileMediaHandlerTestRepository{
		canManageCompanyFn: func(
			context.Context,
			uint,
			uint,
		) (bool, error) {
			return true, nil
		},

		getByIDFn: func(
			context.Context,
			uint,
		) (*ProfileMedia, error) {
			return &ProfileMedia{
				ID:        30,
				CompanyID: &companyID,
				MediaType: MediaTypePortfolio,
			}, nil
		},

		deleteFn: func(
			_ context.Context,
			id uint,
		) error {
			if id != 30 {
				t.Fatalf(
					"expected media ID 30, got %d",
					id,
				)
			}

			deleteCalled = true

			return nil
		},
	}

	handler := newProfileMediaHandlerForTest(
		repo,
	)

	req := httptest.NewRequest(
		http.MethodDelete,
		"/profile-media/companies/7/30",
		nil,
	)

	req = requestWithProfileMediaUser(
		req,
		99,
		"client",
	)

	req = requestWithChiURLParam(
		req,
		"companyID",
		"7",
	)

	req = requestWithChiURLParam(
		req,
		"mediaID",
		"30",
	)

	recorder := httptest.NewRecorder()

	handler.DeleteForCompany(
		recorder,
		req,
	)

	if recorder.Code !=
		http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusOK,
			recorder.Code,
			recorder.Body.String(),
		)
	}

	if !deleteCalled {
		t.Fatal(
			"expected delete repository method to be called",
		)
	}
}

func TestHandler_ServiceRepositoryError(
	t *testing.T,
) {
	repositoryErr := errors.New(
		"database failed",
	)

	repo := &profileMediaHandlerTestRepository{
		listByUserIDFn: func(
			context.Context,
			uint,
		) ([]ProfileMedia, error) {
			return nil, repositoryErr
		},
	}

	handler := newProfileMediaHandlerForTest(
		repo,
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/profile-media/me",
		nil,
	)

	req = requestWithProfileMediaUser(
		req,
		5,
		"cleaner",
	)

	recorder := httptest.NewRecorder()

	handler.ListMine(
		recorder,
		req,
	)

	if recorder.Code !=
		http.StatusInternalServerError {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusInternalServerError,
			recorder.Code,
		)
	}
}

var _ Repository = (*profileMediaHandlerTestRepository)(nil)
