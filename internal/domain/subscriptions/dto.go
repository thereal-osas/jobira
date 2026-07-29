package subscriptions

type CreateSubscriptionRequest struct {
	PlanID uint	`json:"plan_id"`
}

type UpdateSubcriptionStatusRequest struct {
	Status string `json:"status"`
}

 