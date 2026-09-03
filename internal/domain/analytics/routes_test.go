package analytics

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

func analyticsAuthMiddleware(
	next http.Handler,
) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			req := analyticsAuthenticatedRequest(r)
			next.ServeHTTP(w, req)
		},
	)
}

func analyticsAdminMiddleware(
	next http.Handler,
) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
		},
	)
}

func TestRegisterRoutes_Dashboard(t *testing.T) {
	router := chi.NewRouter()

	RegisterRoutes(
		router,
		NewHandler(
			NewService(dashboardBaseMock()),
		),
		analyticsAuthMiddleware,
		analyticsAdminMiddleware,
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/admin/analytics/dashboard",
		nil,
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

func TestRegisterRoutes_MethodNotAllowed(
	t *testing.T,
) {
	router := chi.NewRouter()

	RegisterRoutes(
		router,
		NewHandler(
			NewService(&mockRepository{}),
		),
		analyticsAuthMiddleware,
		analyticsAdminMiddleware,
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/admin/analytics/dashboard",
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

func TestRegisterRoutes_AuthCanBlock(
	t *testing.T,
) {
	adminCalled := false

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
		NewHandler(
			NewService(&mockRepository{}),
		),
		authMiddleware,
		adminMiddleware,
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/admin/analytics/dashboard",
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

func TestRegisterRoutes_AdminCanBlock(
	t *testing.T,
) {
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
		NewHandler(
			NewService(&mockRepository{}),
		),
		authMiddleware,
		adminMiddleware,
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/admin/analytics/dashboard",
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

