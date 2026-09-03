package referrals

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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

func TestHandler_CreateCode_Success(t *testing.T) {
	repo := &mockRepository{
		createCodeFn: func(
			ctx context.Context,
			code *ReferralCode,
		) error {
			if code.UserID != 5 {
				t.Fatalf(
					"expected user ID 5, got %d",
					code.UserID,
				)
			}

			if code.Code != "welcome" {
				t.Fatalf(
					"expected normalized code welcome, got %q",
					code.Code,
				)
			}

			if code.RewardType != "credit" {
				t.Fatalf(
					"expected reward type credit, got %q",
					code.RewardType,
				)
			}

			code.ID = 12

			return nil
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodPost,
		"/referrals/codes",
		strings.NewReader(`{
			"code":"WELCOME",
			"reward_type":"credit",
			"reward_value_pence":1500,
			"max_uses":10
		}`),
	)
	req = requestWithUser(
		req,
		5,
		"user",
	)

	recorder := httptest.NewRecorder()

	handler.CreateCode(recorder, req)

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
		`"code":"welcome"`,
	) {
		t.Fatalf(
			"expected referral code in response: %s",
			recorder.Body.String(),
		)
	}

	if !strings.Contains(
		recorder.Body.String(),
		`"user_id":5`,
	) {
		t.Fatalf(
			"expected user ID in response: %s",
			recorder.Body.String(),
		)
	}
}

func TestHandler_CreateCode_Unauthorized(t *testing.T) {
	handler := newHandlerForTest(
		&mockRepository{},
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/referrals/codes",
		strings.NewReader(`{}`),
	)

	recorder := httptest.NewRecorder()

	handler.CreateCode(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusUnauthorized,
			recorder.Code,
		)
	}
}

