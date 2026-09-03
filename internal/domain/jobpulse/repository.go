package jobpulse

import "context"

type Repository interface {
	GetSnapshot(
		ctx context.Context,
		jobID uint,
	) (*JobPulseSnapshot, error)
}
