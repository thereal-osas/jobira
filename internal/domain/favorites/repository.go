package favorites

import "context"

type Repository interface {
	Create(ctx context.Context, favourite *FavoriteCleaner) error
	Delete(ctx context.Context, clientID uint, cleanerID uint) error
	ListByClientID(ctx context.Context, clientID uint) ([]FavoriteCleaner, error)
	Exists(ctx context.Context, clientID uint, cleanerID uint) (bool, error)
}