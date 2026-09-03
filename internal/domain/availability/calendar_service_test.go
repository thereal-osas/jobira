package availability

import (
	"context"
	"errors"
	"testing"
	"time"
)

type calendarMockRepository struct {
	createFn                     func(context.Context, *CleanerAvailability) error
	getByIDFn                    func(context.Context, uint) (*CleanerAvailability, error)
	listByCleanerIDFn            func(context.Context, uint) ([]CleanerAvailability, error)
	listByCleanerIDRangeFn       func(context.Context, uint, string, string) ([]CleanerAvailability, error)
	updateFn                     func(context.Context, *CleanerAvailability) error
	deleteFn                     func(context.Context, uint) error
	hasConflictFn                func(context.Context, uint, string, string, string, uint) (bool, error)
	createBlockFn                func(context.Context, *AvailabilityBlock) error
	listBlocksByCleanerIDFn      func(context.Context, uint) ([]AvailabilityBlock, error)
	listBlocksByCleanerIDRangeFn func(context.Context, uint, time.Time, time.Time) ([]AvailabilityBlock, error)
	deleteBlockFn                func(context.Context, uint, uint) error
	createRecurringFn            func(context.Context, *RecurringAvailability) error
	listRecurringByCleanerIDFn   func(context.Context, uint) ([]RecurringAvailability, error)
	deleteRecurringFn            func(context.Context, uint, uint) error
	upsertSettingsFn             func(context.Context, *AvailabilitySettings) error
	getSettingsFn                func(context.Context, uint) (*AvailabilitySettings, error)
	createOverrideFn             func(context.Context, *AvailabilityOverride) error
	listOverridesRangeFn         func(context.Context, uint, string, string) ([]AvailabilityOverride, error)
	deleteOverrideFn             func(context.Context, uint, uint) error
}

func (m *calendarMockRepository) Create(
	ctx context.Context,
	availability *CleanerAvailability,
) error {
	if m.createFn != nil {
		return m.createFn(ctx, availability)
	}

	return nil
}

func (m *calendarMockRepository) GetByID(
	ctx context.Context,
	id uint,
) (*CleanerAvailability, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}

	return nil, ErrAvailabilityNotFound
}

func (m *calendarMockRepository) ListByCleanerID(
	ctx context.Context,
	cleanerID uint,
) ([]CleanerAvailability, error) {
	if m.listByCleanerIDFn != nil {
		return m.listByCleanerIDFn(ctx, cleanerID)
	}

	return nil, nil
}

func (m *calendarMockRepository) ListByCleanerIDRange(
	ctx context.Context,
	cleanerID uint,
	fromDate string,
	toDate string,
) ([]CleanerAvailability, error) {
	if m.listByCleanerIDRangeFn != nil {
		return m.listByCleanerIDRangeFn(
			ctx,
			cleanerID,
			fromDate,
			toDate,
		)
	}

	return nil, nil
}

func (m *calendarMockRepository) Update(
	ctx context.Context,
	availability *CleanerAvailability,
) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, availability)
	}

	return nil
}

func (m *calendarMockRepository) Delete(
	ctx context.Context,
	id uint,
) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}

	return nil
}

func (m *calendarMockRepository) HasConflict(
	ctx context.Context,
	cleanerID uint,
	availableDate string,
	startTime string,
	endTime string,
	excludeID uint,
) (bool, error) {
	if m.hasConflictFn != nil {
		return m.hasConflictFn(
			ctx,
			cleanerID,
			availableDate,
			startTime,
			endTime,
			excludeID,
		)
	}

	return false, nil
}

func (m *calendarMockRepository) CreateBlock(
	ctx context.Context,
	block *AvailabilityBlock,
) error {
	if m.createBlockFn != nil {
		return m.createBlockFn(ctx, block)
	}

	return nil
}

func (m *calendarMockRepository) ListBlocksByCleanerID(
	ctx context.Context,
	cleanerID uint,
) ([]AvailabilityBlock, error) {
	if m.listBlocksByCleanerIDFn != nil {
		return m.listBlocksByCleanerIDFn(
			ctx,
			cleanerID,
		)
	}

	return nil, nil
}

