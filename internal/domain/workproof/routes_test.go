package wokrproof

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/rodrigueghenda/jobira/internal/security/identity"
)

func workProofCleanerAuth(
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
					UserID: 8,
					Role:   "cleaner",
				},
			)

			next.ServeHTTP(
				w,
				r.WithContext(ctx),
			)
		},
	)
}

func workProofClientAuth(
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

func TestRegisterRoutes_CreateWorkProof(t *testing.T) {
	repo := &mockRepository{
		getBookingFn: func(
			ctx context.Context,
			bookingID uint,
		) (*BookingSnapshot, error) {
			return &BookingSnapshot{
				ID:        bookingID,
				JobID:     7,
				ClientID:  5,
				CleanerID: 8,
				Status:    "in_progress",
			}, nil
		},

		countByBookingAndTypeFn: func(
			ctx context.Context,
			bookingID uint,
			proofType ProofType,
		) (int, error) {
			return 0, nil
		},

		createFn: func(
			ctx context.Context,
			proof *WorkProof,
		) error {
			proof.ID = 30
			proof.CreatedAt = time.Now().UTC()

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
		workProofCleanerAuth,
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/work-proof/bookings/10",
		strings.NewReader(`{
			"proof_type":"before",
			"photo_url":"https://example.com/before.jpg",
			"caption":"Before cleaning"
		}`),
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(
		recorder,
		req,
	)

	if recorder.Code != http.StatusCreated {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusCreated,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestRegisterRoutes_GetBookingProof(t *testing.T) {
	repo := &mockRepository{
		getBookingFn: func(
			ctx context.Context,
			bookingID uint,
		) (*BookingSnapshot, error) {
			return &BookingSnapshot{
				ID:        bookingID,
				ClientID:  5,
				CleanerID: 8,
				Status:    "completed",
			}, nil
		},

		listByBookingIDFn: func(
			ctx context.Context,
			bookingID uint,
		) ([]WorkProof, error) {
			return []WorkProof{},
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
		workProofClientAuth,
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/work-proof/bookings/10",
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

func TestRegisterRoutes_DeleteWorkProof(t *testing.T) {
	repo := &mockRepository{
		getByIDFn: func(
			ctx context.Context,
			proofID uint,
		) (*WorkProof, error) {
			return &WorkProof{
				ID:        proofID,
				BookingID: 10,
				CleanerID: 8,
			}, nil
		},

		deleteFn: func(
			ctx context.Context,
			proofID uint,
			cleanerID uint,
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
		workProofCleanerAuth,
	)

	req := httptest.NewRequest(
		http.MethodDelete,
		"/work-proof/30",
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

func TestRegisterRoutes_GetVerifiedWorkPublic(
	t *testing.T,
) {
	repo := &mockRepository{
		getVerifiedWorkSummaryFn: func(
			ctx context.Context,
			cleanerID uint,
		) (*VerifiedWorkSummary, error) {
			return &VerifiedWorkSummary{
				CleanerID:    cleanerID,
				VerifiedJobs: 5,
				BeforePhotos: 7,
				AfterPhotos:  8,
				TotalPhotos:  15,
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
		workProofCleanerAuth,
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/work-proof/cleaners/8",
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

func TestRegisterRoutes_PrivateRouteRequiresAuth(
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
		func(
			next http.Handler,
		) http.Handler {
			return next
		},
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/work-proof/bookings/10",
		nil,
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(
		recorder,
		req,
	)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusUnauthorized,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}
