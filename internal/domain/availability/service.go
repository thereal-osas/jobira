package availability

import (
	"context"
	"sort"
	"strings"
	"time"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, cleanerID uint, req CleanerAvailabilityRequest) (*CleanerAvailability, error) {
	req.AvailableDate = strings.TrimSpace(req.AvailableDate)
	req.StartTime = strings.TrimSpace(req.StartTime)
	req.EndTime = strings.TrimSpace(req.EndTime)
	req.Status = strings.TrimSpace(req.Status)
	req.Notes = strings.TrimSpace(req.Notes)

	if cleanerID == 0 || req.AvailableDate == "" || req.StartTime == "" || req.EndTime == "" {
		return nil, ErrInvalidInput
	}

	if _, err := time.Parse("2006-01-02", req.AvailableDate); err != nil {
		return nil, ErrInvalidInput
	}

	if _, err := validateClockRange(
		req.StartTime,
		req.EndTime,
	); err != nil {
		return nil, err
	}

	if req.Status == "" {
		req.Status = "available"
	}

	if !isAllowedStatus(req.Status) {
		return nil, ErrInvalidInput
	}

	conflict, err := s.repo.HasConflict(
		ctx,
		cleanerID,
		req.AvailableDate,
		req.StartTime,
		req.EndTime,
		0,
	)
	if err != nil {
		return nil, err
	}

	if conflict {
		return nil, ErrAvailabilityConflict
	}

	availability := &CleanerAvailability{
		CleanerID:     cleanerID,
		AvailableDate: req.AvailableDate,
		StartTime:     req.StartTime,
		EndTime:       req.EndTime,
		Status:        req.Status,
		Notes:         req.Notes,
	}

	if err := s.repo.Create(ctx, availability); err != nil {
		return nil, err
	}

	return availability, nil
}

func (s *Service) GetByID(ctx context.Context, id uint, userID uint, role string) (*CleanerAvailability, error) {
	if id == 0 || userID == 0 {
		return nil, ErrInvalidInput
	}

	availability, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if role != "admin" && availability.CleanerID != userID {
		return nil, ErrForbidden
	}

	return availability, nil
}

func (s *Service) ListByCleanerID(ctx context.Context, cleanerID uint) ([]CleanerAvailability, error) {
	if cleanerID == 0 {
		return nil, ErrInvalidInput
	}

	return s.repo.ListByCleanerID(ctx, cleanerID)
}

func (s *Service) Update(ctx context.Context, id uint, userID uint, role string, req UpdateAvailabilityRequest) (*CleanerAvailability, error) {
	if id == 0 || userID == 0 {
		return nil, ErrInvalidInput
	}

	foundAvailability, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if role != "admin" && foundAvailability.CleanerID != userID {
		return nil, ErrForbidden
	}

	req.AvailableDate = strings.TrimSpace(req.AvailableDate)
	req.StartTime = strings.TrimSpace(req.StartTime)
	req.EndTime = strings.TrimSpace(req.EndTime)
	req.Status = strings.TrimSpace(req.Status)
	req.Notes = strings.TrimSpace(req.Notes)

	if req.AvailableDate == "" || req.StartTime == "" || req.EndTime == "" {
		return nil, ErrInvalidInput
	}

	if _, err := time.Parse("2006-01-02", req.AvailableDate); err != nil {
		return nil, ErrInvalidInput
	}

	if _, err := validateClockRange(
		req.StartTime,
		req.EndTime,
	); err != nil {
		return nil, err
	}

	if req.Status == "" {
		req.Status = "available"
	}

	if !isAllowedStatus(req.Status) {
		return nil, ErrInvalidStatus
	}

	conflict, err := s.repo.HasConflict(
		ctx,
		foundAvailability.CleanerID,
		req.AvailableDate,
		req.StartTime,
		req.EndTime,
		foundAvailability.ID,
	)
	if err != nil {
		return nil, err
	}

	if conflict {
		return nil, ErrAvailabilityConflict
	}

	foundAvailability.AvailableDate = req.AvailableDate
	foundAvailability.StartTime = req.StartTime
	foundAvailability.EndTime = req.EndTime
	foundAvailability.Status = req.Status
	foundAvailability.Notes = req.Notes

	if err := s.repo.Update(ctx, foundAvailability); err != nil {
		return nil, err
	}

	return s.repo.GetByID(ctx, id)
}

