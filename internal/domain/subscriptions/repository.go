package subscriptions

import "context"

type Repository interface {
	ListPlans(ctx context.Context) ([]SubscriptionPlan, error)
	GetPlanByID(ctx context.Context, id uint) (*SubscriptionPlan, error)
	CreateOrUpdateUserSubscription(ctx context.Context, userID uint, planID uint, status string) (*UserSubscription, error)
	GetByUserID(ctx context.Context, userID uint) (*UserSubscription, error)
	UpdateStatus(ctx context.Context, userID uint, status string) error
}