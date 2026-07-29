package jobs

import (
	"context"
	"errors"
	"testing"
)

type MockRepository struct {
	CreateFunc         func(ctx context.Context, job *Job) error
	GetByIDFunc        func(ctx context.Context, id uint) (*Job, error)
	ListFunc           func(ctx context.Context) ([]Job, error)
	ListByClientIDFunc func(ctx context.Context, clientID uint) ([]Job, error)
	UpdateFunc         func(ctx context.Context, job *Job) error
	DeleteFunc         func(ctx context.Context, id uint) error
	SearchFunc         func(ctx context.Context, req SearchJobRequest) ([]Job, error)
}

func (m *MockRepository) Create(ctx context.Context, job *Job) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, job)
	}
	return nil
}

func (m *MockRepository) GetByID(ctx context.Context, id uint) (*Job, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockRepository) List(ctx context.Context) ([]Job, error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx)
	}
	return nil, nil
}

func (m *MockRepository) ListByClientID(ctx context.Context, clientID uint) ([]Job, error) {
	if m.ListByClientIDFunc != nil {
		return m.ListByClientIDFunc(ctx, clientID)
	}
	return nil, nil
}

func (m *MockRepository) Update(ctx context.Context, job *Job) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, job)
	}
	return nil
}

func (m *MockRepository) Delete(ctx context.Context, id uint) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

func (m *MockRepository) Search(ctx context.Context, req SearchJobRequest) ([]Job, error) {
	if m.SearchFunc != nil {
		return m.SearchFunc(ctx, req)
	}
	return nil, nil
}

func TestNewService(t *testing.T) {
	repo := &MockRepository{}

	service := NewService(repo, nil, nil)

	if service == nil {
		t.Fatal("expected service")
	}

	if service.repo != repo {
		t.Fatal("expected repository to be assigned")
	}

	if service.usageService != nil {
		t.Fatal("expected nil usage service")
	}

	if service.subscriptionAccess != nil {
		t.Fatal("expected nil subscription access")
	}
}

