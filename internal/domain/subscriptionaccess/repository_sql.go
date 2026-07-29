package subscriptionaccess

import (
	"context"
	"database/sql"
	"time"
)

type SQLRepository struct {
	db *sql.DB
}

func NewSQLRepository(db *sql.DB) *SQLRepository {
	return &SQLRepository{db: db}
}

func (r *SQLRepository) EnsureDailyAccess(ctx context.Context, userID uint) error {
	query := `
		INSERT INTO user_daily_access (
			user_id,
			access_date,
			applications_today,
			jobs_posted_today,
			created_at,
			updated_at
		)
		VALUES ($1, CURRENT_DATE, 0, 0, $2, $2)
		ON CONFLICT (user_id, access_date)
		DO NOTHING
	`

	_, err := r.db.ExecContext(ctx, query, userID, time.Now())
	return err
}

func (r *SQLRepository) IsLaunchGraceActive(ctx context.Context) (bool, error) {
	query := `
		SELECT
			launch_grace_enabled
			AND launch_grace_started_at IS NOT NULL
			AND NOW() < launch_grace_started_at + (launch_grace_days || ' days')::interval
		FROM platform_access_settings
		ORDER BY id DESC
		LIMIT 1
	`

	var active bool

	err := r.db.QueryRowContext(ctx, query).Scan(&active)
	if err != nil {
		return false, err
	}

	return active, nil
}

func (r *SQLRepository) GetCleanerAccessStatus(ctx context.Context, userID uint) (*CleanerAccessStatus, error) {
	if err := r.EnsureUserUsage(ctx, userID); err != nil {
		return nil, err
	}

	if err := r.EnsureDailyAccess(ctx, userID); err != nil {
		return nil, err
	}

	query := `
		SELECT
		uu.user_id,
		COALESCE(us.status, 'none'),
		CASE
			WHEN uu.trial_started_at IS NOT NULL
			AND NOW() < uu.trial_started_at + (uu.trial_days || ' days')::interval
			THEN true
			ELSE false
		END,	
		uu.application_count,
		uu.free_application_limit,
		COALESCE(uda.applications_today, 0),
		uu.daily_application_limit,
		uu.free_job_visibility_delay_minutes
	FROM user_usage uu
	LEFT JOIN user_subscriptions us ON us.user_id = uu.user_id
	LEFT JOIN user_daily_access uda 
		ON uda.user_id = uu.user_id 
		AND uda.access_date = CURRENT_DATE
	WHERE uu.user_id = $1
	LIMIT 1
	`

	var status CleanerAccessStatus

	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&status.UserID,
		&status.SubscriptionStatus,
		&status.TrialActive,
		&status.ApplicationCount,
		&status.FreeApplicationLimit,
		&status.ApplicationsToday,
		&status.DailyApplicationLimit,
		&status.FreeJobVisibilityDelayMinutes,
	)

	if err != nil {
		return nil, err
	}

	launchGraceActive, err := r.IsLaunchGraceActive(ctx)
	if err != nil {
		return nil, err
	}

	status.LaunchGraceActive = launchGraceActive
	status.Premium = status.SubscriptionStatus == "active"

	status.CanViewJobs = true
	status.CanApply = status.ApplicationsToday < status.DailyApplicationLimit

	if status.Premium {
		status.CanApply = true
		status.CanViewClientDetails = true
		status.HasInstantAlerts = true
		status.HasPriorityPlacement = true
		status.HasPremiumProfile = true
		status.HasEarlyJobAccess = true
		status.UpgradeMessage = ""
		return &status, nil
	}

	if status.TrialActive {
		status.CanViewClientDetails = true
		status.HasInstantAlerts = true
		status.HasPriorityPlacement = true
		status.HasPremiumProfile = true
		status.HasEarlyJobAccess = true
		status.UpgradeMessage = "you are currently on your free trial. Subscribe to keep premium access after your trial ends."
		return &status, nil
	}

	status.CanViewClientDetails = false
	status.HasInstantAlerts = false
	status.HasPriorityPlacement = false
	status.HasEarlyJobAccess = false
	status.UpgradeMessage = "Upgrade to Premium for instant alerts, priority placement, premium profile features, and full client details."

	return &status, nil
}