func (s *Service) Delete(ctx context.Context, id uint, userID uint, role string) error {
	if id == 0 || userID == 0 {
		return ErrInvalidInput
	}

	foundAvailability, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if role != "admin" && foundAvailability.CleanerID != userID {
		return ErrForbidden
	}

	return s.repo.Delete(ctx, id)
}

func isAllowedStatus(status string) bool {
	if status == "available" {
		return true
	}

	if status == "busy" {
		return true
	}

	if status == "unavailable" {
		return true
	}

	return false
}

func (s *Service) ListMine(ctx context.Context, cleanerID uint) ([]CleanerAvailability, error) {
	if cleanerID == 0 {
		return nil, ErrInvalidInput
	}

	return s.repo.ListByCleanerID(ctx, cleanerID)
}

func (s *Service) CreateBlock(ctx context.Context, cleanerID uint, req CreateAvailabilityBlockRequest) (*AvailabilityBlock, error) {
	req.StartAt = strings.TrimSpace(req.StartAt)
	req.EndAt = strings.TrimSpace(req.EndAt)
	req.Reason = strings.TrimSpace(req.Reason)

	if cleanerID == 0 || req.Reason == "" {
		return nil, ErrInvalidInput
	}

	if req.StartAt == "" || req.EndAt == "" {
		return nil, ErrInvalidInput
	}

	startAt, err := time.Parse(time.RFC3339, req.StartAt)
	if err != nil {
		return nil, ErrInvalidInput
	}

	if startAt.Before(time.Now()) {
		return nil, ErrInvalidInput
	}

	endAt, err := time.Parse(time.RFC3339, req.EndAt)
	if err != nil {
		return nil, ErrInvalidInput
	}

	if !endAt.After(startAt) {
		return nil, ErrInvalidInput
	}

	block := &AvailabilityBlock{
		CleanerID: cleanerID,
		StartAt:   startAt,
		EndAt:     endAt,
		Reason:    req.Reason,
	}

	if err := s.repo.CreateBlock(ctx, block); err != nil {
		return nil, err
	}

	return block, nil
}

func (s *Service) ListBlocks(ctx context.Context, cleanerID uint) ([]AvailabilityBlock, error) {
	if cleanerID == 0 {
		return nil, ErrInvalidInput
	}

	return s.repo.ListBlocksByCleanerID(ctx, cleanerID)
}

func (s *Service) DeleteBlock(ctx context.Context, blockID uint, cleanerID uint) error {
	if blockID == 0 || cleanerID == 0 {
		return ErrInvalidInput
	}

	return s.repo.DeleteBlock(ctx, blockID, cleanerID)
}

func defaultAvailabilitySettings(cleanerID uint) *AvailabilitySettings {
	return &AvailabilitySettings{
		CleanerID:          cleanerID,
		MinNoticeMinutes:   120,
		MinBookingMinutes:  60,
		MaxBookingMinutes:  480,
		BufferMinutes:      30,
		BookingHorizonDays: 90,
		Timezone:           "Europe/London",
	}
}

func (s *Service) getSettingsOrDefault(ctx context.Context, cleanerID uint) (*AvailabilitySettings, error) {
	settings, err := s.repo.GetSettings(
		ctx,
		cleanerID,
	)

	if err == nil {
		return settings, nil
	}

	if err == ErrAvailabilitySettingsNotFound {
		return defaultAvailabilitySettings(cleanerID), nil
	}

	return nil, err
}

