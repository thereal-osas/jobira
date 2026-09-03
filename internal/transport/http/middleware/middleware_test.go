package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/rodrigueghenda/jobira/internal/security/identity"
	jwtsecurity "github.com/rodrigueghenda/jobira/internal/security/jwt"
)

func TestAuth_Success(t *testing.T) {
	issuer := jwtsecurity.NewIssuer("test-secret")

	token, err := issuer.Generate(
		42,
		"user@example.com",
		"cleaner",
	)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	nextCalled := false

	next := http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			nextCalled = true

			currentUser, err := identity.FromContext(
				r.Context(),
			)
			if err != nil {
				t.Fatalf(
					"expected identity: %v",
					err,
				)
			}

			if currentUser.UserID != 42 {
				t.Fatalf(
					"expected user id 42, got %d",
					currentUser.UserID,
				)
			}

			if currentUser.Email != "user@example.com" {
				t.Fatalf(
					"unexpected email %q",
					currentUser.Email,
				)
			}

			if currentUser.Role != "cleaner" {
				t.Fatalf(
					"unexpected role %q",
					currentUser.Role,
				)
			}

			w.WriteHeader(http.StatusOK)
		},
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/protected",
		nil,
	)
	req.Header.Set(
		"Authorization",
		"Bearer "+token,
	)

	rec := httptest.NewRecorder()

	Auth(issuer)(next).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected 200, got %d",
			rec.Code,
		)
	}

	if !nextCalled {
		t.Fatal("expected next handler")
	}
}

func TestAuth_MissingAuthorizationHeader(
	t *testing.T,
) {
	issuer := jwtsecurity.NewIssuer("secret")

	nextCalled := false

	next := http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			nextCalled = true
		},
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/protected",
		nil,
	)

	rec := httptest.NewRecorder()

	Auth(issuer)(next).ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected 401, got %d",
			rec.Code,
		)
	}

	if nextCalled {
		t.Fatal("next handler must not run")
	}
}

func TestAuth_InvalidAuthorizationHeader(
	t *testing.T,
) {
	issuer := jwtsecurity.NewIssuer("secret")

	tests := []string{
		"invalid",
		"Basic abc123",
		"Bearer",
	}

	for _, header := range tests {
		t.Run(header, func(t *testing.T) {
			req := httptest.NewRequest(
				http.MethodGet,
				"/protected",
				nil,
			)
			req.Header.Set(
				"Authorization",
				header,
			)

			rec := httptest.NewRecorder()

			Auth(issuer)(
				http.HandlerFunc(
					func(
						http.ResponseWriter,
						*http.Request,
					) {
						t.Fatal(
							"next should not run",
						)
					},
				),
			).ServeHTTP(rec, req)

			if rec.Code != http.StatusUnauthorized {
				t.Fatalf(
					"expected 401, got %d",
					rec.Code,
				)
			}
		})
	}
}

func TestAuth_InvalidToken(t *testing.T) {
	issuer := jwtsecurity.NewIssuer("secret")

	req := httptest.NewRequest(
		http.MethodGet,
		"/protected",
		nil,
	)
	req.Header.Set(
		"Authorization",
		"Bearer invalid-token",
	)

	rec := httptest.NewRecorder()

	Auth(issuer)(
		http.HandlerFunc(
			func(
				http.ResponseWriter,
				*http.Request,
			) {
				t.Fatal("next should not run")
			},
		),
	).ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected 401, got %d",
			rec.Code,
		)
	}
}

func TestAuth_WrongSecret(t *testing.T) {
	signingIssuer := jwtsecurity.NewIssuer(
		"signing-secret",
	)

	token, err := signingIssuer.Generate(
		1,
		"user@example.com",
		"client",
	)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	parsingIssuer := jwtsecurity.NewIssuer(
		"different-secret",
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/protected",
		nil,
	)
	req.Header.Set(
		"Authorization",
		"Bearer "+token,
	)

	rec := httptest.NewRecorder()

	Auth(parsingIssuer)(
		http.HandlerFunc(
			func(
				http.ResponseWriter,
				*http.Request,
			) {
				t.Fatal("next should not run")
			},
		),
	).ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected 401, got %d",
			rec.Code,
		)
	}
}

func TestRequireRole_Success(t *testing.T) {
	nextCalled := false

	next := http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			nextCalled = true
			w.WriteHeader(http.StatusOK)
		},
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/admin",
		nil,
	)

	ctx := identity.WithUser(
		req.Context(),
		identity.UserIdentity{
			UserID: 1,
			Role:   "admin",
		},
	)

	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()

	RequireRole("admin")(next).
		ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected 200, got %d",
			rec.Code,
		)
	}

	if !nextCalled {
		t.Fatal("expected next handler")
	}
}

