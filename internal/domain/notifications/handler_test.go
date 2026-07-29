package notifications

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/rodrigueghenda/jobira/internal/security/identity"
)

func TestHandler_ListMine_Success(t *testing.T) {
	repo := &MockRepository{
		notifications: []Notification{
			{
				ID:      1,
				UserID:  5,
				Title:   "New message",
				Message: "You received a new message",
				Type:    "chat",
				IsRead:  false,
			},
			{
				ID:      2,
				UserID:  5,
				Title:   "Application update",
				Message: "Your application was accepted",
				Type:    "application",
				IsRead:  true,
			},
		},
	}

	service := NewService(repo)
	handler := NewHandler(service)

	req := httptest.NewRequest(http.MethodGet, "/notifications", nil)

	ctx := identity.WithUser(req.Context(), identity.UserIdentity{
		UserID: 5,
		Email:  "test@example.com",
		Role:   "cleaner",
	})

	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()

	handler.ListMine(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var notifications []Notification

	if err := json.NewDecoder(rr.Body).Decode(&notifications); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(notifications) != 2 {
		t.Fatalf("expected 2 notifications, got %d", len(notifications))
	}

	if notifications[0].Title != "New message" {
		t.Fatalf("unexpected notification title: %s", notifications[0].Title)
	}
}

func TestHandler_ListMine_Unauthorized(t *testing.T) {
	repo := &MockRepository{}
	service := NewService(repo)
	handler := NewHandler(service)

	req := httptest.NewRequest(http.MethodGet, "/notifications", nil)
	rr := httptest.NewRecorder()

	handler.ListMine(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusUnauthorized,
			rr.Code,
		)
	}

	var responseBody map[string]string

	if err := json.NewDecoder(rr.Body).Decode(&responseBody); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if responseBody["error"] != "unauthorized" {
		t.Fatalf(
			"expected error %q, got %q",
			"unauthorized",
			responseBody["error"],
		)
	}
}

func TestHandler_ListMine_RepositoryError(t *testing.T) {
	repo := &MockRepository{
		listErr: errors.New("database unavailable"),
	}

	service := NewService(repo)
	handler := NewHandler(service)

	req := httptest.NewRequest(http.MethodGet, "/notifications", nil)

	ctx := identity.WithUser(req.Context(), identity.UserIdentity{
		UserID: 5,
		Email:  "test@example.com",
		Role:   "cleaner",
	})

	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()

	handler.ListMine(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusInternalServerError,
			rr.Code,
		)
	}

	var responseBody map[string]string

	if err := json.NewDecoder(rr.Body).Decode(&responseBody); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if responseBody["error"] != "database unavailable" {
		t.Fatalf(
			"expected error %q, got %q",
			"database unavailable",
			responseBody["error"],
		)
	}
}

func TestHandler_ListUnreadMine_Success(t *testing.T) {
	repo := &MockRepository{
		notifications: []Notification{
			{
				ID:      1,
				UserID:  5,
				Title:   "New message",
				Message: "You received a new message",
				Type:    "chat",
				IsRead:  false,
			},
		},
	}

	service := NewService(repo)
	handler := NewHandler(service)

	req := httptest.NewRequest(http.MethodGet, "/notifications/unread", nil)

	ctx := identity.WithUser(req.Context(), identity.UserIdentity{
		UserID: 5,
		Email:  "test@example.com",
		Role:   "cleaner",
	})

	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.ListUnreadMine(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var notifications []Notification

	if err := json.NewDecoder(rr.Body).Decode(&notifications); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(notifications) != 1 {
		t.Fatalf("expected 1 unread notification, got %d", len(notifications))
	}

	if notifications[0].IsRead {
		t.Fatal("expected notification to be unread")
	}
}

func TestHandler_ListUnreadMine_Unauthorized(t *testing.T) {
	repo := &MockRepository{}
	service := NewService(repo)
	handler := NewHandler(service)

	req := httptest.NewRequest(http.MethodGet, "/notifications/unread", nil)
	rr := httptest.NewRecorder()

	handler.ListUnreadMine(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rr.Code)
	}

	var responseBody map[string]string

	if err := json.NewDecoder(rr.Body).Decode(&responseBody); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if responseBody["error"] != "unauthorized" {
		t.Fatalf("expected error %q, got %q", "unauthorized", responseBody["error"])
	}
}

