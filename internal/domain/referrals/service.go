package referrals

import (
	"context"
	"strings"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateCode(ctx context.Context, userID uint, req CreateReferralCodeRequest) (*ReferralCode, error) {
	req.Code = strings.ToLower(strings.TrimSpace(req.Code))
	req.RewardType = strings.TrimSpace(req.RewardType)

	if userID == 0 || req.Code == "" {
		return nil, ErrInvalidInput
	}

	if req.RewardType == "" {
		req.RewardType = "credit"
	}

	if !isAllowedRewardType(req.RewardType) {
		return nil, ErrInvalidRewardType
	}

	if req.RewardValuePence < 0 || req.MaxUses < 0 {
		return nil, ErrInvalidInput
	}

	code := &ReferralCode{
		UserID:           userID,
		Code:             req.Code,
		RewardType:       req.RewardType,
		RewardValuePence: req.RewardValuePence,
		MaxUses:          req.MaxUses,
		IsActive:         true,
	}

	if err := s.repo.CreateCode(ctx, code); err != nil {
		return nil, err
	}

	return code, nil
}

func (s *Service) ListMyCodes(ctx context.Context, userID uint) ([]ReferralCode, error) {
	if userID == 0 {
		return nil, ErrInvalidInput
	}

	return s.repo.ListCodesByUserID(ctx, userID)
}

func (s *Service) Redeem(ctx context.Context, userID uint, req RedeemReferralCodeRequest) (*ReferralRedemption, error) {
	req.Code = strings.ToLower(strings.TrimSpace(req.Code))

	if userID == 0 || req.Code == "" {
		return nil, ErrInvalidInput
	}

	code, err := s.repo.GetCodeByCode(ctx, req.Code)
	if err != nil {
		return nil, err
	}

	if !code.IsActive {
		return nil, ErrReferralInactive
	}

	if code.UserID == userID {
		return nil, ErrCannotReferSelf
	}

	if code.MaxUses > 0 && code.TimesUsed >= code.MaxUses {
		return nil, ErrReferralLimitReached
	}

	redemption := &ReferralRedemption{
		ReferralCodeID: code.ID,
		ReferrerID:     code.UserID,
		ReferredUserID: userID,
		Status:         "pending",
		RewardApplied:  false,
	}

	if err := s.repo.CreateRedemption(ctx, redemption); err != nil {
		return nil, err
	}

	if err := s.repo.IncrementCodeUse(ctx, code.ID); err != nil {
		return nil, err
	}

	return redemption, nil
}

func (s *Service) ListMyRedemption(ctx context.Context, userID uint) ([]ReferralRedemption, error) {
	if userID == 0 {
		return nil, ErrInvalidInput
	}

	return s.repo.ListRedemptionsByReferrerID(ctx, userID)
}

func isAllowedRewardType(RewardType string) bool {
	if RewardType == "credit" {
		return true
	}

	if RewardType == "free_month" {
		return true
	}

	return false
}
