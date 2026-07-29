package jobs

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestRegisterRoutes_PublicRoutes(t *testing.T) {
	repo := &MockRepository{}
	service := NewService(repo, nil, nil)
	handler := NewHandler(service)

	router := chi.NewRouter()

	authCalled := false

	auth := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authCalled = true
			next.ServeHTTP(w, r)
		})
	}

	RegisterRoutes(router, handler, auth)

	tests := []struct {
		name   string
		method string
		path   string
	}{
		{"Search", http.MethodGet, "/jobs/search"},
		{"List", http.MethodGet, "/jobs/"},
		{"GetByID", http.MethodGet, "/jobs/1"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			authCalled = false

			req := httptest.NewRequest(tc.method, tc.path, nil)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			if authCalled {
				t.Fatal("public route should not invoke auth middleware")
			}
		})
	}
}

func TestRegisterRoutes_ProtectedRoutes_UseMiddleware(t *testing.T) {
	repo := &MockRepository{}
	service := NewService(repo, nil, nil)
	handler := NewHandler(service)

	router := chi.NewRouter()

	authCalled := false

	auth := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authCalled = true
			w.WriteHeader(http.StatusUnauthorized)
		})
	}

	RegisterRoutes(router, handler, auth)

	tests := []struct {
		name   string
		method string
		path   string
	}{
		{"Create", http.MethodPost, "/jobs/"},
		{"Mine", http.MethodGet, "/jobs/mine"},
		{"Update", http.MethodPut, "/jobs/1"},
		{"Delete", http.MethodDelete, "/jobs/1"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			authCalled = false

			req := httptest.NewRequest(tc.method, tc.path, nil)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			if !authCalled {
				t.Fatal("expected auth middleware to be called")
			}

			if rec.Code != http.StatusUnauthorized {
				t.Fatalf("expected %d got %d", http.StatusUnauthorized, rec.Code)
			}
		})
	}
}

