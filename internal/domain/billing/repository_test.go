package billing

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func newBillingRepositoryTest(
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
		t.Fatal("expected supplied database")
	}
}

func TestSQLRepository_GetPlanByID_Success(t *testing.T) {
	repo, mock := newBillingRepositoryTest(t)

	mock.ExpectQuery(
		`(?s)SELECT.*FROM subscription_plans.*WHERE id = \$1.*is_active = true`,
	).
		WithArgs(uint(2)).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{
					"id",
					"name",
					"role_type",
					"price_pence",
					"stripe_price_id",
				},
			).AddRow(
				2,
				"Standard",
				"cleaner",
				1799,
				"price_123",
			),
		)

	plan, err := repo.GetPlanByID(
		context.Background(),
		2,
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if plan.ID != 2 {
		t.Fatalf(
			"expected ID 2 got %d",
			plan.ID,
		)
	}
}

func TestSQLRepository_GetPlanByID_NotFound(t *testing.T) {
	repo, mock := newBillingRepositoryTest(t)

	mock.ExpectQuery(
		`(?s)SELECT.*FROM subscription_plans.*WHERE id = \$1`,
	).
		WithArgs(uint(999)).
		WillReturnError(sql.ErrNoRows)

	_, err := repo.GetPlanByID(
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

func TestSQLRepository_GetPlanByID_Error(t *testing.T) {
	repo, mock := newBillingRepositoryTest(t)

	dbErr := errors.New("database failed")

	mock.ExpectQuery(
		`(?s)SELECT.*FROM subscription_plans.*WHERE id = \$1`,
	).
		WithArgs(uint(2)).
		WillReturnError(dbErr)

	_, err := repo.GetPlanByID(
		context.Background(),
		2,
	)

	if !errors.Is(err, dbErr) {
		t.Fatalf(
			"expected database error got %v",
			err,
		)
	}
}

func TestSQLRepository_GetUserEmail_Success(t *testing.T) {
	repo, mock := newBillingRepositoryTest(t)

	mock.ExpectQuery(
		`(?s)SELECT email.*FROM users.*WHERE id = \$1`,
	).
		WithArgs(uint(5)).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{"email"},
			).AddRow("user@example.com"),
		)

	email, err := repo.GetUserEmail(
		context.Background(),
		5,
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if email != "user@example.com" {
		t.Fatalf(
			"unexpected email %q",
			email,
		)
	}
}

func TestSQLRepository_GetUserEmail_Error(t *testing.T) {
	repo, mock := newBillingRepositoryTest(t)

	dbErr := errors.New("database failed")

	mock.ExpectQuery(
		`(?s)SELECT email.*FROM users.*WHERE id = \$1`,
	).
		WithArgs(uint(5)).
		WillReturnError(dbErr)

	_, err := repo.GetUserEmail(
		context.Background(),
		5,
	)

	if !errors.Is(err, dbErr) {
		t.Fatalf(
			"expected database error got %v",
			err,
		)
	}
}

func TestSQLRepository_GetStripeCustomerID_Success(
	t *testing.T,
) {
	repo, mock := newBillingRepositoryTest(t)

	mock.ExpectQuery(
		`(?s)SELECT stripe_customer_id.*FROM billing_customers.*WHERE user_id = \$1`,
	).
		WithArgs(uint(5)).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{"stripe_customer_id"},
			).AddRow("cus_123"),
		)

	id, err := repo.GetStripeCustomerID(
		context.Background(),
		5,
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if id != "cus_123" {
		t.Fatalf(
			"expected cus_123 got %q",
			id,
		)
	}
}

func TestSQLRepository_GetStripeCustomerID_NotFound(
	t *testing.T,
) {
	repo, mock := newBillingRepositoryTest(t)

	mock.ExpectQuery(
		`(?s)SELECT stripe_customer_id.*FROM billing_customers.*WHERE user_id = \$1`,
	).
		WithArgs(uint(5)).
		WillReturnError(sql.ErrNoRows)

	_, err := repo.GetStripeCustomerID(
		context.Background(),
		5,
	)

	if !errors.Is(err, ErrCustomerNotFound) {
		t.Fatalf(
			"expected ErrCustomerNotFound got %v",
			err,
		)
	}
}

