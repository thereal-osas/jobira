package companyaccounts

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/rodrigueghenda/jobira/internal/security/identity"
)

func companyTestAuth(
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
					UserID: 5,
				},
			)

			next.ServeHTTP(
				w,
				r.WithContext(ctx),
			)
		},
	)
}

func TestRegisterRoutes_CreateCompany(
	t *testing.T,
) {
	repo := &mockRepository{
		createCompanyFn: func(
			ctx context.Context,
			company *Company,
		) error {
			company.ID = 10
			return nil
		},

		addMemberFn: func(
			ctx context.Context,
			companyID uint,
			userID uint,
			role string,
		) error {
			return nil
		},
	}

	handler := NewHandler(
		NewService(repo),
	)

	router := chi.NewRouter()

	RegisterRoutes(
		router,
		handler,
		companyTestAuth,
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/companies/",
		companyBody(
			`{
				"name":"Jobira Cleaning Ltd",
				"description":"London"
			}`,
		),
	)

	recorder :=
		httptest.NewRecorder()

	router.ServeHTTP(
		recorder,
		req,
	)

	if recorder.Code !=
		http.StatusCreated {
		t.Fatalf(
			"expected status 201, got %d: %s",
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestRegisterRoutes_ListMembers(
	t *testing.T,
) {
	repo := &mockRepository{
		getMemberFn: func(
			ctx context.Context,
			companyID uint,
			userID uint,
		) (*CompanyMember, error) {
			return &CompanyMember{
				CompanyID: companyID,
				UserID:    userID,
				Role:      "owner",
				Status:    "active",
			}, nil
		},

		listMembersFn: func(
			ctx context.Context,
			companyID uint,
		) ([]CompanyMember, error) {
			return []CompanyMember{},
				nil
		},
	}

	handler := NewHandler(
		NewService(repo),
	)

	router := chi.NewRouter()

	RegisterRoutes(
		router,
		handler,
		companyTestAuth,
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/companies/10/members",
		nil,
	)

	recorder :=
		httptest.NewRecorder()

	router.ServeHTTP(
		recorder,
		req,
	)

	if recorder.Code !=
		http.StatusOK {
		t.Fatalf(
			"expected status 200, got %d: %s",
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestRegisterRoutes_AddMember(
	t *testing.T,
) {
	repo := &mockRepository{
		getMemberFn: func(
			ctx context.Context,
			companyID uint,
			userID uint,
		) (*CompanyMember, error) {
			return &CompanyMember{
				CompanyID: companyID,
				UserID:    userID,
				Role:      "owner",
				Status:    "active",
			}, nil
		},

		getCleanerSeatLimitFn: func(
			ctx context.Context,
			companyID uint,
		) (int, error) {
			return 3, nil
		},

		countActiveCleanerSeatsFn: func(
			ctx context.Context,
			companyID uint,
		) (int, error) {
			return 1, nil
		},

		addMemberFn: func(
			ctx context.Context,
			companyID uint,
			userID uint,
			role string,
		) error {
			return nil
		},
	}

	handler := NewHandler(
		NewService(repo),
	)

	router := chi.NewRouter()

	RegisterRoutes(
		router,
		handler,
		companyTestAuth,
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/companies/10/members",
		companyBody(
			`{
				"user_id":8,
				"role":"cleaner"
			}`,
		),
	)

	recorder :=
		httptest.NewRecorder()

	router.ServeHTTP(
		recorder,
		req,
	)

	if recorder.Code !=
		http.StatusCreated {
		t.Fatalf(
			"expected status 201, got %d: %s",
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestRegisterRoutes_UpdateMemberStatus(
	t *testing.T,
) {
	repo := &mockRepository{
		getMemberFn: func(
			ctx context.Context,
			companyID uint,
			userID uint,
		) (*CompanyMember, error) {
			if userID == 5 {
				return &CompanyMember{
					CompanyID: companyID,
					UserID:    5,
					Role:      "owner",
					Status:    "active",
				}, nil
			}

			return &CompanyMember{
				CompanyID: companyID,
				UserID:    userID,
				Role:      "cleaner",
				Status:    "active",
			}, nil
		},

		getByIDFn: func(
			ctx context.Context,
			companyID uint,
		) (*Company, error) {
			return &Company{
				ID:      companyID,
				OwnerID: 5,
			}, nil
		},

		updateMemberStatusFn: func(
			ctx context.Context,
			companyID uint,
			userID uint,
			status string,
		) error {
			return nil
		},
	}

	handler := NewHandler(
		NewService(repo),
	)

	router := chi.NewRouter()

	RegisterRoutes(
		router,
		handler,
		companyTestAuth,
	)

	req := httptest.NewRequest(
		http.MethodPatch,
		"/companies/10/members/8/status",
		companyBody(
			`{"status":"inactive"}`,
		),
	)

	recorder :=
		httptest.NewRecorder()

	router.ServeHTTP(
		recorder,
		req,
	)

	if recorder.Code !=
		http.StatusOK {
		t.Fatalf(
			"expected status 200, got %d: %s",
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestRegisterRoutes_UpdateMemberRole(
	t *testing.T,
) {
	repo := &mockRepository{
		getMemberFn: func(
			ctx context.Context,
			companyID uint,
			userID uint,
		) (*CompanyMember, error) {
			if userID == 5 {
				return &CompanyMember{
					CompanyID: companyID,
					UserID:    5,
					Role:      "owner",
					Status:    "active",
				}, nil
			}

			return &CompanyMember{
				CompanyID: companyID,
				UserID:    userID,
				Role:      "cleaner",
				Status:    "active",
			}, nil
		},

		getByIDFn: func(
			ctx context.Context,
			companyID uint,
		) (*Company, error) {
			return &Company{
				ID:      companyID,
				OwnerID: 5,
			}, nil
		},

		updateMemberRoleFn: func(
			ctx context.Context,
			companyID uint,
			userID uint,
			role string,
		) error {
			return nil
		},
	}

	handler := NewHandler(
		NewService(repo),
	)

	router := chi.NewRouter()

	RegisterRoutes(
		router,
		handler,
		companyTestAuth,
	)

	req := httptest.NewRequest(
		http.MethodPatch,
		"/companies/10/members/8/role",
		companyBody(
			`{"role":"admin"}`,
		),
	)

	recorder :=
		httptest.NewRecorder()

	router.ServeHTTP(
		recorder,
		req,
	)

	if recorder.Code !=
		http.StatusOK {
		t.Fatalf(
			"expected status 200, got %d: %s",
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestRegisterRoutes_GetDashboard(
	t *testing.T,
) {
	repo := &mockRepository{
		getMemberFn: func(
			ctx context.Context,
			companyID uint,
			userID uint,
		) (*CompanyMember, error) {
			return &CompanyMember{
				CompanyID: companyID,
				UserID:    userID,
				Role:      "owner",
				Status:    "active",
			}, nil
		},

		getByIDFn: func(
			ctx context.Context,
			companyID uint,
		) (*Company, error) {
			return &Company{
				ID:      companyID,
				OwnerID: 5,
				Name:    "Jobira Cleaning Ltd",
			}, nil
		},

		listMembersFn: func(
			ctx context.Context,
			companyID uint,
		) ([]CompanyMember, error) {
			return []CompanyMember{
				{
					UserID: 5,
					Role:   "owner",
					Status: "active",
				},
			}, nil
		},

		getCleanerSeatLimitFn: func(
			ctx context.Context,
			companyID uint,
		) (int, error) {
			return 3, nil
		},

		countMembersByRolesAndStatusFn: func(
			ctx context.Context,
			companyID uint,
			role string,
			status string,
		) (int, error) {
			return 0, nil
		},
	}

	handler := NewHandler(
		NewService(repo),
	)

	router := chi.NewRouter()

	RegisterRoutes(
		router,
		handler,
		companyTestAuth,
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/companies/10/dashboard",
		nil,
	)

	recorder :=
		httptest.NewRecorder()

	router.ServeHTTP(
		recorder,
		req,
	)

	if recorder.Code !=
		http.StatusOK {
		t.Fatalf(
			"expected status 200, got %d: %s",
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func companyBody(
	value string,
) *bytes.Reader {
	return bytes.NewReader(
		[]byte(value),
	)
}
