package adminmoderation

import (
	"context"
	"errors"
	"testing"
	"time"
)

type mockRepository struct {
	listReportsFn func(
		context.Context,
	) ([]CleanerReportAdminView, error)

	listOpenReportsFn func(
		context.Context,
	) ([]CleanerReportAdminView, error)

	updateReportStatusFn func(
		context.Context,
		uint,
		uint,
		string,
		string,
	) error

	listBlockCleanersFn func(
		context.Context,
	) ([]BlockedCleanerAdminView, error)
}

func (m *mockRepository) ListReports(
	ctx context.Context,
) ([]CleanerReportAdminView, error) {
	if m.listReportsFn != nil {
		return m.listReportsFn(ctx)
	}

	return nil, nil
}

func (m *mockRepository) ListOpenReports(
	ctx context.Context,
) ([]CleanerReportAdminView, error) {
	if m.listOpenReportsFn != nil {
		return m.listOpenReportsFn(ctx)
	}

	return nil, nil
}

func (m *mockRepository) UpdateReportStatus(
	ctx context.Context,
	reportID uint,
	adminID uint,
	status string,
	adminNote string,
) error {
	if m.updateReportStatusFn != nil {
		return m.updateReportStatusFn(
			ctx,
			reportID,
			adminID,
			status,
			adminNote,
		)
	}

	return nil
}

