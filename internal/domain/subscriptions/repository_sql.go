package subscriptions

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type SQLRepository struct {
	db *sql.DB
}

func NewSQLRepository(db *sql.DB) *SQLRepository {
	return &SQLRepository{db: db}
}

func (r *SQLRepository) ListPlans(ctx context.Context) ([]SubscriptionPlan, error) {
	query := `
		SELECT 
			id, 
			name, 
			role_type, 
			price_pence, 
			billing_interval, 
			application_limit, 
			job_post_limit, 
			cleaner_seat_limit,
			is_active,
			created_at,
			updated_at 
		FROM subscription_plans 
		WHERE is_active = true 
		ORDER BY price_pence ASC 	
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var plans []SubscriptionPlan

	for rows.Next() {
		var plan SubscriptionPlan

		err := rows.Scan(
			&plan.ID,
			&plan.Name,
			&plan.RoleType,
			&plan.PricePence,
			&plan.BillingInterval,
			&plan.ApplicationLimit, 
			&plan.JobPostLimit,
			&plan.CleanerSeatLimit,
			&plan.IsActive, 
			&plan.CreatedAt,
			&plan.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		
		plans = append(plans, plan)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return plans, nil
	
}
func (r *SQLRepository)	GetPlanByID(ctx context.Context, id uint) (*SubscriptionPlan, error) {
	query := `
		SELECT 
			id, 
			name, 
			role_type, 
			price_pence, 
			billing_interval, 
			application_limit, 
			job_post_limit, 
			cleaner_seat_limit,
			is_active,
			created_at,
			updated_at 
		FROM subscription_plans 
		WHERE id = $1  
		LIMIT 1 	
	`

	var plan SubscriptionPlan

		err := r.db.QueryRowContext(ctx, query, id).Scan(
			&plan.ID,
			&plan.Name,
			&plan.RoleType,
			&plan.PricePence,
			&plan.BillingInterval,
			&plan.ApplicationLimit, 
			&plan.JobPostLimit,
			&plan.CleanerSeatLimit,
			&plan.IsActive, 
			&plan.CreatedAt,
			&plan.UpdatedAt,
		)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrPlanNotFound
	}

	if err != nil {
		return nil, err
	}

	return &plan, nil
	
}
func (r *SQLRepository)	CreateOrUpdateUserSubscription(ctx context.Context, userID uint, planID uint, status string) (*UserSubscription, error) {
	now  := time.Now()

	query := `
		INSERT INTO user_subscriptions (
			user_id,
			plan_id,
			status,
			trial_started_at,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $4, $4)
		ON CONFLICT (user_id)
		DO UPDATE SET 
			plan_id = EXCLUDED.plan_id, 
			status = EXCLUDED.status, 
			updated_at = EXCLUDED.updated_at 
		RETURNING 
			id, 
			user_id, 
			plan_id, 
			status, 
			trial_started_at, 
			trial_ends_at, 
			current_period_start, 
			current_period_end, 
			created_at, 
			updated_at  		
	`

	var subscription UserSubscription
	var planIDValue sql.NullInt64
	var trialEndsAt sql.NullTime
	var CurrentPeriodStart sql.NullTime
	var CurrentPeriodEnd sql.NullTime

	err := r.db.QueryRowContext(ctx, query, userID, planID, status, now).Scan(
		&subscription.ID,
		&subscription.UserID,
		&planIDValue,
		&subscription.Status,
		&subscription.TrialStartedAt,
		&trialEndsAt,
		&CurrentPeriodStart,
		&CurrentPeriodEnd,
		&subscription.CreatedAt,
		&subscription.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	if planIDValue.Valid {
		value := uint(planIDValue.Int64)
		subscription.PlanID = &value
	}

	if trialEndsAt.Valid {
		subscription.TrialEndsAt = &trialEndsAt.Time
	}

	if CurrentPeriodStart.Valid {
		subscription.CurrentPeriodStart = &CurrentPeriodStart.Time
	}

	if CurrentPeriodEnd.Valid {
		subscription.CurrentPeriodEnd = &CurrentPeriodEnd.Time
	}

	return &subscription, nil
}


func (r *SQLRepository)	GetByUserID(ctx context.Context, userID uint) (*UserSubscription, error) {
	query := `
		SELECT 
			id, 
			user_id, 
			plan_id, 
			status, 
			trial_started_at, 
			trial_ends_at, 
			current_period_start, 
			current_period_end,
			created_at,
			updated_at 
		FROM user_subscriptions 
		WHERE user_id = $1  
		LIMIT 1 	
	`

	
	var subscription UserSubscription
	var planIDValue sql.NullInt64
	var trialEndsAt sql.NullTime
	var CurrentPeriodStart sql.NullTime
	var CurrentPeriodEnd sql.NullTime

	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&subscription.ID,
		&subscription.UserID,
		&planIDValue,
		&subscription.Status,
		&subscription.TrialStartedAt,
		&trialEndsAt,
		&CurrentPeriodStart,
		&CurrentPeriodEnd,
		&subscription.CreatedAt,
		&subscription.UpdatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrSubscriptionNotFound
	}

	if err != nil {
		return nil, err
	}

	if planIDValue.Valid {
		value := uint(planIDValue.Int64)
		subscription.PlanID = &value
	}

	if trialEndsAt.Valid {
		subscription.TrialEndsAt = &trialEndsAt.Time
	}

		if CurrentPeriodStart.Valid {
		subscription.CurrentPeriodStart = &CurrentPeriodStart.Time
	}

	if CurrentPeriodEnd.Valid {
		subscription.CurrentPeriodEnd = &CurrentPeriodEnd.Time
	}

	return &subscription, nil

}
func (r *SQLRepository)	UpdateStatus(ctx context.Context, userID uint, status string) error {
	query := `
		UPDATE user_subscriptions
		SET status = $1, 
			updated_at = $2 
		WHERE user_id = $3 	
	`

	result, err := r.db.ExecContext(ctx, query, status, time.Now(), userID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrSubscriptionNotFound
	}

	return nil
}