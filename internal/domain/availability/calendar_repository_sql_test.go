package availability

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func newCalendarSQLRepositoryMock(
	t *testing.T,
) (*SQLRepository, sqlmock.Sqlmock) {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}

	t.Cleanup(func() {
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unmet sqlmock expectations: %v", err)
		}

		_ = db.Close()
	})

	return NewSQLRepository(db), mock
}

func TestCalendarSQLRepository_ListByCleanerIDRange_Success(t *testing.T) {
	repo, mock := newCalendarSQLRepositoryMock(t)

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
		1,
		8,
		"2026-12-10",
		"09:00",
		"17:00",
		"available",
		"full day",
		now,
		now,
	)

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
			WHERE cleaner_id = $1
				AND available_date >= $2 
				AND available_date <= $3 
			ORDER BY available_date ASC, start_time ASC
		`),
	).
		WithArgs(
			uint(8),
			"2026-12-01",
			"2026-12-31",
		).
		WillReturnRows(rows)

	records, err := repo.ListByCleanerIDRange(
		context.Background(),
		8,
		"2026-12-01",
		"2026-12-31",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(records) != 1 {
		t.Fatalf(
			"expected 1 record, got %d",
			len(records),
		)
	}

	if records[0].AvailableDate != "2026-12-10" {
		t.Fatalf(
			"unexpected date %s",
			records[0].AvailableDate,
		)
	}
}

func TestCalendarSQLRepository_ListByCleanerIDRange_QueryError(t *testing.T) {
	repo, mock := newCalendarSQLRepositoryMock(t)

	expectedErr := errors.New("range query failed")

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
			WHERE cleaner_id = $1
				AND available_date >= $2 
				AND available_date <= $3 
			ORDER BY available_date ASC, start_time ASC
		`),
	).
		WithArgs(
			uint(8),
			"2026-12-01",
			"2026-12-31",
		).
		WillReturnError(expectedErr)

	_, err := repo.ListByCleanerIDRange(
		context.Background(),
		8,
		"2026-12-01",
		"2026-12-31",
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected %v, got %v",
			expectedErr,
			err,
		)
	}
}

func TestCalendarSQLRepository_ListByCleanerIDRange_ScanError(t *testing.T) {
	repo, mock := newCalendarSQLRepositoryMock(t)

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
		"bad-id",
		8,
		"2026-12-10",
		"09:00",
		"17:00",
		"available",
		"",
		now,
		now,
	)

	mock.ExpectQuery("FROM cleaner_availability").
		WithArgs(
			uint(8),
			"2026-12-01",
			"2026-12-31",
		).
		WillReturnRows(rows)

	records, err := repo.ListByCleanerIDRange(
		context.Background(),
		8,
		"2026-12-01",
		"2026-12-31",
	)

	if records != nil {
		t.Fatalf("expected nil records, got %+v", records)
	}

	if err == nil {
		t.Fatal("expected scan error")
	}
}

func TestCalendarSQLRepository_ListBlocksByCleanerIDRange_Success(t *testing.T) {
	repo, mock := newCalendarSQLRepositoryMock(t)

	startAt := time.Date(
		2026,
		time.December,
		10,
		9,
		0,
		0,
		0,
		time.UTC,
	)

	endAt := startAt.Add(8 * time.Hour)
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
		2,
		8,
		startAt.Add(time.Hour),
		startAt.Add(2*time.Hour),
		"appointment",
		now,
		now,
	)

	mock.ExpectQuery(
		regexp.QuoteMeta(`
			SELECT 
				id, 
				cleaner_id, 
				start_at, 
				end_at, 
				reason, 
				created_at, 
				updated_at
			FROM availability_blocks
			WHERE cleaner_id = $1 
				AND start_at < $3 
				AND end_at > $2 
			ORDER BY start_at ASC
		`),
	).
		WithArgs(
			uint(8),
			startAt,
			endAt,
		).
		WillReturnRows(rows)

	blocks, err := repo.ListBlocksByCleanerIDRange(
		context.Background(),
		8,
		startAt,
		endAt,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(blocks) != 1 {
		t.Fatalf(
			"expected 1 block, got %d",
			len(blocks),
		)
	}
}

