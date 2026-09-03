package subscriptions

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/rodrigueghenda/jobira/internal/security/identity"
)

func subscriptionsTestAuth(
	userID uint,
	role string,
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				ctx := identity.WithUser(
					r.Context(),
					identity.UserIdentity{
						UserID: userID,
						Email:  "test@example.com",
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

func newSubscriptionsTestRouter(
	repo *MockRepository,
	userID uint,
	role string,
) *chi.Mux {
	service := NewService(repo)
	handler := NewHandler(service)

	router := chi.NewRouter()

	RegisterRoutes(
		router,
		handler,
		subscriptionsTestAuth(userID, role),
	)

	return router
}

func TestHandler_NewHandler(t *testing.T) {
	service := NewService(
		&MockRepository{},
	)

	handler := NewHandler(service)

	if handler == nil {
		t.Fatal("expected handler")
	}

	if handler.service != service {
		t.Fatal("expected service assigned")
	}
}

func TestHandler_ListPlans_Success(t *testing.T) {
	repo := &MockRepository{
		plans: []SubscriptionPlan{
			{
				ID:   1,
				Name: "Launch",
			},
			{
				ID:   2,
				Name: "Standard",
			},
		},
	}

	router := newSubscriptionsTestRouter(
		repo,
		0,
		"",
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/subscriptions/plans",
		nil,
	)

	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf(
			"expected 200 got %d: %s",
			res.Code,
			res.Body.String(),
		)
	}

	var plans []SubscriptionPlan

	if err := json.NewDecoder(
		res.Body,
	).Decode(&plans); err != nil {
		t.Fatalf(
			"failed decoding plans: %v",
			err,
		)
	}

	if len(plans) != 2 {
		t.Fatalf(
			"expected 2 plans got %d",
			len(plans),
		)
	}
}

func TestHandler_ListPlans_ServiceError(t *testing.T) {
	router := newSubscriptionsTestRouter(
		&MockRepository{
			listPlansErr: errors.New("database failed"),
		},
		0,
		"",
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/subscriptions/plans",
		nil,
	)

	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected 500 got %d",
			res.Code,
		)
	}
}

func TestHandler_GetMine_Success(t *testing.T) {
	repo := &MockRepository{
		subscription: &UserSubscription{
			ID:     5,
			UserID: 10,
			Status: "active",
		},
	}

	router := newSubscriptionsTestRouter(
		repo,
		10,
		"cleaner",
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/subscriptions/me",
		nil,
	)

	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf(
			"expected 200 got %d: %s",
			res.Code,
			res.Body.String(),
		)
	}

	var subscription UserSubscription

	if err := json.NewDecoder(
		res.Body,
	).Decode(&subscription); err != nil {
		t.Fatalf(
			"failed decoding response: %v",
			err,
		)
	}

	if subscription.ID != 5 {
		t.Fatalf(
			"expected ID 5 got %d",
			subscription.ID,
		)
	}
}

func TestHandler_GetMine_Unauthorized(t *testing.T) {
	handler := NewHandler(
		NewService(
			&MockRepository{},
		),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/subscriptions/me",
		nil,
	)

	res := httptest.NewRecorder()

	handler.GetMine(res, req)

	if res.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected 401 got %d",
			res.Code,
		)
	}
}

func TestHandler_GetMine_InvalidInput(t *testing.T) {
	router := newSubscriptionsTestRouter(
		&MockRepository{},
		0,
		"cleaner",
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/subscriptions/me",
		nil,
	)

	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected 400 got %d",
			res.Code,
		)
	}
}

func TestHandler_GetMine_NotFound(t *testing.T) {
	router := newSubscriptionsTestRouter(
		&MockRepository{
			getByUserErr: ErrSubscriptionNotFound,
		},
		10,
		"cleaner",
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/subscriptions/me",
		nil,
	)

	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusNotFound {
		t.Fatalf(
			"expected 404 got %d",
			res.Code,
		)
	}
}

func TestHandler_GetMine_ServiceError(t *testing.T) {
	router := newSubscriptionsTestRouter(
		&MockRepository{
			getByUserErr: errors.New("database failed"),
		},
		10,
		"cleaner",
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/subscriptions/me",
		nil,
	)

	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected 500 got %d",
			res.Code,
		)
	}
}

func TestHandler_CreateOrUpdateMine_Success(t *testing.T) {
	planID := uint(2)

	repo := &MockRepository{
		plan: &SubscriptionPlan{
			ID:   planID,
			Name: "Standard",
		},
		subscription: &UserSubscription{
			ID:     5,
			UserID: 10,
			PlanID: &planID,
			Status: "trial",
		},
	}

	router := newSubscriptionsTestRouter(
		repo,
		10,
		"cleaner",
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/subscriptions/me",
		bytes.NewBufferString(`{
			"plan_id":2
		}`),
	)

	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusCreated {
		t.Fatalf(
			"expected 201 got %d: %s",
			res.Code,
			res.Body.String(),
		)
	}

	if repo.lastCreateUserID != 10 {
		t.Fatalf(
			"expected user 10 got %d",
			repo.lastCreateUserID,
		)
	}

	if repo.lastCreatePlanID != 2 {
		t.Fatalf(
			"expected plan 2 got %d",
			repo.lastCreatePlanID,
		)
	}

	if repo.lastCreateStatus != "trial" {
		t.Fatalf(
			"expected trial got %q",
			repo.lastCreateStatus,
		)
	}
}

