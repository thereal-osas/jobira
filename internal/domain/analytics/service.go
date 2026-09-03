package analytics

import "context"

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Dashboard(ctx context.Context) (*DashboardStatus, error) {
	totalUsers, err := s.repo.TotalUsers(ctx)
	if err != nil {
		return nil, err
	}

	clients, err := s.repo.UsersByRole(ctx, "client")
	if err != nil {
		return nil, err
	}

	cleaners, err := s.repo.UsersByRole(ctx, "cleaner")
	if err != nil {
		return nil, err
	}

	companies, err := s.repo.TotalCompanies(ctx)
	if err != nil {
		return nil, err
	}

	activeJobs, err := s.repo.ActiveJobs(ctx)
	if err != nil {
		return nil, err
	}

	completedBookings, err := s.repo.CompletedBookings(ctx)
	if err != nil {
		return nil, err
	}

	pendingVerifications, err := s.repo.PendingVerifications(ctx)
	if err != nil {
		return nil, err
	}

	openReports, err := s.repo.OpenReports(ctx)
	if err != nil {
		return nil, err
	}

	bookingsToday, err := s.repo.BookingsToday(ctx)
	if err != nil {
		return nil, err
	}

	status := &DashboardStatus{
		TotalUsers:           totalUsers,
		Clients:              clients,
		Cleaners:             cleaners,
		Companies:            companies,
		ActiveJobs:           activeJobs,
		CompletedBookings:    completedBookings,
		PendingVerifications: pendingVerifications,
		OpenReports:          openReports,
		BookingsToday:        bookingsToday,
	}

	return status, nil
}