func (s *Service) UpdateSettings(ctx context.Context, cleanerID uint, req UpdateAvailabilitySettingsRequest) (*AvailabilitySettings, error) {
	if cleanerID == 0 {
		return nil, ErrInvalidInput
	}

	req.Timezone = strings.TrimSpace(req.Timezone)

	if req.Timezone == "" {
		req.Timezone = "Europe/London"
	}

	if _, err := time.LoadLocation(req.Timezone); err != nil {
		return nil, ErrInvalidInput
	}

	if req.MinNoticeMinutes < 0 ||
		req.MinBookingMinutes <= 0 ||
		req.MaxBookingMinutes <= 0 ||
		req.BufferMinutes < 0 ||
		req.BookingHorizonDays <= 0 {
		return nil, ErrInvalidInput
	}

	if req.MaxBookingMinutes < req.MinBookingMinutes {
		return nil, ErrInvalidInput
	}

	settings := &AvailabilitySettings{
		CleanerID:          cleanerID,
		MinNoticeMinutes:   req.MinNoticeMinutes,
		MinBookingMinutes:  req.MinBookingMinutes,
		MaxBookingMinutes:  req.MaxBookingMinutes,
		BufferMinutes:      req.BufferMinutes,
		BookingHorizonDays: req.BookingHorizonDays,
		Timezone:           req.Timezone,
	}

	if err := s.repo.UpsertSettings(
		ctx, settings,
	); err != nil {
		return nil, err
	}

	return settings, nil
}

func (s *Service) GetSettings(ctx context.Context, cleanerID uint) (*AvailabilitySettings, error) {
	if cleanerID == 0 {
		return nil, ErrInvalidInput
	}

	return s.getSettingsOrDefault(
		ctx,
		cleanerID,
	)
}

func validateClockRange(start string, end string) (time.Duration, error) {
	start = strings.TrimSpace(start)
	end = strings.TrimSpace(end)

	startTime, err := time.Parse(
		"15:04",
		start,
	)
	if err != nil {
		return 0, ErrInvalidTimeRange
	}

	endTime, err := time.Parse(
		"15:04",
		end,
	)

	if !endTime.After(startTime) {
		return 0, ErrInvalidTimeRange
	}

	return endTime.Sub(startTime), nil
}

func parseAvailabilityDateTime(date string, clock string, location *time.Location) (time.Time, error) {
	return time.ParseInLocation(
		"2006-01-02 15:04",
		date+" "+clock,
		location,
	)
}

func (s *Service) validateBookingRules(ctx context.Context, cleanerID uint, date string, start string, end string) error {
	duration, err := validateClockRange(
		start,
		end,
	)
	if err != nil {
		return err
	}

	settings, err := s.getSettingsOrDefault(
		ctx,
		cleanerID,
	)
	if err != nil {
		return err
	}

	location, err := time.LoadLocation(
		settings.Timezone,
	)
	if err != nil {
		return ErrInvalidInput
	}

	startAt, err := parseAvailabilityDateTime(
		date,
		start,
		location,
	)
	if err != nil {
		return ErrInvalidInput
	}

	now := time.Now().In(location)

	minNotice := time.Duration(
		settings.MinNoticeMinutes,
	) * time.Minute

	if startAt.Before(now.Add(minNotice)) {
		return ErrMinimumNotice
	}

	if settings.BookingHorizonDays > 0 {
		horizon := now.AddDate(
			0,
			0,
			settings.BookingHorizonDays,
		)

		if startAt.After(horizon) {
			return ErrBookingHorizonExceeded
		}
	}

	minDuration := time.Duration(
		settings.MinBookingMinutes,
	) * time.Minute

	if duration < minDuration {
		return ErrMinimumDuration
	}

	maxDuration := time.Duration(
		settings.MaxBookingMinutes,
	) * time.Minute

	if duration > maxDuration {
		return ErrMaximumDuration
	}

	return nil
}