func (m *calendarMockRepository) ListBlocksByCleanerIDRange(
	ctx context.Context,
	cleanerID uint,
	startAt time.Time,
	endAt time.Time,
) ([]AvailabilityBlock, error) {
	if m.listBlocksByCleanerIDRangeFn != nil {
		return m.listBlocksByCleanerIDRangeFn(
			ctx,
			cleanerID,
			startAt,
			endAt,
		)
	}

	return nil, nil
}

func (m *calendarMockRepository) DeleteBlock(
	ctx context.Context,
	blockID uint,
	cleanerID uint,
) error {
	if m.deleteBlockFn != nil {
		return m.deleteBlockFn(
			ctx,
			blockID,
			cleanerID,
		)
	}

	return nil
}

func (m *calendarMockRepository) CreateRecurring(
	ctx context.Context,
	recurring *RecurringAvailability,
) error {
	if m.createRecurringFn != nil {
		return m.createRecurringFn(
			ctx,
			recurring,
		)
	}

	return nil
}

func (m *calendarMockRepository) ListRecurringByCleanerID(
	ctx context.Context,
	cleanerID uint,
) ([]RecurringAvailability, error) {
	if m.listRecurringByCleanerIDFn != nil {
		return m.listRecurringByCleanerIDFn(
			ctx,
			cleanerID,
		)
	}

	return nil, nil
}

func (m *calendarMockRepository) DeleteRecurring(
	ctx context.Context,
	recurringID uint,
	cleanerID uint,
) error {
	if m.deleteRecurringFn != nil {
		return m.deleteRecurringFn(
			ctx,
			recurringID,
			cleanerID,
		)
	}

	return nil
}

func (m *calendarMockRepository) UpsertSettings(
	ctx context.Context,
	settings *AvailabilitySettings,
) error {
	if m.upsertSettingsFn != nil {
		return m.upsertSettingsFn(
			ctx,
			settings,
		)
	}

	return nil
}

func (m *calendarMockRepository) GetSettings(
	ctx context.Context,
	cleanerID uint,
) (*AvailabilitySettings, error) {
	if m.getSettingsFn != nil {
		return m.getSettingsFn(
			ctx,
			cleanerID,
		)
	}

	return nil, ErrAvailabilitySettingsNotFound
}

func (m *calendarMockRepository) CreateOverride(
	ctx context.Context,
	override *AvailabilityOverride,
) error {
	if m.createOverrideFn != nil {
		return m.createOverrideFn(
			ctx,
			override,
		)
	}

	return nil
}

func (m *calendarMockRepository) ListOverridesByCleanerIDRange(
	ctx context.Context,
	cleanerID uint,
	fromDate string,
	toDate string,
) ([]AvailabilityOverride, error) {
	if m.listOverridesRangeFn != nil {
		return m.listOverridesRangeFn(
			ctx,
			cleanerID,
			fromDate,
			toDate,
		)
	}

	return nil, nil
}

func (m *calendarMockRepository) DeleteOverride(
	ctx context.Context,
	overrideID uint,
	cleanerID uint,
) error {
	if m.deleteOverrideFn != nil {
		return m.deleteOverrideFn(
			ctx,
			overrideID,
			cleanerID,
		)
	}

	return nil
}

