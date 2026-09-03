package billing

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stripe/stripe-go/v84"
	"github.com/stripe/stripe-go/v84/webhook"
)

type MockRepository struct {
	plan            *BillingPlan
	email           string
	customerID      string
	promo           *PromoCode
	planErr         error
	emailErr        error
	customerErr     error
	upsertErr       error
	saveCheckoutErr error
	markCompleteErr error
	activateErr     error
	updateStatusErr error
	promoErr        error
	incrementErr    error

	markCompleteCalled bool
	markCompleteResult bool
	markedSessionID    string
	markedCustomerID   string
	markedSubID        string

	activateCalled  bool
	activatedUserID uint
	activatedPlanID uint
	activatedSubID  string
	activatedStart  time.Time
	activatedEnd    time.Time

	updateStatusCalled bool
	updatedCustomerID  string
	updatedStatus      string

	updateRenewalCalled   bool
	renewalCustomerID     string
	renewalSubscriptionID string
	renewalPeriodStart    time.Time
	renewalPeriodEnd      time.Time
	updateRenewalErr      error

	markExpiredCalled    bool
	markedExpiredSession string
	markExpiredErr       error
}

type MockStripeSubscriptionRetriever struct {
	subscription *stripe.Subscription
	err          error
	called       bool
}

func (m *MockRepository) MarkCheckoutSessionExpired(
	ctx context.Context,
	stripeSessionID string,
) error {
	m.markExpiredCalled = true
	m.markedExpiredSession = stripeSessionID

	return m.markExpiredErr
}

func (m *MockStripeSubscriptionRetriever) Get(
	subscriptionID string,
) (*stripe.Subscription, error) {
	m.called = true

	if m.err != nil {
		return nil, m.err
	}

	return m.subscription, nil
}

func (m *MockRepository) GetPlanByID(
	ctx context.Context,
	planID uint,
) (*BillingPlan, error) {
	if m.planErr != nil {
		return nil, m.planErr
	}
	return m.plan, nil
}

func (m *MockRepository) GetUserEmail(
	ctx context.Context,
	userID uint,
) (string, error) {
	if m.emailErr != nil {
		return "", m.emailErr
	}
	return m.email, nil
}

func (m *MockRepository) GetStripeCustomerID(
	ctx context.Context,
	userID uint,
) (string, error) {
	if m.customerErr != nil {
		return "", m.customerErr
	}
	return m.customerID, nil
}

func (m *MockRepository) UpsertBillingCustomer(
	ctx context.Context,
	userID uint,
	email string,
	stripeCustomerID string,
) error {
	return m.upsertErr
}

func (m *MockRepository) SaveCheckoutSession(
	ctx context.Context,
	record *CheckoutSessionRecord,
) error {
	return m.saveCheckoutErr
}

func (m *MockRepository) MarkCheckoutSessionComplete(
	ctx context.Context,
	stripeSessionID string,
	stripeCustomerID string,
	stripeSubscriptionID string,
) (bool, error) {
	m.markCompleteCalled = true
	m.markedSessionID = stripeSessionID
	m.markedCustomerID = stripeCustomerID
	m.markedSubID = stripeSubscriptionID

	if m.markCompleteErr != nil {
		return false, m.markCompleteErr
	}

	return m.markCompleteResult, nil
}

func (m *MockRepository) ActivateUserSubscription(
	ctx context.Context,
	userID uint,
	planID uint,
	stripeSubscriptionID string,
	periodStart time.Time,
	periodEnd time.Time,
) error {
	m.activateCalled = true
	m.activatedUserID = userID
	m.activatedPlanID = planID
	m.activatedSubID = stripeSubscriptionID
	m.activatedStart = periodStart
	m.activatedEnd = periodEnd

	return m.activateErr
}

func (m *MockRepository) UpdateUserSubscriptionStatusByCustomer(
	ctx context.Context,
	stripeCustomerID string,
	status string,
) error {
	m.updateStatusCalled = true
	m.updatedCustomerID = stripeCustomerID
	m.updatedStatus = status

	return m.updateStatusErr
}

func (m *MockRepository) UpdateUserSubscriptionRenewalByCustomer(
	ctx context.Context,
	stripeCustomerID string,
	stripeSubscriptionID string,
	periodStart time.Time,
	periodEnd time.Time,
) error {
	m.updateRenewalCalled = true
	m.renewalCustomerID = stripeCustomerID
	m.renewalSubscriptionID = stripeSubscriptionID
	m.renewalPeriodStart = periodStart
	m.renewalPeriodEnd = periodEnd

	return m.updateRenewalErr
}

