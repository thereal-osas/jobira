package messages

import "context"

type Repository interface {
	Create(ctx context.Context, message *Message) error
	ListByJobID(ctx context.Context, jobID uint) ([]Message, error)
	IsJobParticipant(ctx context.Context, jobID uint, userID uint) (bool, error )
}