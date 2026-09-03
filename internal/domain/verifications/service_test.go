package verifications

import (
	"context"
	"errors"
	"testing"
)

type MockRepository struct {
	request  *VerificationRequest
	requests []VerificationRequest

	createErr      error
	getByIDErr     error
	listByUserErr  error
	listAllErr     error
	listPendingErr error
	reviewErr      error

	createdRequest *VerificationRequest

	lastReviewRequestID uint
	lastReviewStatus    string
	lastReviewNotes     string
	lastReviewAdminID   uint
}

func (m *MockRepository) Create(
	ctx context.Context,
	request *VerificationRequest,
) error {
	m.createdRequest = request
	return m.createErr
}

func (m *MockRepository) GetByID(
	ctx context.Context,
	id uint,
) (*VerificationRequest, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}

	return m.request, nil
}

func (m *MockRepository) ListByUserID(
	ctx context.Context,
	userID uint,
) ([]VerificationRequest, error) {
	if m.listByUserErr != nil {
		return nil, m.listByUserErr
	}

	return m.requests, nil
}

func (m *MockRepository) ListAll(
	ctx context.Context,
) ([]VerificationRequest, error) {
	if m.listAllErr != nil {
		return nil, m.listAllErr
	}

	return m.requests, nil
}

func (m *MockRepository) ListPending(
	ctx context.Context,
) ([]VerificationRequest, error) {
	if m.listPendingErr != nil {
		return nil, m.listPendingErr
	}

	return m.requests, nil
}

func (m *MockRepository) Review(
	ctx context.Context,
	requestID uint,
	status string,
	adminNotes string,
	adminID uint,
) error {
	m.lastReviewRequestID = requestID
	m.lastReviewStatus = status
	m.lastReviewNotes = adminNotes
	m.lastReviewAdminID = adminID

	return m.reviewErr
}

func TestNewService(t *testing.T) {
	repo := &MockRepository{}

	service := NewService(repo)

	if service == nil {
		t.Fatal("expected service")
	}

	if service.repo != repo {
		t.Fatal("expected repository assigned")
	}
}

func TestCreate_InvalidInput(t *testing.T) {
	tests := []struct {
		name   string
		userID uint
		req    CreateVerificationRequest
	}{
		{
			name:   "zero user",
			userID: 0,
			req: CreateVerificationRequest{
				VerificationType: "id_check",
				DocumentURL:      "https://test/doc",
			},
		},
		{
			name:   "empty type",
			userID: 1,
			req: CreateVerificationRequest{
				DocumentURL: "https://test/doc",
			},
		},
		{
			name:   "empty document",
			userID: 1,
			req: CreateVerificationRequest{
				VerificationType: "id_check",
			},
		},
		{
			name:   "whitespace values",
			userID: 1,
			req: CreateVerificationRequest{
				VerificationType: "   ",
				DocumentURL:      "   ",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service := NewService(
				&MockRepository{},
			)

			request, err := service.Create(
				context.Background(),
				test.userID,
				test.req,
			)

			if request != nil {
				t.Fatal("expected nil request")
			}

			if !errors.Is(
				err,
				ErrInvalidInput,
			) {
				t.Fatalf(
					"expected ErrInvalidInput got %v",
					err,
				)
			}
		})
	}
}

func TestCreate_InvalidVerificationType(
	t *testing.T,
) {
	service := NewService(
		&MockRepository{},
	)

	_, err := service.Create(
		context.Background(),
		1,
		CreateVerificationRequest{
			VerificationType: "passport_magic",
			DocumentURL:      "https://test/doc",
		},
	)

	if !errors.Is(
		err,
		ErrInvalidVerificationType,
	) {
		t.Fatalf(
			"expected ErrInvalidVerificationType got %v",
			err,
		)
	}
}

func TestCreate_Success(t *testing.T) {
	repo := &MockRepository{}
	service := NewService(repo)

	request, err := service.Create(
		context.Background(),
		5,
		CreateVerificationRequest{
			VerificationType: "  dbs_check  ",
			DocumentURL:      "  https://test/dbs.pdf  ",
		},
	)

	if err != nil {
		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}

	if request == nil {
		t.Fatal("expected request")
	}

	if request.UserID != 5 {
		t.Fatalf(
			"expected user 5 got %d",
			request.UserID,
		)
	}

	if request.VerificationType != "dbs_check" {
		t.Fatalf(
			"expected dbs_check got %q",
			request.VerificationType,
		)
	}

	if request.DocumentURL != "https://test/dbs.pdf" {
		t.Fatalf(
			"expected trimmed URL got %q",
			request.DocumentURL,
		)
	}

	if request.Status != "pending" {
		t.Fatalf(
			"expected pending got %q",
			request.Status,
		)
	}

	if repo.createdRequest != request {
		t.Fatal(
			"expected repository to receive request",
		)
	}
}

