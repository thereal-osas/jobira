package chat

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

func allowRouteTestMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
}

func rejectRouteTestMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(
			w,
			http.StatusText(http.StatusUnauthorized),
			http.StatusUnauthorized,
		)
	})
}

func newChatRouteTestRouter(
	authMiddleware func(http.Handler) http.Handler,
) http.Handler {
	router := chi.NewRouter()

	// The middleware ends the request before the handler is reached.
	// A nil handler is therefore safe for route-registration tests.
	RegisterRoutes(
		router,
		nil,
		authMiddleware,
	)

	return router
}

func TestRegisterRoutes_AllChatRoutesRegistered(t *testing.T) {
	router := newChatRouteTestRouter(allowRouteTestMiddleware)

	tests := []struct {
		name   string
		method string
		path   string
	}{
		{
			name:   "create conversation",
			method: http.MethodPost,
			path:   "/chat/conversations",
		},
		{
			name:   "list conversations",
			method: http.MethodGet,
			path:   "/chat/conversations",
		},
		{
			name:   "get conversation",
			method: http.MethodGet,
			path:   "/chat/conversations/1",
		},
		{
			name:   "send message",
			method: http.MethodPost,
			path:   "/chat/conversations/1/messages",
		},
		{
			name:   "list messages",
			method: http.MethodGet,
			path:   "/chat/conversations/1/messages",
		},
		{
			name:   "mark conversation as read",
			method: http.MethodPost,
			path:   "/chat/conversations/1/read",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(
				test.method,
				test.path,
				nil,
			)

			recorder := httptest.NewRecorder()

			router.ServeHTTP(recorder, request)

			if recorder.Code != http.StatusNoContent {
				t.Fatalf(
					"expected status %d for %s %s, got %d",
					http.StatusNoContent,
					test.method,
					test.path,
					recorder.Code,
				)
			}
		})
	}
}

func TestRegisterRoutes_AllRoutesRequireAuthentication(t *testing.T) {
	router := newChatRouteTestRouter(rejectRouteTestMiddleware)

	tests := []struct {
		name   string
		method string
		path   string
	}{
		{
			name:   "create conversation",
			method: http.MethodPost,
			path:   "/chat/conversations",
		},
		{
			name:   "list conversations",
			method: http.MethodGet,
			path:   "/chat/conversations",
		},
		{
			name:   "get conversation",
			method: http.MethodGet,
			path:   "/chat/conversations/1",
		},
		{
			name:   "send message",
			method: http.MethodPost,
			path:   "/chat/conversations/1/messages",
		},
		{
			name:   "list messages",
			method: http.MethodGet,
			path:   "/chat/conversations/1/messages",
		},
		{
			name:   "mark conversation as read",
			method: http.MethodPost,
			path:   "/chat/conversations/1/read",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(
				test.method,
				test.path,
				nil,
			)

			recorder := httptest.NewRecorder()

			router.ServeHTTP(recorder, request)

			if recorder.Code != http.StatusUnauthorized {
				t.Fatalf(
					"expected status %d for %s %s, got %d",
					http.StatusUnauthorized,
					test.method,
					test.path,
					recorder.Code,
				)
			}
		})
	}
}

func TestRegisterRoutes_RegisteredMethodsAndPaths(t *testing.T) {
	router := chi.NewRouter()

	RegisterRoutes(
		router,
		nil,
		allowRouteTestMiddleware,
	)

	expectedRoutes := map[string]bool{
		http.MethodPost + " /chat/conversations":                           false,
		http.MethodGet + " /chat/conversations":                            false,
		http.MethodGet + " /chat/conversations/{conversationID}":           false,
		http.MethodPost + " /chat/conversations/{conversationID}/messages": false,
		http.MethodGet + " /chat/conversations/{conversationID}/messages":  false,
		http.MethodPost + " /chat/conversations/{conversationID}/read":     false,
	}

	err := chi.Walk(
		router,
		func(
			method string,
			route string,
			handler http.Handler,
			middlewares ...func(http.Handler) http.Handler,
		) error {
			key := method + " " + route

			if _, exists := expectedRoutes[key]; exists {
				expectedRoutes[key] = true
			}

			return nil
		},
	)

	if err != nil {
		t.Fatalf("unexpected route walk error: %v", err)
	}

	for route, registered := range expectedRoutes {
		if !registered {
			t.Errorf("expected route to be registered: %s", route)
		}
	}
}

func TestRegisterRoutes_DoesNotRegisterUnsupportedMethods(t *testing.T) {
	router := chi.NewRouter()

	RegisterRoutes(
		router,
		nil,
		allowRouteTestMiddleware,
	)

	unsupportedRoutes := map[string]bool{
		http.MethodDelete + " /chat/conversations":                           true,
		http.MethodPost + " /chat/conversations/{conversationID}":            true,
		http.MethodDelete + " /chat/conversations/{conversationID}/messages": true,
		http.MethodGet + " /chat/conversations/{conversationID}/read":        true,
	}

	err := chi.Walk(
		router,
		func(
			method string,
			route string,
			handler http.Handler,
			middlewares ...func(http.Handler) http.Handler,
		) error {
			key := method + " " + route

			if unsupportedRoutes[key] {
				t.Errorf("unsupported route was registered: %s", key)
			}

			return nil
		},
	)

	if err != nil {
		t.Fatalf("unexpected route walk error: %v", err)
	}
}
