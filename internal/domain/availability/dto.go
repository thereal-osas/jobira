package availability

type CleanerAvailabilityRequest struct {
	AvailableDate string `json:"available_date"`
	StartTime     string `json:"start_time"`
	EndTime       string `json:"end_time"`
	Status        string `json:"status"`
	Notes         string `json:"notes"`
}

type UpdateAvailabilityRequest struct {
	AvailableDate string `json:"available_date"`
	StartTime     string `json:"start_time"`
	EndTime       string `json:"end_time"`
	Status        string `json:"status"`
	Notes         string `json:"notes"`
}

type CreateAvailabilityBlockRequest struct {
	StartAt string `json:"start_at"`
	EndAt   string `json:"end_at"`
	Reason  string `json:"reason"`
}

type CreateRecurringAvailabilityRequest struct {
	Weekday   int    `json:"weekday"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
	Status    string `json:"status"`
}

type BulkRecurringAvailabilityRequest struct {
	Slots []CreateRecurringAvailabilityRequest `json:"slots"`
}

type UpdateAvailabilitySettingsRequest struct {
	MinNoticeMinutes   int    `json:"min_notice_minutes"`
	MinBookingMinutes  int    `json:"min_booking_minutes"`
	MaxBookingMinutes  int    `json:"max_booking_minutes"`
	BufferMinutes      int    `json:"buffer_minutes"`
	BookingHorizonDays int    `json:"booking_horizon_days"`
	Timezone           string `json:"timezone"`
}

type CreateAvailabilityOverrideRequest struct {
	AvailableDate string `json:"available_date"`
	StartTime     string `json:"start_time"`
	EndTime       string `json:"end_time"`
	Status        string `json:"status"`
	Reason        string `json:"reason"`
}

type BulkAvailabilityRequest struct {
	Slots []CleanerAvailabilityRequest `json:"slots"`
}
