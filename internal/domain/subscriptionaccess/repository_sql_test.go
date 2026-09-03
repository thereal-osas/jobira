package subscriptionaccess

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func newSubscriptionAccessTestDB(
	t *testing.T,
) (*sql.DB, sqlmock.Sqlmock) {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf(
			"failed to create sqlmock: %v",
			err,
		)
	}

	t.Cleanup(func() {
		_ = db.Close()
	})

	return db, mock
}

func TestSQLRepository_EnsureDailyAccess_Success(t *testing.T) {
	db, mock := newSubscriptionAccessTestDB(t)

	repo := NewSQLRepository(db)

	mock.ExpectExec(
		regexp.QuoteMeta(`
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
	`),
	).
		WithArgs(
			uint(7),
			sqlmock.AnyArg(),
		).
		WillReturnResult(
			sqlmock.NewResult(1, 1),
		)

	err := repo.EnsureDailyAccess(
		context.Background(),
		7,
	)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf(
			"unmet expectations: %v",
			err,
		)
	}
}

func TestSQLRepository_EnsureDailyAccess_Error(t *testing.T) {
	db, mock := newSubscriptionAccessTestDB(t)

	repo := NewSQLRepository(db)

	expectedErr := errors.New("insert failed")

	mock.ExpectExec(
		"INSERT INTO user_daily_access",
	).
		WithArgs(
			uint(7),
			sqlmock.AnyArg(),
		).
		WillReturnError(expectedErr)

	err := repo.EnsureDailyAccess(
		context.Background(),
		7,
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected insert error, got %v",
			err,
		)
	}
}

func TestSQLRepository_IsLaunchGraceActive_True(t *testing.T) {
	db, mock := newSubscriptionAccessTestDB(t)

	repo := NewSQLRepository(db)

	mock.ExpectQuery(
		"SELECT",
	).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{"active"},
			).AddRow(true),
		)

	active, err := repo.IsLaunchGraceActive(
		context.Background(),
	)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if !active {
		t.Fatalf("expected launch grace to be active")
	}
}

func TestSQLRepository_IsLaunchGraceActive_False(t *testing.T) {
	db, mock := newSubscriptionAccessTestDB(t)

	repo := NewSQLRepository(db)

	mock.ExpectQuery(
		"SELECT",
	).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{"active"},
			).AddRow(false),
		)

	active, err := repo.IsLaunchGraceActive(
		context.Background(),
	)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if active {
		t.Fatalf("expected launch grace to be inactive")
	}
}

func TestSQLRepository_IsLaunchGraceActive_Error(t *testing.T) {
	db, mock := newSubscriptionAccessTestDB(t)

	repo := NewSQLRepository(db)

	expectedErr := errors.New("query failed")

	mock.ExpectQuery(
		"SELECT",
	).
		WillReturnError(expectedErr)

	active, err := repo.IsLaunchGraceActive(
		context.Background(),
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected query error, got %v",
			err,
		)
	}

	if active {
		t.Fatalf("expected false on error")
	}
}

func TestSQLRepository_EnsureUserUsage_Success(t *testing.T) {
	db, mock := newSubscriptionAccessTestDB(t)

	repo := NewSQLRepository(db)

	mock.ExpectExec(
		"INSERT INTO user_usage",
	).
		WithArgs(
			uint(7),
			sqlmock.AnyArg(),
		).
		WillReturnResult(
			sqlmock.NewResult(1, 1),
		)

	err := repo.EnsureUserUsage(
		context.Background(),
		7,
	)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestSQLRepository_EnsureUserUsage_Error(t *testing.T) {
	db, mock := newSubscriptionAccessTestDB(t)

	repo := NewSQLRepository(db)

	expectedErr := errors.New("insert failed")

	mock.ExpectExec(
		"INSERT INTO user_usage",
	).
		WithArgs(
			uint(7),
			sqlmock.AnyArg(),
		).
		WillReturnError(expectedErr)

	err := repo.EnsureUserUsage(
		context.Background(),
		7,
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected insert error, got %v",
			err,
		)
	}
}

