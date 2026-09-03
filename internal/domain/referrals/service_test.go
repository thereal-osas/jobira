package referrals

import (
	"context"
	"errors"
	"testing"
	"time"
)

type mockRepository struct {
	createCodeFn                  func(context.Context, *ReferralCode) error
	getCodeByCodeFn               func(context.Context, string) (*ReferralCode, error)
	listCodesByUserIDFn           func(context.Context, uint) ([]ReferralCode, error)
	createRedemptionFn            func(context.Context, *ReferralRedemption) error
	incrementCodeUseFn            func(context.Context, uint) error
	listRedemptionsByReferrerIDFn func(context.Context, uint) ([]ReferralRedemption, error)
}

func (m *mockRepository) CreateCode(
	ctx context.Context,
	code *ReferralCode,
) error {
	if m.createCodeFn != nil {
		return m.createCodeFn(ctx, code)
	}

	return nil
}

func (m *mockRepository) GetCodeByCode(
	ctx context.Context,
	code string,
) (*ReferralCode, error) {
	if m.getCodeByCodeFn != nil {
		return m.getCodeByCodeFn(ctx, code)
	}

	return nil, nil
}

func (m *mockRepository) ListCodesByUserID(
	ctx context.Context,
	userID uint,
) ([]ReferralCode, error) {
	if m.listCodesByUserIDFn != nil {
		return m.listCodesByUserIDFn(ctx, userID)
	}

	return nil, nil
}

func (m *mockRepository) CreateRedemption(
	ctx context.Context,
	redemption *ReferralRedemption,
) error {
	if m.createRedemptionFn != nil {
		return m.createRedemptionFn(ctx, redemption)
	}

	return nil
}

func (m *mockRepository) IncrementCodeUse(
	ctx context.Context,
	referralCodeID uint,
) error {
	if m.incrementCodeUseFn != nil {
		return m.incrementCodeUseFn(ctx, referralCodeID)
	}

	return nil
}

func (m *mockRepository) ListRedemptionsByReferrerID(
	ctx context.Context,
	referrerID uint,
) ([]ReferralRedemption, error) {
	if m.listRedemptionsByReferrerIDFn != nil {
		return m.listRedemptionsByReferrerIDFn(ctx, referrerID)
	}

	return nil, nil
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

func TestIsAllowedRewardType(t *testing.T) {
	tests := []struct {
		name       string
		rewardType string
		expected   bool
	}{
		{
			name:       "credit",
			rewardType: "credit",
			expected:   true,
		},
		{
			name:       "free month",
			rewardType: "free_month",
			expected:   true,
		},
		{
			name:       "invalid",
			rewardType: "cash",
			expected:   false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := isAllowedRewardType(test.rewardType)

			if result != test.expected {
				t.Fatalf(
					"expected %v, got %v",
					test.expected,
					result,
				)
			}
		})
	}
}

