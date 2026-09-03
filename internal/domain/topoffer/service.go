package topoffer

import (
	"context"
	"sort"
)

const (
	defaultOfferLimit = 50
	maxOfferLimit     = 100
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

func (s *Service) RankJobOffers(ctx context.Context, jobID uint, clientID uint, req ListTopOffersRequest) (*TopOfferResult, error) {
	if jobID == 0 || clientID == 0 {
		return nil, ErrInvalidInput
	}

	job, err := s.repo.GetJobContext(
		ctx,
		jobID,
	)
	if err != nil {
		return nil, err
	}

	if job == nil {
		return nil, ErrJobNotFound
	}

	if job.ClientID != clientID {
		return nil, ErrForbidden
	}

	if job.Status != "open" {
		return nil, ErrJobClosed
	}

	if req.Limit <= 0 {
		req.Limit = defaultOfferLimit
	}

	if req.Limit > maxOfferLimit {
		req.Limit = maxOfferLimit
	}

	if req.Offset < 0 {
		return nil, ErrInvalidInput
	}

	candidates, err := s.repo.ListOfferCandidates(
		ctx,
		jobID,
		req.Limit,
		req.Offset,
	)
	if err != nil {
		return nil, err
	}

	TotalOffers, err := s.repo.CountOffers(
		ctx,
		jobID,
	)

	if err != nil {
		return nil, err
	}

	offers := make(
		[]RankedOffer,
		0,
		len(candidates),
	)

	for _, candidate := range candidates {
		offers =
			append(
				offers,
				calculateOfferScore(
					job,
					candidate,
				),
			)
	}

	sort.SliceStable(
		offers,
		func(i, j int) bool {
			if offers[i].TopOfferScore ==
				offers[j].TopOfferScore {
				return offers[i].AppliedAt.Before(
					offers[j].AppliedAt,
				)
			}

			return offers[i].TopOfferScore >
				offers[j].TopOfferScore
		},
	)

	var TopOfferApplicationID *uint

	for index := range offers {
		offers[index].Rank =
			index + 1

		offers[index].IsTopOffer =
			index == 0

		if index == 0 {
			id :=
				offers[index].
					ApplicationID

			TopOfferApplicationID =
				&id
		}
	}

	return &TopOfferResult{
		JobID: jobID,

		TotalOffers: TotalOffers,

		TopOfferApplicationID: TopOfferApplicationID,

		Offers: offers,
	}, nil
}

func (s *Service) GetApplicationRanking(ctx context.Context, applicationID uint, cleanerID uint) (*TopOfferSummary, error) {
	if applicationID == 0 || cleanerID == 0 {
		return nil, ErrInvalidInput
	}

	candidate, err := s.repo.GetApplicationCandidate(
		ctx,
		applicationID,
	)
	if err != nil {
		return nil, err
	}

	if candidate == nil {
		return nil, ErrApplicationNotFound
	}

	if candidate.CleanerID != cleanerID {
		return nil, ErrForbidden
	}

	job, err :=
		s.repo.GetJobContext(
			ctx,
			candidate.JobID,
		)

	if err != nil {
		return nil, err
	}

	candidates, err :=
		s.repo.ListOfferCandidates(
			ctx,
			candidate.JobID,
			maxOfferLimit,
			0,
		)
	if err != nil {
		return nil, err
	}

	offers :=
		make(
			[]RankedOffer,
			0,
			len(candidates),
		)
	for _, item := range candidates {
		offers =
			append(
				offers,
				calculateOfferScore(
					job,
					item,
				),
			)
	}

	sort.SliceStable(
		offers,
		func(i, j int) bool {
			if offers[i].TopOfferScore ==
				offers[j].TopOfferScore {
				return offers[i].AppliedAt.Before(
					offers[j].AppliedAt,
				)
			}

			return offers[i].TopOfferScore >
				offers[j].TopOfferScore
		},
	)

	for index, offer := range offers {
		if offer.ApplicationID !=
			applicationID {
			continue
		}

		return &TopOfferSummary{
			ApplicationID: applicationID,

			Rank: index + 1,

			Score: offer.TopOfferScore,

			IsTopOffer: index == 0,
			Reason:     offer.Reasons,
		}, nil
	}

	return nil,
		ErrApplicationNotFound
}

func calculateOfferScore(job *JobOfferContext, candidate OfferCandidate) RankedOffer {
	priceScore := calculatePriceFitScore(
		job.Budget,
		candidate.ProposedRate,
	)

	ratingScore :=
		calculateRatingScore(
			candidate.AverageRating,
		)

	reliabilityScore :=
		clampScore(
			candidate.ReliabilityScore,
		)

	experienceScore :=
		calculateExperienceScore(
			candidate.CompletedJobs,
		)

	recommendationScore :=
		clampScore(
			candidate.RecommendationPercentage,
		)

	responseScore :=
		calculateResponseScore(
			candidate.AverageResponseMinutes,
		)

	verificationsScore := 0

	if candidate.IsVerified {
		verificationsScore = 100
	}

	breakdown :=
		ScoreBreakdown{
			PriceFitScore: priceScore,

			RatingScore: ratingScore,

			ReliabilityScore: reliabilityScore,

			ExperienceScore: experienceScore,

			RecommendationScore: recommendationScore,

			ResponseScore: responseScore,

			VerificationScore: verificationsScore,
		}

	total :=
		(priceScore*20)/100 +
			(ratingScore*20)/100 +
			(reliabilityScore*20)/100 +
			(experienceScore*15)/100 +
			(recommendationScore*10)/100 +
			(responseScore*10)/100 +
			(verificationsScore*5)/100

	reason :=
		buildReason(
			candidate,
			breakdown,
		)

	return RankedOffer{
		ApplicationID: candidate.ApplicationID,

		JobID: candidate.JobID,

		CleanerID: candidate.CleanerID,

		CleanerName: candidate.CleanerName,

		CoverMessage: candidate.CoverMessage,

		ProposedRate: candidate.ProposedRate,

		ApplicationStatus: candidate.ApplicationStatus,

		AverageRating: candidate.AverageRating,

		TotalReviews: candidate.TotalReviews,

		CompletedJobs: candidate.CompletedJobs,

		ReliabilityScore: candidate.ReliabilityScore,

		RecommendationPercentage: candidate.RecommendationPercentage,

		AverageResponseMinutes: candidate.AverageResponseMinutes,

		Badge: candidate.Badge,

		IsVerified: candidate.IsVerified,

		DBSVerified: candidate.DBSVerified,

		TopOfferScore: clampScore(
			total,
		),

		ScoreBreakdown: breakdown,

		Reasons: reason,

		AppliedAt: candidate.AppliedAt,
	}
}

func calculatePriceFitScore(
	budget int,
	proposedRate int,
) int {
	if budget <= 0 || proposedRate <= 0 {
		return 50
	}

	if proposedRate <= budget {
		difference :=
			budget -
				proposedRate

		percentageBelow :=
			(difference * 100) /
				budget

		if percentageBelow <= 10 {
			return 100
		}

		if percentageBelow <= 25 {
			return 90
		}

		if percentageBelow <= 50 {
			return 75
		}

		return 60
	}

	difference :=
		proposedRate -
			budget

	percentageOver :=
		(difference * 100) /
			budget

	switch {
	case percentageOver <= 10:
		return 85

	case percentageOver <= 20:
		return 70

	case percentageOver <= 35:
		return 50

	default:
		return 25
	}
}

func calculateRatingScore(
	rating float64,
) int {
	if rating <= 0 {
		return 0
	}

	score :=
		int(
			(rating / 5.0) *
				100,
		)

	return clampScore(
		score,
	)
}

func calculateExperienceScore(
	CompletedJobs int,
) int {
	switch {
	case CompletedJobs >= 100:
		return 100

	case CompletedJobs >= 50:
		return 90

	case CompletedJobs >= 25:
		return 80

	case CompletedJobs >= 10:
		return 65

	case CompletedJobs >= 5:
		return 50

	case CompletedJobs >= 1:
		return 10

	default:
		return 10
	}
}

func calculateResponseScore(
	minutes int,
) int {
	switch {
	case minutes <= 0:
		return 50

	case minutes <= 15:
		return 100

	case minutes <= 30:
		return 90

	case minutes <= 60:
		return 80

	case minutes <= 180:
		return 65

	case minutes <= 360:
		return 50

	default:
		return 30
	}
}

func buildReason(
	candidate OfferCandidate,
	breakdown ScoreBreakdown,
) []string {
	reasons := make([]string, 0, 6)

	if breakdown.PriceFitScore >= 90 {
		reasons =
			append(
				reasons,
				"Strong price fit",
			)
	}

	if candidate.AverageRating >= 4.8 {
		reasons =
			append(
				reasons,
				"Highly reliable",
			)
	}

	if candidate.CompletedJobs >= 25 {
		reasons =
			append(
				reasons,
				"Strong Jobira experience",
			)
	}

	if candidate.RecommendationPercentage >= 90 {
		reasons =
			append(
				reasons,
				"Highly recommended",
			)
	}

	if candidate.AverageResponseMinutes > 0 &&
		candidate.AverageResponseMinutes <= 30 {
		reasons =
			append(
				reasons,
				"Usually responds quickly",
			)
	}

	if candidate.IsVerified {
		reasons =
			append(
				reasons,
				"Identity verified",
			)
	}

	if candidate.DBSVerified {
		reasons =
			append(
				reasons,
				"DBS verified",
			)
	}

	if len(reasons) == 0 {
		reasons =
			append(
				reasons,
				"Active Jobira applicant",
			)
	}

	return reasons
}

func clampScore(
	score int,
) int {
	if score < 0 {
		return 0
	}

	if score > 100 {
		return 100
	}

	return score
}
