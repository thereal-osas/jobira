package adminsubscriptions

import (
	"context"
	"errors"
	"testing"
)

type mockRepository struct {
	listPlansFn func(context.Context) ([]AdminSubscriptionPlan, error)

	createPlanFn func(
		context.Context,
		*AdminSubscriptionPlan,
	) error

	updatePlanFn func(
		context.Context,
		*AdminSubscriptionPlan,
	) error

	disablePlanFn func(
		context.Context,
		uint,
	) error

	listUserSubscriptionsFn func(
		context.Context,
	) ([]AdminUserSubscription, error)

	getUserSubscriptionsFn func(
		context.Context,
		uint,
	) (*AdminUserSubscription, error)

	createOrUpdateUserSubscriptionFn func(
		context.Context,
		uint,
		uint,
		string,
	) error
}

func (m *mockRepository) ListPlans(
	ctx context.Context,
) ([]AdminSubscriptionPlan, error) {
	if m.listPlansFn != nil {
		return m.listPlansFn(ctx)
	}
	return nil, nil
}

func (m *mockRepository) CreatePlan(
	ctx context.Context,
	plan *AdminSubscriptionPlan,
) error {
	if m.createPlanFn != nil {
		return m.createPlanFn(ctx, plan)
	}
	return nil
}

func (m *mockRepository) UpdatePlan(
	ctx context.Context,
	plan *AdminSubscriptionPlan,
) error {
	if m.updatePlanFn != nil {
		return m.updatePlanFn(ctx, plan)
	}
	return nil
}

func (m *mockRepository) DisablePlan(
	ctx context.Context,
	planID uint,
) error {
	if m.disablePlanFn != nil {
		return m.disablePlanFn(ctx, planID)
	}
	return nil
}

func (m *mockRepository) ListUserSubscriptions(
	ctx context.Context,
) ([]AdminUserSubscription, error) {
	if m.listUserSubscriptionsFn != nil {
		return m.listUserSubscriptionsFn(ctx)
	}
	return nil, nil
}

func (m *mockRepository) GetUserSubscriptions(
	ctx context.Context,
	userID uint,
) (*AdminUserSubscription, error) {
	if m.getUserSubscriptionsFn != nil {
		return m.getUserSubscriptionsFn(ctx, userID)
	}
	return nil, nil
}

func (m *mockRepository) CreateOrUpdateUserSubscription(
	ctx context.Context,
	userID uint,
	planID uint,
	status string,
) error {
	if m.createOrUpdateUserSubscriptionFn != nil {
		return m.createOrUpdateUserSubscriptionFn(
			ctx,
			userID,
			planID,
			status,
		)
	}
	return nil
}

func TestNewService(t *testing.T) {
	repo := &mockRepository{}

	service := NewService(repo)

	if service == nil {
		t.Fatal("expected service")
	}

	if service.repo != repo {
		t.Fatal("expected repository assigned")
	}
}

func TestService_ListPlans_Success(t *testing.T) {
	repo := &mockRepository{
		listPlansFn: func(
			context.Context,
		) ([]AdminSubscriptionPlan, error) {
			return []AdminSubscriptionPlan{
				{ID: 1, Name: "Standard"},
				{ID: 2, Name: "Premium"},
			}, nil
		},
	}

	service := NewService(repo)

	plans, err := service.ListPlans(
		context.Background(),
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(plans) != 2 {
		t.Fatalf(
			"expected 2 plans, got %d",
			len(plans),
		)
	}
}

func TestService_ListPlans_Error(t *testing.T) {
	expectedErr := errors.New("list failed")

	repo := &mockRepository{
		listPlansFn: func(
			context.Context,
		) ([]AdminSubscriptionPlan, error) {
			return nil, expectedErr
		},
	}

	service := NewService(repo)

	_, err := service.ListPlans(
		context.Background(),
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected %v, got %v",
			expectedErr,
			err,
		)
	}
}

