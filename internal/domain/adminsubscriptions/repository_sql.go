package adminsubscriptions

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

func (r *SQLRepository) ListPlans(ctx context.Context) ([]AdminSubscriptionPlan, error) {
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
	ORDER BY id ASC 	
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var plans []AdminSubscriptionPlan

	for rows.Next() {
		var plan AdminSubscriptionPlan

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

func (r *SQLRepository) CreatePlan(ctx context.Context, plan *AdminSubscriptionPlan) error {
	query := `
		INSERT INTO subscription_plans (
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
		)
	VALUES ($1, $2, $3, $4, $5, $6, $7, true, $8, $8)
	RETURNING id, is_active, created_at, updated_at 
	`
	now := time.Now()

	return r.db.QueryRowContext(
		ctx,
		query,
		plan.Name,
		plan.RoleType,
		plan.PricePence,
		plan.BillingInterval,
		plan.ApplicationLimit,
		plan.JobPostLimit,
		plan.CleanerSeatLimit,
		now, 
	).Scan(
		&plan.ID,
		&plan.IsActive,
		&plan.CreatedAt,
		&plan.UpdatedAt,
	)
}

func (r *SQLRepository)	UpdatePlan(ctx context.Context, plan *AdminSubscriptionPlan) error {
	query := `
		UPDATE subscription_plans
		SET 
			name = $1, 
			role_type = $2, 
			price_pence = $3, 
			billing_interval = $4, 
			application_limit = $5, 
			job_post_limit = $6, 
			cleaner_seat_limit = $7, 
			is_active = $8, 
			updated_at = $9
		WHERE id = $10 	 
	`

	result, err := r.db.ExecContext(
		ctx, 
		query,
		plan.Name,
		plan.RoleType,
		plan.PricePence,
		plan.BillingInterval,
		plan.ApplicationLimit,
		plan.JobPostLimit,
		plan.CleanerSeatLimit,
		plan.IsActive,
		time.Now(),
		plan.ID,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrPlanNotFound
	}

	return nil

}

func (r *SQLRepository) DisablePlan(ctx context.Context, planID uint) error {
	query := `
		UPDATE subscription_plans
		SET is_active = false, 
			updated_at = $1 
		WHERE id = $2 	
	`

	result, err := r.db.ExecContext(ctx, query, time.Now(), planID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrPlanNotFound
	}

	return nil
}

func (r *SQLRepository)	ListUserSubscriptions(ctx context.Context) ([]AdminUserSubscription, error) {
	query := `
	SELECT
		us.id, 
		us.user_id,
		u.email,
		u.full_name, 
		us.plan_id, 
		sp.name,
		us.status,
		us.trial_started_at,
		us.trial_ends_at,
		us.current_period_start, 
		us.current_period_end,
		us.created_at,
		us.updated_at 
	FROM user_subscriptions us 
	INNER JOIN users u ON u.id = us.user_id 
	LEFT JOIN subscription_plans sp ON sp.id = us.plan_id 
	ORDER BY us.created_at DESC 	
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subscriptions []AdminUserSubscription

	for rows.Next() {
		var sub AdminUserSubscription
		var planID sql.NullInt64
		var PlanName sql.NullString
		var trialEndsAt sql.NullTime
		var CurrentPeriodStart sql.NullTime
		var CurrentPeriodEnd sql.NullTime

		err := rows.Scan(
			&sub.ID,
			&sub.UserID,
			&sub.UserEmail,
			&sub.UserFullName,
			&planID,
			&PlanName,
			&sub.Status,
			&sub.TrialStartedAt,
			&trialEndsAt,
			&CurrentPeriodStart,
			&CurrentPeriodEnd,
			&sub.CreatedAt,
			&sub.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if planID.Valid {
			value := uint(planID.Int64)
			sub.PlanID = &value
		}

		if PlanName.Valid {
			value := PlanName.String
			sub.PlanName = &value
		}

		if trialEndsAt.Valid {
			sub.TrialEndsAt = &trialEndsAt.Time
		}

		if CurrentPeriodStart.Valid {
			sub.CurrentPeriodStart = &CurrentPeriodStart.Time
		}

		if CurrentPeriodEnd.Valid {
			sub.CurrentPeriodEnd = &CurrentPeriodEnd.Time
		}

		subscriptions = append(subscriptions, sub)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return subscriptions, nil
}

func (r *SQLRepository) GetUserSubscriptions(ctx context.Context, userID uint) (*AdminUserSubscription, error) {
	query := `
		SELECT
			us.id,
			us.user_id,
			u.email,
			u.full_name,
			us.plan_id,
			sp.name,
			us.status,
			us.trial_started_at,
			us.trial_ends_at,
			us.current_period_start,
			us.current_period_end,
			us.created_at,
			us.updated_at
		FROM user_subscriptions us
		INNER JOIN users u ON u.id = us.user_id
		LEFT JOIN subscription_plans sp ON sp.id = us.plan_id
		WHERE us.user_id = $1
		LIMIT 1
	`

	var sub AdminUserSubscription
	var planID sql.NullInt64
	var planName sql.NullString
	var trialEndsAt sql.NullTime
	var currentPeriodStart sql.NullTime
	var currentPeriodEnd sql.NullTime

	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&sub.ID,
		&sub.UserID,
		&sub.UserEmail,
		&sub.UserFullName,
		&planID,
		&planName,
		&sub.Status,
		&sub.TrialStartedAt,
		&trialEndsAt,
		&currentPeriodStart,
		&currentPeriodEnd,
		&sub.CreatedAt,
		&sub.UpdatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrSubscriptionNotFound
	}

	if err != nil {
		return nil, err
	}

	if planID.Valid {
		value := uint(planID.Int64)
		sub.PlanID = &value
	}

	if planName.Valid {
		value := planName.String
		sub.PlanName = &value
	}

	if trialEndsAt.Valid {
		sub.TrialEndsAt = &trialEndsAt.Time
	}

	if currentPeriodStart.Valid {
		sub.CurrentPeriodStart = &currentPeriodStart.Time
	}

	if currentPeriodEnd.Valid {
		sub.CurrentPeriodEnd = &currentPeriodEnd.Time
	}

	return &sub, nil
}

func (r *SQLRepository)	CreateOrUpdateUserSubscription(ctx context.Context, userID uint, planID uint, status string) error {
	now := time.Now()

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
	`

	_, err := r.db.ExecContext(ctx, query, userID, planID, status, now)
	return err
}