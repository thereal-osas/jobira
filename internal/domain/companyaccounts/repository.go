package companyaccounts

import "context"

type Repository interface {
	CreateCompany(ctx context.Context, company *Company) error
	AddMember(ctx context.Context, companyID uint, userID uint, role string) error
	GetByID(ctx context.Context, companyID uint) (*Company, error)
	RemoveMember(ctx context.Context, companyID uint, userID uint) error
	GetMember(ctx context.Context, companyID uint, userID uint) (*CompanyMember, error)
	ListMembers(ctx context.Context, companyID uint) ([]CompanyMember, error)
	CountActiveCleanerSeats(ctx context.Context, companyID uint) (int, error)
	GetCleanerSeatLimit(ctx context.Context, companyID uint) (int, error)
	UpdateMemberStatus(ctx context.Context, companyID uint, userID uint, status string) error
	UpdateMemberRole(ctx context.Context, companyID uint, userID uint, role string) error
	CountMembersByRolesAndStatus(ctx context.Context, comapanyID uint, role string, status string) (int, error)
}