func (m *MockRepository) GetPromoCode(
	ctx context.Context,
	code string,
) (*PromoCode, error) {
	if m.promoErr != nil {
		return nil, m.promoErr
	}
	return m.promo, nil
}

func (m *MockRepository) IncrementPromoUse(
	ctx context.Context,
	code string,
) error {
	return m.incrementErr
}

func TestNewService(t *testing.T) {
	repo := &MockRepository{}

	service := NewService(
		repo,
		StripeConfig{},
	)

	if service == nil {
		t.Fatal("expected service")
	}

	if service.repo != repo {
		t.Fatal("expected repository assigned")
	}
}

func TestCreateCheckoutSession_InvalidInput(t *testing.T) {
	service := NewService(
		&MockRepository{},
		StripeConfig{},
	)

	tests := []struct {
		name   string
		userID uint
		planID uint
	}{
		{
			name:   "zero user",
			userID: 0,
			planID: 1,
		},
		{
			name:   "zero plan",
			userID: 1,
			planID: 0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := service.CreateCheckoutSession(
				context.Background(),
				test.userID,
				CreateCheckoutSessionRequest{
					PlanID: test.planID,
				},
			)

			if result != nil {
				t.Fatal("expected nil response")
			}

			if !errors.Is(err, ErrInvalidInput) {
				t.Fatalf(
					"expected ErrInvalidInput got %v",
					err,
				)
			}
		})
	}
}

func TestCreateCheckoutSession_MissingConfig(t *testing.T) {
	service := NewService(
		&MockRepository{},
		StripeConfig{},
	)

	_, err := service.CreateCheckoutSession(
		context.Background(),
		1,
		CreateCheckoutSessionRequest{
			PlanID: 1,
		},
	)

	if !errors.Is(
		err,
		ErrCheckoutSessionError,
	) {
		t.Fatalf(
			"expected checkout config error got %v",
			err,
		)
	}
}

func checkoutConfig() StripeConfig {
	return StripeConfig{
		SecretKey:  "sk_test_fake",
		SuccessURL: "https://example.com/success",
		CancelURL:  "https://example.com/cancel",
	}
}

func testStripeSubscription(
	periodStart time.Time,
	periodEnd time.Time,
) *stripe.Subscription {
	return &stripe.Subscription{
		ID: "sub_123",
		Items: &stripe.SubscriptionItemList{
			Data: []*stripe.SubscriptionItem{
				{
					ID:                 "si_123",
					CurrentPeriodStart: periodStart.Unix(),
					CurrentPeriodEnd:   periodEnd.Unix(),
				},
			},
		},
	}
}

func signedStripeWebhook(
	payload []byte,
	secret string,
) *webhook.SignedPayload {
	return webhook.GenerateTestSignedPayload(
		&webhook.UnsignedPayload{
			Payload: payload,
			Secret:  secret,
		},
	)
}

func stripeEventPayload(
	eventType string,
	objectJSON string,
) []byte {
	return []byte(
		`{
			"id":"evt_test",
			"object":"event",
			"api_version":"` + stripe.APIVersion + `",
			"type":"` + eventType + `",
			"data":{
				"object":` + objectJSON + `
			}
		}`,
	)
}

func TestCreateCheckoutSession_PlanNotFound(t *testing.T) {
	service := NewService(
		&MockRepository{
			planErr: ErrPlanNotFound,
		},
		checkoutConfig(),
	)

	_, err := service.CreateCheckoutSession(
		context.Background(),
		1,
		CreateCheckoutSessionRequest{
			PlanID: 99,
		},
	)

	if !errors.Is(err, ErrPlanNotFound) {
		t.Fatalf(
			"expected ErrPlanNotFound got %v",
			err,
		)
	}
}

func TestCreateCheckoutSession_PlanLookupError(
	t *testing.T,
) {
	lookupErr := errors.New("plan lookup failed")

	service := NewService(
		&MockRepository{
			planErr: lookupErr,
		},
		checkoutConfig(),
	)

	_, err := service.CreateCheckoutSession(
		context.Background(),
		1,
		CreateCheckoutSessionRequest{
			PlanID: 1,
		},
	)

	if !errors.Is(err, lookupErr) {
		t.Fatalf(
			"expected lookup error got %v",
			err,
		)
	}
}

func TestCreateCheckoutSession_MissingStripePrice(
	t *testing.T,
) {
	service := NewService(
		&MockRepository{
			plan: &BillingPlan{
				ID:            1,
				StripePriceID: "   ",
			},
		},
		checkoutConfig(),
	)

	_, err := service.CreateCheckoutSession(
		context.Background(),
		1,
		CreateCheckoutSessionRequest{
			PlanID: 1,
		},
	)

	if !errors.Is(err, ErrStripePriceMissing) {
		t.Fatalf(
			"expected ErrStripePriceMissing got %v",
			err,
		)
	}
}

