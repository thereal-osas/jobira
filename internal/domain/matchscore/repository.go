package matchscore

import "context"

type Repository interface {
	ListCandidates(ctx context.Context, jobID uint, clientID uint) ([]CandidateData, error)
	JobBelongsToClient(ctx context.Context, jobID uint, clientID uint) (bool, error)
}
