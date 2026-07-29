package referrals

import "errors"

var (
	ErrInvalidInput         = errors.New("invalid input")
	ErrReferralNotFound     = errors.New("referral code not found")
	ErrReferralInactive     = errors.New("referral code inactive")
	ErrReferralLimitReached = errors.New("referral usage limit reached")
	ErrCannotReferSelf      = errors.New("cannot redeem your own referral code")
	ErrAlreadyRedeemed      = errors.New("referral coee already redeemed")
	ErrInvalidRewardType    = errors.New("invalid reward type")
	ErrReferralCodeExists   = errors.New("referral code already exists")
)
