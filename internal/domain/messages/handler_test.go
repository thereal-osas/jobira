package messages

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/rodrigueghenda/jobira/internal/security/identity"
)

func messagesTestAuth(
	userID uint,
	role string,
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				ctx := identity.WithUser(
					r.Context(),
					identity.UserIdentity{
						UserID: userID,
						Email:  "test@example.com",
						Role:   role,
					},
				)

				next.ServeHTTP(
					w,
					r.WithContext(ctx),
				)
			},
		)
	}
}

func newMessagesTestRouter(
	repo *MockRepository,
	userID uint,
	role string,
) *chi.Mux {
	service := NewService(repo, nil)
	handler := NewHandler(service)

	router := chi.NewRouter()

	RegisterRoutes(
		router,
		handler,
		messagesTestAuth(userID, role),
	)

	return router
}

func TestHandler_NewHandler(t *testing.T) {
	service := NewService(
		&MockRepository{},
		nil,
	)

	handler := NewHandler(service)

	if handler == nil {
		t.Fatal("expected handler")
	}

	if handler.service != service {
		t.Fatal("expected service to be assigned")
	}
}

func TestHandler_Send_Success(t *testing.T) {
	repo := &MockRepository{
		participantResults: map[uint]bool{
			1: true,
			2: true,
		},
	}

	router := newMessagesTestRouter(
		repo,
		1,
		"client",
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/messages/jobs/10",
		bytes.NewBufferString(`{
			"receiver_id":2,
			"content":"hello"
		}`),
	)

	req.Header.Set(
		"Content-Type",
		"application/json",
	)

	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusCreated {
		t.Fatalf(
			"expected 201 got %d: %s",
			res.Code,
			res.Body.String(),
		)
	}

	var message Message

	if err := json.NewDecoder(
		res.Body,
	).Decode(&message); err != nil {
		t.Fatalf(
			"failed decoding response: %v",
			err,
		)
	}

	if message.JobID != 10 {
		t.Fatalf(
			"expected job ID 10 got %d",
			message.JobID,
		)
	}

	if message.SenderID != 1 {
		t.Fatalf(
			"expected sender ID 1 got %d",
			message.SenderID,
		)
	}

	if message.ReceiverID != 2 {
		t.Fatalf(
			"expected receiver ID 2 got %d",
			message.ReceiverID,
		)
	}
}

func TestHandler_Send_Unauthorized(t *testing.T) {
	service := NewService(
		&MockRepository{},
		nil,
	)

	handler := NewHandler(service)

	req := httptest.NewRequest(
		http.MethodPost,
		"/messages/jobs/10",
		bytes.NewBufferString(`{}`),
	)

	res := httptest.NewRecorder()

	handler.Send(res, req)

	if res.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected 401 got %d",
			res.Code,
		)
	}
}

func TestHandler_Send_InvalidJobID(t *testing.T) {
	router := newMessagesTestRouter(
		&MockRepository{},
		1,
		"client",
	)

	tests := []string{
		"/messages/jobs/abc",
		"/messages/jobs/0",
	}

	for _, path := range tests {
		t.Run(path, func(t *testing.T) {
			req := httptest.NewRequest(
				http.MethodPost,
				path,
				bytes.NewBufferString(`{
					"receiver_id":2,
					"content":"hello"
				}`),
			)

			res := httptest.NewRecorder()

			router.ServeHTTP(res, req)

			if res.Code != http.StatusBadRequest {
				t.Fatalf(
					"expected 400 got %d",
					res.Code,
				)
			}
		})
	}
}

func TestHandler_Send_InvalidBody(t *testing.T) {
	router := newMessagesTestRouter(
		&MockRepository{},
		1,
		"client",
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/messages/jobs/10",
		bytes.NewBufferString("{invalid"),
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

func TestHandler_Send_InvalidInput(t *testing.T) {
	router := newMessagesTestRouter(
		&MockRepository{},
		1,
		"client",
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/messages/jobs/10",
		bytes.NewBufferString(`{
			"receiver_id":0,
			"content":"hello"
		}`),
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

func TestHandler_Send_Forbidden(t *testing.T) {
	router := newMessagesTestRouter(
		&MockRepository{
			participantResults: map[uint]bool{
				1: false,
			},
		},
		1,
		"client",
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/messages/jobs/10",
		bytes.NewBufferString(`{
			"receiver_id":2,
			"content":"hello"
		}`),
	)

	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusForbidden {
		t.Fatalf(
			"expected 403 got %d: %s",
			res.Code,
			res.Body.String(),
		)
	}
}

func TestHandler_Send_ServiceError(t *testing.T) {
	router := newMessagesTestRouter(
		&MockRepository{
			participantErr: errors.New(
				"database failed",
			),
		},
		1,
		"client",
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/messages/jobs/10",
		bytes.NewBufferString(`{
			"receiver_id":2,
			"content":"hello"
		}`),
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

func TestHandler_ListConversation_Success(
	t *testing.T,
) {
	repo := &MockRepository{
		participantResults: map[uint]bool{
			1: true,
		},
		messages: []Message{
			{
				ID:         1,
				JobID:      10,
				SenderID:   1,
				ReceiverID: 2,
				Content:    "hello",
			},
		},
	}

	router := newMessagesTestRouter(
		repo,
		1,
		"client",
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/messages/jobs/10",
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

	var messages []Message

	if err := json.NewDecoder(
		res.Body,
	).Decode(&messages); err != nil {
		t.Fatalf(
			"failed decoding response: %v",
			err,
		)
	}

	if len(messages) != 1 {
		t.Fatalf(
			"expected one message got %d",
			len(messages),
		)
	}
}

func TestHandler_ListConversation_Unauthorized(
	t *testing.T,
) {
	handler := NewHandler(
		NewService(
			&MockRepository{},
			nil,
		),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/messages/jobs/10",
		nil,
	)

	res := httptest.NewRecorder()

	handler.ListConversation(res, req)

	if res.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected 401 got %d",
			res.Code,
		)
	}
}

func TestHandler_ListConversation_InvalidJobID(
	t *testing.T,
) {
	router := newMessagesTestRouter(
		&MockRepository{},
		1,
		"client",
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/messages/jobs/abc",
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

func TestHandler_ListConversation_Forbidden(
	t *testing.T,
) {
	router := newMessagesTestRouter(
		&MockRepository{
			participantResults: map[uint]bool{
				1: false,
			},
		},
		1,
		"client",
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/messages/jobs/10",
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

func TestHandler_ListConversation_ServiceError(
	t *testing.T,
) {
	router := newMessagesTestRouter(
		&MockRepository{
			participantResults: map[uint]bool{
				1: true,
			},
			listErr: errors.New(
				"database failed",
			),
		},
		1,
		"client",
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/messages/jobs/10",
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
