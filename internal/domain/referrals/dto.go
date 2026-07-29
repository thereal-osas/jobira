package referrals

type CreateReferralCodeRequest struct {
	Code             string `json:"code"`
	RewardType       string `json:"reward_type"`
	RewardValuePence int    `json:"reward_value_pence"`
	MaxUses          int    `json:"max_uses"`
}

type RedeemReferralCodeRequest struct {
	Code string `json:"code"`
}
