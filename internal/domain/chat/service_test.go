package chat

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/rodrigueghenda/jobira/internal/domain/bookings"
)

func stringPtr(value string) *string {
	return &value
}

type MockRepository struct {
	conversation  *Conversation
	conversations []ConversationSummary
	messages      []Message
	message       *Message

	createConversationErr error
	getConversationErr    error
	createMessageErr      error
	listMessageErr        error
	markReadErr           error
	listConversationErr   error
}

type mockBookingReader struct {
	booking *bookings.Booking
	err     error
}

func (m *MockRepository) CreateConversation(ctx context.Context, conversation *Conversation) error {
	if m.createConversationErr != nil {
		return m.createConversationErr
	}

	m.conversation = conversation
	return nil
}

func (m *MockRepository) GetConversationByID(ctx context.Context, conversationID uint) (*Conversation, error) {
	if m.getConversationErr != nil {
		return nil, m.getConversationErr
	}

	return m.conversation, nil
}

func (m *MockRepository) GetConversationByBookingID(ctx context.Context, bookingID uint) (*Conversation, error) {
	if m.getConversationErr != nil {
		return nil, m.getConversationErr
	}

	return m.conversation, nil
}

func (m *MockRepository) ListConversationByUserID(ctx context.Context, userID uint) ([]ConversationSummary, error) {
	if m.listConversationErr != nil {
		return nil, m.listConversationErr
	}

	return m.conversations, nil
}

func (m *MockRepository) CreateMessage(ctx context.Context, message *Message) error {
	if m.createMessageErr != nil {
		return m.createMessageErr
	}

	m.message = message
	return nil
}

func (m *MockRepository) ListMessage(ctx context.Context, conversationID uint, limit int, offset int) ([]Message, error) {
	if m.listMessageErr != nil {
		return nil, m.listMessageErr
	}

	return m.messages, nil
}

func (m *MockRepository) MarkConversationAsRead(ctx context.Context, conversationID uint, userID uint) error {
	return m.markReadErr
}

func (m *MockRepository) MarkAsRead(ctx context.Context, conversationID uint, userID uint) error {
	return m.markReadErr
}

func TestCreateConversation_InvalidInput(t *testing.T) {
	tests := []struct {
		name      string
		userID    uint
		bookingID uint
	}{
		{
			name:      "invalid user ID",
			userID:    0,
			bookingID: 1,
		},
		{
			name:      "invalid booking ID",
			userID:    10,
			bookingID: 0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := &MockRepository{}

			bookingReader := &mockBookingReader{}

			service := NewService(
				repo,
				bookingReader,
			)

			conversation, err := service.CreateConversation(
				context.Background(),
				test.userID,
				"client",
				CreateConversationRequest{
					BookingID: test.bookingID,
				},
			)

			if conversation != nil {
				t.Fatal("expected conversation to be nil")
			}

			if !errors.Is(err, ErrInvalidInput) {
				t.Fatalf(
					"expected ErrInvalidInput, got %v",
					err,
				)
			}
		})
	}
}

func TestCreateConversation_BookingReaderError(t *testing.T) {
	bookingErr := errors.New("booking lookup failed")

	repo := &MockRepository{}

	bookingReader := &mockBookingReader{
		err: bookingErr,
	}

	service := NewService(
		repo,
		bookingReader,
	)

	conversation, err := service.CreateConversation(
		context.Background(),
		10,
		"client",
		CreateConversationRequest{
			BookingID: 5,
		},
	)

	if conversation != nil {
		t.Fatal("expected conversation to be nil")
	}

	if !errors.Is(err, bookingErr) {
		t.Fatalf(
			"expected booking error, got %v",
			err,
		)
	}
}

func TestCreateConversation_NonParticipantForbidden(t *testing.T) {
	repo := &MockRepository{}

	bookingReader := &mockBookingReader{
		booking: &bookings.Booking{
			ID:        5,
			ClientID:  10,
			CleanerID: 20,
		},
	}

	service := NewService(
		repo,
		bookingReader,
	)

	conversation, err := service.CreateConversation(
		context.Background(),
		99,
		"client",
		CreateConversationRequest{
			BookingID: 5,
		},
	)

	if conversation != nil {
		t.Fatal("expected conversation to be nil")
	}

	if !errors.Is(err, ErrForbidden) {
		t.Fatalf(
			"expected ErrForbidden, got %v",
			err,
		)
	}
}

