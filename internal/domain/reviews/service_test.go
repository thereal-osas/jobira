package reviews

import (
	"context"
	"errors"
	"testing"
	"time"
)

type mockRepository struct {
	createFn               func(context.Context, *Review) error
	listByCleanerIDFn      func(context.Context, uint) ([]Review, error)
	listByClientIDFn       func(context.Context, uint) ([]Review, error)
	getJobClientIDFn       func(context.Context, uint) (uint, error)
	getAcceptedCleanerIDFn func(context.Context, uint) (uint, error)
	getJobStatusFn         func(context.Context, uint) (string, error)
}

func (m *mockRepository) Create(
	ctx context.Context,
	review *Review,
) error {
	if m.createFn != nil {
		return m.createFn(ctx, review)
	}

	return nil
}

func (m *mockRepository) ListByCleanerID(
	ctx context.Context,
	cleanerID uint,
) ([]Review, error) {
	if m.listByCleanerIDFn != nil {
		return m.listByCleanerIDFn(ctx, cleanerID)
	}

	return nil, nil
}

func (m *mockRepository) ListByClientID(
	ctx context.Context,
	clientID uint,
) ([]Review, error) {
	if m.listByClientIDFn != nil {
		return m.listByClientIDFn(ctx, clientID)
	}

	return nil, nil
}

func (m *mockRepository) GetJobClientID(
	ctx context.Context,
	jobID uint,
) (uint, error) {
	if m.getJobClientIDFn != nil {
		return m.getJobClientIDFn(ctx, jobID)
	}

	return 0, nil
}

func (m *mockRepository) GetAcceptedCleanerID(
	ctx context.Context,
	jobID uint,
) (uint, error) {
	if m.getAcceptedCleanerIDFn != nil {
		return m.getAcceptedCleanerIDFn(ctx, jobID)
	}

	return 0, nil
}

