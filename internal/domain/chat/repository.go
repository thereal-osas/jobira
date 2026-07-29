package chat

import "context"

type Repository interface {
	CreateConversation(ctx context.Context, conversation *Conversation) error
	GetConversationByID(ctx context.Context, conversationID uint) (*Conversation, error)
	GetConversationByBookingID(ctx context.Context, bookingID uint) (*Conversation, error)
	ListConversationByUserID(ctx context.Context, userID uint) ([]ConversationSummary, error)
	CreateMessage(ctx context.Context, message *Message) error
	ListMessage(ctx context.Context, conversationID uint, limit int, offset int) ([]Message, error)
	MarkConversationAsRead(ctx context.Context, conversationID uint, userID uint) error
}
