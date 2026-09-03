package jobpulse

import (
	"context"
	"errors"
	"testing"
	"time"
)

type mockRepository struct {
	getSnapshotFn func(
		context.Context,
		uint,
	) (*JobPulseSnapshot, error)
}

func (m *mockRepository) GetSnapshot(
	ctx context.Context,
	jobID uint,
) (*JobPulseSnapshot, error) {
	if m.getSnapshotFn != nil {
		return m.getSnapshotFn(
			ctx,
			jobID,
		)
	}

	return nil,
		ErrJobNotFound
}

func TestService_GetJobPulse_Quiet(
	t *testing.T,
) {
	repo := &mockRepository{
		getSnapshotFn: func(
			ctx context.Context,
			jobID uint,
		) (*JobPulseSnapshot, error) {
			return &JobPulseSnapshot{
				JobID: jobID,

				JobStatus: "open",

				CreatedAt: time.Now().
					UTC().
					Add(-48 * time.Hour),

				ApplicationCount: 0,

				RecentApplicationCount: 0,

				BookingCount: 0,
			}, nil
		},
	}

	service := NewService(repo)

	result, err :=
		service.GetJobPulse(
			context.Background(),
			10,
		)
	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}

	if result.PulseLevel !=
		PulseQuiet {
		t.Fatalf(
			"expected quiet, got %q",
			result.PulseLevel,
		)
	}

	if result.ApplicationCount != 0 {
		t.Fatalf(
			"expected zero applications, got %d",
			result.ApplicationCount,
		)
	}

	if result.IsFilled {
		t.Fatal(
			"did not expect filled job",
		)
	}
}

func TestService_GetJobPulse_Active(
	t *testing.T,
) {
	repo := &mockRepository{
		getSnapshotFn: func(
			ctx context.Context,
			jobID uint,
		) (*JobPulseSnapshot, error) {
			return &JobPulseSnapshot{
				JobID: jobID,

				JobStatus: "open",

				CreatedAt: time.Now().
					UTC().
					Add(-12 * time.Hour),

				ApplicationCount: 2,

				RecentApplicationCount: 1,

				BookingCount: 0,
			}, nil
		},
	}

	service := NewService(repo)

	result, err :=
		service.GetJobPulse(
			context.Background(),
			10,
		)
	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}

	if result.PulseLevel !=
		PulseActive {
		t.Fatalf(
			"expected active, got %q with score %d",
			result.PulseLevel,
			result.PulseScore,
		)
	}
}

func TestService_GetJobPulse_HeatingUp(
	t *testing.T,
) {
	repo := &mockRepository{
		getSnapshotFn: func(
			ctx context.Context,
			jobID uint,
		) (*JobPulseSnapshot, error) {
			return &JobPulseSnapshot{
				JobID: jobID,

				JobStatus: "open",

				CreatedAt: time.Now().
					UTC().
					Add(-18 * time.Hour),

				ApplicationCount: 5,

				RecentApplicationCount: 3,

				BookingCount: 0,
			}, nil
		},
	}

	service := NewService(repo)

	result, err :=
		service.GetJobPulse(
			context.Background(),
			10,
		)
	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}

	if result.PulseLevel !=
		PulseHeatingUp {
		t.Fatalf(
			"expected heating_up, got %q with score %d",
			result.PulseLevel,
			result.PulseScore,
		)
	}

	if result.RecentApplicationCount != 3 {
		t.Fatalf(
			"expected 3 recent applications, got %d",
			result.RecentApplicationCount,
		)
	}
}

func TestService_GetJobPulse_HighCompetition(
	t *testing.T,
) {
	repo := &mockRepository{
		getSnapshotFn: func(
			ctx context.Context,
			jobID uint,
		) (*JobPulseSnapshot, error) {
			return &JobPulseSnapshot{
				JobID: jobID,

				JobStatus: "open",

				CreatedAt: time.Now().
					UTC().
					Add(-5 * time.Hour),

				ApplicationCount: 15,

				RecentApplicationCount: 8,

				BookingCount: 0,
			}, nil
		},
	}

	service := NewService(repo)

	result, err :=
		service.GetJobPulse(
			context.Background(),
			10,
		)
	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}

	if result.PulseLevel !=
		PulseCompetitive {
		t.Fatalf(
			"expected high_competition, got %q",
			result.PulseLevel,
		)
	}

	if result.PulseScore != 100 {
		t.Fatalf(
			"expected score 100, got %d",
			result.PulseScore,
		)
	}
}