func TestCreateCheckoutSession_EmailLookupError(
	t *testing.T,
) {
	emailErr := errors.New("email lookup failed")

	service := NewService(
		&MockRepository{
			plan: &BillingPlan{
				ID:            1,
				StripePriceID: "price_test",
			},
			emailErr: emailErr,
		},
		checkoutConfig(),
	)

	_, err := service.CreateCheckoutSession(
		context.Background(),
		1,
		CreateCheckoutSessionRequest{
			PlanID: 1,
		},
	)

	if !errors.Is(err, emailErr) {
		t.Fatalf(
			"expected email error got %v",
			err,
		)
	}
}

func TestCreateCheckoutSession_CustomerLookupError(
	t *testing.T,
) {
	customerErr := errors.New(
		"customer lookup failed",
	)

	service := NewService(
		&MockRepository{
			plan: &BillingPlan{
				ID:            1,
				StripePriceID: "price_test",
			},
			email:       "test@example.com",
			customerErr: customerErr,
		},
		checkoutConfig(),
	)

	_, err := service.CreateCheckoutSession(
		context.Background(),
		1,
		CreateCheckoutSessionRequest{
			PlanID: 1,
		},
	)

	if !errors.Is(err, customerErr) {
		t.Fatalf(
			"expected customer error got %v",
			err,
		)
	}
}

func TestCreateBillingPortalSession_InvalidUser(
	t *testing.T,
) {
	service := NewService(
		&MockRepository{},
		StripeConfig{},
	)

	result, err := service.CreateBillingPortalSession(
		context.Background(),
		0,
	)

	if result != nil {
		t.Fatal("expected nil response")
	}

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf(
			"expected ErrInvalidInput got %v",
			err,
		)
	}
}

func TestCreateBillingPortalSession_MissingConfig(
	t *testing.T,
) {
	service := NewService(
		&MockRepository{},
		StripeConfig{},
	)

	_, err := service.CreateBillingPortalSession(
		context.Background(),
		1,
	)

	if !errors.Is(err, ErrPortalSessionError) {
		t.Fatalf(
			"expected ErrPortalSessionError got %v",
			err,
		)
	}
}

func TestCreateBillingPortalSession_CustomerNotFound(
	t *testing.T,
) {
	service := NewService(
		&MockRepository{
			customerErr: ErrCustomerNotFound,
		},
		StripeConfig{
			SecretKey:       "sk_test_fake",
			PortalReturnURL: "https://example.com/account",
		},
	)

	_, err := service.CreateBillingPortalSession(
		context.Background(),
		1,
	)

	if !errors.Is(err, ErrCustomerNotFound) {
		t.Fatalf(
			"expected ErrCustomerNotFound got %v",
			err,
		)
	}
}

func TestCreateBillingPortalSession_CustomerLookupError(
	t *testing.T,
) {
	lookupErr := errors.New("lookup failed")

	service := NewService(
		&MockRepository{
			customerErr: lookupErr,
		},
		StripeConfig{
			SecretKey:       "sk_test_fake",
			PortalReturnURL: "https://example.com/account",
		},
	)

	_, err := service.CreateBillingPortalSession(
		context.Background(),
		1,
	)

	if !errors.Is(err, lookupErr) {
		t.Fatalf(
			"expected lookup error got %v",
			err,
		)
	}
}

func TestHandleWebhook_MissingSecret(t *testing.T) {
	service := NewService(
		&MockRepository{},
		StripeConfig{},
	)

	err := service.HandleWebhook(
		context.Background(),
		[]byte(`{}`),
		"",
	)

	if !errors.Is(err, ErrWebhookInvalid) {
		t.Fatalf(
			"expected ErrWebhookInvalid got %v",
			err,
		)
	}
}

func TestHandleWebhook_InvalidSignature(t *testing.T) {
	service := NewService(
		&MockRepository{},
		StripeConfig{
			WebhookSecret: "whsec_fake",
		},
	)

	err := service.HandleWebhook(
		context.Background(),
		[]byte(`{"type":"test"}`),
		"invalid-signature",
	)

	if !errors.Is(err, ErrWebhookInvalid) {
		t.Fatalf(
			"expected ErrWebhookInvalid got %v",
			err,
		)
	}
}