func TestCreate_RepositoryError(t *testing.T) {
	createErr := errors.New("create failed")

	service := NewService(
		&MockRepository{
			createErr: createErr,
		},
	)

	request, err := service.Create(
		context.Background(),
		1,
		CreateVerificationRequest{
			VerificationType: "id_check",
			DocumentURL:      "https://test/id",
		},
	)

	if request != nil {
		t.Fatal("expected nil request")
	}

	if !errors.Is(err, createErr) {
		t.Fatalf(
			"expected create error got %v",
			err,
		)
	}
}

func TestGetByID_InvalidInput(t *testing.T) {
	service := NewService(
		&MockRepository{},
	)

	tests := []struct {
		name      string
		requestID uint
		userID    uint
	}{
		{
			name:      "zero request ID",
			requestID: 0,
			userID:    1,
		},
		{
			name:      "zero user ID",
			requestID: 1,
			userID:    0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := service.GetByID(
				context.Background(),
				test.requestID,
				test.userID,
				"cleaner",
			)

			if !errors.Is(
				err,
				ErrInvalidInput,
			) {
				t.Fatalf(
					"expected ErrInvalidInput got %v",
					err,
				)
			}
		})
	}
}

func TestGetByID_RepositoryError(t *testing.T) {
	repositoryErr := errors.New("lookup failed")

	service := NewService(
		&MockRepository{
			getByIDErr: repositoryErr,
		},
	)

	_, err := service.GetByID(
		context.Background(),
		1,
		1,
		"cleaner",
	)

	if !errors.Is(err, repositoryErr) {
		t.Fatalf(
			"expected repository error got %v",
			err,
		)
	}
}

func TestGetByID_Forbidden(t *testing.T) {
	service := NewService(
		&MockRepository{
			request: &VerificationRequest{
				ID:     1,
				UserID: 99,
			},
		},
	)

	request, err := service.GetByID(
		context.Background(),
		1,
		1,
		"cleaner",
	)

	if request != nil {
		t.Fatal("expected nil request")
	}

	if !errors.Is(err, ErrForbidden) {
		t.Fatalf(
			"expected ErrForbidden got %v",
			err,
		)
	}
}

func TestGetByID_OwnerSuccess(t *testing.T) {
	expected := &VerificationRequest{
		ID:     1,
		UserID: 5,
	}

	service := NewService(
		&MockRepository{
			request: expected,
		},
	)

	result, err := service.GetByID(
		context.Background(),
		1,
		5,
		"cleaner",
	)

	if err != nil {
		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}

	if result != expected {
		t.Fatal("expected request")
	}
}

func TestGetByID_AdminCanAccessAnyRequest(
	t *testing.T,
) {
	expected := &VerificationRequest{
		ID:     1,
		UserID: 99,
	}

	service := NewService(
		&MockRepository{
			request: expected,
		},
	)

	result, err := service.GetByID(
		context.Background(),
		1,
		5,
		"admin",
	)

	if err != nil {
		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}

	if result != expected {
		t.Fatal("expected request")
	}
}

