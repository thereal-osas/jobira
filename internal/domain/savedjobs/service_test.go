package savedjobs

import (
	"context"
	"errors"
	"testing"
	"time"
)

type mockRepository struct {
	saveFn         func(context.Context, *SavedJob) error
	listByUserIDFn func(context.Context, uint) ([]SavedJob, error)
	deleteFn       func(context.Context, uint, uint) error
}

func (m *mockRepository) Save(
	ctx context.Context,
	savedJob *SavedJob,
) error {
	if m.saveFn != nil {
		return m.saveFn(ctx, savedJob)
	}

	return nil
}

func (m *mockRepository) ListByUserID(
	ctx context.Context,
	userID uint,
) ([]SavedJob, error) {
	if m.listByUserIDFn != nil {
		return m.listByUserIDFn(ctx, userID)
	}

	return nil, nil
}

func (m *mockRepository) Delete(
	ctx context.Context,
	userID uint,
	jobID uint,
) error {
	if m.deleteFn != nil {
		return m.deleteFn(
			ctx,
			userID,
			jobID,
		)
	}

	return nil
}

func TestNewService(t *testing.T) {
	repo := &mockRepository{}

	service := NewService(repo)

	if service == nil {
		t.Fatal("expected service")
	}

	if service.repo != repo {
		t.Fatal("expected repository assigned")
	}
}

func TestService_Save_Success(t *testing.T) {
	saveCalled := false

	repo := &mockRepository{
		saveFn: func(
			ctx context.Context,
			savedJob *SavedJob,
		) error {
			saveCalled = true

			if savedJob.UserID != 5 {
				t.Fatalf(
					"expected user ID 5, got %d",
					savedJob.UserID,
				)
			}

			if savedJob.JobID != 12 {
				t.Fatalf(
					"expected job ID 12, got %d",
					savedJob.JobID,
				)
			}

			savedJob.ID = 20
			savedJob.CreatedAt = time.Now()

			return nil
		},
	}

	service := NewService(repo)

	savedJob, err := service.Save(
		context.Background(),
		5,
		SavedJbRequest{
			JobID: 12,
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if savedJob == nil {
		t.Fatal("expected saved job")
	}

	if savedJob.ID != 20 {
		t.Fatalf(
			"expected ID 20, got %d",
			savedJob.ID,
		)
	}

	if savedJob.UserID != 5 {
		t.Fatalf(
			"expected user ID 5, got %d",
			savedJob.UserID,
		)
	}

	if savedJob.JobID != 12 {
		t.Fatalf(
			"expected job ID 12, got %d",
			savedJob.JobID,
		)
	}

	if savedJob.CreatedAt.IsZero() {
		t.Fatal("expected created_at")
	}

	if !saveCalled {
		t.Fatal("expected Save repository method to be called")
	}
}

func TestService_Save_InvalidInput(t *testing.T) {
	service := NewService(&mockRepository{})

	tests := []struct {
		name   string
		userID uint
		jobID  uint
	}{
		{
			name:   "zero user ID",
			userID: 0,
			jobID:  12,
		},
		{
			name:   "zero job ID",
			userID: 5,
			jobID:  0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			savedJob, err := service.Save(
				context.Background(),
				test.userID,
				SavedJbRequest{
					JobID: test.jobID,
				},
			)

			if savedJob != nil {
				t.Fatalf(
					"expected nil saved job, got %+v",
					savedJob,
				)
			}

			if !errors.Is(err, ErrInvalidInput) {
				t.Fatalf(
					"expected ErrInvalidInput, got %v",
					err,
				)
			}
		})
	}
}

func TestService_Save_AlreadySaved(t *testing.T) {
	repo := &mockRepository{
		saveFn: func(
			context.Context,
			*SavedJob,
		) error {
			return ErrAlreadySved
		},
	}

	service := NewService(repo)

	savedJob, err := service.Save(
		context.Background(),
		5,
		SavedJbRequest{
			JobID: 12,
		},
	)

	if savedJob != nil {
		t.Fatalf(
			"expected nil saved job, got %+v",
			savedJob,
		)
	}

	if !errors.Is(err, ErrAlreadySved) {
		t.Fatalf(
			"expected ErrAlreadySved, got %v",
			err,
		)
	}
}

func TestService_Save_RepositoryError(t *testing.T) {
	expectedErr := errors.New("save failed")

	repo := &mockRepository{
		saveFn: func(
			context.Context,
			*SavedJob,
		) error {
			return expectedErr
		},
	}

	service := NewService(repo)

	savedJob, err := service.Save(
		context.Background(),
		5,
		SavedJbRequest{
			JobID: 12,
		},
	)

	if savedJob != nil {
		t.Fatalf(
			"expected nil saved job, got %+v",
			savedJob,
		)
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected %v, got %v",
			expectedErr,
			err,
		)
	}
}

