package availability

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
		t.Fatal("expected db assigned")
	}
}

func TestSQLRepository_Create_Success(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	repo := NewSQLRepository(db)

	mock.ExpectQuery(regexp.QuoteMeta(`
		INSERT INTO cleaner_availability (
			cleaner_id,
			available_date,
			start_time,
			end_time,
			status,
			notes,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $7)
		RETURNING id, created_at, updated_at
	`)).
		WithArgs(
			uint(8),
			"2026-08-10",
			"09:00",
			"17:00",
			"available",
			"",
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
				1,
				time.Now(),
				time.Now(),
			),
		)

	availability := &CleanerAvailability{
		CleanerID:     8,
		AvailableDate: "2026-08-10",
		StartTime:     "09:00",
		EndTime:       "17:00",
		Status:        "available",
	}

	err := repo.Create(
		context.Background(),
		availability,
	)

	if err != nil {
		t.Fatal(err)
	}

	if availability.ID != 1 {
		t.Fatalf(
			"expected ID 1 got %d",
			availability.ID,
		)
	}
}

func TestSQLRepository_Create_Duplicate(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	repo := NewSQLRepository(db)

	mock.ExpectQuery("INSERT INTO cleaner_availability").
		WillReturnError(
			errors.New("duplicate key value"),
		)

	err := repo.Create(
		context.Background(),
		&CleanerAvailability{},
	)

	if !errors.Is(err, ErrAvailabilityExists) {
		t.Fatalf(
			"expected ErrAvailabilityExists got %v",
			err,
		)
	}
}

func TestSQLRepository_Create_DatabaseError(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	repo := NewSQLRepository(db)

	expected := errors.New("database failed")

	mock.ExpectQuery("INSERT INTO cleaner_availability").
		WillReturnError(expected)

	err := repo.Create(
		context.Background(),
		&CleanerAvailability{},
	)

	if !errors.Is(err, expected) {
		t.Fatalf(
			"expected %v got %v",
			expected,
			err,
		)
	}
}

func TestSQLRepository_GetByID_Success(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	repo := NewSQLRepository(db)

	now := time.Now()

	mock.ExpectQuery(
		regexp.QuoteMeta(`
			SELECT
				id,
				cleaner_id,
				available_date::text,
				start_time,
				end_time,
				status,
				COALESCE(notes, ''),
				created_at,
				updated_at
			FROM cleaner_availability
			WHERE id = $1
			LIMIT 1
		`),
	).
		WithArgs(uint(12)).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"id",
				"cleaner_id",
				"available_date",
				"start_time",
				"end_time",
				"status",
				"notes",
				"created_at",
				"updated_at",
			}).AddRow(
				12,
				8,
				"2026-08-10",
				"09:00",
				"17:00",
				"available",
				"Available all day",
				now,
				now,
			),
		)

	availability, err := repo.GetByID(
		context.Background(),
		12,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if availability == nil {
		t.Fatal("expected availability")
	}

	if availability.ID != 12 {
		t.Fatalf("expected ID 12, got %d", availability.ID)
	}

	if availability.CleanerID != 8 {
		t.Fatalf(
			"expected cleaner ID 8, got %d",
			availability.CleanerID,
		)
	}

	if availability.AvailableDate != "2026-08-10" {
		t.Fatalf(
			"expected date 2026-08-10, got %q",
			availability.AvailableDate,
		)
	}
}

func TestSQLRepository_GetByID_NotFound(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	repo := NewSQLRepository(db)

	mock.ExpectQuery(
		regexp.QuoteMeta("FROM cleaner_availability"),
	).
		WithArgs(uint(999)).
		WillReturnError(sql.ErrNoRows)

	availability, err := repo.GetByID(
		context.Background(),
		999,
	)

	if availability != nil {
		t.Fatalf(
			"expected nil availability, got %+v",
			availability,
		)
	}

	if !errors.Is(err, ErrAvailabilityNotFound) {
		t.Fatalf(
			"expected ErrAvailabilityNotFound, got %v",
			err,
		)
	}
}

