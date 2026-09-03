package profilemedia

import (
	"context"
	"strings"
)

type Service struct {
	repo Repository
}

func NewService(
	repo Repository,
) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) CreateForUser(ctx context.Context, userID uint, req CreateMediaRequest) (*ProfileMedia, error) {
	if userID == 0 {
		return nil, ErrInvalidInput
	}

	if err := validateCreateRequest(req); err != nil {
		return nil, err
	}

	if !isValidUserMediaType(req.MediaType) {
		return nil, ErrInvalidMediaType
	}

	count, err := s.repo.CountByUserAndType(
		ctx,
		userID,
		req.MediaType,
	)
	if err != nil {
		return nil, err
	}

	if err := validateMediaLimit(
		req.MediaType,
		count,
	); err != nil {
		return nil, err
	}

	ownerUserID := userID

	media := &ProfileMedia{
		OwnerUserID: &ownerUserID,
		CompanyID:   nil,
		MediaType:   req.MediaType,
		URL:         strings.TrimSpace(req.URL),
		Caption:     strings.TrimSpace(req.Caption),
		SortOrder:   req.SortOrder,
	}

	if err := s.repo.Create(
		ctx,
		media,
	); err != nil {
		return nil, err
	}

	return media, nil
}
func (s *Service) CreateForCompany(ctx context.Context, userID uint, companyID uint, req CreateMediaRequest) (*ProfileMedia, error) {
	if userID == 0 || companyID == 0 {
		return nil, ErrInvalidInput
	}

	canManage, err := s.repo.CanManageCompany(
		ctx,
		userID,
		companyID,
	)
	if err != nil {
		return nil, err
	}

	if !canManage {
		return nil, Errforbidden
	}

	if err := validateCreateRequest(req); err != nil {
		return nil, err
	}

	if !isValidCompanyMediaType(req.MediaType) {
		return nil, ErrInvalidMediaType
	}

	count, err := s.repo.CountByCompanyAndType(
		ctx,
		companyID,
		req.MediaType,
	)
	if err != nil {
		return nil, err
	}

	if err := validateMediaLimit(
		req.MediaType,
		count,
	); err != nil {
		return nil, err
	}

	id := companyID

	media := &ProfileMedia{
		OwnerUserID: nil,
		CompanyID:   &id,
		MediaType:   req.MediaType,
		URL:         strings.TrimSpace(req.URL),
		Caption:     strings.TrimSpace(req.Caption),
		SortOrder:   req.SortOrder,
	}

	if err := s.repo.Create(
		ctx,
		media,
	); err != nil {
		return nil, err
	}

	return media, nil
}

func (s *Service) GetByID(ctx context.Context, id uint) (*ProfileMedia, error) {
	if id == 0 {
		return nil, ErrInvalidInput
	}

	return s.repo.GetByID(
		ctx,
		id,
	)
}

func (s *Service) ListForUser(ctx context.Context, userID uint) ([]ProfileMedia, error) {
	if userID == 0 {
		return nil, ErrInvalidInput
	}

	return s.repo.ListByUserID(
		ctx,
		userID,
	)
}

func (s *Service) ListForCompany(ctx context.Context, companyID uint) ([]ProfileMedia, error) {
	if companyID == 0 {
		return nil, ErrInvalidInput
	}

	return s.repo.ListByCompanyID(
		ctx,
		companyID,
	)
}

