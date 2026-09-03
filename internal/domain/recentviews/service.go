package recentviews

import (
	"context"
	"time"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) RecordView(ctx context.Context, clientID uint, cleanerID uint) error {
	if clientID == 0 || cleanerID == 0 {
		return ErrInvalidInput
	}

	if clientID == cleanerID {
		return ErrInvalidInput
	}

	return s.repo.RecordView(ctx, clientID, cleanerID)
}

func (s *Service) ListMine(ctx context.Context, clientID uint) ([]RecentlyViewedCleaner, error) {
	if clientID == 0 {
		return nil, ErrInvalidInput
	}

	return s.repo.ListByClientID(ctx, clientID)
}

func (s *Service) GetProfileViewAnalytics(ctx context.Context, cleanerID uint) (*ProfileViewAnalytics, error) {
	if cleanerID == 0 {
		return nil, ErrInvalidInput
	}

	now := time.Now()

	startOfToday := time.Date(
		now.Year(),
		now.Month(),
		now.Day(),
		0,
		0,
		0,
		0,
		now.Location(),
	)

	startOfThisWeek := startOfWeek(now)

	startOfLastWeek := startOfThisWeek.AddDate(
		0,
		0,
		-7,
	)

	totalViews, err := s.repo.CountByCleanerID(
		ctx,
		cleanerID,
	)
	if err != nil {
		return nil, err
	}

	uniqueViewers, err :=
		s.repo.CountUniqueViewersByCleanerID(
			ctx,
			cleanerID,
		)
	if err != nil {
		return nil, err
	}

	viewsToday, err :=
		s.repo.CountByCleanerIDSince(
			ctx,
			cleanerID,
			startOfToday,
		)
	if err != nil {
		return nil, err
	}

	viewsThisWeek, err :=
		s.repo.CountByCleanerIDSince(
			ctx,
			cleanerID,
			startOfThisWeek,
		)
	if err != nil {
		return nil, err
	}

	viewsLastWeek, err :=
		s.repo.CountByCleanerIDBetween(
			ctx,
			cleanerID,
			startOfLastWeek,
			startOfThisWeek,
		)
	if err != nil {
		return nil, err
	}

	weeklyChange := calculateWeeklyChange(
		viewsThisWeek,
		viewsLastWeek,
	)

	return &ProfileViewAnalytics{
		CleanerID:       cleanerID,
		TotalViews:      totalViews,
		UniqueViewers:   uniqueViewers,
		ViewsToday:      viewsToday,
		ViewsThisWeek:   viewsThisWeek,
		ViewsLastWeek:   viewsLastWeek,
		WeeklyChange:    weeklyChange,
		HasWeeklyGrowth: viewsThisWeek > viewsLastWeek,
	}, nil
}

func startOfWeek(t time.Time) time.Time {
	weekday := int(t.Weekday())

	if weekday == 0 {
		weekday = 7
	}

	start := t.AddDate(
		0,
		0,
		-(weekday - 1),
	)

	return time.Date(
		start.Year(),
		start.Month(),
		start.Day(),
		0,
		0,
		0,
		0,
		start.Location(),
	)
}

func calculateWeeklyChange(current int, previous int) float64 {
	if previous == 0 {
		if current > 0 {
			return 100
		}

		return 0
	}

	return (float64(current-previous) /
		float64(previous)) * 100
}
