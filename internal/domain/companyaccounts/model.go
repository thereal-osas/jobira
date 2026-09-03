package companyaccounts

import "time"

type Company struct {
	ID uint `json:"id"`

	OwnerID uint `json:"owner_id"`

	Name string `json:"name"`

	Description string `json:"description"`

	CreatedAt time.Time `json:"created_at"`

	UpdatedAt time.Time `json:"updated_at"`
}

type CompanyMember struct {
	ID uint `json:"id"`

	CompanyID uint `json:"company_id"`

	UserID uint `json:"user_id"`

	FullName string `json:"full_name"`

	Role string `json:"role"`

	Status string `json:"status"`

	IsVerified bool `json:"is_verified"`

	ReliabilityScore int `json:"reliability_score"`

	Badge string `json:"badge"`

	CreatedAt time.Time `json:"created_at"`
}

type CompanySeatUsage struct {
	CompanyID uint `json:"company_id"`

	SeatLimit int `json:"seat_limit"`

	SeatsUsed int `json:"seats_used"`

	SeatsRemaining int `json:"seats_remaining"`
}

type CompanyDashboard struct {
	Company Company `json:"company"`

	TotalMembers int `json:"total_members"`

	ActiveCleaners int `json:"active_cleaners"`

	InactiveCleaners int `json:"inactive_cleaners"`

	AdminCount int `json:"admin_count"`

	SeatUsage CompanySeatUsage `json:"seat_usage"`
}
