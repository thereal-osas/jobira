package applicationtimeline

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/rodrigueghenda/jobira/internal/security/identity"
)

func TestHandler_GetTimeline_SuccessCleaner(t *testing.T) {
	now := time.Now()

	repo := &mockRepository{
		getApplicationCleanerIDFn: func(
			ctx context.Context,
			applicationID uint,
		) (uint, error) {
			return 20, nil
		},
		getApplicationClientIDFn: func(
			ctx context.Context,
			applicationID uint,
		) (uint, error) {
			return 30, nil
		},
		getCurrentStatusFn: func(
			ctx context.Context,
			applicationID uint,
		) (string, error) {
			return "shortlisted", nil
		},
		listByApplicationIDFn: func(
			ctx context.Context,
			applicationID uint,
		) ([]Event, error) {
			return []Event{
				{
					ID:            1,
					ApplicationID: 10,
					Status:        "pending",
					Note:          "Application submitted",
					CreatedAt:     now,
				},
				{
					ID:            2,
					ApplicationID: 10,
					Status:        "shortlisted",
					Note:          "Application shortlisted",
					CreatedAt:     now.Add(time.Minute),
				},
			}, nil
		},
	}

	service := NewService(repo)
	handler := NewHandler(service)

	req := httptest.NewRequest(
		http.MethodGet,
		"/application-timeline/applications/10",
		nil,
	)

	req = requestWithTimelineURLParam(
		req,
		"applicationID",
		"10",
	)

	req = requestWithTimelineIdentity(
		req,
		20,
		"cleaner",
	)

	rec := httptest.NewRecorder()

	handler.GetTimeline(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusOK,
			rec.Code,
			rec.Body.String(),
		)
	}

	body := rec.Body.String()

	if !containsTimelineString(
		body,
		`"application_id":10`,
	) {
		t.Fatalf(
			"expected application id in response: %s",
			body,
		)
	}

	if !containsTimelineString(
		body,
		`"current_status":"shortlisted"`,
	) {
		t.Fatalf(
			"expected current status in response: %s",
			body,
		)
	}

	if !containsTimelineString(
		body,
		`"status":"pending"`,
	) {
		t.Fatalf(
			"expected pending event in response: %s",
			body,
		)
	}

	if !containsTimelineString(
		body,
		`"status":"shortlisted"`,
	) {
		t.Fatalf(
			"expected shortlisted event in response: %s",
			body,
		)
	}
}

func TestHandler_GetTimeline_SuccessClient(t *testing.T) {
	repo := successfulTimelineRepository()

	service := NewService(repo)
	handler := NewHandler(service)

	req := httptest.NewRequest(
		http.MethodGet,
		"/application-timeline/applications/10",
		nil,
	)

	req = requestWithTimelineURLParam(
		req,
		"applicationID",
		"10",
	)

	req = requestWithTimelineIdentity(
		req,
		30,
		"client",
	)

	rec := httptest.NewRecorder()

	handler.GetTimeline(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusOK,
			rec.Code,
			rec.Body.String(),
		)
	}
}

func TestHandler_GetTimeline_SuccessAdmin(t *testing.T) {
	repo := successfulTimelineRepository()

	service := NewService(repo)
	handler := NewHandler(service)

	req := httptest.NewRequest(
		http.MethodGet,
		"/application-timeline/applications/10",
		nil,
	)

	req = requestWithTimelineURLParam(
		req,
		"applicationID",
		"10",
	)

	req = requestWithTimelineIdentity(
		req,
		999,
		"admin",
	)

	rec := httptest.NewRecorder()

	handler.GetTimeline(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusOK,
			rec.Code,
			rec.Body.String(),
		)
	}
}

func TestHandler_GetTimeline_Unauthorized(t *testing.T) {
	service := NewService(&mockRepository{})
	handler := NewHandler(service)

	req := httptest.NewRequest(
		http.MethodGet,
		"/application-timeline/applications/10",
		nil,
	)

	req = requestWithTimelineURLParam(
		req,
		"applicationID",
		"10",
	)

	rec := httptest.NewRecorder()

	handler.GetTimeline(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusUnauthorized,
			rec.Code,
		)
	}
}

func TestHandler_GetTimeline_InvalidApplicationID(t *testing.T) {
	service := NewService(&mockRepository{})
	handler := NewHandler(service)

	req := httptest.NewRequest(
		http.MethodGet,
		"/application-timeline/applications/invalid",
		nil,
	)

	req = requestWithTimelineURLParam(
		req,
		"applicationID",
		"invalid",
	)

	req = requestWithTimelineIdentity(
		req,
		20,
		"cleaner",
	)

	rec := httptest.NewRecorder()

	handler.GetTimeline(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			rec.Code,
		)
	}
}

func TestHandler_GetTimeline_ZeroApplicationID(t *testing.T) {
	service := NewService(&mockRepository{})
	handler := NewHandler(service)

	req := httptest.NewRequest(
		http.MethodGet,
		"/application-timeline/applications/0",
		nil,
	)

	req = requestWithTimelineURLParam(
		req,
		"applicationID",
		"0",
	)

	req = requestWithTimelineIdentity(
		req,
		20,
		"cleaner",
	)

	rec := httptest.NewRecorder()

	handler.GetTimeline(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			rec.Code,
		)
	}
}