func TestService_CreatePlan_Success(t *testing.T) {
	createCalled := false

	repo := &mockRepository{
		createPlanFn: func(
			ctx context.Context,
			plan *AdminSubscriptionPlan,
		) error {
			createCalled = true

			if plan.Name != "Standard" {
				t.Fatalf(
					"expected trimmed name, got %q",
					plan.Name,
				)
			}

			if plan.RoleType != "cleaner" {
				t.Fatalf(
					"expected cleaner, got %q",
					plan.RoleType,
				)
			}

			if plan.BillingInterval != "monthly" {
				t.Fatalf(
					"expected monthly, got %q",
					plan.BillingInterval,
				)
			}

			if !plan.IsActive {
				t.Fatal("expected active plan")
			}

			plan.ID = 10

			return nil
		},
	}

	service := NewService(repo)

	plan, err := service.CreatePlan(
		context.Background(),
		CreatePlanRequest{
			Name:             " Standard ",
			RoleType:         " cleaner ",
			PricePence:       1799,
			BillingInterval:  " monthly ",
			ApplicationLimit: 50,
			JobPostLimit:     0,
			CleanerSeatLimit: 1,
		},
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !createCalled {
		t.Fatal("expected CreatePlan called")
	}

	if plan.ID != 10 {
		t.Fatalf(
			"expected ID 10, got %d",
			plan.ID,
		)
	}
}

func TestService_CreatePlan_InvalidInput(
	t *testing.T,
) {
	tests := []struct {
		name string
		req  CreatePlanRequest
	}{
		{
			name: "empty name",
			req: CreatePlanRequest{
				RoleType:         "cleaner",
				BillingInterval:  "monthly",
				CleanerSeatLimit: 1,
			},
		},
		{
			name: "empty role",
			req: CreatePlanRequest{
				Name:             "Standard",
				BillingInterval:  "monthly",
				CleanerSeatLimit: 1,
			},
		},
		{
			name: "empty interval",
			req: CreatePlanRequest{
				Name:             "Standard",
				RoleType:         "cleaner",
				CleanerSeatLimit: 1,
			},
		},
		{
			name: "negative price",
			req: CreatePlanRequest{
				Name:             "Standard",
				RoleType:         "cleaner",
				BillingInterval:  "monthly",
				PricePence:       -1,
				CleanerSeatLimit: 1,
			},
		},
		{
			name: "negative application limit",
			req: CreatePlanRequest{
				Name:             "Standard",
				RoleType:         "cleaner",
				BillingInterval:  "monthly",
				ApplicationLimit: -1,
				CleanerSeatLimit: 1,
			},
		},
		{
			name: "negative job post limit",
			req: CreatePlanRequest{
				Name:             "Standard",
				RoleType:         "cleaner",
				BillingInterval:  "monthly",
				JobPostLimit:     -1,
				CleanerSeatLimit: 1,
			},
		},
		{
			name: "zero seat limit",
			req: CreatePlanRequest{
				Name:            "Standard",
				RoleType:        "cleaner",
				BillingInterval: "monthly",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service := NewService(
				&mockRepository{},
			)

			plan, err := service.CreatePlan(
				context.Background(),
				test.req,
			)

			if plan != nil {
				t.Fatal("expected nil plan")
			}

			if !errors.Is(
				err,
				ErrInvalidInput,
			) {
				t.Fatalf(
					"expected ErrInvalidInput, got %v",
					err,
				)
			}
		})
	}
}

func TestService_CreatePlan_RepositoryError(
	t *testing.T,
) {
	expectedErr := errors.New("create failed")

	repo := &mockRepository{
		createPlanFn: func(
			context.Context,
			*AdminSubscriptionPlan,
		) error {
			return expectedErr
		},
	}

	service := NewService(repo)

	plan, err := service.CreatePlan(
		context.Background(),
		CreatePlanRequest{
			Name:             "Standard",
			RoleType:         "cleaner",
			BillingInterval:  "monthly",
			CleanerSeatLimit: 1,
		},
	)

	if plan != nil {
		t.Fatal("expected nil plan")
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected %v, got %v",
			expectedErr,
			err,
		)
	}
}

