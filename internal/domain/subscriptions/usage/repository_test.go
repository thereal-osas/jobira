package usage

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func newUsageRepositoryTest(
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

func usageColumns() []string {
	return []string{
		"id",
		"user_id",
		"application_count",
		"job_post_count",
		"free_application_limit",
		"free_job_post_limit",
		"monetisation_enabled",
		"trial_started_at",
		"trial_ends_at",
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
		t.Fatal(
			"expected repository",
		)
	}

	if repository.db != db {
		t.Fatal(
			"expected supplied database",
		)
	}
}

func TestSQLRepository_CreateIfNotExists_InvalidUser(
	t *testing.T,
) {
	repository, _ :=
		newUsageRepositoryTest(t)

	err := repository.CreateIfNotExists(
		context.Background(),
		0,
	)

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf(
			"expected ErrInvalidInput got %v",
			err,
		)
	}
}

func TestSQLRepository_CreateIfNotExists_Success(
	t *testing.T,
) {
	repository, mock :=
		newUsageRepositoryTest(t)

	mock.ExpectExec(
		`(?s)INSERT INTO user_usage.*ON CONFLICT \(user_id\) DO NOTHING`,
	).
		WithArgs(
			uint(5),
			sqlmock.AnyArg(),
		).
		WillReturnResult(
			sqlmock.NewResult(1, 1),
		)

	err := repository.CreateIfNotExists(
		context.Background(),
		5,
	)

	if err != nil {
		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}
}

func TestSQLRepository_CreateIfNotExists_Error(
	t *testing.T,
) {
	repository, mock :=
		newUsageRepositoryTest(t)

	createErr := errors.New("create failed")

	mock.ExpectExec(
		`(?s)INSERT INTO user_usage.*ON CONFLICT`,
	).
		WithArgs(
			uint(5),
			sqlmock.AnyArg(),
		).
		WillReturnError(createErr)

	err := repository.CreateIfNotExists(
		context.Background(),
		5,
	)

	if !errors.Is(err, createErr) {
		t.Fatalf(
			"expected create error got %v",
			err,
		)
	}
}

func TestSQLRepository_GetByUserID_InvalidUser(
	t *testing.T,
) {
	repository, _ :=
		newUsageRepositoryTest(t)

	usage, err := repository.GetByUserID(
		context.Background(),
		0,
	)

	if usage != nil {
		t.Fatal("expected nil usage")
	}

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf(
			"expected ErrInvalidInput got %v",
			err,
		)
	}
}

func TestSQLRepository_GetByUserID_Success(
	t *testing.T,
) {
	repository, mock :=
		newUsageRepositoryTest(t)

	now := time.Now()
	trialEnds := now.Add(
		30 * 24 * time.Hour,
	)

	mock.ExpectQuery(
		`(?s)SELECT.*FROM user_usage.*WHERE user_id = \$1`,
	).
		WithArgs(uint(5)).
		WillReturnRows(
			sqlmock.NewRows(
				usageColumns(),
			).AddRow(
				1,
				5,
				2,
				3,
				5,
				5,
				true,
				now,
				trialEnds,
				now,
				now,
			),
		)

	result, err := repository.GetByUserID(
		context.Background(),
		5,
	)

	if err != nil {
		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}

	if result.ID != 1 {
		t.Fatalf(
			"expected ID 1 got %d",
			result.ID,
		)
	}

	if result.TrialEndsAt == nil {
		t.Fatal(
			"expected TrialEndsAt",
		)
	}
}

func TestSQLRepository_GetByUserID_NullTrialEnd(
	t *testing.T,
) {
	repository, mock :=
		newUsageRepositoryTest(t)

	now := time.Now()

	mock.ExpectQuery(
		`(?s)SELECT.*FROM user_usage.*WHERE user_id = \$1`,
	).
		WithArgs(uint(5)).
		WillReturnRows(
			sqlmock.NewRows(
				usageColumns(),
			).AddRow(
				1,
				5,
				0,
				0,
				5,
				5,
				false,
				now,
				nil,
				now,
				now,
			),
		)

	result, err := repository.GetByUserID(
		context.Background(),
		5,
	)

	if err != nil {
		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}

	if result.TrialEndsAt != nil {
		t.Fatal(
			"expected nil TrialEndsAt",
		)
	}
}

func TestSQLRepository_GetByUserID_NotFound(
	t *testing.T,
) {
	repository, mock :=
		newUsageRepositoryTest(t)

	mock.ExpectQuery(
		`(?s)SELECT.*FROM user_usage.*WHERE user_id = \$1`,
	).
		WithArgs(uint(999)).
		WillReturnError(sql.ErrNoRows)

	_, err := repository.GetByUserID(
		context.Background(),
		999,
	)

	if !errors.Is(err, ErrUsageNotFound) {
		t.Fatalf(
			"expected ErrUsageNotFound got %v",
			err,
		)
	}
}

