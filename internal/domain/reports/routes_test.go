package reports

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/rodrigueghenda/jobira/internal/security/identity"
)

func reportsAuthMiddleware(
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
					Role:   "client",
				},
			)

			next.ServeHTTP(
				w,
				r.WithContext(ctx),
			)
		},
	)
}

func reportsAdminMiddleware(
	next http.Handler,
) http.Handler {
	return http.HandlerFunc(
		func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			currentUser, err := identity.FromContext(
				r.Context(),
			)
			if err != nil {
				http.Error(
					w,
					"unauthorized",
					http.StatusUnauthorized,
				)
				return
			}

			if currentUser.Role != "admin" {
				http.Error(
					w,
					"forbidden",
					http.StatusForbidden,
				)
				return
			}

			next.ServeHTTP(w, r)
		},
	)
}

func reportsAdminAuthMiddleware(
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
					Role:   "admin",
				},
			)

			next.ServeHTTP(
				w,
				r.WithContext(ctx),
			)
		},
	)
}

func reportsPassthroughMiddleware(
	next http.Handler,
) http.Handler {
	return next
}

func TestRegisterRoutes_Create(t *testing.T) {
	repo := &mockRepository{
		createFn: func(
			context.Context,
			*Report,
		) error {
			return nil
		},
	}

	handler := NewHandler(NewService(repo))

	router := chi.NewRouter()

	RegisterRoutes(
		router,
		handler,
		reportsAuthMiddleware,
		reportsPassthroughMiddleware,
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/reports/",
		http.NoBody,
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	// Route exists and reached handler.
	// Empty body is therefore 400, not 404/405.
	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected %d, got %d",
			http.StatusBadRequest,
			rec.Code,
		)
	}
}

func TestRegisterRoutes_ListMine(t *testing.T) {
	repo := &mockRepository{
		listByReporterIDFn: func(
			context.Context,
			uint,
		) ([]Report, error) {
			return []Report{}, nil
		},
	}

	handler := NewHandler(NewService(repo))

	router := chi.NewRouter()

	RegisterRoutes(
		router,
		handler,
		reportsAuthMiddleware,
		reportsPassthroughMiddleware,
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/reports/me",
		nil,
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected %d, got %d: %s",
			http.StatusOK,
			rec.Code,
			rec.Body.String(),
		)
	}
}

func TestRegisterRoutes_GetByID(t *testing.T) {
	repo := &mockRepository{
		getByIDFn: func(
			context.Context,
			uint,
		) (*Report, error) {
			return &Report{
				ID:         10,
				ReporterID: 5,
			}, nil
		},
	}

	handler := NewHandler(NewService(repo))

	router := chi.NewRouter()

	RegisterRoutes(
		router,
		handler,
		reportsAuthMiddleware,
		reportsPassthroughMiddleware,
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/reports/10",
		nil,
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected %d, got %d: %s",
			http.StatusOK,
			rec.Code,
			rec.Body.String(),
		)
	}
}

func TestRegisterRoutes_AdminListAll(t *testing.T) {
	repo := &mockRepository{
		listAllFn: func(
			context.Context,
		) ([]Report, error) {
			return []Report{}, nil
		},
	}

	handler := NewHandler(NewService(repo))

	router := chi.NewRouter()

	RegisterRoutes(
		router,
		handler,
		reportsAdminAuthMiddleware,
		reportsAdminMiddleware,
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/admin/reports/",
		nil,
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected %d, got %d: %s",
			http.StatusOK,
			rec.Code,
			rec.Body.String(),
		)
	}
}

func TestRegisterRoutes_AdminListOpen(t *testing.T) {
	repo := &mockRepository{
		listOpenFn: func(
			context.Context,
		) ([]Report, error) {
			return []Report{}, nil
		},
	}

	handler := NewHandler(NewService(repo))

	router := chi.NewRouter()

	RegisterRoutes(
		router,
		handler,
		reportsAdminAuthMiddleware,
		reportsAdminMiddleware,
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/admin/reports/open",
		nil,
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected %d, got %d: %s",
			http.StatusOK,
			rec.Code,
			rec.Body.String(),
		)
	}
}

func TestRegisterRoutes_AdminReview(t *testing.T) {
	handler := NewHandler(
		NewService(&mockRepository{}),
	)

	router := chi.NewRouter()

	RegisterRoutes(
		router,
		handler,
		reportsAdminAuthMiddleware,
		reportsAdminMiddleware,
	)

	req := httptest.NewRequest(
		http.MethodPatch,
		"/admin/reports/10",
		http.NoBody,
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	// Route exists and reaches Review.
	// Empty JSON body should produce 400.
	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected %d, got %d",
			http.StatusBadRequest,
			rec.Code,
		)
	}
}

func TestRegisterRoutes_AdminForbidden(
	t *testing.T,
) {
	handler := NewHandler(
		NewService(&mockRepository{}),
	)

	router := chi.NewRouter()

	RegisterRoutes(
		router,
		handler,
		reportsAuthMiddleware,
		reportsAdminMiddleware,
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/admin/reports/",
		nil,
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf(
			"expected %d, got %d",
			http.StatusForbidden,
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
		reportsAuthMiddleware,
		reportsPassthroughMiddleware,
	)

	req := httptest.NewRequest(
		http.MethodDelete,
		"/reports/10",
		nil,
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf(
			"expected %d, got %d",
			http.StatusMethodNotAllowed,
			rec.Code,
		)
	}
}

func TestRegisterRoutes_NotFound(t *testing.T) {
	handler := NewHandler(
		NewService(&mockRepository{}),
	)

	router := chi.NewRouter()

	RegisterRoutes(
		router,
		handler,
		reportsAuthMiddleware,
		reportsPassthroughMiddleware,
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/unknown",
		nil,
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf(
			"expected %d, got %d",
			http.StatusNotFound,
			rec.Code,
		)
	}
}