func TestSQLRepository_GetStripeCustomerID_Error(
	t *testing.T,
) {
	repo, mock := newBillingRepositoryTest(t)

	dbErr := errors.New("database failed")

	mock.ExpectQuery(
		`(?s)SELECT stripe_customer_id.*FROM billing_customers.*WHERE user_id = \$1`,
	).
		WithArgs(uint(5)).
		WillReturnError(dbErr)

	_, err := repo.GetStripeCustomerID(
		context.Background(),
		5,
	)

	if !errors.Is(err, dbErr) {
		t.Fatalf(
			"expected database error got %v",
			err,
		)
	}
}

func TestSQLRepository_UpsertBillingCustomer_Success(
	t *testing.T,
) {
	repo, mock := newBillingRepositoryTest(t)

	mock.ExpectExec(
		`(?s)INSERT INTO billing_customers.*ON CONFLICT \(user_id\).*DO UPDATE`,
	).
		WithArgs(
			uint(5),
			"user@example.com",
			"cus_123",
			sqlmock.AnyArg(),
		).
		WillReturnResult(
			sqlmock.NewResult(1, 1),
		)

	err := repo.UpsertBillingCustomer(
		context.Background(),
		5,
		"user@example.com",
		"cus_123",
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSQLRepository_UpsertBillingCustomer_Error(
	t *testing.T,
) {
	repo, mock := newBillingRepositoryTest(t)

	dbErr := errors.New("insert failed")

	mock.ExpectExec(
		`(?s)INSERT INTO billing_customers`,
	).
		WithArgs(
			uint(5),
			"user@example.com",
			"cus_123",
			sqlmock.AnyArg(),
		).
		WillReturnError(dbErr)

	err := repo.UpsertBillingCustomer(
		context.Background(),
		5,
		"user@example.com",
		"cus_123",
	)

	if !errors.Is(err, dbErr) {
		t.Fatalf(
			"expected insert error got %v",
			err,
		)
	}
}

func TestSQLRepository_SaveCheckoutSession_Success(
	t *testing.T,
) {
	repo, mock := newBillingRepositoryTest(t)

	now := time.Now()

	record := &CheckoutSessionRecord{
		UserID:               5,
		PlanID:               2,
		StripeSessionID:      "cs_123",
		StripeCustomerID:     "cus_123",
		StripeSubscriptionID: "",
		Status:               "created",
	}

	mock.ExpectQuery(
		`(?s)INSERT INTO billing_checkout_sessions.*RETURNING id, created_at, updated_at`,
	).
		WithArgs(
			record.UserID,
			record.PlanID,
			record.StripeSessionID,
			record.StripeCustomerID,
			record.StripeSubscriptionID,
			record.Status,
			sqlmock.AnyArg(),
		).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{
					"id",
					"created_at",
					"updated_at",
				},
			).AddRow(
				10,
				now,
				now,
			),
		)

	err := repo.SaveCheckoutSession(
		context.Background(),
		record,
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if record.ID != 10 {
		t.Fatalf(
			"expected ID 10 got %d",
			record.ID,
		)
	}
}

func TestSQLRepository_SaveCheckoutSession_Error(
	t *testing.T,
) {
	repo, mock := newBillingRepositoryTest(t)

	record := &CheckoutSessionRecord{
		UserID:          5,
		PlanID:          2,
		StripeSessionID: "cs_123",
		Status:          "created",
	}

	dbErr := errors.New("insert failed")

	mock.ExpectQuery(
		`(?s)INSERT INTO billing_checkout_sessions`,
	).
		WithArgs(
			record.UserID,
			record.PlanID,
			record.StripeSessionID,
			record.StripeCustomerID,
			record.StripeSubscriptionID,
			record.Status,
			sqlmock.AnyArg(),
		).
		WillReturnError(dbErr)

	err := repo.SaveCheckoutSession(
		context.Background(),
		record,
	)

	if !errors.Is(err, dbErr) {
		t.Fatalf(
			"expected insert error got %v",
			err,
		)
	}
}

func TestSQLRepository_MarkCheckoutSessionComplete_Success(
	t *testing.T,
) {
	repo, mock := newBillingRepositoryTest(t)

	mock.ExpectExec(
		`(?s)UPDATE billing_checkout_sessions.*status = 'completed'.*WHERE stripe_session_id = \$4`,
	).
		WithArgs(
			"cus_123",
			"sub_123",
			sqlmock.AnyArg(),
			"cs_123",
		).
		WillReturnResult(
			sqlmock.NewResult(0, 1),
		)

	completed, err := repo.MarkCheckoutSessionComplete(
		context.Background(),
		"cs_123",
		"cus_123",
		"sub_123",
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !completed {
		t.Fatal("expected checkout session to be marked complete")
	}

}

func TestSQLRepository_MarkCheckoutSessionComplete_Duplicate(
	t *testing.T,
) {
	repo, mock := newBillingRepositoryTest(t)

	mock.ExpectExec(
		`(?s)UPDATE billing_checkout_sessions.*status = 'completed'.*WHERE stripe_session_id = \$4`,
	).
		WithArgs(
			"cus_123",
			"sub_123",
			sqlmock.AnyArg(),
			"cs_123",
		).
		WillReturnResult(
			sqlmock.NewResult(0, 0),
		)

	mock.ExpectQuery(
		`(?s)SELECT status.*FROM billing_checkout_sessions.*WHERE stripe_session_id = \$1`,
	).
		WithArgs("cs_123").
		WillReturnRows(
			sqlmock.NewRows(
				[]string{"status"},
			).AddRow("completed"),
		)

	completed, err := repo.MarkCheckoutSessionComplete(
		context.Background(),
		"cs_123",
		"cus_123",
		"sub_123",
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if completed {
		t.Fatal("expected duplicate checkout to return false")
	}
}

func TestSQLRepository_MarkCheckoutSessionComplete_NotFound(
	t *testing.T,
) {
	repo, mock := newBillingRepositoryTest(t)

	mock.ExpectExec(
		`(?s)UPDATE billing_checkout_sessions.*status = 'completed'.*WHERE stripe_session_id = \$4`,
	).
		WithArgs(
			"cus_123",
			"sub_123",
			sqlmock.AnyArg(),
			"cs_missing_123",
		).
		WillReturnResult(
			sqlmock.NewResult(0, 0),
		)

	mock.ExpectQuery(
		`(?s)SELECT status.*FROM billing_checkout_sessions.*WHERE stripe_session_id = \$1`,
	).
		WithArgs("cs_missing_123").
		WillReturnError(sql.ErrNoRows)

	completed, err := repo.MarkCheckoutSessionComplete(
		context.Background(),
		"cs_missing_123",
		"cus_123",
		"sub_123",
	)

	if !errors.Is(err, ErrCheckoutSessionNotFound) {
		t.Fatalf(
			"expected ErrCheckoutSessionNotFound got %v",
			err,
		)
	}

	if completed {
		t.Fatal("missing checkout must return false")
	}
}

func TestSQLRepository_MarkCheckoutSessionComplete_Error(
	t *testing.T,
) {
	repo, mock := newBillingRepositoryTest(t)

	dbErr := errors.New("update failed")

	mock.ExpectExec(
		`(?s)UPDATE billing_checkout_sessions`,
	).
		WithArgs(
			"cus_123",
			"sub_123",
			sqlmock.AnyArg(),
			"cs_123",
		).
		WillReturnError(dbErr)

	_, err := repo.MarkCheckoutSessionComplete(
		context.Background(),
		"cs_123",
		"cus_123",
		"sub_123",
	)

	if !errors.Is(err, dbErr) {
		t.Fatalf(
			"expected update error got %v",
			err,
		)
	}
}

func TestSQLRepository_ActivateUserSubscription_UpdateSuccess(
	t *testing.T,
) {
	repo, mock := newBillingRepositoryTest(t)

	periodStart := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
	periodEnd := periodStart.AddDate(0, 1, 0)

	mock.ExpectExec(
		`(?s)UPDATE user_subscriptions.*stripe_subscription_id = \$2.*WHERE user_id = \$6`,
	).
		WithArgs(
			uint(2),
			"sub_123",
			periodStart,
			periodEnd,
			sqlmock.AnyArg(),
			uint(5),
		).
		WillReturnResult(
			sqlmock.NewResult(0, 1),
		)

	err := repo.ActivateUserSubscription(
		context.Background(),
		5,
		2,
		"sub_123",
		periodStart,
		periodEnd,
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSQLRepository_ActivateUserSubscription_UpdateError(
	t *testing.T,
) {
	repo, mock := newBillingRepositoryTest(t)

	periodStart := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
	periodEnd := periodStart.AddDate(0, 1, 0)

	dbErr := errors.New("update failed")

	mock.ExpectExec(
		`(?s)UPDATE user_subscriptions`,
	).
		WithArgs(
			uint(2),
			"sub_123",
			periodStart,
			periodEnd,
			sqlmock.AnyArg(),
			uint(5),
		).
		WillReturnError(dbErr)

	err := repo.ActivateUserSubscription(
		context.Background(),
		5,
		2,
		"sub_123",
		periodStart,
		periodEnd,
	)

	if !errors.Is(err, dbErr) {
		t.Fatalf(
			"expected update error got %v",
			err,
		)
	}
}

func TestSQLRepository_ActivateUserSubscription_InsertFallback(
	t *testing.T,
) {
	repo, mock := newBillingRepositoryTest(t)

	periodStart := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
	periodEnd := periodStart.AddDate(0, 1, 0)

	mock.ExpectExec(
		`(?s)UPDATE user_subscriptions`,
	).
		WithArgs(
			uint(2),
			"sub_123",
			periodStart,
			periodEnd,
			sqlmock.AnyArg(),
			uint(5),
		).
		WillReturnResult(
			sqlmock.NewResult(0, 0),
		)

	mock.ExpectExec(
		`(?s)INSERT INTO user_subscriptions.*VALUES`,
	).
		WithArgs(
			uint(5),
			uint(2),
			"sub_123",
			periodStart,
			periodEnd,
			sqlmock.AnyArg(),
		).
		WillReturnResult(
			sqlmock.NewResult(1, 1),
		)

	err := repo.ActivateUserSubscription(
		context.Background(),
		5,
		2,
		"sub_123",
		periodStart,
		periodEnd,
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSQLRepository_UpdateUserSubscriptionStatusByCustomer_Success(
	t *testing.T,
) {
	repo, mock := newBillingRepositoryTest(t)

	mock.ExpectExec(
		`(?s)UPDATE user_subscriptions.*WHERE user_id = .*billing_customers`,
	).
		WithArgs(
			"cancelled",
			sqlmock.AnyArg(),
			"cus_123",
		).
		WillReturnResult(
			sqlmock.NewResult(0, 1),
		)

	err := repo.UpdateUserSubscriptionStatusByCustomer(
		context.Background(),
		"cus_123",
		"cancelled",
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSQLRepository_UpdateUserSubscriptionStatusByCustomer_Error(
	t *testing.T,
) {
	repo, mock := newBillingRepositoryTest(t)

	dbErr := errors.New("update failed")

	mock.ExpectExec(
		`(?s)UPDATE user_subscriptions`,
	).
		WithArgs(
			"cancelled",
			sqlmock.AnyArg(),
			"cus_123",
		).
		WillReturnError(dbErr)

	err := repo.UpdateUserSubscriptionStatusByCustomer(
		context.Background(),
		"cus_123",
		"cancelled",
	)

	if !errors.Is(err, dbErr) {
		t.Fatalf(
			"expected update error got %v",
			err,
		)
	}
}

func promoColumns() []string {
	return []string{
		"id",
		"code",
		"description",
		"discount_type",
		"percentage_off",
		"fixed_amount_pence",
		"free_months",
		"max_uses",
		"times_used",
		"user_type",
		"plan_id",
		"starts_at",
		"expires_at",
		"is_active",
		"created_at",
		"updated_at",
	}
}

func TestSQLRepository_GetPromoCode_Success(t *testing.T) {
	repo, mock := newBillingRepositoryTest(t)

	now := time.Now()
	startsAt := now.Add(-time.Hour)
	expiresAt := now.Add(time.Hour)

	mock.ExpectQuery(
		`(?s)SELECT.*FROM promo_codes.*WHERE code = \$1`,
	).
		WithArgs("SAVE20").
		WillReturnRows(
			sqlmock.NewRows(
				promoColumns(),
			).AddRow(
				1,
				"SAVE20",
				"20 percent off",
				"percentage",
				20,
				0,
				0,
				100,
				5,
				"cleaner",
				2,
				startsAt,
				expiresAt,
				true,
				now,
				now,
			),
		)

	promo, err := repo.GetPromoCode(
		context.Background(),
		"SAVE20",
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if promo.PlanID == nil ||
		*promo.PlanID != 2 {
		t.Fatal("expected plan ID 2")
	}

	if promo.StartsAt == nil {
		t.Fatal("expected StartsAt")
	}

	if promo.ExpiresAt == nil {
		t.Fatal("expected ExpiresAt")
	}
}

func TestSQLRepository_GetPromoCode_NullOptionalFields(
	t *testing.T,
) {
	repo, mock := newBillingRepositoryTest(t)

	now := time.Now()

	mock.ExpectQuery(
		`(?s)SELECT.*FROM promo_codes.*WHERE code = \$1`,
	).
		WithArgs("GENERAL").
		WillReturnRows(
			sqlmock.NewRows(
				promoColumns(),
			).AddRow(
				1,
				"GENERAL",
				"",
				"percentage",
				10,
				0,
				0,
				0,
				0,
				"",
				nil,
				nil,
				nil,
				true,
				now,
				now,
			),
		)

	promo, err := repo.GetPromoCode(
		context.Background(),
		"GENERAL",
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if promo.PlanID != nil {
		t.Fatal("expected nil PlanID")
	}

	if promo.StartsAt != nil {
		t.Fatal("expected nil StartsAt")
	}

	if promo.ExpiresAt != nil {
		t.Fatal("expected nil ExpiresAt")
	}
}

func TestSQLRepository_GetPromoCode_NotFound(t *testing.T) {
	repo, mock := newBillingRepositoryTest(t)

	mock.ExpectQuery(
		`(?s)SELECT.*FROM promo_codes.*WHERE code = \$1`,
	).
		WithArgs("NOPE").
		WillReturnError(sql.ErrNoRows)

	_, err := repo.GetPromoCode(
		context.Background(),
		"NOPE",
	)

	if !errors.Is(err, ErrPromoCodeNotFound) {
		t.Fatalf(
			"expected ErrPromoCodeNotFound got %v",
			err,
		)
	}
}

func TestSQLRepository_GetPromoCode_Error(t *testing.T) {
	repo, mock := newBillingRepositoryTest(t)

	dbErr := errors.New("database failed")

	mock.ExpectQuery(
		`(?s)SELECT.*FROM promo_codes.*WHERE code = \$1`,
	).
		WithArgs("SAVE20").
		WillReturnError(dbErr)

	_, err := repo.GetPromoCode(
		context.Background(),
		"SAVE20",
	)

	if !errors.Is(err, dbErr) {
		t.Fatalf(
			"expected database error got %v",
			err,
		)
	}
}

func TestSQLRepository_IncrementPromoUse_Success(
	t *testing.T,
) {
	repo, mock := newBillingRepositoryTest(t)

	mock.ExpectExec(
		`(?s)UPDATE promo_codes.*times_used = times_used \+ 1.*WHERE code = \$2`,
	).
		WithArgs(
			sqlmock.AnyArg(),
			"SAVE20",
		).
		WillReturnResult(
			sqlmock.NewResult(0, 1),
		)

	err := repo.IncrementPromoUse(
		context.Background(),
		"SAVE20",
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSQLRepository_IncrementPromoUse_Error(
	t *testing.T,
) {
	repo, mock := newBillingRepositoryTest(t)

	dbErr := errors.New("update failed")

	mock.ExpectExec(
		`(?s)UPDATE promo_codes`,
	).
		WithArgs(
			sqlmock.AnyArg(),
			"SAVE20",
		).
		WillReturnError(dbErr)

	err := repo.IncrementPromoUse(
		context.Background(),
		"SAVE20",
	)

	if !errors.Is(err, dbErr) {
		t.Fatalf(
			"expected update error got %v",
			err,
		)
	}
}

func TestSQLRepository_UpdateUserSubscriptionRenewalByCustomer_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	periodStart := time.Date(
		2026,
		time.September,
		2,
		12,
		0,
		0,
		0,
		time.UTC,
	)

	periodEnd := time.Date(
		2026,
		time.October,
		2,
		12,
		0,
		0,
		0,
		time.UTC,
	)

	mock.ExpectExec(
		regexp.QuoteMeta(`
		UPDATE user_subscriptions
		SET
			status = 'active',
			stripe_subscription_id = $1,
			current_period_start = $2,
			current_period_end = $3,
			updated_at = $4
		WHERE user_id = (
			SELECT user_id
			FROM billing_customers
			WHERE stripe_customer_id = $5
		)
	`),
	).
		WithArgs(
			"sub_123",
			periodStart,
			periodEnd,
			sqlmock.AnyArg(),
			"cus_123",
		).
		WillReturnResult(
			sqlmock.NewResult(0, 1),
		)

	err = repo.UpdateUserSubscriptionRenewalByCustomer(
		context.Background(),
		"cus_123",
		"sub_123",
		periodStart,
		periodEnd,
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestSQLRepository_UpdateUserSubscriptionRenewalByCustomer_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	periodStart := time.Date(
		2026,
		time.September,
		2,
		12,
		0,
		0,
		0,
		time.UTC,
	)

	periodEnd := time.Date(
		2026,
		time.October,
		2,
		12,
		0,
		0,
		0,
		time.UTC,
	)

	mock.ExpectExec(
		regexp.QuoteMeta(`
		UPDATE user_subscriptions
		SET
			status = 'active',
			stripe_subscription_id = $1,
			current_period_start = $2,
			current_period_end = $3,
			updated_at = $4
		WHERE user_id = (
			SELECT user_id
			FROM billing_customers
			WHERE stripe_customer_id = $5
		)
	`),
	).
		WithArgs(
			"sub_123",
			periodStart,
			periodEnd,
			sqlmock.AnyArg(),
			"cus_123",
		).
		WillReturnResult(
			sqlmock.NewResult(0, 0),
		)

	err = repo.UpdateUserSubscriptionRenewalByCustomer(
		context.Background(),
		"cus_123",
		"sub_123",
		periodStart,
		periodEnd,
	)

	if !errors.Is(err, ErrSubscriptionNotFound) {
		t.Fatalf(
			"expected ErrSubscriptionNotFound, got %v",
			err,
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestSQLRepository_UpdateUserSubscriptionRenewalByCustomer_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	periodStart := time.Date(
		2026,
		time.September,
		2,
		12,
		0,
		0,
		0,
		time.UTC,
	)

	periodEnd := time.Date(
		2026,
		time.October,
		2,
		12,
		0,
		0,
		0,
		time.UTC,
	)

	mock.ExpectExec(
		regexp.QuoteMeta(`
		UPDATE user_subscriptions
		SET
			status = 'active',
			stripe_subscription_id = $1,
			current_period_start = $2,
			current_period_end = $3,
			updated_at = $4
		WHERE user_id = (
			SELECT user_id
			FROM billing_customers
			WHERE stripe_customer_id = $5
		)
	`),
	).
		WithArgs(
			"sub_123",
			periodStart,
			periodEnd,
			sqlmock.AnyArg(),
			"cus_123",
		).
		WillReturnError(
			errors.New("database error"),
		)

	err = repo.UpdateUserSubscriptionRenewalByCustomer(
		context.Background(),
		"cus_123",
		"sub_123",
		periodStart,
		periodEnd,
	)

	if err == nil {
		t.Fatal("expected error")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestSQLRepository_MarkCheckoutSessionExpired_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	mock.ExpectExec(
		regexp.QuoteMeta(`
		UPDATE billing_checkout_sessions
		SET
			status = 'expired',
			updated_at = $1
		WHERE stripe_session_id = $2
		  AND status <> 'completed'
		  AND status <> 'expired'
	`),
	).
		WithArgs(
			sqlmock.AnyArg(),
			"cs_expired_123",
		).
		WillReturnResult(
			sqlmock.NewResult(0, 1),
		)

	err = repo.MarkCheckoutSessionExpired(
		context.Background(),
		"cs_expired_123",
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestSQLRepository_MarkCheckoutSessionExpired_AlreadyCompleted_NoOp(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	mock.ExpectExec(
		regexp.QuoteMeta(`
		UPDATE billing_checkout_sessions
		SET
			status = 'expired',
			updated_at = $1
		WHERE stripe_session_id = $2
		  AND status <> 'completed'
		  AND status <> 'expired'
	`),
	).
		WithArgs(
			sqlmock.AnyArg(),
			"cs_completed_123",
		).
		WillReturnResult(
			sqlmock.NewResult(0, 0),
		)

	err = repo.MarkCheckoutSessionExpired(
		context.Background(),
		"cs_completed_123",
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestSQLRepository_MarkCheckoutSessionExpired_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	expectedErr := errors.New("database error")

	mock.ExpectExec(
		regexp.QuoteMeta(`
		UPDATE billing_checkout_sessions
		SET
			status = 'expired',
			updated_at = $1
		WHERE stripe_session_id = $2
		  AND status <> 'completed'
		  AND status <> 'expired'
	`),
	).
		WithArgs(
			sqlmock.AnyArg(),
			"cs_expired_123",
		).
		WillReturnError(expectedErr)

	err = repo.MarkCheckoutSessionExpired(
		context.Background(),
		"cs_expired_123",
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected %v got %v",
			expectedErr,
			err,
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

var _ Repository = (*SQLRepository)(nil)
