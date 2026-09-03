package wokrproof

import (
	"context"
	"errors"
	"testing"
	"time"
)

type mockRepository struct {
	getBookingFn func(
		context.Context,
		uint,
	) (*BookingSnapshot, error)

	createFn func(
		context.Context,
		*WorkProof,
	) error

	getByIDFn func(
		context.Context,
		uint,
	) (*WorkProof, error)

	listByBookingIDFn func(
		context.Context,
		uint,
	) ([]WorkProof, error)

	countByBookingAndTypeFn func(
		context.Context,
		uint,
		ProofType,
	) (int, error)

	deleteFn func(
		context.Context,
		uint,
		uint,
	) error

	getVerifiedWorkSummaryFn func(
		context.Context,
		uint,
	) (*VerifiedWorkSummary, error)
}

func (m *mockRepository) GetBooking(
	ctx context.Context,
	bookingID uint,
) (*BookingSnapshot, error) {
	if m.getBookingFn != nil {
		return m.getBookingFn(
			ctx,
			bookingID,
		)
	}

	return nil, ErrBookingNotFound
}

func (m *mockRepository) Create(
	ctx context.Context,
	proof *WorkProof,
) error {
	if m.createFn != nil {
		return m.createFn(
			ctx,
			proof,
		)
	}

	return nil
}

func (m *mockRepository) GetByID(
	ctx context.Context,
	proofID uint,
) (*WorkProof, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(
			ctx,
			proofID,
		)
	}

	return nil, ErrProofNotFound
}

func (m *mockRepository) ListByBookingID(
	ctx context.Context,
	bookingID uint,
) ([]WorkProof, error) {
	if m.listByBookingIDFn != nil {
		return m.listByBookingIDFn(
			ctx,
			bookingID,
		)
	}

	return []WorkProof{}, nil
}

func (m *mockRepository) CountByBookingAndType(
	ctx context.Context,
	bookingID uint,
	proofType ProofType,
) (int, error) {
	if m.countByBookingAndTypeFn != nil {
		return m.countByBookingAndTypeFn(
			ctx,
			bookingID,
			proofType,
		)
	}

	return 0, nil
}

func (m *mockRepository) Delete(
	ctx context.Context,
	proofID uint,
	cleanerID uint,
) error {
	if m.deleteFn != nil {
		return m.deleteFn(
			ctx,
			proofID,
			cleanerID,
		)
	}

	return nil
}

func (m *mockRepository) GetVerifiedWorkSummary(
	ctx context.Context,
	cleanerID uint,
) (*VerifiedWorkSummary, error) {
	if m.getVerifiedWorkSummaryFn != nil {
		return m.getVerifiedWorkSummaryFn(
			ctx,
			cleanerID,
		)
	}

	return &VerifiedWorkSummary{
		CleanerID: cleanerID,
	}, nil
}

func TestService_Create_BeforeProof_Success(t *testing.T) {
	repo := &mockRepository{
		getBookingFn: func(
			context.Context,
			uint,
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
			context.Context,
			uint,
			ProofType,
		) (int, error) {
			return 2, nil
		},

		createFn: func(
			ctx context.Context,
			proof *WorkProof,
		) error {
			proof.ID = 30
			proof.CreatedAt =
				time.Now().UTC()

			return nil
		},
	}

	service := NewService(repo)

	result, err := service.Create(
		context.Background(),
		10,
		8,
		CreateWorkProofRequest{
			ProofType: " before ",
			PhotoURL:  " https://example.com/before.jpg ",
			Caption:   " Kitchen before ",
		},
	)
	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}

	if result == nil {
		t.Fatal("expected proof")
	}

	if result.ID != 30 {
		t.Fatalf(
			"expected ID 30, got %d",
			result.ID,
		)
	}

	if result.ProofType != ProofTypeBefore {
		t.Fatalf(
			"expected before proof, got %q",
			result.ProofType,
		)
	}

	if result.PhotoURL !=
		"https://example.com/before.jpg" {
		t.Fatalf(
			"unexpected photo URL %q",
			result.PhotoURL,
		)
	}
}

