package jobinvitations

import "context"

type Repository interface {
	Create(ctx context.Context, invitation *JobInvitation) error
	Exists(ctx context.Context, jobID uint, cleanerID uint) (bool, error)
	GetJobClientID(ctx context.Context, jobID uint) (uint, error)
	ListSentByClientID(ctx context.Context, clientID uint) ([]JobInvitation, error)
	ListReceivedByCleanerID(ctx context.Context, cleanerID uint) ([]JobInvitation, error)
}
