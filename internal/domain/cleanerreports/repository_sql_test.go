package cleanerreports

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
		t.Fatalf(
			"failed to create sqlmock database: %v",
			err,
		)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	if repo == nil {
		t.Fatal("expected repository")
	}

	if repo.db != db {
		t.Fatal(
			"expected database to be assigned",
		)
	}
}

func TestSQLRepository_Create_Success(
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
			INSERT INTO cleaner_reports (
				client_id,
				cleaner_id,
				reason,
				details,
				status,
				created_at,
				updated_at
			)
				VALUES ($1, $2, $3, $4, $5, $6, $7)
				RETURNING id, status, created_at, updated_at
		`),
	).
		WithArgs(
			uint(5),
			uint(8),
			"No show",
			"Cleaner did not attend.",
			"open",
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
		).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{
					"id",
					"status",
					"created_at",
					"updated_at",
				},
			).AddRow(
				20,
				"open",
				now,
				now,
			),
		)

	report := &CleanerReport{
		ClientID:  5,
		CleanerID: 8,
		Reason:    "No show",
		Details:   "Cleaner did not attend.",
	}

	err = repo.Create(
		context.Background(),
		report,
	)

	if err != nil {
		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}

	if report.ID != 20 {
		t.Fatalf(
			"expected report ID 20, got %d",
			report.ID,
		)
	}

	if report.Status != "open" {
		t.Fatalf(
			"expected status open, got %q",
			report.Status,
		)
	}

	if !report.CreatedAt.Equal(now) {
		t.Fatalf(
			"expected created_at %v, got %v",
			now,
			report.CreatedAt,
		)
	}

	if !report.UpdatedAt.Equal(now) {
		t.Fatalf(
			"expected updated_at %v, got %v",
			now,
			report.UpdatedAt,
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf(
			"unmet expectations: %v",
			err,
		)
	}
}

func TestSQLRepository_Create_DatabaseError(
	t *testing.T,
) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	expectedErr := errors.New(
		"insert report failed",
	)

	mock.ExpectQuery(
		regexp.QuoteMeta(`
			INSERT INTO cleaner_reports (
				client_id,
				cleaner_id,
				reason,
				details,
				status,
				created_at,
				updated_at
			)
				VALUES ($1, $2, $3, $4, $5, $6, $7)
				RETURNING id, status, created_at, updated_at
		`),
	).
		WithArgs(
			uint(5),
			uint(8),
			"No show",
			"Cleaner did not attend.",
			"open",
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
		).
		WillReturnError(expectedErr)

	report := &CleanerReport{
		ClientID:  5,
		CleanerID: 8,
		Reason:    "No show",
		Details:   "Cleaner did not attend.",
	}

	err = repo.Create(
		context.Background(),
		report,
	)

	if !errors.Is(
		err,
		expectedErr,
	) {
		t.Fatalf(
			"expected %v, got %v",
			expectedErr,
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

func TestSQLRepository_Create_ScanError(
	t *testing.T,
) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	mock.ExpectQuery(
		regexp.QuoteMeta(`
			INSERT INTO cleaner_reports (
				client_id,
				cleaner_id,
				reason,
				details,
				status,
				created_at,
				updated_at
			)
				VALUES ($1, $2, $3, $4, $5, $6, $7)
				RETURNING id, status, created_at, updated_at
		`),
	).
		WithArgs(
			uint(5),
			uint(8),
			"No show",
			"Cleaner did not attend.",
			"open",
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
		).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{
					"id",
					"status",
					"created_at",
					"updated_at",
				},
			).AddRow(
				"invalid-id",
				"open",
				time.Now(),
				time.Now(),
			),
		)

	report := &CleanerReport{
		ClientID:  5,
		CleanerID: 8,
		Reason:    "No show",
		Details:   "Cleaner did not attend.",
	}

	err = repo.Create(
		context.Background(),
		report,
	)

	if err == nil {
		t.Fatal(
			"expected scan error",
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf(
			"unmet expectations: %v",
			err,
		)
	}
}

func TestSQLRepository_ListByClientID_Success(
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
				created_at,
				updated_at
			FROM cleaner_reports
			WHERE client_id = $1
			ORDER BY created_at DESC
		`),
	).
		WithArgs(uint(5)).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{
					"id",
					"client_id",
					"cleaner_id",
					"reason",
					"details",
					"status",
					"created_at",
					"updated_at",
				},
			).
				AddRow(
					2,
					5,
					9,
					"Poor behaviour",
					"Cleaner was rude.",
					"open",
					now,
					now,
				).
				AddRow(
					1,
					5,
					8,
					"No show",
					"Cleaner did not attend.",
					"closed",
					now.Add(-time.Hour),
					now,
				),
		)

	reports, err := repo.ListByClientID(
		context.Background(),
		5,
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

	if reports[0].ID != 2 {
		t.Fatalf(
			"expected first report ID 2, got %d",
			reports[0].ID,
		)
	}

	if reports[0].CleanerID != 9 {
		t.Fatalf(
			"expected cleaner ID 9, got %d",
			reports[0].CleanerID,
		)
	}

	if reports[1].ID != 1 {
		t.Fatalf(
			"expected second report ID 1, got %d",
			reports[1].ID,
		)
	}

	if reports[1].Status != "closed" {
		t.Fatalf(
			"expected status closed, got %q",
			reports[1].Status,
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf(
			"unmet expectations: %v",
			err,
		)
	}
}

func TestSQLRepository_ListByClientID_Empty(
	t *testing.T,
) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	mock.ExpectQuery(
		regexp.QuoteMeta(`
			SELECT
				id,
				client_id,
				cleaner_id,
				reason,
				details,
				status,
				created_at,
				updated_at
			FROM cleaner_reports
			WHERE client_id = $1
			ORDER BY created_at DESC
		`),
	).
		WithArgs(uint(5)).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{
					"id",
					"client_id",
					"cleaner_id",
					"reason",
					"details",
					"status",
					"created_at",
					"updated_at",
				},
			),
		)

	reports, err := repo.ListByClientID(
		context.Background(),
		5,
	)

	if err != nil {
		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}

	if len(reports) != 0 {
		t.Fatalf(
			"expected no reports, got %d",
			len(reports),
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf(
			"unmet expectations: %v",
			err,
		)
	}
}

