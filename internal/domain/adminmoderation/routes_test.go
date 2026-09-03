package adminmoderation

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/rodrigueghenda/jobira/internal/security/identity"
)

func adminModerationTestAuth(
	userID uint,
	role string,
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				ctx := identity.WithUser(
					r.Context(),
					identity.UserIdentity{
						UserID: userID,
						Role:   role,
					},
				)

				next.ServeHTTP(
					w,
					r.WithContext(ctx),
				)
			},
		)
	}
}

func adminModerationAdminOnly(
	next http.Handler,
) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
		},
	)
}

func TestRegisterRoutes_ListReports(t *testing.T) {
	repo := &mockRepository{
		listReportsFn: func(
			context.Context,
		) ([]CleanerReportAdminView, error) {
			return []CleanerReportAdminView{}, nil
		},
	}

	handler := NewHandler(
		NewService(repo),
	)

	router := chi.NewRouter()

	RegisterRoutes(
		router,
		handler,
		adminModerationTestAuth(
			9,
			"admin",
		),
		adminModerationAdminOnly,
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/admin/moderation/reports",
		nil,
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusOK,
			rec.Code,
		)
	}
}

func TestRegisterRoutes_ListOpenReports(
	t *testing.T,
) {
	repo := &mockRepository{
		listOpenReportsFn: func(
			context.Context,
		) ([]CleanerReportAdminView, error) {
			return []CleanerReportAdminView{}, nil
		},
	}

	handler := NewHandler(
		NewService(repo),
	)

	router := chi.NewRouter()

	RegisterRoutes(
		router,
		handler,
		adminModerationTestAuth(
			9,
			"admin",
		),
		adminModerationAdminOnly,
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/admin/moderation/reports/open",
		nil,
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusOK,
			rec.Code,
		)
	}
}

func TestRegisterRoutes_UpdateReportStatus(
	t *testing.T,
) {
	repo := &mockRepository{
		updateReportStatusFn: func(
			context.Context,
			uint,
			uint,
			string,
			string,
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
		adminModerationTestAuth(
			9,
			"admin",
		),
		adminModerationAdminOnly,
	)

	req := httptest.NewRequest(
		http.MethodPatch,
		"/admin/moderation/reports/12",
		strings.NewReader(`{
			"status":"resolved",
			"admin_notes":"Reviewed"
		}`),
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusOK,
			rec.Code,
			rec.Body.String(),
		)
	}
}

func TestRegisterRoutes_ListBlockedCleaners(
	t *testing.T,
) {
	repo := &mockRepository{
		listBlockCleanersFn: func(
			context.Context,
		) ([]BlockedCleanerAdminView, error) {
			return []BlockedCleanerAdminView{}, nil
		},
	}

	handler := NewHandler(
		NewService(repo),
	)

	router := chi.NewRouter()

	RegisterRoutes(
		router,
		handler,
		adminModerationTestAuth(
			9,
			"admin",
		),
		adminModerationAdminOnly,
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/admin/moderation/blocked-cleaners",
		nil,
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusOK,
			rec.Code,
		)
	}
}

func TestRegisterRoutes_MethodNotAllowed(
	t *testing.T,
) {
	handler := NewHandler(
		NewService(&mockRepository{}),
	)

	router := chi.NewRouter()

	RegisterRoutes(
		router,
		handler,
		adminModerationTestAuth(
			9,
			"admin",
		),
		adminModerationAdminOnly,
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/admin/moderation/reports",
		nil,
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusMethodNotAllowed,
			rec.Code,
		)
	}
}

func TestRegisterRoutes_NotFound(
	t *testing.T,
) {
	handler := NewHandler(
		NewService(&mockRepository{}),
	)

	router := chi.NewRouter()

	RegisterRoutes(
		router,
		handler,
		adminModerationTestAuth(
			9,
			"admin",
		),
		adminModerationAdminOnly,
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/admin/moderation",
		nil,
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusNotFound,
			rec.Code,
		)
	}
}

func TestRegisterRoutes_AuthCanBlock(
	t *testing.T,
) {
	handler := NewHandler(
		NewService(&mockRepository{}),
	)

	authMiddleware := func(
		next http.Handler,
	) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				http.Error(
					w,
					"unauthorized",
					http.StatusUnauthorized,
				)
			},
		)
	}

	adminCalled := false

	adminMiddleware := func(
		next http.Handler,
	) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				adminCalled = true
				next.ServeHTTP(w, r)
			},
		)
	}

	router := chi.NewRouter()

	RegisterRoutes(
		router,
		handler,
		authMiddleware,
		adminMiddleware,
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/admin/moderation/reports",
		nil,
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusUnauthorized,
			rec.Code,
		)
	}

	if adminCalled {
		t.Fatal(
			"admin middleware should not run when auth blocks",
		)
	}
}

func TestRegisterRoutes_AdminOnlyCanBlock(
	t *testing.T,
) {
	handler := NewHandler(
		NewService(&mockRepository{}),
	)

	authCalled := false

	authMiddleware := func(
		next http.Handler,
	) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				authCalled = true
				next.ServeHTTP(w, r)
			},
		)
	}

	adminMiddleware := func(
		next http.Handler,
	) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				http.Error(
					w,
					"forbidden",
					http.StatusForbidden,
				)
			},
		)
	}

	router := chi.NewRouter()

	RegisterRoutes(
		router,
		handler,
		authMiddleware,
		adminMiddleware,
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/admin/moderation/reports",
		nil,
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusForbidden,
			rec.Code,
		)
	}

	if !authCalled {
		t.Fatal(
			"expected auth middleware to run first",
		)
	}
}
