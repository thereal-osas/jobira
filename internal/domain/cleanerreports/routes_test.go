package cleanerreports

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/rodrigueghenda/jobira/internal/security/identity"
)

func cleanerReportsTestAuth(
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

func TestRegisterRoutes_Create(t *testing.T) {
	repo := &mockRepository{
		createFn: func(
			context.Context,
			*CleanerReport,
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
		cleanerReportsTestAuth(
			5,
			"client",
		),
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/cleaner-reports/8",
		strings.NewReader(`{
			"reason":"No show",
			"details":"Cleaner did not attend."
		}`),
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusCreated,
			rec.Code,
			rec.Body.String(),
		)
	}
}

func TestRegisterRoutes_ListMine(t *testing.T) {
	repo := &mockRepository{
		listByClientIDFn: func(
			context.Context,
			uint,
		) ([]CleanerReport, error) {
			return []CleanerReport{}, nil
		},
	}

	handler := NewHandler(
		NewService(repo),
	)

	router := chi.NewRouter()

	RegisterRoutes(
		router,
		handler,
		cleanerReportsTestAuth(
			5,
			"client",
		),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/cleaner-reports/me",
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

func TestRegisterRoutes_CreateMethodNotAllowed(
	t *testing.T,
) {
	handler := NewHandler(
		NewService(&mockRepository{}),
	)

	router := chi.NewRouter()

	RegisterRoutes(
		router,
		handler,
		cleanerReportsTestAuth(
			5,
			"client",
		),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/cleaner-reports/8",
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
func TestRegisterRoutes_PostMeInvalidCleanerID(
	t *testing.T,
) {
	handler := NewHandler(
		NewService(&mockRepository{}),
	)

	router := chi.NewRouter()

	RegisterRoutes(
		router,
		handler,
		cleanerReportsTestAuth(
			5,
			"client",
		),
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/cleaner-reports/me",
		nil,
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
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
		cleanerReportsTestAuth(
			5,
			"client",
		),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/cleaner-reports",
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