func TestService_Create_Success(t *testing.T) {
	repo := &MockRepository{
		CreateFunc: func(ctx context.Context, job *Job) error {
			job.ID = 10
			return nil
		},
	}

	service := NewService(repo, nil, nil)

	job, err := service.Create(context.Background(), 5, CreateJobRequest{
		Title:       "Domestic cleaner",
		Description: "Clean a two-bedroom flat",
		Location:    "East London",
		JobType:     "domestic",
		ListingType: "job",
		Budget:      80,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if job == nil {
		t.Fatal("expected job")
	}

	if job.ID != 10 {
		t.Fatalf("expected job ID 10, got %d", job.ID)
	}

	if job.ClientID != 5 {
		t.Fatalf("expected client ID 5, got %d", job.ClientID)
	}

	if job.Status != "open" {
		t.Fatalf("expected open status, got %q", job.Status)
	}

	if job.ListingType != "job" {
		t.Fatalf("expected listing type job, got %q", job.ListingType)
	}
}

func TestService_Create_TrimsInput(t *testing.T) {
	repo := &MockRepository{
		CreateFunc: func(ctx context.Context, job *Job) error {
			if job.Title != "Domestic cleaner" {
				t.Fatalf("unexpected title %q", job.Title)
			}

			if job.Description != "Clean a flat" {
				t.Fatalf("unexpected description %q", job.Description)
			}

			if job.Location != "London" {
				t.Fatalf("unexpected location %q", job.Location)
			}

			if job.JobType != "domestic" {
				t.Fatalf("unexpected job type %q", job.JobType)
			}

			if job.ListingType != "shift" {
				t.Fatalf("unexpected listing type %q", job.ListingType)
			}

			return nil
		},
	}

	service := NewService(repo, nil, nil)

	_, err := service.Create(context.Background(), 5, CreateJobRequest{
		Title:       "  Domestic cleaner  ",
		Description: "  Clean a flat  ",
		Location:    "  London  ",
		JobType:     "  domestic  ",
		ListingType: "  shift  ",
		Budget:      50,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestService_Create_DefaultListingType(t *testing.T) {
	repo := &MockRepository{
		CreateFunc: func(ctx context.Context, job *Job) error {
			if job.ListingType != "shift" {
				t.Fatalf("expected default listing type shift, got %q", job.ListingType)
			}
			return nil
		},
	}

	service := NewService(repo, nil, nil)

	job, err := service.Create(context.Background(), 5, CreateJobRequest{
		Title:       "Cleaner needed",
		Description: "One-day cleaning shift",
		Location:    "London",
		JobType:     "domestic",
		Budget:      60,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if job.ListingType != "shift" {
		t.Fatalf("expected shift, got %q", job.ListingType)
	}
}

func TestService_Create_InvalidInput(t *testing.T) {
	service := NewService(&MockRepository{}, nil, nil)

	tests := []struct {
		name     string
		clientID uint
		request  CreateJobRequest
	}{
		{
			name:     "zero client ID",
			clientID: 0,
			request: CreateJobRequest{
				Title:       "Cleaner",
				Description: "Clean flat",
				Location:    "London",
				JobType:     "domestic",
			},
		},
		{
			name:     "missing title",
			clientID: 5,
			request: CreateJobRequest{
				Description: "Clean flat",
				Location:    "London",
				JobType:     "domestic",
			},
		},
		{
			name:     "missing description",
			clientID: 5,
			request: CreateJobRequest{
				Title:    "Cleaner",
				Location: "London",
				JobType:  "domestic",
			},
		},
		{
			name:     "missing location",
			clientID: 5,
			request: CreateJobRequest{
				Title:       "Cleaner",
				Description: "Clean flat",
				JobType:     "domestic",
			},
		},
		{
			name:     "missing job type",
			clientID: 5,
			request: CreateJobRequest{
				Title:       "Cleaner",
				Description: "Clean flat",
				Location:    "London",
			},
		},
		{
			name:     "negative budget",
			clientID: 5,
			request: CreateJobRequest{
				Title:       "Cleaner",
				Description: "Clean flat",
				Location:    "London",
				JobType:     "domestic",
				Budget:      -1,
			},
		},
		{
			name:     "invalid listing type",
			clientID: 5,
			request: CreateJobRequest{
				Title:       "Cleaner",
				Description: "Clean flat",
				Location:    "London",
				JobType:     "domestic",
				ListingType: "contract",
				Budget:      50,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			job, err := service.Create(
				context.Background(),
				tc.clientID,
				tc.request,
			)

			if job != nil {
				t.Fatalf("expected nil job, got %+v", job)
			}

			if !errors.Is(err, ErrInvalidInput) {
				t.Fatalf("expected ErrInvalidInput, got %v", err)
			}
		})
	}
}

func TestService_Create_RepositoryError(t *testing.T) {
	expected := errors.New("database unavailable")

	repo := &MockRepository{
		CreateFunc: func(ctx context.Context, job *Job) error {
			return expected
		},
	}

	service := NewService(repo, nil, nil)

	job, err := service.Create(context.Background(), 5, CreateJobRequest{
		Title:       "Cleaner",
		Description: "Clean flat",
		Location:    "London",
		JobType:     "domestic",
		ListingType: "shift",
		Budget:      50,
	})

	if job != nil {
		t.Fatalf("expected nil job, got %+v", job)
	}

	if !errors.Is(err, expected) {
		t.Fatalf("expected %v, got %v", expected, err)
	}
}

func TestIsAllowedListingType(t *testing.T) {
	tests := []struct {
		value   string
		allowed bool
	}{
		{value: "job", allowed: true},
		{value: "shift", allowed: true},
		{value: "contract", allowed: false},
		{value: "", allowed: false},
	}

	for _, tc := range tests {
		result := isAllowedListingType(tc.value)

		if result != tc.allowed {
			t.Fatalf(
				"expected %q allowed=%v, got %v",
				tc.value,
				tc.allowed,
				result,
			)
		}
	}
}

func TestService_GetByID_Success(t *testing.T) {
	expected := &Job{
		ID:       12,
		ClientID: 5,
		Title:    "Cleaner needed",
	}

	repo := &MockRepository{
		GetByIDFunc: func(ctx context.Context, id uint) (*Job, error) {
			if id != 12 {
				t.Fatalf("expected ID 12, got %d", id)
			}
			return expected, nil
		},
	}

	service := NewService(repo, nil, nil)

	job, err := service.GetByID(context.Background(), 12)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if job != expected {
		t.Fatal("expected repository job")
	}
}

func TestService_GetByID_InvalidID(t *testing.T) {
	service := NewService(&MockRepository{}, nil, nil)

	job, err := service.GetByID(context.Background(), 0)

	if job != nil {
		t.Fatalf("expected nil job, got %+v", job)
	}

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestService_GetByID_RepositoryError(t *testing.T) {
	expected := errors.New("query failed")

	repo := &MockRepository{
		GetByIDFunc: func(ctx context.Context, id uint) (*Job, error) {
			return nil, expected
		},
	}

	service := NewService(repo, nil, nil)

	job, err := service.GetByID(context.Background(), 12)

	if job != nil {
		t.Fatalf("expected nil job, got %+v", job)
	}

	if !errors.Is(err, expected) {
		t.Fatalf("expected %v, got %v", expected, err)
	}
}

func TestService_List_Success(t *testing.T) {
	expected := []Job{
		{ID: 1, Title: "First job"},
		{ID: 2, Title: "Second job"},
	}

	repo := &MockRepository{
		ListFunc: func(ctx context.Context) ([]Job, error) {
			return expected, nil
		},
	}

	service := NewService(repo, nil, nil)

	jobs, err := service.List(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(jobs) != 2 {
		t.Fatalf("expected 2 jobs, got %d", len(jobs))
	}
}

func TestService_List_RepositoryError(t *testing.T) {
	expected := errors.New("database unavailable")

	repo := &MockRepository{
		ListFunc: func(ctx context.Context) ([]Job, error) {
			return nil, expected
		},
	}

	service := NewService(repo, nil, nil)

	jobs, err := service.List(context.Background())

	if jobs != nil {
		t.Fatalf("expected nil jobs, got %+v", jobs)
	}

	if !errors.Is(err, expected) {
		t.Fatalf("expected %v, got %v", expected, err)
	}
}

func TestService_ListByClientID_Success(t *testing.T) {
	expected := []Job{
		{ID: 1, ClientID: 5},
		{ID: 2, ClientID: 5},
	}

	repo := &MockRepository{
		ListByClientIDFunc: func(ctx context.Context, clientID uint) ([]Job, error) {
			if clientID != 5 {
				t.Fatalf("expected client ID 5, got %d", clientID)
			}
			return expected, nil
		},
	}

	service := NewService(repo, nil, nil)

	jobs, err := service.ListByClientID(context.Background(), 5)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(jobs) != 2 {
		t.Fatalf("expected 2 jobs, got %d", len(jobs))
	}
}

func TestService_ListByClientID_InvalidID(t *testing.T) {
	service := NewService(&MockRepository{}, nil, nil)

	jobs, err := service.ListByClientID(context.Background(), 0)

	if jobs != nil {
		t.Fatalf("expected nil jobs, got %+v", jobs)
	}

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestService_ListByClientID_RepositoryError(t *testing.T) {
	expected := errors.New("database unavailable")

	repo := &MockRepository{
		ListByClientIDFunc: func(ctx context.Context, clientID uint) ([]Job, error) {
			return nil, expected
		},
	}

	service := NewService(repo, nil, nil)

	jobs, err := service.ListByClientID(context.Background(), 5)

	if jobs != nil {
		t.Fatalf("expected nil jobs, got %+v", jobs)
	}

	if !errors.Is(err, expected) {
		t.Fatalf("expected %v, got %v", expected, err)
	}
}

func TestService_Update_OwnerSuccess(t *testing.T) {
	existing := &Job{
		ID:          12,
		ClientID:    5,
		Title:       "Old title",
		Description: "Old description",
		Location:    "London",
		JobType:     "domestic",
		ListingType: "shift",
		Budget:      50,
		Status:      "open",
	}

	repo := &MockRepository{
		GetByIDFunc: func(ctx context.Context, id uint) (*Job, error) {
			return existing, nil
		},
		UpdateFunc: func(ctx context.Context, job *Job) error {
			if job.Title != "Updated title" {
				t.Fatalf("unexpected title %q", job.Title)
			}

			if job.Description != "Updated description" {
				t.Fatalf("unexpected description %q", job.Description)
			}

			if job.Location != "East London" {
				t.Fatalf("unexpected location %q", job.Location)
			}

			if job.Status != "closed" {
				t.Fatalf("unexpected status %q", job.Status)
			}

			return nil
		},
	}

	service := NewService(repo, nil, nil)

	job, err := service.Update(
		context.Background(),
		12,
		5,
		"user",
		UpdateJobRequest{
			Title:       "  Updated title  ",
			Description: "  Updated description  ",
			Location:    "  East London  ",
			JobType:     "  end_of_tenancy  ",
			ListingType: "  job  ",
			Budget:      120,
			Status:      "  closed  ",
		},
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if job != existing {
		t.Fatal("expected updated existing job")
	}

	if job.Budget != 120 {
		t.Fatalf("expected budget 120, got %d", job.Budget)
	}
}

func TestService_Update_AdminSuccess(t *testing.T) {
	existing := &Job{
		ID:       12,
		ClientID: 99,
	}

	repo := &MockRepository{
		GetByIDFunc: func(ctx context.Context, id uint) (*Job, error) {
			return existing, nil
		},
		UpdateFunc: func(ctx context.Context, job *Job) error {
			return nil
		},
	}

	service := NewService(repo, nil, nil)

	job, err := service.Update(
		context.Background(),
		12,
		1,
		"admin",
		UpdateJobRequest{
			Title:       "Updated",
			Description: "Updated description",
			Location:    "London",
			JobType:     "domestic",
			ListingType: "shift",
			Budget:      80,
			Status:      "open",
		},
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if job == nil {
		t.Fatal("expected updated job")
	}
}

func TestService_Update_InvalidIDs(t *testing.T) {
	service := NewService(&MockRepository{}, nil, nil)

	tests := []struct {
		id       uint
		clientID uint
	}{
		{id: 0, clientID: 5},
		{id: 12, clientID: 0},
	}

	for _, tc := range tests {
		job, err := service.Update(
			context.Background(),
			tc.id,
			tc.clientID,
			"user",
			UpdateJobRequest{},
		)

		if job != nil {
			t.Fatalf("expected nil job, got %+v", job)
		}

		if !errors.Is(err, ErrInvalidInput) {
			t.Fatalf("expected ErrInvalidInput, got %v", err)
		}
	}
}

func TestService_Update_GetByIDError(t *testing.T) {
	expected := errors.New("query failed")

	repo := &MockRepository{
		GetByIDFunc: func(ctx context.Context, id uint) (*Job, error) {
			return nil, expected
		},
	}

	service := NewService(repo, nil, nil)

	job, err := service.Update(
		context.Background(),
		12,
		5,
		"user",
		UpdateJobRequest{},
	)

	if job != nil {
		t.Fatalf("expected nil job, got %+v", job)
	}

	if !errors.Is(err, expected) {
		t.Fatalf("expected %v, got %v", expected, err)
	}
}

func TestService_Update_Forbidden(t *testing.T) {
	repo := &MockRepository{
		GetByIDFunc: func(ctx context.Context, id uint) (*Job, error) {
			return &Job{
				ID:       id,
				ClientID: 99,
			}, nil
		},
	}

	service := NewService(repo, nil, nil)

	job, err := service.Update(
		context.Background(),
		12,
		5,
		"user",
		UpdateJobRequest{},
	)

	if job != nil {
		t.Fatalf("expected nil job, got %+v", job)
	}

	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestService_Update_InvalidRequest(t *testing.T) {
	validJob := func() *Job {
		return &Job{
			ID:       12,
			ClientID: 5,
		}
	}

	tests := []struct {
		name    string
		request UpdateJobRequest
	}{
		{
			name: "missing title",
			request: UpdateJobRequest{
				Description: "Description",
				Location:    "London",
				JobType:     "domestic",
				Status:      "open",
			},
		},
		{
			name: "missing description",
			request: UpdateJobRequest{
				Title:    "Title",
				Location: "London",
				JobType:  "domestic",
				Status:   "open",
			},
		},
		{
			name: "missing location",
			request: UpdateJobRequest{
				Title:       "Title",
				Description: "Description",
				JobType:     "domestic",
				Status:      "open",
			},
		},
		{
			name: "missing job type",
			request: UpdateJobRequest{
				Title:       "Title",
				Description: "Description",
				Location:    "London",
				Status:      "open",
			},
		},
		{
			name: "missing status",
			request: UpdateJobRequest{
				Title:       "Title",
				Description: "Description",
				Location:    "London",
				JobType:     "domestic",
			},
		},
		{
			name: "negative budget",
			request: UpdateJobRequest{
				Title:       "Title",
				Description: "Description",
				Location:    "London",
				JobType:     "domestic",
				Status:      "open",
				Budget:      -1,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := &MockRepository{
				GetByIDFunc: func(ctx context.Context, id uint) (*Job, error) {
					return validJob(), nil
				},
			}

			service := NewService(repo, nil, nil)

			job, err := service.Update(
				context.Background(),
				12,
				5,
				"user",
				tc.request,
			)

			if job != nil {
				t.Fatalf("expected nil job, got %+v", job)
			}

			if !errors.Is(err, ErrInvalidInput) {
				t.Fatalf("expected ErrInvalidInput, got %v", err)
			}
		})
	}
}

func TestService_Update_RepositoryError(t *testing.T) {
	expected := errors.New("update failed")

	repo := &MockRepository{
		GetByIDFunc: func(ctx context.Context, id uint) (*Job, error) {
			return &Job{
				ID:       id,
				ClientID: 5,
			}, nil
		},
		UpdateFunc: func(ctx context.Context, job *Job) error {
			return expected
		},
	}

	service := NewService(repo, nil, nil)

	job, err := service.Update(
		context.Background(),
		12,
		5,
		"user",
		UpdateJobRequest{
			Title:       "Updated",
			Description: "Updated description",
			Location:    "London",
			JobType:     "domestic",
			ListingType: "shift",
			Budget:      50,
			Status:      "open",
		},
	)

	if job != nil {
		t.Fatalf("expected nil job, got %+v", job)
	}

	if !errors.Is(err, expected) {
		t.Fatalf("expected %v, got %v", expected, err)
	}
}

func TestService_Delete_OwnerSuccess(t *testing.T) {
	deleted := false

	repo := &MockRepository{
		GetByIDFunc: func(ctx context.Context, id uint) (*Job, error) {
			return &Job{
				ID:       id,
				ClientID: 5,
			}, nil
		},
		DeleteFunc: func(ctx context.Context, id uint) error {
			deleted = true
			return nil
		},
	}

	service := NewService(repo, nil, nil)

	err := service.Delete(context.Background(), 12, 5, "user")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !deleted {
		t.Fatal("expected repository Delete to be called")
	}
}

func TestService_Delete_AdminSuccess(t *testing.T) {
	repo := &MockRepository{
		GetByIDFunc: func(ctx context.Context, id uint) (*Job, error) {
			return &Job{
				ID:       id,
				ClientID: 99,
			}, nil
		},
		DeleteFunc: func(ctx context.Context, id uint) error {
			return nil
		},
	}

	service := NewService(repo, nil, nil)

	err := service.Delete(context.Background(), 12, 1, "admin")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestService_Delete_InvalidIDs(t *testing.T) {
	service := NewService(&MockRepository{}, nil, nil)

	tests := []struct {
		id       uint
		clientID uint
	}{
		{id: 0, clientID: 5},
		{id: 12, clientID: 0},
	}

	for _, tc := range tests {
		err := service.Delete(
			context.Background(),
			tc.id,
			tc.clientID,
			"user",
		)

		if !errors.Is(err, ErrInvalidInput) {
			t.Fatalf("expected ErrInvalidInput, got %v", err)
		}
	}
}

func TestService_Delete_GetByIDError(t *testing.T) {
	expected := errors.New("query failed")

	repo := &MockRepository{
		GetByIDFunc: func(ctx context.Context, id uint) (*Job, error) {
			return nil, expected
		},
	}

	service := NewService(repo, nil, nil)

	err := service.Delete(context.Background(), 12, 5, "user")

	if !errors.Is(err, expected) {
		t.Fatalf("expected %v, got %v", expected, err)
	}
}

func TestService_Delete_Forbidden(t *testing.T) {
	repo := &MockRepository{
		GetByIDFunc: func(ctx context.Context, id uint) (*Job, error) {
			return &Job{
				ID:       id,
				ClientID: 99,
			}, nil
		},
	}

	service := NewService(repo, nil, nil)

	err := service.Delete(context.Background(), 12, 5, "user")

	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestService_Delete_RepositoryError(t *testing.T) {
	expected := errors.New("delete failed")

	repo := &MockRepository{
		GetByIDFunc: func(ctx context.Context, id uint) (*Job, error) {
			return &Job{
				ID:       id,
				ClientID: 5,
			}, nil
		},
		DeleteFunc: func(ctx context.Context, id uint) error {
			return expected
		},
	}

	service := NewService(repo, nil, nil)

	err := service.Delete(context.Background(), 12, 5, "user")

	if !errors.Is(err, expected) {
		t.Fatalf("expected %v, got %v", expected, err)
	}
}

func TestService_Search_Success(t *testing.T) {
	repo := &MockRepository{
		SearchFunc: func(ctx context.Context, req SearchJobRequest) ([]Job, error) {
			if req.Location != "London" {
				t.Fatalf("expected trimmed location, got %q", req.Location)
			}

			if req.JobType != "domestic" {
				t.Fatalf("expected trimmed job type, got %q", req.JobType)
			}

			if req.ListingType != "shift" {
				t.Fatalf("expected trimmed listing type, got %q", req.ListingType)
			}

			return []Job{
				{ID: 1, Title: "Cleaner needed"},
			}, nil
		},
	}

	service := NewService(repo, nil, nil)

	jobs, err := service.Search(context.Background(), SearchJobRequest{
		Location:    "  London  ",
		JobType:     "  domestic  ",
		ListingType: "  shift  ",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(jobs) != 1 {
		t.Fatalf("expected 1 job, got %d", len(jobs))
	}
}

func TestService_Search_RepositoryError(t *testing.T) {
	expected := errors.New("search failed")

	repo := &MockRepository{
		SearchFunc: func(ctx context.Context, req SearchJobRequest) ([]Job, error) {
			return nil, expected
		},
	}

	service := NewService(repo, nil, nil)

	jobs, err := service.Search(context.Background(), SearchJobRequest{
		Location: "London",
	})

	if jobs != nil {
		t.Fatalf("expected nil jobs, got %+v", jobs)
	}

	if !errors.Is(err, expected) {
		t.Fatalf("expected %v, got %v", expected, err)
	}
}
