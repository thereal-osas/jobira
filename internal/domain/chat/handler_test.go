package chat

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/rodrigueghenda/jobira/internal/domain/bookings"
	"github.com/rodrigueghenda/jobira/internal/security/identity"
)

func testAuthMiddleware(userID uint, role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := identity.WithUser(
				r.Context(),
				identity.UserIdentity{
					UserID: userID,
					Email:  "client@example.com",
					Role:   role,
				},
			)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func TestHandler_ListConversations_Success(t *testing.T) {
	repo := &MockRepository{
		conversations: []ConversationSummary{},
	}

	bookingReader := &mockBookingReader{}

	service := NewService(repo, bookingReader)
	handler := NewHandler(service)

	router := chi.NewRouter()

	RegisterRoutes(
		router,
		handler,
		testAuthMiddleware(1, "client"),
	)

	request := httptest.NewRequest(
		http.MethodGet,
		"/chat/conversations",
		nil,
	)

	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusOK,
			response.Code,
			response.Body.String(),
		)
	}

	contentType := response.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf(
			"expected Content-Type application/json, got %q",
			contentType,
		)
	}

	var body ConversationListResponse

	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	if body.Conversations == nil {
		t.Fatal("expected conversations to be an empty slice, got nil")
	}

	if len(body.Conversations) != 0 {
		t.Errorf(
			"expected 0 conversations, got %d",
			len(body.Conversations),
		)
	}
}

func TestHandler_ListConversations_Unauthorized(t *testing.T) {
	repo := &MockRepository{}
	bookingReader := &mockBookingReader{}

	service := NewService(repo, bookingReader)
	handler := NewHandler(service)

	router := chi.NewRouter()

	RegisterRoutes(
		router,
		handler,
		func(next http.Handler) http.Handler {
			return next
		},
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/chat/conversations",
		nil,
	)

	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected %d got %d",
			http.StatusUnauthorized,
			res.Code,
		)
	}
}

func TestHandler_CreateConversation_Success(t *testing.T) {
	repo := &MockRepository{
		conversation: &Conversation{
			ID:        10,
			BookingID: 25,
			ClientID:  1,
			CleanerID: 2,
			Status:    "active",
		},
	}

	bookingReader := &mockBookingReader{
		booking: &bookings.Booking{
			ID:        25,
			ClientID:  1,
			CleanerID: 2,
		},
	}

	service := NewService(repo, bookingReader)
	handler := NewHandler(service)

	router := chi.NewRouter()

	RegisterRoutes(
		router,
		handler,
		testAuthMiddleware(1, "client"),
	)

	requestBody := []byte(`{
		"booking_id": 25
	}`)

	request := httptest.NewRequest(
		http.MethodPost,
		"/chat/conversations",
		bytes.NewReader(requestBody),
	)
	request.Header.Set("Content-Type", "application/json")

	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusCreated,
			response.Code,
			response.Body.String(),
		)
	}

	var body ConversationResponse

	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	if body.Conversation.ID != 10 {
		t.Errorf(
			"expected conversation ID 10, got %d",
			body.Conversation.ID,
		)
	}

	if body.Conversation.BookingID != 25 {
		t.Errorf(
			"expected booking ID 25, got %d",
			body.Conversation.BookingID,
		)
	}

	if body.Conversation.Status != "active" {
		t.Errorf(
			"expected status active, got %q",
			body.Conversation.Status,
		)
	}
}

func TestHandler_CreateConversation_InvalidBody(t *testing.T) {
	repo := &MockRepository{}
	bookingReader := &mockBookingReader{}

	service := NewService(repo, bookingReader)
	handler := NewHandler(service)

	router := chi.NewRouter()

	RegisterRoutes(
		router,
		handler,
		testAuthMiddleware(1, "client"),
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/chat/conversations",
		bytes.NewBufferString("{invalid json"),
	)

	req.Header.Set("Content-Type", "application/json")

	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected %d got %d",
			http.StatusBadRequest,
			res.Code,
		)
	}
}

