package billing

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

func billingTestAuth(
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

func newBillingTestRouter(
	repo *MockRepository,
	userID uint,
	role string,
	cfg StripeConfig,
) *chi.Mux {
	service := NewService(
		repo,
		cfg,
	)

	handler := NewHandler(service)

	router := chi.NewRouter()

	RegisterRoutes(
		router,
		handler,
		billingTestAuth(
			userID,
			role,
		),
	)

	return router
}

func TestHandler_NewHandler(t *testing.T) {
	service := NewService(
		&MockRepository{},
		StripeConfig{},
	)

	handler := NewHandler(service)

	if handler == nil {
		t.Fatal("expected handler")
	}

	if handler.service != service {
		t.Fatal(
			"expected service assigned",
		)
	}
}

func TestHandler_CreateCheckoutSession_Unauthorized(
	t *testing.T,
) {
	handler := NewHandler(
		NewService(
			&MockRepository{},
			StripeConfig{},
		),
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/billing/checkout",
		bytes.NewBufferString(`{}`),
	)

	res := httptest.NewRecorder()

	handler.CreateCheckoutSession(
		res,
		req,
	)

	if res.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected 401 got %d",
			res.Code,
		)
	}
}

func TestHandler_CreateCheckoutSession_InvalidBody(
	t *testing.T,
) {
	router := newBillingTestRouter(
		&MockRepository{},
		1,
		"cleaner",
		StripeConfig{},
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/billing/checkout",
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

func TestHandler_CreateCheckoutSession_InvalidInput(
	t *testing.T,
) {
	router := newBillingTestRouter(
		&MockRepository{},
		1,
		"cleaner",
		StripeConfig{},
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/billing/checkout",
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

func TestHandler_CreateCheckoutSession_PlanNotFound(
	t *testing.T,
) {
	router := newBillingTestRouter(
		&MockRepository{
			planErr: ErrPlanNotFound,
		},
		1,
		"cleaner",
		checkoutConfig(),
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/billing/checkout",
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

func TestHandler_CreateCheckoutSession_MissingStripePrice(
	t *testing.T,
) {
	router := newBillingTestRouter(
		&MockRepository{
			plan: &BillingPlan{
				ID:            1,
				StripePriceID: "",
			},
		},
		1,
		"cleaner",
		checkoutConfig(),
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/billing/checkout",
		bytes.NewBufferString(`{
			"plan_id":1
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

func TestHandler_CreateCheckoutSession_ConfigError(
	t *testing.T,
) {
	router := newBillingTestRouter(
		&MockRepository{},
		1,
		"cleaner",
		StripeConfig{},
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/billing/checkout",
		bytes.NewBufferString(`{
			"plan_id":1
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

func TestHandler_CreateCheckoutSession_ServiceError(
	t *testing.T,
) {
	router := newBillingTestRouter(
		&MockRepository{
			planErr: errors.New(
				"database failed",
			),
		},
		1,
		"cleaner",
		checkoutConfig(),
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/billing/checkout",
		bytes.NewBufferString(`{
			"plan_id":1
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

func TestHandler_CreateBillingPortalSession_Unauthorized(
	t *testing.T,
) {
	handler := NewHandler(
		NewService(
			&MockRepository{},
			StripeConfig{},
		),
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/billing/portal",
		nil,
	)

	res := httptest.NewRecorder()

	handler.CreateBillingPortalSession(
		res,
		req,
	)

	if res.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected 401 got %d",
			res.Code,
		)
	}
}

func TestHandler_CreateBillingPortalSession_InvalidInput(
	t *testing.T,
) {
	router := newBillingTestRouter(
		&MockRepository{},
		0,
		"cleaner",
		StripeConfig{},
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/billing/portal",
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

func TestHandler_CreateBillingPortalSession_CustomerNotFound(
	t *testing.T,
) {
	router := newBillingTestRouter(
		&MockRepository{
			customerErr: ErrCustomerNotFound,
		},
		1,
		"cleaner",
		StripeConfig{
			SecretKey:       "sk_test_fake",
			PortalReturnURL: "https://example.com/account",
		},
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/billing/portal",
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

func TestHandler_CreateBillingPortalSession_ConfigError(
	t *testing.T,
) {
	router := newBillingTestRouter(
		&MockRepository{},
		1,
		"cleaner",
		StripeConfig{},
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/billing/portal",
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

func TestHandler_CreateBillingPortalSession_ServiceError(
	t *testing.T,
) {
	router := newBillingTestRouter(
		&MockRepository{
			customerErr: errors.New(
				"database failed",
			),
		},
		1,
		"cleaner",
		StripeConfig{
			SecretKey:       "sk_test_fake",
			PortalReturnURL: "https://example.com/account",
		},
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/billing/portal",
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

func TestHandler_StripeWebhook_InvalidWebhook(
	t *testing.T,
) {
	router := newBillingTestRouter(
		&MockRepository{},
		0,
		"",
		StripeConfig{
			WebhookSecret: "whsec_fake",
		},
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/stripe/webhook",
		bytes.NewBufferString(`{
			"type":"checkout.session.completed"
		}`),
	)

	req.Header.Set(
		"Stripe-Signature",
		"invalid",
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

func TestHandler_StripeWebhook_MissingSecret(
	t *testing.T,
) {
	router := newBillingTestRouter(
		&MockRepository{},
		0,
		"",
		StripeConfig{},
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/stripe/webhook",
		bytes.NewBufferString(`{}`),
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

func TestHandler_ValidatePromoCode_Success(
	t *testing.T,
) {
	router := newBillingTestRouter(
		&MockRepository{
			promo: &PromoCode{
				Code:          "SAVE20",
				Description:   "20 percent off",
				DiscountType:  "percentage",
				PercentageOff: 20,
				IsActive:      true,
			},
		},
		1,
		"cleaner",
		StripeConfig{},
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/billing/promo/validate",
		bytes.NewBufferString(`{
			"code":"SAVE20",
			"plan_id":2
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

	var promo ValidatePromoCodeRespone

	if err := json.NewDecoder(
		res.Body,
	).Decode(&promo); err != nil {
		t.Fatalf(
			"failed decoding response: %v",
			err,
		)
	}

	if !promo.Valid {
		t.Fatal(
			"expected valid promo",
		)
	}

	if promo.Code != "SAVE20" {
		t.Fatalf(
			"expected SAVE20 got %q",
			promo.Code,
		)
	}
}

func TestHandler_ValidatePromoCode_Unauthorized(
	t *testing.T,
) {
	handler := NewHandler(
		NewService(
			&MockRepository{},
			StripeConfig{},
		),
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/billing/promo/validate",
		bytes.NewBufferString(`{}`),
	)

	res := httptest.NewRecorder()

	handler.ValidatePromoCode(
		res,
		req,
	)

	if res.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected 401 got %d",
			res.Code,
		)
	}
}

func TestHandler_ValidatePromoCode_InvalidBody(
	t *testing.T,
) {
	router := newBillingTestRouter(
		&MockRepository{},
		1,
		"cleaner",
		StripeConfig{},
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/billing/promo/validate",
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

func TestHandler_ValidatePromoCode_InvalidInput(
	t *testing.T,
) {
	router := newBillingTestRouter(
		&MockRepository{},
		1,
		"cleaner",
		StripeConfig{},
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/billing/promo/validate",
		bytes.NewBufferString(`{
			"code":""
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

func TestHandler_ValidatePromoCode_NotFound(
	t *testing.T,
) {
	router := newBillingTestRouter(
		&MockRepository{
			promoErr: ErrPromoCodeNotFound,
		},
		1,
		"cleaner",
		StripeConfig{},
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/billing/promo/validate",
		bytes.NewBufferString(`{
			"code":"NOPE"
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

func TestHandler_ValidatePromoCode_InvalidPromo(
	t *testing.T,
) {
	router := newBillingTestRouter(
		&MockRepository{
			promo: &PromoCode{
				Code:     "OLD",
				IsActive: false,
			},
		},
		1,
		"cleaner",
		StripeConfig{},
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/billing/promo/validate",
		bytes.NewBufferString(`{
			"code":"OLD"
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

func TestHandler_ValidatePromoCode_ServiceError(
	t *testing.T,
) {
	router := newBillingTestRouter(
		&MockRepository{
			promoErr: errors.New(
				"database failed",
			),
		},
		1,
		"cleaner",
		StripeConfig{},
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/billing/promo/validate",
		bytes.NewBufferString(`{
			"code":"SAVE20"
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
