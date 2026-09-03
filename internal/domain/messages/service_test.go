package messages

import (
	"context"
	"errors"
	"testing"
)

type MockRepository struct {
	createErr          error
	listErr            error
	participantErr     error
	participantResults map[uint]bool

	createdMessage *Message
	messages       []Message

	participantCalls []uint
}

func (m *MockRepository) Create(
	ctx context.Context,
	message *Message,
) error {
	m.createdMessage = message
	return m.createErr
}

func (m *MockRepository) ListByJobID(
	ctx context.Context,
	jobID uint,
) ([]Message, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}

	return m.messages, nil
}

func (m *MockRepository) IsJobParticipant(
	ctx context.Context,
	jobID uint,
	userID uint,
) (bool, error) {
	m.participantCalls = append(
		m.participantCalls,
		userID,
	)

	if m.participantErr != nil {
		return false, m.participantErr
	}

	if m.participantResults == nil {
		return false, nil
	}

	return m.participantResults[userID], nil
}

func TestNewService(t *testing.T) {
	repo := &MockRepository{}

	service := NewService(repo, nil)

	if service == nil {
		t.Fatal("expected service")
	}

	if service.repo != repo {
		t.Fatal("expected repository to be assigned")
	}
}

func TestSend_InvalidInput(t *testing.T) {
	tests := []struct {
		name     string
		jobID    uint
		senderID uint
		req      SendMessageRequest
	}{
		{
			name:     "zero job ID",
			jobID:    0,
			senderID: 1,
			req: SendMessageRequest{
				ReceiverID: 2,
				Content:    "hello",
			},
		},
		{
			name:     "zero sender ID",
			jobID:    10,
			senderID: 0,
			req: SendMessageRequest{
				ReceiverID: 2,
				Content:    "hello",
			},
		},
		{
			name:     "zero receiver ID",
			jobID:    10,
			senderID: 1,
			req: SendMessageRequest{
				ReceiverID: 0,
				Content:    "hello",
			},
		},
		{
			name:     "empty content",
			jobID:    10,
			senderID: 1,
			req: SendMessageRequest{
				ReceiverID: 2,
				Content:    "",
			},
		},
		{
			name:     "whitespace content",
			jobID:    10,
			senderID: 1,
			req: SendMessageRequest{
				ReceiverID: 2,
				Content:    "     ",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service := NewService(
				&MockRepository{},
				nil,
			)

			message, err := service.Send(
				context.Background(),
				test.jobID,
				test.senderID,
				test.req,
			)

			if message != nil {
				t.Fatal("expected nil message")
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

func TestSend_SenderParticipantCheckError(t *testing.T) {
	participantErr := errors.New("participant lookup failed")

	service := NewService(
		&MockRepository{
			participantErr: participantErr,
		},
		nil,
	)

	_, err := service.Send(
		context.Background(),
		10,
		1,
		SendMessageRequest{
			ReceiverID: 2,
			Content:    "hello",
		},
	)

	if !errors.Is(err, participantErr) {
		t.Fatalf(
			"expected participant error, got %v",
			err,
		)
	}
}

func TestSend_SenderNotParticipant(t *testing.T) {
	repo := &MockRepository{
		participantResults: map[uint]bool{
			1: false,
			2: true,
		},
	}

	service := NewService(repo, nil)

	message, err := service.Send(
		context.Background(),
		10,
		1,
		SendMessageRequest{
			ReceiverID: 2,
			Content:    "hello",
		},
	)

	if message != nil {
		t.Fatal("expected nil message")
	}

	if !errors.Is(err, ErrForbidden) {
		t.Fatalf(
			"expected ErrForbidden, got %v",
			err,
		)
	}
}

func TestSend_ReceiverNotParticipant(t *testing.T) {
	repo := &MockRepository{
		participantResults: map[uint]bool{
			1: true,
			2: false,
		},
	}

	service := NewService(repo, nil)

	message, err := service.Send(
		context.Background(),
		10,
		1,
		SendMessageRequest{
			ReceiverID: 2,
			Content:    "hello",
		},
	)

	if message != nil {
		t.Fatal("expected nil message")
	}

	if !errors.Is(err, ErrForbidden) {
		t.Fatalf(
			"expected ErrForbidden, got %v",
			err,
		)
	}
}

func TestSend_CreateError(t *testing.T) {
	createErr := errors.New("create failed")

	repo := &MockRepository{
		participantResults: map[uint]bool{
			1: true,
			2: true,
		},
		createErr: createErr,
	}

	service := NewService(repo, nil)

	message, err := service.Send(
		context.Background(),
		10,
		1,
		SendMessageRequest{
			ReceiverID: 2,
			Content:    "hello",
		},
	)

	if message != nil {
		t.Fatal("expected nil message")
	}

	if !errors.Is(err, createErr) {
		t.Fatalf(
			"expected create error, got %v",
			err,
		)
	}
}

func TestSend_Success(t *testing.T) {
	repo := &MockRepository{
		participantResults: map[uint]bool{
			1: true,
			2: true,
		},
	}

	service := NewService(repo, nil)

	message, err := service.Send(
		context.Background(),
		10,
		1,
		SendMessageRequest{
			ReceiverID: 2,
			Content:    "   hello there   ",
		},
	)

	if err != nil {
		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}

	if message == nil {
		t.Fatal("expected message")
	}

	if message.JobID != 10 {
		t.Fatalf(
			"expected job ID 10, got %d",
			message.JobID,
		)
	}

	if message.SenderID != 1 {
		t.Fatalf(
			"expected sender ID 1, got %d",
			message.SenderID,
		)
	}

	if message.ReceiverID != 2 {
		t.Fatalf(
			"expected receiver ID 2, got %d",
			message.ReceiverID,
		)
	}

	if message.Content != "hello there" {
		t.Fatalf(
			"expected trimmed content, got %q",
			message.Content,
		)
	}

	if repo.createdMessage != message {
		t.Fatal(
			"expected repository to receive message",
		)
	}

	if len(repo.participantCalls) != 2 {
		t.Fatalf(
			"expected two participant checks, got %d",
			len(repo.participantCalls),
		)
	}
}

func TestListConversation_InvalidInput(t *testing.T) {
	tests := []struct {
		name   string
		jobID  uint
		userID uint
	}{
		{
			name:   "zero job ID",
			jobID:  0,
			userID: 1,
		},
		{
			name:   "zero user ID",
			jobID:  10,
			userID: 0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service := NewService(
				&MockRepository{},
				nil,
			)

			messages, err := service.ListConversation(
				context.Background(),
				test.jobID,
				test.userID,
			)

			if messages != nil {
				t.Fatal("expected nil messages")
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

func TestListConversation_ParticipantCheckError(
	t *testing.T,
) {
	participantErr := errors.New("lookup failed")

	service := NewService(
		&MockRepository{
			participantErr: participantErr,
		},
		nil,
	)

	_, err := service.ListConversation(
		context.Background(),
		10,
		1,
	)

	if !errors.Is(err, participantErr) {
		t.Fatalf(
			"expected participant error, got %v",
			err,
		)
	}
}

func TestListConversation_Forbidden(t *testing.T) {
	service := NewService(
		&MockRepository{
			participantResults: map[uint]bool{
				1: false,
			},
		},
		nil,
	)

	messages, err := service.ListConversation(
		context.Background(),
		10,
		1,
	)

	if messages != nil {
		t.Fatal("expected nil messages")
	}

	if !errors.Is(err, ErrForbidden) {
		t.Fatalf(
			"expected ErrForbidden, got %v",
			err,
		)
	}
}

func TestListConversation_Success(t *testing.T) {
	expected := []Message{
		{
			ID:         1,
			JobID:      10,
			SenderID:   1,
			ReceiverID: 2,
			Content:    "hello",
		},
	}

	service := NewService(
		&MockRepository{
			participantResults: map[uint]bool{
				1: true,
			},
			messages: expected,
		},
		nil,
	)

	result, err := service.ListConversation(
		context.Background(),
		10,
		1,
	)

	if err != nil {
		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}

	if len(result) != 1 {
		t.Fatalf(
			"expected one message, got %d",
			len(result),
		)
	}

	if result[0].Content != "hello" {
		t.Fatalf(
			"expected hello, got %q",
			result[0].Content,
		)
	}
}

func TestListConversation_RepositoryError(t *testing.T) {
	listErr := errors.New("list failed")

	service := NewService(
		&MockRepository{
			participantResults: map[uint]bool{
				1: true,
			},
			listErr: listErr,
		},
		nil,
	)

	_, err := service.ListConversation(
		context.Background(),
		10,
		1,
	)

	if !errors.Is(err, listErr) {
		t.Fatalf(
			"expected list error, got %v",
			err,
		)
	}
}

var _ Repository = (*MockRepository)(nil)

