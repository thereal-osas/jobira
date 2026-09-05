package workhistory

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/rodrigueghenda/jobira/internal/security/identity"
)

func workHistoryRequestWithUser(
	req *http.Request,
	userID uint,
	role string,
) *http.Request {
	ctx := identity.WithUser(
		req.Context(),
		identity.UserIdentity{
			UserID: userID,
			Email:  "test@jobira.co.uk",
			Role:   role,
		},
	)

	return req.WithContext(ctx)
}

func TestNewHandler(t *testing.T) {
	service := NewService(
		&mockRepository{},
	)

	handler := NewHandler(service)

	if handler == nil {
		t.Fatal("expected handler")
	}

	if handler.service != service {
		t.Fatal("expected service assigned")
	}
}

func TestHandler_GetMine_CleanerSuccess(
	t *testing.T,
) {
	now := time.Now()

	repo := &mockRepository{
		listByCleanerIDFn: func(
			ctx context.Context,
			cleanerID uint,
		) ([]WorkHistoryEntry, error) {
			if cleanerID != 8 {
				t.Fatalf(
					"expected cleaner id 8, got %d",
					cleanerID,
				)
			}

			return []WorkHistoryEntry{
				{
					BookingID: 10,
					JobID:     20,
					ClientID:  5,
					CleanerID: 8,
					JobTitle:  "End of tenancy clean",
					JobType:   "end_of_tenancy",
					Location:  "London",
					Budget:    150,
					Status:    "completed",
					CreatedAt: now,
					UpdatedAt: now,
				},
			}, nil
		},
	}

	handler := NewHandler(
		NewService(repo),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/work-history/me",
		nil,
	)

	req = workHistoryRequestWithUser(
		req,
		8,
		"cleaner",
	)

	rec := httptest.NewRecorder()

	handler.GetMine(
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

	if !strings.Contains(
		rec.Body.String(),
		`"booking_id":10`,
	) {
		t.Fatalf(
			"expected booking id in response: %s",
			rec.Body.String(),
		)
	}
}

func TestHandler_GetMine_ClientSuccess(
	t *testing.T,
) {
	repo := &mockRepository{
		listByClientIDFn: func(
			ctx context.Context,
			clientID uint,
		) ([]WorkHistoryEntry, error) {
			if clientID != 5 {
				t.Fatalf(
					"expected client id 5, got %d",
					clientID,
				)
			}

			return []WorkHistoryEntry{
				{
					BookingID: 10,
					ClientID:  5,
					CleanerID: 8,
					Status:    "completed",
				},
			}, nil
		},
	}

	handler := NewHandler(
		NewService(repo),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/work-history/me",
		nil,
	)

	req = workHistoryRequestWithUser(
		req,
		5,
		"client",
	)

	rec := httptest.NewRecorder()

	handler.GetMine(
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

	if !strings.Contains(
		rec.Body.String(),
		`"client_id":5`,
	) {
		t.Fatalf(
			"expected client id in response: %s",
			rec.Body.String(),
		)
	}
}

func TestHandler_GetMine_Unauthorized(
	t *testing.T,
) {
	handler := NewHandler(
		NewService(
			&mockRepository{},
		),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/work-history/me",
		nil,
	)

	rec := httptest.NewRecorder()

	handler.GetMine(
		rec,
		req,
	)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusUnauthorized,
			rec.Code,
		)
	}
}

func TestHandler_GetMine_Forbidden(
	t *testing.T,
) {
	handler := NewHandler(
		NewService(
			&mockRepository{},
		),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/work-history/me",
		nil,
	)

	req = workHistoryRequestWithUser(
		req,
		7,
		"user",
	)

	rec := httptest.NewRecorder()

	handler.GetMine(
		rec,
		req,
	)

	if rec.Code != http.StatusForbidden {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusForbidden,
			rec.Code,
			rec.Body.String(),
		)
	}
}

func TestHandler_GetMine_RepositoryError(
	t *testing.T,
) {
	expectedErr := errors.New(
		"database failed",
	)

	repo := &mockRepository{
		listByCleanerIDFn: func(
			context.Context,
			uint,
		) ([]WorkHistoryEntry, error) {
			return nil, expectedErr
		},
	}

	handler := NewHandler(
		NewService(repo),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/work-history/me",
		nil,
	)

	req = workHistoryRequestWithUser(
		req,
		8,
		"cleaner",
	)

	rec := httptest.NewRecorder()

	handler.GetMine(
		rec,
		req,
	)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusInternalServerError,
			rec.Code,
			rec.Body.String(),
		)
	}
}

