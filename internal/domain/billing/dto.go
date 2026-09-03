package billing

type CreateCheckoutSessionRequest struct {
	PlanID    uint   `json:"plan_id"`
	PromoCode string `json:"promo_code"`
}

type CheckoutSessionResponse struct {
	URL       string `json:"url"`
	SessionID string `json:"session_id"`
}

type BillingPortalResponse struct {
	URL string `json:"url"`
}

type CreatePromoCodeRequest struct {
	Code             string `json:"code"`
	Description      string `json:"description"`
	DiscountType     string `json:"discount_type"`
	PercentageOff    int    `json:"percentage_off"`
	FixedAmountPence int    `json:"fixed_amount_pence"`
	FreeMonths       int    `json:"free_months"`
	MaxUses          int    `json:"max_uses"`
	UserType         string `json:"user_type"`
	PlanID           *uint  `json:"plan_id"`
}

type ValidatePromoCodeRequest struct {
	Code   string `json:"code"`
	PlanID uint   `json:"plan_id"`
}

type ValidatePromoCodeRespone struct {
	Code             string `json:"code"`
	Valid            bool   `json:"valid"`
	Description      string `json:"description"`
	DiscountType     string `json:"discount_type"`
	PercentageOff    int    `json:"percentage_off"`
	FixedAmountPence int    `json:"fixed_amount_pence"`
	FreeMonths       int    `json:"free_months"`
}
