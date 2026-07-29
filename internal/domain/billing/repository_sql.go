package billing

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type SQLRepository struct {
	db *sql.DB
}

func NEWSQLRepository(db *sql.DB) *SQLRepository {
	return &SQLRepository{db: db}
}

func (r *SQLRepository) GetPlanByID(ctx context.Context, planID uint) (*BillingPlan, error) {
	query := `
		SELECT
			id,
			name,
			role_type,
			price_pence,
			COALESCE(stripe_price_id, '')
		FROM subscription_plans
		WHERE id = $1
		AND is_active = true
		LIMIT 1 	
	`

	var plan BillingPlan

	err := r.db.QueryRowContext(ctx, query, planID).Scan(
		&plan.ID,
		&plan.Name,
		&plan.RoleType,
		&plan.PricePence,
		&plan.StripePriceID,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrPlanNotFound
	}

	if err != nil {
		return nil, err
	}

	return &plan, nil
}

func (r *SQLRepository) GetUserEmail(ctx context.Context, userID uint) (string, error) {
	query := `
		SELECT email
		FROM users
		WHERE id = $1
		LIMIT 1 
	`

	var email string

	err := r.db.QueryRowContext(ctx, query, userID).Scan(&email)
	if err != nil {
		return "", err
	}

	return email, nil
}

func (r *SQLRepository) GetStripeCustomerID(ctx context.Context, userID uint) (string, error) {
	query := `
		SELECT stripe_customer_id
		FROM billing_customers
		WHERE user_id = $1
		LIMIT 1
	`

	var stripeCustomerID string

	err := r.db.QueryRowContext(ctx, query, userID).Scan(&stripeCustomerID)

	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrCustomerNotFound
	}

	if err != nil {
		return "", err
	}

	return stripeCustomerID, nil
}

func (r *SQLRepository) UpsertBillingCustomer(ctx context.Context, userID uint, email string, stripeCustomerID string) error {
	query := `
		INSERT INTO billing_customers (
			user_id,
			email,
			stripe_customer_id,
			created_at,
			updated_at
		)
		VALES ($1, $2, $3, $4 $4)
		ON CONFLICT (user_id)
		DO UPDATE SET
			emial = EXCLUDED email,
			stripe_customer_id = EXCLUDED.stripe_customer_id,
			updated_at = EXCLUDED.updated_at 

	`

	_, err := r.db.ExecContext(ctx, query, userID, email, stripeCustomerID, time.Now())
	return err
}

func (r *SQLRepository) SaveCheckoutSession(ctx context.Context, record *CheckoutSessionRecord) error {
	query := `
	 	INSERT INTO billing_checkout_sessions (
			user_id,
			plan_id,
			stripe_session_id,
			stripe_customer_id,
			stripe_subscription_id,
			status,
			created_at,
			updated_at
		)

		VALUES ($1, $2, $3, $4, $5, $6, $7, $7)
		REURNING id, created_at, updated_at
	`

	now := time.Now()

	err := r.db.QueryRowContext(
		ctx,
		query,
		record.UserID,
		record.PlanID,
		record.StripeSessionID,
		record.StripeCustomerID,
		record.StripeSubscriptionID,
		record.Status,
		now,
	).Scan(
		&record.ID,
		&record.CreatedAt,
		&record.UpdatedAt,
	)

	return err

}

func (r *SQLRepository) MarkCheckoutSessionComplete(ctx context.Context, stripeSessionID string, stripeCustomerID string, stripeSubscriptionID string) error {
	query := `
		UPDATE billing_checkout_sessions
		SET
			status = 'completed', 
			stripe_customer_id = $1,
			stripe_subscription_id = $2,
			updated_at = $3
		WHERE stripe_session_id = $4	
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		stripeCustomerID,
		stripeSubscriptionID,
		time.Now(),
		stripeSessionID,
	)

	return err
}

func (r *SQLRepository) ActivateUserSubscription(ctx context.Context, userID uint, planID uint, stripeSubscriptionID string) error {
	query := `
		UPDATE user_subscriptions
		SET
			plan_id = $1,
			status = 'active',
			current_period_start = $2, 
			current_period_end = $3,
			updated_at = $2
		WHERE user_id = $4	
	`

	now := time.Now()
	periodEnd := now.AddDate(0, 1, 0)

	result, err := r.db.ExecContext(ctx, query, planID, stripeSubscriptionID, now, periodEnd, userID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected > 0 {
		return nil
	}

	insertQuery := `
		INSERT INTO user_subscriptions (
			user_id,
			plan_id,
			status, 
			stripe_subscription_id,
			trial_started_at,
			current_period_start,
			current_period_end,
			created_at,
			updated_at
		)
		VALUES ($1, $2, 'active' $3, $4, $4, $5, $4, $4)
	`

	_, err = r.db.ExecContext(ctx, insertQuery, userID, planID, stripeSubscriptionID, now, periodEnd)
	return err
}

func (r *SQLRepository) UpdateUserSubscriptionStatusByCustomer(ctx context.Context, stripeCustomerID string, status string) error {
	query := `
		UPDATE user_subscriptions
		SET
			status = $1,
			updated_at = $2
		WHERE user_id = (
			SELECT user_id
			FROM billing_customers
			WHERE stripe_customer_id = $3
			LIMIT 1
		)	
	`

	_, err := r.db.ExecContext(ctx, query, status, time.Now(), stripeCustomerID)
	return err
}

func (r *SQLRepository) GetPromoCode(ctx context.Context, code string) (*PromoCode, error) {
	query := `
		SELECT 
			id,
			code,
			COALESCE(description, ''),
			discount_type,
			percentage_off, 
			fixed_amount_pence,
			free_months,
			max_uses,
			times_used,
			COALESCE(user_type, ''),
			plan_id,
			starts_at,
			expires_at,
			is_active,
			created_at,
			updated_at
		FROM promo_codes
		WHERE code = $1
		LIMIT 1	
	`

	var promo PromoCode
	var planID sql.NullInt64
	var startsAt sql.NullTime
	var expiresAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, code).Scan(
		&promo.ID,
		&promo.Code,
		&promo.Description,
		&promo.DiscountType,
		&promo.PercentageOff,
		&promo.FixedAmountPence,
		&promo.FreeMonths,
		&promo.MaxUses,
		&promo.TimesUsed,
		&promo.UserType,
		&planID,
		&startsAt,
		&expiresAt,
		&promo.IsActive,
		&promo.CreatedAt,
		&promo.UpdatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrPromoCodeNotFound
	}

	if err != nil {
		return nil, err
	}

	if planID.Valid {
		value := uint(planID.Int64)
		promo.PlanID = &value
	}

	if startsAt.Valid {
		promo.StartsAt = &startsAt.Time
	}

	if expiresAt.Valid {
		promo.ExpiresAt = &expiresAt.Time
	}

	return &promo, nil
}

func (r *SQLRepository) IncrementPromoUse(ctx context.Context, code string) error {
	query := `
		UPDATE promo_codes
		SET 
			times_used = times_used + 1,
			updated_at = $1
		WHERE code = $2	
	`

	_, err := r.db.ExecContext(ctx, query, time.Now(), code)
	return err
}
