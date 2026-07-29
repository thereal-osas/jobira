package usage

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

func (r *SQLRepository) CreateIfNotExists(ctx context.Context, userID uint) error {
	if userID == 0 {
		return ErrInvalidInput
	}

	query := `
	 	INSERT INTO user_usage (
	 		user_id,
			application_count,
			job_post_count,
			free_application_limit,
			free_job_post_limit,
			monetisation_enabled,
			trial_started_at,
			created_at,
			updated_at 	
		)
			VALUES ($1, 0, 0, 5, 5, false, $2, $2, $2)
			ON CONFLICT (user_id) DO NOTHING 
	 `

	now := time.Now()

	_, err := r.db.ExecContext(ctx, query, userID, now)
	return err
}

func (r *SQLRepository) GetByUserID(ctx context.Context, userID uint) (*UserUsage, error) {
	if userID == 0 {
		return nil, ErrInvalidInput
	}

	query := `
		SELECT
			id, 
			user_id,
			application_count, 
			job_post_count,
			free_application_limit,
			free_job_post_limit,
			monetisation_enabled, 
			trial_started_at, 
			trial_ends_at, 
			created_at, 
			updated_at 
		FROM user_usage 
		WHERE user_id = $1 
		LIMIT 1 	
	`

	var usage UserUsage
	var trialEndsAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&usage.ID,
		&usage.UserID,
		&usage.ApplicationCount,
		&usage.JobPostCount,
		&usage.FreeApplicationLimit,
		&usage.FreeJobPostLimit,
		&usage.MonetisationEnabled,
		&usage.TrialStartedAt,
		&trialEndsAt,
		&usage.CreatedAt,
		&usage.UpdatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrUsageNotFound
	}

	if err != nil {
		return nil, err
	}

	if trialEndsAt.Valid {
		usage.TrialEndsAt = &trialEndsAt.Time
	}

	return &usage, nil
}

func (r *SQLRepository) IncrementApplicationCount(ctx context.Context, userID uint) error {
	if userID == 0 {
		return ErrInvalidInput
	}

	if err := r.CreateIfNotExists(ctx, userID); err != nil {
		return err
	}

	query := `
		UPDATE user_usage
		SET application_count = application_count + 1, 
			updated_at = $1 
		WHERE user_id = $2 	
	`

	_, err := r.db.ExecContext(ctx, query, time.Now(), userID)
	return err
}

func (r *SQLRepository) IncrementJobPostCount(ctx context.Context, userID uint) error {
	if userID == 0 {
		return ErrInvalidInput
	}

	if err := r.CreateIfNotExists(ctx, userID); err != nil {
		return err
	}

	query := `
		UPDATE user_usage
		SET job_post_count = job_post_count + 1, 
			updated_at = $1 
		WHERE user_id = $2 	
	`

	_, err := r.db.ExecContext(ctx, query, time.Now(), userID)
	return err
}
func (r *SQLRepository) SetMonetisationEnabled(ctx context.Context, userID uint, enabled bool) error {
	if userID == 0 {
		return ErrInvalidInput
	}

	if err := r.CreateIfNotExists(ctx, userID); err != nil {
		return err
	}

	query := `
		UPDATE user_usage
		SET monetisation_enabled = $1, 
			updated_at = $2 
		WHERE user_id = $3  	
	`

	_, err := r.db.ExecContext(ctx, query, enabled, time.Now(), userID)
	return err
}

func (r *SQLRepository) GetUserSubscriptionAccess(ctx context.Context, userID uint) (*SubscriptionAccess, error) {
	query := `
		SELECT
			us.status,
			COALESCE(sp.application_limit, 0),
			COALESCE(sp.job_post_limit, 0)
		FROM user_subscriptions us
		LEFT JOIN subscription_plans sp
			ON sp.id = us.plan_id 
		WHERE us.user_id = $1 
		LIMIT 1		
	`

	var access SubscriptionAccess

	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&access.Status,
		&access.ApplicationLimit,
		&access.JobPostLimit,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &access, nil
}
