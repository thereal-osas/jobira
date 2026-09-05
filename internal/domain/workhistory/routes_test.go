package workhistory

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/rodrigueghenda/jobira/internal/security/identity"
)

func workHistoryTestAuth(
	userID uint,
	role string,
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(
				w http.ResponseWriter,
				r *http.Request,
			) {
				ctx := identity.WithUser(
					r.Context(),
					identity.UserIdentity{
						UserID: userID,
						Email:  "test@jobira.co.uk",
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

func TestRegisterRoutes_GetMine(
	t *testing.T,
) {
	repo := &mockRepository{
		listByCleanerIDFn: func(
			context.Context,
			uint,
		) ([]WorkHistoryEntry, error) {
			return []WorkHistoryEntry{
				{
					BookingID: 10,
					CleanerID: 8,
					ClientID:  5,
					Status:    "completed",
				},
			}, nil
		},
	}

	handler := NewHandler(
		NewService(repo),
	)

	router := chi.NewRouter()

	RegisterRoutes(
		router,
		handler,
		workHistoryTestAuth(
			8,
			"cleaner",
		),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/work-history/me",
		nil,
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(
		rec,
		req,
	)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusOK,
			rec.Code,
			rec.Body.String(),
		)
	}
}

func TestRegisterRoutes_GetBookingHistory(
	t *testing.T,
) {
	repo := &mockRepository{
		userCanAccessBookingFn: func(
			context.Context,
			uint,
			uint,
		) (bool, error) {
			return true, nil
		},

		getByBookingIDFn: func(
			ctx context.Context,
			bookingID uint,
			userID uint,
		) (*WorkHistoryDetail, error) {
			return &WorkHistoryDetail{
				WorkHistoryEntry: WorkHistoryEntry{
					BookingID: bookingID,
					ClientID:  userID,
					CleanerID: 8,
					Status:    "completed",
				},
			}, nil
		},
	}

	handler := NewHandler(
		NewService(repo),
	)

	router := chi.NewRouter()

	RegisterRoutes(
		router,
		handler,
		workHistoryTestAuth(
			5,
			"client",
		),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/work-history/bookings/10",
		nil,
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(
		rec,
		req,
	)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusOK,
			rec.Code,
			rec.Body.String(),
		)
	}
}

func TestRegisterRoutes_InvalidBookingID(
	t *testing.T,
) {
	handler := NewHandler(
		NewService(
			&mockRepository{},
		),
	)

	router := chi.NewRouter()

	RegisterRoutes(
		router,
		handler,
		workHistoryTestAuth(
			5,
			"client",
		),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/work-history/bookings/invalid",
		nil,
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(
		rec,
		req,
	)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusBadRequest,
			rec.Code,
			rec.Body.String(),
		)
	}
}

func TestRegisterRoutes_MethodNotAllowed(
	t *testing.T,
) {
	handler := NewHandler(
		NewService(
			&mockRepository{},
		),
	)

	router := chi.NewRouter()

	RegisterRoutes(
		router,
		handler,
		workHistoryTestAuth(
			5,
			"client",
		),
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/work-history/me",
		nil,
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(
		rec,
		req,
	)

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
		NewService(
			&mockRepository{},
		),
	)

	router := chi.NewRouter()

	RegisterRoutes(
		router,
		handler,
		workHistoryTestAuth(
			5,
			"client",
		),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/work-history/unknown",
		nil,
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(
		rec,
		req,
	)

	if rec.Code != http.StatusNotFound {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusNotFound,
			rec.Code,
		)
	}
}
