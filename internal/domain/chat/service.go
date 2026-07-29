package chat

import (
	"context"
	"strings"

	"github.com/rodrigueghenda/jobira/internal/domain/bookings"
)

type BookingReader interface {
	GetByID(ctx context.Context, bookingID uint, userID uint, role string) (*bookings.Booking, error)
}

type Service struct {
	repo           Repository
	bookingService BookingReader
}

func NewService(repo Repository, bookingReader BookingReader) *Service {
	return &Service{
		repo:           repo,
		bookingService: bookingReader,
	}
}

func (s *Service) CreateConversation(ctx context.Context, userID uint, role string, req CreateConversationRequest) (*Conversation, error) {
	if userID == 0 || req.BookingID == 0 {
		return nil, ErrInvalidInput
	}

	booking, err := s.bookingService.GetByID(
		ctx,
		req.BookingID,
		userID,
		role,
	)
	if err != nil {
		return nil, err
	}

	if role != "admin" &&
		booking.ClientID != userID &&
		booking.CleanerID != userID {
		return nil, ErrForbidden
	}

	existing, err := s.repo.GetConversationByBookingID(
		ctx,
		req.BookingID,
	)
	if err == nil {
		return existing, nil
	}

	if err != ErrConversationNotFound {
		return nil, err
	}

	conversation := &Conversation{
		BookingID: req.BookingID,
		ClientID:  booking.ClientID,
		CleanerID: booking.CleanerID,
		Status:    "active",
	}

	if err := s.repo.CreateConversation(ctx, conversation); err != nil {
		return nil, err
	}

	return conversation, nil
}

func (s *Service) GetConversation(ctx context.Context, conversationID uint, userID uint, role string) (*Conversation, error) {
	if conversationID == 0 || userID == 0 {
		return nil, ErrInvalidInput
	}

	conversation, err := s.repo.GetConversationByID(
		ctx,
		conversationID,
	)
	if err != nil {
		return nil, err
	}

	if role != "admin" && conversation.ClientID != userID && conversation.CleanerID != userID {
		return nil, ErrForbidden
	}

	return conversation, nil
}

func (s *Service) ListConversations(ctx context.Context, userID uint) ([]ConversationSummary, error) {
	if userID == 0 {
		return nil, ErrInvalidInput
	}

	return s.repo.ListConversationByUserID(ctx, userID)
}

func (s *Service) SendMessage(ctx context.Context, conversationID uint, userID uint, role string, req SendMessageRequest) (*Message, error) {
	req.Content = strings.TrimSpace(req.Content)

	if conversationID == 0 || userID == 0 {
		return nil, ErrInvalidInput
	}

	if req.Content == "" {
		return nil, ErrInvalidMessage
	}

	if len(req.Content) > 4000 {
		return nil, ErrMessageTooLong
	}

	conversation, err := s.GetConversation(
		ctx,
		conversationID,
		userID,
		role,
	)
	if err != nil {
		return nil, err
	}

	if conversation.Status != "active" {
		return nil, ErrForbidden
	}

	message := &Message{
		ConversationID: conversationID,
		SenderID:       userID,
		MessageType:    "text",
		Content:        req.Content,
	}
	if err := s.repo.CreateMessage(ctx, message); err != nil {
		return nil, err
	}

	return message, nil
}

func (s *Service) ListMessages(ctx context.Context, ConversationID uint, userID uint, role string, limit int, offset int) ([]Message, error) {
	if limit <= 0 {
		limit = 50
	}

	if limit > 100 {
		limit = 100
	}

	if offset < 0 {
		offset = 0
	}

	_, err := s.GetConversation(
		ctx, ConversationID, userID, role,
	)
	if err != nil {
		return nil, err
	}

	messages, err := s.repo.ListMessage(
		ctx, ConversationID, limit, offset,
	)
	if err != nil {
		return nil, err
	}

	err = s.repo.MarkConversationAsRead(
		ctx, ConversationID, userID,
	)
	if err != nil {
		return nil, err
	}

	return messages, nil
}

func (s *Service) MarkAsRead(ctx context.Context, conversationID uint, userID uint, role string) error {
	_, err := s.GetConversation(
		ctx, conversationID, userID, role,
	)
	if err != nil {
		return err
	}

	return s.repo.MarkConversationAsRead(
		ctx, conversationID, userID,
	)
}
