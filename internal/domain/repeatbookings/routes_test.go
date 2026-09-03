package repeatbookings

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/rodrigueghenda/jobira/internal/security/identity"
)

func repeatBookingsAuthMiddleware(
	userID uint,
	role string,
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(
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
		})
	}
}

func TestRegisterRoutes_Create(t *testing.T) {
	repo := &mockRepository{
		createJobFn: func(
			context.Context,
			uint,
			CreateRepeatBookingRequest,
		) (uint, error) {
			return 12, nil
		},
		createInvitationFn: func(
			context.Context,
			uint,
			uint,
			uint,
			string,
		) error {
			return nil
		},
		createRepeatBookingFn: func(
			context.Context,
			*RepeatBooking,
		) error {
			return nil
		},
	}

	handler := newHandlerForTest(repo)
	router := chi.NewRouter()

	RegisterRoutes(
		router,
		handler,
		repeatBookingsAuthMiddleware(5, "client"),
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/repeat-bookings/cleaners/8",
		strings.NewReader(`{
			"title":"Weekly domestic clean",
			"description":"Clean a two-bedroom flat",
			"location":"East London",
			"job_type":"domestic",
			"listing_type":"shift",
			"budget":70,
			"message":"Please clean for me again"
		}`),
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusCreated {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusCreated,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestRegisterRoutes_ListMine(t *testing.T) {
	repo := &mockRepository{
		listByClientIDFn: func(
			context.Context,
			uint,
		) ([]RepeatBooking, error) {
			return []RepeatBooking{}, nil
		},
	}

	handler := newHandlerForTest(repo)
	router := chi.NewRouter()

	RegisterRoutes(
		router,
		handler,
		repeatBookingsAuthMiddleware(5, "client"),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/repeat-bookings/me",
		nil,
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusOK,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestRegisterRoutes_BookAgain(t *testing.T) {
	newBookingID := uint(40)

	repo := &mockRepository{
		getOriginBookingFn: func(
			context.Context,
			uint,
		) (*BookingSnapshot, error) {
			return &BookingSnapshot{
				ID:        10,
				JobID:     7,
				ClientID:  5,
				CleanerID: 8,
				Status:    "closed",
			}, nil
		},
		createBookAgainTransactionFn: func(
			ctx context.Context,
			input BookAgainTransaction,
		) (*RepeatBookingRequest, error) {
			return &RepeatBookingRequest{
				ID:                50,
				OriginalBookingID: input.OriginalBookingID,
				NewBookingID:      &newBookingID,
				ClientID:          input.ClientID,
				CleanerID:         input.CleanerID,
				JobID:             input.JobID,
				ScheduledAt:       input.ScheduledAt,
				Status:            input.Status,
				Message:           input.Message,
			}, nil
		},
	}

	handler := newHandlerForTest(repo)
	router := chi.NewRouter()

	RegisterRoutes(
		router,
		handler,
		repeatBookingsAuthMiddleware(5, "client"),
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/repeat-bookings/bookings/10/book-again",
		strings.NewReader(`{
	"scheduled_at":"2026-09-10T10:00:00Z",
	"scheduled_end_at":"2026-09-10T12:00:00Z",
	"message":"Please bring the same equipment"
}`),
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusCreated {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusCreated,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}