func TestSQLRepository_GetCleanerAccessStatus_Premium(t *testing.T) {
	db, mock := newSubscriptionAccessTestDB(t)

	repo := NewSQLRepository(db)

	mock.ExpectExec(
		"INSERT INTO user_usage",
	).
		WithArgs(
			uint(7),
			sqlmock.AnyArg(),
		).
		WillReturnResult(
			sqlmock.NewResult(1, 1),
		)

	mock.ExpectExec(
		"INSERT INTO user_daily_access",
	).
		WithArgs(
			uint(7),
			sqlmock.AnyArg(),
		).
		WillReturnResult(
			sqlmock.NewResult(1, 1),
		)

	mock.ExpectQuery(
		"SELECT[\\s\\S]*uu.user_id",
	).
		WithArgs(uint(7)).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{
					"user_id",
					"subscription_status",
					"trial_active",
					"application_count",
					"free_application_limit",
					"applications_today",
					"daily_application_limit",
					"free_job_visibility_delay_minutes",
				},
			).AddRow(
				7,
				"active",
				false,
				10,
				5,
				5,
				5,
				20,
			),
		)

	mock.ExpectQuery(
		"SELECT",
	).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{"active"},
			).AddRow(false),
		)

	status, err := repo.GetCleanerAccessStatus(
		context.Background(),
		7,
	)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if !status.Premium {
		t.Fatalf("expected premium status")
	}

	if !status.CanApply {
		t.Fatalf(
			"premium cleaner should be able to apply",
		)
	}

	if !status.CanViewJobs {
		t.Fatalf(
			"expected cleaner to view jobs",
		)
	}

	if !status.CanViewClientDetails {
		t.Fatalf(
			"expected premium client-detail access",
		)
	}

	if !status.HasInstantAlerts {
		t.Fatalf(
			"expected instant alerts",
		)
	}

	if !status.HasPriorityPlacement {
		t.Fatalf(
			"expected priority placement",
		)
	}

	if !status.HasPremiumProfile {
		t.Fatalf(
			"expected premium profile",
		)
	}

	if !status.HasEarlyJobAccess {
		t.Fatalf(
			"expected early job access",
		)
	}

	if status.UpgradeMessage != "" {
		t.Fatalf(
			"expected empty upgrade message",
		)
	}
}

func TestSQLRepository_GetCleanerAccessStatus_Trial(t *testing.T) {
	db, mock := newSubscriptionAccessTestDB(t)

	repo := NewSQLRepository(db)

	mock.ExpectExec(
		"INSERT INTO user_usage",
	).
		WithArgs(
			uint(7),
			sqlmock.AnyArg(),
		).
		WillReturnResult(
			sqlmock.NewResult(1, 1),
		)

	mock.ExpectExec(
		"INSERT INTO user_daily_access",
	).
		WithArgs(
			uint(7),
			sqlmock.AnyArg(),
		).
		WillReturnResult(
			sqlmock.NewResult(1, 1),
		)

	mock.ExpectQuery(
		"SELECT[\\s\\S]*uu.user_id",
	).
		WithArgs(uint(7)).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{
					"user_id",
					"subscription_status",
					"trial_active",
					"application_count",
					"free_application_limit",
					"applications_today",
					"daily_application_limit",
					"free_job_visibility_delay_minutes",
				},
			).AddRow(
				7,
				"none",
				true,
				1,
				5,
				1,
				5,
				20,
			),
		)

	mock.ExpectQuery(
		"SELECT",
	).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{"active"},
			).AddRow(false),
		)

	status, err := repo.GetCleanerAccessStatus(
		context.Background(),
		7,
	)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if !status.TrialActive {
		t.Fatalf("expected trial to be active")
	}

	if status.Premium {
		t.Fatalf("expected premium to be false")
	}

	if !status.CanApply {
		t.Fatalf("expected cleaner to apply")
	}

	if !status.CanViewClientDetails {
		t.Fatalf(
			"expected trial client-detail access",
		)
	}

	if !status.HasInstantAlerts {
		t.Fatalf(
			"expected trial instant alerts",
		)
	}

	if !status.HasPriorityPlacement {
		t.Fatalf(
			"expected trial priority placement",
		)
	}

	if !status.HasPremiumProfile {
		t.Fatalf(
			"expected trial premium profile",
		)
	}

	if !status.HasEarlyJobAccess {
		t.Fatalf(
			"expected trial early access",
		)
	}

	if status.UpgradeMessage == "" {
		t.Fatalf(
			"expected trial upgrade message",
		)
	}
}

