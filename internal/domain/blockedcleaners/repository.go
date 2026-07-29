package blockedcleaners

import "context"

type Repository interface {
	Create(ctx context.Context, block *BlockedCleaner) error
	Delete(ctx context.Context, clientID uint, cleanerID uint) error
	ListByClientID(ctx context.Context, clientID uint) ([]BlockedCleaner, error)
	Exists(ctx context.Context, clientID uint, cleanerID uint) (bool, error)
}