func TestService_Create_ForbiddenCleaner(t *testing.T) {
	repo := &mockRepository{
		getBookingFn: func(
			context.Context,
			uint,
		) (*BookingSnapshot, error) {
			return &BookingSnapshot{
				ID:        10,
				ClientID:  5,
				CleanerID: 8,
				Status:    "in_progress",
			}, nil
		},
	}

	service := NewService(repo)

	result, err := service.Create(
		context.Background(),
		10,
		99,
		CreateWorkProofRequest{
			ProofType: "before",
			PhotoURL:  "https://example.com/photo.jpg",
		},
	)

	if result != nil {
		t.Fatalf(
			"expected nil result, got %+v",
			result,
		)
	}

	if !errors.Is(
		err,
		ErrForbidden,
	) {
		t.Fatalf(
			"expected ErrForbidden, got %v",
			err,
		)
	}
}

func TestService_Create_InvalidProofType(t *testing.T) {
	service := NewService(
		&mockRepository{},
	)

	result, err := service.Create(
		context.Background(),
		10,
		8,
		CreateWorkProofRequest{
			ProofType: "during",
			PhotoURL:  "https://example.com/photo.jpg",
		},
	)

	if result != nil {
		t.Fatalf(
			"expected nil result, got %+v",
			result,
		)
	}

	if !errors.Is(
		err,
		ErrInvalidProofType,
	) {
		t.Fatalf(
			"expected ErrInvalidProofType, got %v",
			err,
		)
	}
}

func TestService_Create_BookingNotEligible(t *testing.T) {
	repo := &mockRepository{
		getBookingFn: func(
			context.Context,
			uint,
		) (*BookingSnapshot, error) {
			return &BookingSnapshot{
				ID:        10,
				ClientID:  5,
				CleanerID: 8,
				Status:    "pending",
			}, nil
		},
	}

	service := NewService(repo)

	result, err := service.Create(
		context.Background(),
		10,
		8,
		CreateWorkProofRequest{
			ProofType: "before",
			PhotoURL:  "https://example.com/photo.jpg",
		},
	)

	if result != nil {
		t.Fatalf(
			"expected nil result, got %+v",
			result,
		)
	}

	if !errors.Is(
		err,
		ErrBookingNotEligible,
	) {
		t.Fatalf(
			"expected ErrBookingNotEligible, got %v",
			err,
		)
	}
}

func TestService_Create_ProofLimitReached(t *testing.T) {
	repo := &mockRepository{
		getBookingFn: func(
			context.Context,
			uint,
		) (*BookingSnapshot, error) {
			return &BookingSnapshot{
				ID:        10,
				ClientID:  5,
				CleanerID: 8,
				Status:    "in_progress",
			}, nil
		},

		countByBookingAndTypeFn: func(
			context.Context,
			uint,
			ProofType,
		) (int, error) {
			return 10, nil
		},
	}

	service := NewService(repo)

	result, err := service.Create(
		context.Background(),
		10,
		8,
		CreateWorkProofRequest{
			ProofType: "before",
			PhotoURL:  "https://example.com/photo.jpg",
		},
	)

	if result != nil {
		t.Fatalf(
			"expected nil result, got %+v",
			result,
		)
	}

	if !errors.Is(
		err,
		ErrProofLimitReached,
	) {
		t.Fatalf(
			"expected ErrProofLimitReached, got %v",
			err,
		)
	}
}

