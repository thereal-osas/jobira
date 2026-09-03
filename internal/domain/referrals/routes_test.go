package referrals

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/rodrigueghenda/jobira/internal/security/identity"
)

func referralsAuthMiddleware(
	userID uint,
	role string,
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			ctx := identity.WithUser(
				r.Context(),
				identity.UserIdentity{
					UserID: userID,
					Role:   role,
				},
			)

			next.ServeHTTP(
				w,
				r.WithContext(ctx),
			)
		})
	}
}

func TestRegisterRoutes_CreateCode(t *testing.T) {
	repo := &mockRepository{
		createCodeFn: func(
			context.Context,
			*ReferralCode,
		) error {
			return nil
		},
	}

	handler := newHandlerForTest(repo)
	router := chi.NewRouter()

	RegisterRoutes(
		router,
		handler,
		referralsAuthMiddleware(5, "user"),
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/referrals/codes",
		strings.NewReader(`{
			"code":"welcome",
			"reward_type":"credit"
		}`),
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusCreated {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusCreated,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestRegisterRoutes_ListMyCodes(t *testing.T) {
	repo := &mockRepository{
		listCodesByUserIDFn: func(
			context.Context,
			uint,
		) ([]ReferralCode, error) {
			return []ReferralCode{}, nil
		},
	}

	handler := newHandlerForTest(repo)
	router := chi.NewRouter()

	RegisterRoutes(
		router,
		handler,
		referralsAuthMiddleware(5, "user"),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/referrals/codes/me",
		nil,
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusOK,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestRegisterRoutes_Redeem(t *testing.T) {
	repo := &mockRepository{
		getCodeByCodeFn: func(
			context.Context,
			string,
		) (*ReferralCode, error) {
			return &ReferralCode{
				ID:       12,
				UserID:   5,
				IsActive: true,
			}, nil
		},
		createRedemptionFn: func(
			context.Context,
			*ReferralRedemption,
		) error {
			return nil
		},
		incrementCodeUseFn: func(
			context.Context,
			uint,
		) error {
			return nil
		},
	}

	handler := newHandlerForTest(repo)
	router := chi.NewRouter()

	RegisterRoutes(
		router,
		handler,
		referralsAuthMiddleware(8, "user"),
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/referrals/redeem",
		strings.NewReader(`{
			"code":"welcome"
		}`),
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusCreated {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusCreated,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestRegisterRoutes_ListMyRedemptions(t *testing.T) {
	repo := &mockRepository{
		listRedemptionsByReferrerIDFn: func(
			context.Context,
			uint,
		) ([]ReferralRedemption, error) {
			return []ReferralRedemption{}, nil
		},
	}

	handler := newHandlerForTest(repo)
	router := chi.NewRouter()

	RegisterRoutes(
		router,
		handler,
		referralsAuthMiddleware(5, "user"),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/referrals/redemptions/me",
		nil,
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusOK,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}
