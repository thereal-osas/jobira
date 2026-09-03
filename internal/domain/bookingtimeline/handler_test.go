package bookingtimeline

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/rodrigueghenda/jobira/internal/security/identity"
)

func newBookingTimelineHandlerForTest(
	repo Repository,
) *Handler {
	service := NewService(repo)

	return NewHandler(service)
}

func requestWithBookingTimelineUser(
	req *http.Request,
	userID uint,
	role string,
) *http.Request {
	ctx := identity.WithUser(
		req.Context(),
		identity.UserIdentity{
			UserID: userID,
			Role:   role,
		},
	)

	return req.WithContext(ctx)
}

func requestWithBookingTimelineID(
	method string,
	target string,
	bookingID string,
) *http.Request {
	req := httptest.NewRequest(
		method,
		target,
		nil,
	)

	routeContext := chi.NewRouteContext()

	routeContext.URLParams.Add(
		"bookingID",
		bookingID,
	)

	ctx := context.WithValue(
		req.Context(),
		chi.RouteCtxKey,
		routeContext,
	)

	return req.WithContext(ctx)
}

func TestHandler_GetByBookingID_Success(
	t *testing.T,
) {
	repo := &mockRepository{
		getBookingAccessFn: func(
			_ context.Context,
			bookingID uint,
		) (uint, uint, string, error) {
			if bookingID != 12 {
				t.Fatalf(
					"expected booking ID 12, got %d",
					bookingID,
				)
			}

			return 5, 8, "accepted", nil
		},

		listByBookingIDFn: func(
			_ context.Context,
			bookingID uint,
		) ([]StatusHistory, error) {
			return []StatusHistory{
				{
					ID:         1,
					BookingID:  bookingID,
					FromStatus: "pending",
					ToStatus:   "accepted",
					Note:       "booking accepted",
				},
			}, nil
		},
	}

	handler := newBookingTimelineHandlerForTest(
		repo,
	)

	req := requestWithBookingTimelineID(
		http.MethodGet,
		"/booking-timeline/12",
		"12",
	)

	req = requestWithBookingTimelineUser(
		req,
		5,
		"client",
	)

	recorder := httptest.NewRecorder()

	handler.GetByBookingID(
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

	body := recorder.Body.String()

	if !strings.Contains(
		body,
		`"booking_id":12`,
	) {
		t.Fatalf(
			"expected booking ID in response: %s",
			body,
		)
	}

	if !strings.Contains(
		body,
		`"current_status":"accepted"`,
	) {
		t.Fatalf(
			"expected current status in response: %s",
			body,
		)
	}

	if !strings.Contains(
		body,
		`"to_status":"accepted"`,
	) {
		t.Fatalf(
			"expected history in response: %s",
			body,
		)
	}
}

func TestHandler_GetByBookingID_Unauthorized(
	t *testing.T,
) {
	handler := newBookingTimelineHandlerForTest(
		&mockRepository{},
	)

	req := requestWithBookingTimelineID(
		http.MethodGet,
		"/booking-timeline/12",
		"12",
	)

	recorder := httptest.NewRecorder()

	handler.GetByBookingID(
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

func TestHandler_GetByBookingID_InvalidBookingID(
	t *testing.T,
) {
	handler := newBookingTimelineHandlerForTest(
		&mockRepository{},
	)

	req := requestWithBookingTimelineID(
		http.MethodGet,
		"/booking-timeline/abc",
		"abc",
	)

	req = requestWithBookingTimelineUser(
		req,
		5,
		"client",
	)

	recorder := httptest.NewRecorder()

	handler.GetByBookingID(
		recorder,
		req,
	)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusBadRequest,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_GetByBookingID_ZeroBookingID(
	t *testing.T,
) {
	handler := newBookingTimelineHandlerForTest(
		&mockRepository{},
	)

	req := requestWithBookingTimelineID(
		http.MethodGet,
		"/booking-timeline/0",
		"0",
	)

	req = requestWithBookingTimelineUser(
		req,
		5,
		"client",
	)

	recorder := httptest.NewRecorder()

	handler.GetByBookingID(
		recorder,
		req,
	)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusBadRequest,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_GetByBookingID_BookingNotFound(
	t *testing.T,
) {
	repo := &mockRepository{
		getBookingAccessFn: func(
			context.Context,
			uint,
		) (uint, uint, string, error) {
			return 0, 0, "", ErrBookingNotFound
		},
	}

	handler := newBookingTimelineHandlerForTest(
		repo,
	)

	req := requestWithBookingTimelineID(
		http.MethodGet,
		"/booking-timeline/999",
		"999",
	)

	req = requestWithBookingTimelineUser(
		req,
		5,
		"client",
	)

	recorder := httptest.NewRecorder()

	handler.GetByBookingID(
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

func TestHandler_GetByBookingID_Forbidden(
	t *testing.T,
) {
	repo := &mockRepository{
		getBookingAccessFn: func(
			context.Context,
			uint,
		) (uint, uint, string, error) {
			return 5, 8, "accepted", nil
		},
	}

	handler := newBookingTimelineHandlerForTest(
		repo,
	)

	req := requestWithBookingTimelineID(
		http.MethodGet,
		"/booking-timeline/12",
		"12",
	)

	req = requestWithBookingTimelineUser(
		req,
		99,
		"client",
	)

	recorder := httptest.NewRecorder()

	handler.GetByBookingID(
		recorder,
		req,
	)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusForbidden,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_GetByBookingID_InternalServerError(
	t *testing.T,
) {
	expectedErr := errors.New(
		"database failed",
	)

	repo := &mockRepository{
		getBookingAccessFn: func(
			context.Context,
			uint,
		) (uint, uint, string, error) {
			return 0, 0, "", expectedErr
		},
	}

	handler := newBookingTimelineHandlerForTest(
		repo,
	)

	req := requestWithBookingTimelineID(
		http.MethodGet,
		"/booking-timeline/12",
		"12",
	)

	req = requestWithBookingTimelineUser(
		req,
		5,
		"client",
	)

	recorder := httptest.NewRecorder()

	handler.GetByBookingID(
		recorder,
		req,
	)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusInternalServerError,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}
