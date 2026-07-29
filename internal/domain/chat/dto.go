package chat

type CreateConversationRequest struct {
	BookingID uint `json:"booking_id"`
}

type SendMessageRequest struct {
	Content string `json:"content"`
}

type MessageResponse struct {
	Message Message `json:"message"`
}

type ConversationResponse struct {
	Conversation Conversation `json:"conversation"`
}

type ConversationListResponse struct {
	Conversations []ConversationSummary `json:"conversations"`
}

type MessageListResponse struct {
	Messages []Message `json:"messages"`
}
