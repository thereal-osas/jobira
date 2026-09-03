package subscriptions

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func newSubscriptionRepositoryTest(
	t *testing.T,
) (*SQLRepository, sqlmock.Sqlmock) {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf(
			"failed creating sqlmock: %v",
			err,
		)
	}

	return NewSQLRepository(db), mock
}

func planColumns() []string {
	return []string{
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
	}
}

func subscriptionColumns() []string {
	return []string{
		"id",
		"user_id",
		"plan_id",
		"status",
		"trial_started_at",
		"trial_ends_at",
		"current_period_start",
		"current_period_end",
		"created_at",
		"updated_at",
	}
}

func TestNewSQLRepository(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repository := NewSQLRepository(db)

	if repository == nil {
		t.Fatal("expected repository")
	}

	if repository.db != db {
		t.Fatal(
			"expected supplied database",
		)
	}
}

func TestSQLRepository_ListPlans_Success(
	t *testing.T,
) {
	repository, mock :=
		newSubscriptionRepositoryTest(t)

	now := time.Now()

	mock.ExpectQuery(
		`(?s)SELECT.*FROM subscription_plans.*WHERE is_active = true.*ORDER BY price_pence ASC`,
	).
		WillReturnRows(
			sqlmock.NewRows(
				planColumns(),
			).
				AddRow(
					1,
					"Launch",
					"cleaner",
					1500,
					"monthly",
					5,
					0,
					1,
					true,
					now,
					now,
				).
				AddRow(
					2,
					"Standard",
					"cleaner",
					1799,
					"monthly",
					100,
					0,
					1,
					true,
					now,
					now,
				),
		)

	plans, err := repository.ListPlans(
		context.Background(),
	)

	if err != nil {
		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}

	if len(plans) != 2 {
		t.Fatalf(
			"expected 2 plans got %d",
			len(plans),
		)
	}
}

func TestSQLRepository_ListPlans_QueryError(
	t *testing.T,
) {
	repository, mock :=
		newSubscriptionRepositoryTest(t)

	queryErr := errors.New("query failed")

	mock.ExpectQuery(
		`(?s)SELECT.*FROM subscription_plans`,
	).
		WillReturnError(queryErr)

	_, err := repository.ListPlans(
		context.Background(),
	)

	if !errors.Is(err, queryErr) {
		t.Fatalf(
			"expected query error got %v",
			err,
		)
	}
}

func TestSQLRepository_ListPlans_ScanError(
	t *testing.T,
) {
	repository, mock :=
		newSubscriptionRepositoryTest(t)

	now := time.Now()

	mock.ExpectQuery(
		`(?s)SELECT.*FROM subscription_plans`,
	).
		WillReturnRows(
			sqlmock.NewRows(
				planColumns(),
			).AddRow(
				"bad-id",
				"Launch",
				"cleaner",
				1500,
				"monthly",
				5,
				0,
				1,
				true,
				now,
				now,
			),
		)

	_, err := repository.ListPlans(
		context.Background(),
	)

	if err == nil {
		t.Fatal("expected scan error")
	}
}

func TestSQLRepository_ListPlans_RowsError(
	t *testing.T,
) {
	repository, mock :=
		newSubscriptionRepositoryTest(t)

	now := time.Now()
	rowsErr := errors.New("rows failed")

	rows := sqlmock.NewRows(
		planColumns(),
	).
		AddRow(
			1,
			"Launch",
			"cleaner",
			1500,
			"monthly",
			5,
			0,
			1,
			true,
			now,
			now,
		).
		RowError(0, rowsErr)

	mock.ExpectQuery(
		`(?s)SELECT.*FROM subscription_plans`,
	).
		WillReturnRows(rows)

	_, err := repository.ListPlans(
		context.Background(),
	)

	if !errors.Is(err, rowsErr) {
		t.Fatalf(
			"expected rows error got %v",
			err,
		)
	}
}

func TestSQLRepository_GetPlanByID_Success(
	t *testing.T,
) {
	repository, mock :=
		newSubscriptionRepositoryTest(t)

	now := time.Now()

	mock.ExpectQuery(
		`(?s)SELECT.*FROM subscription_plans.*WHERE id = \$1`,
	).
		WithArgs(uint(2)).
		WillReturnRows(
			sqlmock.NewRows(
				planColumns(),
			).AddRow(
				2,
				"Standard",
				"cleaner",
				1799,
				"monthly",
				100,
				0,
				1,
				true,
				now,
				now,
			),
		)

	plan, err := repository.GetPlanByID(
		context.Background(),
		2,
	)

	if err != nil {
		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}

	if plan.ID != 2 {
		t.Fatalf(
			"expected ID 2 got %d",
			plan.ID,
		)
	}
}

