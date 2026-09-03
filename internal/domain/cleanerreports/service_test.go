package cleanerreports

import (
	"context"
	"errors"
	"testing"
	"time"
)

type mockRepository struct {
	createFn func(
		context.Context,
		*CleanerReport,
	) error

	listByClientIDFn func(
		context.Context,
		uint,
	) ([]CleanerReport, error)
}

func (m *mockRepository) Create(
	ctx context.Context,
	report *CleanerReport,
) error {
	if m.createFn != nil {
		return m.createFn(
			ctx,
			report,
		)
	}

	return nil
}

func (m *mockRepository) ListByClientID(
	ctx context.Context,
	clientID uint,
) ([]CleanerReport, error) {
	if m.listByClientIDFn != nil {
		return m.listByClientIDFn(
			ctx,
			clientID,
		)
	}

	return nil, nil
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
		createFn: func(
			ctx context.Context,
			report *CleanerReport,
		) error {
			createCalled = true

			if report.ClientID != 5 {
				t.Fatalf(
					"expected client ID 5, got %d",
					report.ClientID,
				)
			}

			if report.CleanerID != 8 {
				t.Fatalf(
					"expected cleaner ID 8, got %d",
					report.CleanerID,
				)
			}

			if report.Reason != "Poor behaviour" {
				t.Fatalf(
					"expected trimmed reason, got %q",
					report.Reason,
				)
			}

			if report.Details != "Cleaner was rude to the client." {
				t.Fatalf(
					"expected trimmed details, got %q",
					report.Details,
				)
			}

			if report.Status != "open" {
				t.Fatalf(
					"expected status open, got %q",
					report.Status,
				)
			}

			report.ID = 20
			report.CreatedAt = time.Now()
			report.UpdatedAt = report.CreatedAt

			return nil
		},
	}

	service := NewService(repo)

	report, err := service.Create(
		context.Background(),
		5,
		8,
		CreateCleanerReportRequest{
			Reason:  "  Poor behaviour  ",
			Details: "  Cleaner was rude to the client.  ",
		},
	)

	if err != nil {
		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}

	if !createCalled {
		t.Fatal(
			"expected repository Create to be called",
		)
	}

	if report == nil {
		t.Fatal("expected report")
	}

	if report.ID != 20 {
		t.Fatalf(
			"expected report ID 20, got %d",
			report.ID,
		)
	}

	if report.ClientID != 5 {
		t.Fatalf(
			"expected client ID 5, got %d",
			report.ClientID,
		)
	}

	if report.CleanerID != 8 {
		t.Fatalf(
			"expected cleaner ID 8, got %d",
			report.CleanerID,
		)
	}

	if report.Reason != "Poor behaviour" {
		t.Fatalf(
			"expected trimmed reason, got %q",
			report.Reason,
		)
	}

	if report.Details != "Cleaner was rude to the client." {
		t.Fatalf(
			"expected trimmed details, got %q",
			report.Details,
		)
	}

	if report.Status != "open" {
		t.Fatalf(
			"expected status open, got %q",
			report.Status,
		)
	}
}

func TestService_Create_EmptyDetailsAllowed(
	t *testing.T,
) {
	repo := &mockRepository{
		createFn: func(
			ctx context.Context,
			report *CleanerReport,
		) error {
			if report.Details != "" {
				t.Fatalf(
					"expected empty details, got %q",
					report.Details,
				)
			}

			return nil
		},
	}

	service := NewService(repo)

	report, err := service.Create(
		context.Background(),
		5,
		8,
		CreateCleanerReportRequest{
			Reason: "No show",
		},
	)

	if err != nil {
		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}

	if report == nil {
		t.Fatal("expected report")
	}

	if report.Reason != "No show" {
		t.Fatalf(
			"expected reason No show, got %q",
			report.Reason,
		)
	}
}

