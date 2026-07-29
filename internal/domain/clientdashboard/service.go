package clientdashboard

import (
	"context"

	subscriptionaccessdomain "github.com/rodrigueghenda/jobira/internal/domain/subscriptionaccess"
)

type Service struct {
	repo                      Repository
	subscriptionAccessService *subscriptionaccessdomain.Service
}

func NewService(repo Repository, subscriptionAccessService *subscriptionaccessdomain.Service) *Service {
	return &Service{
		repo:                      repo,
		subscriptionAccessService: subscriptionAccessService,
	}
}

func (s *Service) GetMine(ctx context.Context, clientID uint) (*ClientDashboard, error) {
	if clientID == 0 {
		return nil, ErrInvalidInput
	}

	accessStatus, err := s.subscriptionAccessService.GetClientAccessStatus(ctx, clientID)
	if err != nil {
		return nil, err
	}

	activeJobs, err := s.repo.ListActiveJobs(ctx, clientID, 5)
	if err != nil {
		return nil, err
	}

	upcomingBookings, err := s.repo.ListUpcomingBookings(ctx, clientID, 5)
	if err != nil {
		return nil, err
	}

	applicationsReceived, err := s.repo.CountApplicationsReceived(ctx, clientID)
	if err != nil {
		return nil, err
	}

	completedBookingCount, err := s.repo.CountCompletedBookings(ctx, clientID)
	if err != nil {
		return nil, err
	}

	favouriteCleanerCount, err := s.repo.CountFavouriteCleaners(ctx, clientID)
	if err != nil {
		return nil, err
	}

	preferredCleanerCount, err := s.repo.CountPreferredCleaners(ctx, clientID)
	if err != nil {
		return nil, err
	}

	repeatBookingCount, err := s.repo.CountRepeatBookings(ctx, clientID)
	if err != nil {
		return nil, err
	}

	dashboard := &ClientDashboard{
		ClientID: clientID,

		SubscriptionStatus: accessStatus.SubscriptionStatus,
		LaunchGraceActive:  accessStatus.LaunchGraceActive,
		Premium:            accessStatus.Premium,

		CanPostJob:        accessStatus.CanPostJob,
		JobPostCount:      accessStatus.JobPostCount,
		FreeJobPostLimit:  accessStatus.FreeJobPostLimit,
		JobsPostedToday:   accessStatus.JobsPostedToday,
		DailyJobPostLimit: accessStatus.DailyJobPostLimit,

		ActiveJobCount:        len(activeJobs),
		ApplicationsReceived:  applicationsReceived,
		ActiveBookingCount:    len(upcomingBookings),
		CompletedBookingCount: completedBookingCount,
		FavouriteCleanerCount: favouriteCleanerCount,
		PreferredCleanerCount: preferredCleanerCount,
		RepeatBookingCount:    repeatBookingCount,

		ActiveJobs:       activeJobs,
		UpcomingBookings: upcomingBookings,

		UpgradeMessage: accessStatus.UpgradeMessage,
	}

	return dashboard, nil
}
