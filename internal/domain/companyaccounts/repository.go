package companyaccounts

import "context"

type Repository interface {
	CreateCompany(ctx context.Context, company *Company) error
	AddMember(ctx context.Context, companyID uint, userID uint, role string) error
	GetByID(ctx context.Context, companyID uint) (*Company, error)
	RemoveMember(ctx context.Context, companyID uint, userID uint) error
	ListMembers(ctx context.Context, companyID uint) ([]CompanyMember, error)
	CountMembers(ctx context.Context, companyID uint) (int, error)
	GetCleanerSeatLimit(ctx context.Context, companyID uint) (int, error)
}
