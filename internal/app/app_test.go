package app

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/rodrigueghenda/jobira/internal/config"
)

func newTestApp(t *testing.T) *App {
	t.Helper()

	database, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}

	t.Cleanup(func() {
		_ = database.Close()
	})

	cfg := config.Config{
		Port:      "8080",
		AppEnv:    "test",
		JWTSecret: "test-secret",
	}

	application, err := New(
		cfg,
		database,
	)
	if err != nil {
		t.Fatalf(
			"New returned error: %v",
			err,
		)
	}

	return application
}

func TestNew_Success(t *testing.T) {
	application := newTestApp(t)

	if application == nil {
		t.Fatal("expected application")
	}

	if application.DB == nil {
		t.Fatal("expected database")
	}

	if application.R == nil {
		t.Fatal("expected router")
	}

	if application.Cfg.AppEnv != "test" {
		t.Fatalf(
			"expected test environment, got %q",
			application.Cfg.AppEnv,
		)
	}

	if application.Cfg.JWTSecret != "test-secret" {
		t.Fatal(
			"expected config to be preserved",
		)
	}
}

func TestNew_NilDatabase(t *testing.T) {
	cfg := config.Config{
		JWTSecret: "test-secret",
	}

	application, err := New(
		cfg,
		nil,
	)

	if err == nil {
		t.Fatal(
			"expected database error",
		)
	}

	if application != nil {
		t.Fatal(
			"expected nil application",
		)
	}

	if err.Error() != "database is required" {
		t.Fatalf(
			"unexpected error %q",
			err.Error(),
		)
	}
}

func TestRouter_ReturnsRouter(t *testing.T) {
	application := newTestApp(t)

	handler := application.Router()

	if handler == nil {
		t.Fatal("expected router")
	}
}

func TestRootRoute(t *testing.T) {
	application := newTestApp(t)

	req := httptest.NewRequest(
		http.MethodGet,
		"/",
		nil,
	)

	rec := httptest.NewRecorder()

	application.Router().
		ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusOK,
			rec.Code,
		)
	}

	expected := "Jobira API is running"

	if strings.TrimSpace(rec.Body.String()) != expected {
		t.Fatalf(
			"expected %q, got %q",
			expected,
			rec.Body.String(),
		)
	}
}

func TestRootRoute_RequestIDMiddleware(
	t *testing.T,
) {
	application := newTestApp(t)

	req := httptest.NewRequest(
		http.MethodGet,
		"/",
		nil,
	)

	rec := httptest.NewRecorder()

	application.Router().
		ServeHTTP(rec, req)

	requestID := rec.Header().Get(
		"X-Request-ID",
	)

	if requestID == "" {
		t.Fatal(
			"expected X-Request-ID header",
		)
	}
}

func TestHealthRouteRegistered(
	t *testing.T,
) {
	application := newTestApp(t)

	req := httptest.NewRequest(
		http.MethodGet,
		"/health",
		nil,
	)

	rec := httptest.NewRecorder()

	application.Router().
		ServeHTTP(rec, req)

	if rec.Code == http.StatusNotFound {
		t.Fatal(
			"expected health route to be registered",
		)
	}
}

func TestUnknownRouteReturnsNotFound(
	t *testing.T,
) {
	application := newTestApp(t)

	req := httptest.NewRequest(
		http.MethodGet,
		"/definitely-not-a-jobira-route",
		nil,
	)

	rec := httptest.NewRecorder()

	application.Router().
		ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusNotFound,
			rec.Code,
		)
	}
}

func TestAuthRouteRegistered(t *testing.T) {
	application := newTestApp(t)

	req := httptest.NewRequest(
		http.MethodPost,
		"/auth/login",
		strings.NewReader(`{}`),
	)

	req.Header.Set(
		"Content-Type",
		"application/json",
	)

	rec := httptest.NewRecorder()

	application.Router().
		ServeHTTP(rec, req)

	if rec.Code == http.StatusNotFound {
		t.Fatal(
			"expected auth login route to be registered",
		)
	}
}
func TestProtectedRouteUsesAuthMiddleware(
	t *testing.T,
) {
	application := newTestApp(t)

	req := httptest.NewRequest(
		http.MethodGet,
		"/admin/dashboard",
		nil,
	)

	rec := httptest.NewRecorder()

	application.Router().
		ServeHTTP(rec, req)

	if rec.Code == http.StatusNotFound {
		t.Fatal(
			"expected admin dashboard route to be registered",
		)
	}

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected unauthorized status, got %d",
			rec.Code,
		)
	}
}