func TestSQLRepository_GetCleanerAccessStatus_FreeDailyLimitReached(
	t *testing.T,
) {
	db, mock := newSubscriptionAccessTestDB(t)

	repo := NewSQLRepository(db)

	mock.ExpectExec(
		"INSERT INTO user_usage",
	).
		WithArgs(
			uint(7),
			sqlmock.AnyArg(),
		).
		WillReturnResult(
			sqlmock.NewResult(1, 1),
		)

	mock.ExpectExec(
		"INSERT INTO user_daily_access",
	).
		WithArgs(
			uint(7),
			sqlmock.AnyArg(),
		).
		WillReturnResult(
			sqlmock.NewResult(1, 1),
		)

	mock.ExpectQuery(
		"SELECT[\\s\\S]*uu.user_id",
	).
		WithArgs(uint(7)).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{
					"user_id",
					"subscription_status",
					"trial_active",
					"application_count",
					"free_application_limit",
					"applications_today",
					"daily_application_limit",
					"free_job_visibility_delay_minutes",
				},
			).AddRow(
				7,
				"none",
				false,
				2,
				5,
				5,
				5,
				20,
			),
		)

	mock.ExpectQuery(
		"SELECT",
	).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{"active"},
			).AddRow(false),
		)

	status, err := repo.GetCleanerAccessStatus(
		context.Background(),
		7,
	)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if status.CanApply {
		t.Fatalf(
			"expected CanApply false at daily limit",
		)
	}

	if status.CanViewClientDetails {
		t.Fatalf(
			"expected free cleaner client details false",
		)
	}

	if status.HasInstantAlerts {
		t.Fatalf(
			"expected free cleaner instant alerts false",
		)
	}

	if status.HasPriorityPlacement {
		t.Fatalf(
			"expected free cleaner priority placement false",
		)
	}

	if status.HasEarlyJobAccess {
		t.Fatalf(
			"expected free cleaner early access false",
		)
	}

	if status.UpgradeMessage == "" {
		t.Fatalf(
			"expected upgrade message",
		)
	}
}

func TestSQLRepository_GetCleanerAccessStatus_EnsureUserUsageError(
	t *testing.T,
) {
	db, mock := newSubscriptionAccessTestDB(t)

	repo := NewSQLRepository(db)

	expectedErr := errors.New("usage insert failed")

	mock.ExpectExec(
		"INSERT INTO user_usage",
	).
		WithArgs(
			uint(7),
			sqlmock.AnyArg(),
		).
		WillReturnError(expectedErr)

	status, err := repo.GetCleanerAccessStatus(
		context.Background(),
		7,
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected usage error, got %v",
			err,
		)
	}

	if status != nil {
		t.Fatalf("expected nil status")
	}
}

func TestSQLRepository_GetCleanerAccessStatus_EnsureDailyAccessError(
	t *testing.T,
) {
	db, mock := newSubscriptionAccessTestDB(t)

	repo := NewSQLRepository(db)

	expectedErr := errors.New("daily insert failed")

	mock.ExpectExec(
		"INSERT INTO user_usage",
	).
		WithArgs(
			uint(7),
			sqlmock.AnyArg(),
		).
		WillReturnResult(
			sqlmock.NewResult(1, 1),
		)

	mock.ExpectExec(
		"INSERT INTO user_daily_access",
	).
		WithArgs(
			uint(7),
			sqlmock.AnyArg(),
		).
		WillReturnError(expectedErr)

	status, err := repo.GetCleanerAccessStatus(
		context.Background(),
		7,
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected daily access error, got %v",
			err,
		)
	}

	if status != nil {
		t.Fatalf("expected nil status")
	}
}

func TestSQLRepository_GetCleanerAccessStatus_QueryError(t *testing.T) {
	db, mock := newSubscriptionAccessTestDB(t)

	repo := NewSQLRepository(db)

	expectedErr := errors.New("query failed")

	mock.ExpectExec(
		"INSERT INTO user_usage",
	).
		WithArgs(
			uint(7),
			sqlmock.AnyArg(),
		).
		WillReturnResult(
			sqlmock.NewResult(1, 1),
		)

	mock.ExpectExec(
		"INSERT INTO user_daily_access",
	).
		WithArgs(
			uint(7),
			sqlmock.AnyArg(),
		).
		WillReturnResult(
			sqlmock.NewResult(1, 1),
		)

	mock.ExpectQuery(
		"SELECT[\\s\\S]*uu.user_id",
	).
		WithArgs(uint(7)).
		WillReturnError(expectedErr)

	status, err := repo.GetCleanerAccessStatus(
		context.Background(),
		7,
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected query error, got %v",
			err,
		)
	}

	if status != nil {
		t.Fatalf("expected nil status")
	}
}

