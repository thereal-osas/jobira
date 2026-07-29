package applications

import (
	"context"
	"errors"
	"testing"
)

type MockRepository struct {
	CreateFunc              func(context.Context, *Application) error
	GetByIDFunc             func(context.Context, uint) (*Application, error)
	ListByJobIDFunc         func(context.Context, uint) ([]Application, error)
	ListByCleanerIDFunc     func(context.Context, uint) ([]Application, error)
	UpdateStatusFunc        func(context.Context, uint, string) error
	GetJobClientIDFunc      func(context.Context, uint) (uint, error)
	GetJobTitleFunc         func(context.Context, uint) (string, error)
	GetUserEmailFunc        func(context.Context, uint) (string, error)
	GetCleanerEmailByIDFunc func(context.Context, uint) (string, error)
}

func (m *MockRepository) Create(ctx context.Context, a *Application) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, a)
	}
	return nil
}

func (m *MockRepository) GetByID(ctx context.Context, id uint) (*Application, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockRepository) ListByJobID(ctx context.Context, id uint) ([]Application, error) {
	if m.ListByJobIDFunc != nil {
		return m.ListByJobIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockRepository) ListByCleanerID(ctx context.Context, id uint) ([]Application, error) {
	if m.ListByCleanerIDFunc != nil {
		return m.ListByCleanerIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockRepository) UpdateStatus(ctx context.Context, id uint, status string) error {
	if m.UpdateStatusFunc != nil {
		return m.UpdateStatusFunc(ctx, id, status)
	}
	return nil
}

func (m *MockRepository) GetJobClientID(ctx context.Context, id uint) (uint, error) {
	if m.GetJobClientIDFunc != nil {
		return m.GetJobClientIDFunc(ctx, id)
	}
	return 0, nil
}

func (m *MockRepository) GetJobTitle(ctx context.Context, id uint) (string, error) {
	if m.GetJobTitleFunc != nil {
		return m.GetJobTitleFunc(ctx, id)
	}
	return "", nil
}

func (m *MockRepository) GetUserEmail(ctx context.Context, id uint) (string, error) {
	if m.GetUserEmailFunc != nil {
		return m.GetUserEmailFunc(ctx, id)
	}
	return "", nil
}

func (m *MockRepository) GetCleanerEmailByID(ctx context.Context, id uint) (string, error) {
	if m.GetCleanerEmailByIDFunc != nil {
		return m.GetCleanerEmailByIDFunc(ctx, id)
	}
	return "", nil
}

func TestNewService(t *testing.T) {
	repo := &MockRepository{}

	service := NewService(
		repo,
		nil,
		nil,
		nil,
		nil,
		nil,
	)

	if service == nil {
		t.Fatal("expected service")
	}

	if service.repo != repo {
		t.Fatal("repository not assigned")
	}
}

func TestService_Apply_Success(t *testing.T) {
	var createdApplication *Application

	repo := &MockRepository{
		GetJobClientIDFunc: func(ctx context.Context, jobID uint) (uint, error) {
			if jobID != 10 {
				t.Fatalf("expected job ID 10, got %d", jobID)
			}

			return 50, nil
		},
		CreateFunc: func(ctx context.Context, application *Application) error {
			createdApplication = application
			application.ID = 1
			return nil
		},
		GetJobTitleFunc: func(ctx context.Context, jobID uint) (string, error) {
			return "Office cleaning", nil
		},
	}

	service := NewService(repo, nil, nil, nil, nil, nil)

	result, err := service.Apply(
		context.Background(),
		10,
		20,
		CreateApplicationRequest{
			CoverMessage: "  I am available tomorrow.  ",
			ProposedRate: 75,
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected application")
	}

	if createdApplication == nil {
		t.Fatal("expected repository Create to be called")
	}

	if result.ID != 1 {
		t.Fatalf("expected ID 1, got %d", result.ID)
	}

	if result.JobID != 10 {
		t.Fatalf("expected job ID 10, got %d", result.JobID)
	}

	if result.CleanerID != 20 {
		t.Fatalf("expected cleaner ID 20, got %d", result.CleanerID)
	}

	if result.CoverMessage != "I am available tomorrow." {
		t.Fatalf(
			"expected trimmed cover message, got %q",
			result.CoverMessage,
		)
	}

	if result.ProposedRate != 75 {
		t.Fatalf(
			"expected proposed rate 75, got %d",
			result.ProposedRate,
		)
	}

	if result.Status != "pending" {
		t.Fatalf("expected pending status, got %q", result.Status)
	}
}

func TestService_Apply_InvalidInput(t *testing.T) {
	tests := []struct {
		name      string
		jobID     uint
		cleanerID uint
		request   CreateApplicationRequest
	}{
		{
			name:      "ZeroJobID",
			jobID:     0,
			cleanerID: 20,
			request: CreateApplicationRequest{
				CoverMessage: "Available",
				ProposedRate: 50,
			},
		},
		{
			name:      "ZeroCleanerID",
			jobID:     10,
			cleanerID: 0,
			request: CreateApplicationRequest{
				CoverMessage: "Available",
				ProposedRate: 50,
			},
		},
		{
			name:      "EmptyCoverMessage",
			jobID:     10,
			cleanerID: 20,
			request: CreateApplicationRequest{
				CoverMessage: "   ",
				ProposedRate: 50,
			},
		},
		{
			name:      "NegativeProposedRate",
			jobID:     10,
			cleanerID: 20,
			request: CreateApplicationRequest{
				CoverMessage: "Available",
				ProposedRate: -1,
			},
		},
	}

	service := NewService(
		&MockRepository{},
		nil,
		nil,
		nil,
		nil,
		nil,
	)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := service.Apply(
				context.Background(),
				tc.jobID,
				tc.cleanerID,
				tc.request,
			)

			if !errors.Is(err, ErrInvalidInput) {
				t.Fatalf(
					"expected ErrInvalidInput, got %v",
					err,
				)
			}

			if result != nil {
				t.Fatalf(
					"expected nil application, got %#v",
					result,
				)
			}
		})
	}
}