func TestSQLRepository_ListByClientID_QueryError(
	t *testing.T,
) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	expectedErr := errors.New(
		"list reports failed",
	)

	mock.ExpectQuery(
		regexp.QuoteMeta(`
			SELECT
				id,
				client_id,
				cleaner_id,
				reason,
				details,
				status,
				created_at,
				updated_at
			FROM cleaner_reports
			WHERE client_id = $1
			ORDER BY created_at DESC
		`),
	).
		WithArgs(uint(5)).
		WillReturnError(expectedErr)

	reports, err := repo.ListByClientID(
		context.Background(),
		5,
	)

	if reports != nil {
		t.Fatalf(
			"expected nil reports, got %+v",
			reports,
		)
	}

	if !errors.Is(
		err,
		expectedErr,
	) {
		t.Fatalf(
			"expected %v, got %v",
			expectedErr,
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

func TestSQLRepository_ListByClientID_ScanError(
	t *testing.T,
) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	mock.ExpectQuery(
		regexp.QuoteMeta(`
			SELECT
				id,
				client_id,
				cleaner_id,
				reason,
				details,
				status,
				created_at,
				updated_at
			FROM cleaner_reports
			WHERE client_id = $1
			ORDER BY created_at DESC
		`),
	).
		WithArgs(uint(5)).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{
					"id",
					"client_id",
					"cleaner_id",
					"reason",
					"details",
					"status",
					"created_at",
					"updated_at",
				},
			).AddRow(
				"invalid-id",
				5,
				8,
				"No show",
				"Details",
				"open",
				time.Now(),
				time.Now(),
			),
		)

	reports, err := repo.ListByClientID(
		context.Background(),
		5,
	)

	if reports != nil {
		t.Fatalf(
			"expected nil reports, got %+v",
			reports,
		)
	}

	if err == nil {
		t.Fatal(
			"expected scan error",
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf(
			"unmet expectations: %v",
			err,
		)
	}
}

func TestSQLRepository_ListByClientID_RowsError(
	t *testing.T,
) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	now := time.Now()

	rows := sqlmock.NewRows(
		[]string{
			"id",
			"client_id",
			"cleaner_id",
			"reason",
			"details",
			"status",
			"created_at",
			"updated_at",
		},
	).
		AddRow(
			1,
			5,
			8,
			"No show",
			"Details",
			"open",
			now,
			now,
		).
		RowError(
			0,
			errors.New("rows iteration failed"),
		)

	mock.ExpectQuery(
		regexp.QuoteMeta(`
			SELECT
				id,
				client_id,
				cleaner_id,
				reason,
				details,
				status,
				created_at,
				updated_at
			FROM cleaner_reports
			WHERE client_id = $1
			ORDER BY created_at DESC
		`),
	).
		WithArgs(uint(5)).
		WillReturnRows(rows)

	reports, err := repo.ListByClientID(
		context.Background(),
		5,
	)

	if reports != nil {
		t.Fatalf(
			"expected nil reports, got %+v",
			reports,
		)
	}

	if err == nil {
		t.Fatal(
			"expected rows error",
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf(
			"unmet expectations: %v",
			err,
		)
	}
}
