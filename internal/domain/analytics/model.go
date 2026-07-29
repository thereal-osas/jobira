package analytics

type DashboardStatus struct {
	TotalUsers           int `json:"total_users"`
	Clients              int `json:"clients"`
	Cleaners             int `json:"cleaners"`
	Companies            int `json:"companies"`
	ActiveJobs           int `json:"active_jobs"`
	CompletedBookings    int `json:"completed_bookings"`
	PendingVerifications int `json:"pending_verifications"`
	OpenReports          int `json:"open_reports"`
	BookingsToday        int `json:"bookings_today"`
}