func TestService_Apply_GetJobClientIDError(t *testing.T) {
	expectedError := errors.New("failed to load job client")

	repo := &MockRepository{
		GetJobClientIDFunc: func(ctx context.Context, jobID uint) (uint, error) {
			return 0, expectedError
		},
	}

	service := NewService(repo, nil, nil, nil, nil, nil)

	result, err := service.Apply(
		context.Background(),
		10,
		20,
		CreateApplicationRequest{
			CoverMessage: "Available",
			ProposedRate: 50,
		},
	)

	if !errors.Is(err, expectedError) {
		t.Fatalf("expected %v, got %v", expectedError, err)
	}

	if result != nil {
		t.Fatalf("expected nil application, got %#v", result)
	}
}

func TestService_Apply_CannotApplyToOwnJob(t *testing.T) {
	createCalled := false

	repo := &MockRepository{
		GetJobClientIDFunc: func(ctx context.Context, jobID uint) (uint, error) {
			return 20, nil
		},
		CreateFunc: func(ctx context.Context, application *Application) error {
			createCalled = true
			return nil
		},
	}

	service := NewService(repo, nil, nil, nil, nil, nil)

	result, err := service.Apply(
		context.Background(),
		10,
		20,
		CreateApplicationRequest{
			CoverMessage: "Available",
			ProposedRate: 50,
		},
	)

	if !errors.Is(err, ErrCannotApplyToOwnJob) {
		t.Fatalf(
			"expected ErrCannotApplyToOwnJob, got %v",
			err,
		)
	}

	if result != nil {
		t.Fatalf("expected nil application, got %#v", result)
	}

	if createCalled {
		t.Fatal("repository Create should not be called")
	}
}

func TestService_Apply_CreateError(t *testing.T) {
	expectedError := errors.New("failed to create application")

	repo := &MockRepository{
		GetJobClientIDFunc: func(ctx context.Context, jobID uint) (uint, error) {
			return 50, nil
		},
		CreateFunc: func(ctx context.Context, application *Application) error {
			return expectedError
		},
	}

	service := NewService(repo, nil, nil, nil, nil, nil)

	result, err := service.Apply(
		context.Background(),
		10,
		20,
		CreateApplicationRequest{
			CoverMessage: "Available",
			ProposedRate: 50,
		},
	)

	if !errors.Is(err, expectedError) {
		t.Fatalf("expected %v, got %v", expectedError, err)
	}

	if result != nil {
		t.Fatalf("expected nil application, got %#v", result)
	}
}