func TestHandler_CreateConversation_Unauthorized(t *testing.T) {
	repo := &MockRepository{}
	bookingReader := &mockBookingReader{}

	service := NewService(repo, bookingReader)
	handler := NewHandler(service)

	router := chi.NewRouter()

	RegisterRoutes(
		router,
		handler,
		func(next http.Handler) http.Handler {
			return next
		},
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/chat/conversations",
		bytes.NewBufferString(`{"booking_id":25}`),
	)

	req.Header.Set("Content-Type", "application/json")

	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusUnauthorized,
			res.Code,
			res.Body.String(),
		)
	}
}

func TestHandler_GetConversation_Success(t *testing.T) {
	repo := &MockRepository{
		conversation: &Conversation{
			ID:        10,
			BookingID: 25,
			ClientID:  1,
			CleanerID: 2,
			Status:    "active",
		},
	}

	bookingReader := &mockBookingReader{}

	service := NewService(repo, bookingReader)
	handler := NewHandler(service)

	router := chi.NewRouter()

	RegisterRoutes(
		router,
		handler,
		testAuthMiddleware(1, "client"),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/chat/conversations/10",
		nil,
	)

	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf(
			"expected %d got %d: %s",
			http.StatusOK,
			res.Code,
			res.Body.String(),
		)
	}

	var body ConversationResponse

	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	if body.Conversation.ID != 10 {
		t.Errorf(
			"expected ID 10, got %d",
			body.Conversation.ID,
		)
	}
}

func TestHandler_GetConversation_InvalidID(t *testing.T) {
	repo := &MockRepository{}
	bookingReader := &mockBookingReader{}

	service := NewService(repo, bookingReader)
	handler := NewHandler(service)

	router := chi.NewRouter()

	RegisterRoutes(
		router,
		handler,
		testAuthMiddleware(1, "client"),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/chat/conversations/not-a-number",
		nil,
	)

	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusBadRequest,
			res.Code,
			res.Body.String(),
		)
	}

	var body map[string]string

	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	if body["error"] != "invalid conversation id" {
		t.Errorf(
			"expected error %q, got %q",
			"invalid conversation id",
			body["error"],
		)
	}
}

func TestHandler_GetConversation_Forbidden(t *testing.T) {
	repo := &MockRepository{
		conversation: &Conversation{
			ID:        10,
			BookingID: 25,
			ClientID:  2,
			CleanerID: 3,
			Status:    "active",
		},
	}

	bookingReader := &mockBookingReader{}

	service := NewService(repo, bookingReader)
	handler := NewHandler(service)

	router := chi.NewRouter()

	RegisterRoutes(
		router,
		handler,
		testAuthMiddleware(1, "client"),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/chat/conversations/10",
		nil,
	)

	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusForbidden {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusForbidden,
			res.Code,
			res.Body.String(),
		)
	}

	var body map[string]string

	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	if body["error"] != ErrForbidden.Error() {
		t.Errorf(
			"expected error %q, got %q",
			ErrForbidden.Error(),
			body["error"],
		)
	}
}

func TestHandler_SendMessage_Success(t *testing.T) {
	repo := &MockRepository{
		conversation: &Conversation{
			ID:        10,
			BookingID: 25,
			ClientID:  1,
			CleanerID: 2,
			Status:    "active",
		},
	}

	bookingReader := &mockBookingReader{}

	service := NewService(repo, bookingReader)
	handler := NewHandler(service)

	router := chi.NewRouter()

	RegisterRoutes(
		router,
		handler,
		testAuthMiddleware(1, "client"),
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/chat/conversations/10/messages",
		bytes.NewBufferString(`{
			"content": "Hello cleaner"
		}`),
	)

	req.Header.Set("Content-Type", "application/json")

	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusCreated {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusCreated,
			res.Code,
			res.Body.String(),
		)
	}

	var body MessageResponse

	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	if body.Message.ConversationID != 10 {
		t.Errorf(
			"expected conversation ID 10, got %d",
			body.Message.ConversationID,
		)
	}

	if body.Message.SenderID != 1 {
		t.Errorf(
			"expected sender ID 1, got %d",
			body.Message.SenderID,
		)
	}

	if body.Message.Content != "Hello cleaner" {
		t.Errorf(
			"expected content %q, got %q",
			"Hello cleaner",
			body.Message.Content,
		)
	}

	if body.Message.MessageType != "text" {
		t.Errorf(
			"expected message type %q, got %q",
			"text",
			body.Message.MessageType,
		)
	}
}