func TestService_GetBookingProof_VerifiedWork(
	t *testing.T,
) {
	now := time.Now().UTC()

	repo := &mockRepository{
		getBookingFn: func(
			context.Context,
			uint,
		) (*BookingSnapshot, error) {
			return &BookingSnapshot{
				ID:        10,
				ClientID:  5,
				CleanerID: 8,
				Status:    "completed",
			}, nil
		},

		listByBookingIDFn: func(
			context.Context,
			uint,
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

	service := NewService(repo)

	result, err :=
		service.GetBookingProof(
			context.Background(),
			10,
			5,
		)
	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}

	if !result.IsVerifiedWork {
		t.Fatal(
			"expected verified work",
		)
	}

	if result.BeforeCount != 1 {
		t.Fatalf(
			"expected 1 before photo, got %d",
			result.BeforeCount,
		)
	}

	if result.AfterCount != 1 {
		t.Fatalf(
			"expected 1 after photo, got %d",
			result.AfterCount,
		)
	}

	if result.LastUploadedAt == nil {
		t.Fatal(
			"expected last uploaded time",
		)
	}
}

func TestService_GetBookingProof_Forbidden(t *testing.T) {
	repo := &mockRepository{
		getBookingFn: func(
			context.Context,
			uint,
		) (*BookingSnapshot, error) {
			return &BookingSnapshot{
				ID:        10,
				ClientID:  5,
				CleanerID: 8,
			}, nil
		},
	}

	service := NewService(repo)

	result, err :=
		service.GetBookingProof(
			context.Background(),
			10,
			99,
		)

	if result != nil {
		t.Fatalf(
			"expected nil result, got %+v",
			result,
		)
	}

	if !errors.Is(
		err,
		ErrForbidden,
	) {
		t.Fatalf(
			"expected ErrForbidden, got %v",
			err,
		)
	}
}

func TestService_Delete_Success(t *testing.T) {
	deleteCalled := false

	repo := &mockRepository{
		getByIDFn: func(
			context.Context,
			uint,
		) (*WorkProof, error) {
			return &WorkProof{
				ID:        30,
				BookingID: 10,
				CleanerID: 8,
			}, nil
		},

		deleteFn: func(
			context.Context,
			uint,
			uint,
		) error {
			deleteCalled = true
			return nil
		},
	}

	service := NewService(repo)

	err := service.Delete(
		context.Background(),
		30,
		8,
	)
	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}

	if !deleteCalled {
		t.Fatal(
			"expected delete call",
		)
	}
}

func TestService_Delete_Forbidden(t *testing.T) {
	repo := &mockRepository{
		getByIDFn: func(
			context.Context,
			uint,
		) (*WorkProof, error) {
			return &WorkProof{
				ID:        30,
				CleanerID: 8,
			}, nil
		},
	}

	service := NewService(repo)

	err := service.Delete(
		context.Background(),
		30,
		99,
	)

	if !errors.Is(
		err,
		ErrForbidden,
	) {
		t.Fatalf(
			"expected ErrForbidden, got %v",
			err,
		)
	}
}

func TestService_GetVerifiedWork_Success(t *testing.T) {
	repo := &mockRepository{
		getVerifiedWorkSummaryFn: func(
			context.Context,
			uint,
		) (*VerifiedWorkSummary, error) {
			return &VerifiedWorkSummary{
				CleanerID:    8,
				VerifiedJobs: 12,
				BeforePhotos: 20,
				AfterPhotos:  22,
				TotalPhotos:  42,
			}, nil
		},
	}

	service := NewService(repo)

	result, err :=
		service.GetVerifiedWork(
			context.Background(),
			8,
		)
	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}

	if result.VerifiedJobs != 12 {
		t.Fatalf(
			"expected 12 verified jobs, got %d",
			result.VerifiedJobs,
		)
	}
}

func TestBookingAllowsProof(t *testing.T) {
	tests := []struct {
		status   string
		expected bool
	}{
		{"confirmed", true},
		{"in_progress", true},
		{"completed", true},
		{"closed", true},
		{"pending", false},
		{"cancelled", false},
		{"", false},
	}

	for _, test := range tests {
		result :=
			bookingAllowsProof(
				test.status,
			)

		if result != test.expected {
			t.Fatalf(
				"status %q expected %v, got %v",
				test.status,
				test.expected,
				result,
			)
		}
	}
}