func TestCalendarService_CreateRecurring_Success(t *testing.T) {
	repo := &calendarMockRepository{
		createRecurringFn: func(
			_ context.Context,
			recurring *RecurringAvailability,
		) error {
			recurring.ID = 15
			return nil
		},
	}

	service := NewService(repo)

	result, err := service.CreateRecurring(
		context.Background(),
		8,
		CreateRecurringAvailabilityRequest{
			Weekday:   1,
			StartTime: "09:00",
			EndTime:   "17:00",
			Status:    "available",
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ID != 15 {
		t.Fatalf("expected id 15, got %d", result.ID)
	}

	if result.CleanerID != 8 {
		t.Fatalf(
			"expected cleaner id 8, got %d",
			result.CleanerID,
		)
	}

	if result.Weekday != 1 {
		t.Fatalf(
			"expected weekday 1, got %d",
			result.Weekday,
		)
	}

	if result.Status != "available" {
		t.Fatalf(
			"expected available, got %s",
			result.Status,
		)
	}
}

func TestCalendarService_CreateRecurring_DefaultStatus(t *testing.T) {
	repo := &calendarMockRepository{}

	service := NewService(repo)

	result, err := service.CreateRecurring(
		context.Background(),
		8,
		CreateRecurringAvailabilityRequest{
			Weekday:   2,
			StartTime: "10:00",
			EndTime:   "14:00",
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Status != "available" {
		t.Fatalf(
			"expected default status available, got %s",
			result.Status,
		)
	}
}

func TestCalendarService_CreateRecurring_InvalidWeekday(t *testing.T) {
	service := NewService(
		&calendarMockRepository{},
	)

	_, err := service.CreateRecurring(
		context.Background(),
		8,
		CreateRecurringAvailabilityRequest{
			Weekday:   7,
			StartTime: "09:00",
			EndTime:   "17:00",
		},
	)

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf(
			"expected ErrInvalidInput, got %v",
			err,
		)
	}
}

func TestCalendarService_CreateRecurring_InvalidTimeRange(t *testing.T) {
	service := NewService(
		&calendarMockRepository{},
	)

	_, err := service.CreateRecurring(
		context.Background(),
		8,
		CreateRecurringAvailabilityRequest{
			Weekday:   1,
			StartTime: "17:00",
			EndTime:   "09:00",
		},
	)

	if !errors.Is(err, ErrInvalidTimeRange) {
		t.Fatalf(
			"expected ErrInvalidTimeRange, got %v",
			err,
		)
	}
}

func TestCalendarService_CreateRecurringBulk_Success(t *testing.T) {
	created := 0

	repo := &calendarMockRepository{
		createRecurringFn: func(
			_ context.Context,
			recurring *RecurringAvailability,
		) error {
			created++
			recurring.ID = uint(created)
			return nil
		},
	}

	service := NewService(repo)

	results, err := service.CreateRecurringBulk(
		context.Background(),
		8,
		BulkRecurringAvailabilityRequest{
			Slots: []CreateRecurringAvailabilityRequest{
				{
					Weekday:   1,
					StartTime: "09:00",
					EndTime:   "17:00",
				},
				{
					Weekday:   3,
					StartTime: "10:00",
					EndTime:   "16:00",
				},
			},
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 2 {
		t.Fatalf(
			"expected 2 records, got %d",
			len(results),
		)
	}

	if created != 2 {
		t.Fatalf(
			"expected repository create twice, got %d",
			created,
		)
	}
}

func TestCalendarService_ListRecurring_Success(t *testing.T) {
	repo := &calendarMockRepository{
		listRecurringByCleanerIDFn: func(
			_ context.Context,
			cleanerID uint,
		) ([]RecurringAvailability, error) {
			return []RecurringAvailability{
				{
					ID:        1,
					CleanerID: cleanerID,
					Weekday:   1,
					StartTime: "09:00",
					EndTime:   "17:00",
					Status:    "available",
				},
			}, nil
		},
	}

	service := NewService(repo)

	results, err := service.ListRecurring(
		context.Background(),
		8,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf(
			"expected 1 record, got %d",
			len(results),
		)
	}
}

func TestCalendarService_DeleteRecurring_Success(t *testing.T) {
	called := false

	repo := &calendarMockRepository{
		deleteRecurringFn: func(
			_ context.Context,
			recurringID uint,
			cleanerID uint,
		) error {
			called = true

			if recurringID != 4 {
				t.Fatalf(
					"expected recurring id 4, got %d",
					recurringID,
				)
			}

			if cleanerID != 8 {
				t.Fatalf(
					"expected cleaner id 8, got %d",
					cleanerID,
				)
			}

			return nil
		},
	}

	service := NewService(repo)

	err := service.DeleteRecurring(
		context.Background(),
		4,
		8,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !called {
		t.Fatal("expected repository delete to be called")
	}
}

func TestCalendarService_UpdateSettings_Success(t *testing.T) {
	repo := &calendarMockRepository{
		upsertSettingsFn: func(
			_ context.Context,
			settings *AvailabilitySettings,
		) error {
			if settings.CleanerID != 8 {
				t.Fatalf(
					"expected cleaner id 8, got %d",
					settings.CleanerID,
				)
			}

			return nil
		},
	}

	service := NewService(repo)

	settings, err := service.UpdateSettings(
		context.Background(),
		8,
		UpdateAvailabilitySettingsRequest{
			MinNoticeMinutes:   180,
			MinBookingMinutes:  90,
			MaxBookingMinutes:  360,
			BufferMinutes:      45,
			BookingHorizonDays: 60,
			Timezone:           "Europe/London",
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if settings.BufferMinutes != 45 {
		t.Fatalf(
			"expected buffer 45, got %d",
			settings.BufferMinutes,
		)
	}

	if settings.BookingHorizonDays != 60 {
		t.Fatalf(
			"expected horizon 60, got %d",
			settings.BookingHorizonDays,
		)
	}
}

func TestCalendarService_UpdateSettings_InvalidDuration(t *testing.T) {
	service := NewService(
		&calendarMockRepository{},
	)

	_, err := service.UpdateSettings(
		context.Background(),
		8,
		UpdateAvailabilitySettingsRequest{
			MinNoticeMinutes:   120,
			MinBookingMinutes:  180,
			MaxBookingMinutes:  60,
			BufferMinutes:      30,
			BookingHorizonDays: 90,
			Timezone:           "Europe/London",
		},
	)

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf(
			"expected ErrInvalidInput, got %v",
			err,
		)
	}
}

func TestCalendarService_GetSettings_Default(t *testing.T) {
	service := NewService(
		&calendarMockRepository{
			getSettingsFn: func(
				context.Context,
				uint,
			) (*AvailabilitySettings, error) {
				return nil, ErrAvailabilitySettingsNotFound
			},
		},
	)

	settings, err := service.GetSettings(
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

	if settings.BufferMinutes != 30 {
		t.Fatalf(
			"expected default buffer 30, got %d",
			settings.BufferMinutes,
		)
	}

	if settings.Timezone != "Europe/London" {
		t.Fatalf(
			"unexpected timezone %s",
			settings.Timezone,
		)
	}
}

func TestCalendarService_CreateOverride_Success(t *testing.T) {
	repo := &calendarMockRepository{
		createOverrideFn: func(
			_ context.Context,
			override *AvailabilityOverride,
		) error {
			override.ID = 12
			return nil
		},
	}

	service := NewService(repo)

	result, err := service.CreateOverride(
		context.Background(),
		8,
		CreateAvailabilityOverrideRequest{
			AvailableDate: "2026-12-10",
			StartTime:     "12:00",
			EndTime:       "17:00",
			Reason:        "holiday",
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ID != 12 {
		t.Fatalf(
			"expected id 12, got %d",
			result.ID,
		)
	}

	if result.Status != "unavailable" {
		t.Fatalf(
			"expected default unavailable status, got %s",
			result.Status,
		)
	}

	if result.AvailabilityDate != "2026-12-10" {
		t.Fatalf(
			"unexpected date %s",
			result.AvailabilityDate,
		)
	}
}

func TestCalendarService_DeleteOverride_Success(t *testing.T) {
	called := false

	repo := &calendarMockRepository{
		deleteOverrideFn: func(
			_ context.Context,
			overrideID uint,
			cleanerID uint,
		) error {
			called = true

			if overrideID != 5 || cleanerID != 8 {
				t.Fatal("unexpected override delete arguments")
			}

			return nil
		},
	}

	service := NewService(repo)

	err := service.DeleteOverride(
		context.Background(),
		5,
		8,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !called {
		t.Fatal("expected delete override call")
	}
}

func TestCalendarService_CreateBulk_Success(t *testing.T) {
	repo := &calendarMockRepository{
		hasConflictFn: func(
			context.Context,
			uint,
			string,
			string,
			string,
			uint,
		) (bool, error) {
			return false, nil
		},
		createFn: func(
			_ context.Context,
			availability *CleanerAvailability,
		) error {
			return nil
		},
	}

	service := NewService(repo)

	results, err := service.CreateBulk(
		context.Background(),
		8,
		BulkAvailabilityRequest{
			Slots: []CleanerAvailabilityRequest{
				{
					AvailableDate: "2026-12-01",
					StartTime:     "09:00",
					EndTime:       "12:00",
				},
				{
					AvailableDate: "2026-12-02",
					StartTime:     "13:00",
					EndTime:       "17:00",
				},
			},
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 2 {
		t.Fatalf(
			"expected 2 results, got %d",
			len(results),
		)
	}
}

func TestCalendarService_GetCalendar_MergesSources(t *testing.T) {
	from := time.Date(
		2026,
		time.December,
		7,
		0,
		0,
		0,
		0,
		time.UTC,
	)

	to := from.AddDate(0, 0, 1)

	repo := &calendarMockRepository{
		listRecurringByCleanerIDFn: func(
			context.Context,
			uint,
		) ([]RecurringAvailability, error) {
			return []RecurringAvailability{
				{
					Weekday:   int(from.Weekday()),
					StartTime: "09:00",
					EndTime:   "12:00",
					Status:    "available",
				},
			}, nil
		},

		listByCleanerIDRangeFn: func(
			context.Context,
			uint,
			string,
			string,
		) ([]CleanerAvailability, error) {
			return []CleanerAvailability{
				{
					AvailableDate: to.Format("2006-01-02"),
					StartTime:     "10:00",
					EndTime:       "14:00",
					Status:        "available",
					Notes:         "extra hours",
				},
			}, nil
		},

		listOverridesRangeFn: func(
			context.Context,
			uint,
			string,
			string,
		) ([]AvailabilityOverride, error) {
			return []AvailabilityOverride{
				{
					AvailabilityDate: from.Format("2006-01-02"),
					StartTime:        "15:00",
					EndTime:          "16:00",
					Status:           "unavailable",
					Reason:           "appointment",
				},
			}, nil
		},

		listBlocksByCleanerIDRangeFn: func(
			context.Context,
			uint,
			time.Time,
			time.Time,
		) ([]AvailabilityBlock, error) {
			return []AvailabilityBlock{
				{
					StartAt: from.Add(17 * time.Hour),
					EndAt:   from.Add(18 * time.Hour),
					Reason:  "blocked",
				},
			}, nil
		},
	}

	service := NewService(repo)

	slots, err := service.GetCalendar(
		context.Background(),
		8,
		from.Format("2006-01-02"),
		to.Format("2006-01-02"),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(slots) != 4 {
		t.Fatalf(
			"expected 4 slots, got %d",
			len(slots),
		)
	}

	foundSources := map[string]bool{}

	for _, slot := range slots {
		foundSources[slot.Source] = true
	}

	for _, expected := range []string{
		"recurring",
		"specific",
		"override",
		"block",
	} {
		if !foundSources[expected] {
			t.Fatalf(
				"expected source %q",
				expected,
			)
		}
	}
}

func TestCalendarService_NextAvailable_Success(t *testing.T) {
	date := time.Now().
		AddDate(0, 0, 14).
		Format("2006-01-02")

	repo := &calendarMockRepository{
		listRecurringByCleanerIDFn: func(
			context.Context,
			uint,
		) ([]RecurringAvailability, error) {
			parsed, _ := time.Parse(
				"2006-01-02",
				date,
			)

			return []RecurringAvailability{
				{
					Weekday:   int(parsed.Weekday()),
					StartTime: "09:00",
					EndTime:   "17:00",
					Status:    "available",
				},
			}, nil
		},
	}

	service := NewService(repo)

	slot, err := service.NextAvailable(
		context.Background(),
		8,
		date,
		7,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if slot == nil {
		t.Fatal("expected next available slot")
	}

	if slot.Status != "available" {
		t.Fatalf(
			"expected available slot, got %s",
			slot.Status,
		)
	}
}

func TestCalendarService_NextAvailable_NotFound(t *testing.T) {
	service := NewService(
		&calendarMockRepository{},
	)

	date := time.Now().
		AddDate(0, 0, 14).
		Format("2006-01-02")

	_, err := service.NextAvailable(
		context.Background(),
		8,
		date,
		7,
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

func TestCalendarService_BookingBufferMinutes_Default(t *testing.T) {
	service := NewService(
		&calendarMockRepository{},
	)

	buffer, err := service.BookingBufferMinutes(
		context.Background(),
		8,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if buffer != 30 {
		t.Fatalf(
			"expected default buffer 30, got %d",
			buffer,
		)
	}
}

func TestCalendarService_ValidateBookingSlot_RecurringAllowed(t *testing.T) {
	location, err := time.LoadLocation(
		"Europe/London",
	)
	if err != nil {
		t.Fatal(err)
	}

	startAt := time.Now().
		In(location).
		AddDate(0, 0, 7)

	startAt = time.Date(
		startAt.Year(),
		startAt.Month(),
		startAt.Day(),
		10,
		0,
		0,
		0,
		location,
	)

	endAt := startAt.Add(2 * time.Hour)

	repo := &calendarMockRepository{
		getSettingsFn: func(
			context.Context,
			uint,
		) (*AvailabilitySettings, error) {
			return &AvailabilitySettings{
				CleanerID:          8,
				MinNoticeMinutes:   0,
				MinBookingMinutes:  60,
				MaxBookingMinutes:  480,
				BufferMinutes:      30,
				BookingHorizonDays: 90,
				Timezone:           "Europe/London",
			}, nil
		},

		listRecurringByCleanerIDFn: func(
			context.Context,
			uint,
		) ([]RecurringAvailability, error) {
			return []RecurringAvailability{
				{
					CleanerID: 8,
					Weekday:   int(startAt.Weekday()),
					StartTime: "09:00",
					EndTime:   "17:00",
					Status:    "available",
				},
			}, nil
		},
	}

	service := NewService(repo)

	err = service.ValidateBookingSlot(
		context.Background(),
		8,
		startAt,
		endAt,
	)
	if err != nil {
		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}
}

func TestCalendarService_ValidateBookingSlot_BlockWins(t *testing.T) {
	location, err := time.LoadLocation(
		"Europe/London",
	)
	if err != nil {
		t.Fatal(err)
	}

	startAt := time.Now().
		In(location).
		AddDate(0, 0, 7)

	startAt = time.Date(
		startAt.Year(),
		startAt.Month(),
		startAt.Day(),
		10,
		0,
		0,
		0,
		location,
	)

	endAt := startAt.Add(2 * time.Hour)

	repo := &calendarMockRepository{
		getSettingsFn: func(
			context.Context,
			uint,
		) (*AvailabilitySettings, error) {
			return &AvailabilitySettings{
				CleanerID:          8,
				MinNoticeMinutes:   0,
				MinBookingMinutes:  60,
				MaxBookingMinutes:  480,
				BufferMinutes:      30,
				BookingHorizonDays: 90,
				Timezone:           "Europe/London",
			}, nil
		},

		listBlocksByCleanerIDRangeFn: func(
			context.Context,
			uint,
			time.Time,
			time.Time,
		) ([]AvailabilityBlock, error) {
			return []AvailabilityBlock{
				{
					CleanerID: 8,
					StartAt:   startAt.Add(30 * time.Minute),
					EndAt:     startAt.Add(90 * time.Minute),
				},
			}, nil
		},

		listRecurringByCleanerIDFn: func(
			context.Context,
			uint,
		) ([]RecurringAvailability, error) {
			return []RecurringAvailability{
				{
					CleanerID: 8,
					Weekday:   int(startAt.Weekday()),
					StartTime: "09:00",
					EndTime:   "17:00",
					Status:    "available",
				},
			}, nil
		},
	}

	service := NewService(repo)

	err = service.ValidateBookingSlot(
		context.Background(),
		8,
		startAt,
		endAt,
	)

	if !errors.Is(
		err,
		ErrAvailabilityConflict,
	) {
		t.Fatalf(
			"expected ErrAvailabilityConflict, got %v",
			err,
		)
	}
}

func TestCalendarService_ValidateBookingSlot_OverrideUnavailableWins(t *testing.T) {
	location, err := time.LoadLocation(
		"Europe/London",
	)
	if err != nil {
		t.Fatal(err)
	}

	startAt := time.Now().
		In(location).
		AddDate(0, 0, 7)

	startAt = time.Date(
		startAt.Year(),
		startAt.Month(),
		startAt.Day(),
		10,
		0,
		0,
		0,
		location,
	)

	endAt := startAt.Add(2 * time.Hour)

	date := startAt.Format("2006-01-02")

	repo := &calendarMockRepository{
		getSettingsFn: func(
			context.Context,
			uint,
		) (*AvailabilitySettings, error) {
			return &AvailabilitySettings{
				CleanerID:          8,
				MinNoticeMinutes:   0,
				MinBookingMinutes:  60,
				MaxBookingMinutes:  480,
				BufferMinutes:      30,
				BookingHorizonDays: 90,
				Timezone:           "Europe/London",
			}, nil
		},

		listOverridesRangeFn: func(
			context.Context,
			uint,
			string,
			string,
		) ([]AvailabilityOverride, error) {
			return []AvailabilityOverride{
				{
					CleanerID:        8,
					AvailabilityDate: date,
					StartTime:        "09:00",
					EndTime:          "17:00",
					Status:           "unavailable",
				},
			}, nil
		},

		listRecurringByCleanerIDFn: func(
			context.Context,
			uint,
		) ([]RecurringAvailability, error) {
			return []RecurringAvailability{
				{
					CleanerID: 8,
					Weekday:   int(startAt.Weekday()),
					StartTime: "09:00",
					EndTime:   "17:00",
					Status:    "available",
				},
			}, nil
		},
	}

	service := NewService(repo)

	err = service.ValidateBookingSlot(
		context.Background(),
		8,
		startAt,
		endAt,
	)

	if !errors.Is(
		err,
		ErrAvailabilityConflict,
	) {
		t.Fatalf(
			"expected ErrAvailabilityConflict, got %v",
			err,
		)
	}
}

func TestCalendarService_ValidateBookingSlot_SpecificAvailabilityWins(t *testing.T) {
	location, err := time.LoadLocation(
		"Europe/London",
	)
	if err != nil {
		t.Fatal(err)
	}

	startAt := time.Now().
		In(location).
		AddDate(0, 0, 7)

	startAt = time.Date(
		startAt.Year(),
		startAt.Month(),
		startAt.Day(),
		11,
		0,
		0,
		0,
		location,
	)

	endAt := startAt.Add(2 * time.Hour)

	date := startAt.Format("2006-01-02")

	repo := &calendarMockRepository{
		getSettingsFn: func(
			context.Context,
			uint,
		) (*AvailabilitySettings, error) {
			return &AvailabilitySettings{
				CleanerID:          8,
				MinNoticeMinutes:   0,
				MinBookingMinutes:  60,
				MaxBookingMinutes:  480,
				BufferMinutes:      30,
				BookingHorizonDays: 90,
				Timezone:           "Europe/London",
			}, nil
		},

		listByCleanerIDRangeFn: func(
			context.Context,
			uint,
			string,
			string,
		) ([]CleanerAvailability, error) {
			return []CleanerAvailability{
				{
					CleanerID:     8,
					AvailableDate: date,
					StartTime:     "10:00",
					EndTime:       "15:00",
					Status:        "available",
				},
			}, nil
		},
	}

	service := NewService(repo)

	err = service.ValidateBookingSlot(
		context.Background(),
		8,
		startAt,
		endAt,
	)
	if err != nil {
		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}
}

func TestCalendarService_ValidateBookingSlot_OutsideAvailability(t *testing.T) {
	location, err := time.LoadLocation(
		"Europe/London",
	)
	if err != nil {
		t.Fatal(err)
	}

	startAt := time.Now().
		In(location).
		AddDate(0, 0, 7)

	startAt = time.Date(
		startAt.Year(),
		startAt.Month(),
		startAt.Day(),
		18,
		0,
		0,
		0,
		location,
	)

	endAt := startAt.Add(2 * time.Hour)

	repo := &calendarMockRepository{
		getSettingsFn: func(
			context.Context,
			uint,
		) (*AvailabilitySettings, error) {
			return &AvailabilitySettings{
				CleanerID:          8,
				MinNoticeMinutes:   0,
				MinBookingMinutes:  60,
				MaxBookingMinutes:  480,
				BufferMinutes:      30,
				BookingHorizonDays: 90,
				Timezone:           "Europe/London",
			}, nil
		},

		listRecurringByCleanerIDFn: func(
			context.Context,
			uint,
		) ([]RecurringAvailability, error) {
			return []RecurringAvailability{
				{
					CleanerID: 8,
					Weekday:   int(startAt.Weekday()),
					StartTime: "09:00",
					EndTime:   "17:00",
					Status:    "available",
				},
			}, nil
		},
	}

	service := NewService(repo)

	err = service.ValidateBookingSlot(
		context.Background(),
		8,
		startAt,
		endAt,
	)

	if !errors.Is(
		err,
		ErrOutsideAvailability,
	) {
		t.Fatalf(
			"expected ErrOutsideAvailability, got %v",
			err,
		)
	}
}

func TestCalendarService_ValidateBookingSlot_MinimumDuration(t *testing.T) {
	location, err := time.LoadLocation(
		"Europe/London",
	)
	if err != nil {
		t.Fatal(err)
	}

	startAt := time.Now().
		In(location).
		AddDate(0, 0, 7)

	startAt = time.Date(
		startAt.Year(),
		startAt.Month(),
		startAt.Day(),
		10,
		0,
		0,
		0,
		location,
	)

	repo := &calendarMockRepository{
		getSettingsFn: func(
			context.Context,
			uint,
		) (*AvailabilitySettings, error) {
			return &AvailabilitySettings{
				CleanerID:          8,
				MinNoticeMinutes:   0,
				MinBookingMinutes:  60,
				MaxBookingMinutes:  480,
				BookingHorizonDays: 90,
				Timezone:           "Europe/London",
			}, nil
		},
	}

	service := NewService(repo)

	err = service.ValidateBookingSlot(
		context.Background(),
		8,
		startAt,
		startAt.Add(30*time.Minute),
	)

	if !errors.Is(
		err,
		ErrMinimumDuration,
	) {
		t.Fatalf(
			"expected ErrMinimumDuration, got %v",
			err,
		)
	}
}

func TestCalendarService_ValidateBookingSlot_MaximumDuration(t *testing.T) {
	location, err := time.LoadLocation(
		"Europe/London",
	)
	if err != nil {
		t.Fatal(err)
	}

	startAt := time.Now().
		In(location).
		AddDate(0, 0, 7)

	startAt = time.Date(
		startAt.Year(),
		startAt.Month(),
		startAt.Day(),
		9,
		0,
		0,
		0,
		location,
	)

	repo := &calendarMockRepository{
		getSettingsFn: func(
			context.Context,
			uint,
		) (*AvailabilitySettings, error) {
			return &AvailabilitySettings{
				CleanerID:          8,
				MinNoticeMinutes:   0,
				MinBookingMinutes:  60,
				MaxBookingMinutes:  120,
				BookingHorizonDays: 90,
				Timezone:           "Europe/London",
			}, nil
		},
	}

	service := NewService(repo)

	err = service.ValidateBookingSlot(
		context.Background(),
		8,
		startAt,
		startAt.Add(3*time.Hour),
	)

	if !errors.Is(
		err,
		ErrMaximumDuration,
	) {
		t.Fatalf(
			"expected ErrMaximumDuration, got %v",
			err,
		)
	}
}
