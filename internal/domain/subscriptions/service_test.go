package subscriptions

import (
	"context"
	"errors"
	"testing"
)

type MockRepository struct {
	plans        []SubscriptionPlan
	plan         *SubscriptionPlan
	subscription *UserSubscription

	listPlansErr      error
	getPlanErr        error
	createOrUpdateErr error
	getByUserErr      error
	updateStatusErr   error

	lastCreateUserID uint
	lastCreatePlanID uint
	lastCreateStatus string

	lastUpdateUserID uint
	lastUpdateStatus string
}

func (m *MockRepository) ListPlans(
	ctx context.Context,
) ([]SubscriptionPlan, error) {
	if m.listPlansErr != nil {
		return nil, m.listPlansErr
	}

	return m.plans, nil
}

func (m *MockRepository) GetPlanByID(
	ctx context.Context,
	id uint,
) (*SubscriptionPlan, error) {
	if m.getPlanErr != nil {
		return nil, m.getPlanErr
	}

	return m.plan, nil
}

func (m *MockRepository) CreateOrUpdateUserSubscription(
	ctx context.Context,
	userID uint,
	planID uint,
	status string,
) (*UserSubscription, error) {
	m.lastCreateUserID = userID
	m.lastCreatePlanID = planID
	m.lastCreateStatus = status

	if m.createOrUpdateErr != nil {
		return nil, m.createOrUpdateErr
	}

	return m.subscription, nil
}

func (m *MockRepository) GetByUserID(
	ctx context.Context,
	userID uint,
) (*UserSubscription, error) {
	if m.getByUserErr != nil {
		return nil, m.getByUserErr
	}

	return m.subscription, nil
}

func (m *MockRepository) UpdateStatus(
	ctx context.Context,
	userID uint,
	status string,
) error {
	m.lastUpdateUserID = userID
	m.lastUpdateStatus = status

	return m.updateStatusErr
}

func TestNewService(t *testing.T) {
	repo := &MockRepository{}

	service := NewService(repo)

	if service == nil {
		t.Fatal("expected service")
	}

	if service.repo != repo {
		t.Fatal("expected repository assigned")
	}
}

func TestListPlans_Success(t *testing.T) {
	service := NewService(
		&MockRepository{
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
		},
	)

	plans, err := service.ListPlans(
		context.Background(),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(plans) != 2 {
		t.Fatalf(
			"expected 2 plans got %d",
			len(plans),
		)
	}
}

func TestListPlans_RepositoryError(t *testing.T) {
	listErr := errors.New("list failed")

	service := NewService(
		&MockRepository{
			listPlansErr: listErr,
		},
	)

	_, err := service.ListPlans(
		context.Background(),
	)

	if !errors.Is(err, listErr) {
		t.Fatalf(
			"expected list error got %v",
			err,
		)
	}
}

func TestGetMine_InvalidUserID(t *testing.T) {
	service := NewService(
		&MockRepository{},
	)

	subscription, err := service.GetMine(
		context.Background(),
		0,
	)

	if subscription != nil {
		t.Fatal("expected nil subscription")
	}

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf(
			"expected ErrInvalidInput got %v",
			err,
		)
	}
}

func TestGetMine_Success(t *testing.T) {
	expected := &UserSubscription{
		ID:     1,
		UserID: 10,
		Status: "active",
	}

	service := NewService(
		&MockRepository{
			subscription: expected,
		},
	)

	result, err := service.GetMine(
		context.Background(),
		10,
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != expected {
		t.Fatal("expected subscription")
	}
}

func TestGetMine_RepositoryError(t *testing.T) {
	repositoryErr := errors.New("lookup failed")

	service := NewService(
		&MockRepository{
			getByUserErr: repositoryErr,
		},
	)

	_, err := service.GetMine(
		context.Background(),
		10,
	)

	if !errors.Is(err, repositoryErr) {
		t.Fatalf(
			"expected repository error got %v",
			err,
		)
	}
}

func TestCreateOrUpdate_InvalidInput(t *testing.T) {
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
			service := NewService(
				&MockRepository{},
			)

			subscription, err := service.CreateOrUpdate(
				context.Background(),
				test.userID,
				CreateSubscriptionRequest{
					PlanID: test.planID,
				},
			)

			if subscription != nil {
				t.Fatal(
					"expected nil subscription",
				)
			}

			if !errors.Is(
				err,
				ErrInvalidInput,
			) {
				t.Fatalf(
					"expected ErrInvalidInput got %v",
					err,
				)
			}
		})
	}
}

