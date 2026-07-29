package referrals

import "time"

type ReferralCode struct {
	ID               uint      `json:"id"`
	UserID           uint      `json:"user_id"`
	Code             string    `json:"code"`
	RewardType       string    `json:"reward_type"`
	RewardValuePence int       `json:"reward_value_pence"`
	MaxUses          int       `json:"max_uses"`
	TimesUsed        int       `json:"times_used"`
	IsActive         bool      `json:"is_active"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type ReferralRedemption struct {
	ID             uint      `json:"id"`
	ReferralCodeID uint      `json:"referral_code_id"`
	ReferrerID     uint      `json:"referrer_id"`
	ReferredUserID uint      `json:"referred_user_id"`
	Status         string    `json:"status"`
	RewardApplied  bool      `json:"reward_applied"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