func (s *Service) CreateRecurring(ctx context.Context, cleanerID uint, req CreateRecurringAvailabilityRequest) (*RecurringAvailability, error) {
	req.StartTime = strings.TrimSpace(req.StartTime)
	req.EndTime = strings.TrimSpace(req.EndTime)
	req.Status = strings.TrimSpace(req.Status)

	if cleanerID == 0 {
		return nil, ErrInvalidInput
	}

	if req.Weekday < 0 || req.Weekday > 6 {
		return nil, ErrInvalidInput
	}

	if _, err := validateClockRange(
		req.StartTime,
		req.EndTime,
	); err != nil {
		return nil, err
	}

	if req.Status == "" {
		req.Status = "available"
	}

	if !isAllowedStatus(req.Status) {
		return nil, ErrInvalidStatus
	}

	recurring := &RecurringAvailability{
		CleanerID: cleanerID,
		Weekday:   req.Weekday,
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
		Status:    req.Status,
	}

	if err := s.repo.CreateRecurring(
		ctx,
		recurring,
	); err != nil {
		return nil, err
	}

	return recurring, nil
}

func (s *Service) CreateRecurringBulk(ctx context.Context, cleanerID uint, req BulkRecurringAvailabilityRequest) ([]RecurringAvailability, error) {
	if cleanerID == 0 || len(req.Slots) == 0 {
		return nil, ErrInvalidInput
	}

	results := make(
		[]RecurringAvailability,
		0,
		len(req.Slots),
	)

	for _, slot := range req.Slots {
		created, err := s.CreateRecurring(
			ctx,
			cleanerID,
			slot,
		)
		if err != nil {
			return nil, err
		}

		results = append(results, *created)
	}

	return results, nil
}

func (s *Service) ListRecurring(ctx context.Context, cleanerID uint) ([]RecurringAvailability, error) {
	if cleanerID == 0 {
		return nil, ErrInvalidInput
	}

	return s.repo.ListRecurringByCleanerID(
		ctx,
		cleanerID,
	)
}

func (s *Service) DeleteRecurring(ctx context.Context, recurringID uint, cleanerID uint) error {
	if recurringID == 0 || cleanerID == 0 {
		return ErrInvalidInput
	}

	return s.repo.DeleteRecurring(
		ctx,
		recurringID,
		cleanerID,
	)
}

func (s *Service) CreateOverride(ctx context.Context, cleanerID uint, req CreateAvailabilityOverrideRequest) (*AvailabilityOverride, error) {
	req.AvailableDate = strings.TrimSpace(req.AvailableDate)
	req.StartTime = strings.TrimSpace(req.StartTime)
	req.EndTime = strings.TrimSpace(req.EndTime)
	req.Status = strings.TrimSpace(req.Status)
	req.Reason = strings.TrimSpace(req.Reason)

	if cleanerID == 0 ||
		req.AvailableDate == "" ||
		req.StartTime == "" ||
		req.EndTime == "" {
		return nil, ErrInvalidInput
	}

	if _, err := time.Parse(
		"2006-01-02",
		req.AvailableDate,
	); err != nil {
		return nil, ErrInvalidInput
	}

	if _, err := validateClockRange(
		req.StartTime,
		req.EndTime,
	); err != nil {
		return nil, err
	}
	if req.Status == "" {
		req.Status = "unavailable"
	}

	if !isAllowedStatus(req.Status) {
		return nil, ErrInvalidStatus
	}

	override := &AvailabilityOverride{
		CleanerID:        cleanerID,
		AvailabilityDate: req.AvailableDate,
		StartTime:        req.StartTime,
		EndTime:          req.EndTime,
		Status:           req.Status,
		Reason:           req.Reason,
	}

	if err := s.repo.CreateOverride(
		ctx,
		override,
	); err != nil {
		return nil, err
	}

	return override, nil
}

func (s *Service) DeleteOverride(ctx context.Context, overrideID uint, cleanerID uint) error {
	if overrideID == 0 || cleanerID == 0 {
		return ErrInvalidInput
	}

	return s.repo.DeleteOverride(
		ctx,
		overrideID,
		cleanerID,
	)
}