func TestSQLRepository_GetByUserID_DatabaseError(
	t *testing.T,
) {
	repository, mock :=
		newUsageRepositoryTest(t)

	databaseErr := errors.New(
		"database failed",
	)

	mock.ExpectQuery(
		`(?s)SELECT.*FROM user_usage.*WHERE user_id = \$1`,
	).
		WithArgs(uint(5)).
		WillReturnError(databaseErr)

	_, err := repository.GetByUserID(
		context.Background(),
		5,
	)

	if !errors.Is(err, databaseErr) {
		t.Fatalf(
			"expected database error got %v",
			err,
		)
	}
}

func TestSQLRepository_IncrementApplicationCount_InvalidUser(
	t *testing.T,
) {
	repository, _ :=
		newUsageRepositoryTest(t)

	err := repository.IncrementApplicationCount(
		context.Background(),
		0,
	)

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf(
			"expected ErrInvalidInput got %v",
			err,
		)
	}
}

func TestSQLRepository_IncrementApplicationCount_Success(
	t *testing.T,
) {
	repository, mock :=
		newUsageRepositoryTest(t)

	mock.ExpectExec(
		`(?s)INSERT INTO user_usage.*ON CONFLICT`,
	).
		WithArgs(
			uint(5),
			sqlmock.AnyArg(),
		).
		WillReturnResult(
			sqlmock.NewResult(0, 0),
		)

	mock.ExpectExec(
		`(?s)UPDATE user_usage.*application_count = application_count \+ 1.*WHERE user_id = \$2`,
	).
		WithArgs(
			sqlmock.AnyArg(),
			uint(5),
		).
		WillReturnResult(
			sqlmock.NewResult(0, 1),
		)

	err := repository.IncrementApplicationCount(
		context.Background(),
		5,
	)

	if err != nil {
		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}
}

func TestSQLRepository_IncrementApplicationCount_CreateError(
	t *testing.T,
) {
	repository, mock :=
		newUsageRepositoryTest(t)

	createErr := errors.New("create failed")

	mock.ExpectExec(
		`(?s)INSERT INTO user_usage.*ON CONFLICT`,
	).
		WithArgs(
			uint(5),
			sqlmock.AnyArg(),
		).
		WillReturnError(createErr)

	err := repository.IncrementApplicationCount(
		context.Background(),
		5,
	)

	if !errors.Is(err, createErr) {
		t.Fatalf(
			"expected create error got %v",
			err,
		)
	}
}

func TestSQLRepository_IncrementApplicationCount_UpdateError(
	t *testing.T,
) {
	repository, mock :=
		newUsageRepositoryTest(t)

	updateErr := errors.New("update failed")

	mock.ExpectExec(
		`(?s)INSERT INTO user_usage.*ON CONFLICT`,
	).
		WithArgs(
			uint(5),
			sqlmock.AnyArg(),
		).
		WillReturnResult(
			sqlmock.NewResult(0, 0),
		)

	mock.ExpectExec(
		`(?s)UPDATE user_usage.*application_count`,
	).
		WithArgs(
			sqlmock.AnyArg(),
			uint(5),
		).
		WillReturnError(updateErr)

	err := repository.IncrementApplicationCount(
		context.Background(),
		5,
	)

	if !errors.Is(err, updateErr) {
		t.Fatalf(
			"expected update error got %v",
			err,
		)
	}
}

func TestSQLRepository_IncrementJobPostCount_InvalidUser(
	t *testing.T,
) {
	repository, _ :=
		newUsageRepositoryTest(t)

	err := repository.IncrementJobPostCount(
		context.Background(),
		0,
	)

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf(
			"expected ErrInvalidInput got %v",
			err,
		)
	}
}

func TestSQLRepository_IncrementJobPostCount_Success(
	t *testing.T,
) {
	repository, mock :=
		newUsageRepositoryTest(t)

	mock.ExpectExec(
		`(?s)INSERT INTO user_usage.*ON CONFLICT`,
	).
		WithArgs(
			uint(5),
			sqlmock.AnyArg(),
		).
		WillReturnResult(
			sqlmock.NewResult(0, 0),
		)

	mock.ExpectExec(
		`(?s)UPDATE user_usage.*job_post_count = job_post_count \+ 1.*WHERE user_id = \$2`,
	).
		WithArgs(
			sqlmock.AnyArg(),
			uint(5),
		).
		WillReturnResult(
			sqlmock.NewResult(0, 1),
		)

	err := repository.IncrementJobPostCount(
		context.Background(),
		5,
	)

	if err != nil {
		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}
}

func TestSQLRepository_IncrementJobPostCount_CreateError(
	t *testing.T,
) {
	repository, mock :=
		newUsageRepositoryTest(t)

	createErr := errors.New("create failed")

	mock.ExpectExec(
		`(?s)INSERT INTO user_usage.*ON CONFLICT`,
	).
		WithArgs(
			uint(5),
			sqlmock.AnyArg(),
		).
		WillReturnError(createErr)

	err := repository.IncrementJobPostCount(
		context.Background(),
		5,
	)

	if !errors.Is(err, createErr) {
		t.Fatalf(
			"expected create error got %v",
			err,
		)
	}
}

