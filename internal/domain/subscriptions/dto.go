package subscriptions

type CreateSubscriptionRequest struct {
	PlanID uint	`json:"plan_id"`
}

type UpdateSubscriptionStatusRequest struct {
	Status string `json:"status"`
}

 