package referrals

import "context"

type Repository interface {
	CreateCode(ctx context.Context, code *ReferralCode) error
	GetCodeByCode(ctx context.Context, code string) (*ReferralCode, error)
	ListCodesByUserID(ctx context.Context, userID uint) ([]ReferralCode, error)
	CreateRedemption(ctx context.Context, redemption *ReferralRedemption) error
	IncrementCodeUse(ctx context.Context, referralCodeID uint) error
	ListRedemptionsByReferrerID(ctx context.Context, referrerID uint) ([]ReferralRedemption, error)
}
