package adminmoderation

import "time"

type CleanerReportAdminView struct {
	ID         uint      `json:"id"`
	ClientID   uint      `json:"client_id"`
	CleanerID  uint      `json:"cleaner_id"`
	Reason     string    `json:"reason"`
	Details    string    `json:"details"`
	Status     string    `json:"status"`
	AdminNotes string    `json:"admin_notes"`
	ResolvedBy *uint     `json:"resolved_by"`
	ResolvedAt *time.Time `json:"resolved_at"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type BlockedCleanerAdminView struct {
	ID        uint      `json:"id"`
	ClientID  uint      `json:"client_id"`
	CleanerID uint      `json:"cleaner_id"`
	Reason    string    `json:"reason"`
	CreatedAt time.Time `json:"created_at"`
}