func TestSQLRepository_GetCleanerAccessStatus_LaunchGraceQueryError(
	t *testing.T,
) {
	db, mock := newSubscriptionAccessTestDB(t)

	repo := NewSQLRepository(db)

	expectedErr := errors.New("launch grace query failed")

	mock.ExpectExec(
		"INSERT INTO user_usage",
	).
		WithArgs(
			uint(7),
			sqlmock.AnyArg(),
		).
		WillReturnResult(
			sqlmock.NewResult(1, 1),
		)

	mock.ExpectExec(
		"INSERT INTO user_daily_access",
	).
		WithArgs(
			uint(7),
			sqlmock.AnyArg(),
		).
		WillReturnResult(
			sqlmock.NewResult(1, 1),
		)

	mock.ExpectQuery(
		"SELECT[\\s\\S]*uu.user_id",
	).
		WithArgs(uint(7)).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{
					"user_id",
					"subscription_status",
					"trial_active",
					"application_count",
					"free_application_limit",
					"applications_today",
					"daily_application_limit",
					"free_job_visibility_delay_minutes",
				},
			).AddRow(
				7,
				"none",
				false,
				1,
				5,
				1,
				5,
				20,
			),
		)

	mock.ExpectQuery(
		"SELECT",
	).
		WillReturnError(expectedErr)

	status, err := repo.GetCleanerAccessStatus(
		context.Background(),
		7,
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected launch grace error, got %v",
			err,
		)
	}

	if status != nil {
		t.Fatalf("expected nil status")
	}
}

func TestSQLRepository_GetClientAccessStatus_Premium(t *testing.T) {
	db, mock := newSubscriptionAccessTestDB(t)

	repo := NewSQLRepository(db)

	mock.ExpectExec(
		"INSERT INTO user_usage",
	).
		WithArgs(
			uint(11),
			sqlmock.AnyArg(),
		).
		WillReturnResult(
			sqlmock.NewResult(1, 1),
		)

	mock.ExpectExec(
		"INSERT INTO user_daily_access",
	).
		WithArgs(
			uint(11),
			sqlmock.AnyArg(),
		).
		WillReturnResult(
			sqlmock.NewResult(1, 1),
		)

	mock.ExpectQuery(
		"SELECT[\\s\\S]*uu.user_id",
	).
		WithArgs(uint(11)).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{
					"user_id",
					"subscription_status",
					"job_post_count",
					"free_job_post_limit",
					"jobs_posted_today",
					"daily_job_post_limit",
				},
			).AddRow(
				11,
				"active",
				9,
				5,
				5,
				5,
			),
		)

	mock.ExpectQuery(
		"SELECT",
	).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{"active"},
			).AddRow(false),
		)

	status, err := repo.GetClientAccessStatus(
		context.Background(),
		11,
	)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if !status.Premium {
		t.Fatalf("expected premium")
	}

	if !status.CanPostJob {
		t.Fatalf(
			"premium client should be able to post",
		)
	}

	if !status.CanViewApplicants {
		t.Fatalf(
			"expected applicant access",
		)
	}

	if !status.CanContactCleaners {
		t.Fatalf(
			"expected cleaner contact access",
		)
	}

	if !status.CanUseRepeatBooking {
		t.Fatalf(
			"expected repeat booking access",
		)
	}

	if status.UpgradeMessage != "" {
		t.Fatalf(
			"expected empty upgrade message",
		)
	}
}

func TestSQLRepository_GetClientAccessStatus_LaunchGrace(t *testing.T) {
	db, mock := newSubscriptionAccessTestDB(t)

	repo := NewSQLRepository(db)

	mock.ExpectExec(
		"INSERT INTO user_usage",
	).
		WithArgs(
			uint(11),
			sqlmock.AnyArg(),
		).
		WillReturnResult(
			sqlmock.NewResult(1, 1),
		)

	mock.ExpectExec(
		"INSERT INTO user_daily_access",
	).
		WithArgs(
			uint(11),
			sqlmock.AnyArg(),
		).
		WillReturnResult(
			sqlmock.NewResult(1, 1),
		)

	mock.ExpectQuery(
		"SELECT[\\s\\S]*uu.user_id",
	).
		WithArgs(uint(11)).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{
					"user_id",
					"subscription_status",
					"job_post_count",
					"free_job_post_limit",
					"jobs_posted_today",
					"daily_job_post_limit",
				},
			).AddRow(
				11,
				"none",
				5,
				5,
				1,
				5,
			),
		)

	mock.ExpectQuery(
		"SELECT",
	).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{"active"},
			).AddRow(true),
		)

	status, err := repo.GetClientAccessStatus(
		context.Background(),
		11,
	)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if !status.LaunchGraceActive {
		t.Fatalf(
			"expected launch grace active",
		)
	}

	if !status.CanPostJob {
		t.Fatalf(
			"expected posting during launch grace",
		)
	}

	if status.UpgradeMessage == "" {
		t.Fatalf(
			"expected launch upgrade message",
		)
	}
}