func TestHandler_ListUnreadMine_RepositoryError(t *testing.T) {
	repo := &MockRepository{
		listErr: errors.New("database unavailable"),
	}

	service := NewService(repo)
	handler := NewHandler(service)

	req := httptest.NewRequest(http.MethodGet, "/notifications", nil)

	ctx := identity.WithUser(req.Context(), identity.UserIdentity{
		UserID: 5,
		Email:  "test@example.com",
		Role:   "cleaner",
	})

	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()

	handler.ListUnreadMine(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusInternalServerError,
			rr.Code,
		)
	}

	var responseBody map[string]string

	if err := json.NewDecoder(rr.Body).Decode(&responseBody); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if responseBody["error"] != "database unavailable" {
		t.Fatalf(
			"expected error %q, got %q",
			"database unavailable",
			responseBody["error"],
		)
	}
}

func TestHandler_CountUnreadMine_Success(t *testing.T) {
	repo := &MockRepository{
		unreadCount: 5,
	}

	service := NewService(repo)
	handler := NewHandler(service)

	req := httptest.NewRequest(http.MethodGet, "/notifications/unread/count", nil)

	ctx := identity.WithUser(req.Context(), identity.UserIdentity{
		UserID: 5,
		Email:  "test@example.com",
		Role:   "cleaner",
	})

	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()

	handler.CountUnreadMine(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var responseBody map[string]int

	if err := json.NewDecoder(rr.Body).Decode(&responseBody); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if responseBody["unread_count"] != 5 {
		t.Fatalf(
			"expected unread_count 5, got %d",
			responseBody["unread_count"],
		)
	}
}

func TestHandler_CountUnreadMine_Unauthorized(t *testing.T) {
	repo := &MockRepository{}
	service := NewService(repo)
	handler := NewHandler(service)

	req := httptest.NewRequest(http.MethodGet, "/notifications/unread/count", nil)
	rr := httptest.NewRecorder()

	handler.CountUnreadMine(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusUnauthorized,
			rr.Code,
		)
	}
}
func TestHandler_CountUnreadMine_RepositoryError(t *testing.T) {
	repo := &MockRepository{
		countErr: errors.New("database unavailable"),
	}

	service := NewService(repo)
	handler := NewHandler(service)

	req := httptest.NewRequest(
		http.MethodGet,
		"/notifications/unread/count",
		nil,
	)

	ctx := identity.WithUser(
		req.Context(),
		identity.UserIdentity{
			UserID: 5,
			Email:  "test@example.com",
			Role:   "cleaner",
		},
	)

	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.CountUnreadMine(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusInternalServerError,
			rr.Code,
		)
	}
}

func TestHandler_MarkAllAsRead_Success(t *testing.T) {
	repo := &MockRepository{}

	service := NewService(repo)
	handler := NewHandler(service)

	req := httptest.NewRequest(http.MethodPost, "/notifications/read-all", nil)

	ctx := identity.WithUser(req.Context(), identity.UserIdentity{
		UserID: 5,
		Email:  "test@example.com",
		Role:   "cleaner",
	})

	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()

	handler.MarkAllAsRead(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var responseBody map[string]string

	if err := json.NewDecoder(rr.Body).Decode(&responseBody); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if responseBody["message"] != "all notifications marked as read" {
		t.Fatalf(
			"expected success message, got %q",
			responseBody["message"],
		)
	}
}

func TestHandler_MarkAllAsRead_Unauthorized(t *testing.T) {
	repo := &MockRepository{}
	service := NewService(repo)
	handler := NewHandler(service)

	req := httptest.NewRequest(http.MethodPost, "/notifications/read-all", nil)
	rr := httptest.NewRecorder()

	handler.MarkAllAsRead(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusUnauthorized,
			rr.Code,
		)
	}
}
func TestHandler_MarkAllAsRead_RepositoryError(t *testing.T) {
	repo := &MockRepository{
		markAllErr: errors.New("database unavailable"),
	}

	service := NewService(repo)
	handler := NewHandler(service)

	req := httptest.NewRequest(
		http.MethodPatch,
		"/notifications/read-all",
		nil,
	)

	ctx := identity.WithUser(
		req.Context(),
		identity.UserIdentity{
			UserID: 5,
			Email:  "test@example.com",
			Role:   "cleaner",
		},
	)

	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.MarkAllAsRead(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusInternalServerError,
			rr.Code,
		)
	}
}

