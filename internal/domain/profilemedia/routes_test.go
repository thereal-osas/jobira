package profilemedia

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/rodrigueghenda/jobira/internal/security/identity"
)

func profileMediaAuthMiddlewareForTest(
	next http.Handler,
) http.Handler {
	return http.HandlerFunc(
		func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			ctx := identity.WithUser(
				r.Context(),
				identity.UserIdentity{
					UserID: 99,
					Role:   "cleaner",
				},
			)

			next.ServeHTTP(
				w,
				r.WithContext(ctx),
			)
		},
	)
}

func newProfileMediaRouterForTest(
	repo Repository,
) http.Handler {
	service := NewService(repo)
	handler := NewHandler(service)

	router := chi.NewRouter()

	RegisterRoutes(
		router,
		handler,
		profileMediaAuthMiddlewareForTest,
	)

	return router
}

func TestProfileMediaRoutes_PublicUserMedia(
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
					ID:          1,
					OwnerUserID: &userID,
					MediaType:   MediaTypePortfolio,
					URL:         "https://example.com/work.jpg",
				},
			}, nil
		},
	}

	router := newProfileMediaRouterForTest(
		repo,
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/profile-media/users/12",
		nil,
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(
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
			"expected portfolio media in response: %s",
			recorder.Body.String(),
		)
	}
}

func TestProfileMediaRoutes_PublicCompanyMedia(
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
					ID:        2,
					CompanyID: &companyID,
					MediaType: MediaTypeCompanyLogo,
					URL:       "https://example.com/logo.jpg",
				},
			}, nil
		},
	}

	router := newProfileMediaRouterForTest(
		repo,
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/profile-media/companies/7",
		nil,
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(
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

func TestProfileMediaRoutes_CreateMine(
	t *testing.T,
) {
	repo := &profileMediaHandlerTestRepository{
		countByUserAndTypeFn: func(
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
			media.ID = 10

			return nil
		},
	}

	router := newProfileMediaRouterForTest(
		repo,
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/profile-media/me",
		strings.NewReader(`{
			"media_type":"portfolio",
			"url":"https://example.com/work.jpg",
			"caption":"Completed work",
			"sort_order":1
		}`),
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(
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

func TestProfileMediaRoutes_ListMine(
	t *testing.T,
) {
	userID := uint(99)

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
					MediaType:   MediaTypeProfilePhoto,
					URL:         "https://example.com/profile.jpg",
				},
			}, nil
		},
	}

	router := newProfileMediaRouterForTest(
		repo,
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/profile-media/me",
		nil,
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(
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

func TestProfileMediaRoutes_UpdateMine(
	t *testing.T,
) {
	userID := uint(99)

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

		updateFn: func(
			context.Context,
			*ProfileMedia,
		) error {
			return nil
		},
	}

	router := newProfileMediaRouterForTest(
		repo,
	)

	req := httptest.NewRequest(
		http.MethodPut,
		"/profile-media/me/10",
		strings.NewReader(`{
			"caption":"Updated",
			"sort_order":2
		}`),
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(
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

func TestProfileMediaRoutes_DeleteMine(
	t *testing.T,
) {
	userID := uint(99)

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
			context.Context,
			uint,
		) error {
			return nil
		},
	}

	router := newProfileMediaRouterForTest(
		repo,
	)

	req := httptest.NewRequest(
		http.MethodDelete,
		"/profile-media/me/10",
		nil,
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(
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

func TestProfileMediaRoutes_CreateCompany(
	t *testing.T,
) {
	repo := &profileMediaHandlerTestRepository{
		canManageCompanyFn: func(
			context.Context,
			uint,
			uint,
		) (bool, error) {
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
			media.ID = 50

			return nil
		},
	}

	router := newProfileMediaRouterForTest(
		repo,
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/profile-media/companies/7",
		strings.NewReader(`{
			"media_type":"company_logo",
			"url":"https://example.com/logo.jpg"
		}`),
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(
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

func TestProfileMediaRoutes_UpdateCompany(
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
				ID:        50,
				CompanyID: &companyID,
				MediaType: MediaTypeCompanyLogo,
			}, nil
		},

		updateFn: func(
			context.Context,
			*ProfileMedia,
		) error {
			return nil
		},
	}

	router := newProfileMediaRouterForTest(
		repo,
	)

	req := httptest.NewRequest(
		http.MethodPut,
		"/profile-media/companies/7/50",
		strings.NewReader(`{
			"caption":"Updated logo",
			"sort_order":1
		}`),
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(
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

func TestProfileMediaRoutes_DeleteCompany(
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
				ID:        50,
				CompanyID: &companyID,
				MediaType: MediaTypeCompanyLogo,
			}, nil
		},

		deleteFn: func(
			context.Context,
			uint,
		) error {
			return nil
		},
	}

	router := newProfileMediaRouterForTest(
		repo,
	)

	req := httptest.NewRequest(
		http.MethodDelete,
		"/profile-media/companies/7/50",
		nil,
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(
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

func TestProfileMediaRoutes_PublicRouteDoesNotRequireAuth(
	t *testing.T,
) {
	repo := &profileMediaHandlerTestRepository{
		listByUserIDFn: func(
			context.Context,
			uint,
		) ([]ProfileMedia, error) {
			return []ProfileMedia{}, nil
		},
	}

	router := newProfileMediaRouterForTest(
		repo,
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/profile-media/users/5",
		nil,
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(
		recorder,
		req,
	)

	if recorder.Code !=
		http.StatusOK {
		t.Fatalf(
			"expected public route status %d, got %d",
			http.StatusOK,
			recorder.Code,
		)
	}
}

func TestProfileMediaRoutes_NotFound(
	t *testing.T,
) {
	router := newProfileMediaRouterForTest(
		&profileMediaHandlerTestRepository{},
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/profile-media/not-a-route",
		nil,
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(
		recorder,
		req,
	)

	if recorder.Code !=
		http.StatusNotFound {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusNotFound,
			recorder.Code,
		)
	}
}
