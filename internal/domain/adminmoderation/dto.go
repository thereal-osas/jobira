package adminmoderation

type UpdateReportStatusRequest struct {
	Status     string `json:"status"`
	AdminNote string `json:"admin_notes"`
}