func TestHandler_ListMessages_Success(t *testing.T) {
	repo := &MockRepository{
		conversation: &Conversation{
			ID:        10,
			BookingID: 25,
			ClientID:  1,
			CleanerID: 2,
			Status:    "active",
		},
		messages: []Message{
			{
				ID:             1,
				ConversationID: 10,
				SenderID:       1,
				MessageType:    "text",
				Content:        "Hello cleaner",
			},
			{
				ID:             2,
				ConversationID: 10,
				SenderID:       2,
				MessageType:    "text",
				Content:        "Hello client",
			},
		},
	}

	bookingReader := &mockBookingReader{}

	service := NewService(repo, bookingReader)
	handler := NewHandler(service)

	router := chi.NewRouter()

	RegisterRoutes(
		router,
		handler,
		testAuthMiddleware(1, "client"),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/chat/conversations/10/messages?limit=25&offset=5",
		nil,
	)

	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusOK,
			res.Code,
			res.Body.String(),
		)
	}

	var body MessageListResponse

	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	if len(body.Messages) != 2 {
		t.Fatalf(
			"expected 2 messages, got %d",
			len(body.Messages),
		)
	}

	if body.Messages[0].Content != "Hello cleaner" {
		t.Errorf(
			"expected first message content %q, got %q",
			"Hello cleaner",
			body.Messages[0].Content,
		)
	}

	if body.Messages[1].SenderID != 2 {
		t.Errorf(
			"expected second sender ID 2, got %d",
			body.Messages[1].SenderID,
		)
	}
}

func TestHandler_MarkAsRead_Success(t *testing.T) {
	repo := &MockRepository{
		conversation: &Conversation{
			ID:        10,
			BookingID: 25,
			ClientID:  1,
			CleanerID: 2,
			Status:    "active",
		},
	}

	bookingReader := &mockBookingReader{}

	service := NewService(repo, bookingReader)
	handler := NewHandler(service)

	router := chi.NewRouter()

	RegisterRoutes(
		router,
		handler,
		testAuthMiddleware(1, "client"),
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/chat/conversations/10/read",
		nil,
	)

	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusOK,
			res.Code,
			res.Body.String(),
		)
	}

	var body map[string]string

	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	if body["message"] != "conversation marked as read" {
		t.Errorf(
			"expected message %q, got %q",
			"conversation marked as read",
			body["message"],
		)
	}
}

func TestHandler_CreateConversation_ServiceError(t *testing.T) {
	repo := &MockRepository{
		getConversationErr: errors.New("database error"),
	}

	bookingReader := &mockBookingReader{
		booking: &bookings.Booking{
			ID:        25,
			ClientID:  1,
			CleanerID: 2,
		},
	}

	service := NewService(repo, bookingReader)
	handler := NewHandler(service)

	router := chi.NewRouter()

	RegisterRoutes(
		router,
		handler,
		testAuthMiddleware(1, "client"),
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/chat/conversations",
		bytes.NewBufferString(`{"booking_id":25}`),
	)

	req.Header.Set("Content-Type", "application/json")

	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected 500 got %d: %s",
			res.Code,
			res.Body.String(),
		)
	}
}

func TestHandler_SendMessage_InvalidBody(t *testing.T) {
	repo := &MockRepository{}

	service := NewService(
		repo,
		&mockBookingReader{},
	)

	handler := NewHandler(service)

	router := chi.NewRouter()

	RegisterRoutes(
		router,
		handler,
		testAuthMiddleware(1, "client"),
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/chat/conversations/10/messages",
		bytes.NewBufferString("{invalid"),
	)

	req.Header.Set("Content-Type", "application/json")

	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected 400 got %d",
			res.Code,
		)
	}
}

func TestHandler_SendMessage_ServiceError(t *testing.T) {
	repo := &MockRepository{
		conversation: &Conversation{
			ID:        10,
			ClientID:  1,
			CleanerID: 2,
			Status:    "active",
		},
		createMessageErr: errors.New("message failed"),
	}

	service := NewService(
		repo,
		&mockBookingReader{},
	)

	handler := NewHandler(service)

	router := chi.NewRouter()

	RegisterRoutes(
		router,
		handler,
		testAuthMiddleware(1, "client"),
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/chat/conversations/10/messages",
		bytes.NewBufferString(`{"content":"hello"}`),
	)

	req.Header.Set("Content-Type", "application/json")

	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected 500 got %d",
			res.Code,
		)
	}
}

