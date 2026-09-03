package recentviews

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

func newHandlerForTest(repo Repository) *Handler {
	service := NewService(repo)

	return NewHandler(service)
}

func requestWithUser(
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

func requestWithURLParam(
	req *http.Request,
	name string,
	value string,
) *http.Request {
	routeContext := chi.NewRouteContext()
	routeContext.URLParams.Add(name, value)

	ctx := context.WithValue(
		req.Context(),
		chi.RouteCtxKey,
		routeContext,
	)

	return req.WithContext(ctx)
}

func TestNewHandler(t *testing.T) {
	service := NewService(&mockRepository{})

	handler := NewHandler(service)

	if handler == nil {
		t.Fatal("expected handler")
	}

	if handler.service != service {
		t.Fatal("expected service assigned")
	}
}

func TestParseIDParam_Success(t *testing.T) {
	req := httptest.NewRequest(
		http.MethodPost,
		"/recent-views/cleaners/8",
		nil,
	)
	req = requestWithURLParam(
		req,
		"cleanerID",
		"8",
	)

	id, err := parseIDParam(req, "cleanerID")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if id != 8 {
		t.Fatalf(
			"expected ID 8, got %d",
			id,
		)
	}
}

func TestParseIDParam_InvalidID(t *testing.T) {
	req := httptest.NewRequest(
		http.MethodPost,
		"/recent-views/cleaners/nope",
		nil,
	)
	req = requestWithURLParam(
		req,
		"cleanerID",
		"nope",
	)

	id, err := parseIDParam(req, "cleanerID")

	if err == nil {
		t.Fatal("expected parsing error")
	}

	if id != 0 {
		t.Fatalf(
			"expected ID 0, got %d",
			id,
		)
	}
}

func TestParseIDParam_MissingID(t *testing.T) {
	req := httptest.NewRequest(
		http.MethodPost,
		"/recent-views/cleaners/",
		nil,
	)

	id, err := parseIDParam(req, "cleanerID")

	if err == nil {
		t.Fatal("expected parsing error")
	}

	if id != 0 {
		t.Fatalf(
			"expected ID 0, got %d",
			id,
		)
	}
}

func TestHandler_RecordView_Success(t *testing.T) {
	repo := &mockRepository{
		recordViewFn: func(
			ctx context.Context,
			clientID uint,
			cleanerID uint,
		) error {
			if clientID != 5 {
				t.Fatalf(
					"expected client ID 5, got %d",
					clientID,
				)
			}

			if cleanerID != 8 {
				t.Fatalf(
					"expected cleaner ID 8, got %d",
					cleanerID,
				)
			}

			return nil
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodPost,
		"/recent-views/cleaners/8",
		nil,
	)
	req = requestWithURLParam(
		req,
		"cleanerID",
		"8",
	)
	req = requestWithUser(
		req,
		5,
		"client",
	)

	recorder := httptest.NewRecorder()

	handler.RecordView(recorder, req)

	if recorder.Code != http.StatusCreated {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusCreated,
			recorder.Code,
			recorder.Body.String(),
		)
	}

	if !strings.Contains(
		recorder.Body.String(),
		`"message":"cleaner view recorded"`,
	) {
		t.Fatalf(
			"expected success message: %s",
			recorder.Body.String(),
		)
	}
}

func TestHandler_RecordView_Unauthorized(t *testing.T) {
	handler := newHandlerForTest(
		&mockRepository{},
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/recent-views/cleaners/8",
		nil,
	)

	recorder := httptest.NewRecorder()

	handler.RecordView(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusUnauthorized,
			recorder.Code,
		)
	}
}

