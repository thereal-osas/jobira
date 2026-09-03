package adminsubscriptions

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
)

func adminSubscriptionsAuthMiddleware(
	next http.Handler,
) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
		},
	)
}

func adminSubscriptionsAdminMiddleware(
	next http.Handler,
) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
		},
	)
}

func TestRegisterRoutes_ListPlans(t *testing.T) {
	repo := &mockRepository{
		listPlansFn: func(
			context.Context,
		) ([]AdminSubscriptionPlan, error) {
			return []AdminSubscriptionPlan{}, nil
		},
	}

	router := chi.NewRouter()

	RegisterRoutes(
		router,
		NewHandler(NewService(repo)),
		adminSubscriptionsAuthMiddleware,
		adminSubscriptionsAdminMiddleware,
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/admin/subscriptions/plans",
		nil,
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusOK,
			rec.Code,
		)
	}
}

func TestRegisterRoutes_CreatePlan(t *testing.T) {
	router := chi.NewRouter()

	RegisterRoutes(
		router,
		NewHandler(NewService(&mockRepository{})),
		adminSubscriptionsAuthMiddleware,
		adminSubscriptionsAdminMiddleware,
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/admin/subscriptions/plans",
		strings.NewReader(`{
			"name":"Standard",
			"role_type":"cleaner",
			"billing_interval":"monthly",
			"cleaner_seat_limit":1
		}`),
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusCreated,
			rec.Code,
			rec.Body.String(),
		)
	}
}

func TestRegisterRoutes_UpdatePlan(t *testing.T) {
	router := chi.NewRouter()

	RegisterRoutes(
		router,
		NewHandler(NewService(&mockRepository{})),
		adminSubscriptionsAuthMiddleware,
		adminSubscriptionsAdminMiddleware,
	)

	req := httptest.NewRequest(
		http.MethodPut,
		"/admin/subscriptions/plans/2",
		strings.NewReader(`{
			"name":"Premium",
			"role_type":"cleaner",
			"billing_interval":"monthly",
			"cleaner_seat_limit":1,
			"is_active":true
		}`),
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusOK,
			rec.Code,
			rec.Body.String(),
		)
	}
}

func TestRegisterRoutes_DisablePlan(t *testing.T) {
	router := chi.NewRouter()

	RegisterRoutes(
		router,
		NewHandler(NewService(&mockRepository{})),
		adminSubscriptionsAuthMiddleware,
		adminSubscriptionsAdminMiddleware,
	)

	req := httptest.NewRequest(
		http.MethodDelete,
		"/admin/subscriptions/plans/2",
		nil,
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusOK,
			rec.Code,
		)
	}
}

func TestRegisterRoutes_ListUsers(t *testing.T) {
	repo := &mockRepository{
		listUserSubscriptionsFn: func(
			context.Context,
		) ([]AdminUserSubscription, error) {
			return []AdminUserSubscription{}, nil
		},
	}

	router := chi.NewRouter()

	RegisterRoutes(
		router,
		NewHandler(NewService(repo)),
		adminSubscriptionsAuthMiddleware,
		adminSubscriptionsAdminMiddleware,
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/admin/subscriptions/users",
		nil,
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusOK,
			rec.Code,
		)
	}
}

func TestRegisterRoutes_GetUser(t *testing.T) {
	repo := &mockRepository{
		getUserSubscriptionsFn: func(
			context.Context,
			uint,
		) (*AdminUserSubscription, error) {
			return &AdminUserSubscription{
				ID:     1,
				UserID: 8,
			}, nil
		},
	}

	router := chi.NewRouter()

	RegisterRoutes(
		router,
		NewHandler(NewService(repo)),
		adminSubscriptionsAuthMiddleware,
		adminSubscriptionsAdminMiddleware,
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/admin/subscriptions/users/8",
		nil,
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusOK,
			rec.Code,
		)
	}
}

func TestRegisterRoutes_UpdateUser(t *testing.T) {
	router := chi.NewRouter()

	RegisterRoutes(
		router,
		NewHandler(NewService(&mockRepository{})),
		adminSubscriptionsAuthMiddleware,
		adminSubscriptionsAdminMiddleware,
	)

	req := httptest.NewRequest(
		http.MethodPatch,
		"/admin/subscriptions/users/8",
		strings.NewReader(`{
			"plan_id":2,
			"status":"active"
		}`),
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusOK,
			rec.Code,
			rec.Body.String(),
		)
	}
}

func TestRegisterRoutes_MethodNotAllowed(
	t *testing.T,
) {
	router := chi.NewRouter()

	RegisterRoutes(
		router,
		NewHandler(NewService(&mockRepository{})),
		adminSubscriptionsAuthMiddleware,
		adminSubscriptionsAdminMiddleware,
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/admin/subscriptions/users",
		nil,
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusMethodNotAllowed,
			rec.Code,
		)
	}
}

func TestRegisterRoutes_AuthCanBlock(t *testing.T) {
	authMiddleware := func(
		next http.Handler,
	) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				http.Error(
					w,
					"unauthorized",
					http.StatusUnauthorized,
				)
			},
		)
	}

	adminCalled := false

	adminMiddleware := func(
		next http.Handler,
	) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				adminCalled = true
				next.ServeHTTP(w, r)
			},
		)
	}

	router := chi.NewRouter()

	RegisterRoutes(
		router,
		NewHandler(NewService(&mockRepository{})),
		authMiddleware,
		adminMiddleware,
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/admin/subscriptions/plans",
		nil,
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusUnauthorized,
			rec.Code,
		)
	}

	if adminCalled {
		t.Fatal(
			"admin middleware should not run when auth blocks",
		)
	}
}

func TestRegisterRoutes_AdminCanBlock(t *testing.T) {
	authCalled := false

	authMiddleware := func(
		next http.Handler,
	) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				authCalled = true
				next.ServeHTTP(w, r)
			},
		)
	}

	adminMiddleware := func(
		next http.Handler,
	) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				http.Error(
					w,
					"forbidden",
					http.StatusForbidden,
				)
			},
		)
	}

	router := chi.NewRouter()

	RegisterRoutes(
		router,
		NewHandler(NewService(&mockRepository{})),
		authMiddleware,
		adminMiddleware,
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/admin/subscriptions/plans",
		nil,
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusForbidden,
			rec.Code,
		)
	}

	if !authCalled {
		t.Fatal(
			"expected auth middleware to run first",
		)
	}
}