func TestCalendarSQLRepository_ListBlocksByCleanerIDRange_QueryError(t *testing.T) {
	repo, mock := newCalendarSQLRepositoryMock(t)

	startAt := time.Now()
	endAt := startAt.Add(time.Hour)
	expectedErr := errors.New("block range query failed")

	mock.ExpectQuery("FROM availability_blocks").
		WithArgs(
			uint(8),
			startAt,
			endAt,
		).
		WillReturnError(expectedErr)

	_, err := repo.ListBlocksByCleanerIDRange(
		context.Background(),
		8,
		startAt,
		endAt,
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected %v, got %v",
			expectedErr,
			err,
		)
	}
}

func TestCalendarSQLRepository_CreateRecurring_Success(t *testing.T) {
	repo, mock := newCalendarSQLRepositoryMock(t)

	now := time.Now()

	mock.ExpectQuery(
		regexp.QuoteMeta(`
			INSERT INTO recurring_availability (
				cleaner_id, 
				weekday,
				start_time,
				end_time,
				status, 
				created_at, 
				updated_at
			) 
			VALUES ($1, $2, $3, $4, $5, $6, $6)
			RETURNING id, created_at, updated_at
		`),
	).
		WithArgs(
			uint(8),
			1,
			"09:00",
			"17:00",
			"available",
			sqlmock.AnyArg(),
		).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"id",
				"created_at",
				"updated_at",
			}).AddRow(
				10,
				now,
				now,
			),
		)

	recurring := &RecurringAvailability{
		CleanerID: 8,
		Weekday:   1,
		StartTime: "09:00",
		EndTime:   "17:00",
		Status:    "available",
	}

	err := repo.CreateRecurring(
		context.Background(),
		recurring,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if recurring.ID != 10 {
		t.Fatalf(
			"expected id 10, got %d",
			recurring.ID,
		)
	}
}

func TestCalendarSQLRepository_CreateRecurring_Error(t *testing.T) {
	repo, mock := newCalendarSQLRepositoryMock(t)

	expectedErr := errors.New("create recurring failed")

	mock.ExpectQuery("INSERT INTO recurring_availability").
		WillReturnError(expectedErr)

	err := repo.CreateRecurring(
		context.Background(),
		&RecurringAvailability{},
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected %v, got %v",
			expectedErr,
			err,
		)
	}
}