func TestHandler_CreateOrUpdateMine_Unauthorized(
	t *testing.T,
) {
	handler := NewHandler(
		NewService(
			&MockRepository{},
		),
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/subscriptions/me",
		bytes.NewBufferString(`{}`),
	)

	res := httptest.NewRecorder()

	handler.CreateOrUpdateMine(res, req)

	if res.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected 401 got %d",
			res.Code,
		)
	}
}

func TestHandler_CreateOrUpdateMine_InvalidBody(
	t *testing.T,
) {
	router := newSubscriptionsTestRouter(
		&MockRepository{},
		10,
		"cleaner",
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/subscriptions/me",
		bytes.NewBufferString("{bad"),
	)

	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected 400 got %d",
			res.Code,
		)
	}
}

func TestHandler_CreateOrUpdateMine_InvalidInput(
	t *testing.T,
) {
	router := newSubscriptionsTestRouter(
		&MockRepository{},
		10,
		"cleaner",
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/subscriptions/me",
		bytes.NewBufferString(`{
			"plan_id":0
		}`),
	)

	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected 400 got %d",
			res.Code,
		)
	}
}

func TestHandler_CreateOrUpdateMine_PlanNotFound(
	t *testing.T,
) {
	router := newSubscriptionsTestRouter(
		&MockRepository{
			getPlanErr: ErrPlanNotFound,
		},
		10,
		"cleaner",
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/subscriptions/me",
		bytes.NewBufferString(`{
			"plan_id":999
		}`),
	)

	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusNotFound {
		t.Fatalf(
			"expected 404 got %d",
			res.Code,
		)
	}
}

func TestHandler_CreateOrUpdateMine_ServiceError(
	t *testing.T,
) {
	router := newSubscriptionsTestRouter(
		&MockRepository{
			getPlanErr: errors.New("database failed"),
		},
		10,
		"cleaner",
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/subscriptions/me",
		bytes.NewBufferString(`{
			"plan_id":2
		}`),
	)

	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected 500 got %d",
			res.Code,
		)
	}
}

func TestHandler_UpdateMineStatus_Success(t *testing.T) {
	repo := &MockRepository{}

	router := newSubscriptionsTestRouter(
		repo,
		10,
		"cleaner",
	)

	req := httptest.NewRequest(
		http.MethodPatch,
		"/subscriptions/me/status",
		bytes.NewBufferString(`{
			"status":"active"
		}`),
	)

	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf(
			"expected 200 got %d: %s",
			res.Code,
			res.Body.String(),
		)
	}

	if repo.lastUpdateUserID != 10 {
		t.Fatalf(
			"expected user 10 got %d",
			repo.lastUpdateUserID,
		)
	}

	if repo.lastUpdateStatus != "active" {
		t.Fatalf(
			"expected active got %q",
			repo.lastUpdateStatus,
		)
	}
}

func TestHandler_UpdateMineStatus_Unauthorized(
	t *testing.T,
) {
	handler := NewHandler(
		NewService(
			&MockRepository{},
		),
	)

	req := httptest.NewRequest(
		http.MethodPatch,
		"/subscriptions/me/status",
		bytes.NewBufferString(`{}`),
	)

	res := httptest.NewRecorder()

	handler.UpdateMineStatus(res, req)

	if res.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected 401 got %d",
			res.Code,
		)
	}
}

func TestHandler_UpdateMineStatus_InvalidBody(
	t *testing.T,
) {
	router := newSubscriptionsTestRouter(
		&MockRepository{},
		10,
		"cleaner",
	)

	req := httptest.NewRequest(
		http.MethodPatch,
		"/subscriptions/me/status",
		bytes.NewBufferString("{bad"),
	)

	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected 400 got %d",
			res.Code,
		)
	}
}

func TestHandler_UpdateMineStatus_InvalidInput(
	t *testing.T,
) {
	router := newSubscriptionsTestRouter(
		&MockRepository{},
		10,
		"cleaner",
	)

	req := httptest.NewRequest(
		http.MethodPatch,
		"/subscriptions/me/status",
		bytes.NewBufferString(`{
			"status":""
		}`),
	)

	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected 400 got %d",
			res.Code,
		)
	}
}

func TestHandler_UpdateMineStatus_InvalidStatus(
	t *testing.T,
) {
	router := newSubscriptionsTestRouter(
		&MockRepository{},
		10,
		"cleaner",
	)

	req := httptest.NewRequest(
		http.MethodPatch,
		"/subscriptions/me/status",
		bytes.NewBufferString(`{
			"status":"banana"
		}`),
	)

	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected 400 got %d",
			res.Code,
		)
	}
}

func TestHandler_UpdateMineStatus_NotFound(
	t *testing.T,
) {
	router := newSubscriptionsTestRouter(
		&MockRepository{
			updateStatusErr: ErrSubscriptionNotFound,
		},
		10,
		"cleaner",
	)

	req := httptest.NewRequest(
		http.MethodPatch,
		"/subscriptions/me/status",
		bytes.NewBufferString(`{
			"status":"active"
		}`),
	)

	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusNotFound {
		t.Fatalf(
			"expected 404 got %d",
			res.Code,
		)
	}
}

func TestHandler_UpdateMineStatus_ServiceError(
	t *testing.T,
) {
	router := newSubscriptionsTestRouter(
		&MockRepository{
			updateStatusErr: errors.New("database failed"),
		},
		10,
		"cleaner",
	)

	req := httptest.NewRequest(
		http.MethodPatch,
		"/subscriptions/me/status",
		bytes.NewBufferString(`{
			"status":"active"
		}`),
	)

	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected 500 got %d",
			res.Code,
		)
	}
}