func (m *mockRepository) GetJobStatus(
	ctx context.Context,
	jobID uint,
) (string, error) {
	if m.getJobStatusFn != nil {
		return m.getJobStatusFn(ctx, jobID)
	}

	return "", nil
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

func TestService_Create_Success(t *testing.T) {
	createCalled := false

	repo := &mockRepository{
		getJobClientIDFn: func(
			ctx context.Context,
			jobID uint,
		) (uint, error) {
			if jobID != 12 {
				t.Fatalf(
					"expected job ID 12, got %d",
					jobID,
				)
			}

			return 5, nil
		},
		getJobStatusFn: func(
			ctx context.Context,
			jobID uint,
		) (string, error) {
			if jobID != 12 {
				t.Fatalf(
					"expected job ID 12, got %d",
					jobID,
				)
			}

			return "completed", nil
		},
		getAcceptedCleanerIDFn: func(
			ctx context.Context,
			jobID uint,
		) (uint, error) {
			if jobID != 12 {
				t.Fatalf(
					"expected job ID 12, got %d",
					jobID,
				)
			}

			return 8, nil
		},
		createFn: func(
			ctx context.Context,
			review *Review,
		) error {
			createCalled = true

			if review.BookingID != 20 {
				t.Fatalf(
					"expected booking ID 20, got %d",
					review.BookingID,
				)
			}

			if review.JobID != 12 {
				t.Fatalf(
					"expected job ID 12, got %d",
					review.JobID,
				)
			}

			if review.CleanerID != 8 {
				t.Fatalf(
					"expected cleaner ID 8, got %d",
					review.CleanerID,
				)
			}

			if review.ClientID != 5 {
				t.Fatalf(
					"expected client ID 5, got %d",
					review.ClientID,
				)
			}

			if review.Rating != 5 {
				t.Fatalf(
					"expected rating 5, got %d",
					review.Rating,
				)
			}

			if review.Comment != "Excellent cleaner" {
				t.Fatalf(
					"expected trimmed comment, got %q",
					review.Comment,
				)
			}

			now := time.Now()

			review.ID = 30
			review.CreatedAt = now
			review.UpdatedAt = now

			return nil
		},
	}

	service := NewService(repo)

	review, err := service.Create(
		context.Background(),
		12,
		5,
		CreateReviewRequest{
			Rating:    5,
			Comment:   "  Excellent cleaner  ",
			CleanerID: 999,
			BookingID: 20,
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if review == nil {
		t.Fatal("expected review")
	}

	if review.ID != 30 {
		t.Fatalf(
			"expected review ID 30, got %d",
			review.ID,
		)
	}

	if review.CleanerID != 8 {
		t.Fatalf(
			"expected accepted cleaner ID 8, got %d",
			review.CleanerID,
		)
	}

	if review.Comment != "Excellent cleaner" {
		t.Fatalf(
			"expected trimmed comment, got %q",
			review.Comment,
		)
	}

	if !createCalled {
		t.Fatal("expected Create repository call")
	}
}

func TestService_Create_InvalidInput(t *testing.T) {
	service := NewService(&mockRepository{})

	tests := []struct {
		name     string
		jobID    uint
		clientID uint
		request  CreateReviewRequest
	}{
		{
			name:     "zero job ID",
			jobID:    0,
			clientID: 5,
			request: CreateReviewRequest{
				Rating:    5,
				Comment:   "Excellent",
				BookingID: 20,
			},
		},
		{
			name:     "zero client ID",
			jobID:    12,
			clientID: 0,
			request: CreateReviewRequest{
				Rating:    5,
				Comment:   "Excellent",
				BookingID: 20,
			},
		},
		{
			name:     "blank comment",
			jobID:    12,
			clientID: 5,
			request: CreateReviewRequest{
				Rating:    5,
				Comment:   "   ",
				BookingID: 20,
			},
		},
		{
			name:     "rating below one",
			jobID:    12,
			clientID: 5,
			request: CreateReviewRequest{
				Rating:    0,
				Comment:   "Poor",
				BookingID: 20,
			},
		},
		{
			name:     "rating above five",
			jobID:    12,
			clientID: 5,
			request: CreateReviewRequest{
				Rating:    6,
				Comment:   "Excellent",
				BookingID: 20,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			review, err := service.Create(
				context.Background(),
				test.jobID,
				test.clientID,
				test.request,
			)

			if review != nil {
				t.Fatalf(
					"expected nil review, got %+v",
					review,
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

func TestService_Create_GetJobClientIDError(t *testing.T) {
	expectedErr := errors.New("get job client failed")

	repo := &mockRepository{
		getJobClientIDFn: func(
			context.Context,
			uint,
		) (uint, error) {
			return 0, expectedErr
		},
	}

	service := NewService(repo)

	review, err := service.Create(
		context.Background(),
		12,
		5,
		CreateReviewRequest{
			Rating:    5,
			Comment:   "Excellent",
			BookingID: 20,
		},
	)

	if review != nil {
		t.Fatalf(
			"expected nil review, got %+v",
			review,
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

func TestService_Create_Forbidden(t *testing.T) {
	repo := &mockRepository{
		getJobClientIDFn: func(
			context.Context,
			uint,
		) (uint, error) {
			return 99, nil
		},
	}

	service := NewService(repo)

	review, err := service.Create(
		context.Background(),
		12,
		5,
		CreateReviewRequest{
			Rating:    5,
			Comment:   "Excellent",
			BookingID: 20,
		},
	)

	if review != nil {
		t.Fatalf(
			"expected nil review, got %+v",
			review,
		)
	}

	if !errors.Is(err, ErrForbidden) {
		t.Fatalf(
			"expected ErrForbidden, got %v",
			err,
		)
	}
}

func TestService_Create_GetJobStatusError(t *testing.T) {
	expectedErr := errors.New("get job status failed")

	repo := &mockRepository{
		getJobClientIDFn: func(
			context.Context,
			uint,
		) (uint, error) {
			return 5, nil
		},
		getJobStatusFn: func(
			context.Context,
			uint,
		) (string, error) {
			return "", expectedErr
		},
	}

	service := NewService(repo)

	review, err := service.Create(
		context.Background(),
		12,
		5,
		CreateReviewRequest{
			Rating:    5,
			Comment:   "Excellent",
			BookingID: 20,
		},
	)

	if review != nil {
		t.Fatalf(
			"expected nil review, got %+v",
			review,
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

func TestService_Create_JobNotCompleted(t *testing.T) {
	repo := &mockRepository{
		getJobClientIDFn: func(
			context.Context,
			uint,
		) (uint, error) {
			return 5, nil
		},
		getJobStatusFn: func(
			context.Context,
			uint,
		) (string, error) {
			return "open", nil
		},
	}

	service := NewService(repo)

	review, err := service.Create(
		context.Background(),
		12,
		5,
		CreateReviewRequest{
			Rating:    5,
			Comment:   "Excellent",
			BookingID: 20,
		},
	)

	if review != nil {
		t.Fatalf(
			"expected nil review, got %+v",
			review,
		)
	}

	if !errors.Is(err, ErrJobNotCompleted) {
		t.Fatalf(
			"expected ErrJobNotCompleted, got %v",
			err,
		)
	}
}

func TestService_Create_GetAcceptedCleanerIDError(t *testing.T) {
	expectedErr := errors.New("get accepted cleaner failed")

	repo := &mockRepository{
		getJobClientIDFn: func(
			context.Context,
			uint,
		) (uint, error) {
			return 5, nil
		},
		getJobStatusFn: func(
			context.Context,
			uint,
		) (string, error) {
			return "completed", nil
		},
		getAcceptedCleanerIDFn: func(
			context.Context,
			uint,
		) (uint, error) {
			return 0, expectedErr
		},
	}

	service := NewService(repo)

	review, err := service.Create(
		context.Background(),
		12,
		5,
		CreateReviewRequest{
			Rating:    5,
			Comment:   "Excellent",
			BookingID: 20,
		},
	)

	if review != nil {
		t.Fatalf(
			"expected nil review, got %+v",
			review,
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

func TestService_Create_MissingBookingID(t *testing.T) {
	createCalled := false

	repo := &mockRepository{
		getJobClientIDFn: func(
			context.Context,
			uint,
		) (uint, error) {
			return 5, nil
		},
		getJobStatusFn: func(
			context.Context,
			uint,
		) (string, error) {
			return "completed", nil
		},
		getAcceptedCleanerIDFn: func(
			context.Context,
			uint,
		) (uint, error) {
			return 8, nil
		},
		createFn: func(
			context.Context,
			*Review,
		) error {
			createCalled = true
			return nil
		},
	}

	service := NewService(repo)

	review, err := service.Create(
		context.Background(),
		12,
		5,
		CreateReviewRequest{
			Rating:  5,
			Comment: "Excellent",
		},
	)

	if review != nil {
		t.Fatalf(
			"expected nil review, got %+v",
			review,
		)
	}

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf(
			"expected ErrInvalidInput, got %v",
			err,
		)
	}

	if createCalled {
		t.Fatal("did not expect Create repository call")
	}
}

func TestService_Create_RepositoryError(t *testing.T) {
	expectedErr := errors.New("create review failed")

	repo := &mockRepository{
		getJobClientIDFn: func(
			context.Context,
			uint,
		) (uint, error) {
			return 5, nil
		},
		getJobStatusFn: func(
			context.Context,
			uint,
		) (string, error) {
			return "completed", nil
		},
		getAcceptedCleanerIDFn: func(
			context.Context,
			uint,
		) (uint, error) {
			return 8, nil
		},
		createFn: func(
			context.Context,
			*Review,
		) error {
			return expectedErr
		},
	}

	service := NewService(repo)

	review, err := service.Create(
		context.Background(),
		12,
		5,
		CreateReviewRequest{
			Rating:    5,
			Comment:   "Excellent",
			BookingID: 20,
		},
	)

	if review != nil {
		t.Fatalf(
			"expected nil review, got %+v",
			review,
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

func TestService_Create_AlreadyExists(t *testing.T) {
	repo := &mockRepository{
		getJobClientIDFn: func(
			context.Context,
			uint,
		) (uint, error) {
			return 5, nil
		},
		getJobStatusFn: func(
			context.Context,
			uint,
		) (string, error) {
			return "completed", nil
		},
		getAcceptedCleanerIDFn: func(
			context.Context,
			uint,
		) (uint, error) {
			return 8, nil
		},
		createFn: func(
			context.Context,
			*Review,
		) error {
			return ErrReviewAlreadyExist
		},
	}

	service := NewService(repo)

	review, err := service.Create(
		context.Background(),
		12,
		5,
		CreateReviewRequest{
			Rating:    5,
			Comment:   "Excellent",
			BookingID: 20,
		},
	)

	if review != nil {
		t.Fatalf(
			"expected nil review, got %+v",
			review,
		)
	}

	if !errors.Is(err, ErrReviewAlreadyExist) {
		t.Fatalf(
			"expected ErrReviewAlreadyExist, got %v",
			err,
		)
	}
}

func TestService_ListByCleanerID_Success(t *testing.T) {
	now := time.Now()

	repo := &mockRepository{
		listByCleanerIDFn: func(
			ctx context.Context,
			cleanerID uint,
		) ([]Review, error) {
			if cleanerID != 8 {
				t.Fatalf(
					"expected cleaner ID 8, got %d",
					cleanerID,
				)
			}

			return []Review{
				{
					ID:        1,
					BookingID: 20,
					CleanerID: 8,
					ClientID:  5,
					JobID:     12,
					Rating:    5,
					Comment:   "Excellent",
					CreatedAt: now,
					UpdatedAt: now,
				},
			}, nil
		},
	}

	service := NewService(repo)

	reviews, err := service.ListByCleanerID(
		context.Background(),
		8,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(reviews) != 1 {
		t.Fatalf(
			"expected 1 review, got %d",
			len(reviews),
		)
	}

	if reviews[0].BookingID != 20 {
		t.Fatalf(
			"expected booking ID 20, got %d",
			reviews[0].BookingID,
		)
	}
}

func TestService_ListByCleanerID_Empty(t *testing.T) {
	repo := &mockRepository{
		listByCleanerIDFn: func(
			context.Context,
			uint,
		) ([]Review, error) {
			return []Review{}, nil
		},
	}

	service := NewService(repo)

	reviews, err := service.ListByCleanerID(
		context.Background(),
		8,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(reviews) != 0 {
		t.Fatalf(
			"expected no reviews, got %d",
			len(reviews),
		)
	}
}

func TestService_ListByCleanerID_InvalidInput(t *testing.T) {
	service := NewService(&mockRepository{})

	reviews, err := service.ListByCleanerID(
		context.Background(),
		0,
	)

	if reviews != nil {
		t.Fatalf(
			"expected nil reviews, got %+v",
			reviews,
		)
	}

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf(
			"expected ErrInvalidInput, got %v",
			err,
		)
	}
}

func TestService_ListByCleanerID_RepositoryError(t *testing.T) {
	expectedErr := errors.New("list cleaner reviews failed")

	repo := &mockRepository{
		listByCleanerIDFn: func(
			context.Context,
			uint,
		) ([]Review, error) {
			return nil, expectedErr
		},
	}

	service := NewService(repo)

	reviews, err := service.ListByCleanerID(
		context.Background(),
		8,
	)

	if reviews != nil {
		t.Fatalf(
			"expected nil reviews, got %+v",
			reviews,
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
		listByClientIDFn: func(
			ctx context.Context,
			clientID uint,
		) ([]Review, error) {
			if clientID != 5 {
				t.Fatalf(
					"expected client ID 5, got %d",
					clientID,
				)
			}

			return []Review{
				{
					ID:        1,
					BookingID: 20,
					CleanerID: 8,
					ClientID:  5,
					JobID:     12,
					Rating:    5,
					Comment:   "Excellent",
					CreatedAt: now,
					UpdatedAt: now,
				},
			}, nil
		},
	}

	service := NewService(repo)

	reviews, err := service.ListMine(
		context.Background(),
		5,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(reviews) != 1 {
		t.Fatalf(
			"expected 1 review, got %d",
			len(reviews),
		)
	}

	if reviews[0].ClientID != 5 {
		t.Fatalf(
			"expected client ID 5, got %d",
			reviews[0].ClientID,
		)
	}
}

func TestService_ListMine_Empty(t *testing.T) {
	repo := &mockRepository{
		listByClientIDFn: func(
			context.Context,
			uint,
		) ([]Review, error) {
			return []Review{}, nil
		},
	}

	service := NewService(repo)

	reviews, err := service.ListMine(
		context.Background(),
		5,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(reviews) != 0 {
		t.Fatalf(
			"expected no reviews, got %d",
			len(reviews),
		)
	}
}

func TestService_ListMine_InvalidInput(t *testing.T) {
	service := NewService(&mockRepository{})

	reviews, err := service.ListMine(
		context.Background(),
		0,
	)

	if reviews != nil {
		t.Fatalf(
			"expected nil reviews, got %+v",
			reviews,
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
	expectedErr := errors.New("list client reviews failed")

	repo := &mockRepository{
		listByClientIDFn: func(
			context.Context,
			uint,
		) ([]Review, error) {
			return nil, expectedErr
		},
	}

	service := NewService(repo)

	reviews, err := service.ListMine(
		context.Background(),
		5,
	)

	if reviews != nil {
		t.Fatalf(
			"expected nil reviews, got %+v",
			reviews,
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
