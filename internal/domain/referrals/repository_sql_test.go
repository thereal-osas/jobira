package referrals

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

func TestSQLRepository_CreateCode_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)
	now := time.Now()

	mock.ExpectQuery(
		regexp.QuoteMeta("INSERT INTO referral_codes"),
	).
		WithArgs(
			uint(5),
			"welcome",
			"credit",
			1500,
			10,
			sqlmock.AnyArg(),
		).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"id",
				"created_at",
				"updated_at",
			}).AddRow(
				12,
				now,
				now,
			),
		)

	code := &ReferralCode{
		UserID:           5,
		Code:             "welcome",
		RewardType:       "credit",
		RewardValuePence: 1500,
		MaxUses:          10,
	}

	err = repo.CreateCode(
		context.Background(),
		code,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if code.ID != 12 {
		t.Fatalf(
			"expected ID 12, got %d",
			code.ID,
		)
	}

	if code.CreatedAt.IsZero() {
		t.Fatal("expected created_at")
	}

	if code.UpdatedAt.IsZero() {
		t.Fatal("expected updated_at")
	}
}

func TestSQLRepository_CreateCode_Duplicate(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	mock.ExpectQuery(
		regexp.QuoteMeta("INSERT INTO referral_codes"),
	).
		WillReturnError(
			errors.New(
				"duplicate key value violates unique constraint",
			),
		)

	err = repo.CreateCode(
		context.Background(),
		&ReferralCode{},
	)

	if !errors.Is(err, ErrReferralCodeExists) {
		t.Fatalf(
			"expected ErrReferralCodeExists, got %v",
			err,
		)
	}
}

func TestSQLRepository_CreateCode_DatabaseError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)
	expectedErr := errors.New("create referral code failed")

	mock.ExpectQuery(
		regexp.QuoteMeta("INSERT INTO referral_codes"),
	).
		WillReturnError(expectedErr)

	err = repo.CreateCode(
		context.Background(),
		&ReferralCode{},
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected %v, got %v",
			expectedErr,
			err,
		)
	}
}

func TestSQLRepository_GetCodeByCode_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)
	now := time.Now()

	mock.ExpectQuery(
		regexp.QuoteMeta("FROM referral_codes"),
	).
		WithArgs("welcome").
		WillReturnRows(
			sqlmock.NewRows([]string{
				"id",
				"user_id",
				"code",
				"reward_type",
				"reward_value_pence",
				"max_uses",
				"times_used",
				"is_active",
				"created_at",
				"updated_at",
			}).AddRow(
				12,
				5,
				"welcome",
				"credit",
				1500,
				10,
				2,
				true,
				now,
				now,
			),
		)

	code, err := repo.GetCodeByCode(
		context.Background(),
		"welcome",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if code == nil {
		t.Fatal("expected referral code")
	}

	if code.ID != 12 {
		t.Fatalf(
			"expected ID 12, got %d",
			code.ID,
		)
	}

	if code.UserID != 5 {
		t.Fatalf(
			"expected user ID 5, got %d",
			code.UserID,
		)
	}

	if code.Code != "welcome" {
		t.Fatalf(
			"expected code welcome, got %q",
			code.Code,
		)
	}
}

func TestSQLRepository_GetCodeByCode_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	mock.ExpectQuery(
		regexp.QuoteMeta("FROM referral_codes"),
	).
		WithArgs("missing").
		WillReturnError(sql.ErrNoRows)

	code, err := repo.GetCodeByCode(
		context.Background(),
		"missing",
	)

	if code != nil {
		t.Fatalf(
			"expected nil code, got %+v",
			code,
		)
	}

	if !errors.Is(err, ErrReferralNotFound) {
		t.Fatalf(
			"expected ErrReferralNotFound, got %v",
			err,
		)
	}
}