func TestHandleWebhook_CheckoutCompleted_Success(
	t *testing.T,
) {

	periodStart := time.Date(
		2026,
		time.September,
		2,
		12,
		0,
		0,
		0,
		time.UTC,
	)

	periodEnd := time.Date(
		2026,
		time.October,
		2,
		12,
		0,
		0,
		0,
		time.UTC,
	)

	const secret = "whsec_test_secret"

	repo := &MockRepository{
		markCompleteResult: true,
	}

	service := NewService(
		repo,
		StripeConfig{
			WebhookSecret: secret,
		},
	)

	service.subscriptionRetriever = &MockStripeSubscriptionRetriever{
		subscription: testStripeSubscription(
			periodStart,
			periodEnd,
		),
	}

	payload := stripeEventPayload(
		"checkout.session.completed",
		`{
		"id":"cs_123",
		"object":"checkout.session",
		"customer":"cus_123",
		"subscription":"sub_123",
		"metadata":{
			"user_id":"5",
			"plan_id":"2"
		}
	}`,
	)

	err := service.HandleWebhook(
		context.Background(),
		payload,
		signedStripeWebhook(payload, secret).Header,
	)
	if err != nil {
		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}

	if !repo.markCompleteCalled {
		t.Fatal(
			"expected checkout session to be marked complete",
		)
	}

	if repo.markedSessionID != "cs_123" {
		t.Fatalf(
			"expected session cs_123 got %q",
			repo.markedSessionID,
		)
	}

	if repo.markedCustomerID != "cus_123" {
		t.Fatalf(
			"expected customer cus_123 got %q",
			repo.markedCustomerID,
		)
	}

	if repo.markedSubID != "sub_123" {
		t.Fatalf(
			"expected subscription sub_123 got %q",
			repo.markedSubID,
		)
	}

	if !repo.activateCalled {
		t.Fatal(
			"expected subscription activation",
		)
	}

	if repo.activatedUserID != 5 {
		t.Fatalf(
			"expected user 5 got %d",
			repo.activatedUserID,
		)
	}

	if repo.activatedPlanID != 2 {
		t.Fatalf(
			"expected plan 2 got %d",
			repo.activatedPlanID,
		)
	}

	if repo.activatedSubID != "sub_123" {
		t.Fatalf(
			"expected subscription sub_123 got %q",
			repo.activatedSubID,
		)
	}

	if !repo.activatedStart.Equal(periodStart) {
		t.Fatalf(
			"expected period start %v got %v",
			periodStart,
			repo.activatedStart,
		)
	}

	if !repo.activatedEnd.Equal(periodEnd) {
		t.Fatalf(
			"expected period end %v got %v",
			periodEnd,
			repo.activatedEnd,
		)
	}
}

func TestHandleWebhook_CheckoutCompleted_MissingMetadata(
	t *testing.T,
) {
	const secret = "whsec_test_secret"

	repo := &MockRepository{}

	service := NewService(
		repo,
		StripeConfig{
			WebhookSecret: secret,
		},
	)

	payload := stripeEventPayload(
		"checkout.session.completed",
		`{
		"id":"cs_123",
		"object":"checkout.session",
		"customer":"cus_123",
		"subscription":"sub_123",
		"metadata":{}
	}`,
	)

	err := service.HandleWebhook(
		context.Background(),
		payload,
		signedStripeWebhook(payload, secret).Header,
	)

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf(
			"expected ErrInvalidInput got %v",
			err,
		)
	}

	if repo.markCompleteCalled {
		t.Fatal(
			"checkout should not be marked complete",
		)
	}

	if repo.activateCalled {
		t.Fatal(
			"subscription should not be activated",
		)
	}
}

func TestHandleWebhook_CheckoutCompleted_MissingStripeIDs(
	t *testing.T,
) {
	const secret = "whsec_test_secret"

	repo := &MockRepository{}

	service := NewService(
		repo,
		StripeConfig{
			WebhookSecret: secret,
		},
	)

	payload := stripeEventPayload(
		"checkout.session.completed",
		`{
		"id":"cs_123",
		"object":"checkout.session",
		"metadata":{
			"user_id":"5",
			"plan_id":"2"
		}
	}`,
	)

	err := service.HandleWebhook(
		context.Background(),
		payload,
		signedStripeWebhook(payload, secret).Header,
	)

	if !errors.Is(err, ErrWebhookInvalid) {
		t.Fatalf(
			"expected ErrWebhookInvalid got %v",
			err,
		)
	}

	if repo.markCompleteCalled {
		t.Fatal(
			"checkout should not be marked complete",
		)
	}

	if repo.activateCalled {
		t.Fatal(
			"subscription should not be activated",
		)
	}
}