func TestHandler_MarkAsRead_Success(t *testing.T) {
	repo := &MockRepository{}
	service := NewService(repo)
	handler := NewHandler(service)

	req := httptest.NewRequest(
		http.MethodPatch,
		"/notifications/1/read",
		nil,
	)

	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add("id", "1")

	req = req.WithContext(
		context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx),
	)

	ctx := identity.WithUser(req.Context(), identity.UserIdentity{
		UserID: 5,
		Email:  "test@example.com",
		Role:   "cleaner",
	})

	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()

	handler.MarkAsRead(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var responseBody map[string]string

	if err := json.NewDecoder(rr.Body).Decode(&responseBody); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if responseBody["message"] != "notification marked as read" {
		t.Fatalf(
			"expected success message, got %q",
			responseBody["message"],
		)
	}
}

func TestHandler_MarkAsRead_Unauthorized(t *testing.T) {
	repo := &MockRepository{}
	service := NewService(repo)
	handler := NewHandler(service)

	req := httptest.NewRequest(
		http.MethodPatch,
		"/notifications/1/read",
		nil,
	)

	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add("id", "1")

	req = req.WithContext(
		context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx),
	)

	rr := httptest.NewRecorder()

	handler.MarkAsRead(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusUnauthorized,
			rr.Code,
		)
	}
}

func TestHandler_MarkAsRead_InvalidID(t *testing.T) {
	repo := &MockRepository{}
	service := NewService(repo)
	handler := NewHandler(service)

	req := httptest.NewRequest(
		http.MethodPatch,
		"/notifications/not-a-number/read",
		nil,
	)

	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add("id", "not-a-number")

	req = req.WithContext(
		context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx),
	)

	ctx := identity.WithUser(req.Context(), identity.UserIdentity{
		UserID: 5,
		Email:  "test@example.com",
		Role:   "cleaner",
	})

	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()

	handler.MarkAsRead(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			rr.Code,
		)
	}

	var responseBody map[string]string

	if err := json.NewDecoder(rr.Body).Decode(&responseBody); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if responseBody["error"] != "invalid notification id" {
		t.Fatalf(
			"expected error %q, got %q",
			"invalid notification id",
			responseBody["error"],
		)
	}
}

func TestHandler_MarkAsRead_NotFound(t *testing.T) {
	repo := &MockRepository{
		markErr: ErrNotificationNotFound,
	}

	service := NewService(repo)
	handler := NewHandler(service)

	req := httptest.NewRequest(
		http.MethodPatch,
		"/notifications/1/read",
		nil,
	)

	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add("id", "1")

	req = req.WithContext(
		context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx),
	)

	ctx := identity.WithUser(req.Context(), identity.UserIdentity{
		UserID: 5,
		Email:  "test@example.com",
		Role:   "cleaner",
	})

	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()

	handler.MarkAsRead(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, rr.Code)
	}
}
func TestHandler_MarkAsRead_RepositoryError(t *testing.T) {
	repoErr := errors.New("database unavailable")

	repo := &MockRepository{
		markErr: repoErr,
	}

	service := NewService(repo)
	handler := NewHandler(service)

	req := httptest.NewRequest(
		http.MethodPatch,
		"/notifications/1/read",
		nil,
	)

	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add("id", "1")

	req = req.WithContext(
		context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx),
	)

	ctx := identity.WithUser(req.Context(), identity.UserIdentity{
		UserID: 5,
		Email:  "test@example.com",
		Role:   "cleaner",
	})

	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.MarkAsRead(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusInternalServerError,
			rr.Code,
		)
	}

	var responseBody map[string]string

	if err := json.NewDecoder(rr.Body).Decode(&responseBody); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if responseBody["error"] != repoErr.Error() {
		t.Fatalf(
			"expected error %q, got %q",
			repoErr.Error(),
			responseBody["error"],
		)
	}
}

