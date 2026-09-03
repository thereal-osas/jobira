package adminmoderation

import (
	"context"
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

func TestSQLRepository_ListReports_Success(
	t *testing.T,
) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	now := time.Now()
	resolvedAt := now.Add(-time.Hour)

	mock.ExpectQuery(
		regexp.QuoteMeta(`
			SELECT
				id,
				client_id,
				cleaner_id,
				reason,
				details,
				status,
				COALESCE(admin_notes, ''),
				resolved_by,
				resolved_at,
				created_at,
				updated_at
			FROM cleaner_reports
			ORDER BY created_at DESC
		`),
	).WillReturnRows(
		sqlmock.NewRows(
			[]string{
				"id",
				"client_id",
				"cleaner_id",
				"reason",
				"details",
				"status",
				"admin_notes",
				"resolved_by",
				"resolved_at",
				"created_at",
				"updated_at",
			},
		).
			AddRow(
				2,
				5,
				8,
				"No show",
				"Details",
				"resolved",
				"Reviewed",
				9,
				resolvedAt,
				now,
				now,
			).
			AddRow(
				1,
				6,
				10,
				"Behaviour",
				"Details",
				"open",
				"",
				nil,
				nil,
				now.Add(-time.Hour),
				now,
			),
	)

	reports, err := repo.ListReports(
		context.Background(),
	)

	if err != nil {
		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}

	if len(reports) != 2 {
		t.Fatalf(
			"expected 2 reports, got %d",
			len(reports),
		)
	}

	if reports[0].ResolvedBy == nil ||
		*reports[0].ResolvedBy != 9 {
		t.Fatal(
			"expected resolved_by 9",
		)
	}

	if reports[0].ResolvedAt == nil {
		t.Fatal(
			"expected resolved_at",
		)
	}

	if !reports[0].ResolvedAt.Equal(
		resolvedAt,
	) {
		t.Fatalf(
			"expected resolved_at %v, got %v",
			resolvedAt,
			*reports[0].ResolvedAt,
		)
	}

	if reports[1].ResolvedAt != nil {
		t.Fatal(
			"expected nil resolved_at",
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf(
			"unmet expectations: %v",
			err,
		)
	}
}

func TestSQLRepository_ListReports_QueryError(
	t *testing.T,
) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	expectedErr := errors.New(
		"query failed",
	)

	mock.ExpectQuery(
		"SELECT",
	).WillReturnError(expectedErr)

	reports, err := repo.ListReports(
		context.Background(),
	)

	if reports != nil {
		t.Fatalf(
			"expected nil reports, got %+v",
			reports,
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

func TestSQLRepository_ListReports_ScanError(
	t *testing.T,
) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	mock.ExpectQuery(
		"SELECT",
	).WillReturnRows(
		sqlmock.NewRows(
			[]string{
				"id",
				"client_id",
				"cleaner_id",
				"reason",
				"details",
				"status",
				"admin_notes",
				"resolved_by",
				"resolved_at",
				"created_at",
				"updated_at",
			},
		).AddRow(
			"invalid",
			5,
			8,
			"Reason",
			"Details",
			"open",
			"",
			nil,
			nil,
			time.Now(),
			time.Now(),
		),
	)

	reports, err := repo.ListReports(
		context.Background(),
	)

	if reports != nil {
		t.Fatalf(
			"expected nil reports, got %+v",
			reports,
		)
	}

	if err == nil {
		t.Fatal("expected scan error")
	}
}

func TestSQLRepository_ListOpenReports_Success(
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
		regexp.QuoteMeta(`
			SELECT
				id,
				client_id,
				cleaner_id,
				reason,
				details,
				status,
				COALESCE(admin_notes, ''),
				resolved_by,
				resolved_at,
				created_at,
				updated_at
			FROM cleaner_reports
			WHERE status = 'open'
			ORDER BY created_at DESC
		`),
	).WillReturnRows(
		sqlmock.NewRows(
			[]string{
				"id",
				"client_id",
				"cleaner_id",
				"reason",
				"details",
				"status",
				"admin_notes",
				"resolved_by",
				"resolved_at",
				"created_at",
				"updated_at",
			},
		).AddRow(
			1,
			5,
			8,
			"No show",
			"Details",
			"open",
			"",
			nil,
			nil,
			now,
			now,
		),
	)

	reports, err := repo.ListOpenReports(
		context.Background(),
	)

	if err != nil {
		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}

	if len(reports) != 1 {
		t.Fatalf(
			"expected 1 report, got %d",
			len(reports),
		)
	}

	if reports[0].Status != "open" {
		t.Fatalf(
			"expected open status, got %q",
			reports[0].Status,
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf(
			"unmet expectations: %v",
			err,
		)
	}
}

func TestSQLRepository_ListOpenReports_QueryError(
	t *testing.T,
) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	expectedErr := errors.New(
		"query failed",
	)

	mock.ExpectQuery(
		"SELECT",
	).WillReturnError(expectedErr)

	reports, err := repo.ListOpenReports(
		context.Background(),
	)

	if reports != nil {
		t.Fatalf(
			"expected nil reports, got %+v",
			reports,
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

func TestSQLRepository_UpdateReportStatus_Success(
	t *testing.T,
) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	mock.ExpectExec(
		regexp.QuoteMeta(`
			UPDATE cleaner_reports
			SET
				status = $1,
				admin_notes = $2,
				resolved_by = $3,
				resolved_at = $4,
				updated_at = $5
			WHERE id = $6
		`),
	).
		WithArgs(
			"resolved",
			"Reviewed",
			uint(9),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			uint(12),
		).
		WillReturnResult(
			sqlmock.NewResult(0, 1),
		)

	err = repo.UpdateReportStatus(
		context.Background(),
		12,
		9,
		"resolved",
		"Reviewed",
	)

	if err != nil {
		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf(
			"unmet expectations: %v",
			err,
		)
	}
}

func TestSQLRepository_UpdateReportStatus_NotFound(
	t *testing.T,
) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	mock.ExpectExec(
		"UPDATE cleaner_reports",
	).
		WithArgs(
			"resolved",
			"Reviewed",
			uint(9),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			uint(999),
		).
		WillReturnResult(
			sqlmock.NewResult(0, 0),
		)

	err = repo.UpdateReportStatus(
		context.Background(),
		999,
		9,
		"resolved",
		"Reviewed",
	)

	if !errors.Is(
		err,
		ErrReportNotFound,
	) {
		t.Fatalf(
			"expected ErrReportNotFound, got %v",
			err,
		)
	}
}

func TestSQLRepository_UpdateReportStatus_DatabaseError(
	t *testing.T,
) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	expectedErr := errors.New(
		"update failed",
	)

	mock.ExpectExec(
		"UPDATE cleaner_reports",
	).
		WithArgs(
			"resolved",
			"Reviewed",
			uint(9),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			uint(12),
		).
		WillReturnError(expectedErr)

	err = repo.UpdateReportStatus(
		context.Background(),
		12,
		9,
		"resolved",
		"Reviewed",
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected %v, got %v",
			expectedErr,
			err,
		)
	}
}

