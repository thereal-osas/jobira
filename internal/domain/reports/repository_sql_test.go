package reports

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func newReportsSQLMock(
	t *testing.T,
) (*SQLRepository, sqlmock.Sqlmock) {
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

	return NewSQLRepository(db), mock
}

func TestNewSQLRepository(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf(
			"failed to create sqlmock: %v",
			err,
		)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	if repo == nil {
		t.Fatal("expected repository")
	}

	if repo.db != db {
		t.Fatal("expected db assigned")
	}
}

func TestSQLRepository_Create_Success(t *testing.T) {
	repo, mock := newReportsSQLMock(t)

	report := &Report{
		ReporterID: 5,
		ReportType: "user",
		Reason:     "abuse",
		Details:    "details",
		Status:     "open",
	}

	createdAt := time.Now()

	mock.ExpectQuery(
		`INSERT INTO reports`,
	).
		WithArgs(
			report.ReporterID,
			report.ReportedUserID,
			report.JobID,
			report.BookingID,
			report.ReportType,
			report.Reason,
			report.Details,
			report.Status,
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
				createdAt,
				createdAt,
			),
		)

	err := repo.Create(
		context.Background(),
		report,
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if report.ID != 10 {
		t.Fatalf(
			"expected id 10, got %d",
			report.ID,
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLRepository_Create_Error(t *testing.T) {
	repo, mock := newReportsSQLMock(t)

	expectedErr := errors.New("insert failed")

	report := &Report{
		ReporterID: 1,
		ReportType: "user",
		Reason:     "reason",
		Status:     "open",
	}

	mock.ExpectQuery(
		`INSERT INTO reports`,
	).
		WithArgs(
			report.ReporterID,
			report.ReportedUserID,
			report.JobID,
			report.BookingID,
			report.ReportType,
			report.Reason,
			report.Details,
			report.Status,
			sqlmock.AnyArg(),
		).
		WillReturnError(expectedErr)

	err := repo.Create(
		context.Background(),
		report,
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected insert error, got %v",
			err,
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLRepository_GetByID_Success(
	t *testing.T,
) {
	repo, mock := newReportsSQLMock(t)

	now := time.Now()
	reportedUserID := int64(20)
	jobID := int64(30)
	bookingID := int64(40)
	reviewedBy := int64(99)

	mock.ExpectQuery(
		`SELECT[\s\S]*FROM reports[\s\S]*WHERE id = \$1`,
	).
		WithArgs(uint(10)).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{
					"id",
					"reporter_id",
					"reported_user_id",
					"job_id",
					"booking_id",
					"report_type",
					"reason",
					"details",
					"status",
					"admin_notes",
					"reviewed_by",
					"reviewed_at",
					"created_at",
					"updated_at",
				},
			).AddRow(
				10,
				5,
				reportedUserID,
				jobID,
				bookingID,
				"user",
				"reason",
				"details",
				"resolved",
				"reviewed",
				reviewedBy,
				now,
				now,
				now,
			),
		)

	report, err := repo.GetByID(
		context.Background(),
		10,
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if report.ID != 10 {
		t.Fatalf(
			"expected id 10, got %d",
			report.ID,
		)
	}

	if report.ReportedUserID == nil ||
		*report.ReportedUserID != 20 {
		t.Fatal(
			"expected reported user id 20",
		)
	}

	if report.JobID == nil || *report.JobID != 30 {
		t.Fatal("expected job id 30")
	}

	if report.BookingID == nil ||
		*report.BookingID != 40 {
		t.Fatal("expected booking id 40")
	}

	if report.ReviewedBy == nil ||
		*report.ReviewedBy != 99 {
		t.Fatal("expected reviewed by 99")
	}

	if report.ReviewedAt == nil {
		t.Fatal("expected reviewed at")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLRepository_GetByID_NullableFields(
	t *testing.T,
) {
	repo, mock := newReportsSQLMock(t)

	now := time.Now()

	mock.ExpectQuery(
		`SELECT[\s\S]*FROM reports[\s\S]*WHERE id = \$1`,
	).
		WithArgs(uint(10)).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{
					"id",
					"reporter_id",
					"reported_user_id",
					"job_id",
					"booking_id",
					"report_type",
					"reason",
					"details",
					"status",
					"admin_notes",
					"reviewed_by",
					"reviewed_at",
					"created_at",
					"updated_at",
				},
			).AddRow(
				10,
				5,
				nil,
				nil,
				nil,
				"spam",
				"reason",
				"",
				"open",
				"",
				nil,
				nil,
				now,
				now,
			),
		)

	report, err := repo.GetByID(
		context.Background(),
		10,
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if report.ReportedUserID != nil {
		t.Fatal(
			"expected nil reported user id",
		)
	}

	if report.JobID != nil {
		t.Fatal("expected nil job id")
	}

	if report.BookingID != nil {
		t.Fatal("expected nil booking id")
	}

	if report.ReviewedBy != nil {
		t.Fatal("expected nil reviewed by")
	}

	if report.ReviewedAt != nil {
		t.Fatal("expected nil reviewed at")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLRepository_GetByID_NotFound(
	t *testing.T,
) {
	repo, mock := newReportsSQLMock(t)

	mock.ExpectQuery(
		`SELECT[\s\S]*FROM reports[\s\S]*WHERE id = \$1`,
	).
		WithArgs(uint(999)).
		WillReturnError(sql.ErrNoRows)

	report, err := repo.GetByID(
		context.Background(),
		999,
	)

	if report != nil {
		t.Fatal("expected nil report")
	}

	if !errors.Is(err, ErrReportNotFound) {
		t.Fatalf(
			"expected ErrReportNotFound, got %v",
			err,
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLRepository_GetByID_Error(t *testing.T) {
	repo, mock := newReportsSQLMock(t)

	expectedErr := errors.New("query failed")

	mock.ExpectQuery(
		`SELECT[\s\S]*FROM reports[\s\S]*WHERE id = \$1`,
	).
		WithArgs(uint(10)).
		WillReturnError(expectedErr)

	_, err := repo.GetByID(
		context.Background(),
		10,
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected query error, got %v",
			err,
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLRepository_Review_Success(t *testing.T) {
	repo, mock := newReportsSQLMock(t)

	mock.ExpectExec(
		regexp.QuoteMeta("UPDATE reports"),
	).
		WithArgs(
			"resolved",
			"reviewed",
			uint(99),
			sqlmock.AnyArg(),
			uint(10),
		).
		WillReturnResult(
			sqlmock.NewResult(0, 1),
		)

	err := repo.Review(
		context.Background(),
		10,
		"resolved",
		"reviewed",
		99,
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLRepository_Review_NotFound(t *testing.T) {
	repo, mock := newReportsSQLMock(t)

	mock.ExpectExec(
		regexp.QuoteMeta("UPDATE reports"),
	).
		WithArgs(
			"resolved",
			"reviewed",
			uint(99),
			sqlmock.AnyArg(),
			uint(999),
		).
		WillReturnResult(
			sqlmock.NewResult(0, 0),
		)

	err := repo.Review(
		context.Background(),
		999,
		"resolved",
		"reviewed",
		99,
	)

	if !errors.Is(err, ErrReportNotFound) {
		t.Fatalf(
			"expected ErrReportNotFound, got %v",
			err,
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLRepository_Review_ExecError(t *testing.T) {
	repo, mock := newReportsSQLMock(t)

	expectedErr := errors.New("update failed")

	mock.ExpectExec(
		regexp.QuoteMeta("UPDATE reports"),
	).
		WithArgs(
			"resolved",
			"reviewed",
			uint(99),
			sqlmock.AnyArg(),
			uint(10),
		).
		WillReturnError(expectedErr)

	err := repo.Review(
		context.Background(),
		10,
		"resolved",
		"reviewed",
		99,
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected update error, got %v",
			err,
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLRepository_Review_RowsAffectedError(
	t *testing.T,
) {
	repo, mock := newReportsSQLMock(t)

	expectedErr := errors.New(
		"rows affected failed",
	)

	mock.ExpectExec(
		regexp.QuoteMeta("UPDATE reports"),
	).
		WithArgs(
			"resolved",
			"reviewed",
			uint(99),
			sqlmock.AnyArg(),
			uint(10),
		).
		WillReturnResult(
			sqlmock.NewErrorResult(expectedErr),
		)

	err := repo.Review(
		context.Background(),
		10,
		"resolved",
		"reviewed",
		99,
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected rows affected error, got %v",
			err,
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
