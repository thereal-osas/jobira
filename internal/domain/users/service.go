package users

import "context"

 type Service struct {
	repo Repository
 }

 func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
	}
 }

 func (s *Service) GetCurrentUser(ctx context.Context, userID uint) (*User, error) {
	return s.repo.GetByID(ctx, userID)
 }