func TestService_Apply_GetJobTitleErrorIsIgnored(t *testing.T) {
	expectedError := errors.New("failed to load job title")

	repo := &MockRepository{
		GetJobClientIDFunc: func(ctx context.Context, jobID uint) (uint, error) {
			return 50, nil
		},
		CreateFunc: func(ctx context.Context, application *Application) error {
			application.ID = 1
			return nil
		},
		GetJobTitleFunc: func(ctx context.Context, jobID uint) (string, error) {
			return "", expectedError
		},
	}

	service := NewService(repo, nil, nil, nil, nil, nil)

	result, err := service.Apply(
		context.Background(),
		10,
		20,
		CreateApplicationRequest{
			CoverMessage: "Available",
			ProposedRate: 50,
		},
	)
	if err != nil {
		t.Fatalf("expected title error to be ignored, got %v", err)
	}

	if result == nil {
		t.Fatal("expected application")
	}

	if result.ID != 1 {
		t.Fatalf("expected ID 1, got %d", result.ID)
	}
}

func TestService_GetByID_CleanerSuccess(t *testing.T) {
	expected := &Application{
		ID:        1,
		JobID:     10,
		CleanerID: 20,
		Status:    "pending",
	}

	repo := &MockRepository{
		GetByIDFunc: func(ctx context.Context, id uint) (*Application, error) {
			return expected, nil
		},
		GetJobClientIDFunc: func(ctx context.Context, jobID uint) (uint, error) {
			return 50, nil
		},
	}

	service := NewService(repo, nil, nil, nil, nil, nil)

	result, err := service.GetByID(
		context.Background(),
		1,
		20,
		"cleaner",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != expected {
		t.Fatal("expected returned application")
	}
}

func TestService_GetByID_ClientSuccess(t *testing.T) {
	expected := &Application{
		ID:        1,
		JobID:     10,
		CleanerID: 20,
	}

	repo := &MockRepository{
		GetByIDFunc: func(ctx context.Context, id uint) (*Application, error) {
			return expected, nil
		},
		GetJobClientIDFunc: func(ctx context.Context, jobID uint) (uint, error) {
			return 50, nil
		},
	}

	service := NewService(repo, nil, nil, nil, nil, nil)

	result, err := service.GetByID(
		context.Background(),
		1,
		50,
		"client",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != expected {
		t.Fatal("expected returned application")
	}
}

func TestService_GetByID_AdminSuccess(t *testing.T) {
	expected := &Application{
		ID:        1,
		JobID:     10,
		CleanerID: 20,
	}

	repo := &MockRepository{
		GetByIDFunc: func(ctx context.Context, id uint) (*Application, error) {
			return expected, nil
		},
		GetJobClientIDFunc: func(ctx context.Context, jobID uint) (uint, error) {
			return 50, nil
		},
	}

	service := NewService(repo, nil, nil, nil, nil, nil)

	result, err := service.GetByID(
		context.Background(),
		1,
		99,
		"admin",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != expected {
		t.Fatal("expected returned application")
	}
}

func TestService_GetByID_InvalidInput(t *testing.T) {
	tests := []struct {
		name   string
		id     uint
		userID uint
	}{
		{
			name:   "ZeroApplicationID",
			id:     0,
			userID: 20,
		},
		{
			name:   "ZeroUserID",
			id:     1,
			userID: 0,
		},
	}

	service := NewService(
		&MockRepository{},
		nil,
		nil,
		nil,
		nil,
		nil,
	)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := service.GetByID(
				context.Background(),
				tc.id,
				tc.userID,
				"cleaner",
			)

			if !errors.Is(err, ErrInvalidInput) {
				t.Fatalf("expected ErrInvalidInput, got %v", err)
			}

			if result != nil {
				t.Fatalf("expected nil result, got %#v", result)
			}
		})
	}
}

func TestService_GetByID_RepositoryError(t *testing.T) {
	expectedError := errors.New("failed to load application")

	repo := &MockRepository{
		GetByIDFunc: func(ctx context.Context, id uint) (*Application, error) {
			return nil, expectedError
		},
	}

	service := NewService(repo, nil, nil, nil, nil, nil)

	result, err := service.GetByID(
		context.Background(),
		1,
		20,
		"cleaner",
	)

	if !errors.Is(err, expectedError) {
		t.Fatalf("expected %v, got %v", expectedError, err)
	}

	if result != nil {
		t.Fatalf("expected nil result, got %#v", result)
	}
}