func (s *Service) UpdateForUser(ctx context.Context, userID uint, mediaID uint, req UpdateMediaRequest) (*ProfileMedia, error) {
	if userID == 0 || mediaID == 0 {
		return nil, ErrInvalidInput
	}

	media, err := s.repo.GetByID(
		ctx,
		mediaID,
	)
	if err != nil {
		return nil, err
	}

	if media.OwnerUserID == nil || *media.OwnerUserID != userID {
		return nil, Errforbidden
	}

	media.Caption = strings.TrimSpace(
		req.Caption,
	)

	media.SortOrder = req.SortOrder

	if err := s.repo.Update(
		ctx,
		media,
	); err != nil {
		return nil, err
	}

	return media, nil
}
func (s *Service) UpdateForCompany(ctx context.Context, userID uint, companyID uint, mediaID uint, req UpdateMediaRequest) (*ProfileMedia, error) {
	if userID == 0 ||
		companyID == 0 ||
		mediaID == 0 {
		return nil, ErrInvalidInput
	}

	canManage, err := s.repo.CanManageCompany(
		ctx,
		userID,
		companyID,
	)
	if err != nil {
		return nil, err
	}

	if !canManage {
		return nil, Errforbidden
	}

	media, err := s.repo.GetByID(
		ctx,
		mediaID,
	)
	if err != nil {
		return nil, err
	}

	if media.CompanyID == nil ||
		*media.CompanyID != companyID {
		return nil, Errforbidden
	}

	media.Caption = strings.TrimSpace(
		req.Caption,
	)
	media.SortOrder = req.SortOrder

	if err := s.repo.Update(
		ctx,
		media,
	); err != nil {
		return nil, err
	}

	return media, nil
}
func (s *Service) DeleteForUser(ctx context.Context, userID uint, mediaID uint) error {
	if userID == 0 || mediaID == 0 {
		return ErrInvalidInput
	}

	media, err := s.repo.GetByID(
		ctx,
		mediaID,
	)
	if err != nil {
		return err
	}

	if media.OwnerUserID == nil || *media.OwnerUserID != userID {
		return Errforbidden
	}

	return s.repo.Delete(
		ctx,
		mediaID,
	)
}

func (s *Service) DeleteForCompany(ctx context.Context, userID uint, companyID uint, mediaID uint) error {
	if userID == 0 ||
		companyID == 0 ||
		mediaID == 0 {
		return ErrInvalidInput
	}

	canManage, err := s.repo.CanManageCompany(
		ctx,
		userID,
		companyID,
	)
	if err != nil {
		return err
	}

	if !canManage {
		return Errforbidden
	}

	media, err := s.repo.GetByID(
		ctx,
		mediaID,
	)
	if err != nil {
		return err
	}

	if media.CompanyID == nil ||
		*media.CompanyID != companyID {
		return Errforbidden
	}

	return s.repo.Delete(
		ctx,
		mediaID,
	)
}

func validateCreateRequest(req CreateMediaRequest) error {
	if strings.TrimSpace(req.URL) == "" {
		return ErrInvalidInput
	}

	if req.SortOrder < 0 {
		return ErrInvalidInput
	}

	if !isKnownMediaType(req.MediaType) {
		return ErrInvalidMediaType
	}

	return nil
}

func isKnownMediaType(mediaType MediaType) bool {
	switch mediaType {
	case MediaTypeProfilePhoto,
		MediaTypePortfolio,
		MediaTypePricing,
		MediaTypeCompanyLogo:
		return true

	default:
		return false
	}
}

func isValidUserMediaType(mediaType MediaType) bool {
	switch mediaType {
	case MediaTypeProfilePhoto,
		MediaTypePortfolio,
		MediaTypePricing:
		return true

	default:
		return false
	}
}

func isValidCompanyMediaType(mediaType MediaType) bool {
	switch mediaType {
	case MediaTypePortfolio,
		MediaTypePricing,
		MediaTypeCompanyLogo:
		return true

	default:
		return false
	}
}

func validateMediaLimit(mediaType MediaType, currentCount int) error {
	limits := DefaultMediaLimits()

	switch mediaType {
	case MediaTypeProfilePhoto:
		if currentCount >= limits.ProfilePhoto {
			return ErrProfilePhotoLimitReached
		}

	case MediaTypePortfolio:
		if currentCount >= limits.Portfolio {
			return ErrPortfolioLimitReached
		}

	case MediaTypePricing:
		if currentCount >= limits.Pricing {
			return ErrPricingLimitReached
		}

	case MediaTypeCompanyLogo:
		if currentCount >= limits.CompanyLogo {
			return ErrCompanyLogoLimitReached
		}

	default:
		return ErrInvalidMediaType
	}

	return nil
}