func TestHandler_GetBookingHistory_ClientSuccess(
	t *testing.T,
) {
	repo := &mockRepository{
		userCanAccessBookingFn: func(
			ctx context.Context,
			bookingID uint,
			userID uint,
		) (bool, error) {
			return bookingID == 10 &&
				userID == 5, nil
		},

		getByBookingIDFn: func(
			ctx context.Context,
			bookingID uint,
			userID uint,
		) (*WorkHistoryDetail, error) {
			return &WorkHistoryDetail{
				WorkHistoryEntry: WorkHistoryEntry{
					BookingID: bookingID,
					ClientID:  5,
					CleanerID: 8,
					Status:    "completed",
				},
				BeforePhotos: []WorkProofPhoto{
					{
						ID:        1,
						ProofType: "before",
						PhotoURL:  "before.jpg",
					},
				},
				AfterPhoto: []WorkProofPhoto{
					{
						ID:        2,
						ProofType: "after",
						PhotoURL:  "after.jpg",
					},
				},
			}, nil
		},
	}

	handler := NewHandler(
		NewService(repo),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/work-history/bookings/10",
		nil,
	)

	routeCtx := chi.NewRouteContext()

	routeCtx.URLParams.Add(
		"bookingID",
		"10",
	)

	req = req.WithContext(
		context.WithValue(
			req.Context(),
			chi.RouteCtxKey,
			routeCtx,
		),
	)

	req = workHistoryRequestWithUser(
		req,
		5,
		"client",
	)

	rec := httptest.NewRecorder()

	handler.GetBookingHistory(
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

	if !strings.Contains(
		rec.Body.String(),
		`"booking_id":10`,
	) {
		t.Fatalf(
			"expected booking id in response: %s",
			rec.Body.String(),
		)
	}

	if !strings.Contains(
		rec.Body.String(),
		`"before_photos"`,
	) {
		t.Fatalf(
			"expected before photos in response: %s",
			rec.Body.String(),
		)
	}
}

func TestHandler_GetBookingHistory_Unauthorized(
	t *testing.T,
) {
	handler := NewHandler(
		NewService(
			&mockRepository{},
		),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/work-history/bookings/10",
		nil,
	)

	rec := httptest.NewRecorder()

	handler.GetBookingHistory(
		rec,
		req,
	)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusUnauthorized,
			rec.Code,
		)
	}
}

func TestHandler_GetBookingHistory_InvalidBookingID(
	t *testing.T,
) {
	handler := NewHandler(
		NewService(
			&mockRepository{},
		),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/work-history/bookings/abc",
		nil,
	)

	routeCtx := chi.NewRouteContext()

	routeCtx.URLParams.Add(
		"bookingID",
		"abc",
	)

	req = req.WithContext(
		context.WithValue(
			req.Context(),
			chi.RouteCtxKey,
			routeCtx,
		),
	)

	req = workHistoryRequestWithUser(
		req,
		5,
		"client",
	)

	rec := httptest.NewRecorder()

	handler.GetBookingHistory(
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

func TestHandler_GetBookingHistory_ZeroBookingID(
	t *testing.T,
) {
	handler := NewHandler(
		NewService(
			&mockRepository{},
		),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/work-history/bookings/0",
		nil,
	)

	routeCtx := chi.NewRouteContext()

	routeCtx.URLParams.Add(
		"bookingID",
		"0",
	)

	req = req.WithContext(
		context.WithValue(
			req.Context(),
			chi.RouteCtxKey,
			routeCtx,
		),
	)

	req = workHistoryRequestWithUser(
		req,
		5,
		"client",
	)

	rec := httptest.NewRecorder()

	handler.GetBookingHistory(
		rec,
		req,
	)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			rec.Code,
		)
	}
}

func TestHandler_GetBookingHistory_Forbidden(
	t *testing.T,
) {
	repo := &mockRepository{
		userCanAccessBookingFn: func(
			context.Context,
			uint,
			uint,
		) (bool, error) {
			return false, nil
		},
	}

	handler := NewHandler(
		NewService(repo),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/work-history/bookings/10",
		nil,
	)

	routeCtx := chi.NewRouteContext()

	routeCtx.URLParams.Add(
		"bookingID",
		"10",
	)

	req = req.WithContext(
		context.WithValue(
			req.Context(),
			chi.RouteCtxKey,
			routeCtx,
		),
	)

	req = workHistoryRequestWithUser(
		req,
		50,
		"client",
	)

	rec := httptest.NewRecorder()

	handler.GetBookingHistory(
		rec,
		req,
	)

	if rec.Code != http.StatusForbidden {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusForbidden,
			rec.Code,
			rec.Body.String(),
		)
	}
}

func TestHandler_GetBookingHistory_NotFound(
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
			context.Context,
			uint,
			uint,
		) (*WorkHistoryDetail, error) {
			return nil, ErrHistoryNotFound
		},
	}

	handler := NewHandler(
		NewService(repo),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/work-history/bookings/999",
		nil,
	)

	routeCtx := chi.NewRouteContext()

	routeCtx.URLParams.Add(
		"bookingID",
		"999",
	)

	req = req.WithContext(
		context.WithValue(
			req.Context(),
			chi.RouteCtxKey,
			routeCtx,
		),
	)

	req = workHistoryRequestWithUser(
		req,
		5,
		"client",
	)

	rec := httptest.NewRecorder()

	handler.GetBookingHistory(
		rec,
		req,
	)

	if rec.Code != http.StatusNotFound {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusNotFound,
			rec.Code,
			rec.Body.String(),
		)
	}
}

func TestHandler_GetBookingHistory_RepositoryError(
	t *testing.T,
) {
	expectedErr := errors.New(
		"database failed",
	)

	repo := &mockRepository{
		userCanAccessBookingFn: func(
			context.Context,
			uint,
			uint,
		) (bool, error) {
			return true, nil
		},

		getByBookingIDFn: func(
			context.Context,
			uint,
			uint,
		) (*WorkHistoryDetail, error) {
			return nil, expectedErr
		},
	}

	handler := NewHandler(
		NewService(repo),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/work-history/bookings/10",
		nil,
	)

	routeCtx := chi.NewRouteContext()

	routeCtx.URLParams.Add(
		"bookingID",
		"10",
	)

	req = req.WithContext(
		context.WithValue(
			req.Context(),
			chi.RouteCtxKey,
			routeCtx,
		),
	)

	req = workHistoryRequestWithUser(
		req,
		5,
		"client",
	)

	rec := httptest.NewRecorder()

	handler.GetBookingHistory(
		rec,
		req,
	)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusInternalServerError,
			rec.Code,
			rec.Body.String(),
		)
	}
}