func TestCalendarSQLRepository_ListRecurringByCleanerID_Success(t *testing.T) {
	repo, mock := newCalendarSQLRepositoryMock(t)

	now := time.Now()

	rows := sqlmock.NewRows([]string{
		"id",
		"cleaner_id",
		"weekday",
		"start_time",
		"end_time",
		"status",
		"created_at",
		"updated_at",
	}).
		AddRow(
			1,
			8,
			1,
			"09:00",
			"17:00",
			"available",
			now,
			now,
		).
		AddRow(
			2,
			8,
			3,
			"10:00",
			"15:00",
			"available",
			now,
			now,
		)

	mock.ExpectQuery(
		regexp.QuoteMeta(`
			SELECT 	
				id, 
				cleaner_id, 
				weekday, 
				start_time, 
				end_time, 
				status, 
				created_at,
				updated_at 
			FROM recurring_availability 
			WHERE cleaner_id = $1 
			ORDER BY weekday ASC, start_time ASC
		`),
	).
		WithArgs(uint(8)).
		WillReturnRows(rows)

	records, err := repo.ListRecurringByCleanerID(
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

func TestCalendarSQLRepository_ListRecurringByCleanerID_QueryError(t *testing.T) {
	repo, mock := newCalendarSQLRepositoryMock(t)

	expectedErr := errors.New("recurring query failed")

	mock.ExpectQuery("FROM recurring_availability").
		WithArgs(uint(8)).
		WillReturnError(expectedErr)

	_, err := repo.ListRecurringByCleanerID(
		context.Background(),
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

func TestCalendarSQLRepository_ListRecurringByCleanerID_ScanError(t *testing.T) {
	repo, mock := newCalendarSQLRepositoryMock(t)

	now := time.Now()

	rows := sqlmock.NewRows([]string{
		"id",
		"cleaner_id",
		"weekday",
		"start_time",
		"end_time",
		"status",
		"created_at",
		"updated_at",
	}).AddRow(
		"bad-id",
		8,
		1,
		"09:00",
		"17:00",
		"available",
		now,
		now,
	)

	mock.ExpectQuery("FROM recurring_availability").
		WithArgs(uint(8)).
		WillReturnRows(rows)

	records, err := repo.ListRecurringByCleanerID(
		context.Background(),
		8,
	)

	if records != nil {
		t.Fatalf("expected nil records, got %+v", records)
	}

	if err == nil {
		t.Fatal("expected scan error")
	}
}

func TestCalendarSQLRepository_DeleteRecurring_Success(t *testing.T) {
	repo, mock := newCalendarSQLRepositoryMock(t)

	mock.ExpectExec(
		regexp.QuoteMeta(`
			DELETE FROM recurring_availability
			WHERE id = $1
				AND cleaner_id = $2
		`),
	).
		WithArgs(
			uint(4),
			uint(8),
		).
		WillReturnResult(
			sqlmock.NewResult(0, 1),
		)

	err := repo.DeleteRecurring(
		context.Background(),
		4,
		8,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCalendarSQLRepository_DeleteRecurring_NotFound(t *testing.T) {
	repo, mock := newCalendarSQLRepositoryMock(t)

	mock.ExpectExec("DELETE FROM recurring_availability").
		WithArgs(
			uint(4),
			uint(8),
		).
		WillReturnResult(
			sqlmock.NewResult(0, 0),
		)

	err := repo.DeleteRecurring(
		context.Background(),
		4,
		8,
	)

	if !errors.Is(
		err,
		ErrRecurringAvailabilityNotFound,
	) {
		t.Fatalf(
			"expected ErrRecurringAvailabilityNotFound, got %v",
			err,
		)
	}
}

func TestCalendarSQLRepository_DeleteRecurring_ExecError(t *testing.T) {
	repo, mock := newCalendarSQLRepositoryMock(t)

	expectedErr := errors.New("delete recurring failed")

	mock.ExpectExec("DELETE FROM recurring_availability").
		WithArgs(
			uint(4),
			uint(8),
		).
		WillReturnError(expectedErr)

	err := repo.DeleteRecurring(
		context.Background(),
		4,
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

func TestCalendarSQLRepository_UpsertSettings_Success(t *testing.T) {
	repo, mock := newCalendarSQLRepositoryMock(t)

	now := time.Now()

	mock.ExpectQuery(
		regexp.QuoteMeta(`
			INSERT INTO availability_settings (
				cleaner_id,
				min_notice_minutes,
				min_booking_minutes,
				max_booking_minutes,
				buffer_minutes,
				booking_horizon_days,
				timezone,
				created_at,
				updated_at
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $8)
			ON CONFLICT (cleaner_id)
			DO UPDATE SET 
				min_notice_minutes = EXCLUDED.min_notice_minutes, 
				min_booking_minutes = EXCLUDED.min_booking_minutes,
				max_booking_minutes = EXCLUDED.max_booking_minutes,
				buffer_minutes = EXCLUDED.buffer_minutes,
				booking_horizon_days = EXCLUDED.booking_horizon_days,
				timezone = EXCLUDED.timezone, 
				updated_at = EXCLUDED.updated_at
			RETURNING created_at, updated_at
		`),
	).
		WithArgs(
			uint(8),
			120,
			60,
			480,
			30,
			90,
			"Europe/London",
			sqlmock.AnyArg(),
		).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"created_at",
				"updated_at",
			}).AddRow(
				now,
				now,
			),
		)

	settings := &AvailabilitySettings{
		CleanerID:          8,
		MinNoticeMinutes:   120,
		MinBookingMinutes:  60,
		MaxBookingMinutes:  480,
		BufferMinutes:      30,
		BookingHorizonDays: 90,
		Timezone:           "Europe/London",
	}

	err := repo.UpsertSettings(
		context.Background(),
		settings,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if settings.CreatedAt.IsZero() {
		t.Fatal("expected created_at")
	}

	if settings.UpdatedAt.IsZero() {
		t.Fatal("expected updated_at")
	}
}

func TestCalendarSQLRepository_UpsertSettings_Error(t *testing.T) {
	repo, mock := newCalendarSQLRepositoryMock(t)

	expectedErr := errors.New("upsert settings failed")

	mock.ExpectQuery("INSERT INTO availability_settings").
		WillReturnError(expectedErr)

	err := repo.UpsertSettings(
		context.Background(),
		&AvailabilitySettings{},
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected %v, got %v",
			expectedErr,
			err,
		)
	}
}

func TestCalendarSQLRepository_GetSettings_Success(t *testing.T) {
	repo, mock := newCalendarSQLRepositoryMock(t)

	now := time.Now()

	mock.ExpectQuery(
		regexp.QuoteMeta(`
			SELECT 
				cleaner_id,
				min_notice_minutes,
				min_booking_minutes,
				max_booking_minutes,
				buffer_minutes,
				booking_horizon_days, 
				timezone,
				created_at,
				updated_at 
			FROM availability_settings
			WHERE cleaner_id = $1 
			LIMIT 1
		`),
	).
		WithArgs(uint(8)).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"cleaner_id",
				"min_notice_minutes",
				"min_booking_minutes",
				"max_booking_minutes",
				"buffer_minutes",
				"booking_horizon_days",
				"timezone",
				"created_at",
				"updated_at",
			}).AddRow(
				8,
				120,
				60,
				480,
				30,
				90,
				"Europe/London",
				now,
				now,
			),
		)

	settings, err := repo.GetSettings(
		context.Background(),
		8,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if settings.CleanerID != 8 {
		t.Fatalf(
			"expected cleaner id 8, got %d",
			settings.CleanerID,
		)
	}
}

func TestCalendarSQLRepository_GetSettings_NotFound(t *testing.T) {
	repo, mock := newCalendarSQLRepositoryMock(t)

	mock.ExpectQuery("FROM availability_settings").
		WithArgs(uint(8)).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"cleaner_id",
				"min_notice_minutes",
				"min_booking_minutes",
				"max_booking_minutes",
				"buffer_minutes",
				"booking_horizon_days",
				"timezone",
				"created_at",
				"updated_at",
			}),
		)

	_, err := repo.GetSettings(
		context.Background(),
		8,
	)

	if !errors.Is(
		err,
		ErrAvailabilitySettingsNotFound,
	) {
		t.Fatalf(
			"expected ErrAvailabilitySettingsNotFound, got %v",
			err,
		)
	}
}