func TestHandler_MarkAsRead_ServiceError(t *testing.T) {
	repo := &MockRepository{
		conversation: &Conversation{
			ID:        10,
			ClientID:  1,
			CleanerID: 2,
			Status:    "active",
		},
		markReadErr: errors.New("mark read failed"),
	}

	service := NewService(
		repo,
		&mockBookingReader{},
	)

	handler := NewHandler(service)

	router := chi.NewRouter()

	RegisterRoutes(
		router,
		handler,
		testAuthMiddleware(1, "client"),
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/chat/conversations/10/read",
		nil,
	)

	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected 500 got %d",
			res.Code,
		)
	}
}

func TestHandler_ListMessages_InvalidLimitUsesFallback(t *testing.T) {
	repo := &MockRepository{
		conversation: &Conversation{
			ID:        10,
			ClientID:  1,
			CleanerID: 2,
			Status:    "active",
		},
	}

	service := NewService(
		repo,
		&mockBookingReader{},
	)

	handler := NewHandler(service)

	router := chi.NewRouter()

	RegisterRoutes(
		router,
		handler,
		testAuthMiddleware(1, "client"),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/chat/conversations/10/messages?limit=abc",
		nil,
	)

	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf(
			"expected 200 got %d: %s",
			res.Code,
			res.Body.String(),
		)
	}
}

func TestHandler_ListMessages_InvalidOffsetUsesFallback(t *testing.T) {
	repo := &MockRepository{
		conversation: &Conversation{
			ID:        10,
			ClientID:  1,
			CleanerID: 2,
			Status:    "active",
		},
	}

	service := NewService(
		repo,
		&mockBookingReader{},
	)

	handler := NewHandler(service)

	router := chi.NewRouter()

	RegisterRoutes(
		router,
		handler,
		testAuthMiddleware(1, "client"),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/chat/conversations/10/messages?offset=abc",
		nil,
	)

	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf(
			"expected 200 got %d: %s",
			res.Code,
			res.Body.String(),
		)
	}
}

func TestHandler_ListMessages_ServiceError(t *testing.T) {
	repo := &MockRepository{
		conversation: &Conversation{
			ID:        10,
			ClientID:  1,
			CleanerID: 2,
			Status:    "active",
		},
		listMessageErr: errors.New("messages failed"),
	}

	service := NewService(
		repo,
		&mockBookingReader{},
	)

	handler := NewHandler(service)

	router := chi.NewRouter()

	RegisterRoutes(
		router,
		handler,
		testAuthMiddleware(1, "client"),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/chat/conversations/10/messages",
		nil,
	)

	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected 500 got %d",
			res.Code,
		)
	}
}

func TestHandler_GetConversation_RepositoryForbidden(t *testing.T) {
	repo := &MockRepository{
		getConversationErr: ErrForbidden,
	}

	service := NewService(
		repo,
		&mockBookingReader{},
	)

	handler := NewHandler(service)

	router := chi.NewRouter()

	RegisterRoutes(
		router,
		handler,
		testAuthMiddleware(1, "client"),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/chat/conversations/10",
		nil,
	)

	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusForbidden {
		t.Fatalf(
			"expected 403 got %d",
			res.Code,
		)
	}
}

func TestHandler_GetConversation_NotFound(t *testing.T) {
	repo := &MockRepository{
		getConversationErr: ErrConversationNotFound,
	}

	service := NewService(
		repo,
		&mockBookingReader{},
	)

	handler := NewHandler(service)

	router := chi.NewRouter()

	RegisterRoutes(
		router,
		handler,
		testAuthMiddleware(1, "client"),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/chat/conversations/10",
		nil,
	)

	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusNotFound {
		t.Fatalf(
			"expected 404 got %d",
			res.Code,
		)
	}
}