func TestService_GetByID_GetJobClientIDError(t *testing.T) {
	expectedError := errors.New("failed to load job owner")

	repo := &MockRepository{
		GetByIDFunc: func(ctx context.Context, id uint) (*Application, error) {
			return &Application{
				ID:        id,
				JobID:     10,
				CleanerID: 20,
			}, nil
		},
		GetJobClientIDFunc: func(ctx context.Context, jobID uint) (uint, error) {
			return 0, expectedError
		},
	}

	service := NewService(repo, nil, nil, nil, nil, nil)

	result, err := service.GetByID(
		context.Background(),
		1,
		20,
		"cleaner",
	)

	if !errors.Is(err, expectedError) {
		t.Fatalf("expected %v, got %v", expectedError, err)
	}

	if result != nil {
		t.Fatalf("expected nil result, got %#v", result)
	}
}

func TestService_GetByID_Forbidden(t *testing.T) {
	repo := &MockRepository{
		GetByIDFunc: func(ctx context.Context, id uint) (*Application, error) {
			return &Application{
				ID:        id,
				JobID:     10,
				CleanerID: 20,
			}, nil
		},
		GetJobClientIDFunc: func(ctx context.Context, jobID uint) (uint, error) {
			return 50, nil
		},
	}

	service := NewService(repo, nil, nil, nil, nil, nil)

	result, err := service.GetByID(
		context.Background(),
		1,
		999,
		"cleaner",
	)

	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}

	if result != nil {
		t.Fatalf("expected nil result, got %#v", result)
	}
}

func TestService_ListByJobID_Success(t *testing.T) {
	expected := []Application{
		{
			ID:        1,
			JobID:     10,
			CleanerID: 20,
		},
		{
			ID:        2,
			JobID:     10,
			CleanerID: 21,
		},
	}

	repo := &MockRepository{
		GetJobClientIDFunc: func(ctx context.Context, jobID uint) (uint, error) {
			return 50, nil
		},
		ListByJobIDFunc: func(ctx context.Context, jobID uint) ([]Application, error) {
			return expected, nil
		},
	}

	service := NewService(repo, nil, nil, nil, nil, nil)

	result, err := service.ListByJobID(
		context.Background(),
		10,
		50,
		"client",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 2 {
		t.Fatalf("expected 2 applications, got %d", len(result))
	}
}

func TestService_ListByJobID_AdminSuccess(t *testing.T) {
	repo := &MockRepository{
		GetJobClientIDFunc: func(ctx context.Context, jobID uint) (uint, error) {
			return 50, nil
		},
		ListByJobIDFunc: func(ctx context.Context, jobID uint) ([]Application, error) {
			return []Application{{ID: 1, JobID: jobID}}, nil
		},
	}

	service := NewService(repo, nil, nil, nil, nil, nil)

	result, err := service.ListByJobID(
		context.Background(),
		10,
		999,
		"admin",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("expected 1 application, got %d", len(result))
	}
}

func TestService_ListByJobID_InvalidInput(t *testing.T) {
	tests := []struct {
		name   string
		jobID  uint
		userID uint
	}{
		{
			name:   "ZeroJobID",
			jobID:  0,
			userID: 50,
		},
		{
			name:   "ZeroUserID",
			jobID:  10,
			userID: 0,
		},
	}

	service := NewService(
		&MockRepository{},
		nil,
		nil,
		nil,
		nil,
		nil,
	)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := service.ListByJobID(
				context.Background(),
				tc.jobID,
				tc.userID,
				"client",
			)

			if !errors.Is(err, ErrInvalidInput) {
				t.Fatalf("expected ErrInvalidInput, got %v", err)
			}

			if result != nil {
				t.Fatalf("expected nil result, got %#v", result)
			}
		})
	}
}

func TestService_ListByJobID_GetJobClientIDError(t *testing.T) {
	expectedError := errors.New("failed to load client")

	repo := &MockRepository{
		GetJobClientIDFunc: func(ctx context.Context, jobID uint) (uint, error) {
			return 0, expectedError
		},
	}

	service := NewService(repo, nil, nil, nil, nil, nil)

	result, err := service.ListByJobID(
		context.Background(),
		10,
		50,
		"client",
	)

	if !errors.Is(err, expectedError) {
		t.Fatalf("expected %v, got %v", expectedError, err)
	}

	if result != nil {
		t.Fatalf("expected nil result, got %#v", result)
	}
}