func TestCalendarSQLRepository_GetSettings_DatabaseError(t *testing.T) {
	repo, mock := newCalendarSQLRepositoryMock(t)

	expectedErr := errors.New("settings query failed")

	mock.ExpectQuery("FROM availability_settings").
		WithArgs(uint(8)).
		WillReturnError(expectedErr)

	_, err := repo.GetSettings(
		context.Background(),
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

func TestCalendarSQLRepository_CreateOverride_Success(t *testing.T) {
	repo, mock := newCalendarSQLRepositoryMock(t)

	now := time.Now()

	mock.ExpectQuery(
		regexp.QuoteMeta(`
			INSERT INTO availability_overrides (
				cleaner_id,
				available_date,
				start_time,
				end_time, 
				status, 
				reason, 
				created_at, 
				updated_at
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $7)
			RETURNING id, created_at, updated_at
		`),
	).
		WithArgs(
			uint(8),
			"2026-12-25",
			"09:00",
			"17:00",
			"unavailable",
			"Christmas",
			sqlmock.AnyArg(),
		).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"id",
				"created_at",
				"updated_at",
			}).AddRow(
				22,
				now,
				now,
			),
		)

	override := &AvailabilityOverride{
		CleanerID:        8,
		AvailabilityDate: "2026-12-25",
		StartTime:        "09:00",
		EndTime:          "17:00",
		Status:           "unavailable",
		Reason:           "Christmas",
	}

	err := repo.CreateOverride(
		context.Background(),
		override,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if override.ID != 22 {
		t.Fatalf(
			"expected id 22, got %d",
			override.ID,
		)
	}
}

func TestCalendarSQLRepository_CreateOverride_Error(t *testing.T) {
	repo, mock := newCalendarSQLRepositoryMock(t)

	expectedErr := errors.New("create override failed")

	mock.ExpectQuery("INSERT INTO availability_overrides").
		WillReturnError(expectedErr)

	err := repo.CreateOverride(
		context.Background(),
		&AvailabilityOverride{},
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected %v, got %v",
			expectedErr,
			err,
		)
	}
}