func TestSQLRepository_GetClientAccessStatus_FreeLimitReached(
	t *testing.T,
) {
	db, mock := newSubscriptionAccessTestDB(t)

	repo := NewSQLRepository(db)

	mock.ExpectExec(
		"INSERT INTO user_usage",
	).
		WithArgs(
			uint(11),
			sqlmock.AnyArg(),
		).
		WillReturnResult(
			sqlmock.NewResult(1, 1),
		)

	mock.ExpectExec(
		"INSERT INTO user_daily_access",
	).
		WithArgs(
			uint(11),
			sqlmock.AnyArg(),
		).
		WillReturnResult(
			sqlmock.NewResult(1, 1),
		)

	mock.ExpectQuery(
		"SELECT[\\s\\S]*uu.user_id",
	).
		WithArgs(uint(11)).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{
					"user_id",
					"subscription_status",
					"job_post_count",
					"free_job_post_limit",
					"jobs_posted_today",
					"daily_job_post_limit",
				},
			).AddRow(
				11,
				"none",
				5,
				5,
				1,
				5,
			),
		)

	mock.ExpectQuery(
		"SELECT",
	).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{"active"},
			).AddRow(false),
		)

	status, err := repo.GetClientAccessStatus(
		context.Background(),
		11,
	)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if status.CanPostJob {
		t.Fatalf(
			"expected posting blocked after free limit",
		)
	}

	if status.UpgradeMessage == "" {
		t.Fatalf(
			"expected upgrade message",
		)
	}
}

func TestSQLRepository_GetClientAccessStatus_FreeCanPost(t *testing.T) {
	db, mock := newSubscriptionAccessTestDB(t)

	repo := NewSQLRepository(db)

	mock.ExpectExec(
		"INSERT INTO user_usage",
	).
		WithArgs(
			uint(11),
			sqlmock.AnyArg(),
		).
		WillReturnResult(
			sqlmock.NewResult(1, 1),
		)

	mock.ExpectExec(
		"INSERT INTO user_daily_access",
	).
		WithArgs(
			uint(11),
			sqlmock.AnyArg(),
		).
		WillReturnResult(
			sqlmock.NewResult(1, 1),
		)

	mock.ExpectQuery(
		"SELECT[\\s\\S]*uu.user_id",
	).
		WithArgs(uint(11)).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{
					"user_id",
					"subscription_status",
					"job_post_count",
					"free_job_post_limit",
					"jobs_posted_today",
					"daily_job_post_limit",
				},
			).AddRow(
				11,
				"none",
				2,
				5,
				1,
				5,
			),
		)

	mock.ExpectQuery(
		"SELECT",
	).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{"active"},
			).AddRow(false),
		)

	status, err := repo.GetClientAccessStatus(
		context.Background(),
		11,
	)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if !status.CanPostJob {
		t.Fatalf(
			"expected free client to be able to post",
		)
	}

	if status.UpgradeMessage == "" {
		t.Fatalf(
			"expected upgrade message",
		)
	}
}

func TestSQLRepository_GetClientAccessStatus_QueryError(t *testing.T) {
	db, mock := newSubscriptionAccessTestDB(t)

	repo := NewSQLRepository(db)

	expectedErr := errors.New("query failed")

	mock.ExpectExec(
		"INSERT INTO user_usage",
	).
		WithArgs(
			uint(11),
			sqlmock.AnyArg(),
		).
		WillReturnResult(
			sqlmock.NewResult(1, 1),
		)

	mock.ExpectExec(
		"INSERT INTO user_daily_access",
	).
		WithArgs(
			uint(11),
			sqlmock.AnyArg(),
		).
		WillReturnResult(
			sqlmock.NewResult(1, 1),
		)

	mock.ExpectQuery(
		"SELECT[\\s\\S]*uu.user_id",
	).
		WithArgs(uint(11)).
		WillReturnError(expectedErr)

	status, err := repo.GetClientAccessStatus(
		context.Background(),
		11,
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected query error, got %v",
			err,
		)
	}

	if status != nil {
		t.Fatalf("expected nil status")
	}
}

