package availability

import "time"

type CleanerAvailability struct {
	ID            uint      `json:"id"`
	CleanerID     uint      `json:"cleaner_id"`
	AvailableDate string    `json:"available_date"`
	StartTime     string    `json:"start_time"`
	EndTime       string    `json:"end_time"`
	Status        string    `json:"status"`
	Notes         string    `json:"notes"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type AvailabilityBlock struct {
	ID        uint      `json:"id"`
	CleanerID uint      `json:"cleaner_id"`
	StartAt   time.Time `json:"start_at"`
	EndAt     time.Time `json:"end_at"`
	Reason    string    `json:"reason"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type RecurringAvailability struct {
	ID        uint      `json:"id"`
	CleanerID uint      `json:"cleaner_id"`
	Weekday   int       `json:"weekday"`
	StartTime string    `json:"start_time"`
	EndTime   string    `json:"end_time"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type AvailabilitySettings struct {
	CleanerID          uint      `json:"cleaner_id"`
	MinNoticeMinutes   int       `json:"min_notice_minutes"`
	MinBookingMinutes  int       `json:"min_booking_minutes"`
	MaxBookingMinutes  int       `json:"max_booking_minutes"`
	BufferMinutes      int       `json:"buffer_minutes"`
	BookingHorizonDays int       `json:"booking_horizon_days"`
	Timezone           string    `json:"timezone"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type AvailabilityOverride struct {
	ID               uint      `json:"id"`
	CleanerID        uint      `json:"cleaner_id"`
	AvailabilityDate string    `json:"availability_date"`
	StartTime        string    `json:"start_time"`
	EndTime          string    `json:"end_time"`
	Status           string    `json:"status"`
	Reason           string    `json:"reason"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type CalendarSlot struct {
	Date      string `json:"date"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
	Status    string `json:"status"`
	Source    string `json:"source"`
	Reason    string `json:"reason"`
}
