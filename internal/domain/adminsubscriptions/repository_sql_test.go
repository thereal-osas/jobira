package adminsubscriptions

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestNewSQLRepository(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	if repo == nil {
		t.Fatal("expected repository")
	}

	if repo.db != db {
		t.Fatal("expected database assigned")
	}
}

func TestSQLRepository_ListPlans_Success(
	t *testing.T,
) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)
	now := time.Now()

	mock.ExpectQuery(
		"SELECT",
	).WillReturnRows(
		sqlmock.NewRows([]string{
			"id",
			"name",
			"role_type",
			"price_pence",
			"billing_interval",
			"application_limit",
			"job_post_limit",
			"cleaner_seat_limit",
			"is_active",
			"created_at",
			"updated_at",
		}).AddRow(
			1,
			"Standard",
			"cleaner",
			1799,
			"monthly",
			50,
			0,
			1,
			true,
			now,
			now,
		),
	)

	plans, err := repo.ListPlans(
		context.Background(),
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(plans) != 1 {
		t.Fatalf(
			"expected 1 plan, got %d",
			len(plans),
		)
	}

	if plans[0].Name != "Standard" {
		t.Fatalf(
			"expected Standard, got %q",
			plans[0].Name,
		)
	}
}

func TestSQLRepository_ListPlans_QueryError(
	t *testing.T,
) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)
	expectedErr := errors.New("query failed")

	mock.ExpectQuery("SELECT").
		WillReturnError(expectedErr)

	plans, err := repo.ListPlans(
		context.Background(),
	)

	if plans != nil {
		t.Fatal("expected nil plans")
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected %v, got %v",
			expectedErr,
			err,
		)
	}
}

func TestSQLRepository_ListPlans_ScanError(
	t *testing.T,
) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	mock.ExpectQuery("SELECT").
		WillReturnRows(
			sqlmock.NewRows([]string{
				"id",
				"name",
				"role_type",
				"price_pence",
				"billing_interval",
				"application_limit",
				"job_post_limit",
				"cleaner_seat_limit",
				"is_active",
				"created_at",
				"updated_at",
			}).AddRow(
				"invalid",
				"Standard",
				"cleaner",
				1799,
				"monthly",
				50,
				0,
				1,
				true,
				time.Now(),
				time.Now(),
			),
		)

	_, err = repo.ListPlans(
		context.Background(),
	)

	if err == nil {
		t.Fatal("expected scan error")
	}
}

func TestSQLRepository_CreatePlan_Success(
	t *testing.T,
) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	now := time.Now()

	mock.ExpectQuery(
		regexp.QuoteMeta(
			"INSERT INTO subscription_plans",
		),
	).
		WithArgs(
			"Standard",
			"cleaner",
			1799,
			"monthly",
			50,
			0,
			1,
			sqlmock.AnyArg(),
		).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"id",
				"is_active",
				"created_at",
				"updated_at",
			}).AddRow(
				10,
				true,
				now,
				now,
			),
		)

	plan := &AdminSubscriptionPlan{
		Name:             "Standard",
		RoleType:         "cleaner",
		PricePence:       1799,
		BillingInterval:  "monthly",
		ApplicationLimit: 50,
		JobPostLimit:     0,
		CleanerSeatLimit: 1,
	}

	err = repo.CreatePlan(
		context.Background(),
		plan,
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if plan.ID != 10 {
		t.Fatalf(
			"expected ID 10, got %d",
			plan.ID,
		)
	}

	if !plan.IsActive {
		t.Fatal("expected active plan")
	}
}

func TestSQLRepository_CreatePlan_Error(
	t *testing.T,
) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	expectedErr := errors.New("create failed")

	mock.ExpectQuery(
		"INSERT INTO subscription_plans",
	).WillReturnError(expectedErr)

	plan := &AdminSubscriptionPlan{
		Name:             "Standard",
		RoleType:         "cleaner",
		BillingInterval:  "monthly",
		CleanerSeatLimit: 1,
	}

	err = repo.CreatePlan(
		context.Background(),
		plan,
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected %v, got %v",
			expectedErr,
			err,
		)
	}
}