func TestService_UpdatePlan_Success(t *testing.T) {
	repo := &mockRepository{
		updatePlanFn: func(
			ctx context.Context,
			plan *AdminSubscriptionPlan,
		) error {
			if plan.ID != 7 {
				t.Fatalf(
					"expected ID 7, got %d",
					plan.ID,
				)
			}

			if plan.Name != "Premium" {
				t.Fatalf(
					"expected Premium, got %q",
					plan.Name,
				)
			}

			if !plan.IsActive {
				t.Fatal("expected active")
			}

			return nil
		},
	}

	service := NewService(repo)

	err := service.UpdatePlan(
		context.Background(),
		7,
		UpdatePlanRequest{
			Name:             " Premium ",
			RoleType:         " cleaner ",
			PricePence:       2499,
			BillingInterval:  " monthly ",
			ApplicationLimit: 100,
			CleanerSeatLimit: 1,
			IsActive:         true,
		},
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestService_UpdatePlan_InvalidInput(
	t *testing.T,
) {
	tests := []struct {
		name   string
		planID uint
		req    UpdatePlanRequest
	}{
		{
			name:   "zero ID",
			planID: 0,
			req: UpdatePlanRequest{
				Name:             "Standard",
				RoleType:         "cleaner",
				BillingInterval:  "monthly",
				CleanerSeatLimit: 1,
			},
		},
		{
			name:   "empty name",
			planID: 1,
			req: UpdatePlanRequest{
				RoleType:         "cleaner",
				BillingInterval:  "monthly",
				CleanerSeatLimit: 1,
			},
		},
		{
			name:   "negative price",
			planID: 1,
			req: UpdatePlanRequest{
				Name:             "Standard",
				RoleType:         "cleaner",
				BillingInterval:  "monthly",
				PricePence:       -1,
				CleanerSeatLimit: 1,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service := NewService(
				&mockRepository{},
			)

			err := service.UpdatePlan(
				context.Background(),
				test.planID,
				test.req,
			)

			if !errors.Is(
				err,
				ErrInvalidInput,
			) {
				t.Fatalf(
					"expected ErrInvalidInput, got %v",
					err,
				)
			}
		})
	}
}

func TestService_UpdatePlan_RepositoryError(
	t *testing.T,
) {
	expectedErr := ErrPlanNotFound

	repo := &mockRepository{
		updatePlanFn: func(
			context.Context,
			*AdminSubscriptionPlan,
		) error {
			return expectedErr
		},
	}

	service := NewService(repo)

	err := service.UpdatePlan(
		context.Background(),
		7,
		UpdatePlanRequest{
			Name:             "Standard",
			RoleType:         "cleaner",
			BillingInterval:  "monthly",
			CleanerSeatLimit: 1,
		},
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected %v, got %v",
			expectedErr,
			err,
		)
	}
}

func TestService_DisablePlan_Success(t *testing.T) {
	called := false

	repo := &mockRepository{
		disablePlanFn: func(
			ctx context.Context,
			planID uint,
		) error {
			called = true

			if planID != 5 {
				t.Fatalf(
					"expected ID 5, got %d",
					planID,
				)
			}

			return nil
		},
	}

	service := NewService(repo)

	err := service.DisablePlan(
		context.Background(),
		5,
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !called {
		t.Fatal("expected DisablePlan called")
	}
}

func TestService_DisablePlan_InvalidInput(
	t *testing.T,
) {
	service := NewService(
		&mockRepository{},
	)

	err := service.DisablePlan(
		context.Background(),
		0,
	)

	if !errors.Is(
		err,
		ErrInvalidInput,
	) {
		t.Fatalf(
			"expected ErrInvalidInput, got %v",
			err,
		)
	}
}

func TestService_ListUserSubscriptions_Success(
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

	service := NewService(repo)

	subs, err := service.ListUserSubscriptions(
		context.Background(),
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(subs) != 1 {
		t.Fatalf(
			"expected 1 subscription, got %d",
			len(subs),
		)
	}
}

func TestService_ListUserSubscriptions_Error(
	t *testing.T,
) {
	expectedErr := errors.New("list failed")

	repo := &mockRepository{
		listUserSubscriptionsFn: func(
			context.Context,
		) ([]AdminUserSubscription, error) {
			return nil, expectedErr
		},
	}

	service := NewService(repo)

	_, err := service.ListUserSubscriptions(
		context.Background(),
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected %v, got %v",
			expectedErr,
			err,
		)
	}
}

func TestService_GetUserSubscriptions_Success(
	t *testing.T,
) {
	repo := &mockRepository{
		getUserSubscriptionsFn: func(
			ctx context.Context,
			userID uint,
		) (*AdminUserSubscription, error) {
			if userID != 8 {
				t.Fatalf(
					"expected user ID 8, got %d",
					userID,
				)
			}

			return &AdminUserSubscription{
				ID:     1,
				UserID: 8,
				Status: "active",
			}, nil
		},
	}

	service := NewService(repo)

	sub, err := service.GetUserSubscriptions(
		context.Background(),
		8,
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if sub.UserID != 8 {
		t.Fatalf(
			"expected user ID 8, got %d",
			sub.UserID,
		)
	}
}

func TestService_GetUserSubscriptions_InvalidInput(
	t *testing.T,
) {
	service := NewService(
		&mockRepository{},
	)

	sub, err := service.GetUserSubscriptions(
		context.Background(),
		0,
	)

	if sub != nil {
		t.Fatal("expected nil subscription")
	}

	if !errors.Is(
		err,
		ErrInvalidInput,
	) {
		t.Fatalf(
			"expected ErrInvalidInput, got %v",
			err,
		)
	}
}

func TestService_GetUserSubscriptions_Error(
	t *testing.T,
) {
	expectedErr := ErrSubscriptionNotFound

	repo := &mockRepository{
		getUserSubscriptionsFn: func(
			context.Context,
			uint,
		) (*AdminUserSubscription, error) {
			return nil, expectedErr
		},
	}

	service := NewService(repo)

	_, err := service.GetUserSubscriptions(
		context.Background(),
		8,
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected %v, got %v",
			expectedErr,
			err,
		)
	}
}

func TestService_CreateOrUpdateUserSubscriptions_Success(
	t *testing.T,
) {
	repo := &mockRepository{
		createOrUpdateUserSubscriptionFn: func(
			ctx context.Context,
			userID uint,
			planID uint,
			status string,
		) error {
			if userID != 8 {
				t.Fatalf(
					"expected user ID 8, got %d",
					userID,
				)
			}

			if planID != 2 {
				t.Fatalf(
					"expected plan ID 2, got %d",
					planID,
				)
			}

			if status != "active" {
				t.Fatalf(
					"expected active, got %q",
					status,
				)
			}

			return nil
		},
	}

	service := NewService(repo)

	err := service.CreateOrUpdateUserSubscriptions(
		context.Background(),
		8,
		UpdateUserSubscriptionRequest{
			PlanID: 2,
			Status: " active ",
		},
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestService_CreateOrUpdateUserSubscriptions_InvalidInput(
	t *testing.T,
) {
	tests := []struct {
		name   string
		userID uint
		req    UpdateUserSubscriptionRequest
	}{
		{
			name:   "zero user ID",
			userID: 0,
			req: UpdateUserSubscriptionRequest{
				PlanID: 1,
				Status: "active",
			},
		},
		{
			name:   "zero plan ID",
			userID: 8,
			req: UpdateUserSubscriptionRequest{
				Status: "active",
			},
		},
		{
			name:   "empty status",
			userID: 8,
			req: UpdateUserSubscriptionRequest{
				PlanID: 1,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service := NewService(
				&mockRepository{},
			)

			err := service.CreateOrUpdateUserSubscriptions(
				context.Background(),
				test.userID,
				test.req,
			)

			if !errors.Is(
				err,
				ErrInvalidInput,
			) {
				t.Fatalf(
					"expected ErrInvalidInput, got %v",
					err,
				)
			}
		})
	}
}

func TestService_CreateOrUpdateUserSubscriptions_InvalidStatus(
	t *testing.T,
) {
	service := NewService(
		&mockRepository{},
	)

	err := service.CreateOrUpdateUserSubscriptions(
		context.Background(),
		8,
		UpdateUserSubscriptionRequest{
			PlanID: 1,
			Status: "pending",
		},
	)

	if !errors.Is(
		err,
		ErrInvalidStatus,
	) {
		t.Fatalf(
			"expected ErrInvalidStatus, got %v",
			err,
		)
	}
}

func TestService_CreateOrUpdateUserSubscriptions_AllStatuses(
	t *testing.T,
) {
	statuses := []string{
		"trial",
		"active",
		"past_due",
		"cancelled",
		"expired",
	}

	for _, status := range statuses {
		t.Run(status, func(t *testing.T) {
			service := NewService(
				&mockRepository{},
			)

			err := service.CreateOrUpdateUserSubscriptions(
				context.Background(),
				8,
				UpdateUserSubscriptionRequest{
					PlanID: 1,
					Status: status,
				},
			)

			if err != nil {
				t.Fatalf(
					"unexpected error for %q: %v",
					status,
					err,
				)
			}
		})
	}
}

func TestService_CreateOrUpdateUserSubscriptions_RepositoryError(
	t *testing.T,
) {
	expectedErr := errors.New("update failed")

	repo := &mockRepository{
		createOrUpdateUserSubscriptionFn: func(
			context.Context,
			uint,
			uint,
			string,
		) error {
			return expectedErr
		},
	}

	service := NewService(repo)

	err := service.CreateOrUpdateUserSubscriptions(
		context.Background(),
		8,
		UpdateUserSubscriptionRequest{
			PlanID: 1,
			Status: "active",
		},
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected %v, got %v",
			expectedErr,
			err,
		)
	}
}

func TestIsAllowedStatus(t *testing.T) {
	tests := []struct {
		status   string
		expected bool
	}{
		{"trial", true},
		{"active", true},
		{"past_due", true},
		{"cancelled", true},
		{"expired", true},
		{"pending", false},
		{"", false},
	}

	for _, test := range tests {
		t.Run(test.status, func(t *testing.T) {
			actual := isAllowedStatus(
				test.status,
			)

			if actual != test.expected {
				t.Fatalf(
					"expected %v for %q, got %v",
					test.expected,
					test.status,
					actual,
				)
			}
		})
	}
}