func TestHandler_GetTimeline_ApplicationNotFound(t *testing.T) {
	repo := &mockRepository{
		getApplicationCleanerIDFn: func(
			ctx context.Context,
			applicationID uint,
		) (uint, error) {
			return 0, ErrApplicationNotFound
		},
	}

	service := NewService(repo)
	handler := NewHandler(service)

	req := httptest.NewRequest(
		http.MethodGet,
		"/application-timeline/applications/999",
		nil,
	)

	req = requestWithTimelineURLParam(
		req,
		"applicationID",
		"999",
	)

	req = requestWithTimelineIdentity(
		req,
		20,
		"cleaner",
	)

	rec := httptest.NewRecorder()

	handler.GetTimeline(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusNotFound,
			rec.Code,
			rec.Body.String(),
		)
	}
}

func TestHandler_GetTimeline_TimelineNotFound(t *testing.T) {
	repo := &mockRepository{
		getApplicationCleanerIDFn: func(
			ctx context.Context,
			applicationID uint,
		) (uint, error) {
			return 20, nil
		},
		getApplicationClientIDFn: func(
			ctx context.Context,
			applicationID uint,
		) (uint, error) {
			return 30, nil
		},
		getCurrentStatusFn: func(
			ctx context.Context,
			applicationID uint,
		) (string, error) {
			return "pending", nil
		},
		listByApplicationIDFn: func(
			ctx context.Context,
			applicationID uint,
		) ([]Event, error) {
			return nil, ErrTimelineNotFound
		},
	}

	service := NewService(repo)
	handler := NewHandler(service)

	req := httptest.NewRequest(
		http.MethodGet,
		"/application-timeline/applications/10",
		nil,
	)

	req = requestWithTimelineURLParam(
		req,
		"applicationID",
		"10",
	)

	req = requestWithTimelineIdentity(
		req,
		20,
		"cleaner",
	)

	rec := httptest.NewRecorder()

	handler.GetTimeline(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusNotFound,
			rec.Code,
			rec.Body.String(),
		)
	}
}

func TestHandler_GetTimeline_Forbidden(t *testing.T) {
	repo := &mockRepository{
		getApplicationCleanerIDFn: func(
			ctx context.Context,
			applicationID uint,
		) (uint, error) {
			return 20, nil
		},
		getApplicationClientIDFn: func(
			ctx context.Context,
			applicationID uint,
		) (uint, error) {
			return 30, nil
		},
	}

	service := NewService(repo)
	handler := NewHandler(service)

	req := httptest.NewRequest(
		http.MethodGet,
		"/application-timeline/applications/10",
		nil,
	)

	req = requestWithTimelineURLParam(
		req,
		"applicationID",
		"10",
	)

	req = requestWithTimelineIdentity(
		req,
		99,
		"user",
	)

	rec := httptest.NewRecorder()

	handler.GetTimeline(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusForbidden,
			rec.Code,
			rec.Body.String(),
		)
	}
}

func TestHandler_GetTimeline_InternalServerError(t *testing.T) {
	expectedErr := errors.New("database exploded")

	repo := &mockRepository{
		getApplicationCleanerIDFn: func(
			ctx context.Context,
			applicationID uint,
		) (uint, error) {
			return 0, expectedErr
		},
	}

	service := NewService(repo)
	handler := NewHandler(service)

	req := httptest.NewRequest(
		http.MethodGet,
		"/application-timeline/applications/10",
		nil,
	)

	req = requestWithTimelineURLParam(
		req,
		"applicationID",
		"10",
	)

	req = requestWithTimelineIdentity(
		req,
		20,
		"cleaner",
	)

	rec := httptest.NewRecorder()

	handler.GetTimeline(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusInternalServerError,
			rec.Code,
			rec.Body.String(),
		)
	}
}

func TestParseIDParam_Success(t *testing.T) {
	req := httptest.NewRequest(
		http.MethodGet,
		"/10",
		nil,
	)

	req = requestWithTimelineURLParam(
		req,
		"applicationID",
		"10",
	)

	id, err := parseIDParam(
		req,
		"applicationID",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if id != 10 {
		t.Fatalf(
			"expected id 10, got %d",
			id,
		)
	}
}

func TestParseIDParam_Invalid(t *testing.T) {
	req := httptest.NewRequest(
		http.MethodGet,
		"/invalid",
		nil,
	)

	req = requestWithTimelineURLParam(
		req,
		"applicationID",
		"invalid",
	)

	_, err := parseIDParam(
		req,
		"applicationID",
	)

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf(
			"expected ErrInvalidInput, got %v",
			err,
		)
	}
}

func TestParseIDParam_Zero(t *testing.T) {
	req := httptest.NewRequest(
		http.MethodGet,
		"/0",
		nil,
	)

	req = requestWithTimelineURLParam(
		req,
		"applicationID",
		"0",
	)

	_, err := parseIDParam(
		req,
		"applicationID",
	)

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf(
			"expected ErrInvalidInput, got %v",
			err,
		)
	}
}

func requestWithTimelineURLParam(
	req *http.Request,
	key string,
	value string,
) *http.Request {
	routeCtx := chi.RouteContext(req.Context())

	if routeCtx == nil {
		routeCtx = chi.NewRouteContext()

		ctx := context.WithValue(
			req.Context(),
			chi.RouteCtxKey,
			routeCtx,
		)

		req = req.WithContext(ctx)
	}

	routeCtx.URLParams.Add(key, value)

	return req
}

func requestWithTimelineIdentity(
	req *http.Request,
	userID uint,
	role string,
) *http.Request {
	currentUser := identity.UserIdentity{
		UserID: userID,
		Role:   role,
	}

	ctx := identity.WithUser(
		req.Context(),
		currentUser,
	)

	return req.WithContext(ctx)
}

func containsTimelineString(
	value string,
	expected string,
) bool {
	for i := 0; i+len(expected) <= len(value); i++ {
		if value[i:i+len(expected)] == expected {
			return true
		}
	}

	return false
}