func TestSQLRepository_UpdatePlan_Success(
	t *testing.T,
) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	mock.ExpectExec(
		"UPDATE subscription_plans",
	).
		WithArgs(
			"Premium",
			"cleaner",
			2499,
			"monthly",
			100,
			0,
			1,
			true,
			sqlmock.AnyArg(),
			uint(2),
		).
		WillReturnResult(
			sqlmock.NewResult(0, 1),
		)

	err = repo.UpdatePlan(
		context.Background(),
		&AdminSubscriptionPlan{
			ID:               2,
			Name:             "Premium",
			RoleType:         "cleaner",
			PricePence:       2499,
			BillingInterval:  "monthly",
			ApplicationLimit: 100,
			JobPostLimit:     0,
			CleanerSeatLimit: 1,
			IsActive:         true,
		},
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSQLRepository_UpdatePlan_NotFound(
	t *testing.T,
) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	mock.ExpectExec(
		"UPDATE subscription_plans",
	).WillReturnResult(
		sqlmock.NewResult(0, 0),
	)

	err = repo.UpdatePlan(
		context.Background(),
		&AdminSubscriptionPlan{
			ID:               999,
			Name:             "Plan",
			RoleType:         "cleaner",
			BillingInterval:  "monthly",
			CleanerSeatLimit: 1,
		},
	)

	if !errors.Is(
		err,
		ErrPlanNotFound,
	) {
		t.Fatalf(
			"expected ErrPlanNotFound, got %v",
			err,
		)
	}
}

func TestSQLRepository_UpdatePlan_Error(
	t *testing.T,
) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	expectedErr := errors.New("update failed")

	mock.ExpectExec(
		"UPDATE subscription_plans",
	).WillReturnError(expectedErr)

	err = repo.UpdatePlan(
		context.Background(),
		&AdminSubscriptionPlan{
			ID:               2,
			Name:             "Plan",
			RoleType:         "cleaner",
			BillingInterval:  "monthly",
			CleanerSeatLimit: 1,
		},
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected %v, got %v",
			expectedErr,
			err,
		)
	}
}

func TestSQLRepository_DisablePlan_Success(
	t *testing.T,
) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	mock.ExpectExec(
		"UPDATE subscription_plans",
	).
		WithArgs(
			sqlmock.AnyArg(),
			uint(2),
		).
		WillReturnResult(
			sqlmock.NewResult(0, 1),
		)

	err = repo.DisablePlan(
		context.Background(),
		2,
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSQLRepository_DisablePlan_NotFound(
	t *testing.T,
) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	mock.ExpectExec(
		"UPDATE subscription_plans",
	).WillReturnResult(
		sqlmock.NewResult(0, 0),
	)

	err = repo.DisablePlan(
		context.Background(),
		999,
	)

	if !errors.Is(
		err,
		ErrPlanNotFound,
	) {
		t.Fatalf(
			"expected ErrPlanNotFound, got %v",
			err,
		)
	}
}

func TestSQLRepository_ListUserSubscriptions_Success(
	t *testing.T,
) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	now := time.Now()
	planID := 2
	planName := "Premium"
	trialEnd := now.AddDate(0, 0, 7)
	periodStart := now
	periodEnd := now.AddDate(0, 1, 0)

	mock.ExpectQuery(
		"SELECT",
	).WillReturnRows(
		sqlmock.NewRows([]string{
			"id",
			"user_id",
			"email",
			"full_name",
			"plan_id",
			"plan_name",
			"status",
			"trial_started_at",
			"trial_ends_at",
			"current_period_start",
			"current_period_end",
			"created_at",
			"updated_at",
		}).AddRow(
			1,
			8,
			"cleaner@example.com",
			"Cleaner User",
			planID,
			planName,
			"active",
			now,
			trialEnd,
			periodStart,
			periodEnd,
			now,
			now,
		),
	)

	subs, err := repo.ListUserSubscriptions(
		context.Background(),
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(subs) != 1 {
		t.Fatalf(
			"expected 1 subscription, got %d",
			len(subs),
		)
	}

	if subs[0].PlanID == nil ||
		*subs[0].PlanID != 2 {
		t.Fatal("expected plan ID 2")
	}

	if subs[0].PlanName == nil ||
		*subs[0].PlanName != "Premium" {
		t.Fatal("expected Premium plan")
	}

	if subs[0].TrialEndsAt == nil {
		t.Fatal("expected trial end")
	}

	if subs[0].CurrentPeriodStart == nil {
		t.Fatal("expected current period start")
	}

	if subs[0].CurrentPeriodEnd == nil {
		t.Fatal("expected current period end")
	}
}

