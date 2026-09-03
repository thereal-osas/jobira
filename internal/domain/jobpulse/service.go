package jobpulse

import (
	"context"
	"strings"
	"time"
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

func (s *Service) GetJobPulse(ctx context.Context, jobID uint) (*JobPulse, error) {
	if jobID == 0 {
		return nil, ErrInvalidInput
	}

	snapshot, err := s.repo.GetSnapshot(
		ctx,
		jobID,
	)
	if err != nil {
		return nil, err
	}

	if snapshot == nil {
		return nil, ErrJobNotFound
	}

	now := time.Now().UTC()

	jobAge := now.Sub(
		snapshot.CreatedAt,
	)

	jobAgeHours := int(
		jobAge.Hours(),
	)

	if jobAgeHours < 0 {
		jobAgeHours = 0
	}

	hasBooking :=
		snapshot.BookingCount > 0

	isFilled := hasBooking || jobStatusIsFilled(
		snapshot.JobStatus,
	)

	score :=
		calculatePulseScore(
			snapshot.ApplicationCount,
			snapshot.RecentApplicationCount,
			jobAgeHours,
			isFilled,
		)

	level :=
		determinePulseLevel(
			score,
			isFilled,
		)

	message :=
		buildPulseMessage(
			level,
			snapshot.ApplicationCount,
			snapshot.RecentApplicationCount,
			isFilled,
		)

	return &JobPulse{
		JobID: snapshot.JobID,

		ApplicationCount: snapshot.ApplicationCount,

		RecentApplicationCount: snapshot.RecentApplicationCount,

		JobAgeHours: jobAgeHours,

		HasBooking: hasBooking,

		IsFilled: isFilled,

		PulseLevel: level,

		PulseScore: score,

		Message: message,

		UpdatedAt: now,
	}, nil
}

func calculatePulseScore(applicationCount int, recentApplicationCount int, jobAgeHours int, isFilled bool) int {
	if isFilled {
		return 100
	}

	score := 0

	switch {
	case applicationCount >= 15:
		score += 45

	case applicationCount >= 10:
		score += 38

	case applicationCount >= 5:
		score += 28

	case applicationCount >= 2:
		score += 18

	case applicationCount == 1:
		score += 10
	}

	switch {
	case recentApplicationCount >= 8:
		score += 40

	case recentApplicationCount >= 5:
		score += 32

	case recentApplicationCount >= 3:
		score += 24

	case recentApplicationCount >= 1:
		score += 12
	}

	switch {
	case jobAgeHours <= 6:
		score += 15

	case jobAgeHours <= 24:
		score += 10

	case jobAgeHours <= 72:
		score += 5
	}

	return clampPulseScore(
		score,
	)
}

func determinePulseLevel(score int, isFilled bool) PulseLevel {
	if isFilled {
		return PulseCompetitive
	}

	switch {
	case score >= 70:
		return PulseCompetitive

	case score >= 45:
		return PulseHeatingUp

	case score >= 20:
		return PulseActive

	default:
		return PulseQuiet
	}
}

func buildPulseMessage(level PulseLevel, applicationCount int, recentApplicationCount int, isFilled bool) string {
	if isFilled {
		return "This job has already move into booking."
	}

	switch level {
	case PulseCompetitive:
		return "This job is receiving high interest."

	case PulseHeatingUp:
		if recentApplicationCount > 0 {
			return "This job is heating up with recent application."
		}

		return "This job is attracting strong interest."

	case PulseActive:
		if applicationCount == 1 {
			return "1 cleaner has applied so far."
		}

		return "Several cleaners have applied."

	default:
		return "Be one of the first cleaners to apply."
	}
}

func jobStatusIsFilled(status string) bool {
	switch strings.ToLower(
		strings.TrimSpace(
			status,
		),
	) {
	case "filled",
		"booked",
		"completed",
		"closed":
		return true

	default:
		return false
	}
}

func clampPulseScore(score int) int {
	if score < 0 {
		return 0
	}

	if score > 100 {
		return 100
	}

	return score
}