func TestService_ListByJobID_Forbidden(t *testing.T) {
	listCalled := false

	repo := &MockRepository{
		GetJobClientIDFunc: func(ctx context.Context, jobID uint) (uint, error) {
			return 50, nil
		},
		ListByJobIDFunc: func(ctx context.Context, jobID uint) ([]Application, error) {
			listCalled = true
			return nil, nil
		},
	}

	service := NewService(repo, nil, nil, nil, nil, nil)

	result, err := service.ListByJobID(
		context.Background(),
		10,
		999,
		"client",
	)

	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}

	if result != nil {
		t.Fatalf("expected nil result, got %#v", result)
	}

	if listCalled {
		t.Fatal("ListByJobID should not be called")
	}
}

func TestService_ListByJobID_RepositoryError(t *testing.T) {
	expectedError := errors.New("failed to list applications")

	repo := &MockRepository{
		GetJobClientIDFunc: func(ctx context.Context, jobID uint) (uint, error) {
			return 50, nil
		},
		ListByJobIDFunc: func(ctx context.Context, jobID uint) ([]Application, error) {
			return nil, expectedError
		},
	}

	service := NewService(repo, nil, nil, nil, nil, nil)

	result, err := service.ListByJobID(
		context.Background(),
		10,
		50,
		"client",
	)

	if !errors.Is(err, expectedError) {
		t.Fatalf("expected %v, got %v", expectedError, err)
	}

	if result != nil {
		t.Fatalf("expected nil result, got %#v", result)
	}
}

func TestService_ListMine_Success(t *testing.T) {
	expected := []Application{
		{
			ID:        1,
			JobID:     10,
			CleanerID: 20,
		},
	}

	repo := &MockRepository{
		ListByCleanerIDFunc: func(ctx context.Context, cleanerID uint) ([]Application, error) {
			if cleanerID != 20 {
				t.Fatalf("expected cleaner ID 20, got %d", cleanerID)
			}

			return expected, nil
		},
	}

	service := NewService(repo, nil, nil, nil, nil, nil)

	result, err := service.ListMine(
		context.Background(),
		20,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("expected 1 application, got %d", len(result))
	}
}

func TestService_ListMine_InvalidInput(t *testing.T) {
	service := NewService(
		&MockRepository{},
		nil,
		nil,
		nil,
		nil,
		nil,
	)

	result, err := service.ListMine(
		context.Background(),
		0,
	)

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}

	if result != nil {
		t.Fatalf("expected nil result, got %#v", result)
	}
}

func TestService_ListMine_RepositoryError(t *testing.T) {
	expectedError := errors.New("failed to list cleaner applications")

	repo := &MockRepository{
		ListByCleanerIDFunc: func(ctx context.Context, cleanerID uint) ([]Application, error) {
			return nil, expectedError
		},
	}

	service := NewService(repo, nil, nil, nil, nil, nil)

	result, err := service.ListMine(
		context.Background(),
		20,
	)

	if !errors.Is(err, expectedError) {
		t.Fatalf("expected %v, got %v", expectedError, err)
	}

	if result != nil {
		t.Fatalf("expected nil result, got %#v", result)
	}
}

