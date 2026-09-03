package jobalerts

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

func TestSQLRepository_Create_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)
	now := time.Now()

	mock.ExpectQuery(
		regexp.QuoteMeta("INSERT INTO job_alerts"),
	).
		WithArgs(
			uint(5),
			"London",
			"domestic",
			80,
			true,
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

	alert := &JobAlert{
		UserID:        5,
		Location:      "London",
		JobType:       "domestic",
		MinimumBudget: 80,
		IsActive:      true,
	}

	err = repo.Create(
		context.Background(),
		alert,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if alert.ID != 12 {
		t.Fatalf(
			"expected alert ID 12, got %d",
			alert.ID,
		)
	}

	if alert.CreatedAt.IsZero() {
		t.Fatal("expected created_at")
	}

	if alert.UpdatedAt.IsZero() {
		t.Fatal("expected updated_at")
	}
}

func TestSQLRepository_Create_DatabaseError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)
	expectedErr := errors.New("create alert failed")

	mock.ExpectQuery(
		regexp.QuoteMeta("INSERT INTO job_alerts"),
	).
		WillReturnError(expectedErr)

	err = repo.Create(
		context.Background(),
		&JobAlert{},
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected %v, got %v",
			expectedErr,
			err,
		)
	}
}

func TestSQLRepository_GetByID_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)
	now := time.Now()

	mock.ExpectQuery(
		regexp.QuoteMeta("FROM job_alerts"),
	).
		WithArgs(uint(12)).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"id",
				"user_id",
				"location",
				"job_type",
				"minimum_budget",
				"is_active",
				"created_at",
				"updated_at",
			}).AddRow(
				12,
				5,
				"London",
				"domestic",
				80,
				true,
				now,
				now,
			),
		)

	alert, err := repo.GetByID(
		context.Background(),
		12,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if alert == nil {
		t.Fatal("expected alert")
	}

	if alert.ID != 12 {
		t.Fatalf(
			"expected alert ID 12, got %d",
			alert.ID,
		)
	}

	if alert.UserID != 5 {
		t.Fatalf(
			"expected user ID 5, got %d",
			alert.UserID,
		)
	}
}

func TestSQLRepository_GetByID_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	mock.ExpectQuery(
		regexp.QuoteMeta("FROM job_alerts"),
	).
		WithArgs(uint(12)).
		WillReturnError(sql.ErrNoRows)

	alert, err := repo.GetByID(
		context.Background(),
		12,
	)

	if alert != nil {
		t.Fatalf(
			"expected nil alert, got %+v",
			alert,
		)
	}

	if !errors.Is(err, ErrAlertNotFound) {
		t.Fatalf(
			"expected ErrAlertNotFound, got %v",
			err,
		)
	}
}

func TestSQLRepository_GetByID_DatabaseError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)
	expectedErr := errors.New("get alert failed")

	mock.ExpectQuery(
		regexp.QuoteMeta("FROM job_alerts"),
	).
		WithArgs(uint(12)).
		WillReturnError(expectedErr)

	alert, err := repo.GetByID(
		context.Background(),
		12,
	)

	if alert != nil {
		t.Fatalf(
			"expected nil alert, got %+v",
			alert,
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

func TestSQLRepository_ListByUserID_Success(t *testing.T) {
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
		"location",
		"job_type",
		"minimum_budget",
		"is_active",
		"created_at",
		"updated_at",
	}).
		AddRow(
			1,
			5,
			"London",
			"domestic",
			80,
			true,
			now,
			now,
		).
		AddRow(
			2,
			5,
			"Essex",
			"commercial",
			120,
			false,
			now.Add(-time.Hour),
			now.Add(-time.Hour),
		)

	mock.ExpectQuery(
		regexp.QuoteMeta("FROM job_alerts"),
	).
		WithArgs(uint(5)).
		WillReturnRows(rows)

	alerts, err := repo.ListByUserID(
		context.Background(),
		5,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(alerts) != 2 {
		t.Fatalf(
			"expected 2 alerts, got %d",
			len(alerts),
		)
	}

	if alerts[0].Location != "London" {
		t.Fatalf(
			"expected London, got %q",
			alerts[0].Location,
		)
	}
}

func TestSQLRepository_ListByUserID_Empty(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	mock.ExpectQuery(
		regexp.QuoteMeta("FROM job_alerts"),
	).
		WithArgs(uint(5)).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"id",
				"user_id",
				"location",
				"job_type",
				"minimum_budget",
				"is_active",
				"created_at",
				"updated_at",
			}),
		)

	alerts, err := repo.ListByUserID(
		context.Background(),
		5,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(alerts) != 0 {
		t.Fatalf(
			"expected no alerts, got %d",
			len(alerts),
		)
	}
}

