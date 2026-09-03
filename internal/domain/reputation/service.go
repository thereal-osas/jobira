package reputation

import (
	"context"
	"math"
)

var jobMilestones = []struct {
	Jobs        int
	Code        string
	Name        string
	Description string
}{
	{1, "first_job", "First Job", "Completed the first Jobira job"},
	{5, "getting_started", "Getting started", "Completed 5 Jobira jobs"},
	{10, "established_cleaner", "Established Cleaner", "Completed 10 Jobira jobs"},
	{15, "rising_pro", "Rising Pro", "Completed 15 Jobira jobs"},
	{25, "experienced_cleaner", "Experience Cleaner", "Completed 25 Jobira jobs"},
	{50, "proven_professional", "Proven Professional", "Completed 50 Jobira jobs"},
	{100, "century_cleaner", "Century Cleaner", "Completed 100 Jobira jobs"},
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetByCleanerID(ctx context.Context, cleanerID uint) (*CleanerReputation, error) {
	if cleanerID == 0 {
		return nil, ErrInvalidInput
	}

	if err := s.repo.Refresh(
		ctx,
		cleanerID,
	); err != nil {
		return nil, err
	}

	reputation, err := s.repo.GetByCleanerID(
		ctx,
		cleanerID,
	)
	if err != nil {
		return nil, err
	}

	if reputation == nil {
		return nil, ErrReputationNotFound
	}

	enrichReputation(reputation)

	return reputation, nil
}

func (s *Service) Refresh(ctx context.Context, cleanerID uint) (*CleanerReputation, error) {
	if cleanerID == 0 {
		return nil, ErrInvalidInput
	}

	if err := s.repo.Refresh(ctx, cleanerID); err != nil {
		return nil, err
	}

	reputation, err := s.repo.GetByCleanerID(ctx, cleanerID)
	if err != nil {
		return nil, err
	}

	if reputation == nil {
		return nil, ErrReputationNotFound
	}

	enrichReputation(reputation)

	if err := s.repo.UpdateBadge(ctx, cleanerID, reputation.Badge); err != nil {
		return nil, err
	}
	return reputation, nil
}

func enrichReputation(reputation *CleanerReputation) {
	if reputation == nil {
		return
	}

	reputation.CompletionRate = calculateCompletionRate(
		reputation,
	)

	reputation.CancellationRate = calculateCancellationRate(
		reputation,
	)

	reputation.ResponseRate = calculateResponseRate(
		reputation,
	)

	reputation.ReliabilityScore = calculateReliabilityScore(
		reputation,
	)

	reputation.ReliabilityStatus = calculateReliabilityStatus(
		reputation,
	)

	reputation.Badge = calculateBadge(
		reputation,
	)

	reputation.Achievements = calculateAchievements(
		reputation,
	)

	setMilestoneProgress(reputation)
}
func calculateCompletionRate(
	reputation *CleanerReputation,
) int {
	if reputation == nil {
		return 0
	}

	reliabilityBookings :=
		reputation.CompletedJobs +
			reputation.CleanerCancellations

	if reliabilityBookings <= 0 {
		return 0
	}

	return clampPercentage(
		percentage(
			reputation.CompletedJobs,
			reliabilityBookings,
		),
	)
}

func calculateCancellationRate(
	reputation *CleanerReputation,
) int {
	if reputation == nil {
		return 0
	}

	reliabilityBookings :=
		reputation.CompletedJobs +
			reputation.CleanerCancellations

	if reliabilityBookings <= 0 {
		return 0
	}

	return clampPercentage(
		percentage(
			reputation.CleanerCancellations,
			reliabilityBookings,
		),
	)
}

func calculateResponseRate(
	reputation *CleanerReputation,
) int {
	if reputation == nil {
		return 0
	}

	if reputation.EligibleResponseMessages <= 0 {
		return 0
	}

	rate := percentage(
		reputation.RespondedMessages,
		reputation.EligibleResponseMessages,
	)

	return clampPercentage(rate)
}

func calculateReliabilityScore(
	reputation *CleanerReputation,
) int {
	if reputation == nil {
		return 0
	}

	// No booking history yet means there is not enough evidence
	// for a reliability score.
	reliabilityBookings :=
		reputation.CompletedJobs +
			reputation.CleanerCancellations

	if reliabilityBookings <= 0 {
		return 0
	}
	completionScore := float64(
		reputation.CompletionRate,
	) * 0.25

	ratingScore := 0.0

	if reputation.TotalReviews > 0 {
		ratingPercentage := (reputation.AverageRating / 5.0) * 100
		ratingScore = ratingPercentage * 0.20
	}

	recommendationScore := float64(
		reputation.RecommendationPercentage,
	) * 0.20

	cancellationReliability := 100 -
		reputation.CancellationRate

	cancellationScore := float64(
		clampPercentage(
			cancellationReliability,
		),
	) * 0.15

	repeatClientRate := 0

	if reputation.CompletedJobs > 0 {
		repeatClientRate = percentage(
			reputation.RepeatClients,
			reputation.CompletedJobs,
		)
	}

	repeatClientScore := float64(
		clampPercentage(
			repeatClientRate,
		),
	) * 0.10

	responseScore := 0.0

	if reputation.EligibleResponseMessages > 0 {
		responseScore = float64(
			reputation.ResponseRate,
		) * 0.10
	}

	rawScore := completionScore +
		ratingScore +
		recommendationScore +
		cancellationScore +
		repeatClientScore +
		responseScore

		/*
			Confidence adjustment.

			A cleaner with one successful booking should not
			immediately look equivalent to somebody with 100.

			1-4 bookings:
				score is still developing.

			5+ bookings:
				full calculated score is used.
		*/
	if reliabilityBookings < 5 {
		confidence := 0.60 + (float64(reliabilityBookings) * 0.08)
		rawScore *= confidence
	}

	score := int(
		math.Round(rawScore),
	)

	return clampPercentage(score)
}

func calculateReliabilityStatus(
	reputation *CleanerReputation,
) string {
	if reputation == nil {
		return "New"
	}

	if reputation.CompletedJobs == 0 {
		return "New"
	}

	if reputation.CompletedJobs < 5 {
		return "Building Reliability"
	}

	if reputation.ReliabilityScore >= 90 {
		return "Highly Reliable"
	}

	if reputation.ReliabilityScore >= 80 {
		return "Reliable"
	}

	if reputation.ReliabilityScore >= 70 {
		return "Established Reliability"
	}

	return "Building Reliability"
}
func calculateBadge(
	reputation *CleanerReputation,
) string {
	if reputation == nil {
		return "New Cleaner"
	}

	if reputation.TotalReviews >= 20 &&
		reputation.AverageRating >= 4.9 &&
		reputation.RecommendationPercentage >= 95 &&
		reputation.ReliabilityScore >= 90 {
		return "Elite Cleaner"
	}

	if reputation.TotalReviews >= 10 &&
		reputation.AverageRating >= 4.8 &&
		reputation.RecommendationPercentage >= 90 &&
		reputation.ReliabilityScore >= 80 {
		return "Top Rated"
	}

	if reputation.CompletedJobs >= 50 {
		return "Experienced Cleaner"
	}

	if reputation.CompletedJobs >= 1 {
		return "Active Cleaner"
	}

	return "New Cleaner"
}

func calculateAchievements(reputation *CleanerReputation) []Achievement {
	if reputation == nil {
		return []Achievement{}
	}

	achievements := make(
		[]Achievement,
		0,
		len(jobMilestones)+6,
	)

	for _, milestone := range jobMilestones {
		achievements = append(
			achievements,
			Achievement{
				Code:        milestone.Code,
				Name:        milestone.Name,
				Description: milestone.Description,
				Category:    "jobs",
				Earned:      reputation.CompletedJobs >= milestone.Jobs,
				Current:     reputation.CompletedJobs,
				Target:      milestone.Jobs,
			},
		)
	}

	achievements = append(
		achievements,
		calculateHighlyReliableAchievement(
			reputation,
		),
		calculateHighlyRecommendedAchievement(
			reputation,
		),
		calculateRepeatClientAchievement(
			reputation,
		),
		calculateQuickResponderAchievement(
			reputation,
		),
		calculateFastResponderAchievement(
			reputation,
		),
		calculateRapidResponderAchievement(
			reputation,
		),
	)

	return achievements
}

func calculateHighlyReliableAchievement(
	reputation *CleanerReputation,
) Achievement {
	return Achievement{
		Code:        "highly_reliable",
		Name:        "Highly_Reliable",
		Description: "Maintained excellent reliability across Jobira bookings",
		Category:    "reliabilty",
		Earned: reputation.CompletedJobs >= 10 &&
			reputation.ReliabilityScore >= 90,
		Current: reputation.ReliabilityScore,
		Target:  90,
	}
}

func calculateHighlyRecommendedAchievement(
	reputation *CleanerReputation,
) Achievement {
	return Achievement{
		Code:        "highly_recommended",
		Name:        "Highly_Recommended",
		Description: "Consistently recommended by Jobira clients",
		Category:    "performance",
		Earned: reputation.TotalReviews >= 10 &&
			reputation.RecommendationPercentage >= 90,
		Current: reputation.RecommendationPercentage,
		Target:  90,
	}
}

func calculateRepeatClientAchievement(
	reputation *CleanerReputation,
) Achievement {
	return Achievement{
		Code:        "repeat_client_favourite",
		Name:        "Repeat Client Favourite",
		Description: "Clients regularly choose to book this cleaner again",
		Category:    "performance",
		Earned: reputation.CompletedJobs >= 10 &&
			reputation.RepeatClients >= 3,
		Current: reputation.RepeatClients,
		Target:  3,
	}
}

func calculateQuickResponderAchievement(
	reputation *CleanerReputation,
) Achievement {
	return Achievement{
		Code:        "quick_responder",
		Name:        "Quick Responder",
		Description: "Usually responds to clients within 60 minutes",
		Category:    "response",
		Earned: reputation.EligibleResponseMessages >= 10 &&
			reputation.ResponseRate >= 80 &&
			reputation.AverageResponseMinutes > 0 &&
			reputation.AverageResponseMinutes <= 60,
		Current: reputation.EligibleResponseMessages,
		Target:  10,
	}
}

func calculateFastResponderAchievement(
	reputation *CleanerReputation,
) Achievement {
	return Achievement{
		Code:        "fast_responder",
		Name:        "Fast Responder",
		Description: "Usually responds to clients within 30 minutes",
		Category:    "response",
		Earned: reputation.EligibleResponseMessages >= 20 &&
			reputation.ResponseRate >= 85 &&
			reputation.AverageResponseMinutes > 0 &&
			reputation.AverageResponseMinutes <= 30,
		Current: reputation.EligibleResponseMessages,
		Target:  20,
	}
}

func calculateRapidResponderAchievement(
	reputation *CleanerReputation,
) Achievement {
	return Achievement{
		Code:        "rapid_responder",
		Name:        "Rapid Responder",
		Description: "Usually responds to clients within 10 minutes",
		Category:    "response",
		Earned: reputation.EligibleResponseMessages >= 30 &&
			reputation.ResponseRate >= 90 &&
			reputation.AverageResponseMinutes > 0 &&
			reputation.AverageResponseMinutes <= 10,
		Current: reputation.EligibleResponseMessages,
		Target:  30,
	}
}

func setMilestoneProgress(
	reputation *CleanerReputation,
) {
	if reputation == nil {
		return
	}

	completed := reputation.CompletedJobs

	reputation.CurrentMilestone = 0
	reputation.NextMilestone = 1
	reputation.JobsUntilNextMilestone = 1
	reputation.MilestoneProgress = 0

	previousMilestone := 0

	for _, milestone := range jobMilestones {
		if completed >= milestone.Jobs {
			reputation.CurrentMilestone = milestone.Jobs
			previousMilestone = milestone.Jobs
			continue
		}

		reputation.NextMilestone = milestone.Jobs
		reputation.JobsUntilNextMilestone = milestone.Jobs - completed

		rangeSize := milestone.Jobs - previousMilestone
		progressIntoRange := completed - previousMilestone

		if rangeSize > 0 {
			reputation.MilestoneProgress = clampPercentage(
				int(
					math.Round(
						(float64(progressIntoRange) /
							float64(rangeSize)) * 100,
					),
				),
			)
		}

		return
	}

	//Cleaner has reached the final current milestone.
	lastMilestone := jobMilestones[len(jobMilestones)-1].Jobs

	reputation.CurrentMilestone = lastMilestone
	reputation.NextMilestone = 0
	reputation.JobsUntilNextMilestone = 0
	reputation.MilestoneProgress = 100
}
func percentage(
	value int,
	total int,
) int {
	if total <= 0 {
		return 0
	}

	return int(
		math.Round(
			(float64(value) / float64(total)) * 100,
		),
	)
}

func clampPercentage(
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
