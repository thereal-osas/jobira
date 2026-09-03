package admin

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

func testAuthMiddleware(
	called *bool,
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				*called = true
				next.ServeHTTP(w, r)
			},
		)
	}
}

func testAdminOnlyMiddleware(
	called *bool,
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				*called = true
				next.ServeHTTP(w, r)
			},
		)
	}
}

func TestRegisterRoutes_Dashboard_Success(
	t *testing.T,
) {
	service := NewService()
	handler := NewHandler(service)

	authCalled := false
	adminCalled := false

	router := chi.NewRouter()

	RegisterRoutes(
		router,
		handler,
		testAuthMiddleware(&authCalled),
		testAdminOnlyMiddleware(&adminCalled),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/admin/dashboard",
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

	if !authCalled {
		t.Fatal(
			"expected auth middleware to be called",
		)
	}

	if !adminCalled {
		t.Fatal(
			"expected admin-only middleware to be called",
		)
	}
}

func TestRegisterRoutes_Dashboard_MethodNotAllowed(
	t *testing.T,
) {
	service := NewService()
	handler := NewHandler(service)

	authCalled := false
	adminCalled := false

	router := chi.NewRouter()

	RegisterRoutes(
		router,
		handler,
		testAuthMiddleware(&authCalled),
		testAdminOnlyMiddleware(&adminCalled),
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/admin/dashboard",
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
	service := NewService()
	handler := NewHandler(service)

	authCalled := false
	adminCalled := false

	router := chi.NewRouter()

	RegisterRoutes(
		router,
		handler,
		testAuthMiddleware(&authCalled),
		testAdminOnlyMiddleware(&adminCalled),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/admin",
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

func TestRegisterRoutes_AuthMiddlewareCanBlock(
	t *testing.T,
) {
	service := NewService()
	handler := NewHandler(service)

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

	router := chi.NewRouter()

	RegisterRoutes(
		router,
		handler,
		authMiddleware,
		testAdminOnlyMiddleware(&adminCalled),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/admin/dashboard",
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
			"admin middleware should not run when auth blocks request",
		)
	}
}

func TestRegisterRoutes_AdminMiddlewareCanBlock(
	t *testing.T,
) {
	service := NewService()
	handler := NewHandler(service)

	authCalled := false

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
		testAuthMiddleware(&authCalled),
		adminMiddleware,
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/admin/dashboard",
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
			"expected auth middleware to run before admin middleware",
		)
	}
}
