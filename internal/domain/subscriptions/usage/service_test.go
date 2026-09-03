package usage

import (
	"context"
	"errors"
	"testing"
)

type MockRepository struct {
	usage  *UserUsage
	access *SubscriptionAccess

	createErr                error
	getByUserErr             error
	getSubscriptionAccessErr error
	incrementApplicationErr  error
	incrementJobPostErr      error
	setMonetisationErr       error

	createCalls int

	lastApplicationUserID uint
	lastJobPostUserID     uint

	lastMonetisationUserID  uint
	lastMonetisationEnabled bool
}

func (m *MockRepository) CreateIfNotExists(
	ctx context.Context,
	userID uint,
) error {
	m.createCalls++
	return m.createErr
}

func (m *MockRepository) GetByUserID(
	ctx context.Context,
	userID uint,
) (*UserUsage, error) {
	if m.getByUserErr != nil {
		return nil, m.getByUserErr
	}

	return m.usage, nil
}

func (m *MockRepository) GetUserSubscriptionAccess(
	ctx context.Context,
	userID uint,
) (*SubscriptionAccess, error) {
	if m.getSubscriptionAccessErr != nil {
		return nil, m.getSubscriptionAccessErr
	}

	return m.access, nil
}

func (m *MockRepository) IncrementApplicationCount(
	ctx context.Context,
	userID uint,
) error {
	m.lastApplicationUserID = userID
	return m.incrementApplicationErr
}

func (m *MockRepository) IncrementJobPostCount(
	ctx context.Context,
	userID uint,
) error {
	m.lastJobPostUserID = userID
	return m.incrementJobPostErr
}

func (m *MockRepository) SetMonetisationEnabled(
	ctx context.Context,
	userID uint,
	enabled bool,
) error {
	m.lastMonetisationUserID = userID
	m.lastMonetisationEnabled = enabled

	return m.setMonetisationErr
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

func TestEnsureUsage_InvalidUser(t *testing.T) {
	service := NewService(
		&MockRepository{},
	)

	err := service.EnsureUsage(
		context.Background(),
		0,
	)

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf(
			"expected ErrInvalidInput got %v",
			err,
		)
	}
}

func TestEnsureUsage_Success(t *testing.T) {
	repo := &MockRepository{}
	service := NewService(repo)

	err := service.EnsureUsage(
		context.Background(),
		5,
	)
	if err != nil {
		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}

	if repo.createCalls != 1 {
		t.Fatalf(
			"expected one create call got %d",
			repo.createCalls,
		)
	}
}

func TestEnsureUsage_RepositoryError(t *testing.T) {
	createErr := errors.New("create failed")

	service := NewService(
		&MockRepository{
			createErr: createErr,
		},
	)

	err := service.EnsureUsage(
		context.Background(),
		5,
	)

	if !errors.Is(err, createErr) {
		t.Fatalf(
			"expected create error got %v",
			err,
		)
	}
}

func TestGetByUserID_InvalidUser(t *testing.T) {
	service := NewService(
		&MockRepository{},
	)

	result, err := service.GetByUserID(
		context.Background(),
		0,
	)

	if result != nil {
		t.Fatal("expected nil usage")
	}

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf(
			"expected ErrInvalidInput got %v",
			err,
		)
	}
}

func TestGetByUserID_CreateError(t *testing.T) {
	createErr := errors.New("create failed")

	service := NewService(
		&MockRepository{
			createErr: createErr,
		},
	)

	_, err := service.GetByUserID(
		context.Background(),
		5,
	)

	if !errors.Is(err, createErr) {
		t.Fatalf(
			"expected create error got %v",
			err,
		)
	}
}

func TestGetByUserID_Success(t *testing.T) {
	expected := &UserUsage{
		ID:                   1,
		UserID:               5,
		ApplicationCount:     2,
		JobPostCount:         1,
		FreeApplicationLimit: 5,
		FreeJobPostLimit:     5,
	}

	service := NewService(
		&MockRepository{
			usage: expected,
		},
	)

	result, err := service.GetByUserID(
		context.Background(),
		5,
	)
	if err != nil {
		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}

	if result != expected {
		t.Fatal("expected usage")
	}
}

func TestGetByUserID_RepositoryError(t *testing.T) {
	getErr := errors.New("lookup failed")

	service := NewService(
		&MockRepository{
			getByUserErr: getErr,
		},
	)

	_, err := service.GetByUserID(
		context.Background(),
		5,
	)

	if !errors.Is(err, getErr) {
		t.Fatalf(
			"expected lookup error got %v",
			err,
		)
	}
}