func TestService_GetJobPulse_BookingMakesFilled(
	t *testing.T,
) {
	repo := &mockRepository{
		getSnapshotFn: func(
			ctx context.Context,
			jobID uint,
		) (*JobPulseSnapshot, error) {
			return &JobPulseSnapshot{
				JobID: jobID,

				JobStatus: "open",

				CreatedAt: time.Now().
					UTC().
					Add(-24 * time.Hour),

				ApplicationCount: 4,

				RecentApplicationCount: 1,

				BookingCount: 1,
			}, nil
		},
	}

	service := NewService(repo)

	result, err :=
		service.GetJobPulse(
			context.Background(),
			10,
		)
	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}

	if !result.HasBooking {
		t.Fatal(
			"expected booking",
		)
	}

	if !result.IsFilled {
		t.Fatal(
			"expected filled job",
		)
	}

	if result.PulseScore != 100 {
		t.Fatalf(
			"expected filled score 100, got %d",
			result.PulseScore,
		)
	}
}

func TestService_GetJobPulse_ClosedStatusMakesFilled(
	t *testing.T,
) {
	repo := &mockRepository{
		getSnapshotFn: func(
			ctx context.Context,
			jobID uint,
		) (*JobPulseSnapshot, error) {
			return &JobPulseSnapshot{
				JobID: jobID,

				JobStatus: "closed",

				CreatedAt: time.Now().
					UTC().
					Add(-72 * time.Hour),
			}, nil
		},
	}

	service := NewService(repo)

	result, err :=
		service.GetJobPulse(
			context.Background(),
			10,
		)
	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}

	if !result.IsFilled {
		t.Fatal(
			"expected closed job to be filled",
		)
	}
}

func TestService_GetJobPulse_InvalidJobID(
	t *testing.T,
) {
	service := NewService(
		&mockRepository{},
	)

	result, err :=
		service.GetJobPulse(
			context.Background(),
			0,
		)

	if result != nil {
		t.Fatalf(
			"expected nil result, got %+v",
			result,
		)
	}

	if !errors.Is(
		err,
		ErrInvalidInput,
	) {
		t.Fatalf(
			"expected ErrInvalidInput, got %v",
			err,
		)
	}
}

func TestService_GetJobPulse_NotFound(
	t *testing.T,
) {
	service := NewService(
		&mockRepository{},
	)

	result, err :=
		service.GetJobPulse(
			context.Background(),
			10,
		)

	if result != nil {
		t.Fatalf(
			"expected nil result, got %+v",
			result,
		)
	}

	if !errors.Is(
		err,
		ErrJobNotFound,
	) {
		t.Fatalf(
			"expected ErrJobNotFound, got %v",
			err,
		)
	}
}

func TestCalculatePulseScore(
	t *testing.T,
) {
	tests := []struct {
		name     string
		apps     int
		recent   int
		age      int
		filled   bool
		expected int
	}{
		{
			name:     "quiet old job",
			apps:     0,
			recent:   0,
			age:      100,
			expected: 0,
		},
		{
			name:     "single fresh application",
			apps:     1,
			recent:   1,
			age:      5,
			expected: 37,
		},
		{
			name:     "heating up",
			apps:     5,
			recent:   3,
			age:      18,
			expected: 62,
		},
		{
			name:     "high competition",
			apps:     15,
			recent:   8,
			age:      5,
			expected: 100,
		},
		{
			name:     "filled",
			apps:     0,
			recent:   0,
			age:      100,
			filled:   true,
			expected: 100,
		},
	}

	for _, test := range tests {
		t.Run(
			test.name,
			func(t *testing.T) {
				result :=
					calculatePulseScore(
						test.apps,
						test.recent,
						test.age,
						test.filled,
					)

				if result !=
					test.expected {
					t.Fatalf(
						"expected %d, got %d",
						test.expected,
						result,
					)
				}
			},
		)
	}
}

func TestJobStatusIsFilled(
	t *testing.T,
) {
	tests := []struct {
		status   string
		expected bool
	}{
		{"open", false},
		{"filled", true},
		{"booked", true},
		{"completed", true},
		{"closed", true},
		{"cancelled", false},
	}

	for _, test := range tests {
		result :=
			jobStatusIsFilled(
				test.status,
			)

		if result !=
			test.expected {
			t.Fatalf(
				"status %q expected %v, got %v",
				test.status,
				test.expected,
				result,
			)
		}
	}
}