func TestSQLRepository_ListUserSubscriptions_Nulls(
	t *testing.T,
) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)
	now := time.Now()

	mock.ExpectQuery(
		"SELECT",
	).WillReturnRows(
		sqlmock.NewRows([]string{
			"id",
			"user_id",
			"email",
			"full_name",
			"plan_id",
			"plan_name",
			"status",
			"trial_started_at",
			"trial_ends_at",
			"current_period_start",
			"current_period_end",
			"created_at",
			"updated_at",
		}).AddRow(
			1,
			8,
			"user@example.com",
			"User",
			nil,
			nil,
			"trial",
			now,
			nil,
			nil,
			nil,
			now,
			now,
		),
	)

	subs, err := repo.ListUserSubscriptions(
		context.Background(),
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if subs[0].PlanID != nil {
		t.Fatal("expected nil plan ID")
	}

	if subs[0].PlanName != nil {
		t.Fatal("expected nil plan name")
	}

	if subs[0].TrialEndsAt != nil {
		t.Fatal("expected nil trial end")
	}
}

func TestSQLRepository_ListUserSubscriptions_Error(
	t *testing.T,
) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)
	expectedErr := errors.New("query failed")

	mock.ExpectQuery("SELECT").
		WillReturnError(expectedErr)

	_, err = repo.ListUserSubscriptions(
		context.Background(),
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected %v, got %v",
			expectedErr,
			err,
		)
	}
}

func TestSQLRepository_GetUserSubscriptions_Success(
	t *testing.T,
) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)
	now := time.Now()

	mock.ExpectQuery("SELECT").
		WithArgs(uint(8)).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"id",
				"user_id",
				"email",
				"full_name",
				"plan_id",
				"plan_name",
				"status",
				"trial_started_at",
				"trial_ends_at",
				"current_period_start",
				"current_period_end",
				"created_at",
				"updated_at",
			}).AddRow(
				1,
				8,
				"user@example.com",
				"User",
				2,
				"Premium",
				"active",
				now,
				nil,
				nil,
				nil,
				now,
				now,
			),
		)

	sub, err := repo.GetUserSubscriptions(
		context.Background(),
		8,
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if sub.UserID != 8 {
		t.Fatalf(
			"expected user ID 8, got %d",
			sub.UserID,
		)
	}
}

func TestSQLRepository_GetUserSubscriptions_NotFound(
	t *testing.T,
) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	mock.ExpectQuery("SELECT").
		WithArgs(uint(999)).
		WillReturnError(sql.ErrNoRows)

	sub, err := repo.GetUserSubscriptions(
		context.Background(),
		999,
	)

	if sub != nil {
		t.Fatal("expected nil subscription")
	}

	if !errors.Is(
		err,
		ErrSubscriptionNotFound,
	) {
		t.Fatalf(
			"expected ErrSubscriptionNotFound, got %v",
			err,
		)
	}
}

func TestSQLRepository_GetUserSubscriptions_DatabaseError(
	t *testing.T,
) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	expectedErr := errors.New("database failed")

	mock.ExpectQuery("SELECT").
		WithArgs(uint(8)).
		WillReturnError(expectedErr)

	sub, err := repo.GetUserSubscriptions(
		context.Background(),
		8,
	)

	if sub != nil {
		t.Fatal("expected nil subscription")
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected %v, got %v",
			expectedErr,
			err,
		)
	}
}

func TestSQLRepository_CreateOrUpdateUserSubscription(
	t *testing.T,
) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	mock.ExpectExec(
		"INSERT INTO user_subscriptions",
	).
		WithArgs(
			uint(8),
			uint(2),
			"active",
			sqlmock.AnyArg(),
		).
		WillReturnResult(
			sqlmock.NewResult(1, 1),
		)

	err = repo.CreateOrUpdateUserSubscription(
		context.Background(),
		8,
		2,
		"active",
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
