package reports

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func reportColumns() []string {
	return []string{
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
	}
}

func TestSQLRepository_ListByReporterID_Success(
	t *testing.T,
) {
	repo, mock := newReportsSQLMock(t)

	now := time.Now()

	rows := sqlmock.NewRows(reportColumns()).
		AddRow(
			2,
			5,
			20,
			30,
			40,
			"user",
			"reason two",
			"details two",
			"resolved",
			"reviewed",
			99,
			now,
			now,
			now,
		).
		AddRow(
			1,
			5,
			nil,
			nil,
			nil,
			"spam",
			"reason one",
			"",
			"open",
			"",
			nil,
			nil,
			now,
			now,
		)

	mock.ExpectQuery(
		`SELECT[\s\S]*FROM reports[\s\S]*WHERE reporter_id = \$1[\s\S]*ORDER BY created_at DESC`,
	).
		WithArgs(uint(5)).
		WillReturnRows(rows)

	reports, err := repo.ListByReporterID(
		context.Background(),
		5,
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(reports) != 2 {
		t.Fatalf(
			"expected 2 reports, got %d",
			len(reports),
		)
	}

	if reports[0].ID != 2 {
		t.Fatalf(
			"expected first report id 2, got %d",
			reports[0].ID,
		)
	}

	if reports[0].ReportedUserID == nil ||
		*reports[0].ReportedUserID != 20 {
		t.Fatal("expected reported user id 20")
	}

	if reports[0].JobID == nil ||
		*reports[0].JobID != 30 {
		t.Fatal("expected job id 30")
	}

	if reports[0].BookingID == nil ||
		*reports[0].BookingID != 40 {
		t.Fatal("expected booking id 40")
	}

	if reports[0].ReviewedBy == nil ||
		*reports[0].ReviewedBy != 99 {
		t.Fatal("expected reviewed by 99")
	}

	if reports[0].ReviewedAt == nil {
		t.Fatal("expected reviewed at")
	}

	if reports[1].ReportedUserID != nil {
		t.Fatal(
			"expected nil reported user id",
		)
	}

	if reports[1].JobID != nil {
		t.Fatal("expected nil job id")
	}

	if reports[1].BookingID != nil {
		t.Fatal("expected nil booking id")
	}

	if reports[1].ReviewedBy != nil {
		t.Fatal("expected nil reviewed by")
	}

	if reports[1].ReviewedAt != nil {
		t.Fatal("expected nil reviewed at")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLRepository_ListByReporterID_Empty(
	t *testing.T,
) {
	repo, mock := newReportsSQLMock(t)

	mock.ExpectQuery(
		`SELECT[\s\S]*FROM reports[\s\S]*WHERE reporter_id = \$1`,
	).
		WithArgs(uint(5)).
		WillReturnRows(
			sqlmock.NewRows(reportColumns()),
		)

	reports, err := repo.ListByReporterID(
		context.Background(),
		5,
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(reports) != 0 {
		t.Fatalf(
			"expected 0 reports, got %d",
			len(reports),
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLRepository_ListByReporterID_QueryError(
	t *testing.T,
) {
	repo, mock := newReportsSQLMock(t)

	expectedErr := errors.New("query failed")

	mock.ExpectQuery(
		`SELECT[\s\S]*FROM reports[\s\S]*WHERE reporter_id = \$1`,
	).
		WithArgs(uint(5)).
		WillReturnError(expectedErr)

	_, err := repo.ListByReporterID(
		context.Background(),
		5,
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

func TestSQLRepository_ListByReporterID_ScanError(
	t *testing.T,
) {
	repo, mock := newReportsSQLMock(t)

	now := time.Now()

	rows := sqlmock.NewRows(reportColumns()).
		AddRow(
			"invalid-id",
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
		)

	mock.ExpectQuery(
		`SELECT[\s\S]*FROM reports[\s\S]*WHERE reporter_id = \$1`,
	).
		WithArgs(uint(5)).
		WillReturnRows(rows)

	_, err := repo.ListByReporterID(
		context.Background(),
		5,
	)

	if err == nil {
		t.Fatal("expected scan error")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLRepository_ListAll_Success(t *testing.T) {
	repo, mock := newReportsSQLMock(t)

	now := time.Now()

	rows := sqlmock.NewRows(reportColumns()).
		AddRow(
			2,
			6,
			nil,
			30,
			nil,
			"job",
			"bad listing",
			"details",
			"reviewing",
			"checking",
			99,
			now,
			now,
			now,
		).
		AddRow(
			1,
			5,
			20,
			nil,
			nil,
			"user",
			"abuse",
			"",
			"open",
			"",
			nil,
			nil,
			now,
			now,
		)

	mock.ExpectQuery(
		`SELECT[\s\S]*FROM reports[\s\S]*ORDER BY created_at DESC`,
	).
		WillReturnRows(rows)

	reports, err := repo.ListAll(
		context.Background(),
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(reports) != 2 {
		t.Fatalf(
			"expected 2 reports, got %d",
			len(reports),
		)
	}

	if reports[0].Status != "reviewing" {
		t.Fatalf(
			"expected reviewing, got %q",
			reports[0].Status,
		)
	}

	if reports[1].ReportedUserID == nil ||
		*reports[1].ReportedUserID != 20 {
		t.Fatal(
			"expected reported user id 20",
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLRepository_ListAll_Empty(t *testing.T) {
	repo, mock := newReportsSQLMock(t)

	mock.ExpectQuery(
		`SELECT[\s\S]*FROM reports[\s\S]*ORDER BY created_at DESC`,
	).
		WillReturnRows(
			sqlmock.NewRows(reportColumns()),
		)

	reports, err := repo.ListAll(
		context.Background(),
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(reports) != 0 {
		t.Fatalf(
			"expected 0 reports, got %d",
			len(reports),
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLRepository_ListAll_QueryError(
	t *testing.T,
) {
	repo, mock := newReportsSQLMock(t)

	expectedErr := errors.New("query failed")

	mock.ExpectQuery(
		`SELECT[\s\S]*FROM reports[\s\S]*ORDER BY created_at DESC`,
	).
		WillReturnError(expectedErr)

	_, err := repo.ListAll(
		context.Background(),
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

func TestSQLRepository_ListAll_ScanError(t *testing.T) {
	repo, mock := newReportsSQLMock(t)

	now := time.Now()

	rows := sqlmock.NewRows(reportColumns()).
		AddRow(
			"invalid-id",
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
		)

	mock.ExpectQuery(
		`SELECT[\s\S]*FROM reports[\s\S]*ORDER BY created_at DESC`,
	).
		WillReturnRows(rows)

	_, err := repo.ListAll(
		context.Background(),
	)

	if err == nil {
		t.Fatal("expected scan error")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLRepository_ListOpen_Success(t *testing.T) {
	repo, mock := newReportsSQLMock(t)

	now := time.Now()

	rows := sqlmock.NewRows(reportColumns()).
		AddRow(
			1,
			5,
			nil,
			nil,
			40,
			"booking",
			"no show",
			"cleaner did not arrive",
			"open",
			"",
			nil,
			nil,
			now,
			now,
		)

	mock.ExpectQuery(
		`SELECT[\s\S]*FROM reports[\s\S]*WHERE status = 'open'[\s\S]*ORDER BY created_at DESC`,
	).
		WillReturnRows(rows)

	reports, err := repo.ListOpen(
		context.Background(),
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(reports) != 1 {
		t.Fatalf(
			"expected 1 report, got %d",
			len(reports),
		)
	}

	if reports[0].Status != "open" {
		t.Fatalf(
			"expected open, got %q",
			reports[0].Status,
		)
	}

	if reports[0].BookingID == nil ||
		*reports[0].BookingID != 40 {
		t.Fatal("expected booking id 40")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLRepository_ListOpen_Empty(t *testing.T) {
	repo, mock := newReportsSQLMock(t)

	mock.ExpectQuery(
		`SELECT[\s\S]*FROM reports[\s\S]*WHERE status = 'open'`,
	).
		WillReturnRows(
			sqlmock.NewRows(reportColumns()),
		)

	reports, err := repo.ListOpen(
		context.Background(),
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(reports) != 0 {
		t.Fatalf(
			"expected 0 reports, got %d",
			len(reports),
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLRepository_ListOpen_QueryError(
	t *testing.T,
) {
	repo, mock := newReportsSQLMock(t)

	expectedErr := errors.New("query failed")

	mock.ExpectQuery(
		`SELECT[\s\S]*FROM reports[\s\S]*WHERE status = 'open'`,
	).
		WillReturnError(expectedErr)

	_, err := repo.ListOpen(
		context.Background(),
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

func TestSQLRepository_ListOpen_ScanError(
	t *testing.T,
) {
	repo, mock := newReportsSQLMock(t)

	now := time.Now()

	rows := sqlmock.NewRows(reportColumns()).
		AddRow(
			"invalid-id",
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
		)

	mock.ExpectQuery(
		`SELECT[\s\S]*FROM reports[\s\S]*WHERE status = 'open'`,
	).
		WillReturnRows(rows)

	_, err := repo.ListOpen(
		context.Background(),
	)

	if err == nil {
		t.Fatal("expected scan error")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