func TestListMine_InvalidUser(t *testing.T) {
	service := NewService(
		&MockRepository{},
	)

	_, err := service.ListMine(
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

func TestListMine_Success(t *testing.T) {
	expected := []VerificationRequest{
		{
			ID:     1,
			UserID: 5,
		},
	}

	service := NewService(
		&MockRepository{
			requests: expected,
		},
	)

	result, err := service.ListMine(
		context.Background(),
		5,
	)

	if err != nil {
		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}

	if len(result) != 1 {
		t.Fatalf(
			"expected one request got %d",
			len(result),
		)
	}
}

func TestListMine_RepositoryError(t *testing.T) {
	listErr := errors.New("list failed")

	service := NewService(
		&MockRepository{
			listByUserErr: listErr,
		},
	)

	_, err := service.ListMine(
		context.Background(),
		5,
	)

	if !errors.Is(err, listErr) {
		t.Fatalf(
			"expected list error got %v",
			err,
		)
	}
}

func TestListAll_Success(t *testing.T) {
	service := NewService(
		&MockRepository{
			requests: []VerificationRequest{
				{ID: 1},
				{ID: 2},
			},
		},
	)

	result, err := service.ListAll(
		context.Background(),
	)

	if err != nil {
		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}

	if len(result) != 2 {
		t.Fatalf(
			"expected 2 requests got %d",
			len(result),
		)
	}
}

func TestListAll_Error(t *testing.T) {
	listErr := errors.New("list all failed")

	service := NewService(
		&MockRepository{
			listAllErr: listErr,
		},
	)

	_, err := service.ListAll(
		context.Background(),
	)

	if !errors.Is(err, listErr) {
		t.Fatalf(
			"expected list error got %v",
			err,
		)
	}
}

func TestListPending_Success(t *testing.T) {
	service := NewService(
		&MockRepository{
			requests: []VerificationRequest{
				{
					ID:     1,
					Status: "pending",
				},
			},
		},
	)

	result, err := service.ListPending(
		context.Background(),
	)

	if err != nil {
		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}

	if len(result) != 1 {
		t.Fatalf(
			"expected one request got %d",
			len(result),
		)
	}
}

func TestListPending_Error(t *testing.T) {
	listErr := errors.New(
		"list pending failed",
	)

	service := NewService(
		&MockRepository{
			listPendingErr: listErr,
		},
	)

	_, err := service.ListPending(
		context.Background(),
	)

	if !errors.Is(err, listErr) {
		t.Fatalf(
			"expected list error got %v",
			err,
		)
	}
}

func TestReview_InvalidInput(t *testing.T) {
	service := NewService(
		&MockRepository{},
	)

	tests := []struct {
		name      string
		requestID uint
		adminID   uint
	}{
		{
			name:      "zero request ID",
			requestID: 0,
			adminID:   1,
		},
		{
			name:      "zero admin ID",
			requestID: 1,
			adminID:   0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := service.Review(
				context.Background(),
				test.requestID,
				test.adminID,
				ReviewVerificationRequest{
					Status: "approved",
				},
			)

			if !errors.Is(
				err,
				ErrInvalidInput,
			) {
				t.Fatalf(
					"expected ErrInvalidInput got %v",
					err,
				)
			}
		})
	}
}

func TestReview_InvalidStatus(t *testing.T) {
	service := NewService(
		&MockRepository{},
	)

	_, err := service.Review(
		context.Background(),
		1,
		2,
		ReviewVerificationRequest{
			Status: "maybe",
		},
	)

	if !errors.Is(
		err,
		ErrInValidVerificationStatus,
	) {
		t.Fatalf(
			"expected invalid status error got %v",
			err,
		)
	}
}

func TestReview_RepositoryError(t *testing.T) {
	reviewErr := errors.New("review failed")

	service := NewService(
		&MockRepository{
			reviewErr: reviewErr,
		},
	)

	_, err := service.Review(
		context.Background(),
		1,
		2,
		ReviewVerificationRequest{
			Status:     "approved",
			AdminNotes: "Looks good",
		},
	)

	if !errors.Is(err, reviewErr) {
		t.Fatalf(
			"expected review error got %v",
			err,
		)
	}
}

func TestReview_Success(t *testing.T) {
	expected := &VerificationRequest{
		ID:         1,
		Status:     "approved",
		AdminNotes: "Verified",
	}

	repo := &MockRepository{
		request: expected,
	}

	service := NewService(repo)

	result, err := service.Review(
		context.Background(),
		1,
		99,
		ReviewVerificationRequest{
			Status:     "  approved ",
			AdminNotes: "  Verified  ",
		},
	)

	if err != nil {
		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}

	if repo.lastReviewRequestID != 1 {
		t.Fatalf(
			"expected request 1 got %d",
			repo.lastReviewRequestID,
		)
	}

	if repo.lastReviewAdminID != 99 {
		t.Fatalf(
			"expected admin 99 got %d",
			repo.lastReviewAdminID,
		)
	}

	if repo.lastReviewStatus != "approved" {
		t.Fatalf(
			"expected approved got %q",
			repo.lastReviewStatus,
		)
	}

	if repo.lastReviewNotes != "Verified" {
		t.Fatalf(
			"expected trimmed notes got %q",
			repo.lastReviewNotes,
		)
	}

	if result != expected {
		t.Fatal(
			"expected refreshed request",
		)
	}
}

func TestAllowedVerificationTypes(t *testing.T) {
	valid := []string{
		"id_check",
		"dbs_check",
		"insurance",
	}

	for _, value := range valid {
		if !isAllowedVerficationType(value) {
			t.Fatalf(
				"expected %q to be allowed",
				value,
			)
		}
	}

	if isAllowedVerficationType("passport") {
		t.Fatal(
			"expected passport to be rejected",
		)
	}
}

func TestAllowedVerificationStatuses(
	t *testing.T,
) {
	valid := []string{
		"pending",
		"approved",
		"rejected",
	}

	for _, value := range valid {
		if !isAllowedVerificationStatus(value) {
			t.Fatalf(
				"expected %q allowed",
				value,
			)
		}
	}

	if isAllowedVerificationStatus("maybe") {
		t.Fatal(
			"expected maybe to be rejected",
		)
	}
}

var _ Repository = (*MockRepository)(nil)
