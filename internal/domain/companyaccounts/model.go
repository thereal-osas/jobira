package companyaccounts

import "time"

type Company struct {
	ID          uint      `json:"id"`
	OwnerID     uint      `json:"owner_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CompanyMember struct {
	ID        uint      `json:"id"`
	CompanyID uint      `json:"company_id"`
	UserID    uint      `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}