func TestSQLRepository_ListByUserID_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)
	expectedErr := errors.New("query alerts failed")

	mock.ExpectQuery(
		regexp.QuoteMeta("FROM job_alerts"),
	).
		WithArgs(uint(5)).
		WillReturnError(expectedErr)

	alerts, err := repo.ListByUserID(
		context.Background(),
		5,
	)

	if alerts != nil {
		t.Fatalf(
			"expected nil alerts, got %+v",
			alerts,
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

func TestSQLRepository_ListByUserID_ScanError(t *testing.T) {
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
		"location",
		"job_type",
		"minimum_budget",
		"is_active",
		"created_at",
		"updated_at",
	}).AddRow(
		"invalid-id",
		5,
		"London",
		"domestic",
		80,
		true,
		now,
		now,
	)

	mock.ExpectQuery(
		regexp.QuoteMeta("FROM job_alerts"),
	).
		WithArgs(uint(5)).
		WillReturnRows(rows)

	alerts, err := repo.ListByUserID(
		context.Background(),
		5,
	)

	if alerts != nil {
		t.Fatalf(
			"expected nil alerts, got %+v",
			alerts,
		)
	}

	if err == nil {
		t.Fatal("expected scan error")
	}
}

func TestSQLRepository_ListByUserID_RowsError(t *testing.T) {
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
		"location",
		"job_type",
		"minimum_budget",
		"is_active",
		"created_at",
		"updated_at",
	}).
		AddRow(
			1,
			5,
			"London",
			"domestic",
			80,
			true,
			now,
			now,
		).
		RowError(0, expectedErr)

	mock.ExpectQuery(
		regexp.QuoteMeta("FROM job_alerts"),
	).
		WithArgs(uint(5)).
		WillReturnRows(rows)

	alerts, err := repo.ListByUserID(
		context.Background(),
		5,
	)

	if alerts != nil {
		t.Fatalf(
			"expected nil alerts, got %+v",
			alerts,
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

func TestSQLRepository_Update_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	mock.ExpectExec(
		regexp.QuoteMeta("UPDATE job_alerts"),
	).
		WithArgs(
			"East London",
			"airbnb",
			120,
			false,
			sqlmock.AnyArg(),
			uint(12),
		).
		WillReturnResult(
			sqlmock.NewResult(0, 1),
		)

	alert := &JobAlert{
		ID:            12,
		Location:      "East London",
		JobType:       "airbnb",
		MinimumBudget: 120,
		IsActive:      false,
	}

	err = repo.Update(
		context.Background(),
		alert,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if alert.UpdatedAt.IsZero() {
		t.Fatal("expected updated_at")
	}
}

func TestSQLRepository_Update_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	mock.ExpectExec(
		regexp.QuoteMeta("UPDATE job_alerts"),
	).
		WithArgs(
			"",
			"",
			0,
			false,
			sqlmock.AnyArg(),
			uint(12),
		).
		WillReturnResult(
			sqlmock.NewResult(0, 0),
		)

	err = repo.Update(
		context.Background(),
		&JobAlert{ID: 12},
	)

	if !errors.Is(err, ErrAlertNotFound) {
		t.Fatalf(
			"expected ErrAlertNotFound, got %v",
			err,
		)
	}
}

func TestSQLRepository_Update_ExecError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)
	expectedErr := errors.New("update alert failed")

	mock.ExpectExec(
		regexp.QuoteMeta("UPDATE job_alerts"),
	).
		WithArgs(
			"",
			"",
			0,
			false,
			sqlmock.AnyArg(),
			uint(12),
		).
		WillReturnError(expectedErr)

	err = repo.Update(
		context.Background(),
		&JobAlert{ID: 12},
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected %v, got %v",
			expectedErr,
			err,
		)
	}
}

func TestSQLRepository_Update_RowsAffectedError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)
	expectedErr := errors.New("rows affected failed")

	mock.ExpectExec(
		regexp.QuoteMeta("UPDATE job_alerts"),
	).
		WithArgs(
			"",
			"",
			0,
			false,
			sqlmock.AnyArg(),
			uint(12),
		).
		WillReturnResult(
			sqlmock.NewErrorResult(expectedErr),
		)

	err = repo.Update(
		context.Background(),
		&JobAlert{ID: 12},
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected %v, got %v",
			expectedErr,
			err,
		)
	}
}

func TestSQLRepository_Delete_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	mock.ExpectExec(
		regexp.QuoteMeta("DELETE FROM job_alerts"),
	).
		WithArgs(uint(12)).
		WillReturnResult(
			sqlmock.NewResult(0, 1),
		)

	err = repo.Delete(
		context.Background(),
		12,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSQLRepository_Delete_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	mock.ExpectExec(
		regexp.QuoteMeta("DELETE FROM job_alerts"),
	).
		WithArgs(uint(12)).
		WillReturnResult(
			sqlmock.NewResult(0, 0),
		)

	err = repo.Delete(
		context.Background(),
		12,
	)

	if !errors.Is(err, ErrAlertNotFound) {
		t.Fatalf(
			"expected ErrAlertNotFound, got %v",
			err,
		)
	}
}

func TestSQLRepository_Delete_ExecError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)
	expectedErr := errors.New("delete alert failed")

	mock.ExpectExec(
		regexp.QuoteMeta("DELETE FROM job_alerts"),
	).
		WithArgs(uint(12)).
		WillReturnError(expectedErr)

	err = repo.Delete(
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

func TestSQLRepository_Delete_RowsAffectedError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)
	expectedErr := errors.New("rows affected failed")

	mock.ExpectExec(
		regexp.QuoteMeta("DELETE FROM job_alerts"),
	).
		WithArgs(uint(12)).
		WillReturnResult(
			sqlmock.NewErrorResult(expectedErr),
		)

	err = repo.Delete(
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