func TestHandler_RecordView_InvalidCleanerID(t *testing.T) {
	handler := newHandlerForTest(
		&mockRepository{},
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/recent-views/cleaners/nope",
		nil,
	)
	req = requestWithURLParam(
		req,
		"cleanerID",
		"nope",
	)
	req = requestWithUser(
		req,
		5,
		"client",
	)

	recorder := httptest.NewRecorder()

	handler.RecordView(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusBadRequest,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_RecordView_InvalidInput(t *testing.T) {
	handler := newHandlerForTest(
		&mockRepository{},
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/recent-views/cleaners/8",
		nil,
	)
	req = requestWithURLParam(
		req,
		"cleanerID",
		"8",
	)
	req = requestWithUser(
		req,
		0,
		"client",
	)

	recorder := httptest.NewRecorder()

	handler.RecordView(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusBadRequest,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_RecordView_SameClientAndCleaner(t *testing.T) {
	handler := newHandlerForTest(
		&mockRepository{},
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/recent-views/cleaners/5",
		nil,
	)
	req = requestWithURLParam(
		req,
		"cleanerID",
		"5",
	)
	req = requestWithUser(
		req,
		5,
		"client",
	)

	recorder := httptest.NewRecorder()

	handler.RecordView(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusBadRequest,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_RecordView_InternalServerError(t *testing.T) {
	expectedErr := errors.New("record view failed")

	repo := &mockRepository{
		recordViewFn: func(
			context.Context,
			uint,
			uint,
		) error {
			return expectedErr
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodPost,
		"/recent-views/cleaners/8",
		nil,
	)
	req = requestWithURLParam(
		req,
		"cleanerID",
		"8",
	)
	req = requestWithUser(
		req,
		5,
		"client",
	)

	recorder := httptest.NewRecorder()

	handler.RecordView(recorder, req)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusInternalServerError,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_ListMine_Success(t *testing.T) {
	repo := &mockRepository{
		listByClientIDFn: func(
			context.Context,
			uint,
		) ([]RecentlyViewedCleaner, error) {
			return []RecentlyViewedCleaner{
				{
					ID:        1,
					ClientID:  5,
					CleanerID: 8,
				},
			}, nil
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodGet,
		"/recent-views/me",
		nil,
	)
	req = requestWithUser(
		req,
		5,
		"client",
	)

	recorder := httptest.NewRecorder()

	handler.ListMine(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusOK,
			recorder.Code,
			recorder.Body.String(),
		)
	}

	if !strings.Contains(
		recorder.Body.String(),
		`"cleaner_id":8`,
	) {
		t.Fatalf(
			"expected cleaner ID in response: %s",
			recorder.Body.String(),
		)
	}
}

func TestHandler_ListMine_Empty(t *testing.T) {
	repo := &mockRepository{
		listByClientIDFn: func(
			context.Context,
			uint,
		) ([]RecentlyViewedCleaner, error) {
			return []RecentlyViewedCleaner{}, nil
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodGet,
		"/recent-views/me",
		nil,
	)
	req = requestWithUser(
		req,
		5,
		"client",
	)

	recorder := httptest.NewRecorder()

	handler.ListMine(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusOK,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_ListMine_Unauthorized(t *testing.T) {
	handler := newHandlerForTest(
		&mockRepository{},
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/recent-views/me",
		nil,
	)

	recorder := httptest.NewRecorder()

	handler.ListMine(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusUnauthorized,
			recorder.Code,
		)
	}
}

func TestHandler_ListMine_InvalidInput(t *testing.T) {
	handler := newHandlerForTest(
		&mockRepository{},
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/recent-views/me",
		nil,
	)
	req = requestWithUser(
		req,
		0,
		"client",
	)

	recorder := httptest.NewRecorder()

	handler.ListMine(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusBadRequest,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_ListMine_InternalServerError(t *testing.T) {
	expectedErr := errors.New("list views failed")

	repo := &mockRepository{
		listByClientIDFn: func(
			context.Context,
			uint,
		) ([]RecentlyViewedCleaner, error) {
			return nil, expectedErr
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodGet,
		"/recent-views/me",
		nil,
	)
	req = requestWithUser(
		req,
		5,
		"client",
	)

	recorder := httptest.NewRecorder()

	handler.ListMine(recorder, req)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusInternalServerError,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_GetMyProfileViewAnalytics_Success(
	t *testing.T,
) {
	repo := &mockRepository{
		countByCleanerIDFn: func(
			context.Context,
			uint,
		) (int, error) {
			return 50, nil
		},

		countUniqueViewersByCleanerIDFn: func(
			context.Context,
			uint,
		) (int, error) {
			return 20, nil
		},

		countByCleanerIDSinceFn: func(
			context.Context,
			uint,
			time.Time,
		) (int, error) {
			return 7, nil
		},

		countByCleanerIDBetweenFn: func(
			context.Context,
			uint,
			time.Time,
			time.Time,
		) (int, error) {
			return 5, nil
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodGet,
		"/recent-views/me/analytics",
		nil,
	)

	req = requestWithUser(
		req,
		8,
		"cleaner",
	)

	recorder := httptest.NewRecorder()

	handler.GetMyProfileViewAnalytics(
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

	if !strings.Contains(
		recorder.Body.String(),
		`"cleaner_id":8`,
	) {
		t.Fatalf(
			"expected cleaner ID in response: %s",
			recorder.Body.String(),
		)
	}

	if !strings.Contains(
		recorder.Body.String(),
		`"total_views":50`,
	) {
		t.Fatalf(
			"expected total views in response: %s",
			recorder.Body.String(),
		)
	}

	if !strings.Contains(
		recorder.Body.String(),
		`"unique_viewers":20`,
	) {
		t.Fatalf(
			"expected unique viewers in response: %s",
			recorder.Body.String(),
		)
	}
}

func TestHandler_GetMyProfileViewAnalytics_Unauthorized(
	t *testing.T,
) {
	handler := newHandlerForTest(
		&mockRepository{},
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/recent-views/me/analytics",
		nil,
	)

	recorder := httptest.NewRecorder()

	handler.GetMyProfileViewAnalytics(
		recorder,
		req,
	)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusUnauthorized,
			recorder.Code,
		)
	}
}

func TestHandler_GetMyProfileViewAnalytics_InvalidInput(
	t *testing.T,
) {
	handler := newHandlerForTest(
		&mockRepository{},
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/recent-views/me/analytics",
		nil,
	)

	req = requestWithUser(
		req,
		0,
		"cleaner",
	)

	recorder := httptest.NewRecorder()

	handler.GetMyProfileViewAnalytics(
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

func TestHandler_GetMyProfileViewAnalytics_InternalServerError(
	t *testing.T,
) {
	expectedErr := errors.New(
		"analytics failed",
	)

	repo := &mockRepository{
		countByCleanerIDFn: func(
			context.Context,
			uint,
		) (int, error) {
			return 0, expectedErr
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodGet,
		"/recent-views/me/analytics",
		nil,
	)

	req = requestWithUser(
		req,
		8,
		"cleaner",
	)

	recorder := httptest.NewRecorder()

	handler.GetMyProfileViewAnalytics(
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