func TestSQLRepository_ListBlockCleaners_Success(
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
		regexp.QuoteMeta(`
			SELECT
				id,
				client_id,
				cleaner_id,
				reason,
				created_at
			FROM blocked_cleaners
			ORDER BY created_at DESC
		`),
	).WillReturnRows(
		sqlmock.NewRows(
			[]string{
				"id",
				"client_id",
				"cleaner_id",
				"reason",
				"created_at",
			},
		).
			AddRow(
				2,
				5,
				8,
				"Unprofessional",
				now,
			).
			AddRow(
				1,
				6,
				9,
				"No show",
				now.Add(-time.Hour),
			),
	)

	blocks, err := repo.ListBlockCleaners(
		context.Background(),
	)

	if err != nil {
		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}

	if len(blocks) != 2 {
		t.Fatalf(
			"expected 2 blocks, got %d",
			len(blocks),
		)
	}

	if blocks[0].CleanerID != 8 {
		t.Fatalf(
			"expected cleaner ID 8, got %d",
			blocks[0].CleanerID,
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf(
			"unmet expectations: %v",
			err,
		)
	}
}

func TestSQLRepository_ListBlockCleaners_QueryError(
	t *testing.T,
) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	expectedErr := errors.New(
		"query failed",
	)

	mock.ExpectQuery(
		"SELECT",
	).WillReturnError(expectedErr)

	blocks, err := repo.ListBlockCleaners(
		context.Background(),
	)

	if blocks != nil {
		t.Fatalf(
			"expected nil blocks, got %+v",
			blocks,
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

func TestSQLRepository_ListBlockCleaners_ScanError(
	t *testing.T,
) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	mock.ExpectQuery(
		"SELECT",
	).WillReturnRows(
		sqlmock.NewRows(
			[]string{
				"id",
				"client_id",
				"cleaner_id",
				"reason",
				"created_at",
			},
		).AddRow(
			"invalid",
			5,
			8,
			"Reason",
			time.Now(),
		),
	)

	blocks, err := repo.ListBlockCleaners(
		context.Background(),
	)

	if blocks != nil {
		t.Fatalf(
			"expected nil blocks, got %+v",
			blocks,
		)
	}

	if err == nil {
		t.Fatal("expected scan error")
	}
}
