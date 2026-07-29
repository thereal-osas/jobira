package preferredcleaners

import "context"

type Repository interface {
	Create(ctx context.Context, preferred *PreferredCleaners) error
	Delete(ctx context.Context, clientID uint, cleanerID uint) error
	ListByClientID(ctx context.Context, clientID uint) ([]PreferredCleaners, error)
	Exists(ctx context.Context, clientID uint, cleanerID uint) (bool, error)
}