func TestCanApply_InvalidUser(t *testing.T) {
	service := NewService(
		&MockRepository{},
	)

	err := service.CanApply(
		context.Background(),
		0,
	)

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf(
			"expected ErrInvalidInput got %v",
			err,
		)
	}
}

func TestCanApply_CreateUsageError(t *testing.T) {
	createErr := errors.New("create failed")

	service := NewService(
		&MockRepository{
			createErr: createErr,
		},
	)

	err := service.CanApply(
		context.Background(),
		5,
	)

	if !errors.Is(err, createErr) {
		t.Fatalf(
			"expected create error got %v",
			err,
		)
	}
}

func TestCanApply_GetUsageError(t *testing.T) {
	getErr := errors.New("lookup failed")

	service := NewService(
		&MockRepository{
			getByUserErr: getErr,
		},
	)

	err := service.CanApply(
		context.Background(),
		5,
	)

	if !errors.Is(err, getErr) {
		t.Fatalf(
			"expected lookup error got %v",
			err,
		)
	}
}

func TestCanApply_MonetisationDisabled(t *testing.T) {
	repo := &MockRepository{
		usage: &UserUsage{
			UserID:               5,
			ApplicationCount:     999,
			FreeApplicationLimit: 5,
			MonetisationEnabled:  false,
		},
	}

	service := NewService(repo)

	err := service.CanApply(
		context.Background(),
		5,
	)

	if err != nil {
		t.Fatalf(
			"expected unrestricted access got %v",
			err,
		)
	}
}

func TestCanApply_SubscriptionAccessError(t *testing.T) {
	accessErr := errors.New(
		"subscription lookup failed",
	)

	service := NewService(
		&MockRepository{
			usage: &UserUsage{
				UserID:              5,
				MonetisationEnabled: true,
			},
			getSubscriptionAccessErr: accessErr,
		},
	)

	err := service.CanApply(
		context.Background(),
		5,
	)

	if !errors.Is(err, accessErr) {
		t.Fatalf(
			"expected access error got %v",
			err,
		)
	}
}

func TestCanApply_ActiveSubscriptionUnlimitedZeroRejected(
	t *testing.T,
) {
	service := NewService(
		&MockRepository{
			usage: &UserUsage{
				UserID:              5,
				MonetisationEnabled: true,
			},
			access: &SubscriptionAccess{
				Status:           "active",
				ApplicationLimit: 0,
			},
		},
	)

	err := service.CanApply(
		context.Background(),
		5,
	)

	if !errors.Is(
		err,
		ErrApplicationLimitReached,
	) {
		t.Fatalf(
			"expected application limit error got %v",
			err,
		)
	}
}

func TestCanApply_ActiveSubscriptionLimitReached(
	t *testing.T,
) {
	service := NewService(
		&MockRepository{
			usage: &UserUsage{
				UserID:              5,
				ApplicationCount:    10,
				MonetisationEnabled: true,
			},
			access: &SubscriptionAccess{
				Status:           "active",
				ApplicationLimit: 10,
			},
		},
	)

	err := service.CanApply(
		context.Background(),
		5,
	)

	if !errors.Is(
		err,
		ErrApplicationLimitReached,
	) {
		t.Fatalf(
			"expected limit reached got %v",
			err,
		)
	}
}

func TestCanApply_ActiveSubscriptionAllowed(
	t *testing.T,
) {
	service := NewService(
		&MockRepository{
			usage: &UserUsage{
				UserID:              5,
				ApplicationCount:    4,
				MonetisationEnabled: true,
			},
			access: &SubscriptionAccess{
				Status:           "active",
				ApplicationLimit: 10,
			},
		},
	)

	err := service.CanApply(
		context.Background(),
		5,
	)

	if err != nil {
		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}
}

func TestCanApply_TrialSubscriptionAllowed(
	t *testing.T,
) {
	service := NewService(
		&MockRepository{
			usage: &UserUsage{
				UserID:              5,
				ApplicationCount:    1,
				MonetisationEnabled: true,
			},
			access: &SubscriptionAccess{
				Status:           "trial",
				ApplicationLimit: 5,
			},
		},
	)

	err := service.CanApply(
		context.Background(),
		5,
	)

	if err != nil {
		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}
}