func TestSQLRepository_GetPlanByID_NotFound(
	t *testing.T,
) {
	repository, mock :=
		newSubscriptionRepositoryTest(t)

	mock.ExpectQuery(
		`(?s)SELECT.*FROM subscription_plans.*WHERE id = \$1`,
	).
		WithArgs(uint(999)).
		WillReturnError(sql.ErrNoRows)

	_, err := repository.GetPlanByID(
		context.Background(),
		999,
	)

	if !errors.Is(err, ErrPlanNotFound) {
		t.Fatalf(
			"expected ErrPlanNotFound got %v",
			err,
		)
	}
}

func TestSQLRepository_GetPlanByID_DatabaseError(
	t *testing.T,
) {
	repository, mock :=
		newSubscriptionRepositoryTest(t)

	databaseErr := errors.New(
		"database failed",
	)

	mock.ExpectQuery(
		`(?s)SELECT.*FROM subscription_plans.*WHERE id = \$1`,
	).
		WithArgs(uint(2)).
		WillReturnError(databaseErr)

	_, err := repository.GetPlanByID(
		context.Background(),
		2,
	)

	if !errors.Is(err, databaseErr) {
		t.Fatalf(
			"expected database error got %v",
			err,
		)
	}
}

func TestSQLRepository_CreateOrUpdateUserSubscription_Success(
	t *testing.T,
) {
	repository, mock :=
		newSubscriptionRepositoryTest(t)

	now := time.Now()
	trialEnds := now.Add(30 * 24 * time.Hour)
	periodStart := now
	periodEnd := now.Add(30 * 24 * time.Hour)

	mock.ExpectQuery(
		`(?s)INSERT INTO user_subscriptions.*ON CONFLICT.*RETURNING`,
	).
		WithArgs(
			uint(10),
			uint(2),
			"trial",
			sqlmock.AnyArg(),
		).
		WillReturnRows(
			sqlmock.NewRows(
				subscriptionColumns(),
			).AddRow(
				5,
				10,
				2,
				"trial",
				now,
				trialEnds,
				periodStart,
				periodEnd,
				now,
				now,
			),
		)

	subscription, err :=
		repository.CreateOrUpdateUserSubscription(
			context.Background(),
			10,
			2,
			"trial",
		)

	if err != nil {
		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}

	if subscription.ID != 5 {
		t.Fatalf(
			"expected ID 5 got %d",
			subscription.ID,
		)
	}

	if subscription.PlanID == nil ||
		*subscription.PlanID != 2 {
		t.Fatal("expected plan ID 2")
	}

	if subscription.TrialEndsAt == nil {
		t.Fatal(
			"expected TrialEndsAt",
		)
	}

	if subscription.CurrentPeriodStart == nil {
		t.Fatal(
			"expected CurrentPeriodStart",
		)
	}

	if subscription.CurrentPeriodEnd == nil {
		t.Fatal(
			"expected CurrentPeriodEnd",
		)
	}
}