func TestSQLRepository_IncrementApplicationsToday_Success(t *testing.T) {
	db, mock := newSubscriptionAccessTestDB(t)

	repo := NewSQLRepository(db)

	mock.ExpectExec(
		"INSERT INTO user_daily_access",
	).
		WithArgs(
			uint(7),
			sqlmock.AnyArg(),
		).
		WillReturnResult(
			sqlmock.NewResult(1, 1),
		)

	mock.ExpectExec(
		"UPDATE user_daily_access",
	).
		WithArgs(
			sqlmock.AnyArg(),
			uint(7),
		).
		WillReturnResult(
			sqlmock.NewResult(0, 1),
		)

	err := repo.IncrementApplicationsToday(
		context.Background(),
		7,
	)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestSQLRepository_IncrementApplicationsToday_EnsureError(
	t *testing.T,
) {
	db, mock := newSubscriptionAccessTestDB(t)

	repo := NewSQLRepository(db)

	expectedErr := errors.New("ensure failed")

	mock.ExpectExec(
		"INSERT INTO user_daily_access",
	).
		WithArgs(
			uint(7),
			sqlmock.AnyArg(),
		).
		WillReturnError(expectedErr)

	err := repo.IncrementApplicationsToday(
		context.Background(),
		7,
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected ensure error, got %v",
			err,
		)
	}
}

func TestSQLRepository_IncrementApplicationsToday_UpdateError(
	t *testing.T,
) {
	db, mock := newSubscriptionAccessTestDB(t)

	repo := NewSQLRepository(db)

	expectedErr := errors.New("update failed")

	mock.ExpectExec(
		"INSERT INTO user_daily_access",
	).
		WithArgs(
			uint(7),
			sqlmock.AnyArg(),
		).
		WillReturnResult(
			sqlmock.NewResult(1, 1),
		)

	mock.ExpectExec(
		"UPDATE user_daily_access",
	).
		WithArgs(
			sqlmock.AnyArg(),
			uint(7),
		).
		WillReturnError(expectedErr)

	err := repo.IncrementApplicationsToday(
		context.Background(),
		7,
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected update error, got %v",
			err,
		)
	}
}

func TestSQLRepository_IncrementJobsPostToday_Success(t *testing.T) {
	db, mock := newSubscriptionAccessTestDB(t)

	repo := NewSQLRepository(db)

	mock.ExpectExec(
		"INSERT INTO user_daily_access",
	).
		WithArgs(
			uint(11),
			sqlmock.AnyArg(),
		).
		WillReturnResult(
			sqlmock.NewResult(1, 1),
		)

	mock.ExpectExec(
		"UPDATE user_daily_access",
	).
		WithArgs(
			sqlmock.AnyArg(),
			uint(11),
		).
		WillReturnResult(
			sqlmock.NewResult(0, 1),
		)

	err := repo.IncrementJobsPostToday(
		context.Background(),
		11,
	)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestSQLRepository_IncrementJobsPostToday_EnsureError(
	t *testing.T,
) {
	db, mock := newSubscriptionAccessTestDB(t)

	repo := NewSQLRepository(db)

	expectedErr := errors.New("ensure failed")

	mock.ExpectExec(
		"INSERT INTO user_daily_access",
	).
		WithArgs(
			uint(11),
			sqlmock.AnyArg(),
		).
		WillReturnError(expectedErr)

	err := repo.IncrementJobsPostToday(
		context.Background(),
		11,
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected ensure error, got %v",
			err,
		)
	}
}

func TestSQLRepository_IncrementJobsPostToday_UpdateError(
	t *testing.T,
) {
	db, mock := newSubscriptionAccessTestDB(t)

	repo := NewSQLRepository(db)

	expectedErr := errors.New("update failed")

	mock.ExpectExec(
		"INSERT INTO user_daily_access",
	).
		WithArgs(
			uint(11),
			sqlmock.AnyArg(),
		).
		WillReturnResult(
			sqlmock.NewResult(1, 1),
		)

	mock.ExpectExec(
		"UPDATE user_daily_access",
	).
		WithArgs(
			sqlmock.AnyArg(),
			uint(11),
		).
		WillReturnError(expectedErr)

	err := repo.IncrementJobsPostToday(
		context.Background(),
		11,
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected update error, got %v",
			err,
		)
	}
}