func TestService_CreateCode_Success(t *testing.T) {
	createCalled := false

	repo := &mockRepository{
		createCodeFn: func(
			ctx context.Context,
			code *ReferralCode,
		) error {
			createCalled = true

			if code.UserID != 5 {
				t.Fatalf(
					"expected user ID 5, got %d",
					code.UserID,
				)
			}

			if code.Code != "summer25" {
				t.Fatalf(
					"expected normalized code summer25, got %q",
					code.Code,
				)
			}

			if code.RewardType != "free_month" {
				t.Fatalf(
					"expected free_month, got %q",
					code.RewardType,
				)
			}

			if code.RewardValuePence != 2500 {
				t.Fatalf(
					"expected reward value 2500, got %d",
					code.RewardValuePence,
				)
			}

			if code.MaxUses != 10 {
				t.Fatalf(
					"expected max uses 10, got %d",
					code.MaxUses,
				)
			}

			if !code.IsActive {
				t.Fatal("expected code to be active")
			}

			now := time.Now()

			code.ID = 12
			code.CreatedAt = now
			code.UpdatedAt = now

			return nil
		},
	}

	service := NewService(repo)

	code, err := service.CreateCode(
		context.Background(),
		5,
		CreateReferralCodeRequest{
			Code:             "  SUMMER25  ",
			RewardType:       "free_month",
			RewardValuePence: 2500,
			MaxUses:          10,
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if code == nil {
		t.Fatal("expected referral code")
	}

	if code.ID != 12 {
		t.Fatalf("expected ID 12, got %d", code.ID)
	}

	if code.Code != "summer25" {
		t.Fatalf(
			"expected normalized code summer25, got %q",
			code.Code,
		)
	}

	if !createCalled {
		t.Fatal("expected CreateCode repository call")
	}
}

func TestService_CreateCode_DefaultRewardType(t *testing.T) {
	repo := &mockRepository{
		createCodeFn: func(
			ctx context.Context,
			code *ReferralCode,
		) error {
			if code.RewardType != "credit" {
				t.Fatalf(
					"expected default reward type credit, got %q",
					code.RewardType,
				)
			}

			return nil
		},
	}

	service := NewService(repo)

	code, err := service.CreateCode(
		context.Background(),
		5,
		CreateReferralCodeRequest{
			Code: "WELCOME",
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if code == nil {
		t.Fatal("expected referral code")
	}

	if code.RewardType != "credit" {
		t.Fatalf(
			"expected reward type credit, got %q",
			code.RewardType,
		)
	}
}

func TestService_CreateCode_InvalidInput(t *testing.T) {
	service := NewService(&mockRepository{})

	tests := []struct {
		name    string
		userID  uint
		request CreateReferralCodeRequest
	}{
		{
			name:   "zero user ID",
			userID: 0,
			request: CreateReferralCodeRequest{
				Code: "welcome",
			},
		},
		{
			name:   "blank code",
			userID: 5,
			request: CreateReferralCodeRequest{
				Code: "   ",
			},
		},
		{
			name:   "negative reward value",
			userID: 5,
			request: CreateReferralCodeRequest{
				Code:             "welcome",
				RewardType:       "credit",
				RewardValuePence: -1,
			},
		},
		{
			name:   "negative maximum uses",
			userID: 5,
			request: CreateReferralCodeRequest{
				Code:       "welcome",
				RewardType: "credit",
				MaxUses:    -1,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			code, err := service.CreateCode(
				context.Background(),
				test.userID,
				test.request,
			)

			if code != nil {
				t.Fatalf(
					"expected nil code, got %+v",
					code,
				)
			}

			if !errors.Is(err, ErrInvalidInput) {
				t.Fatalf(
					"expected ErrInvalidInput, got %v",
					err,
				)
			}
		})
	}
}

func TestService_CreateCode_InvalidRewardType(t *testing.T) {
	service := NewService(&mockRepository{})

	code, err := service.CreateCode(
		context.Background(),
		5,
		CreateReferralCodeRequest{
			Code:       "welcome",
			RewardType: "cash",
		},
	)

	if code != nil {
		t.Fatalf(
			"expected nil code, got %+v",
			code,
		)
	}

	if !errors.Is(err, ErrInvalidRewardType) {
		t.Fatalf(
			"expected ErrInvalidRewardType, got %v",
			err,
		)
	}
}

func TestService_CreateCode_RepositoryError(t *testing.T) {
	expectedErr := errors.New("create code failed")

	repo := &mockRepository{
		createCodeFn: func(
			context.Context,
			*ReferralCode,
		) error {
			return expectedErr
		},
	}

	service := NewService(repo)

	code, err := service.CreateCode(
		context.Background(),
		5,
		CreateReferralCodeRequest{
			Code:       "welcome",
			RewardType: "credit",
		},
	)

	if code != nil {
		t.Fatalf(
			"expected nil code, got %+v",
			code,
		)
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected %v, got %v",
			expectedErr,
			err,
		)
	}
}

func TestService_ListMyCodes_Success(t *testing.T) {
	now := time.Now()

	repo := &mockRepository{
		listCodesByUserIDFn: func(
			ctx context.Context,
			userID uint,
		) ([]ReferralCode, error) {
			if userID != 5 {
				t.Fatalf(
					"expected user ID 5, got %d",
					userID,
				)
			}

			return []ReferralCode{
				{
					ID:         1,
					UserID:     5,
					Code:       "welcome",
					RewardType: "credit",
					IsActive:   true,
					CreatedAt:  now,
					UpdatedAt:  now,
				},
				{
					ID:         2,
					UserID:     5,
					Code:       "summer",
					RewardType: "free_month",
					IsActive:   true,
					CreatedAt:  now.Add(-time.Hour),
					UpdatedAt:  now.Add(-time.Hour),
				},
			}, nil
		},
	}

	service := NewService(repo)

	codes, err := service.ListMyCodes(
		context.Background(),
		5,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(codes) != 2 {
		t.Fatalf(
			"expected 2 codes, got %d",
			len(codes),
		)
	}

	if codes[0].Code != "welcome" {
		t.Fatalf(
			"expected first code welcome, got %q",
			codes[0].Code,
		)
	}
}

func TestService_ListMyCodes_Empty(t *testing.T) {
	repo := &mockRepository{
		listCodesByUserIDFn: func(
			context.Context,
			uint,
		) ([]ReferralCode, error) {
			return []ReferralCode{}, nil
		},
	}

	service := NewService(repo)

	codes, err := service.ListMyCodes(
		context.Background(),
		5,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(codes) != 0 {
		t.Fatalf(
			"expected no codes, got %d",
			len(codes),
		)
	}
}

func TestService_ListMyCodes_InvalidInput(t *testing.T) {
	service := NewService(&mockRepository{})

	codes, err := service.ListMyCodes(
		context.Background(),
		0,
	)

	if codes != nil {
		t.Fatalf(
			"expected nil codes, got %+v",
			codes,
		)
	}

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf(
			"expected ErrInvalidInput, got %v",
			err,
		)
	}
}

func TestService_ListMyCodes_RepositoryError(t *testing.T) {
	expectedErr := errors.New("list codes failed")

	repo := &mockRepository{
		listCodesByUserIDFn: func(
			context.Context,
			uint,
		) ([]ReferralCode, error) {
			return nil, expectedErr
		},
	}

	service := NewService(repo)

	codes, err := service.ListMyCodes(
		context.Background(),
		5,
	)

	if codes != nil {
		t.Fatalf(
			"expected nil codes, got %+v",
			codes,
		)
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected %v, got %v",
			expectedErr,
			err,
		)
	}
}

func TestService_Redeem_Success(t *testing.T) {
	createCalled := false
	incrementCalled := false

	repo := &mockRepository{
		getCodeByCodeFn: func(
			ctx context.Context,
			code string,
		) (*ReferralCode, error) {
			if code != "welcome" {
				t.Fatalf(
					"expected normalized code welcome, got %q",
					code,
				)
			}

			return &ReferralCode{
				ID:         12,
				UserID:     5,
				Code:       "welcome",
				MaxUses:    10,
				TimesUsed:  2,
				IsActive:   true,
				RewardType: "credit",
			}, nil
		},
		createRedemptionFn: func(
			ctx context.Context,
			redemption *ReferralRedemption,
		) error {
			createCalled = true

			if redemption.ReferralCodeID != 12 {
				t.Fatalf(
					"expected referral code ID 12, got %d",
					redemption.ReferralCodeID,
				)
			}

			if redemption.ReferrerID != 5 {
				t.Fatalf(
					"expected referrer ID 5, got %d",
					redemption.ReferrerID,
				)
			}

			if redemption.ReferredUserID != 8 {
				t.Fatalf(
					"expected referred user ID 8, got %d",
					redemption.ReferredUserID,
				)
			}

			if redemption.Status != "pending" {
				t.Fatalf(
					"expected status pending, got %q",
					redemption.Status,
				)
			}

			if redemption.RewardApplied {
				t.Fatal("expected reward_applied false")
			}

			redemption.ID = 20
			redemption.CreatedAt = time.Now()
			redemption.UpdatedAt = time.Now()

			return nil
		},
		incrementCodeUseFn: func(
			ctx context.Context,
			referralCodeID uint,
		) error {
			incrementCalled = true

			if referralCodeID != 12 {
				t.Fatalf(
					"expected referral code ID 12, got %d",
					referralCodeID,
				)
			}

			return nil
		},
	}

	service := NewService(repo)

	redemption, err := service.Redeem(
		context.Background(),
		8,
		RedeemReferralCodeRequest{
			Code: "  WELCOME  ",
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if redemption == nil {
		t.Fatal("expected redemption")
	}

	if redemption.ID != 20 {
		t.Fatalf(
			"expected redemption ID 20, got %d",
			redemption.ID,
		)
	}

	if !createCalled {
		t.Fatal("expected CreateRedemption repository call")
	}

	if !incrementCalled {
		t.Fatal("expected IncrementCodeUse repository call")
	}
}

func TestService_Redeem_UnlimitedCode(t *testing.T) {
	repo := &mockRepository{
		getCodeByCodeFn: func(
			context.Context,
			string,
		) (*ReferralCode, error) {
			return &ReferralCode{
				ID:        12,
				UserID:    5,
				MaxUses:   0,
				TimesUsed: 500,
				IsActive:  true,
			}, nil
		},
	}

	service := NewService(repo)

	redemption, err := service.Redeem(
		context.Background(),
		8,
		RedeemReferralCodeRequest{
			Code: "welcome",
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if redemption == nil {
		t.Fatal("expected redemption")
	}
}

func TestService_Redeem_InvalidInput(t *testing.T) {
	service := NewService(&mockRepository{})

	tests := []struct {
		name   string
		userID uint
		code   string
	}{
		{
			name:   "zero user ID",
			userID: 0,
			code:   "welcome",
		},
		{
			name:   "blank code",
			userID: 8,
			code:   "   ",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			redemption, err := service.Redeem(
				context.Background(),
				test.userID,
				RedeemReferralCodeRequest{
					Code: test.code,
				},
			)

			if redemption != nil {
				t.Fatalf(
					"expected nil redemption, got %+v",
					redemption,
				)
			}

			if !errors.Is(err, ErrInvalidInput) {
				t.Fatalf(
					"expected ErrInvalidInput, got %v",
					err,
				)
			}
		})
	}
}

func TestService_Redeem_GetCodeError(t *testing.T) {
	expectedErr := errors.New("get code failed")

	repo := &mockRepository{
		getCodeByCodeFn: func(
			context.Context,
			string,
		) (*ReferralCode, error) {
			return nil, expectedErr
		},
	}

	service := NewService(repo)

	redemption, err := service.Redeem(
		context.Background(),
		8,
		RedeemReferralCodeRequest{
			Code: "welcome",
		},
	)

	if redemption != nil {
		t.Fatalf(
			"expected nil redemption, got %+v",
			redemption,
		)
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected %v, got %v",
			expectedErr,
			err,
		)
	}
}

func TestService_Redeem_Inactive(t *testing.T) {
	repo := &mockRepository{
		getCodeByCodeFn: func(
			context.Context,
			string,
		) (*ReferralCode, error) {
			return &ReferralCode{
				ID:       12,
				UserID:   5,
				IsActive: false,
			}, nil
		},
	}

	service := NewService(repo)

	redemption, err := service.Redeem(
		context.Background(),
		8,
		RedeemReferralCodeRequest{
			Code: "welcome",
		},
	)

	if redemption != nil {
		t.Fatalf(
			"expected nil redemption, got %+v",
			redemption,
		)
	}

	if !errors.Is(err, ErrReferralInactive) {
		t.Fatalf(
			"expected ErrReferralInactive, got %v",
			err,
		)
	}
}

func TestService_Redeem_CannotReferSelf(t *testing.T) {
	repo := &mockRepository{
		getCodeByCodeFn: func(
			context.Context,
			string,
		) (*ReferralCode, error) {
			return &ReferralCode{
				ID:       12,
				UserID:   8,
				IsActive: true,
			}, nil
		},
	}

	service := NewService(repo)

	redemption, err := service.Redeem(
		context.Background(),
		8,
		RedeemReferralCodeRequest{
			Code: "welcome",
		},
	)

	if redemption != nil {
		t.Fatalf(
			"expected nil redemption, got %+v",
			redemption,
		)
	}

	if !errors.Is(err, ErrCannotReferSelf) {
		t.Fatalf(
			"expected ErrCannotReferSelf, got %v",
			err,
		)
	}
}

func TestService_Redeem_LimitReached(t *testing.T) {
	repo := &mockRepository{
		getCodeByCodeFn: func(
			context.Context,
			string,
		) (*ReferralCode, error) {
			return &ReferralCode{
				ID:        12,
				UserID:    5,
				MaxUses:   3,
				TimesUsed: 3,
				IsActive:  true,
			}, nil
		},
	}

	service := NewService(repo)

	redemption, err := service.Redeem(
		context.Background(),
		8,
		RedeemReferralCodeRequest{
			Code: "welcome",
		},
	)

	if redemption != nil {
		t.Fatalf(
			"expected nil redemption, got %+v",
			redemption,
		)
	}

	if !errors.Is(err, ErrReferralLimitReached) {
		t.Fatalf(
			"expected ErrReferralLimitReached, got %v",
			err,
		)
	}
}

func TestService_Redeem_CreateRedemptionError(t *testing.T) {
	expectedErr := errors.New("create redemption failed")

	repo := &mockRepository{
		getCodeByCodeFn: func(
			context.Context,
			string,
		) (*ReferralCode, error) {
			return &ReferralCode{
				ID:       12,
				UserID:   5,
				IsActive: true,
			}, nil
		},
		createRedemptionFn: func(
			context.Context,
			*ReferralRedemption,
		) error {
			return expectedErr
		},
	}

	service := NewService(repo)

	redemption, err := service.Redeem(
		context.Background(),
		8,
		RedeemReferralCodeRequest{
			Code: "welcome",
		},
	)

	if redemption != nil {
		t.Fatalf(
			"expected nil redemption, got %+v",
			redemption,
		)
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected %v, got %v",
			expectedErr,
			err,
		)
	}
}

func TestService_Redeem_IncrementError(t *testing.T) {
	expectedErr := errors.New("increment failed")

	repo := &mockRepository{
		getCodeByCodeFn: func(
			context.Context,
			string,
		) (*ReferralCode, error) {
			return &ReferralCode{
				ID:       12,
				UserID:   5,
				IsActive: true,
			}, nil
		},
		incrementCodeUseFn: func(
			context.Context,
			uint,
		) error {
			return expectedErr
		},
	}

	service := NewService(repo)

	redemption, err := service.Redeem(
		context.Background(),
		8,
		RedeemReferralCodeRequest{
			Code: "welcome",
		},
	)

	if redemption != nil {
		t.Fatalf(
			"expected nil redemption, got %+v",
			redemption,
		)
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected %v, got %v",
			expectedErr,
			err,
		)
	}
}

func TestService_ListMyRedemption_Success(t *testing.T) {
	now := time.Now()

	repo := &mockRepository{
		listRedemptionsByReferrerIDFn: func(
			ctx context.Context,
			referrerID uint,
		) ([]ReferralRedemption, error) {
			if referrerID != 5 {
				t.Fatalf(
					"expected referrer ID 5, got %d",
					referrerID,
				)
			}

			return []ReferralRedemption{
				{
					ID:             1,
					ReferralCodeID: 12,
					ReferrerID:     5,
					ReferredUserID: 8,
					Status:         "pending",
					CreatedAt:      now,
					UpdatedAt:      now,
				},
			}, nil
		},
	}

	service := NewService(repo)

	redemptions, err := service.ListMyRedemption(
		context.Background(),
		5,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(redemptions) != 1 {
		t.Fatalf(
			"expected 1 redemption, got %d",
			len(redemptions),
		)
	}

	if redemptions[0].ReferredUserID != 8 {
		t.Fatalf(
			"expected referred user ID 8, got %d",
			redemptions[0].ReferredUserID,
		)
	}
}

func TestService_ListMyRedemption_Empty(t *testing.T) {
	repo := &mockRepository{
		listRedemptionsByReferrerIDFn: func(
			context.Context,
			uint,
		) ([]ReferralRedemption, error) {
			return []ReferralRedemption{}, nil
		},
	}

	service := NewService(repo)

	redemptions, err := service.ListMyRedemption(
		context.Background(),
		5,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(redemptions) != 0 {
		t.Fatalf(
			"expected no redemptions, got %d",
			len(redemptions),
		)
	}
}

func TestService_ListMyRedemption_InvalidInput(t *testing.T) {
	service := NewService(&mockRepository{})

	redemptions, err := service.ListMyRedemption(
		context.Background(),
		0,
	)

	if redemptions != nil {
		t.Fatalf(
			"expected nil redemptions, got %+v",
			redemptions,
		)
	}

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf(
			"expected ErrInvalidInput, got %v",
			err,
		)
	}
}

func TestService_ListMyRedemption_RepositoryError(t *testing.T) {
	expectedErr := errors.New("list redemptions failed")

	repo := &mockRepository{
		listRedemptionsByReferrerIDFn: func(
			context.Context,
			uint,
		) ([]ReferralRedemption, error) {
			return nil, expectedErr
		},
	}

	service := NewService(repo)

	redemptions, err := service.ListMyRedemption(
		context.Background(),
		5,
	)

	if redemptions != nil {
		t.Fatalf(
			"expected nil redemptions, got %+v",
			redemptions,
		)
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected %v, got %v",
			expectedErr,
			err,
		)
	}
}
