package workhistory

import (
	"context"
	"strings"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) ListMine(ctx context.Context, userID uint, role string) ([]WorkHistoryEntry, error) {
	if userID == 0 {
		return nil, ErrInvalidInput
	}

	role = strings.ToLower(
		strings.TrimSpace(role),
	)

	switch role {
	case "cleaner":
		return s.repo.ListByCleanerID(
			ctx,
			userID,
		)

	case "client":
		return s.repo.ListByClientID(
			ctx,
			userID,
		)

	default:
		return nil, ErrForbidden
	}
}

func (s *Service) GetBookingHistory(ctx context.Context, bookingID uint, userID uint, role string) (*WorkHistoryDetail, error) {
	if bookingID == 0 || userID == 0 {
		return nil, ErrInvalidInput
	}

	role = strings.ToLower(
		strings.TrimSpace(role),
	)

	if role != "client" &&
		role != "cleaner" && 
		role != "admin" {
			return nil, ErrForbidden
		}

		if role != "admin" {
			allowed, err := 
				s.repo.UserCanAccessBooking(
					ctx,
					bookingID,
					userID,
				)
			if err != nil {
				return nil, err
			}	

			if !allowed {
				return nil, ErrForbidden
			}
		}

		history, err := 
			s.repo.GetByBookingID(
				ctx,
				bookingID,
				userID,
			)
		if err != nil {
			return nil, err

		}	

		return history, nil
}