func TestSQLRepository_GetCodeByCode_DatabaseError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)
	expectedErr := errors.New("get referral code failed")

	mock.ExpectQuery(
		regexp.QuoteMeta("FROM referral_codes"),
	).
		WithArgs("welcome").
		WillReturnError(expectedErr)

	code, err := repo.GetCodeByCode(
		context.Background(),
		"welcome",
	)

	if code != nil {
		t.Fatalf(
			"expected nil code, got %+v",
			code,
		)
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected %v, got %v",
			expectedErr,
			err,
		)
	}
}

func TestSQLRepository_ListCodesByUserID_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)
	now := time.Now()

	rows := sqlmock.NewRows([]string{
		"id",
		"user_id",
		"code",
		"reward_type",
		"reward_value_pence",
		"max_uses",
		"times_used",
		"is_active",
		"created_at",
		"updated_at",
	}).
		AddRow(
			1,
			5,
			"welcome",
			"credit",
			1500,
			10,
			2,
			true,
			now,
			now,
		).
		AddRow(
			2,
			5,
			"summer",
			"free_month",
			0,
			5,
			1,
			true,
			now.Add(-time.Hour),
			now.Add(-time.Hour),
		)

	mock.ExpectQuery(
		regexp.QuoteMeta("FROM referral_codes"),
	).
		WithArgs(uint(5)).
		WillReturnRows(rows)

	codes, err := repo.ListCodesByUserID(
		context.Background(),
		5,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(codes) != 2 {
		t.Fatalf(
			"expected 2 codes, got %d",
			len(codes),
		)
	}

	if codes[0].Code != "welcome" {
		t.Fatalf(
			"expected first code welcome, got %q",
			codes[0].Code,
		)
	}
}

func TestSQLRepository_ListCodesByUserID_Empty(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	mock.ExpectQuery(
		regexp.QuoteMeta("FROM referral_codes"),
	).
		WithArgs(uint(5)).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"id",
				"user_id",
				"code",
				"reward_type",
				"reward_value_pence",
				"max_uses",
				"times_used",
				"is_active",
				"created_at",
				"updated_at",
			}),
		)

	codes, err := repo.ListCodesByUserID(
		context.Background(),
		5,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(codes) != 0 {
		t.Fatalf(
			"expected no codes, got %d",
			len(codes),
		)
	}
}

func TestSQLRepository_ListCodesByUserID_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)
	expectedErr := errors.New("query codes failed")

	mock.ExpectQuery(
		regexp.QuoteMeta("FROM referral_codes"),
	).
		WithArgs(uint(5)).
		WillReturnError(expectedErr)

	codes, err := repo.ListCodesByUserID(
		context.Background(),
		5,
	)

	if codes != nil {
		t.Fatalf(
			"expected nil codes, got %+v",
			codes,
		)
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected %v, got %v",
			expectedErr,
			err,
		)
	}
}

func TestSQLRepository_ListCodesByUserID_ScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)
	now := time.Now()

	rows := sqlmock.NewRows([]string{
		"id",
		"user_id",
		"code",
		"reward_type",
		"reward_value_pence",
		"max_uses",
		"times_used",
		"is_active",
		"created_at",
		"updated_at",
	}).AddRow(
		"invalid-id",
		5,
		"welcome",
		"credit",
		1500,
		10,
		2,
		true,
		now,
		now,
	)

	mock.ExpectQuery(
		regexp.QuoteMeta("FROM referral_codes"),
	).
		WithArgs(uint(5)).
		WillReturnRows(rows)

	codes, err := repo.ListCodesByUserID(
		context.Background(),
		5,
	)

	if codes != nil {
		t.Fatalf(
			"expected nil codes, got %+v",
			codes,
		)
	}

	if err == nil {
		t.Fatal("expected scan error")
	}
}

