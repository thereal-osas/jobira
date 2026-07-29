package chat

import "time"

type Conversation struct {
	ID             uint       `json:"id"`
	BookingID      uint       `json:"booking_id"`
	ClientID       uint       `json:"client_id"`
	CleanerID      uint       `json:"cleaner_id"`
	Status         string     `json:"status"`
	LastMessagesAt *time.Time `json:"last_message_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type Message struct {
	ID             uint       `json:"id"`
	ConversationID uint       `json:"conversation_id"`
	SenderID       uint       `json:"sender_id"`
	MessageType    string     `json:"message_type"`
	Content        string     `json:"content"`
	IsRead         bool       `json:"is_read"`
	ReadAt         *time.Time `json:"read_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type ConversationSummary struct {
	ID                uint       `json:"id"`
	BookingID         uint       `json:"booking_id"`
	ClientID          uint       `json:"client_id"`
	CleanerID         uint       `json:"cleaner_id"`
	Status            string     `json:"status"`
	OtherUserID       uint       `json:"other_user_id"`
	OtherUserName     string     `json:"other_user_name"`
	LastMessage       *string    `json:"last_message"`
	LastMessageSender *uint      `json:"last_message_sender"`
	LastMessageAt     *time.Time `json:"last_message_at"`
	UnreadCount       int        `json:"unread_count"`
	CreatedAt         time.Time  `json:"created_at"`
}