func TestSQLRepository_CreateOrUpdateUserSubscription_NullOptionalFields(
	t *testing.T,
) {
	repository, mock :=
		newSubscriptionRepositoryTest(t)

	now := time.Now()

	mock.ExpectQuery(
		`(?s)INSERT INTO user_subscriptions.*ON CONFLICT.*RETURNING`,
	).
		WithArgs(
			uint(10),
			uint(2),
			"trial",
			sqlmock.AnyArg(),
		).
		WillReturnRows(
			sqlmock.NewRows(
				subscriptionColumns(),
			).AddRow(
				5,
				10,
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

	subscription, err :=
		repository.CreateOrUpdateUserSubscription(
			context.Background(),
			10,
			2,
			"trial",
		)

	if err != nil {
		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}

	if subscription.PlanID != nil {
		t.Fatal(
			"expected nil PlanID",
		)
	}

	if subscription.TrialEndsAt != nil {
		t.Fatal(
			"expected nil TrialEndsAt",
		)
	}

	if subscription.CurrentPeriodStart != nil {
		t.Fatal(
			"expected nil CurrentPeriodStart",
		)
	}

	if subscription.CurrentPeriodEnd != nil {
		t.Fatal(
			"expected nil CurrentPeriodEnd",
		)
	}
}

func TestSQLRepository_CreateOrUpdateUserSubscription_Error(
	t *testing.T,
) {
	repository, mock :=
		newSubscriptionRepositoryTest(t)

	createErr := errors.New("create failed")

	mock.ExpectQuery(
		`(?s)INSERT INTO user_subscriptions.*ON CONFLICT`,
	).
		WithArgs(
			uint(10),
			uint(2),
			"trial",
			sqlmock.AnyArg(),
		).
		WillReturnError(createErr)

	_, err :=
		repository.CreateOrUpdateUserSubscription(
			context.Background(),
			10,
			2,
			"trial",
		)

	if !errors.Is(err, createErr) {
		t.Fatalf(
			"expected create error got %v",
			err,
		)
	}
}

func TestSQLRepository_GetByUserID_Success(
	t *testing.T,
) {
	repository, mock :=
		newSubscriptionRepositoryTest(t)

	now := time.Now()

	mock.ExpectQuery(
		`(?s)SELECT.*FROM user_subscriptions.*WHERE user_id = \$1`,
	).
		WithArgs(uint(10)).
		WillReturnRows(
			sqlmock.NewRows(
				subscriptionColumns(),
			).AddRow(
				5,
				10,
				2,
				"active",
				now,
				nil,
				nil,
				nil,
				now,
				now,
			),
		)

	subscription, err := repository.GetByUserID(
		context.Background(),
		10,
	)

	if err != nil {
		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}

	if subscription.UserID != 10 {
		t.Fatalf(
			"expected user 10 got %d",
			subscription.UserID,
		)
	}

	if subscription.PlanID == nil ||
		*subscription.PlanID != 2 {
		t.Fatal("expected plan ID 2")
	}
}

func TestSQLRepository_GetByUserID_NotFound(
	t *testing.T,
) {
	repository, mock :=
		newSubscriptionRepositoryTest(t)

	mock.ExpectQuery(
		`(?s)SELECT.*FROM user_subscriptions.*WHERE user_id = \$1`,
	).
		WithArgs(uint(999)).
		WillReturnError(sql.ErrNoRows)

	_, err := repository.GetByUserID(
		context.Background(),
		999,
	)

	if !errors.Is(
		err,
		ErrSubscriptionNotFound,
	) {
		t.Fatalf(
			"expected not found got %v",
			err,
		)
	}
}

func TestSQLRepository_GetByUserID_DatabaseError(
	t *testing.T,
) {
	repository, mock :=
		newSubscriptionRepositoryTest(t)

	databaseErr := errors.New(
		"database failed",
	)

	mock.ExpectQuery(
		`(?s)SELECT.*FROM user_subscriptions.*WHERE user_id = \$1`,
	).
		WithArgs(uint(10)).
		WillReturnError(databaseErr)

	_, err := repository.GetByUserID(
		context.Background(),
		10,
	)

	if !errors.Is(err, databaseErr) {
		t.Fatalf(
			"expected database error got %v",
			err,
		)
	}
}

func TestSQLRepository_UpdateStatus_Success(
	t *testing.T,
) {
	repository, mock :=
		newSubscriptionRepositoryTest(t)

	mock.ExpectExec(
		`(?s)UPDATE user_subscriptions.*SET status = \$1.*WHERE user_id = \$3`,
	).
		WithArgs(
			"active",
			sqlmock.AnyArg(),
			uint(10),
		).
		WillReturnResult(
			sqlmock.NewResult(0, 1),
		)

	err := repository.UpdateStatus(
		context.Background(),
		10,
		"active",
	)

	if err != nil {
		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}
}

func TestSQLRepository_UpdateStatus_NotFound(
	t *testing.T,
) {
	repository, mock :=
		newSubscriptionRepositoryTest(t)

	mock.ExpectExec(
		`(?s)UPDATE user_subscriptions.*WHERE user_id = \$3`,
	).
		WithArgs(
			"active",
			sqlmock.AnyArg(),
			uint(999),
		).
		WillReturnResult(
			sqlmock.NewResult(0, 0),
		)

	err := repository.UpdateStatus(
		context.Background(),
		999,
		"active",
	)

	if !errors.Is(
		err,
		ErrSubscriptionNotFound,
	) {
		t.Fatalf(
			"expected not found got %v",
			err,
		)
	}
}

func TestSQLRepository_UpdateStatus_DatabaseError(
	t *testing.T,
) {
	repository, mock :=
		newSubscriptionRepositoryTest(t)

	updateErr := errors.New("update failed")

	mock.ExpectExec(
		`(?s)UPDATE user_subscriptions.*WHERE user_id = \$3`,
	).
		WithArgs(
			"active",
			sqlmock.AnyArg(),
			uint(10),
		).
		WillReturnError(updateErr)

	err := repository.UpdateStatus(
		context.Background(),
		10,
		"active",
	)

	if !errors.Is(err, updateErr) {
		t.Fatalf(
			"expected update error got %v",
			err,
		)
	}
}

var _ Repository = (*SQLRepository)(nil)