func TestSQLRepository_ListCodesByUserID_RowsError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)
	now := time.Now()
	expectedErr := errors.New("rows failed")

	rows := sqlmock.NewRows([]string{
		"id",
		"user_id",
		"code",
		"reward_type",
		"reward_value_pence",
		"max_uses",
		"times_used",
		"is_active",
		"created_at",
		"updated_at",
	}).
		AddRow(
			1,
			5,
			"welcome",
			"credit",
			1500,
			10,
			2,
			true,
			now,
			now,
		).
		RowError(0, expectedErr)

	mock.ExpectQuery(
		regexp.QuoteMeta("FROM referral_codes"),
	).
		WithArgs(uint(5)).
		WillReturnRows(rows)

	codes, err := repo.ListCodesByUserID(
		context.Background(),
		5,
	)

	if codes != nil {
		t.Fatalf(
			"expected nil codes, got %+v",
			codes,
		)
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected %v, got %v",
			expectedErr,
			err,
		)
	}
}

func TestSQLRepository_CreateRedemption_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)
	now := time.Now()

	mock.ExpectQuery(
		regexp.QuoteMeta("INSERT INTO referral_redemptions"),
	).
		WithArgs(
			uint(12),
			uint(5),
			uint(8),
			"pending",
			false,
			sqlmock.AnyArg(),
		).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"id",
				"created_at",
				"updated_at",
			}).AddRow(
				20,
				now,
				now,
			),
		)

	redemption := &ReferralRedemption{
		ReferralCodeID: 12,
		ReferrerID:     5,
		ReferredUserID: 8,
		Status:         "pending",
		RewardApplied:  false,
	}

	err = repo.CreateRedemption(
		context.Background(),
		redemption,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if redemption.ID != 20 {
		t.Fatalf(
			"expected ID 20, got %d",
			redemption.ID,
		)
	}

	if redemption.CreatedAt.IsZero() {
		t.Fatal("expected created_at")
	}
}

func TestSQLRepository_CreateRedemption_Duplicate(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	mock.ExpectQuery(
		regexp.QuoteMeta("INSERT INTO referral_redemptions"),
	).
		WillReturnError(
			errors.New(
				"duplicate key value violates unique constraint",
			),
		)

	err = repo.CreateRedemption(
		context.Background(),
		&ReferralRedemption{},
	)

	if !errors.Is(err, ErrAlreadyRedeemed) {
		t.Fatalf(
			"expected ErrAlreadyRedeemed, got %v",
			err,
		)
	}
}

func TestSQLRepository_CreateRedemption_DatabaseError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)
	expectedErr := errors.New("create redemption failed")

	mock.ExpectQuery(
		regexp.QuoteMeta("INSERT INTO referral_redemptions"),
	).
		WillReturnError(expectedErr)

	err = repo.CreateRedemption(
		context.Background(),
		&ReferralRedemption{},
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected %v, got %v",
			expectedErr,
			err,
		)
	}
}

func TestSQLRepository_IncrementCodeUse_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	mock.ExpectExec(
		regexp.QuoteMeta("UPDATE referral_codes"),
	).
		WithArgs(
			sqlmock.AnyArg(),
			uint(12),
		).
		WillReturnResult(
			sqlmock.NewResult(0, 1),
		)

	err = repo.IncrementCodeUse(
		context.Background(),
		12,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSQLRepository_IncrementCodeUse_DatabaseError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)
	expectedErr := errors.New("increment failed")

	mock.ExpectExec(
		regexp.QuoteMeta("UPDATE referral_codes"),
	).
		WithArgs(
			sqlmock.AnyArg(),
			uint(12),
		).
		WillReturnError(expectedErr)

	err = repo.IncrementCodeUse(
		context.Background(),
		12,
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected %v, got %v",
			expectedErr,
			err,
		)
	}
}