func TestHandleWebhook_CheckoutCompleted_MarkCompleteError(
	t *testing.T,
) {
	const secret = "whsec_test_secret"

	expectedErr := errors.New(
		"mark checkout failed",
	)

	repo := &MockRepository{
		markCompleteErr: expectedErr,
	}

	service := NewService(
		repo,
		StripeConfig{
			WebhookSecret: secret,
		},
	)

	payload := stripeEventPayload(
		"checkout.session.completed",
		`{
		"id":"cs_123",
		"object":"checkout.session",
		"customer":"cus_123",
		"subscription":"sub_123",
		"metadata":{
			"user_id":"5",
			"plan_id":"2"
		}
	}`,
	)

	err := service.HandleWebhook(
		context.Background(),
		payload,
		signedStripeWebhook(payload, secret).Header,
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected mark error got %v",
			err,
		)
	}

	if !repo.markCompleteCalled {
		t.Fatal(
			"expected mark complete attempt",
		)
	}

	if repo.activateCalled {
		t.Fatal(
			"activation should not run after mark failure",
		)
	}
}

func TestHandleWebhook_CheckoutCompleted_ActivationError(
	t *testing.T,
) {
	const secret = "whsec_test_secret"

	expectedErr := errors.New(
		"activation failed",
	)

	repo := &MockRepository{
		activateErr:        expectedErr,
		markCompleteResult: true,
	}

	service := NewService(
		repo,
		StripeConfig{
			WebhookSecret: secret,
		},
	)

	service.subscriptionRetriever = &MockStripeSubscriptionRetriever{
		subscription: testStripeSubscription(
			time.Date(
				2026,
				time.September,
				2,
				12,
				0,
				0,
				0,
				time.UTC,
			),
			time.Date(
				2026,
				time.October,
				2,
				12,
				0,
				0,
				0,
				time.UTC,
			),
		),
	}

	payload := stripeEventPayload(
		"checkout.session.completed",
		`{
		"id":"cs_123",
		"object":"checkout.session",
		"customer":"cus_123",
		"subscription":"sub_123",
		"metadata":{
			"user_id":"5",
			"plan_id":"2"
		}
	}`,
	)

	err := service.HandleWebhook(
		context.Background(),
		payload,
		signedStripeWebhook(payload, secret).Header,
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected activation error got %v",
			err,
		)
	}

	if !repo.markCompleteCalled {
		t.Fatal(
			"expected checkout marked complete",
		)
	}

	if !repo.activateCalled {
		t.Fatal(
			"expected activation attempt",
		)
	}
}

func TestHandleWebhook_SubscriptionDeleted_Success(
	t *testing.T,
) {
	const secret = "whsec_test_secret"

	repo := &MockRepository{}

	service := NewService(
		repo,
		StripeConfig{
			WebhookSecret: secret,
		},
	)

	payload := stripeEventPayload(
		"customer.subscription.deleted",
		`{
		"id":"sub_123",
		"object":"subscription",
		"customer":"cus_123",
		"status":"canceled"
	}`,
	)

	err := service.HandleWebhook(
		context.Background(),
		payload,
		signedStripeWebhook(payload, secret).Header,
	)
	if err != nil {
		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}

	if !repo.updateStatusCalled {
		t.Fatal(
			"expected status update",
		)
	}

	if repo.updatedCustomerID != "cus_123" {
		t.Fatalf(
			"expected customer cus_123 got %q",
			repo.updatedCustomerID,
		)
	}

	if repo.updatedStatus != "cancelled" {
		t.Fatalf(
			"expected cancelled got %q",
			repo.updatedStatus,
		)
	}
}

func TestHandleWebhook_SubscriptionUpdated_Success(
	t *testing.T,
) {
	const secret = "whsec_test_secret"

	repo := &MockRepository{}

	service := NewService(
		repo,
		StripeConfig{
			WebhookSecret: secret,
		},
	)

	payload := stripeEventPayload(
		"customer.subscription.updated",
		`{
		"id":"sub_123",
		"object":"subscription",
		"customer":"cus_123",
		"status":"past_due"
	}`,
	)

	err := service.HandleWebhook(
		context.Background(),
		payload,
		signedStripeWebhook(payload, secret).Header,
	)
	if err != nil {
		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}

	if !repo.updateStatusCalled {
		t.Fatal(
			"expected status update",
		)
	}

	if repo.updatedCustomerID != "cus_123" {
		t.Fatalf(
			"expected customer cus_123 got %q",
			repo.updatedCustomerID,
		)
	}

	if repo.updatedStatus != "past_due" {
		t.Fatalf(
			"expected past_due got %q",
			repo.updatedStatus,
		)
	}
}