func TestCanApply_NoSubscriptionFreeLimitReached(
	t *testing.T,
) {
	service := NewService(
		&MockRepository{
			usage: &UserUsage{
				UserID:               5,
				ApplicationCount:     5,
				FreeApplicationLimit: 5,
				MonetisationEnabled:  true,
			},
			access: nil,
		},
	)

	err := service.CanApply(
		context.Background(),
		5,
	)

	if !errors.Is(
		err,
		ErrApplicationLimitReached,
	) {
		t.Fatalf(
			"expected limit reached got %v",
			err,
		)
	}
}

func TestCanApply_NoSubscriptionFreeAccessAllowed(
	t *testing.T,
) {
	service := NewService(
		&MockRepository{
			usage: &UserUsage{
				UserID:               5,
				ApplicationCount:     4,
				FreeApplicationLimit: 5,
				MonetisationEnabled:  true,
			},
		},
	)

	err := service.CanApply(
		context.Background(),
		5,
	)

	if err != nil {
		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}
}

func TestCanPostJob_InvalidUser(t *testing.T) {
	service := NewService(
		&MockRepository{},
	)

	err := service.CanPostJob(
		context.Background(),
		0,
	)

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf(
			"expected ErrInvalidInput got %v",
			err,
		)
	}
}

func TestCanPostJob_CreateUsageError(t *testing.T) {
	createErr := errors.New("create failed")

	service := NewService(
		&MockRepository{
			createErr: createErr,
		},
	)

	err := service.CanPostJob(
		context.Background(),
		5,
	)

	if !errors.Is(err, createErr) {
		t.Fatalf(
			"expected create error got %v",
			err,
		)
	}
}

func TestCanPostJob_GetUsageError(t *testing.T) {
	getErr := errors.New("lookup failed")

	service := NewService(
		&MockRepository{
			getByUserErr: getErr,
		},
	)

	err := service.CanPostJob(
		context.Background(),
		5,
	)

	if !errors.Is(err, getErr) {
		t.Fatalf(
			"expected lookup error got %v",
			err,
		)
	}
}

func TestCanPostJob_MonetisationDisabled(t *testing.T) {
	service := NewService(
		&MockRepository{
			usage: &UserUsage{
				UserID:              5,
				JobPostCount:        999,
				MonetisationEnabled: false,
			},
		},
	)

	err := service.CanPostJob(
		context.Background(),
		5,
	)

	if err != nil {
		t.Fatalf(
			"expected unrestricted access got %v",
			err,
		)
	}
}

func TestCanPostJob_SubscriptionAccessError(
	t *testing.T,
) {
	accessErr := errors.New(
		"subscription lookup failed",
	)

	service := NewService(
		&MockRepository{
			usage: &UserUsage{
				UserID:              5,
				MonetisationEnabled: true,
			},
			getSubscriptionAccessErr: accessErr,
		},
	)

	err := service.CanPostJob(
		context.Background(),
		5,
	)

	if !errors.Is(err, accessErr) {
		t.Fatalf(
			"expected access error got %v",
			err,
		)
	}
}

func TestCanPostJob_ActiveSubscriptionZeroLimit(
	t *testing.T,
) {
	service := NewService(
		&MockRepository{
			usage: &UserUsage{
				UserID:              5,
				MonetisationEnabled: true,
			},
			access: &SubscriptionAccess{
				Status:       "active",
				JobPostLimit: 0,
			},
		},
	)

	err := service.CanPostJob(
		context.Background(),
		5,
	)

	if !errors.Is(
		err,
		ErrJobPostLimitReached,
	) {
		t.Fatalf(
			"expected job limit error got %v",
			err,
		)
	}
}

func TestCanPostJob_ActiveSubscriptionLimitReached(
	t *testing.T,
) {
	service := NewService(
		&MockRepository{
			usage: &UserUsage{
				UserID:              5,
				JobPostCount:        10,
				MonetisationEnabled: true,
			},
			access: &SubscriptionAccess{
				Status:       "active",
				JobPostLimit: 10,
			},
		},
	)

	err := service.CanPostJob(
		context.Background(),
		5,
	)

	if !errors.Is(
		err,
		ErrJobPostLimitReached,
	) {
		t.Fatalf(
			"expected job limit reached got %v",
			err,
		)
	}
}

func TestCanPostJob_ActiveSubscriptionAllowed(
	t *testing.T,
) {
	service := NewService(
		&MockRepository{
			usage: &UserUsage{
				UserID:              5,
				JobPostCount:        3,
				MonetisationEnabled: true,
			},
			access: &SubscriptionAccess{
				Status:       "active",
				JobPostLimit: 10,
			},
		},
	)

	err := service.CanPostJob(
		context.Background(),
		5,
	)

	if err != nil {
		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}
}