func TestService_UpdateStatus_Success(t *testing.T) {
	updateCalls := 0
	getCalls := 0

	repo := &MockRepository{
		GetByIDFunc: func(ctx context.Context, id uint) (*Application, error) {
			getCalls++

			if getCalls == 1 {
				return &Application{
					ID:        id,
					JobID:     10,
					CleanerID: 20,
					Status:    "pending",
				}, nil
			}

			return &Application{
				ID:        id,
				JobID:     10,
				CleanerID: 20,
				Status:    "accepted",
			}, nil
		},
		GetJobClientIDFunc: func(ctx context.Context, jobID uint) (uint, error) {
			return 50, nil
		},
		UpdateStatusFunc: func(ctx context.Context, id uint, status string) error {
			updateCalls++

			if id != 1 {
				t.Fatalf("expected application ID 1, got %d", id)
			}

			if status != "accepted" {
				t.Fatalf("expected accepted status, got %q", status)
			}

			return nil
		},
	}

	service := NewService(repo, nil, nil, nil, nil, nil)

	result, err := service.UpdateStatus(
		context.Background(),
		1,
		50,
		"client",
		UpdateApplicationStatusRequest{
			Status: "  accepted  ",
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected updated application")
	}

	if result.Status != "accepted" {
		t.Fatalf("expected accepted status, got %q", result.Status)
	}

	if updateCalls != 1 {
		t.Fatalf("expected UpdateStatus to be called once, got %d", updateCalls)
	}

	if getCalls != 2 {
		t.Fatalf("expected GetByID to be called twice, got %d", getCalls)
	}
}

func TestService_UpdateStatus_AdminSuccess(t *testing.T) {
	getCalls := 0

	repo := &MockRepository{
		GetByIDFunc: func(ctx context.Context, id uint) (*Application, error) {
			getCalls++

			return &Application{
				ID:        id,
				JobID:     10,
				CleanerID: 20,
				Status:    "shortlisted",
			}, nil
		},
		GetJobClientIDFunc: func(ctx context.Context, jobID uint) (uint, error) {
			return 50, nil
		},
		UpdateStatusFunc: func(ctx context.Context, id uint, status string) error {
			return nil
		},
	}

	service := NewService(repo, nil, nil, nil, nil, nil)

	result, err := service.UpdateStatus(
		context.Background(),
		1,
		999,
		"admin",
		UpdateApplicationStatusRequest{
			Status: "rejected",
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected updated application")
	}

	if getCalls != 2 {
		t.Fatalf("expected GetByID twice, got %d", getCalls)
	}
}

func TestService_UpdateStatus_InvalidInput(t *testing.T) {
	tests := []struct {
		name          string
		applicationID uint
		userID        uint
	}{
		{
			name:          "ZeroApplicationID",
			applicationID: 0,
			userID:        50,
		},
		{
			name:          "ZeroUserID",
			applicationID: 1,
			userID:        0,
		},
	}

	service := NewService(
		&MockRepository{},
		nil,
		nil,
		nil,
		nil,
		nil,
	)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := service.UpdateStatus(
				context.Background(),
				tc.applicationID,
				tc.userID,
				"client",
				UpdateApplicationStatusRequest{
					Status: "accepted",
				},
			)

			if !errors.Is(err, ErrInvalidInput) {
				t.Fatalf("expected ErrInvalidInput, got %v", err)
			}

			if result != nil {
				t.Fatalf("expected nil result, got %#v", result)
			}
		})
	}
}

func TestService_UpdateStatus_InvalidStatus(t *testing.T) {
	service := NewService(
		&MockRepository{},
		nil,
		nil,
		nil,
		nil,
		nil,
	)

	result, err := service.UpdateStatus(
		context.Background(),
		1,
		50,
		"client",
		UpdateApplicationStatusRequest{
			Status: "  unknown  ",
		},
	)

	if !errors.Is(err, ErrInvalidApplicationStatus) {
		t.Fatalf(
			"expected ErrInvalidApplicationStatus, got %v",
			err,
		)
	}

	if result != nil {
		t.Fatalf("expected nil result, got %#v", result)
	}
}

func TestService_UpdateStatus_GetByIDError(t *testing.T) {
	expectedError := errors.New("failed to load application")

	repo := &MockRepository{
		GetByIDFunc: func(ctx context.Context, id uint) (*Application, error) {
			return nil, expectedError
		},
	}

	service := NewService(repo, nil, nil, nil, nil, nil)

	result, err := service.UpdateStatus(
		context.Background(),
		1,
		50,
		"client",
		UpdateApplicationStatusRequest{
			Status: "accepted",
		},
	)

	if !errors.Is(err, expectedError) {
		t.Fatalf("expected %v, got %v", expectedError, err)
	}

	if result != nil {
		t.Fatalf("expected nil result, got %#v", result)
	}
}

func TestService_UpdateStatus_GetJobClientIDError(t *testing.T) {
	expectedError := errors.New("failed to load job owner")

	updateCalled := false

	repo := &MockRepository{
		GetByIDFunc: func(ctx context.Context, id uint) (*Application, error) {
			return &Application{
				ID:        id,
				JobID:     10,
				CleanerID: 20,
				Status:    "pending",
			}, nil
		},
		GetJobClientIDFunc: func(ctx context.Context, jobID uint) (uint, error) {
			return 0, expectedError
		},
		UpdateStatusFunc: func(ctx context.Context, id uint, status string) error {
			updateCalled = true
			return nil
		},
	}

	service := NewService(repo, nil, nil, nil, nil, nil)

	result, err := service.UpdateStatus(
		context.Background(),
		1,
		50,
		"client",
		UpdateApplicationStatusRequest{
			Status: "accepted",
		},
	)

	if !errors.Is(err, expectedError) {
		t.Fatalf("expected %v, got %v", expectedError, err)
	}

	if result != nil {
		t.Fatalf("expected nil result, got %#v", result)
	}

	if updateCalled {
		t.Fatal("UpdateStatus should not run before ownership is verified")
	}
}

