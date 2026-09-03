package adminsubscriptions

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

func adminSubscriptionsRequestWithParam(
	req *http.Request,
	name string,
	value string,
) *http.Request {
	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add(name, value)

	ctx := context.WithValue(
		req.Context(),
		chi.RouteCtxKey,
		routeCtx,
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

func TestHandler_ListPlans_Success(t *testing.T) {
	repo := &mockRepository{
		listPlansFn: func(
			context.Context,
		) ([]AdminSubscriptionPlan, error) {
			return []AdminSubscriptionPlan{
				{
					ID:   1,
					Name: "Standard",
				},
			}, nil
		},
	}

	handler := NewHandler(
		NewService(repo),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/admin/subscriptions/plans",
		nil,
	)

	rec := httptest.NewRecorder()

	handler.ListPlans(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusOK,
			rec.Code,
		)
	}
}

func TestHandler_ListPlans_InternalServerError(
	t *testing.T,
) {
	repo := &mockRepository{
		listPlansFn: func(
			context.Context,
		) ([]AdminSubscriptionPlan, error) {
			return nil, errors.New("failed")
		},
	}

	handler := NewHandler(
		NewService(repo),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/admin/subscriptions/plans",
		nil,
	)

	rec := httptest.NewRecorder()

	handler.ListPlans(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusInternalServerError,
			rec.Code,
		)
	}
}

func TestHandler_CreatePlan_Success(t *testing.T) {
	repo := &mockRepository{
		createPlanFn: func(
			context.Context,
			*AdminSubscriptionPlan,
		) error {
			return nil
		},
	}

	handler := NewHandler(
		NewService(repo),
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/admin/subscriptions/plans",
		bytes.NewBufferString(`{
			"name":"Standard",
			"role_type":"cleaner",
			"price_pence":1799,
			"billing_interval":"monthly",
			"application_limit":50,
			"job_post_limit":0,
			"cleaner_seat_limit":1
		}`),
	)

	rec := httptest.NewRecorder()

	handler.CreatePlan(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusCreated,
			rec.Code,
			rec.Body.String(),
		)
	}
}

func TestHandler_CreatePlan_InvalidBody(
	t *testing.T,
) {
	handler := NewHandler(
		NewService(&mockRepository{}),
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/admin/subscriptions/plans",
		bytes.NewBufferString(`{invalid}`),
	)

	rec := httptest.NewRecorder()

	handler.CreatePlan(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			rec.Code,
		)
	}
}

func TestHandler_CreatePlan_InvalidInput(
	t *testing.T,
) {
	handler := NewHandler(
		NewService(&mockRepository{}),
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/admin/subscriptions/plans",
		bytes.NewBufferString(`{
			"name":"",
			"role_type":"cleaner",
			"billing_interval":"monthly",
			"cleaner_seat_limit":1
		}`),
	)

	rec := httptest.NewRecorder()

	handler.CreatePlan(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			rec.Code,
		)
	}
}

func TestHandler_CreatePlan_InternalServerError(
	t *testing.T,
) {
	repo := &mockRepository{
		createPlanFn: func(
			context.Context,
			*AdminSubscriptionPlan,
		) error {
			return errors.New("failed")
		},
	}

	handler := NewHandler(
		NewService(repo),
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/admin/subscriptions/plans",
		bytes.NewBufferString(`{
			"name":"Standard",
			"role_type":"cleaner",
			"billing_interval":"monthly",
			"cleaner_seat_limit":1
		}`),
	)

	rec := httptest.NewRecorder()

	handler.CreatePlan(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusInternalServerError,
			rec.Code,
		)
	}
}

func TestHandler_UpdatePlan_Success(t *testing.T) {
	repo := &mockRepository{
		updatePlanFn: func(
			context.Context,
			*AdminSubscriptionPlan,
		) error {
			return nil
		},
	}

	handler := NewHandler(
		NewService(repo),
	)

	req := httptest.NewRequest(
		http.MethodPut,
		"/admin/subscriptions/plans/2",
		bytes.NewBufferString(`{
			"name":"Premium",
			"role_type":"cleaner",
			"price_pence":2499,
			"billing_interval":"monthly",
			"application_limit":100,
			"job_post_limit":0,
			"cleaner_seat_limit":1,
			"is_active":true
		}`),
	)

	req = adminSubscriptionsRequestWithParam(
		req,
		"planID",
		"2",
	)

	rec := httptest.NewRecorder()

	handler.UpdatePlan(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusOK,
			rec.Code,
			rec.Body.String(),
		)
	}
}