func TestCanPostJob_FreeLimitReached(
	t *testing.T,
) {
	service := NewService(
		&MockRepository{
			usage: &UserUsage{
				UserID:              5,
				JobPostCount:        5,
				FreeJobPostLimit:    5,
				MonetisationEnabled: true,
			},
		},
	)

	err := service.CanPostJob(
		context.Background(),
		5,
	)

	if !errors.Is(
		err,
		ErrJobPostLimitReached,
	) {
		t.Fatalf(
			"expected free limit reached got %v",
			err,
		)
	}
}

func TestCanPostJob_FreeAccessAllowed(
	t *testing.T,
) {
	service := NewService(
		&MockRepository{
			usage: &UserUsage{
				UserID:              5,
				JobPostCount:        4,
				FreeJobPostLimit:    5,
				MonetisationEnabled: true,
			},
		},
	)

	err := service.CanPostJob(
		context.Background(),
		5,
	)

	if err != nil {
		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}
}

func TestIncrementApplicationCount_InvalidUser(
	t *testing.T,
) {
	service := NewService(
		&MockRepository{},
	)

	err := service.IncrementApplicationCount(
		context.Background(),
		0,
	)

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf(
			"expected ErrInvalidInput got %v",
			err,
		)
	}
}

func TestIncrementApplicationCount_Success(
	t *testing.T,
) {
	repo := &MockRepository{}
	service := NewService(repo)

	err := service.IncrementApplicationCount(
		context.Background(),
		5,
	)
	if err != nil {
		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}

	if repo.lastApplicationUserID != 5 {
		t.Fatalf(
			"expected user 5 got %d",
			repo.lastApplicationUserID,
		)
	}
}

func TestIncrementApplicationCount_Error(
	t *testing.T,
) {
	incrementErr := errors.New(
		"increment failed",
	)

	service := NewService(
		&MockRepository{
			incrementApplicationErr: incrementErr,
		},
	)

	err := service.IncrementApplicationCount(
		context.Background(),
		5,
	)

	if !errors.Is(err, incrementErr) {
		t.Fatalf(
			"expected increment error got %v",
			err,
		)
	}
}

func TestIncrementJobPostCount_InvalidUser(
	t *testing.T,
) {
	service := NewService(
		&MockRepository{},
	)

	err := service.IncrementJobPostCount(
		context.Background(),
		0,
	)

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf(
			"expected ErrInvalidInput got %v",
			err,
		)
	}
}

func TestIncrementJobPostCount_Success(
	t *testing.T,
) {
	repo := &MockRepository{}
	service := NewService(repo)

	err := service.IncrementJobPostCount(
		context.Background(),
		5,
	)

	if err != nil {
		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}

	if repo.lastJobPostUserID != 5 {
		t.Fatalf(
			"expected user 5 got %d",
			repo.lastJobPostUserID,
		)
	}
}

func TestSetMonetisationEnabled_InvalidUser(
	t *testing.T,
) {
	service := NewService(
		&MockRepository{},
	)

	err := service.SetMonetisationEnabled(
		context.Background(),
		0,
		true,
	)

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf(
			"expected ErrInvalidInput got %v",
			err,
		)
	}
}

func TestSetMonetisationEnabled_Success(
	t *testing.T,
) {
	repo := &MockRepository{}
	service := NewService(repo)

	err := service.SetMonetisationEnabled(
		context.Background(),
		5,
		true,
	)

	if err != nil {
		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}

	if repo.lastMonetisationUserID != 5 {
		t.Fatalf(
			"expected user 5 got %d",
			repo.lastMonetisationUserID,
		)
	}

	if !repo.lastMonetisationEnabled {
		t.Fatal(
			"expected monetisation enabled",
		)
	}
}

func TestIsSubscriptionAllowed(t *testing.T) {
	if !isSubscriptionAllowed("active") {
		t.Fatal(
			"active should be allowed",
		)
	}

	if !isSubscriptionAllowed("trial") {
		t.Fatal(
			"trial should be allowed",
		)
	}

	invalid := []string{
		"past_due",
		"cancelled",
		"expired",
		"",
	}

	for _, status := range invalid {
		if isSubscriptionAllowed(status) {
			t.Fatalf(
				"%q should not be allowed",
				status,
			)
		}
	}
}

var _ Repository = (*MockRepository)(nil)
