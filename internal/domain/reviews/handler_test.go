package reviews

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
		http.MethodGet,
		"/reviews/cleaners/8",
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
		http.MethodGet,
		"/reviews/cleaners/nope",
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
		http.MethodGet,
		"/reviews/cleaners/",
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

func TestHandler_Create_Success(t *testing.T) {
	repo := &mockRepository{
		getJobClientIDFn: func(
			context.Context,
			uint,
		) (uint, error) {
			return 5, nil
		},
		getJobStatusFn: func(
			context.Context,
			uint,
		) (string, error) {
			return "completed", nil
		},
		getAcceptedCleanerIDFn: func(
			context.Context,
			uint,
		) (uint, error) {
			return 8, nil
		},
		createFn: func(
			ctx context.Context,
			review *Review,
		) error {
			if review.JobID != 12 {
				t.Fatalf(
					"expected job ID 12, got %d",
					review.JobID,
				)
			}

			if review.BookingID != 20 {
				t.Fatalf(
					"expected booking ID 20, got %d",
					review.BookingID,
				)
			}

			if review.ClientID != 5 {
				t.Fatalf(
					"expected client ID 5, got %d",
					review.ClientID,
				)
			}

			if review.CleanerID != 8 {
				t.Fatalf(
					"expected cleaner ID 8, got %d",
					review.CleanerID,
				)
			}

			if review.Rating != 5 {
				t.Fatalf(
					"expected rating 5, got %d",
					review.Rating,
				)
			}

			if review.Comment != "Excellent cleaner" {
				t.Fatalf(
					"expected trimmed comment, got %q",
					review.Comment,
				)
			}

			review.ID = 30

			return nil
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodPost,
		"/reviews/jobs/12",
		strings.NewReader(`{
			"rating":5,
			"comment":"Excellent cleaner",
			"cleaner_id":999,
			"booking_id":20
		}`),
	)
	req = requestWithURLParam(
		req,
		"jobID",
		"12",
	)
	req = requestWithUser(
		req,
		5,
		"client",
	)

	recorder := httptest.NewRecorder()

	handler.Create(recorder, req)

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
		`"id":30`,
	) {
		t.Fatalf(
			"expected review ID in response: %s",
			recorder.Body.String(),
		)
	}

	if !strings.Contains(
		recorder.Body.String(),
		`"booking_id":20`,
	) {
		t.Fatalf(
			"expected booking ID in response: %s",
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

func TestHandler_Create_Unauthorized(t *testing.T) {
	handler := newHandlerForTest(
		&mockRepository{},
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/reviews/jobs/12",
		strings.NewReader(`{}`),
	)

	recorder := httptest.NewRecorder()

	handler.Create(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusUnauthorized,
			recorder.Code,
		)
	}
}

func TestHandler_Create_InvalidJobID(t *testing.T) {
	handler := newHandlerForTest(
		&mockRepository{},
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/reviews/jobs/nope",
		strings.NewReader(`{}`),
	)
	req = requestWithURLParam(
		req,
		"jobID",
		"nope",
	)
	req = requestWithUser(
		req,
		5,
		"client",
	)

	recorder := httptest.NewRecorder()

	handler.Create(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusBadRequest,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Create_InvalidJSON(t *testing.T) {
	handler := newHandlerForTest(
		&mockRepository{},
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/reviews/jobs/12",
		strings.NewReader(`{"rating":`),
	)
	req = requestWithURLParam(
		req,
		"jobID",
		"12",
	)
	req = requestWithUser(
		req,
		5,
		"client",
	)

	recorder := httptest.NewRecorder()

	handler.Create(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusBadRequest,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Create_InvalidInput(t *testing.T) {
	handler := newHandlerForTest(
		&mockRepository{},
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/reviews/jobs/12",
		strings.NewReader(`{
			"rating":0,
			"comment":"",
			"booking_id":20
		}`),
	)
	req = requestWithURLParam(
		req,
		"jobID",
		"12",
	)
	req = requestWithUser(
		req,
		5,
		"client",
	)

	recorder := httptest.NewRecorder()

	handler.Create(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusBadRequest,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Create_ZeroUserID(t *testing.T) {
	handler := newHandlerForTest(
		&mockRepository{},
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/reviews/jobs/12",
		strings.NewReader(`{
			"rating":5,
			"comment":"Excellent",
			"booking_id":20
		}`),
	)
	req = requestWithURLParam(
		req,
		"jobID",
		"12",
	)
	req = requestWithUser(
		req,
		0,
		"client",
	)

	recorder := httptest.NewRecorder()

	handler.Create(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusBadRequest,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Create_Forbidden(t *testing.T) {
	repo := &mockRepository{
		getJobClientIDFn: func(
			context.Context,
			uint,
		) (uint, error) {
			return 99, nil
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodPost,
		"/reviews/jobs/12",
		strings.NewReader(`{
			"rating":5,
			"comment":"Excellent",
			"booking_id":20
		}`),
	)
	req = requestWithURLParam(
		req,
		"jobID",
		"12",
	)
	req = requestWithUser(
		req,
		5,
		"client",
	)

	recorder := httptest.NewRecorder()

	handler.Create(recorder, req)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusForbidden,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Create_JobNotCompleted(t *testing.T) {
	repo := &mockRepository{
		getJobClientIDFn: func(
			context.Context,
			uint,
		) (uint, error) {
			return 5, nil
		},
		getJobStatusFn: func(
			context.Context,
			uint,
		) (string, error) {
			return "open", nil
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodPost,
		"/reviews/jobs/12",
		strings.NewReader(`{
			"rating":5,
			"comment":"Excellent",
			"booking_id":20
		}`),
	)
	req = requestWithURLParam(
		req,
		"jobID",
		"12",
	)
	req = requestWithUser(
		req,
		5,
		"client",
	)

	recorder := httptest.NewRecorder()

	handler.Create(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusBadRequest,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Create_AlreadyExists(t *testing.T) {
	repo := &mockRepository{
		getJobClientIDFn: func(
			context.Context,
			uint,
		) (uint, error) {
			return 5, nil
		},
		getJobStatusFn: func(
			context.Context,
			uint,
		) (string, error) {
			return "completed", nil
		},
		getAcceptedCleanerIDFn: func(
			context.Context,
			uint,
		) (uint, error) {
			return 8, nil
		},
		createFn: func(
			context.Context,
			*Review,
		) error {
			return ErrReviewAlreadyExist
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodPost,
		"/reviews/jobs/12",
		strings.NewReader(`{
			"rating":5,
			"comment":"Excellent",
			"booking_id":20
		}`),
	)
	req = requestWithURLParam(
		req,
		"jobID",
		"12",
	)
	req = requestWithUser(
		req,
		5,
		"client",
	)

	recorder := httptest.NewRecorder()

	handler.Create(recorder, req)

	if recorder.Code != http.StatusConflict {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusConflict,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Create_InternalServerError(t *testing.T) {
	expectedErr := errors.New("create review failed")

	repo := &mockRepository{
		getJobClientIDFn: func(
			context.Context,
			uint,
		) (uint, error) {
			return 0, expectedErr
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodPost,
		"/reviews/jobs/12",
		strings.NewReader(`{
			"rating":5,
			"comment":"Excellent",
			"booking_id":20
		}`),
	)
	req = requestWithURLParam(
		req,
		"jobID",
		"12",
	)
	req = requestWithUser(
		req,
		5,
		"client",
	)

	recorder := httptest.NewRecorder()

	handler.Create(recorder, req)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusInternalServerError,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_ListByCleaner_Success(t *testing.T) {
	repo := &mockRepository{
		listByCleanerIDFn: func(
			ctx context.Context,
			cleanerID uint,
		) ([]Review, error) {
			if cleanerID != 8 {
				t.Fatalf(
					"expected cleaner ID 8, got %d",
					cleanerID,
				)
			}

			return []Review{
				{
					ID:        30,
					BookingID: 20,
					CleanerID: 8,
					ClientID:  5,
					JobID:     12,
					Rating:    5,
					Comment:   "Excellent",
				},
			}, nil
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodGet,
		"/reviews/cleaners/8",
		nil,
	)
	req = requestWithURLParam(
		req,
		"cleanerID",
		"8",
	)

	recorder := httptest.NewRecorder()

	handler.ListByCleaner(recorder, req)

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
		`"booking_id":20`,
	) {
		t.Fatalf(
			"expected booking ID in response: %s",
			recorder.Body.String(),
		)
	}

	if !strings.Contains(
		recorder.Body.String(),
		`"comment":"Excellent"`,
	) {
		t.Fatalf(
			"expected review in response: %s",
			recorder.Body.String(),
		)
	}
}

func TestHandler_ListByCleaner_Empty(t *testing.T) {
	repo := &mockRepository{
		listByCleanerIDFn: func(
			context.Context,
			uint,
		) ([]Review, error) {
			return []Review{}, nil
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodGet,
		"/reviews/cleaners/8",
		nil,
	)
	req = requestWithURLParam(
		req,
		"cleanerID",
		"8",
	)

	recorder := httptest.NewRecorder()

	handler.ListByCleaner(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusOK,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_ListByCleaner_InvalidCleanerID(t *testing.T) {
	handler := newHandlerForTest(
		&mockRepository{},
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/reviews/cleaners/nope",
		nil,
	)
	req = requestWithURLParam(
		req,
		"cleanerID",
		"nope",
	)

	recorder := httptest.NewRecorder()

	handler.ListByCleaner(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusBadRequest,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_ListByCleaner_InvalidInput(t *testing.T) {
	handler := newHandlerForTest(
		&mockRepository{},
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/reviews/cleaners/0",
		nil,
	)
	req = requestWithURLParam(
		req,
		"cleanerID",
		"0",
	)

	recorder := httptest.NewRecorder()

	handler.ListByCleaner(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusBadRequest,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_ListByCleaner_InternalServerError(t *testing.T) {
	expectedErr := errors.New("list reviews failed")

	repo := &mockRepository{
		listByCleanerIDFn: func(
			context.Context,
			uint,
		) ([]Review, error) {
			return nil, expectedErr
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodGet,
		"/reviews/cleaners/8",
		nil,
	)
	req = requestWithURLParam(
		req,
		"cleanerID",
		"8",
	)

	recorder := httptest.NewRecorder()

	handler.ListByCleaner(recorder, req)

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
			ctx context.Context,
			clientID uint,
		) ([]Review, error) {
			if clientID != 5 {
				t.Fatalf(
					"expected client ID 5, got %d",
					clientID,
				)
			}

			return []Review{
				{
					ID:        30,
					BookingID: 20,
					CleanerID: 8,
					ClientID:  5,
					JobID:     12,
					Rating:    5,
					Comment:   "Excellent",
				},
			}, nil
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodGet,
		"/reviews/me",
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
		`"client_id":5`,
	) {
		t.Fatalf(
			"expected client ID in response: %s",
			recorder.Body.String(),
		)
	}
}

func TestHandler_ListMine_Empty(t *testing.T) {
	repo := &mockRepository{
		listByClientIDFn: func(
			context.Context,
			uint,
		) ([]Review, error) {
			return []Review{}, nil
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodGet,
		"/reviews/me",
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
		"/reviews/me",
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
		"/reviews/me",
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
	expectedErr := errors.New("list reviews failed")

	repo := &mockRepository{
		listByClientIDFn: func(
			context.Context,
			uint,
		) ([]Review, error) {
			return nil, expectedErr
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodGet,
		"/reviews/me",
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