func TestRequireRole_Unauthorized(t *testing.T) {
	nextCalled := false

	req := httptest.NewRequest(
		http.MethodGet,
		"/admin",
		nil,
	)

	rec := httptest.NewRecorder()

	RequireRole("admin")(
		http.HandlerFunc(
			func(
				http.ResponseWriter,
				*http.Request,
			) {
				nextCalled = true
			},
		),
	).ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected 401, got %d",
			rec.Code,
		)
	}

	if nextCalled {
		t.Fatal("next handler must not run")
	}
}

func TestRequireRole_Forbidden(t *testing.T) {
	nextCalled := false

	req := httptest.NewRequest(
		http.MethodGet,
		"/admin",
		nil,
	)

	ctx := identity.WithUser(
		req.Context(),
		identity.UserIdentity{
			UserID: 1,
			Role:   "cleaner",
		},
	)

	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()

	RequireRole("admin")(
		http.HandlerFunc(
			func(
				http.ResponseWriter,
				*http.Request,
			) {
				nextCalled = true
			},
		),
	).ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf(
			"expected 403, got %d",
			rec.Code,
		)
	}

	if nextCalled {
		t.Fatal("next handler must not run")
	}
}

func TestRequestID(t *testing.T) {
	next := http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			requestID, ok := r.Context().
				Value(RequestIDKey).(string)

			if !ok {
				t.Fatal(
					"expected request id in context",
				)
			}

			if requestID == "" {
				t.Fatal(
					"expected non-empty request id",
				)
			}

			header := w.Header().Get(
				"X-Request-ID",
			)

			if header != requestID {
				t.Fatalf(
					"expected header %q, got %q",
					requestID,
					header,
				)
			}

			w.WriteHeader(http.StatusOK)
		},
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/",
		nil,
	)

	rec := httptest.NewRecorder()

	RequestID(next).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected 200, got %d",
			rec.Code,
		)
	}
}

func TestRequestID_GeneratesUniqueIDs(
	t *testing.T,
) {
	var ids []string

	next := http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			requestID, _ := r.Context().
				Value(RequestIDKey).(string)

			ids = append(ids, requestID)
		},
	)

	handler := RequestID(next)

	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(
			http.MethodGet,
			"/",
			nil,
		)

		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)
	}

	if len(ids) != 2 {
		t.Fatalf(
			"expected 2 ids, got %d",
			len(ids),
		)
	}

	if ids[0] == ids[1] {
		t.Fatal(
			"expected unique request ids",
		)
	}
}

func TestRecover_NoPanic(t *testing.T) {
	next := http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusCreated)
		},
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/",
		nil,
	)

	rec := httptest.NewRecorder()

	Recover(next).ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf(
			"expected 201, got %d",
			rec.Code,
		)
	}
}

func TestRecover_Panic(t *testing.T) {
	next := http.HandlerFunc(
		func(
			http.ResponseWriter,
			*http.Request,
		) {
			panic("boom")
		},
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/",
		nil,
	)

	rec := httptest.NewRecorder()

	Recover(next).ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected 500, got %d",
			rec.Code,
		)
	}

	if !strings.Contains(
		rec.Body.String(),
		"internal server error",
	) {
		t.Fatalf(
			"unexpected body %q",
			rec.Body.String(),
		)
	}
}

func TestStatusRecorder_WriteHeader(
	t *testing.T,
) {
	rec := httptest.NewRecorder()

	statusRec := &statusRecorder{
		ResponseWriter: rec,
		statusCode:     http.StatusOK,
	}

	statusRec.WriteHeader(http.StatusCreated)

	if statusRec.statusCode != http.StatusCreated {
		t.Fatalf(
			"expected 201, got %d",
			statusRec.statusCode,
		)
	}

	if rec.Code != http.StatusCreated {
		t.Fatalf(
			"expected underlying 201, got %d",
			rec.Code,
		)
	}
}

func TestLogger_PassesRequest(t *testing.T) {
	nextCalled := false

	next := http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			nextCalled = true
			w.WriteHeader(http.StatusAccepted)
		},
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/test",
		nil,
	)

	rec := httptest.NewRecorder()

	Logger(next).ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf(
			"expected 202, got %d",
			rec.Code,
		)
	}

	if !nextCalled {
		t.Fatal("expected next handler")
	}
}

func TestRateLimiter_AllowsWithinLimit(
	t *testing.T,
) {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		limit:    2,
		window:   time.Minute,
	}

	calls := 0

	handler := rl.Middleware(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				calls++
				w.WriteHeader(http.StatusOK)
			},
		),
	)

	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(
			http.MethodGet,
			"/",
			nil,
		)
		req.RemoteAddr = "192.0.2.1:1234"

		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf(
				"request %d: expected 200, got %d",
				i+1,
				rec.Code,
			)
		}
	}

	if calls != 2 {
		t.Fatalf(
			"expected 2 calls, got %d",
			calls,
		)
	}
}