func (s *Service) CreateBulk(ctx context.Context, cleanerID uint, req BulkAvailabilityRequest) ([]CleanerAvailability, error) {
	if cleanerID == 0 || len(req.Slots) == 0 {
		return nil, ErrInvalidInput
	}

	results := make(
		[]CleanerAvailability,
		0,
		len(req.Slots),
	)

	for _, slot := range req.Slots {
		created, err := s.Create(
			ctx,
			cleanerID,
			slot,
		)
		if err != nil {
			return nil, err
		}

		results = append(results, *created)
	}

	return results, nil
}

func (s *Service) GetCalendar(ctx context.Context, cleanerID uint, fromDate string, toDate string) ([]CalendarSlot, error) {
	if cleanerID == 0 {
		return nil, ErrInvalidInput
	}

	from, err := time.Parse(
		"2006-01-02",
		fromDate,
	)
	if err != nil {
		return nil, ErrInvalidInput
	}

	to, err := time.Parse(
		"2006-01-02",
		toDate,
	)
	if err != nil || to.Before(from) {
		return nil, ErrInvalidInput
	}

	if to.Sub(from) > 366*24*time.Hour {
		return nil, ErrInvalidInput
	}

	recurring, err := s.repo.ListRecurringByCleanerID(
		ctx,
		cleanerID,
	)
	if err != nil {
		return nil, err
	}

	specific, err := s.repo.ListByCleanerIDRange(
		ctx,
		cleanerID,
		fromDate,
		toDate,
	)
	if err != nil {
		return nil, err
	}

	overrides, err := s.repo.ListOverridesByCleanerIDRange(
		ctx,
		cleanerID,
		fromDate,
		toDate,
	)
	if err != nil {
		return nil, err
	}

	rangeEnd := to.Add(24 * time.Hour)

	blocks, err := s.repo.ListBlocksByCleanerIDRange(
		ctx,
		cleanerID,
		from,
		rangeEnd,
	)
	if err != nil {
		return nil, err
	}

	var slots []CalendarSlot

	for date := from; !date.After(to); date = date.AddDate(0, 0, 1) {
		dateString := date.Format("2006-01-02")

		for _, recurringSlot := range recurring {
			if recurringSlot.Weekday != int(date.Weekday()) {
				continue
			}

			slots = append(
				slots,
				CalendarSlot{
					Date:      dateString,
					StartTime: recurringSlot.StartTime,
					EndTime:   recurringSlot.EndTime,
					Status:    recurringSlot.Status,
					Source:    "recurring",
				},
			)
		}
	}

	for _, record := range specific {
		slots = append(
			slots,
			CalendarSlot{
				Date:      record.AvailableDate,
				StartTime: record.StartTime,
				EndTime:   record.EndTime,
				Status:    record.Status,
				Source:    "specific",
				Reason:    record.Notes,
			},
		)
	}

	for _, override := range overrides {
		slots = append(
			slots,
			CalendarSlot{
				Date:      override.AvailabilityDate,
				StartTime: override.StartTime,
				EndTime:   override.EndTime,
				Status:    override.Status,
				Source:    "override",
				Reason:    override.Reason,
			},
		)
	}

	for _, block := range blocks {
		slots = append(
			slots,
			CalendarSlot{
				Date: block.StartAt.Format(
					"2006-01-02",
				),
				StartTime: block.StartAt.Format(
					"15:04",
				),
				EndTime: block.EndAt.Format(
					"15:04",
				),
				Status: "unavailable",
				Source: "block",
				Reason: block.Reason,
			},
		)
	}

	sort.Slice(
		slots,
		func(i, j int) bool {
			if slots[i].Date == slots[j].Date {
				return slots[i].StartTime <
					slots[j].StartTime
			}

			return slots[i].Date <
				slots[j].Date
		},
	)

	return slots, nil
}

func (s *Service) NextAvailable(ctx context.Context, cleanerID uint, fromDate string, days int) (*CalendarSlot, error) {
	if days <= 0 {
		days = 30
	}

	if days > 90 {
		days = 90
	}

	from, err := time.Parse(
		"2006-01-02",
		fromDate,
	)
	if err != nil {
		return nil, ErrInvalidInput
	}

	to := from.AddDate(
		0,
		0,
		days,
	)

	slots, err := s.GetCalendar(
		ctx,
		cleanerID,
		fromDate,
		to.Format("2006-01-02"),
	)
	if err != nil {
		return nil, err
	}

	for _, slot := range slots {
		if slot.Status == "available" {
			result := slot
			return &result, nil
		}
	}

	return nil, ErrAvailabilityNotFound
}

