package clientnotes

import "context"

type Repository interface {
	Create(ctx context.Context, note *ClientCleanerNote) error
	GetByCleanerID(ctx context.Context, clientID uint, cleanerID uint) ([]ClientCleanerNote, error)
	Update(ctx context.Context, noteID uint, clientID uint, note string) error
	Delete(ctx context.Context, noteID uint, clientID uint) error
}