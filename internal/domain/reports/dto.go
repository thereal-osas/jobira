package reports

type CreateReportRequest struct {
	ReportedUserID *uint  `json:"reported_user_id"`
	JobID          *uint  `json:"job_id"`
	BookingID      *uint  `json:"booking_id"`
	ReportType     string `json:"report_type"`
	Reason         string `json:"reason"`
	Details        string `json:"details"`
}

type ReviewReportRequest struct {
	Status     string `json:"status"`
	AdminNotes string `json:"admin_notes"`
}
