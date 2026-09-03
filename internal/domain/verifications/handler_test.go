package verifications

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

func verificationsTestAuth(
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

func verificationsPassthroughAdmin(
	next http.Handler,
) http.Handler {
	return next
}

func newVerificationTestRouter(
	repo *MockRepository,
	userID uint,
	role string,
) *chi.Mux {
	service := NewService(repo)
	handler := NewHandler(service)

	router := chi.NewRouter()

	RegisterRoutes(
		router,
		handler,
		verificationsTestAuth(userID, role),
		verificationsPassthroughAdmin,
	)

	return router
}

func TestHandler_NewHandler(t *testing.T) {
	service := NewService(
		&MockRepository{},
	)

	handler := NewHandler(service)

	if handler == nil {
		t.Fatal("expected handler")
	}

	if handler.service != service {
		t.Fatal("expected service assigned")
	}
}

func TestHandler_Create_Success(t *testing.T) {
	repo := &MockRepository{}

	router := newVerificationTestRouter(
		repo,
		5,
		"cleaner",
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/verifications/",
		bytes.NewBufferString(`{
			"verification_type":"dbs_check",
			"document_url":"https://test/dbs.pdf"
		}`),
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

	if repo.createdRequest == nil {
		t.Fatal("expected verification request created")
	}

	if repo.createdRequest.UserID != 5 {
		t.Fatalf(
			"expected user ID 5 got %d",
			repo.createdRequest.UserID,
		)
	}

	if repo.createdRequest.Status != "pending" {
		t.Fatalf(
			"expected pending got %q",
			repo.createdRequest.Status,
		)
	}
}

func TestHandler_Create_Unauthorized(t *testing.T) {
	handler := NewHandler(
		NewService(
			&MockRepository{},
		),
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/verifications/",
		bytes.NewBufferString(`{}`),
	)

	res := httptest.NewRecorder()

	handler.Create(res, req)

	if res.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected 401 got %d",
			res.Code,
		)
	}
}

