package profilemedia

import "context"

type Repository interface {
	Create(ctx context.Context, media *ProfileMedia) error

	GetByID(ctx context.Context, id uint) (*ProfileMedia, error)

	ListByUserID(ctx context.Context, userID uint) ([]ProfileMedia, error)

	ListByCompanyID(ctx context.Context, companyID uint) ([]ProfileMedia, error)

	CountByUserAndType(ctx context.Context, userID uint, mediaType MediaType) (int, error)

	CountByCompanyAndType(ctx context.Context, companyID uint, mediaType MediaType) (int, error)

	CanManageCompany(ctx context.Context, userID uint, companyID uint) (bool, error)

	Update(ctx context.Context, media *ProfileMedia) error

	Delete(ctx context.Context, id uint) error
}