func TestSQLRepository_ListRedemptionsByReferrerID_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)
	now := time.Now()

	rows := sqlmock.NewRows([]string{
		"id",
		"referral_code_id",
		"referrer_id",
		"referred_user_id",
		"status",
		"reward_applied",
		"created_at",
		"updated_at",
	}).
		AddRow(
			1,
			12,
			5,
			8,
			"pending",
			false,
			now,
			now,
		).
		AddRow(
			2,
			12,
			5,
			9,
			"completed",
			true,
			now.Add(-time.Hour),
			now.Add(-time.Hour),
		)

	mock.ExpectQuery(
		regexp.QuoteMeta("FROM referral_redemptions"),
	).
		WithArgs(uint(5)).
		WillReturnRows(rows)

	redemptions, err := repo.ListRedemptionsByReferrerID(
		context.Background(),
		5,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(redemptions) != 2 {
		t.Fatalf(
			"expected 2 redemptions, got %d",
			len(redemptions),
		)
	}

	if redemptions[0].ReferredUserID != 8 {
		t.Fatalf(
			"expected first referred user ID 8, got %d",
			redemptions[0].ReferredUserID,
		)
	}
}

func TestSQLRepository_ListRedemptionsByReferrerID_Empty(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	mock.ExpectQuery(
		regexp.QuoteMeta("FROM referral_redemptions"),
	).
		WithArgs(uint(5)).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"id",
				"referral_code_id",
				"referrer_id",
				"referred_user_id",
				"status",
				"reward_applied",
				"created_at",
				"updated_at",
			}),
		)

	redemptions, err := repo.ListRedemptionsByReferrerID(
		context.Background(),
		5,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(redemptions) != 0 {
		t.Fatalf(
			"expected no redemptions, got %d",
			len(redemptions),
		)
	}
}

func TestSQLRepository_ListRedemptionsByReferrerID_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)
	expectedErr := errors.New("query redemptions failed")

	mock.ExpectQuery(
		regexp.QuoteMeta("FROM referral_redemptions"),
	).
		WithArgs(uint(5)).
		WillReturnError(expectedErr)

	redemptions, err := repo.ListRedemptionsByReferrerID(
		context.Background(),
		5,
	)

	if redemptions != nil {
		t.Fatalf(
			"expected nil redemptions, got %+v",
			redemptions,
		)
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected %v, got %v",
			expectedErr,
			err,
		)
	}
}

func TestSQLRepository_ListRedemptionsByReferrerID_ScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)
	now := time.Now()

	rows := sqlmock.NewRows([]string{
		"id",
		"referral_code_id",
		"referrer_id",
		"referred_user_id",
		"status",
		"reward_applied",
		"created_at",
		"updated_at",
	}).AddRow(
		"invalid-id",
		12,
		5,
		8,
		"pending",
		false,
		now,
		now,
	)

	mock.ExpectQuery(
		regexp.QuoteMeta("FROM referral_redemptions"),
	).
		WithArgs(uint(5)).
		WillReturnRows(rows)

	redemptions, err := repo.ListRedemptionsByReferrerID(
		context.Background(),
		5,
	)

	if redemptions != nil {
		t.Fatalf(
			"expected nil redemptions, got %+v",
			redemptions,
		)
	}

	if err == nil {
		t.Fatal("expected scan error")
	}
}

func TestSQLRepository_ListRedemptionsByReferrerID_RowsError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)
	now := time.Now()
	expectedErr := errors.New("rows failed")

	rows := sqlmock.NewRows([]string{
		"id",
		"referral_code_id",
		"referrer_id",
		"referred_user_id",
		"status",
		"reward_applied",
		"created_at",
		"updated_at",
	}).
		AddRow(
			1,
			12,
			5,
			8,
			"pending",
			false,
			now,
			now,
		).
		RowError(0, expectedErr)

	mock.ExpectQuery(
		regexp.QuoteMeta("FROM referral_redemptions"),
	).
		WithArgs(uint(5)).
		WillReturnRows(rows)

	redemptions, err := repo.ListRedemptionsByReferrerID(
		context.Background(),
		5,
	)

	if redemptions != nil {
		t.Fatalf(
			"expected nil redemptions, got %+v",
			redemptions,
		)
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected %v, got %v",
			expectedErr,
			err,
		)
	}
}