func TestHandleWebhook_SubscriptionWithoutCustomer_NoOp(
	t *testing.T,
) {
	const secret = "whsec_test_secret"

	repo := &MockRepository{}

	service := NewService(
		repo,
		StripeConfig{
			WebhookSecret: secret,
		},
	)

	payload := stripeEventPayload(
		"customer.subscription.deleted",
		`{
		"id":"sub_123",
		"object":"subscription",
		"status":"canceled"
	}`,
	)

	err := service.HandleWebhook(
		context.Background(),
		payload,
		signedStripeWebhook(payload, secret).Header,
	)
	if err != nil {
		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}

	if repo.updateStatusCalled {
		t.Fatal(
			"status update should not run without customer",
		)
	}
}

func TestHandleWebhook_UnknownEvent_NoOp(
	t *testing.T,
) {
	const secret = "whsec_test_secret"

	repo := &MockRepository{}

	service := NewService(
		repo,
		StripeConfig{
			WebhookSecret: secret,
		},
	)

	payload := stripeEventPayload(
		"payment_intent.created",
		`{
		"id":"pi_123",
		"object":"payment_intent"
	}`,
	)

	err := service.HandleWebhook(
		context.Background(),
		payload,
		signedStripeWebhook(payload, secret).Header,
	)
	if err != nil {
		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}

	if repo.markCompleteCalled {
		t.Fatal(
			"checkout operation should not run",
		)
	}

	if repo.activateCalled {
		t.Fatal(
			"activation should not run",
		)
	}

	if repo.updateStatusCalled {
		t.Fatal(
			"status update should not run",
		)
	}
}

func TestMetadataUint_Success(t *testing.T) {
	result, err := metadataUint(
		map[string]string{
			"user_id": "42",
		},
		"user_id",
	)

	if err != nil {
		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}

	if result != 42 {
		t.Fatalf(
			"expected 42 got %d",
			result,
		)
	}
}

func TestMetadataUint_Missing(t *testing.T) {
	_, err := metadataUint(
		map[string]string{},
		"user_id",
	)

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf(
			"expected ErrInvalidInput got %v",
			err,
		)
	}
}

func TestMetadataUint_InvalidNumber(t *testing.T) {
	_, err := metadataUint(
		map[string]string{
			"user_id": "abc",
		},
		"user_id",
	)

	if err == nil {
		t.Fatal("expected parse error")
	}
}

func TestValidatePromoCode_InvalidInput(t *testing.T) {
	service := NewService(
		&MockRepository{},
		StripeConfig{},
	)

	tests := []struct {
		name   string
		userID uint
		code   string
	}{
		{
			name:   "zero user",
			userID: 0,
			code:   "SAVE10",
		},
		{
			name:   "empty code",
			userID: 1,
			code:   "",
		},
		{
			name:   "whitespace code",
			userID: 1,
			code:   "   ",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := service.ValidatePromoCode(
				context.Background(),
				test.userID,
				ValidatePromoCodeRequest{
					Code: test.code,
				},
			)

			if result != nil {
				t.Fatal("expected nil response")
			}

			if !errors.Is(err, ErrInvalidInput) {
				t.Fatalf(
					"expected ErrInvalidInput got %v",
					err,
				)
			}
		})
	}
}

func TestValidatePromoCode_NotFound(t *testing.T) {
	service := NewService(
		&MockRepository{
			promoErr: ErrPromoCodeNotFound,
		},
		StripeConfig{},
	)

	_, err := service.ValidatePromoCode(
		context.Background(),
		1,
		ValidatePromoCodeRequest{
			Code: "NOPE",
		},
	)

	if !errors.Is(err, ErrPromoCodeNotFound) {
		t.Fatalf(
			"expected ErrPromoCodeNotFound got %v",
			err,
		)
	}
}

func TestValidatePromoCode_RepositoryError(
	t *testing.T,
) {
	repositoryErr := errors.New("database failed")

	service := NewService(
		&MockRepository{
			promoErr: repositoryErr,
		},
		StripeConfig{},
	)

	_, err := service.ValidatePromoCode(
		context.Background(),
		1,
		ValidatePromoCodeRequest{
			Code: "SAVE10",
		},
	)

	if !errors.Is(err, repositoryErr) {
		t.Fatalf(
			"expected repository error got %v",
			err,
		)
	}
}

func TestValidatePromoCode_Inactive(t *testing.T) {
	service := NewService(
		&MockRepository{
			promo: &PromoCode{
				Code:     "SAVE10",
				IsActive: false,
			},
		},
		StripeConfig{},
	)

	_, err := service.ValidatePromoCode(
		context.Background(),
		1,
		ValidatePromoCodeRequest{
			Code: "SAVE10",
		},
	)

	if !errors.Is(err, ErrPromoCodeInvalid) {
		t.Fatalf(
			"expected invalid promo got %v",
			err,
		)
	}
}

