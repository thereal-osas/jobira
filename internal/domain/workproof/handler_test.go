package wokrproof

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/rodrigueghenda/jobira/internal/security/identity"
)

func workProofRequestWithUser(
	req *http.Request,
	userID uint,
	role string,
) *http.Request {
	ctx := identity.WithUser(
		req.Context(),
		identity.UserIdentity{
			UserID: userID,
			Role:   role,
		},
	)

	return req.WithContext(ctx)
}

func withWorkProofURLParam(
	req *http.Request,
	key string,
	value string,
) *http.Request {
	routeCtx := chi.NewRouteContext()

	routeCtx.URLParams.Add(
		key,
		value,
	)

	ctx := context.WithValue(
		req.Context(),
		chi.RouteCtxKey,
		routeCtx,
	)

	return req.WithContext(ctx)
}

func TestHandler_Create_Success(t *testing.T) {
	repo := &mockRepository{
		getBookingFn: func(
			ctx context.Context,
			bookingID uint,
		) (*BookingSnapshot, error) {
			return &BookingSnapshot{
				ID:        10,
				JobID:     7,
				ClientID:  5,
				CleanerID: 8,
				Status:    "in_progress",
			}, nil
		},

		countByBookingAndTypeFn: func(
			ctx context.Context,
			bookingID uint,
			proofType ProofType,
		) (int, error) {
			return 0, nil
		},

		createFn: func(
			ctx context.Context,
			proof *WorkProof,
		) error {
			proof.ID = 30
			proof.CreatedAt = time.Now().UTC()

			return nil
		},
	}

	handler := NewHandler(
		NewService(repo),
	)

	body := []byte(`{
		"proof_type":"before",
		"photo_url":"https://example.com/before.jpg",
		"caption":"Kitchen before cleaning"
	}`)

	req := httptest.NewRequest(
		http.MethodPost,
		"/work-proof/bookings/10",
		bytes.NewReader(body),
	)

	req = workProofRequestWithUser(
		req,
		8,
		"cleaner",
	)

	req = withWorkProofURLParam(
		req,
		"bookingID",
		"10",
	)

	recorder := httptest.NewRecorder()

	handler.Create(
		recorder,
		req,
	)

	if recorder.Code != http.StatusCreated {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusCreated,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Create_Unauthorized(t *testing.T) {
	handler := NewHandler(
		NewService(
			&mockRepository{},
		),
	)

	body := []byte(`{
		"proof_type":"before",
		"photo_url":"https://example.com/before.jpg"
	}`)

	req := httptest.NewRequest(
		http.MethodPost,
		"/work-proof/bookings/10",
		bytes.NewReader(body),
	)

	req = withWorkProofURLParam(
		req,
		"bookingID",
		"10",
	)

	recorder := httptest.NewRecorder()

	handler.Create(
		recorder,
		req,
	)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusUnauthorized,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Create_ForbiddenRole(t *testing.T) {
	handler := NewHandler(
		NewService(
			&mockRepository{},
		),
	)

	body := []byte(`{
		"proof_type":"before",
		"photo_url":"https://example.com/before.jpg"
	}`)

	req := httptest.NewRequest(
		http.MethodPost,
		"/work-proof/bookings/10",
		bytes.NewReader(body),
	)

	req = workProofRequestWithUser(
		req,
		5,
		"client",
	)

	req = withWorkProofURLParam(
		req,
		"bookingID",
		"10",
	)

	recorder := httptest.NewRecorder()

	handler.Create(
		recorder,
		req,
	)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusForbidden,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Create_InvalidBookingID(t *testing.T) {
	handler := NewHandler(
		NewService(
			&mockRepository{},
		),
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/work-proof/bookings/nope",
		bytes.NewBufferString(`{}`),
	)

	req = workProofRequestWithUser(
		req,
		8,
		"cleaner",
	)

	req = withWorkProofURLParam(
		req,
		"bookingID",
		"nope",
	)

	recorder := httptest.NewRecorder()

	handler.Create(
		recorder,
		req,
	)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusBadRequest,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Create_InvalidBody(t *testing.T) {
	handler := NewHandler(
		NewService(
			&mockRepository{},
		),
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/work-proof/bookings/10",
		bytes.NewBufferString(`{invalid`),
	)

	req = workProofRequestWithUser(
		req,
		8,
		"cleaner",
	)

	req = withWorkProofURLParam(
		req,
		"bookingID",
		"10",
	)

	recorder := httptest.NewRecorder()

	handler.Create(
		recorder,
		req,
	)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusBadRequest,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Create_BookingNotEligible(t *testing.T) {
	repo := &mockRepository{
		getBookingFn: func(
			ctx context.Context,
			bookingID uint,
		) (*BookingSnapshot, error) {
			return &BookingSnapshot{
				ID:        10,
				ClientID:  5,
				CleanerID: 8,
				Status:    "pending",
			}, nil
		},
	}

	handler := NewHandler(
		NewService(repo),
	)

	body := []byte(`{
		"proof_type":"before",
		"photo_url":"https://example.com/before.jpg"
	}`)

	req := httptest.NewRequest(
		http.MethodPost,
		"/work-proof/bookings/10",
		bytes.NewReader(body),
	)

	req = workProofRequestWithUser(
		req,
		8,
		"cleaner",
	)

	req = withWorkProofURLParam(
		req,
		"bookingID",
		"10",
	)

	recorder := httptest.NewRecorder()

	handler.Create(
		recorder,
		req,
	)

	if recorder.Code != http.StatusConflict {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusConflict,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Create_InternalServerError(t *testing.T) {
	repo := &mockRepository{
		getBookingFn: func(
			ctx context.Context,
			bookingID uint,
		) (*BookingSnapshot, error) {
			return nil,
				errors.New("database failed")
		},
	}

	handler := NewHandler(
		NewService(repo),
	)

	body := []byte(`{
		"proof_type":"before",
		"photo_url":"https://example.com/before.jpg"
	}`)

	req := httptest.NewRequest(
		http.MethodPost,
		"/work-proof/bookings/10",
		bytes.NewReader(body),
	)

	req = workProofRequestWithUser(
		req,
		8,
		"cleaner",
	)

	req = withWorkProofURLParam(
		req,
		"bookingID",
		"10",
	)

	recorder := httptest.NewRecorder()

	handler.Create(
		recorder,
		req,
	)

	if recorder.Code !=
		http.StatusInternalServerError {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusInternalServerError,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_GetBookingProof_Success(t *testing.T) {
	now := time.Now().UTC()

	repo := &mockRepository{
		getBookingFn: func(
			ctx context.Context,
			bookingID uint,
		) (*BookingSnapshot, error) {
			return &BookingSnapshot{
				ID:        10,
				ClientID:  5,
				CleanerID: 8,
				Status:    "completed",
			}, nil
		},

		listByBookingIDFn: func(
			ctx context.Context,
			bookingID uint,
		) ([]WorkProof, error) {
			return []WorkProof{
				{
					ID:        1,
					BookingID: 10,
					CleanerID: 8,
					ProofType: ProofTypeBefore,
					PhotoURL:  "before.jpg",
					CreatedAt: now,
				},
				{
					ID:        2,
					BookingID: 10,
					CleanerID: 8,
					ProofType: ProofTypeAfter,
					PhotoURL:  "after.jpg",
					CreatedAt: now.Add(time.Hour),
				},
			}, nil
		},
	}

	handler := NewHandler(
		NewService(repo),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/work-proof/bookings/10",
		nil,
	)

	req = workProofRequestWithUser(
		req,
		5,
		"client",
	)

	req = withWorkProofURLParam(
		req,
		"bookingID",
		"10",
	)

	recorder := httptest.NewRecorder()

	handler.GetBookingProof(
		recorder,
		req,
	)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusOK,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_GetBookingProof_Forbidden(t *testing.T) {
	repo := &mockRepository{
		getBookingFn: func(
			ctx context.Context,
			bookingID uint,
		) (*BookingSnapshot, error) {
			return &BookingSnapshot{
				ID:        10,
				ClientID:  5,
				CleanerID: 8,
			}, nil
		},
	}

	handler := NewHandler(
		NewService(repo),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/work-proof/bookings/10",
		nil,
	)

	req = workProofRequestWithUser(
		req,
		99,
		"client",
	)

	req = withWorkProofURLParam(
		req,
		"bookingID",
		"10",
	)

	recorder := httptest.NewRecorder()

	handler.GetBookingProof(
		recorder,
		req,
	)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusForbidden,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Delete_Success(t *testing.T) {
	repo := &mockRepository{
		getByIDFn: func(
			ctx context.Context,
			proofID uint,
		) (*WorkProof, error) {
			return &WorkProof{
				ID:        30,
				BookingID: 10,
				CleanerID: 8,
			}, nil
		},

		deleteFn: func(
			ctx context.Context,
			proofID uint,
			cleanerID uint,
		) error {
			return nil
		},
	}

	handler := NewHandler(
		NewService(repo),
	)

	req := httptest.NewRequest(
		http.MethodDelete,
		"/work-proof/30",
		nil,
	)

	req = workProofRequestWithUser(
		req,
		8,
		"cleaner",
	)

	req = withWorkProofURLParam(
		req,
		"proofID",
		"30",
	)

	recorder := httptest.NewRecorder()

	handler.Delete(
		recorder,
		req,
	)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusOK,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Delete_ForbiddenRole(t *testing.T) {
	handler := NewHandler(
		NewService(
			&mockRepository{},
		),
	)

	req := httptest.NewRequest(
		http.MethodDelete,
		"/work-proof/30",
		nil,
	)

	req = workProofRequestWithUser(
		req,
		5,
		"client",
	)

	req = withWorkProofURLParam(
		req,
		"proofID",
		"30",
	)

	recorder := httptest.NewRecorder()

	handler.Delete(
		recorder,
		req,
	)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusForbidden,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Delete_NotFound(t *testing.T) {
	repo := &mockRepository{
		getByIDFn: func(
			ctx context.Context,
			proofID uint,
		) (*WorkProof, error) {
			return nil,
				ErrProofNotFound
		},
	}

	handler := NewHandler(
		NewService(repo),
	)

	req := httptest.NewRequest(
		http.MethodDelete,
		"/work-proof/30",
		nil,
	)

	req = workProofRequestWithUser(
		req,
		8,
		"cleaner",
	)

	req = withWorkProofURLParam(
		req,
		"proofID",
		"30",
	)

	recorder := httptest.NewRecorder()

	handler.Delete(
		recorder,
		req,
	)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusNotFound,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_GetVerifiedWork_Success(t *testing.T) {
	repo := &mockRepository{
		getVerifiedWorkSummaryFn: func(
			ctx context.Context,
			cleanerID uint,
		) (*VerifiedWorkSummary, error) {
			return &VerifiedWorkSummary{
				CleanerID:    cleanerID,
				VerifiedJobs: 12,
				BeforePhotos: 20,
				AfterPhotos:  22,
				TotalPhotos:  42,
			}, nil
		},
	}

	handler := NewHandler(
		NewService(repo),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/work-proof/cleaners/8",
		nil,
	)

	req = withWorkProofURLParam(
		req,
		"cleanerID",
		"8",
	)

	recorder := httptest.NewRecorder()

	handler.GetVerifiedWork(
		recorder,
		req,
	)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusOK,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_GetVerifiedWork_InvalidCleanerID(
	t *testing.T,
) {
	handler := NewHandler(
		NewService(
			&mockRepository{},
		),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/work-proof/cleaners/nope",
		nil,
	)

	req = withWorkProofURLParam(
		req,
		"cleanerID",
		"nope",
	)

	recorder := httptest.NewRecorder()

	handler.GetVerifiedWork(
		recorder,
		req,
	)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusBadRequest,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}
