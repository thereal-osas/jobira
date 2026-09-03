package bookingtimeline

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/rodrigueghenda/jobira/internal/security/identity"
)

func bookingTimelineAuthMiddleware(
	userID uint,
	role string,
) func(http.Handler) http.Handler {
	return func(
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

func TestRegisterRoutes_GetBookingTimeline(
	t *testing.T,
) {
	repo := &mockRepository{
		getBookingAccessFn: func(
			context.Context,
			uint,
		) (uint, uint, string, error) {
			return 5, 8, "accepted", nil
		},

		listByBookingIDFn: func(
			context.Context,
			uint,
		) ([]StatusHistory, error) {
			return []StatusHistory{
				{
					ID:         1,
					BookingID:  12,
					FromStatus: "pending",
					ToStatus:   "accepted",
				},
			}, nil
		},
	}

	handler := newBookingTimelineHandlerForTest(
		repo,
	)

	router := chi.NewRouter()

	RegisterRoutes(
		router,
		handler,
		bookingTimelineAuthMiddleware(
			5,
			"client",
		),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/booking-timeline/12",
		nil,
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(
		recorder,
		req,
	)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusOK,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestRegisterRoutes_GetBookingTimeline_Cleaner(
	t *testing.T,
) {
	repo := &mockRepository{
		getBookingAccessFn: func(
			context.Context,
			uint,
		) (uint, uint, string, error) {
			return 5, 8, "confirmed", nil
		},
	}

	handler := newBookingTimelineHandlerForTest(
		repo,
	)

	router := chi.NewRouter()

	RegisterRoutes(
		router,
		handler,
		bookingTimelineAuthMiddleware(
			8,
			"cleaner",
		),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/booking-timeline/12",
		nil,
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(
		recorder,
		req,
	)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusOK,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestRegisterRoutes_GetBookingTimeline_Admin(
	t *testing.T,
) {
	repo := &mockRepository{
		getBookingAccessFn: func(
			context.Context,
			uint,
		) (uint, uint, string, error) {
			return 5, 8, "completed", nil
		},
	}

	handler := newBookingTimelineHandlerForTest(
		repo,
	)

	router := chi.NewRouter()

	RegisterRoutes(
		router,
		handler,
		bookingTimelineAuthMiddleware(
			99,
			"admin",
		),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/booking-timeline/12",
		nil,
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(
		recorder,
		req,
	)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusOK,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestRegisterRoutes_BookingTimeline_MethodNotAllowed(
	t *testing.T,
) {
	handler := newBookingTimelineHandlerForTest(
		&mockRepository{},
	)

	router := chi.NewRouter()

	RegisterRoutes(
		router,
		handler,
		bookingTimelineAuthMiddleware(
			5,
			"client",
		),
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/booking-timeline/12",
		nil,
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(
		recorder,
		req,
	)

	if recorder.Code != http.StatusMethodNotAllowed {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusMethodNotAllowed,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestRegisterRoutes_BookingTimeline_NotFound(
	t *testing.T,
) {
	handler := newBookingTimelineHandlerForTest(
		&mockRepository{},
	)

	router := chi.NewRouter()

	RegisterRoutes(
		router,
		handler,
		bookingTimelineAuthMiddleware(
			5,
			"client",
		),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/booking-timeline",
		nil,
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(
		recorder,
		req,
	)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusNotFound,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}