func TestService_Create_InvalidInput(
	t *testing.T,
) {
	tests := []struct {
		name      string
		clientID  uint
		cleanerID uint
		req       CreateCleanerReportRequest
	}{
		{
			name:      "zero client ID",
			clientID:  0,
			cleanerID: 8,
			req: CreateCleanerReportRequest{
				Reason: "No show",
			},
		},
		{
			name:      "zero cleaner ID",
			clientID:  5,
			cleanerID: 0,
			req: CreateCleanerReportRequest{
				Reason: "No show",
			},
		},
		{
			name:      "empty reason",
			clientID:  5,
			cleanerID: 8,
			req: CreateCleanerReportRequest{
				Reason: "",
			},
		},
		{
			name:      "whitespace reason",
			clientID:  5,
			cleanerID: 8,
			req: CreateCleanerReportRequest{
				Reason: "   ",
			},
		},
		{
			name:      "same client and cleaner",
			clientID:  5,
			cleanerID: 5,
			req: CreateCleanerReportRequest{
				Reason: "No show",
			},
		},
	}

	for _, test := range tests {
		t.Run(
			test.name,
			func(t *testing.T) {
				createCalled := false

				repo := &mockRepository{
					createFn: func(
						context.Context,
						*CleanerReport,
					) error {
						createCalled = true
						return nil
					},
				}

				service := NewService(repo)

				report, err := service.Create(
					context.Background(),
					test.clientID,
					test.cleanerID,
					test.req,
				)

				if report != nil {
					t.Fatalf(
						"expected nil report, got %+v",
						report,
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

				if createCalled {
					t.Fatal(
						"repository Create should not be called",
					)
				}
			},
		)
	}
}

func TestService_Create_RepositoryError(
	t *testing.T,
) {
	expectedErr := errors.New(
		"create report failed",
	)

	repo := &mockRepository{
		createFn: func(
			context.Context,
			*CleanerReport,
		) error {
			return expectedErr
		},
	}

	service := NewService(repo)

	report, err := service.Create(
		context.Background(),
		5,
		8,
		CreateCleanerReportRequest{
			Reason:  "No show",
			Details: "Cleaner did not attend.",
		},
	)

	if report != nil {
		t.Fatalf(
			"expected nil report, got %+v",
			report,
		)
	}

	if !errors.Is(
		err,
		expectedErr,
	) {
		t.Fatalf(
			"expected %v, got %v",
			expectedErr,
			err,
		)
	}
}

func TestService_ListMine_Success(
	t *testing.T,
) {
	now := time.Now()

	repo := &mockRepository{
		listByClientIDFn: func(
			ctx context.Context,
			clientID uint,
		) ([]CleanerReport, error) {
			if clientID != 5 {
				t.Fatalf(
					"expected client ID 5, got %d",
					clientID,
				)
			}

			return []CleanerReport{
				{
					ID:        2,
					ClientID:  5,
					CleanerID: 9,
					Reason:    "Poor behaviour",
					Details:   "Details",
					Status:    "open",
					CreatedAt: now,
					UpdatedAt: now,
				},
				{
					ID:        1,
					ClientID:  5,
					CleanerID: 8,
					Reason:    "No show",
					Details:   "Details",
					Status:    "closed",
					CreatedAt: now.Add(-time.Hour),
					UpdatedAt: now,
				},
			}, nil
		},
	}

	service := NewService(repo)

	reports, err := service.ListMine(
		context.Background(),
		5,
	)

	if err != nil {
		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}

	if len(reports) != 2 {
		t.Fatalf(
			"expected 2 reports, got %d",
			len(reports),
		)
	}

	if reports[0].ID != 2 {
		t.Fatalf(
			"expected first report ID 2, got %d",
			reports[0].ID,
		)
	}

	if reports[1].ID != 1 {
		t.Fatalf(
			"expected second report ID 1, got %d",
			reports[1].ID,
		)
	}
}

func TestService_ListMine_Empty(
	t *testing.T,
) {
	repo := &mockRepository{
		listByClientIDFn: func(
			context.Context,
			uint,
		) ([]CleanerReport, error) {
			return []CleanerReport{}, nil
		},
	}

	service := NewService(repo)

	reports, err := service.ListMine(
		context.Background(),
		5,
	)

	if err != nil {
		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}

	if len(reports) != 0 {
		t.Fatalf(
			"expected no reports, got %d",
			len(reports),
		)
	}
}

func TestService_ListMine_InvalidInput(
	t *testing.T,
) {
	service := NewService(
		&mockRepository{},
	)

	reports, err := service.ListMine(
		context.Background(),
		0,
	)

	if reports != nil {
		t.Fatalf(
			"expected nil reports, got %+v",
			reports,
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

func TestService_ListMine_RepositoryError(
	t *testing.T,
) {
	expectedErr := errors.New(
		"list reports failed",
	)

	repo := &mockRepository{
		listByClientIDFn: func(
			context.Context,
			uint,
		) ([]CleanerReport, error) {
			return nil, expectedErr
		},
	}

	service := NewService(repo)

	reports, err := service.ListMine(
		context.Background(),
		5,
	)

	if reports != nil {
		t.Fatalf(
			"expected nil reports, got %+v",
			reports,
		)
	}

	if !errors.Is(
		err,
		expectedErr,
	) {
		t.Fatalf(
			"expected %v, got %v",
			expectedErr,
			err,
		)
	}
}