func TestValidatePromoCode_NotStarted(t *testing.T) {
	startsAt := time.Now().Add(time.Hour)

	service := NewService(
		&MockRepository{
			promo: &PromoCode{
				Code:     "FUTURE",
				IsActive: true,
				StartsAt: &startsAt,
			},
		},
		StripeConfig{},
	)

	_, err := service.ValidatePromoCode(
		context.Background(),
		1,
		ValidatePromoCodeRequest{
			Code: "FUTURE",
		},
	)

	if !errors.Is(err, ErrPromoCodeInvalid) {
		t.Fatalf(
			"expected invalid promo got %v",
			err,
		)
	}
}

func TestValidatePromoCode_Expired(t *testing.T) {
	expiresAt := time.Now().Add(-time.Hour)

	service := NewService(
		&MockRepository{
			promo: &PromoCode{
				Code:      "OLD",
				IsActive:  true,
				ExpiresAt: &expiresAt,
			},
		},
		StripeConfig{},
	)

	_, err := service.ValidatePromoCode(
		context.Background(),
		1,
		ValidatePromoCodeRequest{
			Code: "OLD",
		},
	)

	if !errors.Is(err, ErrPromoCodeInvalid) {
		t.Fatalf(
			"expected invalid promo got %v",
			err,
		)
	}
}

func TestValidatePromoCode_MaxUsesReached(
	t *testing.T,
) {
	service := NewService(
		&MockRepository{
			promo: &PromoCode{
				Code:      "FULL",
				IsActive:  true,
				MaxUses:   10,
				TimesUsed: 10,
			},
		},
		StripeConfig{},
	)

	_, err := service.ValidatePromoCode(
		context.Background(),
		1,
		ValidatePromoCodeRequest{
			Code: "FULL",
		},
	)

	if !errors.Is(err, ErrPromoCodeInvalid) {
		t.Fatalf(
			"expected invalid promo got %v",
			err,
		)
	}
}

func TestValidatePromoCode_WrongPlan(t *testing.T) {
	planID := uint(2)

	service := NewService(
		&MockRepository{
			promo: &PromoCode{
				Code:     "PLAN2",
				IsActive: true,
				PlanID:   &planID,
			},
		},
		StripeConfig{},
	)

	_, err := service.ValidatePromoCode(
		context.Background(),
		1,
		ValidatePromoCodeRequest{
			Code:   "PLAN2",
			PlanID: 3,
		},
	)

	if !errors.Is(err, ErrPromoCodeInvalid) {
		t.Fatalf(
			"expected invalid promo got %v",
			err,
		)
	}
}

func TestValidatePromoCode_Success(t *testing.T) {
	planID := uint(2)

	service := NewService(
		&MockRepository{
			promo: &PromoCode{
				Code:             "SAVE20",
				Description:      "20 percent off",
				DiscountType:     "percentage",
				PercentageOff:    20,
				FixedAmountPence: 0,
				FreeMonths:       0,
				IsActive:         true,
				PlanID:           &planID,
			},
		},
		StripeConfig{},
	)

	result, err := service.ValidatePromoCode(
		context.Background(),
		1,
		ValidatePromoCodeRequest{
			Code:   "  SAVE20  ",
			PlanID: 2,
		},
	)

	if err != nil {
		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}

	if result == nil {
		t.Fatal("expected response")
	}

	if !result.Valid {
		t.Fatal("expected valid promo")
	}

	if result.Code != "SAVE20" {
		t.Fatalf(
			"expected SAVE20 got %q",
			result.Code,
		)
	}

	if result.PercentageOff != 20 {
		t.Fatalf(
			"expected 20 got %d",
			result.PercentageOff,
		)
	}
}

func TestValidatePromoCode_GenericPromoAllowsAnyPlan(
	t *testing.T,
) {
	service := NewService(
		&MockRepository{
			promo: &PromoCode{
				Code:     "ANYPLAN",
				IsActive: true,
				PlanID:   nil,
			},
		},
		StripeConfig{},
	)

	result, err := service.ValidatePromoCode(
		context.Background(),
		1,
		ValidatePromoCodeRequest{
			Code:   "ANYPLAN",
			PlanID: 999,
		},
	)

	if err != nil {
		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}

	if result == nil || !result.Valid {
		t.Fatal("expected valid promo")
	}
}

