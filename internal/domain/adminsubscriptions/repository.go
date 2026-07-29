package adminsubscriptions

import "context"

type Repository interface {
	ListPlans(ctx context.Context) ([]AdminSubscriptionPlan, error) 
	CreatePlan(ctx context.Context, plan *AdminSubscriptionPlan) error
	UpdatePlan(ctx context.Context, plan *AdminSubscriptionPlan) error
	DisablePlan(ctx context.Context, planID uint) error

	ListUserSubscriptions(ctx context.Context) ([]AdminUserSubscription, error)
	GetUserSubscriptions(ctx context.Context, userID uint) (*AdminUserSubscription, error)
	CreateOrUpdateUserSubscription(ctx context.Context, uerID uint, planID uint, status string) error
}