func TestService_ListMine_Success(t *testing.T) {
	now := time.Now()

	repo := &mockRepository{
		listByUserIDFn: func(
			ctx context.Context,
			userID uint,
		) ([]SavedJob, error) {
			if userID != 5 {
				t.Fatalf(
					"expected user ID 5, got %d",
					userID,
				)
			}

			return []SavedJob{
				{
					ID:        1,
					UserID:    5,
					JobID:     12,
					CreatedAt: now,
				},
				{
					ID:        2,
					UserID:    5,
					JobID:     15,
					CreatedAt: now.Add(-time.Hour),
				},
			}, nil
		},
	}

	service := NewService(repo)

	savedJobs, err := service.ListMine(
		context.Background(),
		5,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(savedJobs) != 2 {
		t.Fatalf(
			"expected 2 saved jobs, got %d",
			len(savedJobs),
		)
	}

	if savedJobs[0].JobID != 12 {
		t.Fatalf(
			"expected first job ID 12, got %d",
			savedJobs[0].JobID,
		)
	}

	if savedJobs[1].JobID != 15 {
		t.Fatalf(
			"expected second job ID 15, got %d",
			savedJobs[1].JobID,
		)
	}
}

func TestService_ListMine_Empty(t *testing.T) {
	repo := &mockRepository{
		listByUserIDFn: func(
			context.Context,
			uint,
		) ([]SavedJob, error) {
			return []SavedJob{}, nil
		},
	}

	service := NewService(repo)

	savedJobs, err := service.ListMine(
		context.Background(),
		5,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(savedJobs) != 0 {
		t.Fatalf(
			"expected no saved jobs, got %d",
			len(savedJobs),
		)
	}
}

func TestService_ListMine_InvalidInput(t *testing.T) {
	service := NewService(&mockRepository{})

	savedJobs, err := service.ListMine(
		context.Background(),
		0,
	)

	if savedJobs != nil {
		t.Fatalf(
			"expected nil saved jobs, got %+v",
			savedJobs,
		)
	}

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf(
			"expected ErrInvalidInput, got %v",
			err,
		)
	}
}

func TestService_ListMine_RepositoryError(t *testing.T) {
	expectedErr := errors.New("list saved jobs failed")

	repo := &mockRepository{
		listByUserIDFn: func(
			context.Context,
			uint,
		) ([]SavedJob, error) {
			return nil, expectedErr
		},
	}

	service := NewService(repo)

	savedJobs, err := service.ListMine(
		context.Background(),
		5,
	)

	if savedJobs != nil {
		t.Fatalf(
			"expected nil saved jobs, got %+v",
			savedJobs,
		)
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected %v, got %v",
			expectedErr,
			err,
		)
	}
}

func TestService_Delete_Success(t *testing.T) {
	deleteCalled := false

	repo := &mockRepository{
		deleteFn: func(
			ctx context.Context,
			userID uint,
			jobID uint,
		) error {
			deleteCalled = true

			if userID != 5 {
				t.Fatalf(
					"expected user ID 5, got %d",
					userID,
				)
			}

			if jobID != 12 {
				t.Fatalf(
					"expected job ID 12, got %d",
					jobID,
				)
			}

			return nil
		},
	}

	service := NewService(repo)

	err := service.Delete(
		context.Background(),
		5,
		12,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !deleteCalled {
		t.Fatal("expected Delete repository method to be called")
	}
}

func TestService_Delete_InvalidInput(t *testing.T) {
	service := NewService(&mockRepository{})

	tests := []struct {
		name   string
		userID uint
		jobID  uint
	}{
		{
			name:   "zero user ID",
			userID: 0,
			jobID:  12,
		},
		{
			name:   "zero job ID",
			userID: 5,
			jobID:  0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := service.Delete(
				context.Background(),
				test.userID,
				test.jobID,
			)

			if !errors.Is(err, ErrInvalidInput) {
				t.Fatalf(
					"expected ErrInvalidInput, got %v",
					err,
				)
			}
		})
	}
}

func TestService_Delete_NotFound(t *testing.T) {
	repo := &mockRepository{
		deleteFn: func(
			context.Context,
			uint,
			uint,
		) error {
			return ErrSavedJobNotFound
		},
	}

	service := NewService(repo)

	err := service.Delete(
		context.Background(),
		5,
		12,
	)

	if !errors.Is(err, ErrSavedJobNotFound) {
		t.Fatalf(
			"expected ErrSavedJobNotFound, got %v",
			err,
		)
	}
}

func TestService_Delete_RepositoryError(t *testing.T) {
	expectedErr := errors.New("delete saved job failed")

	repo := &mockRepository{
		deleteFn: func(
			context.Context,
			uint,
			uint,
		) error {
			return expectedErr
		},
	}

	service := NewService(repo)

	err := service.Delete(
		context.Background(),
		5,
		12,
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected %v, got %v",
			expectedErr,
			err,
		)
	}
}