func TestCalendarSQLRepository_ListOverridesByCleanerIDRange_Success(t *testing.T) {
	repo, mock := newCalendarSQLRepositoryMock(t)

	now := time.Now()

	rows := sqlmock.NewRows([]string{
		"id",
		"cleaner_id",
		"available_date",
		"start_time",
		"end_time",
		"status",
		"reason",
		"created_at",
		"updated_at",
	}).AddRow(
		5,
		8,
		"2026-12-25",
		"09:00",
		"17:00",
		"unavailable",
		"Christmas",
		now,
		now,
	)

	mock.ExpectQuery(
		regexp.QuoteMeta(`
			SELECT 
				id,
				cleaner_id,
				available_date::text, 
				start_time,
				end_time,
				status,
				COALESCE(reason, ''), 
				created_at,
				updated_at
			FROM availability_overrides
			WHERE cleaner_id = $1 
				AND available_date >= $2 
				AND available_date <= $3 
			ORDER BY available_date ASC, start_time ASC
		`),
	).
		WithArgs(
			uint(8),
			"2026-12-01",
			"2026-12-31",
		).
		WillReturnRows(rows)

	overrides, err := repo.ListOverridesByCleanerIDRange(
		context.Background(),
		8,
		"2026-12-01",
		"2026-12-31",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(overrides) != 1 {
		t.Fatalf(
			"expected 1 override, got %d",
			len(overrides),
		)
	}

	if overrides[0].Reason != "Christmas" {
		t.Fatalf(
			"unexpected reason %s",
			overrides[0].Reason,
		)
	}
}

func TestCalendarSQLRepository_ListOverridesByCleanerIDRange_QueryError(t *testing.T) {
	repo, mock := newCalendarSQLRepositoryMock(t)

	expectedErr := errors.New("override query failed")

	mock.ExpectQuery("FROM availability_overrides").
		WithArgs(
			uint(8),
			"2026-12-01",
			"2026-12-31",
		).
		WillReturnError(expectedErr)

	_, err := repo.ListOverridesByCleanerIDRange(
		context.Background(),
		8,
		"2026-12-01",
		"2026-12-31",
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected %v, got %v",
			expectedErr,
			err,
		)
	}
}

func TestCalendarSQLRepository_ListOverridesByCleanerIDRange_ScanError(t *testing.T) {
	repo, mock := newCalendarSQLRepositoryMock(t)

	now := time.Now()

	rows := sqlmock.NewRows([]string{
		"id",
		"cleaner_id",
		"available_date",
		"start_time",
		"end_time",
		"status",
		"reason",
		"created_at",
		"updated_at",
	}).AddRow(
		"bad-id",
		8,
		"2026-12-25",
		"09:00",
		"17:00",
		"unavailable",
		"Christmas",
		now,
		now,
	)

	mock.ExpectQuery("FROM availability_overrides").
		WithArgs(
			uint(8),
			"2026-12-01",
			"2026-12-31",
		).
		WillReturnRows(rows)

	overrides, err := repo.ListOverridesByCleanerIDRange(
		context.Background(),
		8,
		"2026-12-01",
		"2026-12-31",
	)

	if overrides != nil {
		t.Fatalf(
			"expected nil overrides, got %+v",
			overrides,
		)
	}

	if err == nil {
		t.Fatal("expected scan error")
	}
}

func TestCalendarSQLRepository_DeleteOverride_Success(t *testing.T) {
	repo, mock := newCalendarSQLRepositoryMock(t)

	mock.ExpectExec(
		regexp.QuoteMeta(`
			DELETE FROM availability_overrides
			WHERE id = $1 
				AND cleaner_id = $2
		`),
	).
		WithArgs(
			uint(5),
			uint(8),
		).
		WillReturnResult(
			sqlmock.NewResult(0, 1),
		)

	err := repo.DeleteOverride(
		context.Background(),
		5,
		8,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCalendarSQLRepository_DeleteOverride_NotFound(t *testing.T) {
	repo, mock := newCalendarSQLRepositoryMock(t)

	mock.ExpectExec("DELETE FROM availability_overrides").
		WithArgs(
			uint(5),
			uint(8),
		).
		WillReturnResult(
			sqlmock.NewResult(0, 0),
		)

	err := repo.DeleteOverride(
		context.Background(),
		5,
		8,
	)

	if !errors.Is(
		err,
		ErrAvailabilityNotFound,
	) {
		t.Fatalf(
			"expected ErrAvailabilityNotFound, got %v",
			err,
		)
	}
}

func TestCalendarSQLRepository_DeleteOverride_ExecError(t *testing.T) {
	repo, mock := newCalendarSQLRepositoryMock(t)

	expectedErr := errors.New("delete override failed")

	mock.ExpectExec("DELETE FROM availability_overrides").
		WithArgs(
			uint(5),
			uint(8),
		).
		WillReturnError(expectedErr)

	err := repo.DeleteOverride(
		context.Background(),
		5,
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
