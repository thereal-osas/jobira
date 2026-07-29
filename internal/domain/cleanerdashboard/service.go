package cleanerdashboard

import (
	"context"

	reputationdomain "github.com/rodrigueghenda/jobira/internal/domain/reputation"
	subscriptionaccessdomain "github.com/rodrigueghenda/jobira/internal/domain/subscriptionaccess"
)

type Service struct {
	repo                      Repository
	reputationService         *reputationdomain.Service
	subscriptionaccessService *subscriptionaccessdomain.Service
}

func NewService(
	repo Repository,
	reputationService *reputationdomain.Service,
	subscriptionAccessService *subscriptionaccessdomain.Service,
) *Service {
	return &Service{
		repo:                      repo,
		reputationService:         reputationService,
		subscriptionaccessService: subscriptionAccessService,
	}
}

func (s *Service) GetMine(ctx context.Context, cleanerID uint) (*CleanerDashboard, error) {
	if cleanerID == 0 {
		return nil, ErrInvalidInput
	}

	accessStatus, err := s.subscriptionaccessService.GetCleanerAccessStatus(ctx, cleanerID)
	if err != nil {
		return nil, err
	}

	reputation, err := s.reputationService.GetByCleanerID(ctx, cleanerID)
	if err != nil {
		return nil, err
	}

	favouriteCount, err := s.repo.CountFavourites(ctx, cleanerID)
	if err != nil {
		return nil, err
	}

	preferredClientCount, err := s.repo.CountPreferredClients(ctx, cleanerID)
	if err != nil {
		return nil, err
	}

	upcomingBookings, err := s.repo.ListUpcomingBookings(ctx, cleanerID, 5)
	if err != nil {
		return nil, err
	}

	dashboard := &CleanerDashboard{
		CleanerID: cleanerID,

		SubscriptionStatus: accessStatus.SubscriptionStatus,
		TrialActive:        accessStatus.TrialActive,
		LaunchGraceActive:  accessStatus.LaunchGraceActive,
		Premium:            accessStatus.Premium,

		CanApply:              accessStatus.CanApply,
		ApplicationsToday:     accessStatus.ApplicationsToday,
		DailyApplicationLimit: accessStatus.DailyApplicationLimit,

		AverageRating:            reputation.AverageRating,
		TotalReviews:             reputation.TotalReviews,
		CompletedJobs:            reputation.CompletedJobs,
		RepeatClients:            reputation.RepeatClients,
		RecommendationPercentage: reputation.RecommendationPercentage,
		Badge:                    reputation.Badge,

		FavouriteCount:       favouriteCount,
		PreferredClientCount: preferredClientCount,
		UpcomingBookingCount: len(upcomingBookings),
		UpcomingBookings:     upcomingBookings,

		UpgradeMessage: accessStatus.UpgradeMessage,
	}

	return dashboard, nil
}