func TestHandler_ListMine_InvalidInput(t *testing.T) {
	repo := &MockRepository{
		listErr: ErrInvalidInput,
	}

	service := NewService(repo)
	handler := NewHandler(service)

	req := httptest.NewRequest(http.MethodGet, "/notifications", nil)

	ctx := identity.WithUser(req.Context(), identity.UserIdentity{
		UserID: 5,
		Email:  "test@example.com",
		Role:   "cleaner",
	})

	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.ListMine(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			rr.Code,
		)
	}

	var responseBody map[string]string

	if err := json.NewDecoder(rr.Body).Decode(&responseBody); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if responseBody["error"] != ErrInvalidInput.Error() {
		t.Fatalf(
			"expected error %q, got %q",
			ErrInvalidInput.Error(),
			responseBody["error"],
		)
	}
}

func TestHandler_ListUnreadMine_InvalidInput(t *testing.T) {
	repo := &MockRepository{
		listErr: ErrInvalidInput,
	}

	service := NewService(repo)
	handler := NewHandler(service)

	req := httptest.NewRequest(http.MethodGet, "/notifications/unread", nil)

	ctx := identity.WithUser(req.Context(), identity.UserIdentity{
		UserID: 5,
		Email:  "test@example.com",
		Role:   "cleaner",
	})

	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.ListUnreadMine(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			rr.Code,
		)
	}

	var responseBody map[string]string

	if err := json.NewDecoder(rr.Body).Decode(&responseBody); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if responseBody["error"] != ErrInvalidInput.Error() {
		t.Fatalf(
			"expected error %q, got %q",
			ErrInvalidInput.Error(),
			responseBody["error"],
		)
	}
}

func TestHandler_CountUnreadMine_InvalidInput(t *testing.T) {
	repo := &MockRepository{
		countErr: ErrInvalidInput,
	}

	service := NewService(repo)
	handler := NewHandler(service)

	req := httptest.NewRequest(http.MethodGet, "/notifications/unread/count", nil)

	ctx := identity.WithUser(req.Context(), identity.UserIdentity{
		UserID: 5,
		Email:  "test@example.com",
		Role:   "cleaner",
	})

	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.CountUnreadMine(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			rr.Code,
		)
	}

	var responseBody map[string]string

	if err := json.NewDecoder(rr.Body).Decode(&responseBody); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if responseBody["error"] != ErrInvalidInput.Error() {
		t.Fatalf(
			"expected error %q, got %q",
			ErrInvalidInput.Error(),
			responseBody["error"],
		)
	}
}
func TestHandler_MarkAllAsRead_InvalidInput(t *testing.T) {
	repo := &MockRepository{
		markAllErr: ErrInvalidInput,
	}

	service := NewService(repo)
	handler := NewHandler(service)

	req := httptest.NewRequest(http.MethodPatch, "/notifications/read-all", nil)

	ctx := identity.WithUser(req.Context(), identity.UserIdentity{
		UserID: 5,
		Email:  "test@example.com",
		Role:   "cleaner",
	})

	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.MarkAllAsRead(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			rr.Code,
		)
	}

	var responseBody map[string]string

	if err := json.NewDecoder(rr.Body).Decode(&responseBody); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if responseBody["error"] != ErrInvalidInput.Error() {
		t.Fatalf(
			"expected error %q, got %q",
			ErrInvalidInput.Error(),
			responseBody["error"],
		)
	}
}

func TestHandler_MarkAsRead_InvalidInput(t *testing.T) {
	repo := &MockRepository{
		markErr: ErrInvalidInput,
	}

	service := NewService(repo)
	handler := NewHandler(service)

	req := httptest.NewRequest(
		http.MethodPatch,
		"/notifications/1/read",
		nil,
	)

	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add("id", "1")

	req = req.WithContext(
		context.WithValue(
			req.Context(),
			chi.RouteCtxKey,
			routeCtx,
		),
	)

	ctx := identity.WithUser(
		req.Context(),
		identity.UserIdentity{
			UserID: 5,
			Email:  "test@example.com",
			Role:   "cleaner",
		},
	)

	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.MarkAsRead(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			rr.Code,
		)
	}

	var responseBody map[string]string

	if err := json.NewDecoder(rr.Body).Decode(&responseBody); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if responseBody["error"] != ErrInvalidInput.Error() {
		t.Fatalf(
			"expected error %q, got %q",
			ErrInvalidInput.Error(),
			responseBody["error"],
		)
	}
}