func TestHandler_UpdatePlan_InvalidID(t *testing.T) {
	handler := NewHandler(
		NewService(&mockRepository{}),
	)

	req := httptest.NewRequest(
		http.MethodPut,
		"/admin/subscriptions/plans/invalid",
		bytes.NewBufferString(`{}`),
	)

	req = adminSubscriptionsRequestWithParam(
		req,
		"planID",
		"invalid",
	)

	rec := httptest.NewRecorder()

	handler.UpdatePlan(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			rec.Code,
		)
	}
}

func TestHandler_UpdatePlan_InvalidBody(t *testing.T) {
	handler := NewHandler(
		NewService(&mockRepository{}),
	)

	req := httptest.NewRequest(
		http.MethodPut,
		"/admin/subscriptions/plans/2",
		bytes.NewBufferString(`{invalid}`),
	)

	req = adminSubscriptionsRequestWithParam(
		req,
		"planID",
		"2",
	)

	rec := httptest.NewRecorder()

	handler.UpdatePlan(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			rec.Code,
		)
	}
}

func TestHandler_UpdatePlan_InvalidInput(
	t *testing.T,
) {
	handler := NewHandler(
		NewService(&mockRepository{}),
	)

	req := httptest.NewRequest(
		http.MethodPut,
		"/admin/subscriptions/plans/2",
		bytes.NewBufferString(`{
			"name":"",
			"role_type":"cleaner",
			"billing_interval":"monthly",
			"cleaner_seat_limit":1
		}`),
	)

	req = adminSubscriptionsRequestWithParam(
		req,
		"planID",
		"2",
	)

	rec := httptest.NewRecorder()

	handler.UpdatePlan(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			rec.Code,
		)
	}
}

func TestHandler_UpdatePlan_NotFound(t *testing.T) {
	repo := &mockRepository{
		updatePlanFn: func(
			context.Context,
			*AdminSubscriptionPlan,
		) error {
			return ErrPlanNotFound
		},
	}

	handler := NewHandler(
		NewService(repo),
	)

	req := httptest.NewRequest(
		http.MethodPut,
		"/admin/subscriptions/plans/999",
		bytes.NewBufferString(`{
			"name":"Plan",
			"role_type":"cleaner",
			"billing_interval":"monthly",
			"cleaner_seat_limit":1
		}`),
	)

	req = adminSubscriptionsRequestWithParam(
		req,
		"planID",
		"999",
	)

	rec := httptest.NewRecorder()

	handler.UpdatePlan(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusNotFound,
			rec.Code,
		)
	}
}

func TestHandler_DisablePlan_Success(t *testing.T) {
	repo := &mockRepository{
		disablePlanFn: func(
			context.Context,
			uint,
		) error {
			return nil
		},
	}

	handler := NewHandler(
		NewService(repo),
	)

	req := httptest.NewRequest(
		http.MethodDelete,
		"/admin/subscriptions/plans/2",
		nil,
	)

	req = adminSubscriptionsRequestWithParam(
		req,
		"planID",
		"2",
	)

	rec := httptest.NewRecorder()

	handler.DisablePlan(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusOK,
			rec.Code,
		)
	}
}

func TestHandler_DisablePlan_InvalidID(t *testing.T) {
	handler := NewHandler(
		NewService(&mockRepository{}),
	)

	req := httptest.NewRequest(
		http.MethodDelete,
		"/admin/subscriptions/plans/x",
		nil,
	)

	req = adminSubscriptionsRequestWithParam(
		req,
		"planID",
		"x",
	)

	rec := httptest.NewRecorder()

	handler.DisablePlan(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			rec.Code,
		)
	}
}

func TestHandler_DisablePlan_NotFound(t *testing.T) {
	repo := &mockRepository{
		disablePlanFn: func(
			context.Context,
			uint,
		) error {
			return ErrPlanNotFound
		},
	}

	handler := NewHandler(
		NewService(repo),
	)

	req := httptest.NewRequest(
		http.MethodDelete,
		"/admin/subscriptions/plans/999",
		nil,
	)

	req = adminSubscriptionsRequestWithParam(
		req,
		"planID",
		"999",
	)

	rec := httptest.NewRecorder()

	handler.DisablePlan(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusNotFound,
			rec.Code,
		)
	}
}