func (s *Service) ValidateBookingSlot(
	ctx context.Context,
	cleanerID uint,
	startAt time.Time,
	endAt time.Time,
) error {
	if cleanerID == 0 {
		return ErrInvalidInput
	}

	if startAt.IsZero() || endAt.IsZero() {
		return ErrInvalidInput
	}

	if !endAt.After(startAt) {
		return ErrInvalidTimeRange
	}

	settings, err := s.getSettingsOrDefault(
		ctx,
		cleanerID,
	)
	if err != nil {
		return err
	}

	location, err := time.LoadLocation(
		settings.Timezone,
	)
	if err != nil {
		return ErrInvalidInput
	}

	localStart := startAt.In(location)
	localEnd := endAt.In(location)
	now := time.Now().In(location)

	/*
		For the current marketplace calendar model,
		a booking must begin and end on the same
		local calendar day.
	*/
	if localStart.Format("2006-01-02") !=
		localEnd.Format("2006-01-02") {
		return ErrInvalidTimeRange
	}

	if localStart.Before(now) {
		return ErrInvalidInput
	}

	/*
		MINIMUM NOTICE
	*/
	minimumNotice := time.Duration(
		settings.MinNoticeMinutes,
	) * time.Minute

	if localStart.Before(
		now.Add(minimumNotice),
	) {
		return ErrMinimumNotice
	}

	/*
		BOOKING HORIZON
	*/
	if settings.BookingHorizonDays > 0 {
		horizon := now.AddDate(
			0,
			0,
			settings.BookingHorizonDays,
		)

		if localStart.After(horizon) {
			return ErrBookingHorizonExceeded
		}
	}

	/*
		BOOKING DURATION
	*/
	duration := localEnd.Sub(localStart)

	minimumDuration := time.Duration(
		settings.MinBookingMinutes,
	) * time.Minute

	if duration < minimumDuration {
		return ErrMinimumDuration
	}

	maximumDuration := time.Duration(
		settings.MaxBookingMinutes,
	) * time.Minute

	if duration > maximumDuration {
		return ErrMaximumDuration
	}

	date := localStart.Format("2006-01-02")

	/*
		1. HARD BLOCKS

		A block always wins over availability.
	*/
	blocks, err := s.repo.ListBlocksByCleanerIDRange(
		ctx,
		cleanerID,
		localStart,
		localEnd,
	)
	if err != nil {
		return err
	}

	for _, block := range blocks {
		blockStart := block.StartAt.In(location)
		blockEnd := block.EndAt.In(location)

		if timeRangesOverlap(
			localStart,
			localEnd,
			blockStart,
			blockEnd,
		) {
			return ErrAvailabilityConflict
		}
	}

	/*
		2. ONE-OFF OVERRIDES

		Overrides take priority over the cleaner's
		normal recurring schedule.
	*/
	overrides, err := s.repo.ListOverridesByCleanerIDRange(
		ctx,
		cleanerID,
		date,
		date,
	)
	if err != nil {
		return err
	}

	hasOverrides := len(overrides) > 0
	overrideAllowsBooking := false

	for _, override := range overrides {
		overrideStart, err := availabilityClockTime(
			override.AvailabilityDate,
			override.StartTime,
			location,
		)
		if err != nil {
			return err
		}

		overrideEnd, err := availabilityClockTime(
			override.AvailabilityDate,
			override.EndTime,
			location,
		)
		if err != nil {
			return err
		}

		if !timeRangesOverlap(
			localStart,
			localEnd,
			overrideStart,
			overrideEnd,
		) {
			continue
		}

		if override.Status == "busy" ||
			override.Status == "unavailable" {
			return ErrAvailabilityConflict
		}

		if override.Status == "available" &&
			timeRangeContains(
				overrideStart,
				overrideEnd,
				localStart,
				localEnd,
			) {
			overrideAllowsBooking = true
		}
	}

	/*
		If the cleaner created overrides for this date,
		the overrides become authoritative for the date.
	*/
	if hasOverrides {
		if overrideAllowsBooking {
			return nil
		}

		return ErrOutsideAvailability
	}

	/*
		3. DATE-SPECIFIC AVAILABILITY

		Specific availability takes priority over
		recurring weekly availability.
	*/
	specific, err := s.repo.ListByCleanerIDRange(
		ctx,
		cleanerID,
		date,
		date,
	)
	if err != nil {
		return err
	}

	hasSpecific := len(specific) > 0
	specificAllowsBooking := false

	for _, slot := range specific {
		slotStart, err := availabilityClockTime(
			slot.AvailableDate,
			slot.StartTime,
			location,
		)
		if err != nil {
			return err
		}

		slotEnd, err := availabilityClockTime(
			slot.AvailableDate,
			slot.EndTime,
			location,
		)
		if err != nil {
			return err
		}

		if !timeRangesOverlap(
			localStart,
			localEnd,
			slotStart,
			slotEnd,
		) {
			continue
		}

		if slot.Status == "busy" ||
			slot.Status == "unavailable" {
			return ErrAvailabilityConflict
		}

		if slot.Status == "available" &&
			timeRangeContains(
				slotStart,
				slotEnd,
				localStart,
				localEnd,
			) {
			specificAllowsBooking = true
		}
	}

	if hasSpecific {
		if specificAllowsBooking {
			return nil
		}

		return ErrOutsideAvailability
	}

	/*
		4. RECURRING WEEKLY AVAILABILITY
	*/
	recurring, err := s.repo.ListRecurringByCleanerID(
		ctx,
		cleanerID,
	)
	if err != nil {
		return err
	}

	weekday := int(localStart.Weekday())

	for _, slot := range recurring {
		if slot.Weekday != weekday {
			continue
		}

		recurringStart, err := availabilityClockTime(
			date,
			slot.StartTime,
			location,
		)
		if err != nil {
			return err
		}

		recurringEnd, err := availabilityClockTime(
			date,
			slot.EndTime,
			location,
		)
		if err != nil {
			return err
		}

		if !timeRangesOverlap(
			localStart,
			localEnd,
			recurringStart,
			recurringEnd,
		) {
			continue
		}

		if slot.Status == "busy" ||
			slot.Status == "unavailable" {
			return ErrAvailabilityConflict
		}

		if slot.Status == "available" &&
			timeRangeContains(
				recurringStart,
				recurringEnd,
				localStart,
				localEnd,
			) {
			return nil
		}
	}

	return ErrOutsideAvailability
}

