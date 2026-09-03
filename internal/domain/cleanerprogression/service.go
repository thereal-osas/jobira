package cleanerprogression

import (
	"context"
	"errors"

	reputationdomain "github.com/rodrigueghenda/jobira/internal/domain/reputation"
)

var (
	ErrInvalidInput = errors.New("invalid input")
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

func (s *Service) GetMyProgression(ctx context.Context, cleanerID uint) (*ProgressSnapshot, error) {
	if cleanerID == 0 {
		return nil, ErrInvalidInput
	}

	reputation, err := s.repo.GetCleanerReputation(
		ctx,
		cleanerID,
	)
	if err != nil {
		return nil, err
	}

	if reputation == nil {
		return nil, reputationdomain.ErrReputationNotFound
	}

	enrichProgressionMetrics(
		reputation,
	)

	badgeProgress := calculateBadgeProgress(
		reputation,
	)

	milestones := calculateMilestones(
		reputation,
	)

	overallProgress := calculateOverallProgress(
		reputation,
		badgeProgress,
	)

	return &ProgressSnapshot{
		CleanerID: reputation.CleanerID,

		CurrentBadge: reputation.Badge,

		ReliabilityScore: reputation.ReliabilityScore,

		CompletionRate: reputation.CompletionRate,

		CancellationRate: reputation.CancellationRate,

		ResponseRate: reputation.ResponseRate,

		AverageRating: reputation.AverageRating,

		TotalReviews: reputation.TotalReviews,

		CompletedJobs: reputation.CompletedJobs,

		RepeatClients: reputation.RepeatClients,

		RecommendationPercentage: reputation.RecommendationPercentage,

		BadgeProgress: badgeProgress,

		Milestones: milestones,

		OverallProgress: overallProgress,

		Encouragement: calculateEncouragement(
			reputation,
			badgeProgress,
		),
	}, nil
}

func calculateBadgeProgress(reputation *reputationdomain.CleanerReputation) BadgeProgress {
	if reputation == nil {
		return BadgeProgress{}
	}

	switch reputation.Badge {
	case "Elite Cleaner":
		return BadgeProgress{
			CurrentBadge: "Elite Cleaner",

			Progress: 100,

			IsHighestLevel: true,
		}

	case "Top Rated":
		return badgeProgressTowardsElite(
			reputation,
		)
	case "Expirenced Cleaner":
		return badgesProgressTowardsTopRated(
			reputation,
		)
	case "Aactive Cleaner":
		return badgeProgressTowardsExperienced(
			reputation,
		)
	default:
		return badgesProgressTowardsActive(
			reputation,
		)
	}
}

func badgesProgressTowardsActive(reputation *reputationdomain.CleanerReputation,
) BadgeProgress {
	current := reputation.CompletedJobs
	target := 1

	return BadgeProgress{
		CurrentBadge: reputation.Badge,

		NextBadge: "Active Cleaner",

		Progress: percentage(
			current,
			target,
		),

		JobsRemaining: remaining(
			current,

			target,
		),
	}
}

func badgeProgressTowardsExperienced(reputation *reputationdomain.CleanerReputation,
) BadgeProgress {
	const target = 50

	return BadgeProgress{
		CurrentBadge: reputation.Badge,

		NextBadge: "Experienced Cleaner",

		Progress: percentage(
			reputation.CompletedJobs,
			target,
		),

		JobsRemaining: remaining(
			reputation.CompletedJobs,
			target,
		),
	}
}

func badgesProgressTowardsTopRated(reputation *reputationdomain.CleanerReputation) BadgeProgress {

	reviewProgress := percentage(
		reputation.TotalReviews,
		10,
	)

	ratingProgress :=
		floatPercentage(
			reputation.AverageRating,
			4.8,
		)

	recommendationProgress :=
		percentage(
			reputation.RecommendationPercentage,
			90,
		)

	progress := average(
		reviewProgress,
		ratingProgress,
		recommendationProgress,
	)

	return BadgeProgress{
		CurrentBadge: reputation.Badge,

		NextBadge: "Top Rated",

		Progress: progress,

		ReviewsNeeded: remaining(
			reputation.TotalReviews,
			10,
		),

		RatingNeeded: floatRemaining(
			reputation.AverageRating,
			4.8,
		),

		RecommendationNeeded: remaining(
			reputation.RecommendationPercentage,
			90,
		),
	}
}

func badgeProgressTowardsElite(reputation *reputationdomain.CleanerReputation) BadgeProgress {
	reviewProgress := percentage(
		reputation.TotalReviews,
		20,
	)

	ratingProgress :=
		floatPercentage(
			reputation.AverageRating,
			4.9,
		)

	recommendationProgress :=
		percentage(
			reputation.RecommendationPercentage,
			95,
		)

	progress := average(
		reviewProgress,
		ratingProgress,
		recommendationProgress,
	)

	return BadgeProgress{
		CurrentBadge: reputation.Badge,

		NextBadge: "Elite Cleaner",

		Progress: progress,

		ReviewsNeeded: remaining(
			reputation.TotalReviews,
			20,
		),

		RatingNeeded: floatRemaining(
			reputation.AverageRating,
			4.9,
		),

		RecommendationNeeded: remaining(
			reputation.RecommendationPercentage,
			95,
		),
	}
}

func calculateMilestones(reputation *reputationdomain.CleanerReputation) []ProgressMetric {
	if reputation == nil {
		return []ProgressMetric{}
	}

	return []ProgressMetric{
		buildMetric(
			"first_job",
			"First job",
			reputation.CompletedJobs,
			1,
			"Complete your first Jobira job",
		),

		buildMetric(
			"five_jobs",
			"Getting Established",
			reputation.CompletedJobs,
			5,
			"Complete 5 jobs",
		),

		buildMetric(
			"ten_jobs",
			"10 Jobs Completed",
			reputation.CompletedJobs,
			10,
			"Complete 10 successful jobs",
		),

		buildMetric(
			"fifty_jobs",
			"50 Job Milestone",
			reputation.CompletedJobs,
			50,
			"Complete 50 successful jobs",
		),

		buildMetric(
			"ten_reviews",
			"Trusted by Clients",
			reputation.TotalReviews,
			10,
			"Receive 10 client reviews",
		),

		buildMetric(
			"repeat_clients",
			"Repeat Favourite",
			reputation.RepeatClients,
			3,
			"Have at least 3 repeat clients",
		),

		buildMetric(
			"recommended",
			"Highly Recommended",
			reputation.RecommendationPercentage,
			90,
			"Reach a 90% recommendation rate",
		),

		buildMetric(
			"responsive",
			"Quick Responder",
			reputation.ResponseRate,
			90,
			"Reach a 90% response rate",
		),
	}
}

func buildMetric(code string, name string, current int, target int, description string) ProgressMetric {
	return ProgressMetric{
		Code: code,

		Name: name,

		Current: current,

		Target: target,

		Percentage: percentage(
			current,
			target,
		),

		Completed: current >= target,

		Description: description,
	}
}

func calculateOverallProgress(
	reputation *reputationdomain.CleanerReputation,
	badgeProgress BadgeProgress,
) int {
	if reputation == nil {
		return 0
	}

	if badgeProgress.IsHighestLevel {
		return 100
	}

	reliability := clamp(
		reputation.ReliabilityScore,
	)

	rating := floatPercentage(
		reputation.AverageRating,
		5.0,
	)

	recommendation := clamp(
		reputation.RecommendationPercentage,
	)

	return average(
		badgeProgress.Progress,
		reliability,
		rating,
		recommendation,
	)
}

func calculateEncouragement(reputation *reputationdomain.CleanerReputation, progress BadgeProgress) string {
	if reputation == nil {
		return ""
	}

	if progress.IsHighestLevel {
		return "You have reached Jobira's highest cleaner level."
	}

	if progress.Progress >= 90 {
		return "You're very close to your next badge."
	}

	if progress.Progress >= 70 {
		return "Strong progress. keep building your client history."
	}

	if reputation.CompletedJobs == 0 {
		return "Complete your first job to start building your Jobira reputation."
	}

	return "Every completed job strengthens your Jobira reputation."
}

func percentage(value int, target int) int {
	if target <= 0 {
		return 0
	}

	if value >= target {
		return 100
	}

	if value <= 0 {
		return 0
	}

	return clamp(
		(value * 100) / target,
	)
}

func floatPercentage(
	value float64,
	target float64,
) int {
	if target <= 0 {
		return 0
	}

	if value >= target {
		return 0
	}

	if value >= target {
		return 100
	}

	if value <= 0 {
		return 0
	}

	return clamp(
		int(
			(value / target) *
				100,
		),
	)
}

func remaining(
	current int,
	target int,
) int {
	if current >= target {
		return 0
	}

	return target - current
}

func floatRemaining(current float64, target float64) float64 {
	if current >= target {
		return 0
	}

	value := target - current

	return float64(
		int(
			value*100,
		),
	) / 100
}

func average(
	values ...int,
) int {
	if len(values) == 0 {
		return 0
	}

	total := 0

	for _, value := range values {
		total += clamp(
			value,
		)
	}

	return clamp(
		total / len(values),
	)
}

func clamp(
	value int,
) int {
	if value < 0 {
		return 0
	}

	if value > 100 {
		return 100
	}

	return value
}

func enrichProgressionMetrics(
	reputation *reputationdomain.CleanerReputation,
) {
	if reputation == nil {
		return
	}

	if reputation.TotalBookings > 0 {
		reputation.CompletionRate =
			percentage(
				reputation.CompletedJobs,
				reputation.TotalBookings,
			)

		reputation.CancellationRate =
			percentage(
				reputation.CleanerCancellations,
				reputation.TotalBookings,
			)
	}

	if reputation.EligibleResponseMessages > 0 {
		reputation.ResponseRate =
			percentage(
				reputation.RespondedMessages,
				reputation.EligibleResponseMessages,
			)
	}
}