func TestSQLRepository_SetMonetisationEnabled_InvalidUser(
	t *testing.T,
) {
	repository, _ :=
		newUsageRepositoryTest(t)

	err := repository.SetMonetisationEnabled(
		context.Background(),
		0,
		true,
	)

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf(
			"expected ErrInvalidInput got %v",
			err,
		)
	}
}

func TestSQLRepository_SetMonetisationEnabled_Success(
	t *testing.T,
) {
	repository, mock :=
		newUsageRepositoryTest(t)

	mock.ExpectExec(
		`(?s)INSERT INTO user_usage.*ON CONFLICT`,
	).
		WithArgs(
			uint(5),
			sqlmock.AnyArg(),
		).
		WillReturnResult(
			sqlmock.NewResult(0, 0),
		)

	mock.ExpectExec(
		`(?s)UPDATE user_usage.*monetisation_enabled = \$1.*WHERE user_id = \$3`,
	).
		WithArgs(
			true,
			sqlmock.AnyArg(),
			uint(5),
		).
		WillReturnResult(
			sqlmock.NewResult(0, 1),
		)

	err := repository.SetMonetisationEnabled(
		context.Background(),
		5,
		true,
	)

	if err != nil {
		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}
}

func TestSQLRepository_SetMonetisationEnabled_CreateError(
	t *testing.T,
) {
	repository, mock :=
		newUsageRepositoryTest(t)

	createErr := errors.New("create failed")

	mock.ExpectExec(
		`(?s)INSERT INTO user_usage.*ON CONFLICT`,
	).
		WithArgs(
			uint(5),
			sqlmock.AnyArg(),
		).
		WillReturnError(createErr)

	err := repository.SetMonetisationEnabled(
		context.Background(),
		5,
		true,
	)

	if !errors.Is(err, createErr) {
		t.Fatalf(
			"expected create error got %v",
			err,
		)
	}
}

func TestSQLRepository_GetUserSubscriptionAccess_InvalidUser(
	t *testing.T,
) {
	repository, _ :=
		newUsageRepositoryTest(t)

	access, err :=
		repository.GetUserSubscriptionAccess(
			context.Background(),
			0,
		)

	if access != nil {
		t.Fatal("expected nil access")
	}

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf(
			"expected ErrInvalidInput got %v",
			err,
		)
	}
}

func TestSQLRepository_GetUserSubscriptionAccess_Success(
	t *testing.T,
) {
	repository, mock :=
		newUsageRepositoryTest(t)

	mock.ExpectQuery(
		`(?s)SELECT.*FROM user_subscriptions us.*LEFT JOIN subscription_plans sp.*WHERE us.user_id = \$1`,
	).
		WithArgs(uint(5)).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{
					"status",
					"application_limit",
					"job_post_limit",
				},
			).AddRow(
				"active",
				100,
				20,
			),
		)

	access, err :=
		repository.GetUserSubscriptionAccess(
			context.Background(),
			5,
		)

	if err != nil {
		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}

	if access == nil {
		t.Fatal("expected access")
	}

	if access.Status != "active" {
		t.Fatalf(
			"expected active got %q",
			access.Status,
		)
	}

	if access.ApplicationLimit != 100 {
		t.Fatalf(
			"expected 100 application limit got %d",
			access.ApplicationLimit,
		)
	}
}

func TestSQLRepository_GetUserSubscriptionAccess_NotFound(
	t *testing.T,
) {
	repository, mock :=
		newUsageRepositoryTest(t)

	mock.ExpectQuery(
		`(?s)SELECT.*FROM user_subscriptions us.*WHERE us.user_id = \$1`,
	).
		WithArgs(uint(5)).
		WillReturnError(sql.ErrNoRows)

	access, err :=
		repository.GetUserSubscriptionAccess(
			context.Background(),
			5,
		)

	if err != nil {
		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}

	if access != nil {
		t.Fatal(
			"expected nil access when subscription missing",
		)
	}
}

func TestSQLRepository_GetUserSubscriptionAccess_DatabaseError(
	t *testing.T,
) {
	repository, mock :=
		newUsageRepositoryTest(t)

	databaseErr := errors.New(
		"database failed",
	)

	mock.ExpectQuery(
		`(?s)SELECT.*FROM user_subscriptions us.*WHERE us.user_id = \$1`,
	).
		WithArgs(uint(5)).
		WillReturnError(databaseErr)

	_, err :=
		repository.GetUserSubscriptionAccess(
			context.Background(),
			5,
		)

	if !errors.Is(err, databaseErr) {
		t.Fatalf(
			"expected database error got %v",
			err,
		)
	}
}

var _ Repository = (*SQLRepository)(nil)