func availabilityClockTime(
	date string,
	clock string,
	location *time.Location,
) (time.Time, error) {
	value, err := time.ParseInLocation(
		"2006-01-02 15:04",
		date+" "+clock,
		location,
	)
	if err != nil {
		return time.Time{}, ErrInvalidInput
	}

	return value, nil
}

func timeRangesOverlap(
	firstStart time.Time,
	firstEnd time.Time,
	secondStart time.Time,
	secondEnd time.Time,
) bool {
	return firstStart.Before(secondEnd) &&
		firstEnd.After(secondStart)
}

func timeRangeContains(
	containerStart time.Time,
	containerEnd time.Time,
	requestStart time.Time,
	requestEnd time.Time,
) bool {
	return !requestStart.Before(containerStart) &&
		!requestEnd.After(containerEnd)
}

func (s *Service) BookingBufferMinutes(
	ctx context.Context,
	cleanerID uint,
) (int, error) {
	if cleanerID == 0 {
		return 0, ErrInvalidInput
	}

	settings, err := s.getSettingsOrDefault(
		ctx,
		cleanerID,
	)
	if err != nil {
		return 0, err
	}

	if settings.BufferMinutes < 0 {
		return 0, ErrInvalidInput
	}

	return settings.BufferMinutes, nil
}