func TestHandler_CreateConversation_AlreadyExists(t *testing.T) {
	repo := &MockRepository{
		getConversationErr: ErrConversationExists,
	}

	service := NewService(
		repo,
		&mockBookingReader{
			booking: &bookings.Booking{
				ID:        10,
				ClientID:  1,
				CleanerID: 2,
			},
		},
	)

	handler := NewHandler(service)

	router := chi.NewRouter()

	RegisterRoutes(
		router,
		handler,
		testAuthMiddleware(1, "client"),
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/chat/conversations",
		bytes.NewBufferString(`{"booking_id":10}`),
	)

	req.Header.Set("Content-Type", "application/json")

	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusConflict {
		t.Fatalf(
			"expected 409 got %d",
			res.Code,
		)
	}
}

func TestHandler_SendMessage_InvalidMessage(t *testing.T) {
	repo := &MockRepository{
		conversation: &Conversation{
			ID:        10,
			ClientID:  1,
			CleanerID: 2,
			Status:    "active",
		},
	}

	service := NewService(
		repo,
		&mockBookingReader{},
	)

	handler := NewHandler(service)

	router := chi.NewRouter()

	RegisterRoutes(
		router,
		handler,
		testAuthMiddleware(1, "client"),
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/chat/conversations/10/messages",
		bytes.NewBufferString(`{"content":""}`),
	)

	req.Header.Set("Content-Type", "application/json")

	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected 400 got %d",
			res.Code,
		)
	}
}

func TestHandler_SendMessage_MessageTooLong(t *testing.T) {
	repo := &MockRepository{
		conversation: &Conversation{
			ID:        10,
			ClientID:  1,
			CleanerID: 2,
			Status:    "active",
		},
	}

	service := NewService(
		repo,
		&mockBookingReader{},
	)

	handler := NewHandler(service)

	router := chi.NewRouter()

	RegisterRoutes(
		router,
		handler,
		testAuthMiddleware(1, "client"),
	)

	longMessage := strings.Repeat("a", 4001)

	req := httptest.NewRequest(
		http.MethodPost,
		"/chat/conversations/10/messages",
		bytes.NewBufferString(
			`{"content":"`+longMessage+`"}`,
		),
	)

	req.Header.Set("Content-Type", "application/json")

	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected 400 got %d",
			res.Code,
		)
	}
}

func TestHandler_GenericInternalError(t *testing.T) {
	repo := &MockRepository{
		getConversationErr: errors.New("database connection failed"),
	}

	service := NewService(
		repo,
		&mockBookingReader{},
	)

	handler := NewHandler(service)

	router := chi.NewRouter()

	RegisterRoutes(
		router,
		handler,
		testAuthMiddleware(1, "client"),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/chat/conversations/10",
		nil,
	)

	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected 500 got %d",
			res.Code,
		)
	}
}

func TestHandler_ListConversations_ServiceError(t *testing.T) {
	repo := &MockRepository{
		listConversationErr: errors.New("failed loading conversations"),
	}

	service := NewService(
		repo,
		&mockBookingReader{},
	)

	handler := NewHandler(service)

	router := chi.NewRouter()

	RegisterRoutes(
		router,
		handler,
		testAuthMiddleware(1, "client"),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/chat/conversations",
		nil,
	)

	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected 500 got %d",
			res.Code,
		)
	}
}

func TestHandler_MarkAsRead_InvalidConversationID(t *testing.T) {
	repo := &MockRepository{}

	service := NewService(
		repo,
		&mockBookingReader{},
	)

	handler := NewHandler(service)

	router := chi.NewRouter()

	RegisterRoutes(
		router,
		handler,
		testAuthMiddleware(1, "client"),
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/chat/conversations/abc/read",
		nil,
	)

	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected 400 got %d",
			res.Code,
		)
	}
}

func TestListConversations_InvalidUser(t *testing.T) {
	service := NewService(
		&MockRepository{},
		&mockBookingReader{},
	)

	_, err := service.ListConversations(
		context.Background(),
		0,
	)

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf(
			"expected ErrInvalidInput got %v",
			err,
		)
	}
}

func TestListMessages_DefaultPagination(t *testing.T) {
	repo := &MockRepository{
		conversation: &Conversation{
			ID:        10,
			ClientID:  1,
			CleanerID: 2,
			Status:    "active",
		},
	}

	service := NewService(
		repo,
		&mockBookingReader{},
	)

	_, err := service.ListMessages(
		context.Background(),
		10,
		1,
		"client",
		0,
		-1,
	)

	if err != nil {
		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}
}
