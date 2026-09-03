package availablenow

import (
	"context"
	"strings"
	"time"
)

const (
	defaultAvailableMinutes = 120
	minAvailableMinutes     = 30
	maxAvailableMinutes     = 720

	defaultSearchLimit = 20
	maxSearchLimit     = 100
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

func (s *Service) SetAvailableNow(
	ctx context.Context,
	cleanerID uint,
	req SetAvailableNowRequest,
) (*CleanerAvailableNow, error) {
	if cleanerID == 0 {
		return nil, ErrInvalidInput
	}

	req.Location = strings.TrimSpace(
		req.Location,
	)

	req.JobTypes = normaliseJobTypes(
		req.JobTypes,
	)

	if req.DurationMinutes == 0 {
		req.DurationMinutes = defaultAvailableMinutes
	}

	if req.DurationMinutes < minAvailableMinutes ||
		req.DurationMinutes > maxAvailableMinutes {
		return nil, ErrInvalidInput
	}

	if req.Location == "" {
		return nil, ErrInvalidInput
	}

	if req.TravelRadiusMiles < 0 ||
		req.TravelRadiusMiles > 100 {
		return nil, ErrInvalidInput
	}

	if len(req.JobTypes) == 0 {
		return nil, ErrInvalidInput
	}

	now := time.Now().UTC()

	availability := &CleanerAvailableNow{
		CleanerID:     cleanerID,
		IsAvailable:   true,
		AvailableFrom: now,
		AvailableUntil: now.Add(
			time.Duration(req.DurationMinutes) *
				time.Minute,
		),
		Location:          req.Location,
		TravelRadiusMiles: req.TravelRadiusMiles,
		JobTypes:          req.JobTypes,
	}

	err := s.repo.Upsert(
		ctx,
		availability,
	)
	if err != nil {
		return nil, err
	}

	return availability, nil
}

func (s *Service) GetStatus(ctx context.Context, cleanerID uint) (*AvailableNowStatusResponse, error) {
	if cleanerID == 0 {
		return nil, ErrInvalidInput
	}

	availability, err := s.repo.GetByCleanerID(
		ctx, cleanerID,
	)

	if err != nil {
		if err == ErrAvailableNowNotFound {
			return &AvailableNowStatusResponse{
				IsAvailable: false,
			}, nil
		}

		return nil, err
	}

	now := time.Now().UTC()

	active := availability.IsAvailable && availability.AvailableFrom.Before(
		now.Add(
			time.Second,
		),
	) &&
		availability.AvailableUntil.After(
			now,
		)

	if !active {
		return &AvailableNowStatusResponse{
			IsAvailable: false,
		}, nil
	}

	remaining := int(time.Until(
		availability.AvailableUntil,
	).Minutes(),
	)

	if remaining < 0 {
		remaining = 0
	}

	return &AvailableNowStatusResponse{
		IsAvailable: true,

		AvailableUntil: availability.AvailableUntil.Format(
			time.RFC3339,
		),

		MinutesRemaining: remaining,
	}, nil

}

func (s *Service) Disable(ctx context.Context, cleanerID uint) error {
	if cleanerID == 0 {
		return ErrInvalidInput
	}

	return s.repo.Disable(
		ctx,
		cleanerID,
	)
}

func (s *Service) Search(ctx context.Context, search AvailableNowSearchRequest) ([]AvailableCleaner, error) {
	search.Location = strings.TrimSpace(search.Location)

	search.JobType = strings.ToLower(
		strings.TrimSpace(
			search.JobType,
		),
	)

	if search.MinimumRating < 0 || search.MinimumRating > 5 {
		return nil, ErrInvalidInput
	}

	if search.Limit <= 0 {
		search.Limit = defaultSearchLimit
	}

	if search.Offset < 0 {
		return nil, ErrInvalidInput
	}

	return s.repo.ListAvailableCleaners(
		ctx,
		search,
		time.Now().UTC(),
	)
}

func normaliseJobTypes(
	values []string,
) []string {
	if len(values) == 0 {
		return []string{}
	}

	result := make(
		[]string,
		0,
		len(values),
	)

	seen := make(
		map[string]struct{},
	)

	for _, value := range values {
		value = strings.ToLower(
			strings.TrimSpace(
				value,
			),
		)
		if value == "" {
			continue
		}

		if _, exists := seen[value]; exists {
			continue
		}

		seen[value] = struct{}{}

		result = append(result, value)
	}

	return result
}