func TestHandler_Create_InvalidBody(t *testing.T) {
	router := newVerificationTestRouter(
		&MockRepository{},
		1,
		"cleaner",
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/verifications/",
		bytes.NewBufferString("{bad"),
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

func TestHandler_Create_InvalidInput(t *testing.T) {
	router := newVerificationTestRouter(
		&MockRepository{},
		1,
		"cleaner",
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/verifications/",
		bytes.NewBufferString(`{
			"verification_type":"",
			"document_url":""
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

func TestHandler_Create_InvalidVerificationType(
	t *testing.T,
) {
	router := newVerificationTestRouter(
		&MockRepository{},
		1,
		"cleaner",
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/verifications/",
		bytes.NewBufferString(`{
			"verification_type":"passport_magic",
			"document_url":"https://test/doc"
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

func TestHandler_Create_ServiceError(t *testing.T) {
	router := newVerificationTestRouter(
		&MockRepository{
			createErr: errors.New("database failed"),
		},
		1,
		"cleaner",
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/verifications/",
		bytes.NewBufferString(`{
			"verification_type":"id_check",
			"document_url":"https://test/doc"
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

func TestHandler_GetByID_Success(t *testing.T) {
	repo := &MockRepository{
		request: &VerificationRequest{
			ID:               10,
			UserID:           5,
			VerificationType: "dbs_check",
			Status:           "pending",
		},
	}

	router := newVerificationTestRouter(
		repo,
		5,
		"cleaner",
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/verifications/10",
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

	var request VerificationRequest

	if err := json.NewDecoder(
		res.Body,
	).Decode(&request); err != nil {
		t.Fatalf(
			"failed decoding response: %v",
			err,
		)
	}

	if request.ID != 10 {
		t.Fatalf(
			"expected ID 10 got %d",
			request.ID,
		)
	}
}

func TestHandler_GetByID_Unauthorized(t *testing.T) {
	handler := NewHandler(
		NewService(
			&MockRepository{},
		),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/verifications/10",
		nil,
	)

	res := httptest.NewRecorder()

	handler.GetByID(res, req)

	if res.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected 401 got %d",
			res.Code,
		)
	}
}

func TestHandler_GetByID_InvalidID(t *testing.T) {
	router := newVerificationTestRouter(
		&MockRepository{},
		1,
		"cleaner",
	)

	tests := []string{
		"/verifications/abc",
		"/verifications/0",
	}

	for _, path := range tests {
		t.Run(path, func(t *testing.T) {
			req := httptest.NewRequest(
				http.MethodGet,
				path,
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
		})
	}
}

func TestHandler_GetByID_NotFound(t *testing.T) {
	router := newVerificationTestRouter(
		&MockRepository{
			getByIDErr: ErrVerificationNotFound,
		},
		1,
		"cleaner",
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/verifications/10",
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

func TestHandler_GetByID_Forbidden(t *testing.T) {
	router := newVerificationTestRouter(
		&MockRepository{
			request: &VerificationRequest{
				ID:     10,
				UserID: 99,
			},
		},
		1,
		"cleaner",
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/verifications/10",
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

func TestHandler_GetByID_ServiceError(t *testing.T) {
	router := newVerificationTestRouter(
		&MockRepository{
			getByIDErr: errors.New("database failed"),
		},
		1,
		"cleaner",
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/verifications/10",
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

func TestHandler_ListMine_Success(t *testing.T) {
	repo := &MockRepository{
		requests: []VerificationRequest{
			{
				ID:     1,
				UserID: 5,
			},
		},
	}

	router := newVerificationTestRouter(
		repo,
		5,
		"cleaner",
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/verifications/me",
		nil,
	)

	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf(
			"expected 200 got %d",
			res.Code,
		)
	}

	var requests []VerificationRequest

	if err := json.NewDecoder(
		res.Body,
	).Decode(&requests); err != nil {
		t.Fatalf(
			"failed decoding response: %v",
			err,
		)
	}

	if len(requests) != 1 {
		t.Fatalf(
			"expected one request got %d",
			len(requests),
		)
	}
}

func TestHandler_ListMine_Unauthorized(t *testing.T) {
	handler := NewHandler(
		NewService(
			&MockRepository{},
		),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/verifications/me",
		nil,
	)

	res := httptest.NewRecorder()

	handler.ListMine(res, req)

	if res.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected 401 got %d",
			res.Code,
		)
	}
}

func TestHandler_ListMine_ServiceError(t *testing.T) {
	router := newVerificationTestRouter(
		&MockRepository{
			listByUserErr: errors.New(
				"database failed",
			),
		},
		5,
		"cleaner",
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/verifications/me",
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

func TestHandler_ListAll_Success(t *testing.T) {
	repo := &MockRepository{
		requests: []VerificationRequest{
			{
				ID: 1,
			},
			{
				ID: 2,
			},
		},
	}

	router := newVerificationTestRouter(
		repo,
		1,
		"admin",
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/admin/verifications/",
		nil,
	)

	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf(
			"expected 200 got %d",
			res.Code,
		)
	}
}

func TestHandler_ListAll_ServiceError(t *testing.T) {
	router := newVerificationTestRouter(
		&MockRepository{
			listAllErr: errors.New("database failed"),
		},
		1,
		"admin",
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/admin/verifications/",
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

func TestHandler_ListPending_Success(t *testing.T) {
	repo := &MockRepository{
		requests: []VerificationRequest{
			{
				ID:     1,
				Status: "pending",
			},
		},
	}

	router := newVerificationTestRouter(
		repo,
		1,
		"admin",
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/admin/verifications/pending",
		nil,
	)

	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf(
			"expected 200 got %d",
			res.Code,
		)
	}
}

func TestHandler_ListPending_ServiceError(
	t *testing.T,
) {
	router := newVerificationTestRouter(
		&MockRepository{
			listPendingErr: errors.New(
				"database failed",
			),
		},
		1,
		"admin",
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/admin/verifications/pending",
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

func TestHandler_Review_Success(t *testing.T) {
	repo := &MockRepository{
		request: &VerificationRequest{
			ID:     10,
			UserID: 5,
			Status: "approved",
		},
	}

	router := newVerificationTestRouter(
		repo,
		99,
		"admin",
	)

	req := httptest.NewRequest(
		http.MethodPatch,
		"/admin/verifications/10",
		bytes.NewBufferString(`{
			"status":"approved",
			"admin_notes":"verified"
		}`),
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

	if repo.lastReviewRequestID != 10 {
		t.Fatalf(
			"expected request ID 10 got %d",
			repo.lastReviewRequestID,
		)
	}

	if repo.lastReviewAdminID != 99 {
		t.Fatalf(
			"expected admin ID 99 got %d",
			repo.lastReviewAdminID,
		)
	}
}

func TestHandler_Review_Unauthorized(t *testing.T) {
	handler := NewHandler(
		NewService(
			&MockRepository{},
		),
	)

	req := httptest.NewRequest(
		http.MethodPatch,
		"/admin/verifications/10",
		bytes.NewBufferString(`{}`),
	)

	res := httptest.NewRecorder()

	handler.Review(res, req)

	if res.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected 401 got %d",
			res.Code,
		)
	}
}

func TestHandler_Review_InvalidID(t *testing.T) {
	router := newVerificationTestRouter(
		&MockRepository{},
		99,
		"admin",
	)

	tests := []string{
		"/admin/verifications/abc",
		"/admin/verifications/0",
	}

	for _, path := range tests {
		t.Run(path, func(t *testing.T) {
			req := httptest.NewRequest(
				http.MethodPatch,
				path,
				bytes.NewBufferString(`{
					"status":"approved"
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

func TestHandler_Review_InvalidBody(t *testing.T) {
	router := newVerificationTestRouter(
		&MockRepository{},
		99,
		"admin",
	)

	req := httptest.NewRequest(
		http.MethodPatch,
		"/admin/verifications/10",
		bytes.NewBufferString("{bad"),
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

func TestHandler_Review_InvalidStatus(t *testing.T) {
	router := newVerificationTestRouter(
		&MockRepository{},
		99,
		"admin",
	)

	req := httptest.NewRequest(
		http.MethodPatch,
		"/admin/verifications/10",
		bytes.NewBufferString(`{
			"status":"maybe",
			"admin_notes":"test"
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

func TestHandler_Review_NotFound(t *testing.T) {
	router := newVerificationTestRouter(
		&MockRepository{
			reviewErr: ErrVerificationNotFound,
		},
		99,
		"admin",
	)

	req := httptest.NewRequest(
		http.MethodPatch,
		"/admin/verifications/10",
		bytes.NewBufferString(`{
			"status":"approved",
			"admin_notes":"verified"
		}`),
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

func TestHandler_Review_ServiceError(t *testing.T) {
	router := newVerificationTestRouter(
		&MockRepository{
			reviewErr: errors.New("database failed"),
		},
		99,
		"admin",
	)

	req := httptest.NewRequest(
		http.MethodPatch,
		"/admin/verifications/10",
		bytes.NewBufferString(`{
			"status":"approved"
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

func TestParseIDParam_Valid(t *testing.T) {
	router := chi.NewRouter()

	router.Get(
		"/test/{requestID}",
		func(w http.ResponseWriter, r *http.Request) {
			id, err := parseIDParam(
				r,
				"requestID",
			)

			if err != nil {
				t.Fatalf(
					"unexpected error: %v",
					err,
				)
			}

			if id != 25 {
				t.Fatalf(
					"expected 25 got %d",
					id,
				)
			}

			w.WriteHeader(http.StatusOK)
		},
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/test/25",
		nil,
	)

	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf(
			"expected 200 got %d",
			res.Code,
		)
	}
}