func (m *mockRepository) ListBlockCleaners(
	ctx context.Context,
) ([]BlockedCleanerAdminView, error) {
	if m.listBlockCleanersFn != nil {
		return m.listBlockCleanersFn(ctx)
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

func TestService_ListReports_Success(t *testing.T) {
	now := time.Now()

	repo := &mockRepository{
		listReportsFn: func(
			context.Context,
		) ([]CleanerReportAdminView, error) {
			return []CleanerReportAdminView{
				{
					ID:        1,
					ClientID:  5,
					CleanerID: 8,
					Reason:    "No show",
					Status:    "open",
					CreatedAt: now,
					UpdatedAt: now,
				},
				{
					ID:        2,
					ClientID:  6,
					CleanerID: 9,
					Reason:    "Poor behaviour",
					Status:    "resolved",
					CreatedAt: now,
					UpdatedAt: now,
				},
			}, nil
		},
	}

	service := NewService(repo)

	reports, err := service.ListReports(
		context.Background(),
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

	if reports[0].ID != 1 {
		t.Fatalf(
			"expected first report ID 1, got %d",
			reports[0].ID,
		)
	}
}

func TestService_ListReports_RepositoryError(
	t *testing.T,
) {
	expectedErr := errors.New(
		"list reports failed",
	)

	repo := &mockRepository{
		listReportsFn: func(
			context.Context,
		) ([]CleanerReportAdminView, error) {
			return nil, expectedErr
		},
	}

	service := NewService(repo)

	reports, err := service.ListReports(
		context.Background(),
	)

	if reports != nil {
		t.Fatalf(
			"expected nil reports, got %+v",
			reports,
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

func TestService_ListOpenReports_Success(
	t *testing.T,
) {
	repo := &mockRepository{
		listOpenReportsFn: func(
			context.Context,
		) ([]CleanerReportAdminView, error) {
			return []CleanerReportAdminView{
				{
					ID:        1,
					ClientID:  5,
					CleanerID: 8,
					Status:    "open",
				},
			}, nil
		},
	}

	service := NewService(repo)

	reports, err := service.ListOpenReports(
		context.Background(),
	)

	if err != nil {
		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}

	if len(reports) != 1 {
		t.Fatalf(
			"expected 1 report, got %d",
			len(reports),
		)
	}

	if reports[0].Status != "open" {
		t.Fatalf(
			"expected status open, got %q",
			reports[0].Status,
		)
	}
}

func TestService_ListOpenReports_RepositoryError(
	t *testing.T,
) {
	expectedErr := errors.New(
		"list open reports failed",
	)

	repo := &mockRepository{
		listOpenReportsFn: func(
			context.Context,
		) ([]CleanerReportAdminView, error) {
			return nil, expectedErr
		},
	}

	service := NewService(repo)

	reports, err := service.ListOpenReports(
		context.Background(),
	)

	if reports != nil {
		t.Fatalf(
			"expected nil reports, got %+v",
			reports,
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

func TestService_ListBlockCleaners_Success(
	t *testing.T,
) {
	repo := &mockRepository{
		listBlockCleanersFn: func(
			context.Context,
		) ([]BlockedCleanerAdminView, error) {
			return []BlockedCleanerAdminView{
				{
					ID:        1,
					ClientID:  5,
					CleanerID: 8,
					Reason:    "Blocked",
				},
			}, nil
		},
	}

	service := NewService(repo)

	blocks, err := service.ListBlockCleaners(
		context.Background(),
	)

	if err != nil {
		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}

	if len(blocks) != 1 {
		t.Fatalf(
			"expected 1 block, got %d",
			len(blocks),
		)
	}

	if blocks[0].CleanerID != 8 {
		t.Fatalf(
			"expected cleaner ID 8, got %d",
			blocks[0].CleanerID,
		)
	}
}

func TestService_ListBlockCleaners_RepositoryError(
	t *testing.T,
) {
	expectedErr := errors.New(
		"list blocked cleaners failed",
	)

	repo := &mockRepository{
		listBlockCleanersFn: func(
			context.Context,
		) ([]BlockedCleanerAdminView, error) {
			return nil, expectedErr
		},
	}

	service := NewService(repo)

	blocks, err := service.ListBlockCleaners(
		context.Background(),
	)

	if blocks != nil {
		t.Fatalf(
			"expected nil blocks, got %+v",
			blocks,
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

func TestService_UpdateReportStatus_Success(
	t *testing.T,
) {
	updateCalled := false

	repo := &mockRepository{
		updateReportStatusFn: func(
			ctx context.Context,
			reportID uint,
			adminID uint,
			status string,
			adminNote string,
		) error {
			updateCalled = true

			if reportID != 12 {
				t.Fatalf(
					"expected report ID 12, got %d",
					reportID,
				)
			}

			if adminID != 9 {
				t.Fatalf(
					"expected admin ID 9, got %d",
					adminID,
				)
			}

			if status != "resolved" {
				t.Fatalf(
					"expected status resolved, got %q",
					status,
				)
			}

			if adminNote != "Issue reviewed" {
				t.Fatalf(
					"expected trimmed admin note, got %q",
					adminNote,
				)
			}

			return nil
		},
	}

	service := NewService(repo)

	err := service.UpdateReportStatus(
		context.Background(),
		12,
		9,
		UpdateReportStatusRequest{
			Status:    "  resolved  ",
			AdminNote: "  Issue reviewed  ",
		},
	)

	if err != nil {
		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}

	if !updateCalled {
		t.Fatal(
			"expected repository UpdateReportStatus to be called",
		)
	}
}

func TestService_UpdateReportStatus_AllAllowedStatuses(
	t *testing.T,
) {
	statuses := []string{
		"open",
		"resolved",
		"dismissed",
	}

	for _, status := range statuses {
		t.Run(
			status,
			func(t *testing.T) {
				repo := &mockRepository{
					updateReportStatusFn: func(
						ctx context.Context,
						reportID uint,
						adminID uint,
						actualStatus string,
						adminNote string,
					) error {
						if actualStatus != status {
							t.Fatalf(
								"expected status %q, got %q",
								status,
								actualStatus,
							)
						}

						return nil
					},
				}

				service := NewService(repo)

				err := service.UpdateReportStatus(
					context.Background(),
					12,
					9,
					UpdateReportStatusRequest{
						Status: status,
					},
				)

				if err != nil {
					t.Fatalf(
						"unexpected error: %v",
						err,
					)
				}
			},
		)
	}
}

func TestService_UpdateReportStatus_InvalidInput(
	t *testing.T,
) {
	tests := []struct {
		name     string
		reportID uint
		adminID  uint
	}{
		{
			name:     "zero report ID",
			reportID: 0,
			adminID:  9,
		},
		{
			name:     "zero admin ID",
			reportID: 12,
			adminID:  0,
		},
	}

	for _, test := range tests {
		t.Run(
			test.name,
			func(t *testing.T) {
				service := NewService(
					&mockRepository{},
				)

				err := service.UpdateReportStatus(
					context.Background(),
					test.reportID,
					test.adminID,
					UpdateReportStatusRequest{
						Status: "resolved",
					},
				)

				if !errors.Is(
					err,
					ErrInvalidInput,
				) {
					t.Fatalf(
						"expected ErrInvalidInput, got %v",
						err,
					)
				}
			},
		)
	}
}

func TestService_UpdateReportStatus_InvalidStatus(
	t *testing.T,
) {
	updateCalled := false

	repo := &mockRepository{
		updateReportStatusFn: func(
			context.Context,
			uint,
			uint,
			string,
			string,
		) error {
			updateCalled = true
			return nil
		},
	}

	service := NewService(repo)

	tests := []string{
		"",
		"pending",
		"closed",
		"invalid",
		"   ",
	}

	for _, status := range tests {
		t.Run(
			status,
			func(t *testing.T) {
				err := service.UpdateReportStatus(
					context.Background(),
					12,
					9,
					UpdateReportStatusRequest{
						Status: status,
					},
				)

				if !errors.Is(
					err,
					ErrInvalidStatus,
				) {
					t.Fatalf(
						"expected ErrInvalidStatus, got %v",
						err,
					)
				}
			},
		)
	}

	if updateCalled {
		t.Fatal(
			"repository should not be called for invalid status",
		)
	}
}

func TestService_UpdateReportStatus_RepositoryError(
	t *testing.T,
) {
	expectedErr := errors.New(
		"update failed",
	)

	repo := &mockRepository{
		updateReportStatusFn: func(
			context.Context,
			uint,
			uint,
			string,
			string,
		) error {
			return expectedErr
		},
	}

	service := NewService(repo)

	err := service.UpdateReportStatus(
		context.Background(),
		12,
		9,
		UpdateReportStatusRequest{
			Status: "resolved",
		},
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected %v, got %v",
			expectedErr,
			err,
		)
	}
}

func TestIsAllowedReportStatus(t *testing.T) {
	tests := []struct {
		status   string
		expected bool
	}{
		{"open", true},
		{"resolved", true},
		{"dismissed", true},
		{"pending", false},
		{"closed", false},
		{"", false},
	}

	for _, test := range tests {
		t.Run(
			test.status,
			func(t *testing.T) {
				result := isAllowedReportStatus(
					test.status,
				)

				if result != test.expected {
					t.Fatalf(
						"expected %v for %q, got %v",
						test.expected,
						test.status,
						result,
					)
				}
			},
		)
	}
}