func TestHandler_CreateCode_InvalidJSON(t *testing.T) {
	handler := newHandlerForTest(
		&mockRepository{},
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/referrals/codes",
		strings.NewReader(`{"code":`),
	)
	req = requestWithUser(
		req,
		5,
		"user",
	)

	recorder := httptest.NewRecorder()

	handler.CreateCode(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusBadRequest,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_CreateCode_InvalidInput(t *testing.T) {
	handler := newHandlerForTest(
		&mockRepository{},
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/referrals/codes",
		strings.NewReader(`{
			"code":""
		}`),
	)
	req = requestWithUser(
		req,
		5,
		"user",
	)

	recorder := httptest.NewRecorder()

	handler.CreateCode(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusBadRequest,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_CreateCode_ZeroUserID(t *testing.T) {
	handler := newHandlerForTest(
		&mockRepository{},
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/referrals/codes",
		strings.NewReader(`{
			"code":"welcome"
		}`),
	)
	req = requestWithUser(
		req,
		0,
		"user",
	)

	recorder := httptest.NewRecorder()

	handler.CreateCode(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusBadRequest,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_CreateCode_InvalidRewardType(t *testing.T) {
	handler := newHandlerForTest(
		&mockRepository{},
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/referrals/codes",
		strings.NewReader(`{
			"code":"welcome",
			"reward_type":"cash"
		}`),
	)
	req = requestWithUser(
		req,
		5,
		"user",
	)

	recorder := httptest.NewRecorder()

	handler.CreateCode(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusBadRequest,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_CreateCode_AlreadyExists(t *testing.T) {
	repo := &mockRepository{
		createCodeFn: func(
			context.Context,
			*ReferralCode,
		) error {
			return ErrReferralCodeExists
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodPost,
		"/referrals/codes",
		strings.NewReader(`{
			"code":"welcome",
			"reward_type":"credit"
		}`),
	)
	req = requestWithUser(
		req,
		5,
		"user",
	)

	recorder := httptest.NewRecorder()

	handler.CreateCode(recorder, req)

	if recorder.Code != http.StatusConflict {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusConflict,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_CreateCode_InternalServerError(t *testing.T) {
	expectedErr := errors.New("create code failed")

	repo := &mockRepository{
		createCodeFn: func(
			context.Context,
			*ReferralCode,
		) error {
			return expectedErr
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodPost,
		"/referrals/codes",
		strings.NewReader(`{
			"code":"welcome",
			"reward_type":"credit"
		}`),
	)
	req = requestWithUser(
		req,
		5,
		"user",
	)

	recorder := httptest.NewRecorder()

	handler.CreateCode(recorder, req)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusInternalServerError,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_ListMyCodes_Success(t *testing.T) {
	repo := &mockRepository{
		listCodesByUserIDFn: func(
			context.Context,
			uint,
		) ([]ReferralCode, error) {
			return []ReferralCode{
				{
					ID:         12,
					UserID:     5,
					Code:       "welcome",
					RewardType: "credit",
					IsActive:   true,
				},
			}, nil
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodGet,
		"/referrals/codes/me",
		nil,
	)
	req = requestWithUser(
		req,
		5,
		"user",
	)

	recorder := httptest.NewRecorder()

	handler.ListMyCodes(recorder, req)

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
		`"code":"welcome"`,
	) {
		t.Fatalf(
			"expected referral code in response: %s",
			recorder.Body.String(),
		)
	}
}

func TestHandler_ListMyCodes_Empty(t *testing.T) {
	repo := &mockRepository{
		listCodesByUserIDFn: func(
			context.Context,
			uint,
		) ([]ReferralCode, error) {
			return []ReferralCode{}, nil
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodGet,
		"/referrals/codes/me",
		nil,
	)
	req = requestWithUser(
		req,
		5,
		"user",
	)

	recorder := httptest.NewRecorder()

	handler.ListMyCodes(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusOK,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_ListMyCodes_Unauthorized(t *testing.T) {
	handler := newHandlerForTest(
		&mockRepository{},
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/referrals/codes/me",
		nil,
	)

	recorder := httptest.NewRecorder()

	handler.ListMyCodes(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusUnauthorized,
			recorder.Code,
		)
	}
}

func TestHandler_ListMyCodes_InvalidInput(t *testing.T) {
	handler := newHandlerForTest(
		&mockRepository{},
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/referrals/codes/me",
		nil,
	)
	req = requestWithUser(
		req,
		0,
		"user",
	)

	recorder := httptest.NewRecorder()

	handler.ListMyCodes(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusBadRequest,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_ListMyCodes_InternalServerError(t *testing.T) {
	expectedErr := errors.New("list codes failed")

	repo := &mockRepository{
		listCodesByUserIDFn: func(
			context.Context,
			uint,
		) ([]ReferralCode, error) {
			return nil, expectedErr
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodGet,
		"/referrals/codes/me",
		nil,
	)
	req = requestWithUser(
		req,
		5,
		"user",
	)

	recorder := httptest.NewRecorder()

	handler.ListMyCodes(recorder, req)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusInternalServerError,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Redeem_Success(t *testing.T) {
	repo := &mockRepository{
		getCodeByCodeFn: func(
			context.Context,
			string,
		) (*ReferralCode, error) {
			return &ReferralCode{
				ID:       12,
				UserID:   5,
				Code:     "welcome",
				IsActive: true,
			}, nil
		},
		createRedemptionFn: func(
			ctx context.Context,
			redemption *ReferralRedemption,
		) error {
			redemption.ID = 20
			return nil
		},
		incrementCodeUseFn: func(
			context.Context,
			uint,
		) error {
			return nil
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodPost,
		"/referrals/redeem",
		strings.NewReader(`{
			"code":"WELCOME"
		}`),
	)
	req = requestWithUser(
		req,
		8,
		"user",
	)

	recorder := httptest.NewRecorder()

	handler.Redeem(recorder, req)

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
		`"referral_code_id":12`,
	) {
		t.Fatalf(
			"expected referral code ID in response: %s",
			recorder.Body.String(),
		)
	}

	if !strings.Contains(
		recorder.Body.String(),
		`"referred_user_id":8`,
	) {
		t.Fatalf(
			"expected referred user ID in response: %s",
			recorder.Body.String(),
		)
	}
}

func TestHandler_Redeem_Unauthorized(t *testing.T) {
	handler := newHandlerForTest(
		&mockRepository{},
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/referrals/redeem",
		strings.NewReader(`{}`),
	)

	recorder := httptest.NewRecorder()

	handler.Redeem(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusUnauthorized,
			recorder.Code,
		)
	}
}

func TestHandler_Redeem_InvalidJSON(t *testing.T) {
	handler := newHandlerForTest(
		&mockRepository{},
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/referrals/redeem",
		strings.NewReader(`{"code":`),
	)
	req = requestWithUser(
		req,
		8,
		"user",
	)

	recorder := httptest.NewRecorder()

	handler.Redeem(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusBadRequest,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Redeem_InvalidInput(t *testing.T) {
	handler := newHandlerForTest(
		&mockRepository{},
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/referrals/redeem",
		strings.NewReader(`{
			"code":""
		}`),
	)
	req = requestWithUser(
		req,
		8,
		"user",
	)

	recorder := httptest.NewRecorder()

	handler.Redeem(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusBadRequest,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Redeem_NotFound(t *testing.T) {
	repo := &mockRepository{
		getCodeByCodeFn: func(
			context.Context,
			string,
		) (*ReferralCode, error) {
			return nil, ErrReferralNotFound
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodPost,
		"/referrals/redeem",
		strings.NewReader(`{
			"code":"missing"
		}`),
	)
	req = requestWithUser(
		req,
		8,
		"user",
	)

	recorder := httptest.NewRecorder()

	handler.Redeem(recorder, req)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusNotFound,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Redeem_Inactive(t *testing.T) {
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

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodPost,
		"/referrals/redeem",
		strings.NewReader(`{
			"code":"welcome"
		}`),
	)
	req = requestWithUser(
		req,
		8,
		"user",
	)

	recorder := httptest.NewRecorder()

	handler.Redeem(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusBadRequest,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Redeem_LimitReached(t *testing.T) {
	repo := &mockRepository{
		getCodeByCodeFn: func(
			context.Context,
			string,
		) (*ReferralCode, error) {
			return &ReferralCode{
				ID:        12,
				UserID:    5,
				IsActive:  true,
				MaxUses:   2,
				TimesUsed: 2,
			}, nil
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodPost,
		"/referrals/redeem",
		strings.NewReader(`{
			"code":"welcome"
		}`),
	)
	req = requestWithUser(
		req,
		8,
		"user",
	)

	recorder := httptest.NewRecorder()

	handler.Redeem(recorder, req)

	if recorder.Code != http.StatusConflict {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusConflict,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Redeem_CannotReferSelf(t *testing.T) {
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

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodPost,
		"/referrals/redeem",
		strings.NewReader(`{
			"code":"welcome"
		}`),
	)
	req = requestWithUser(
		req,
		8,
		"user",
	)

	recorder := httptest.NewRecorder()

	handler.Redeem(recorder, req)

	if recorder.Code != http.StatusConflict {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusConflict,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Redeem_AlreadyRedeemed(t *testing.T) {
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
			return ErrAlreadyRedeemed
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodPost,
		"/referrals/redeem",
		strings.NewReader(`{
			"code":"welcome"
		}`),
	)
	req = requestWithUser(
		req,
		8,
		"user",
	)

	recorder := httptest.NewRecorder()

	handler.Redeem(recorder, req)

	if recorder.Code != http.StatusConflict {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusConflict,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Redeem_InternalServerError(t *testing.T) {
	expectedErr := errors.New("get code failed")

	repo := &mockRepository{
		getCodeByCodeFn: func(
			context.Context,
			string,
		) (*ReferralCode, error) {
			return nil, expectedErr
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodPost,
		"/referrals/redeem",
		strings.NewReader(`{
			"code":"welcome"
		}`),
	)
	req = requestWithUser(
		req,
		8,
		"user",
	)

	recorder := httptest.NewRecorder()

	handler.Redeem(recorder, req)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusInternalServerError,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_ListMyRedemptions_Success(t *testing.T) {
	repo := &mockRepository{
		listRedemptionsByReferrerIDFn: func(
			context.Context,
			uint,
		) ([]ReferralRedemption, error) {
			return []ReferralRedemption{
				{
					ID:             1,
					ReferralCodeID: 12,
					ReferrerID:     5,
					ReferredUserID: 8,
					Status:         "pending",
				},
			}, nil
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodGet,
		"/referrals/redemptions/me",
		nil,
	)
	req = requestWithUser(
		req,
		5,
		"user",
	)

	recorder := httptest.NewRecorder()

	handler.ListMyRedemptions(recorder, req)

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
		`"referred_user_id":8`,
	) {
		t.Fatalf(
			"expected redemption in response: %s",
			recorder.Body.String(),
		)
	}
}

func TestHandler_ListMyRedemptions_Empty(t *testing.T) {
	repo := &mockRepository{
		listRedemptionsByReferrerIDFn: func(
			context.Context,
			uint,
		) ([]ReferralRedemption, error) {
			return []ReferralRedemption{}, nil
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodGet,
		"/referrals/redemptions/me",
		nil,
	)
	req = requestWithUser(
		req,
		5,
		"user",
	)

	recorder := httptest.NewRecorder()

	handler.ListMyRedemptions(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusOK,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_ListMyRedemptions_Unauthorized(t *testing.T) {
	handler := newHandlerForTest(
		&mockRepository{},
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/referrals/redemptions/me",
		nil,
	)

	recorder := httptest.NewRecorder()

	handler.ListMyRedemptions(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusUnauthorized,
			recorder.Code,
		)
	}
}

func TestHandler_ListMyRedemptions_InvalidInput(t *testing.T) {
	handler := newHandlerForTest(
		&mockRepository{},
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/referrals/redemptions/me",
		nil,
	)
	req = requestWithUser(
		req,
		0,
		"user",
	)

	recorder := httptest.NewRecorder()

	handler.ListMyRedemptions(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusBadRequest,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_ListMyRedemptions_InternalServerError(t *testing.T) {
	expectedErr := errors.New("list redemptions failed")

	repo := &mockRepository{
		listRedemptionsByReferrerIDFn: func(
			context.Context,
			uint,
		) ([]ReferralRedemption, error) {
			return nil, expectedErr
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodGet,
		"/referrals/redemptions/me",
		nil,
	)
	req = requestWithUser(
		req,
		5,
		"user",
	)

	recorder := httptest.NewRecorder()

	handler.ListMyRedemptions(recorder, req)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusInternalServerError,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}