func TestService_UpdateStatus_Forbidden(t *testing.T) {
	updateCalled := false

	repo := &MockRepository{
		GetByIDFunc: func(ctx context.Context, id uint) (*Application, error) {
			return &Application{
				ID:        id,
				JobID:     10,
				CleanerID: 20,
				Status:    "pending",
			}, nil
		},
		GetJobClientIDFunc: func(ctx context.Context, jobID uint) (uint, error) {
			return 50, nil
		},
		UpdateStatusFunc: func(ctx context.Context, id uint, status string) error {
			updateCalled = true
			return nil
		},
	}

	service := NewService(repo, nil, nil, nil, nil, nil)

	result, err := service.UpdateStatus(
		context.Background(),
		1,
		999,
		"client",
		UpdateApplicationStatusRequest{
			Status: "accepted",
		},
	)

	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}

	if result != nil {
		t.Fatalf("expected nil result, got %#v", result)
	}

	if updateCalled {
		t.Fatal("unauthorized user must not update application status")
	}
}

func TestService_UpdateStatus_RepositoryError(t *testing.T) {
	expectedError := errors.New("failed to update status")

	repo := &MockRepository{
		GetByIDFunc: func(ctx context.Context, id uint) (*Application, error) {
			return &Application{
				ID:        id,
				JobID:     10,
				CleanerID: 20,
				Status:    "pending",
			}, nil
		},
		GetJobClientIDFunc: func(ctx context.Context, jobID uint) (uint, error) {
			return 50, nil
		},
		UpdateStatusFunc: func(ctx context.Context, id uint, status string) error {
			return expectedError
		},
	}

	service := NewService(repo, nil, nil, nil, nil, nil)

	result, err := service.UpdateStatus(
		context.Background(),
		1,
		50,
		"client",
		UpdateApplicationStatusRequest{
			Status: "accepted",
		},
	)

	if !errors.Is(err, expectedError) {
		t.Fatalf("expected %v, got %v", expectedError, err)
	}

	if result != nil {
		t.Fatalf("expected nil result, got %#v", result)
	}
}

func TestService_UpdateStatus_FinalGetByIDError(t *testing.T) {
	expectedError := errors.New("failed to reload application")
	getCalls := 0

	repo := &MockRepository{
		GetByIDFunc: func(ctx context.Context, id uint) (*Application, error) {
			getCalls++

			if getCalls == 1 {
				return &Application{
					ID:        id,
					JobID:     10,
					CleanerID: 20,
					Status:    "pending",
				}, nil
			}

			return nil, expectedError
		},
		GetJobClientIDFunc: func(ctx context.Context, jobID uint) (uint, error) {
			return 50, nil
		},
		UpdateStatusFunc: func(ctx context.Context, id uint, status string) error {
			return nil
		},
	}

	service := NewService(repo, nil, nil, nil, nil, nil)

	result, err := service.UpdateStatus(
		context.Background(),
		1,
		50,
		"client",
		UpdateApplicationStatusRequest{
			Status: "accepted",
		},
	)

	if !errors.Is(err, expectedError) {
		t.Fatalf("expected %v, got %v", expectedError, err)
	}

	if result != nil {
		t.Fatalf("expected nil result, got %#v", result)
	}
}

func TestIsAllowedStatus(t *testing.T) {
	tests := []struct {
		status  string
		allowed bool
	}{
		{"pending", true},
		{"shortlisted", true},
		{"invited", true},
		{"accepted", true},
		{"rejected", true},
		{"completed", true},
		{"cancelled", true},
		{"unknown", false},
		{"", false},
		{"ACCEPTED", false},
	}

	for _, tc := range tests {
		t.Run(tc.status, func(t *testing.T) {
			result := isAllowedStatus(tc.status)

			if result != tc.allowed {
				t.Fatalf(
					"expected %q allowed=%v, got %v",
					tc.status,
					tc.allowed,
					result,
				)
			}
		})
	}
}


