package reports

import "time"

type Report struct {
	ID             uint       `json:"id"`
	ReporterID     uint       `json:"reporter_id"`
	ReportedUserID *uint      `json:"reported_user_id"`
	JobID          *uint      `json:"job_id"`
	BookingID      *uint      `json:"booking_id"`
	ReportType     string     `json:"report_type"`
	Reason         string     `json:"reason"`
	Details        string     `json:"details"`
	Status         string     `json:"status"`
	AdminNotes     string     `json:"admin_notes"`
	ReviewedBy     *uint      `json:"reviewed_by"`
	ReviewedAt     *time.Time `json:"reviewed_at"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}