func TestCreateOrUpdate_PlanNotFound(t *testing.T) {
	service := NewService(
		&MockRepository{
			getPlanErr: ErrPlanNotFound,
		},
	)

	subscription, err := service.CreateOrUpdate(
		context.Background(),
		1,
		CreateSubscriptionRequest{
			PlanID: 999,
		},
	)

	if subscription != nil {
		t.Fatal(
			"expected nil subscription",
		)
	}

	if !errors.Is(err, ErrPlanNotFound) {
		t.Fatalf(
			"expected ErrPlanNotFound got %v",
			err,
		)
	}
}

func TestCreateOrUpdate_PlanLookupError(
	t *testing.T,
) {
	lookupErr := errors.New("lookup failed")

	service := NewService(
		&MockRepository{
			getPlanErr: lookupErr,
		},
	)

	_, err := service.CreateOrUpdate(
		context.Background(),
		1,
		CreateSubscriptionRequest{
			PlanID: 2,
		},
	)

	if !errors.Is(err, lookupErr) {
		t.Fatalf(
			"expected lookup error got %v",
			err,
		)
	}
}

func TestCreateOrUpdate_Success(t *testing.T) {
	expected := &UserSubscription{
		ID:     5,
		UserID: 10,
		Status: "trial",
	}

	repo := &MockRepository{
		plan: &SubscriptionPlan{
			ID:   2,
			Name: "Standard",
		},
		subscription: expected,
	}

	service := NewService(repo)

	result, err := service.CreateOrUpdate(
		context.Background(),
		10,
		CreateSubscriptionRequest{
			PlanID: 2,
		},
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != expected {
		t.Fatal(
			"expected created subscription",
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

func TestCreateOrUpdate_RepositoryError(
	t *testing.T,
) {
	createErr := errors.New("create failed")

	service := NewService(
		&MockRepository{
			plan: &SubscriptionPlan{
				ID: 1,
			},
			createOrUpdateErr: createErr,
		},
	)

	_, err := service.CreateOrUpdate(
		context.Background(),
		1,
		CreateSubscriptionRequest{
			PlanID: 1,
		},
	)

	if !errors.Is(err, createErr) {
		t.Fatalf(
			"expected create error got %v",
			err,
		)
	}
}

func TestUpdateStatus_InvalidInput(t *testing.T) {
	tests := []struct {
		name   string
		userID uint
		status string
	}{
		{
			name:   "zero user",
			userID: 0,
			status: "active",
		},
		{
			name:   "empty status",
			userID: 1,
			status: "",
		},
		{
			name:   "whitespace status",
			userID: 1,
			status: "   ",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service := NewService(
				&MockRepository{},
			)

			err := service.UpdateStatus(
				context.Background(),
				test.userID,
				UpdateSubscriptionStatusRequest{
					Status: test.status,
				},
			)

			if !errors.Is(
				err,
				ErrInvalidInput,
			) {
				t.Fatalf(
					"expected ErrInvalidInput got %v",
					err,
				)
			}
		})
	}
}

func TestUpdateStatus_InvalidStatus(t *testing.T) {
	service := NewService(
		&MockRepository{},
	)

	err := service.UpdateStatus(
		context.Background(),
		1,
		UpdateSubscriptionStatusRequest{
			Status: "banana",
		},
	)

	if !errors.Is(err, ErrInvalidStatus) {
		t.Fatalf(
			"expected ErrInvalidStatus got %v",
			err,
		)
	}
}

func TestUpdateStatus_Success(t *testing.T) {
	repo := &MockRepository{}
	service := NewService(repo)

	err := service.UpdateStatus(
		context.Background(),
		10,
		UpdateSubscriptionStatusRequest{
			Status: "  active  ",
		},
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
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

func TestUpdateStatus_RepositoryError(
	t *testing.T,
) {
	updateErr := errors.New("update failed")

	service := NewService(
		&MockRepository{
			updateStatusErr: updateErr,
		},
	)

	err := service.UpdateStatus(
		context.Background(),
		10,
		UpdateSubscriptionStatusRequest{
			Status: "cancelled",
		},
	)

	if !errors.Is(err, updateErr) {
		t.Fatalf(
			"expected update error got %v",
			err,
		)
	}
}

func TestIsAllowedStatus(t *testing.T) {
	valid := []string{
		"trial",
		"active",
		"past_due",
		"cancelled",
		"expired",
	}

	for _, status := range valid {
		if !isAllowedStatus(status) {
			t.Fatalf(
				"expected %q to be valid",
				status,
			)
		}
	}

	invalid := []string{
		"",
		"banana",
		"pending",
		"deleted",
	}

	for _, status := range invalid {
		if isAllowedStatus(status) {
			t.Fatalf(
				"expected %q to be invalid",
				status,
			)
		}
	}
}

var _ Repository = (*MockRepository)(nil)