func (r *SQLRepository) GetClientAccessStatus(ctx context.Context, userID uint) (*ClientAccessStatus, error) {
	if err := r.EnsureUserUsage(ctx, userID); err != nil {
		return nil, err
	}

	if err := r.EnsureDailyAccess(ctx, userID); err != nil {
		return nil, err
	}

	query := `
		SELECT 
			uu.user_id, 
			COALESCE(us.status, 'none'),
			uu.job_post_count,
			uu.free_job_post_limit,
			uda.jobs_posted_today, 
			uu.daily_job_post_limit 
		FROM user_usage uu 
		LEFT JOIN user_subscriptions us ON us.user_id = uu.user_id
		LEFT JOIN user_daily_access uda 
			ON uda.user_id = uu.user_id 
			AND uda.access_date = CURRENT_DATE
		WHERE uu.user_id = $1 
		LIMIT 1		
	`

	var status ClientAccessStatus

	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&status.UserID,
		&status.SubscriptionStatus,
		&status.JobPostCount,
		&status.FreeJobPostLimit,
		&status.JobsPostedToday,
		&status.DailyJobPostLimit,
	)
	if err != nil {
		return nil, err
	}

	launchGraceActive, err := r.IsLaunchGraceActive(ctx)
	if err != nil {
		return nil, err
	}

	status.LaunchGraceActive = launchGraceActive
	status.Premium = status.SubscriptionStatus == "active"

	status.CanPostJob = status.JobsPostedToday < status.DailyJobPostLimit
	status.CanViewApplicants = true
	status.CanContactCleaners = true
	status.CanUseRepeatBooking = true

	if status.Premium {
		status.CanPostJob = true
		status.UpgradeMessage = ""
		return &status, nil
	}

	if status.LaunchGraceActive {
		status.UpgradeMessage = "Launch access is active. Subscribe before the launch period ends to keep full posting access."
		return &status, nil
	}

	if status.JobPostCount >= status.FreeJobPostLimit {
		status.CanPostJob = false
		status.UpgradeMessage = "You have used your free job posts. Upgrade or pay per post to continue posting jobs."
		return &status, nil
	}

	status.UpgradeMessage = "Upgrade for more job posts, repeat booking tools, and priority access to cleaners."

	return &status, nil
}

func (r *SQLRepository) IncrementApplicationsToday(ctx context.Context, userID uint) error {
	if err := r.EnsureDailyAccess(ctx, userID); err != nil {
		return err
	}

	query := `
		UPDATE user_daily_access
		SET 
			applications_today = applications_today + 1, 
			updated_at = $1 
		WHERE user_id = $2
		AND access_date = CURRENT_DATE	
	`

	_, err := r.db.ExecContext(ctx, query, time.Now(), userID)
	return err
}

func (r *SQLRepository) IncrementJobsPostToday(ctx context.Context, userID uint) error {
	if err := r.EnsureDailyAccess(ctx, userID); err != nil {
		return err
	}

	query := `
		UPDATE user_daily_access
		SET
			jobs_posted_today = jobs_posted_today + 1, 
			updated_at = $1
		WHERE user_id = $2
		AND access_date = CURRENT_DATE	
	`

	_, err := r.db.ExecContext(ctx, query, time.Now(), userID)
	return err
}

func (r *SQLRepository) EnsureUserUsage(ctx context.Context, userID uint) error {
	query := `
		INSERT INTO user_usage (
			user_id,
			application_count,
			job_post_count,
			free_application_limit,
			free_job_post_limit,
			monetisation_enabled,
			trial_started_at,
			trial_days,
			daily_application_limit,
			daily_job_post_limit,
			free_job_visibility_delay_minutes,
			created_at,
			updated_at
		)
		VALUES ($1, 0, 0, 5, 5, true, $2, 14, 5, 5, 20, $2, $2)
		ON CONFLICT (user_id) DO NOTHING
	`

	_, err := r.db.ExecContext(ctx, query, userID, time.Now())
	return err
}