func TestHandler_ListUserSubscriptions_Success(
	t *testing.T,
) {
	repo := &mockRepository{
		listUserSubscriptionsFn: func(
			context.Context,
		) ([]AdminUserSubscription, error) {
			return []AdminUserSubscription{
				{
					ID:     1,
					UserID: 8,
					Status: "active",
				},
			}, nil
		},
	}

	handler := NewHandler(
		NewService(repo),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/admin/subscriptions/users",
		nil,
	)

	rec := httptest.NewRecorder()

	handler.ListUserSubscriptions(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusOK,
			rec.Code,
		)
	}
}

func TestHandler_ListUserSubscriptions_InternalServerError(
	t *testing.T,
) {
	repo := &mockRepository{
		listUserSubscriptionsFn: func(
			context.Context,
		) ([]AdminUserSubscription, error) {
			return nil, errors.New("failed")
		},
	}

	handler := NewHandler(
		NewService(repo),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/admin/subscriptions/users",
		nil,
	)

	rec := httptest.NewRecorder()

	handler.ListUserSubscriptions(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusInternalServerError,
			rec.Code,
		)
	}
}

func TestHandler_GetUserSubscriptions_Success(
	t *testing.T,
) {
	repo := &mockRepository{
		getUserSubscriptionsFn: func(
			context.Context,
			uint,
		) (*AdminUserSubscription, error) {
			return &AdminUserSubscription{
				ID:     1,
				UserID: 8,
				Status: "active",
			}, nil
		},
	}

	handler := NewHandler(
		NewService(repo),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/admin/subscriptions/users/8",
		nil,
	)

	req = adminSubscriptionsRequestWithParam(
		req,
		"userID",
		"8",
	)

	rec := httptest.NewRecorder()

	handler.GetUserSubscriptions(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusOK,
			rec.Code,
		)
	}
}

func TestHandler_GetUserSubscriptions_InvalidID(
	t *testing.T,
) {
	handler := NewHandler(
		NewService(&mockRepository{}),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/admin/subscriptions/users/invalid",
		nil,
	)

	req = adminSubscriptionsRequestWithParam(
		req,
		"userID",
		"invalid",
	)

	rec := httptest.NewRecorder()

	handler.GetUserSubscriptions(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			rec.Code,
		)
	}
}

func TestHandler_GetUserSubscriptions_NotFound(
	t *testing.T,
) {
	repo := &mockRepository{
		getUserSubscriptionsFn: func(
			context.Context,
			uint,
		) (*AdminUserSubscription, error) {
			return nil, ErrSubscriptionNotFound
		},
	}

	handler := NewHandler(
		NewService(repo),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/admin/subscriptions/users/999",
		nil,
	)

	req = adminSubscriptionsRequestWithParam(
		req,
		"userID",
		"999",
	)

	rec := httptest.NewRecorder()

	handler.GetUserSubscriptions(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusNotFound,
			rec.Code,
		)
	}
}

func TestHandler_UpdateUserSubscription_Success(
	t *testing.T,
) {
	repo := &mockRepository{
		createOrUpdateUserSubscriptionFn: func(
			context.Context,
			uint,
			uint,
			string,
		) error {
			return nil
		},
	}

	handler := NewHandler(
		NewService(repo),
	)

	req := httptest.NewRequest(
		http.MethodPatch,
		"/admin/subscriptions/users/8",
		bytes.NewBufferString(`{
			"plan_id":2,
			"status":"active"
		}`),
	)

	req = adminSubscriptionsRequestWithParam(
		req,
		"userID",
		"8",
	)

	rec := httptest.NewRecorder()

	handler.UpdateUserSubscription(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusOK,
			rec.Code,
			rec.Body.String(),
		)
	}
}

func TestHandler_UpdateUserSubscription_InvalidStatus(
	t *testing.T,
) {
	handler := NewHandler(
		NewService(&mockRepository{}),
	)

	req := httptest.NewRequest(
		http.MethodPatch,
		"/admin/subscriptions/users/8",
		bytes.NewBufferString(`{
			"plan_id":2,
			"status":"pending"
		}`),
	)

	req = adminSubscriptionsRequestWithParam(
		req,
		"userID",
		"8",
	)

	rec := httptest.NewRecorder()

	handler.UpdateUserSubscription(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			rec.Code,
		)
	}
}
