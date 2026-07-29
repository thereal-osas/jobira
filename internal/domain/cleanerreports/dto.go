package cleanerreports

type CreateCleanerReportRequest struct {
	Reason  string `json:"reason"`
	Details string `json:"details"`
}