func TestSQLRepository_GetByID_DatabaseError(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	repo := NewSQLRepository(db)

	expectedErr := errors.New("query failed")

	mock.ExpectQuery(
		regexp.QuoteMeta("FROM cleaner_availability"),
	).
		WithArgs(uint(12)).
		WillReturnError(expectedErr)

	availability, err := repo.GetByID(
		context.Background(),
		12,
	)

	if availability != nil {
		t.Fatalf(
			"expected nil availability, got %+v",
			availability,
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

func TestSQLRepository_ListByCleanerID_Success(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	repo := NewSQLRepository(db)

	now := time.Now()

	rows := sqlmock.NewRows([]string{
		"id",
		"cleaner_id",
		"available_date",
		"start_time",
		"end_time",
		"status",
		"notes",
		"created_at",
		"updated_at",
	}).
		AddRow(
			1,
			8,
			"2026-08-10",
			"09:00",
			"13:00",
			"available",
			"Morning",
			now,
			now,
		).
		AddRow(
			2,
			8,
			"2026-08-10",
			"14:00",
			"18:00",
			"available",
			"Afternoon",
			now,
			now,
		)

	mock.ExpectQuery(
		regexp.QuoteMeta("FROM cleaner_availability"),
	).
		WithArgs(uint(8)).
		WillReturnRows(rows)

	records, err := repo.ListByCleanerID(
		context.Background(),
		8,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(records) != 2 {
		t.Fatalf(
			"expected 2 records, got %d",
			len(records),
		)
	}
}

func TestSQLRepository_ListByCleanerID_Empty(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	repo := NewSQLRepository(db)

	mock.ExpectQuery(
		regexp.QuoteMeta("FROM cleaner_availability"),
	).
		WithArgs(uint(8)).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"id",
				"cleaner_id",
				"available_date",
				"start_time",
				"end_time",
				"status",
				"notes",
				"created_at",
				"updated_at",
			}),
		)

	records, err := repo.ListByCleanerID(
		context.Background(),
		8,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(records) != 0 {
		t.Fatalf(
			"expected no records, got %d",
			len(records),
		)
	}
}

func TestSQLRepository_ListByCleanerID_QueryError(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	repo := NewSQLRepository(db)

	expectedErr := errors.New("query failed")

	mock.ExpectQuery(
		regexp.QuoteMeta("FROM cleaner_availability"),
	).
		WithArgs(uint(8)).
		WillReturnError(expectedErr)

	records, err := repo.ListByCleanerID(
		context.Background(),
		8,
	)

	if records != nil {
		t.Fatalf(
			"expected nil records, got %+v",
			records,
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

func TestSQLRepository_ListByCleanerID_ScanError(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	repo := NewSQLRepository(db)

	now := time.Now()

	rows := sqlmock.NewRows([]string{
		"id",
		"cleaner_id",
		"available_date",
		"start_time",
		"end_time",
		"status",
		"notes",
		"created_at",
		"updated_at",
	}).AddRow(
		"invalid-id",
		8,
		"2026-08-10",
		"09:00",
		"17:00",
		"available",
		"",
		now,
		now,
	)

	mock.ExpectQuery(
		regexp.QuoteMeta("FROM cleaner_availability"),
	).
		WithArgs(uint(8)).
		WillReturnRows(rows)

	records, err := repo.ListByCleanerID(
		context.Background(),
		8,
	)

	if records != nil {
		t.Fatalf(
			"expected nil records, got %+v",
			records,
		)
	}

	if err == nil {
		t.Fatal("expected scan error")
	}
}

func TestSQLRepository_ListByCleanerID_RowsError(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	repo := NewSQLRepository(db)

	now := time.Now()
	expectedErr := errors.New("rows failed")

	rows := sqlmock.NewRows([]string{
		"id",
		"cleaner_id",
		"available_date",
		"start_time",
		"end_time",
		"status",
		"notes",
		"created_at",
		"updated_at",
	}).
		AddRow(
			1,
			8,
			"2026-08-10",
			"09:00",
			"17:00",
			"available",
			"",
			now,
			now,
		).
		RowError(0, expectedErr)

	mock.ExpectQuery(
		regexp.QuoteMeta("FROM cleaner_availability"),
	).
		WithArgs(uint(8)).
		WillReturnRows(rows)

	records, err := repo.ListByCleanerID(
		context.Background(),
		8,
	)

	if records != nil {
		t.Fatalf(
			"expected nil records, got %+v",
			records,
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
	db, mock, _ := sqlmock.New()
	defer db.Close()

	repo := NewSQLRepository(db)

	mock.ExpectExec(
		regexp.QuoteMeta(`
			UPDATE cleaner_availability
			SET
				available_date = $1,
				start_time = $2,
				end_time = $3,
				status = $4,
				notes = $5,
				updated_at = $6
			WHERE id = $7
		`),
	).
		WithArgs(
			"2026-08-11",
			"10:00",
			"18:00",
			"busy",
			"Updated",
			sqlmock.AnyArg(),
			uint(1),
		).
		WillReturnResult(
			sqlmock.NewResult(0, 1),
		)

	err := repo.Update(
		context.Background(),
		&CleanerAvailability{
			ID:            1,
			AvailableDate: "2026-08-11",
			StartTime:     "10:00",
			EndTime:       "18:00",
			Status:        "busy",
			Notes:         "Updated",
		},
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSQLRepository_Update_NotFound(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	repo := NewSQLRepository(db)

	mock.ExpectExec(
		regexp.QuoteMeta("UPDATE cleaner_availability"),
	).
		WillReturnResult(
			sqlmock.NewResult(0, 0),
		)

	err := repo.Update(
		context.Background(),
		&CleanerAvailability{ID: 999},
	)

	if !errors.Is(err, ErrAvailabilityNotFound) {
		t.Fatalf(
			"expected ErrAvailabilityNotFound, got %v",
			err,
		)
	}
}

func TestSQLRepository_Update_ExecError(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	repo := NewSQLRepository(db)

	expectedErr := errors.New("update failed")

	mock.ExpectExec(
		regexp.QuoteMeta("UPDATE cleaner_availability"),
	).
		WillReturnError(expectedErr)

	err := repo.Update(
		context.Background(),
		&CleanerAvailability{ID: 1},
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
	db, mock, _ := sqlmock.New()
	defer db.Close()

	repo := NewSQLRepository(db)

	expectedErr := errors.New("rows affected failed")

	mock.ExpectExec(
		regexp.QuoteMeta("UPDATE cleaner_availability"),
	).
		WillReturnResult(
			sqlmock.NewErrorResult(expectedErr),
		)

	err := repo.Update(
		context.Background(),
		&CleanerAvailability{ID: 1},
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
	db, mock, _ := sqlmock.New()
	defer db.Close()

	repo := NewSQLRepository(db)

	mock.ExpectExec(
		regexp.QuoteMeta(`
			DELETE FROM cleaner_availability
			WHERE id = $1
		`),
	).
		WithArgs(uint(1)).
		WillReturnResult(
			sqlmock.NewResult(0, 1),
		)

	err := repo.Delete(
		context.Background(),
		1,
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSQLRepository_Delete_NotFound(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	repo := NewSQLRepository(db)

	mock.ExpectExec(
		regexp.QuoteMeta("DELETE FROM cleaner_availability"),
	).
		WithArgs(uint(999)).
		WillReturnResult(
			sqlmock.NewResult(0, 0),
		)

	err := repo.Delete(
		context.Background(),
		999,
	)

	if !errors.Is(err, ErrAvailabilityNotFound) {
		t.Fatalf(
			"expected ErrAvailabilityNotFound, got %v",
			err,
		)
	}
}

func TestSQLRepository_Delete_ExecError(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	repo := NewSQLRepository(db)

	expectedErr := errors.New("delete failed")

	mock.ExpectExec(
		regexp.QuoteMeta("DELETE FROM cleaner_availability"),
	).
		WithArgs(uint(1)).
		WillReturnError(expectedErr)

	err := repo.Delete(
		context.Background(),
		1,
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
	db, mock, _ := sqlmock.New()
	defer db.Close()

	repo := NewSQLRepository(db)

	expectedErr := errors.New("rows affected failed")

	mock.ExpectExec(
		regexp.QuoteMeta("DELETE FROM cleaner_availability"),
	).
		WithArgs(uint(1)).
		WillReturnResult(
			sqlmock.NewErrorResult(expectedErr),
		)

	err := repo.Delete(
		context.Background(),
		1,
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected %v, got %v",
			expectedErr,
			err,
		)
	}
}

func TestSQLRepository_HasConflict_True(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	repo := NewSQLRepository(db)

	mock.ExpectQuery(
		regexp.QuoteMeta("SELECT EXISTS"),
	).
		WithArgs(
			uint(8),
			"2026-08-10",
			uint(0),
			"17:00",
			"09:00",
		).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{"exists"},
			).AddRow(true),
		)

	conflict, err := repo.HasConflict(
		context.Background(),
		8,
		"2026-08-10",
		"09:00",
		"17:00",
		0,
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !conflict {
		t.Fatal("expected conflict to be true")
	}
}

func TestSQLRepository_HasConflict_False(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	repo := NewSQLRepository(db)

	mock.ExpectQuery(
		regexp.QuoteMeta("SELECT EXISTS"),
	).
		WithArgs(
			uint(8),
			"2026-08-10",
			uint(1),
			"17:00",
			"09:00",
		).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{"exists"},
			).AddRow(false),
		)

	conflict, err := repo.HasConflict(
		context.Background(),
		8,
		"2026-08-10",
		"09:00",
		"17:00",
		1,
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if conflict {
		t.Fatal("expected conflict to be false")
	}
}

func TestSQLRepository_HasConflict_DatabaseError(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	repo := NewSQLRepository(db)

	expectedErr := errors.New("conflict query failed")

	mock.ExpectQuery(
		regexp.QuoteMeta("SELECT EXISTS"),
	).
		WillReturnError(expectedErr)

	conflict, err := repo.HasConflict(
		context.Background(),
		8,
		"2026-08-10",
		"09:00",
		"17:00",
		0,
	)

	if conflict {
		t.Fatal("expected conflict to be false")
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected %v, got %v",
			expectedErr,
			err,
		)
	}
}

func TestSQLRepository_CreateBlock_Success(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	repo := NewSQLRepository(db)

	startAt := time.Date(
		2026,
		time.August,
		10,
		9,
		0,
		0,
		0,
		time.UTC,
	)

	endAt := time.Date(
		2026,
		time.August,
		10,
		17,
		0,
		0,
		0,
		time.UTC,
	)

	mock.ExpectQuery(
		regexp.QuoteMeta(`
			INSERT INTO availability_blocks (
				cleaner_id,
				start_at,
				end_at,
				reason,
				created_at,
				updated_at
			)
			VALUES ($1, $2, $3, $4, $5, $5)
			RETURNING id
		`),
	).
		WithArgs(
			uint(8),
			startAt,
			endAt,
			"Holiday",
			sqlmock.AnyArg(),
		).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{"id"},
			).AddRow(20),
		)

	block := &AvailabilityBlock{
		CleanerID: 8,
		StartAt:   startAt,
		EndAt:     endAt,
		Reason:    "Holiday",
	}

	err := repo.CreateBlock(
		context.Background(),
		block,
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if block.ID != 20 {
		t.Fatalf(
			"expected block ID 20, got %d",
			block.ID,
		)
	}

	if block.CreatedAt.IsZero() {
		t.Fatal("expected created_at")
	}

	if block.UpdatedAt.IsZero() {
		t.Fatal("expected updated_at")
	}
}

func TestSQLRepository_CreateBlock_DatabaseError(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	repo := NewSQLRepository(db)

	expectedErr := errors.New("create block failed")

	mock.ExpectQuery(
		regexp.QuoteMeta("INSERT INTO availability_blocks"),
	).
		WillReturnError(expectedErr)

	err := repo.CreateBlock(
		context.Background(),
		&AvailabilityBlock{},
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected %v, got %v",
			expectedErr,
			err,
		)
	}
}

func TestSQLRepository_ListBlocksByCleanerID_Success(
	t *testing.T,
) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	repo := NewSQLRepository(db)

	now := time.Now()

	rows := sqlmock.NewRows([]string{
		"id",
		"cleaner_id",
		"start_at",
		"end_at",
		"reason",
		"created_at",
		"updated_at",
	}).
		AddRow(
			1,
			8,
			now,
			now.Add(2*time.Hour),
			"Appointment",
			now,
			now,
		).
		AddRow(
			2,
			8,
			now.Add(24*time.Hour),
			now.Add(26*time.Hour),
			"Holiday",
			now,
			now,
		)

	mock.ExpectQuery(
		regexp.QuoteMeta("FROM availability_blocks"),
	).
		WithArgs(uint(8)).
		WillReturnRows(rows)

	blocks, err := repo.ListBlocksByCleanerID(
		context.Background(),
		8,
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(blocks) != 2 {
		t.Fatalf(
			"expected 2 blocks, got %d",
			len(blocks),
		)
	}
}

func TestSQLRepository_ListBlocksByCleanerID_Empty(
	t *testing.T,
) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	repo := NewSQLRepository(db)

	mock.ExpectQuery(
		regexp.QuoteMeta("FROM availability_blocks"),
	).
		WithArgs(uint(8)).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"id",
				"cleaner_id",
				"start_at",
				"end_at",
				"reason",
				"created_at",
				"updated_at",
			}),
		)

	blocks, err := repo.ListBlocksByCleanerID(
		context.Background(),
		8,
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(blocks) != 0 {
		t.Fatalf(
			"expected no blocks, got %d",
			len(blocks),
		)
	}
}

func TestSQLRepository_ListBlocksByCleanerID_QueryError(
	t *testing.T,
) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	repo := NewSQLRepository(db)

	expectedErr := errors.New("query failed")

	mock.ExpectQuery(
		regexp.QuoteMeta("FROM availability_blocks"),
	).
		WithArgs(uint(8)).
		WillReturnError(expectedErr)

	blocks, err := repo.ListBlocksByCleanerID(
		context.Background(),
		8,
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

func TestSQLRepository_ListBlocksByCleanerID_ScanError(
	t *testing.T,
) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	repo := NewSQLRepository(db)

	now := time.Now()

	rows := sqlmock.NewRows([]string{
		"id",
		"cleaner_id",
		"start_at",
		"end_at",
		"reason",
		"created_at",
		"updated_at",
	}).AddRow(
		"invalid-id",
		8,
		now,
		now.Add(time.Hour),
		"Appointment",
		now,
		now,
	)

	mock.ExpectQuery(
		regexp.QuoteMeta("FROM availability_blocks"),
	).
		WithArgs(uint(8)).
		WillReturnRows(rows)

	blocks, err := repo.ListBlocksByCleanerID(
		context.Background(),
		8,
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

func TestSQLRepository_ListBlocksByCleanerID_RowsError(
	t *testing.T,
) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	repo := NewSQLRepository(db)

	now := time.Now()
	expectedErr := errors.New("rows failed")

	rows := sqlmock.NewRows([]string{
		"id",
		"cleaner_id",
		"start_at",
		"end_at",
		"reason",
		"created_at",
		"updated_at",
	}).
		AddRow(
			1,
			8,
			now,
			now.Add(time.Hour),
			"Appointment",
			now,
			now,
		).
		RowError(0, expectedErr)

	mock.ExpectQuery(
		regexp.QuoteMeta("FROM availability_blocks"),
	).
		WithArgs(uint(8)).
		WillReturnRows(rows)

	blocks, err := repo.ListBlocksByCleanerID(
		context.Background(),
		8,
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

func TestSQLRepository_DeleteBlock_Success(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	repo := NewSQLRepository(db)

	mock.ExpectExec(
		regexp.QuoteMeta(`
			DELETE FROM availability_blocks
			WHERE id = $1
				AND cleaner_id = $2
		`),
	).
		WithArgs(
			uint(1),
			uint(8),
		).
		WillReturnResult(
			sqlmock.NewResult(0, 1),
		)

	err := repo.DeleteBlock(
		context.Background(),
		1,
		8,
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSQLRepository_DeleteBlock_NotFound(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	repo := NewSQLRepository(db)

	mock.ExpectExec(
		regexp.QuoteMeta("DELETE FROM availability_blocks"),
	).
		WithArgs(
			uint(999),
			uint(8),
		).
		WillReturnResult(
			sqlmock.NewResult(0, 0),
		)

	err := repo.DeleteBlock(
		context.Background(),
		999,
		8,
	)

	if !errors.Is(err, ErrAvailabilityNotFound) {
		t.Fatalf(
			"expected ErrAvailabilityNotFound, got %v",
			err,
		)
	}
}

func TestSQLRepository_DeleteBlock_ExecError(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	repo := NewSQLRepository(db)

	expectedErr := errors.New("delete block failed")

	mock.ExpectExec(
		regexp.QuoteMeta("DELETE FROM availability_blocks"),
	).
		WithArgs(
			uint(1),
			uint(8),
		).
		WillReturnError(expectedErr)

	err := repo.DeleteBlock(
		context.Background(),
		1,
		8,
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected %v, got %v",
			expectedErr,
			err,
		)
	}
}

func TestSQLRepository_DeleteBlock_RowsAffectedError(
	t *testing.T,
) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	repo := NewSQLRepository(db)

	expectedErr := errors.New("rows affected failed")

	mock.ExpectExec(
		regexp.QuoteMeta("DELETE FROM availability_blocks"),
	).
		WithArgs(
			uint(1),
			uint(8),
		).
		WillReturnResult(
			sqlmock.NewErrorResult(expectedErr),
		)

	err := repo.DeleteBlock(
		context.Background(),
		1,
		8,
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected %v, got %v",
			expectedErr,
			err,
		)
	}
}
