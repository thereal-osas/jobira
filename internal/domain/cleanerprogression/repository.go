package cleanerprogression

import (
	"context"

	reputationdomain "github.com/rodrigueghenda/jobira/internal/domain/reputation"
)

type Repository interface {
	GetCleanerReputation(ctx context.Context, cleanerID uint,) (*reputationdomain.CleanerReputation, error)
}