func TestHandleWebhook_CheckoutCompleted_Duplicate_NoOp(t *testing.T) {
	repo := &MockRepository{
		markCompleteResult: false,
	}

	cfg := checkoutConfig()
	cfg.WebhookSecret = "whsec_test"

	service := NewService(repo, cfg)

	periodStart := time.Date(
		2026,
		time.September,
		2,
		12,
		0,
		0,
		0,
		time.UTC,
	)

	periodEnd := time.Date(
		2026,
		time.October,
		2,
		12,
		0,
		0,
		0,
		time.UTC,
	)

	service.subscriptionRetriever = &MockStripeSubscriptionRetriever{
		subscription: testStripeSubscription(
			periodStart,
			periodEnd,
		),
	}

	payload := stripeEventPayload(
		"checkout.session.completed",
		`{
			"id":"cs_123",
			"object":"checkout.session",
			"customer":"cus_123",
			"subscription":"sub_123",
			"metadata":{
				"user_id":"5",
				"plan_id":"2"
			}
		}`,
	)

	err := service.HandleWebhook(
		context.Background(),
		payload,
		signedStripeWebhook(
			payload,
			cfg.WebhookSecret,
		).Header,
	)

	if err != nil {
		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}

	if !repo.markCompleteCalled {
		t.Fatal("expected checkout completion check")
	}

	if repo.activateCalled {
		t.Fatal("duplicate webhook must not activate subscription")
	}
}

func TestHandleWebhook_InvoicePaid_Success(t *testing.T) {
	periodStart := time.Date(
		2026,
		time.September,
		2,
		12,
		0,
		0,
		0,
		time.UTC,
	)

	periodEnd := time.Date(
		2026,
		time.October,
		2,
		12,
		0,
		0,
		0,
		time.UTC,
	)

	repo := &MockRepository{}

	cfg := checkoutConfig()
	cfg.WebhookSecret = "whsec_test"

	service := NewService(repo, cfg)

	retriever := &MockStripeSubscriptionRetriever{
		subscription: testStripeSubscription(
			periodStart,
			periodEnd,
		),
	}

	service.subscriptionRetriever = retriever

	payload := stripeEventPayload(
		"invoice.paid",
		`{
			"id":"in_123",
			"object":"invoice",
			"customer":"cus_123",
			"parent":{
				"type":"subscription_details",
				"subscription_details":{
					"subscription":"sub_123"
				}
			}
		}`,
	)

	err := service.HandleWebhook(
		context.Background(),
		payload,
		signedStripeWebhook(
			payload,
			cfg.WebhookSecret,
		).Header,
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !retriever.called {
		t.Fatal("expected Stripe subscription retrieval")
	}

	if !repo.updateRenewalCalled {
		t.Fatal("expected subscription renewal update")
	}

	if repo.renewalCustomerID != "cus_123" {
		t.Fatalf(
			"expected customer cus_123 got %s",
			repo.renewalCustomerID,
		)
	}

	if repo.renewalSubscriptionID != "sub_123" {
		t.Fatalf(
			"expected subscription sub_123 got %s",
			repo.renewalSubscriptionID,
		)
	}

	if !repo.renewalPeriodStart.Equal(periodStart) {
		t.Fatalf(
			"expected period start %v got %v",
			periodStart,
			repo.renewalPeriodStart,
		)
	}

	if !repo.renewalPeriodEnd.Equal(periodEnd) {
		t.Fatalf(
			"expected period end %v got %v",
			periodEnd,
			repo.renewalPeriodEnd,
		)
	}
}

func TestHandleWebhook_CheckoutSessionExpired_Success(t *testing.T) {
	repo := &MockRepository{}

	cfg := checkoutConfig()
	cfg.WebhookSecret = "whsec_test"

	service := NewService(repo, cfg)

	payload := stripeEventPayload(
		"checkout.session.expired",
		`{
			"id":"cs_expired_123",
			"object":"checkout.session"
		}`,
	)

	err := service.HandleWebhook(
		context.Background(),
		payload,
		signedStripeWebhook(
			payload,
			cfg.WebhookSecret,
		).Header,
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !repo.markExpiredCalled {
		t.Fatal("expected expired checkout update")
	}

	if repo.markedExpiredSession != "cs_expired_123" {
		t.Fatalf(
			"expected cs_expired_123 got %s",
			repo.markedExpiredSession,
		)
	}
}

func TestHandleWebhook_CheckoutSessionExpired_Error(t *testing.T) {
	expectedErr := errors.New("database error")

	repo := &MockRepository{
		markExpiredErr: expectedErr,
	}

	cfg := checkoutConfig()
	cfg.WebhookSecret = "whsec_test"

	service := NewService(repo, cfg)

	payload := stripeEventPayload(
		"checkout.session.expired",
		`{
			"id":"cs_expired_123",
			"object":"checkout.session"
		}`,
	)

	err := service.HandleWebhook(
		context.Background(),
		payload,
		signedStripeWebhook(
			payload,
			cfg.WebhookSecret,
		).Header,
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected %v got %v",
			expectedErr,
			err,
		)
	}
}

var _ Repository = (*MockRepository)(nil)
