package topoffer

import "context"

type Repository interface {
	GetJobContext(ctx context.Context, jobID uint) (*JobOfferContext, error)
	ListOfferCandidates(ctx context.Context, jobID uint, limit int, offset int) ([]OfferCandidate, error)
	GetApplicationCandidate(ctx context.Context, applicationID uint) (*OfferCandidate, error)
	CountOffers(ctx context.Context, jobID uint) (int, error)
}
