package matchscore

import (
	"context"
	"errors"
	"math"
	"sort"
	"strings"
)

var (
	ErrInvalidInput = errors.New("invalid input")
	ErrForbidden    = errors.New("forbidden")
	ErrJobNotFound  = errors.New("job not found")
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

func (s *Service) RankJobCandidates(ctx context.Context, jobID uint, clientID uint) (*JobMatches, error) {
	if jobID == 0 || clientID == 0 {
		return nil, ErrInvalidInput
	}

	ownsJob, err := s.repo.JobBelongsToClient(
		ctx, jobID, clientID,
	)
	if err != nil {
		return nil, err
	}

	if !ownsJob {
		return nil, ErrForbidden
	}

	candidates, err := s.repo.ListCandidates(
		ctx, jobID, clientID,
	)
	if err != nil {
		return nil, err
	}

	matches := make(
		[]CleanerMatch,
		0,
		len(candidates),
	)

	for _, candidate := range candidates {
		match := calculateCandidateMatch(
			candidate,
		)

		matches = append(
			matches,
			match,
		)
	}

	sort.SliceStable(
		matches,
		func(i int, j int) bool {
			if matches[i].MatchScore ==
				matches[j].MatchScore {
				return matches[i].ReliabilityScore >
					matches[j].ReliabilityScore
			}

			return matches[i].MatchScore >
				matches[j].MatchScore
		},
	)

	for i := range matches {
		matches[i].Rank = i + 1

		matches[i].IsTopOffer =
			i == 0 &&
				matches[i].MatchScore >= 70

		matches[i].IsRecommended =
			i < 3 &&
				matches[i].MatchScore >= 70
	}

	result := &JobMatches{
		JobID:           jobID,
		TotalApplicants: len(matches),
		Matches:         matches,
	}

	if len(matches) > 0 &&
		matches[0].IsTopOffer {
		topOffer := matches[0]

		result.TopOffer = &topOffer
	}

	return result, nil
}

func calculateCandidateMatch(
	candidate CandidateData,
) CleanerMatch {
	breakdown := MatchScoreBreakdown{
		ServiceFit: calculateServiceFitScore(
			candidate.JobType,
			candidate.ServicesOffered,
		),

		Availability: calculateAvailabilityScore(
			candidate.AvailabilityStatus,
		),

		Reliability: calculateReliabilityPoints(
			candidate.ReliabilityScore,
		),

		Location: calculateLocationScore(
			candidate.JobLocation,
			candidate.CleanerLocation,
			candidate.CleanerPostcodeArea,
		),

		Verification: calculateVerificationScore(
			candidate.IsVerified,
		),

		Response: calculateResponseScore(
			candidate.ResponseRate,
			candidate.AverageResponseMinutes,
		),
	}

	total := breakdown.ServiceFit +
		breakdown.Availability +
		breakdown.Reliability +
		breakdown.Location +
		breakdown.Verification +
		breakdown.Response

	total = clampScore(
		total,
		0,
		100,
	)

	return CleanerMatch{
		ApplicationID: candidate.ApplicationID,
		JobID:         candidate.JobID,
		CleanerID:     candidate.CleanerID,

		CleanerName: candidate.CleanerName,

		MatchScore: total,
		MatchLabel: MatchLabel(
			total,
		),

		Breakdown: breakdown,

		WhyYoureSeeingThis: buildMatchReasons(
			breakdown,
		),

		AverageRating: candidate.AverageRating,
		TotalReviews:  candidate.TotalReviews,

		ReliabilityScore: candidate.ReliabilityScore,

		ResponseRate: candidate.ResponseRate,

		AverageResponseMinutes: candidate.AverageResponseMinutes,

		IsVerified: candidate.IsVerified,

		AvailabilityStatus: candidate.AvailabilityStatus,

		Badge: candidate.Badge,
	}
}

func buildMatchReasons(
	breakdown MatchScoreBreakdown,
) []MatchReason {
	reasons := make(
		[]MatchReason,
		0,
		6,
	)

	if breakdown.ServiceFit > 0 {
		message := "This cleaner offers services relevant to this job."

		if breakdown.ServiceFit >= 25 {
			message = "This cleaner offers the service required for this job."
		}

		reasons = append(
			reasons,
			MatchReason{
				Code:    "service_fit",
				Label:   "Service match",
				Message: message,
				Points:  breakdown.ServiceFit,
			},
		)
	}

	if breakdown.Availability > 0 {
		message := "This cleaner has some availability for work."

		if breakdown.Availability >= 25 {
			message = "This cleaner is currently available for work."
		}

		reasons = append(
			reasons,
			MatchReason{
				Code:    "availability",
				Label:   "Availability",
				Message: message,
				Points:  breakdown.Availability,
			},
		)
	}

	if breakdown.Location > 0 {
		message := "This cleaner works near the job location."

		if breakdown.Location >= 15 {
			message = "This cleaner matches the job location."
		}

		reasons = append(
			reasons,
			MatchReason{
				Code:    "location",
				Label:   "Location match",
				Message: message,
				Points:  breakdown.Location,
			},
		)
	}

	if breakdown.Reliability > 0 {
		reasons = append(
			reasons,
			MatchReason{
				Code:    "reliability",
				Label:   "Reliability",
				Message: "This cleaner's reliability history contributes to the match.",
				Points:  breakdown.Reliability,
			},
		)
	}

	if breakdown.Verification > 0 {
		reasons = append(
			reasons,
			MatchReason{
				Code:    "verification",
				Label:   "Verified profile",
				Message: "This cleaner has a verified profile.",
				Points:  breakdown.Verification,
			},
		)
	}

	if breakdown.Response > 0 {
		reasons = append(
			reasons,
			MatchReason{
				Code:    "response",
				Label:   "Responsive",
				Message: "This cleaner's response history contributes to the match.",
				Points:  breakdown.Response,
			},
		)
	}

	return reasons
}

func calculateServiceFitScore(jobType string, servicesOffered string) int {
	jobType = normalizeMatchText(
		jobType,
	)

	servicesOffered = normalizeMatchText(
		servicesOffered,
	)

	if jobType == "" || servicesOffered == "" {
		return 0
	}

	if strings.Contains(servicesOffered, jobType) {
		return 25
	}

	jobWords := strings.Fields(
		jobType,
	)

	if len(jobWords) == 0 {
		return 0
	}

	matchedWords := 0

	for _, word := range jobWords {
		if len(word) < 3 {
			continue
		}

		if strings.Contains(
			servicesOffered,
			word,
		) {
			matchedWords++
		}
	}

	if matchedWords == len(jobWords) {
		return 25
	}

	if matchedWords > 0 {
		return 15
	}

	return 0
}

func calculateAvailabilityScore(status string) int {
	status = normalizeMatchText(
		status,
	)

	switch status {
	case "available now":
		return 25

	case "availability today":
		return 25

	case "available":
		return 25

	case "today":
		return 25

	case "limited availability":
		return 15

	case "limited":
		return 15

	case "busy":
		return 5

	case "unavailability":
		return 0

	default:
		return 0
	}
}

func calculateReliabilityPoints(
	reliabilityScore int,
) int {
	reliabilityScore = clampScore(
		reliabilityScore,
		0,
		100,
	)

	points := (float64(reliabilityScore) / 100.0) * 20.0

	return int(
		math.Round(points),
	)
}

func calculateLocationScore(jobLocation string, cleanerLocation string, postcodeArea string) int {
	jobLocation = normalizeMatchText(
		jobLocation,
	)

	cleanerLocation = normalizeMatchText(
		cleanerLocation,
	)

	postcodeArea = normalizeMatchText(
		postcodeArea,
	)

	if jobLocation == "" {
		return 0
	}

	if cleanerLocation != "" &&
		jobLocation == cleanerLocation {
		return 15
	}

	if postcodeArea != "" &&
		strings.Contains(
			jobLocation,
			postcodeArea,
		) {
		return 15
	}

	if cleanerLocation != "" &&
		jobLocation == cleanerLocation {
		return 15
	}

	if postcodeArea != "" &&
		strings.Contains(
			jobLocation,
			postcodeArea,
		) {
		return 15
	}

	if cleanerLocation != "" &&
		(strings.Contains(jobLocation, cleanerLocation) ||
			strings.Contains(cleanerLocation, jobLocation)) {
		return 10
	}

	return 0

}

func calculateVerificationScore(
	verified bool,
) int {
	if verified {
		return 10
	}

	return 0
}

func calculateResponseScore(
	responseRate int,
	averageMinutes int,
) int {
	responseRate = clampScore(
		responseRate,
		0,
		100,
	)

	points := 0

	switch {
	case responseRate >= 90:
		points += 3

	case responseRate >= 75:
		points += 2

	case responseRate >= 50:
		points++
	}

	if averageMinutes > 0 {
		switch {
		case averageMinutes <= 15:
			points += 2

		case averageMinutes <= 60:
			points++
		}
	}

	return clampScore(
		points,
		0,
		5,
	)
}

func MatchLabel(score int) string {
	switch {
	case score >= 90:
		return "Excellent Match"

	case score >= 80:
		return "Strong Match"

	case score >= 70:
		return "Good Match"

	case score >= 50:
		return "Potential Match"

	default:
		return "Low Match"
	}
}

func normalizeMatchText(
	value string,
) string {
	value = strings.TrimSpace(
		strings.ToLower(value),
	)

	replacer := strings.NewReplacer(
		"_", " ",
		"-", " ",
		"/", " ",
		",", " ",
	)

	value = replacer.Replace(
		value,
	)

	return strings.Join(
		strings.Fields(value),
		" ",
	)
}

func clampScore(value int, minimum int, maximum int) int {
	if value < minimum {
		return minimum
	}

	if value > maximum {
		return maximum
	}

	return value
}