func TestCreateConversation_ReturnsExistingConversation(t *testing.T) {
	existingConversation := &Conversation{
		ID:        7,
		BookingID: 5,
		ClientID:  10,
		CleanerID: 20,
		Status:    "active",
	}

	repo := &MockRepository{
		conversation: existingConversation,
	}

	bookingReader := &mockBookingReader{
		booking: &bookings.Booking{
			ID:        5,
			ClientID:  10,
			CleanerID: 20,
		},
	}

	service := NewService(
		repo,
		bookingReader,
	)

	conversation, err := service.CreateConversation(
		context.Background(),
		10,
		"client",
		CreateConversationRequest{
			BookingID: 5,
		},
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if conversation == nil {
		t.Fatal("expected conversation")
	}

	if conversation.ID != 7 {
		t.Fatalf(
			"expected conversation ID 7, got %d",
			conversation.ID,
		)
	}

	if conversation != existingConversation {
		t.Fatal("expected existing conversation to be returned")
	}
}

func TestCreateConversation_ConversationLookupError(t *testing.T) {
	lookupErr := errors.New("conversation lookup failed")

	repo := &MockRepository{
		getConversationErr: lookupErr,
	}

	bookingReader := &mockBookingReader{
		booking: &bookings.Booking{
			ID:        5,
			ClientID:  10,
			CleanerID: 20,
		},
	}

	service := NewService(
		repo,
		bookingReader,
	)

	conversation, err := service.CreateConversation(
		context.Background(),
		10,
		"client",
		CreateConversationRequest{
			BookingID: 5,
		},
	)

	if conversation != nil {
		t.Fatal("expected conversation to be nil")
	}

	if !errors.Is(err, lookupErr) {
		t.Fatalf(
			"expected conversation lookup error, got %v",
			err,
		)
	}
}

func TestCreateConversation_Success(t *testing.T) {
	repo := &MockRepository{
		getConversationErr: ErrConversationNotFound,
	}

	bookingReader := &mockBookingReader{
		booking: &bookings.Booking{
			ID:        5,
			ClientID:  10,
			CleanerID: 20,
		},
	}

	service := NewService(
		repo,
		bookingReader,
	)

	conversation, err := service.CreateConversation(
		context.Background(),
		10,
		"client",
		CreateConversationRequest{
			BookingID: 5,
		},
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if conversation == nil {
		t.Fatal("expected conversation")
	}

	if conversation.BookingID != 5 {
		t.Fatalf(
			"expected booking ID 5, got %d",
			conversation.BookingID,
		)
	}

	if conversation.Status != "active" {
		t.Fatalf(
			"expected active status, got %s",
			conversation.Status,
		)
	}
}

func TestCreateConversation_CreateRepositoryError(t *testing.T) {
	createErr := errors.New("create conversation failed")

	repo := &MockRepository{
		getConversationErr:    ErrConversationNotFound,
		createConversationErr: createErr,
	}

	bookingReader := &mockBookingReader{
		booking: &bookings.Booking{
			ID:        5,
			ClientID:  10,
			CleanerID: 20,
		},
	}

	service := NewService(
		repo,
		bookingReader,
	)

	conversation, err := service.CreateConversation(
		context.Background(),
		10,
		"client",
		CreateConversationRequest{
			BookingID: 5,
		},
	)

	if conversation != nil {
		t.Fatal("expected conversation to be nil")
	}

	if !errors.Is(err, createErr) {
		t.Fatalf(
			"expected create error, got %v",
			err,
		)
	}
}

func TestSendMessage_EmptyMessage(t *testing.T) {
	repo := &MockRepository{
		conversation: &Conversation{
			ID:        1,
			ClientID:  10,
			CleanerID: 11,
			Status:    "active",
		},
	}

	service := NewService(repo, &mockBookingReader{})

	_, err := service.SendMessage(
		context.Background(),
		1,
		10,
		"client",
		SendMessageRequest{
			Content: "",
		},
	)

	if !errors.Is(err, ErrInvalidMessage) {
		t.Fatalf("expected ErrInvalidMessage, got %v", err)
	}
}

func TestSendMessage_WhitespaceOnly(t *testing.T) {
	repo := &MockRepository{
		conversation: &Conversation{
			ID:        1,
			ClientID:  10,
			CleanerID: 11,
			Status:    "active",
		},
	}

	service := NewService(repo, &mockBookingReader{})

	_, err := service.SendMessage(
		context.Background(),
		1,
		10,
		"client",
		SendMessageRequest{
			Content: "      ",
		},
	)

	if !errors.Is(err, ErrInvalidMessage) {
		t.Fatalf("expected ErrInvalidMessage, got %v", err)
	}
}

func (m *mockBookingReader) GetByID(ctx context.Context, bookingID uint, userID uint, role string) (*bookings.Booking, error) {
	if m.err != nil {
		return nil, m.err
	}

	return m.booking, nil
}

func TestSendMessage_MessageTooLong(t *testing.T) {
	repo := &MockRepository{
		conversation: &Conversation{
			ID:        1,
			ClientID:  10,
			CleanerID: 11,
			Status:    "active",
		},
	}

	service := NewService(repo, &mockBookingReader{})

	longMessage := strings.Repeat("A", 4001)

	_, err := service.SendMessage(
		context.Background(),
		1,
		10,
		"client",
		SendMessageRequest{
			Content: longMessage,
		},
	)

	if !errors.Is(err, ErrMessageTooLong) {
		t.Fatalf("expected ErrMessageTooLong, got %v", err)
	}
}

func TestSendMessage_NonParticipantForbidden(t *testing.T) {
	repo := &MockRepository{
		conversation: &Conversation{
			ID:        1,
			ClientID:  10,
			CleanerID: 11,
			Status:    "active",
		},
	}

	service := NewService(repo, &mockBookingReader{})

	_, err := service.SendMessage(
		context.Background(),
		1,
		99,
		"user",
		SendMessageRequest{
			Content: "Hello",
		},
	)

	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestSendMessage_InactiveConversationForbidden(t *testing.T) {
	repo := &MockRepository{
		conversation: &Conversation{
			ID:        1,
			ClientID:  10,
			CleanerID: 11,
			Status:    "closed",
		},
	}

	service := NewService(repo, &mockBookingReader{})

	_, err := service.SendMessage(
		context.Background(),
		1,
		10,
		"client",
		SendMessageRequest{
			Content: "Hello",
		},
	)

	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestSendMessage_Success(t *testing.T) {
	repo := &MockRepository{
		conversation: &Conversation{
			ID:        1,
			ClientID:  10,
			CleanerID: 11,
			Status:    "active",
		},
	}

	service := NewService(repo, &mockBookingReader{})

	message, err := service.SendMessage(
		context.Background(),
		1,
		10,
		"client",
		SendMessageRequest{
			Content: "Hello cleaner!",
		},
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if message == nil {
		t.Fatal("expected message to be returned")
	}

	if repo.message == nil {
		t.Fatal("expected repository CreateMessage to be called")
	}

	if repo.message.Content != "Hello cleaner!" {
		t.Fatalf("expected content 'Hello cleaner!', got %q", repo.message.Content)
	}

	if repo.message.SenderID != 10 {
		t.Fatalf("expected sender ID 10, got %d", repo.message.SenderID)
	}

	if repo.message.ConversationID != 1 {
		t.Fatalf("expected conversation ID 1, got %d", repo.message.ConversationID)
	}
}

func TestSendMessage_CreateMessageFails(t *testing.T) {
	repo := &MockRepository{
		conversation: &Conversation{
			ID:        1,
			ClientID:  10,
			CleanerID: 11,
			Status:    "active",
		},
		createMessageErr: errors.New("database error"),
	}

	service := NewService(repo, &mockBookingReader{})

	_, err := service.SendMessage(
		context.Background(),
		1,
		10,
		"client",
		SendMessageRequest{
			Content: "Hello",
		},
	)

	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGetConversation_Success(t *testing.T) {
	repo := &MockRepository{
		conversation: &Conversation{
			ID:        1,
			ClientID:  10,
			CleanerID: 11,
			Status:    "active",
		},
	}

	service := NewService(repo, &mockBookingReader{})

	conversation, err := service.GetConversation(
		context.Background(),
		1,
		10,
		"client",
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if conversation == nil {
		t.Fatal("expected conversation")
	}

	if conversation.ID != 1 {
		t.Fatalf("expected conversation ID 1, got %d", conversation.ID)
	}
}

func TestGetConversation_NonParticipantForbidden(t *testing.T) {
	repo := &MockRepository{
		conversation: &Conversation{
			ID:        1,
			ClientID:  10,
			CleanerID: 11,
			Status:    "active",
		},
	}

	service := NewService(repo, &mockBookingReader{})

	_, err := service.GetConversation(
		context.Background(),
		1,
		99,
		"user",
	)

	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestGetConversation_RepositoryError(t *testing.T) {
	repositoryErr := errors.New("database error")

	repo := &MockRepository{
		getConversationErr: repositoryErr,
	}

	service := NewService(repo, &mockBookingReader{})

	_, err := service.GetConversation(
		context.Background(),
		1,
		10,
		"client",
	)

	if !errors.Is(err, repositoryErr) {
		t.Fatalf("expected repository error, got %v", err)
	}
}

func TestListConversations_Success(t *testing.T) {
	repo := &MockRepository{
		conversations: []ConversationSummary{
			{
				ID:            1,
				OtherUserName: "Cleaner One",
				LastMessage:   stringPtr("Hello"),
			},
			{
				ID:            2,
				OtherUserName: "Cleaner Two",
				LastMessage:   stringPtr("Morning"),
			},
		},
	}

	service := NewService(repo, &mockBookingReader{})

	conversations, err := service.ListConversations(
		context.Background(),
		10,
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(conversations) != 2 {
		t.Fatalf("expected 2 conversations, got %d", len(conversations))
	}
}

func TestListConversations_Empty(t *testing.T) {
	repo := &MockRepository{
		conversations: []ConversationSummary{},
	}

	service := NewService(repo, &mockBookingReader{})

	conversations, err := service.ListConversations(
		context.Background(),
		10,
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(conversations) != 0 {
		t.Fatalf("expected empty list")
	}
}

func TestListConversations_RepositoryError(t *testing.T) {
	repositoryErr := errors.New("database error")

	repo := &MockRepository{
		listConversationErr: repositoryErr,
	}

	service := NewService(repo, &mockBookingReader{})

	_, err := service.ListConversations(
		context.Background(),
		10,
	)

	if !errors.Is(err, repositoryErr) {
		t.Fatalf("expected repository error, got %v", err)
	}
}

func TestListMessages_Success(t *testing.T) {
	repo := &MockRepository{
		conversation: &Conversation{
			ID:        1,
			ClientID:  10,
			CleanerID: 11,
			Status:    "active",
		},
		messages: []Message{
			{
				ID:             1,
				ConversationID: 1,
				SenderID:       10,
				Content:        "Hello",
			},
			{
				ID:             2,
				ConversationID: 1,
				SenderID:       11,
				Content:        "Hi",
			},
		},
	}

	service := NewService(repo, &mockBookingReader{})

	messages, err := service.ListMessages(
		context.Background(),
		1,
		10,
		"client",
		20,
		0,
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(messages) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(messages))
	}
}

func TestListMessages_NonParticipantForbidden(t *testing.T) {
	repo := &MockRepository{
		conversation: &Conversation{
			ID:        1,
			ClientID:  10,
			CleanerID: 11,
			Status:    "active",
		},
	}

	service := NewService(repo, &mockBookingReader{})

	_, err := service.ListMessages(
		context.Background(),
		1,
		99,
		"user",
		20,
		0,
	)

	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestListMessages_RepositoryError(t *testing.T) {
	repositoryErr := errors.New("database error")

	repo := &MockRepository{
		conversation: &Conversation{
			ID:        1,
			ClientID:  10,
			CleanerID: 11,
			Status:    "active",
		},

		listMessageErr: repositoryErr,
	}

	service := NewService(repo, &mockBookingReader{})

	_, err := service.ListMessages(
		context.Background(),
		1,
		10,
		"client",
		20,
		0,
	)

	if !errors.Is(err, repositoryErr) {
		t.Fatalf("expected repository error, got %v", err)
	}
}

func TestMarkConversationAsRead_Success(t *testing.T) {
	repo := &MockRepository{
		conversation: &Conversation{
			ID:        1,
			ClientID:  10,
			CleanerID: 11,
			Status:    "active",
		},
	}

	service := NewService(repo, &mockBookingReader{})

	err := service.MarkAsRead(
		context.Background(),
		1,
		10,
		"client",
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMarkConversationAsRead_NonParticipantForbidden(t *testing.T) {
	repo := &MockRepository{
		conversation: &Conversation{
			ID:        1,
			ClientID:  10,
			CleanerID: 11,
			Status:    "active",
		},
	}

	service := NewService(repo, &mockBookingReader{})

	err := service.MarkAsRead(
		context.Background(),
		1,
		99,
		"user",
	)

	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestMarkConversationAsRead_RepositoryError(t *testing.T) {
	repositoryErr := errors.New("database error")

	repo := &MockRepository{
		conversation: &Conversation{
			ID:        1,
			ClientID:  10,
			CleanerID: 11,
			Status:    "active",
		},
		markReadErr: repositoryErr,
	}

	service := NewService(repo, &mockBookingReader{})

	err := service.MarkAsRead(
		context.Background(),
		1,
		10,
		"client",
	)

	if !errors.Is(err, repositoryErr) {
		t.Fatalf("expected repository error, got %v", err)
	}
}