func TestRateLimiter_BlocksOverLimit(
	t *testing.T,
) {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		limit:    2,
		window:   time.Minute,
	}

	handler := rl.Middleware(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
		),
	)

	for i := 1; i <= 3; i++ {
		req := httptest.NewRequest(
			http.MethodGet,
			"/",
			nil,
		)
		req.RemoteAddr = "192.0.2.1:1234"

		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if i <= 2 {
			if rec.Code != http.StatusOK {
				t.Fatalf(
					"request %d: expected 200, got %d",
					i,
					rec.Code,
				)
			}
			continue
		}

		if rec.Code != http.StatusTooManyRequests {
			t.Fatalf(
				"request 3: expected 429, got %d",
				rec.Code,
			)
		}
	}
}

func TestRateLimiter_UsesIPNotPort(
	t *testing.T,
) {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		limit:    1,
		window:   time.Minute,
	}

	handler := rl.Middleware(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
		),
	)

	first := httptest.NewRequest(
		http.MethodGet,
		"/",
		nil,
	)
	first.RemoteAddr = "192.0.2.10:1111"

	firstRec := httptest.NewRecorder()
	handler.ServeHTTP(firstRec, first)

	if firstRec.Code != http.StatusOK {
		t.Fatalf(
			"expected first request 200, got %d",
			firstRec.Code,
		)
	}

	second := httptest.NewRequest(
		http.MethodGet,
		"/",
		nil,
	)
	second.RemoteAddr = "192.0.2.10:2222"

	secondRec := httptest.NewRecorder()
	handler.ServeHTTP(secondRec, second)

	if secondRec.Code != http.StatusTooManyRequests {
		t.Fatalf(
			"expected second request 429, got %d",
			secondRec.Code,
		)
	}
}

func TestRateLimiter_ResetWindow(t *testing.T) {
	rl := &RateLimiter{
		visitors: map[string]*visitor{
			"192.0.2.1": {
				request:   100,
				lastSeen:  time.Now().Add(-time.Hour),
				resetTime: time.Now().Add(-time.Minute),
			},
		},
		limit:  1,
		window: time.Minute,
	}

	req := httptest.NewRequest(
		http.MethodGet,
		"/",
		nil,
	)
	req.RemoteAddr = "192.0.2.1:1234"

	rec := httptest.NewRecorder()

	rl.Middleware(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
		),
	).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected reset request 200, got %d",
			rec.Code,
		)
	}

	v := rl.visitors["192.0.2.1"]

	if v == nil {
		t.Fatal("expected visitor")
	}

	if v.request != 1 {
		t.Fatalf(
			"expected request count 1, got %d",
			v.request,
		)
	}
}

func TestRateLimiter_RemoteAddrWithoutPort(
	t *testing.T,
) {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		limit:    1,
		window:   time.Minute,
	}

	req := httptest.NewRequest(
		http.MethodGet,
		"/",
		nil,
	)
	req.RemoteAddr = "192.0.2.20"

	rec := httptest.NewRecorder()

	rl.Middleware(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
		),
	).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected 200, got %d",
			rec.Code,
		)
	}

	if _, exists := rl.visitors["192.0.2.20"]; !exists {
		t.Fatal(
			"expected raw remote address fallback",
		)
	}
}

func TestMiddleware_ContextPreserved(
	t *testing.T,
) {
	type key string

	const testKey key = "test"

	req := httptest.NewRequest(
		http.MethodGet,
		"/",
		nil,
	)

	req = req.WithContext(
		context.WithValue(
			req.Context(),
			testKey,
			"value",
		),
	)

	rec := httptest.NewRecorder()

	RequestID(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				if got := r.Context().Value(testKey); got != "value" {
					t.Fatalf(
						"expected context value, got %v",
						got,
					)
				}
			},
		),
	).ServeHTTP(rec, req)
}

func TestRecover_ResponseIsJSON(t *testing.T) {
	req := httptest.NewRequest(
		http.MethodGet,
		"/",
		nil,
	)

	rec := httptest.NewRecorder()

	Recover(
		http.HandlerFunc(
			func(
				http.ResponseWriter,
				*http.Request,
			) {
				panic("boom")
			},
		),
	).ServeHTTP(rec, req)

	if got := rec.Header().Get(
		"Content-Type",
	); got != "application/json" {
		t.Fatalf(
			"expected application/json, got %q",
			got,
		)
	}

	var body map[string]string

	if err := json.Unmarshal(
		rec.Body.Bytes(),
		&body,
	); err != nil {
		t.Fatalf(
			"invalid JSON response: %v",
			err,
		)
	}

	if body["error"] != "internal server error" {
		t.Fatalf(
			"unexpected error %q",
			body["error"],
		)
	}
}
