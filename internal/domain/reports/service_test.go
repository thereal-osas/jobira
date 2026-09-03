package reports

import (
	"context"
	"errors"
	"testing"
)

type mockRepository struct {
	createFn           func(context.Context, *Report) error
	getByIDFn          func(context.Context, uint) (*Report, error)
	listByReporterIDFn func(context.Context, uint) ([]Report, error)
	listAllFn          func(context.Context) ([]Report, error)
	listOpenFn         func(context.Context) ([]Report, error)
	reviewFn           func(context.Context, uint, string, string, uint) error
}

func (m *mockRepository) Create(
	ctx context.Context,
	report *Report,
) error {
	if m.createFn != nil {
		return m.createFn(ctx, report)
	}
	return nil
}

func (m *mockRepository) GetByID(
	ctx context.Context,
	id uint,
) (*Report, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return &Report{}, nil
}

func (m *mockRepository) ListByReporterID(
	ctx context.Context,
	reporterID uint,
) ([]Report, error) {
	if m.listByReporterIDFn != nil {
		return m.listByReporterIDFn(ctx, reporterID)
	}
	return nil, nil
}

func (m *mockRepository) ListAll(
	ctx context.Context,
) ([]Report, error) {
	if m.listAllFn != nil {
		return m.listAllFn(ctx)
	}
	return nil, nil
}

func (m *mockRepository) ListOpen(
	ctx context.Context,
) ([]Report, error) {
	if m.listOpenFn != nil {
		return m.listOpenFn(ctx)
	}
	return nil, nil
}

func (m *mockRepository) Review(
	ctx context.Context,
	reportID uint,
	status string,
	adminNotes string,
	adminID uint,
) error {
	if m.reviewFn != nil {
		return m.reviewFn(
			ctx,
			reportID,
			status,
			adminNotes,
			adminID,
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

func TestService_Create_Success(t *testing.T) {
	var received *Report

	repo := &mockRepository{
		createFn: func(
			ctx context.Context,
			report *Report,
		) error {
			received = report
			report.ID = 10
			return nil
		},
	}

	service := NewService(repo)

	reportedUserID := uint(20)
	jobID := uint(30)

	report, err := service.Create(
		context.Background(),
		5,
		CreateReportRequest{
			ReportedUserID: &reportedUserID,
			JobID:          &jobID,
			ReportType:     "  user  ",
			Reason:         "  abusive behaviour  ",
			Details:        "  additional details  ",
		},
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if report.ID != 10 {
		t.Fatalf("expected id 10, got %d", report.ID)
	}

	if received.ReporterID != 5 {
		t.Fatalf(
			"expected reporter id 5, got %d",
			received.ReporterID,
		)
	}

	if received.ReportType != "user" {
		t.Fatalf(
			"expected trimmed report type, got %q",
			received.ReportType,
		)
	}

	if received.Reason != "abusive behaviour" {
		t.Fatalf(
			"expected trimmed reason, got %q",
			received.Reason,
		)
	}

	if received.Details != "additional details" {
		t.Fatalf(
			"expected trimmed details, got %q",
			received.Details,
		)
	}

	if received.Status != "open" {
		t.Fatalf(
			"expected open status, got %q",
			received.Status,
		)
	}
}

func TestService_Create_InvalidReporter(t *testing.T) {
	service := NewService(&mockRepository{})

	_, err := service.Create(
		context.Background(),
		0,
		CreateReportRequest{
			ReportType: "user",
			Reason:     "reason",
		},
	)

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf(
			"expected ErrInvalidInput, got %v",
			err,
		)
	}
}

func TestService_Create_EmptyReportType(t *testing.T) {
	service := NewService(&mockRepository{})

	_, err := service.Create(
		context.Background(),
		1,
		CreateReportRequest{
			ReportType: "   ",
			Reason:     "reason",
		},
	)

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf(
			"expected ErrInvalidInput, got %v",
			err,
		)
	}
}

func TestService_Create_EmptyReason(t *testing.T) {
	service := NewService(&mockRepository{})

	_, err := service.Create(
		context.Background(),
		1,
		CreateReportRequest{
			ReportType: "user",
			Reason:     "   ",
		},
	)

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf(
			"expected ErrInvalidInput, got %v",
			err,
		)
	}
}

func TestService_Create_InvalidReportType(t *testing.T) {
	service := NewService(&mockRepository{})

	_, err := service.Create(
		context.Background(),
		1,
		CreateReportRequest{
			ReportType: "something_invalid",
			Reason:     "reason",
		},
	)

	if !errors.Is(err, ErrInvalidReportType) {
		t.Fatalf(
			"expected ErrInvalidReportType, got %v",
			err,
		)
	}
}

func TestService_Create_RepositoryError(t *testing.T) {
	expectedErr := errors.New("database error")

	repo := &mockRepository{
		createFn: func(
			context.Context,
			*Report,
		) error {
			return expectedErr
		},
	}

	service := NewService(repo)

	_, err := service.Create(
		context.Background(),
		1,
		CreateReportRequest{
			ReportType: "user",
			Reason:     "reason",
		},
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected repository error, got %v",
			err,
		)
	}
}

func TestService_GetByID_Owner(t *testing.T) {
	repo := &mockRepository{
		getByIDFn: func(
			context.Context,
			uint,
		) (*Report, error) {
			return &Report{
				ID:         10,
				ReporterID: 5,
			}, nil
		},
	}

	service := NewService(repo)

	report, err := service.GetByID(
		context.Background(),
		10,
		5,
		"client",
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if report.ID != 10 {
		t.Fatalf("expected report 10, got %d", report.ID)
	}
}

func TestService_GetByID_Admin(t *testing.T) {
	repo := &mockRepository{
		getByIDFn: func(
			context.Context,
			uint,
		) (*Report, error) {
			return &Report{
				ID:         10,
				ReporterID: 5,
			}, nil
		},
	}

	service := NewService(repo)

	_, err := service.GetByID(
		context.Background(),
		10,
		99,
		"admin",
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestService_GetByID_InvalidInput(t *testing.T) {
	service := NewService(&mockRepository{})

	_, err := service.GetByID(
		context.Background(),
		0,
		1,
		"client",
	)

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf(
			"expected ErrInvalidInput, got %v",
			err,
		)
	}
}

func TestService_GetByID_RepositoryError(t *testing.T) {
	expectedErr := errors.New("database error")

	repo := &mockRepository{
		getByIDFn: func(
			context.Context,
			uint,
		) (*Report, error) {
			return nil, expectedErr
		},
	}

	service := NewService(repo)

	_, err := service.GetByID(
		context.Background(),
		10,
		5,
		"client",
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected repository error, got %v",
			err,
		)
	}
}

func TestService_GetByID_Forbidden(t *testing.T) {
	repo := &mockRepository{
		getByIDFn: func(
			context.Context,
			uint,
		) (*Report, error) {
			return &Report{
				ID:         10,
				ReporterID: 5,
			}, nil
		},
	}

	service := NewService(repo)

	_, err := service.GetByID(
		context.Background(),
		10,
		6,
		"client",
	)

	if !errors.Is(err, ErrForbidden) {
		t.Fatalf(
			"expected ErrForbidden, got %v",
			err,
		)
	}
}

func TestService_ListMine_Success(t *testing.T) {
	repo := &mockRepository{
		listByReporterIDFn: func(
			ctx context.Context,
			reporterID uint,
		) ([]Report, error) {
			if reporterID != 5 {
				t.Fatalf(
					"expected reporter 5, got %d",
					reporterID,
				)
			}

			return []Report{
				{ID: 1},
				{ID: 2},
			}, nil
		},
	}

	service := NewService(repo)

	reports, err := service.ListMine(
		context.Background(),
		5,
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(reports) != 2 {
		t.Fatalf(
			"expected 2 reports, got %d",
			len(reports),
		)
	}
}

func TestService_ListMine_InvalidInput(t *testing.T) {
	service := NewService(&mockRepository{})

	_, err := service.ListMine(
		context.Background(),
		0,
	)

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf(
			"expected ErrInvalidInput, got %v",
			err,
		)
	}
}

func TestService_ListMine_RepositoryError(t *testing.T) {
	expectedErr := errors.New("database error")

	repo := &mockRepository{
		listByReporterIDFn: func(
			context.Context,
			uint,
		) ([]Report, error) {
			return nil, expectedErr
		},
	}

	service := NewService(repo)

	_, err := service.ListMine(
		context.Background(),
		5,
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected repository error, got %v",
			err,
		)
	}
}

func TestService_ListAll_Success(t *testing.T) {
	repo := &mockRepository{
		listAllFn: func(
			context.Context,
		) ([]Report, error) {
			return []Report{
				{ID: 1},
				{ID: 2},
			}, nil
		},
	}

	service := NewService(repo)

	reports, err := service.ListAll(
		context.Background(),
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(reports) != 2 {
		t.Fatalf(
			"expected 2 reports, got %d",
			len(reports),
		)
	}
}

func TestService_ListAll_Error(t *testing.T) {
	expectedErr := errors.New("database error")

	repo := &mockRepository{
		listAllFn: func(
			context.Context,
		) ([]Report, error) {
			return nil, expectedErr
		},
	}

	service := NewService(repo)

	_, err := service.ListAll(
		context.Background(),
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected repository error, got %v",
			err,
		)
	}
}

func TestService_ListOpen_Success(t *testing.T) {
	repo := &mockRepository{
		listOpenFn: func(
			context.Context,
		) ([]Report, error) {
			return []Report{
				{
					ID:     1,
					Status: "open",
				},
			}, nil
		},
	}

	service := NewService(repo)

	reports, err := service.ListOpen(
		context.Background(),
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(reports) != 1 {
		t.Fatalf(
			"expected 1 report, got %d",
			len(reports),
		)
	}
}

func TestService_ListOpen_Error(t *testing.T) {
	expectedErr := errors.New("database error")

	repo := &mockRepository{
		listOpenFn: func(
			context.Context,
		) ([]Report, error) {
			return nil, expectedErr
		},
	}

	service := NewService(repo)

	_, err := service.ListOpen(
		context.Background(),
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected repository error, got %v",
			err,
		)
	}
}

func TestService_Review_Success(t *testing.T) {
	reviewCalled := false

	repo := &mockRepository{
		reviewFn: func(
			ctx context.Context,
			reportID uint,
			status string,
			adminNotes string,
			adminID uint,
		) error {
			reviewCalled = true

			if reportID != 10 {
				t.Fatalf(
					"expected report 10, got %d",
					reportID,
				)
			}

			if status != "resolved" {
				t.Fatalf(
					"expected resolved, got %q",
					status,
				)
			}

			if adminNotes != "review complete" {
				t.Fatalf(
					"unexpected admin notes %q",
					adminNotes,
				)
			}

			if adminID != 99 {
				t.Fatalf(
					"expected admin 99, got %d",
					adminID,
				)
			}

			return nil
		},
		getByIDFn: func(
			context.Context,
			uint,
		) (*Report, error) {
			return &Report{
				ID:         10,
				Status:     "resolved",
				AdminNotes: "review complete",
			}, nil
		},
	}

	service := NewService(repo)

	report, err := service.Review(
		context.Background(),
		10,
		99,
		ReviewReportRequest{
			Status:     "  resolved  ",
			AdminNotes: "  review complete  ",
		},
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !reviewCalled {
		t.Fatal("expected repository Review called")
	}

	if report.Status != "resolved" {
		t.Fatalf(
			"expected resolved, got %q",
			report.Status,
		)
	}
}

func TestService_Review_InvalidInput(t *testing.T) {
	service := NewService(&mockRepository{})

	_, err := service.Review(
		context.Background(),
		0,
		99,
		ReviewReportRequest{
			Status: "resolved",
		},
	)

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf(
			"expected ErrInvalidInput, got %v",
			err,
		)
	}
}

func TestService_Review_InvalidStatus(t *testing.T) {
	service := NewService(&mockRepository{})

	_, err := service.Review(
		context.Background(),
		10,
		99,
		ReviewReportRequest{
			Status: "invalid",
		},
	)

	if !errors.Is(err, ErrInvalidReportStatus) {
		t.Fatalf(
			"expected ErrInvalidReportStatus, got %v",
			err,
		)
	}
}

func TestService_Review_RepositoryError(t *testing.T) {
	expectedErr := errors.New("database error")

	repo := &mockRepository{
		reviewFn: func(
			context.Context,
			uint,
			string,
			string,
			uint,
		) error {
			return expectedErr
		},
	}

	service := NewService(repo)

	_, err := service.Review(
		context.Background(),
		10,
		99,
		ReviewReportRequest{
			Status: "resolved",
		},
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected repository error, got %v",
			err,
		)
	}
}

func TestService_Review_GetByIDError(t *testing.T) {
	expectedErr := errors.New("reload error")

	repo := &mockRepository{
		getByIDFn: func(
			context.Context,
			uint,
		) (*Report, error) {
			return nil, expectedErr
		},
	}

	service := NewService(repo)

	_, err := service.Review(
		context.Background(),
		10,
		99,
		ReviewReportRequest{
			Status: "resolved",
		},
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected reload error, got %v",
			err,
		)
	}
}

func TestAllowedReportTypes(t *testing.T) {
	allowed := []string{
		"user",
		"job",
		"booking",
		"message",
		"abuse",
		"no_show",
		"spam",
	}

	for _, reportType := range allowed {
		if !isAllowedReportType(reportType) {
			t.Fatalf(
				"expected %q allowed",
				reportType,
			)
		}
	}

	if isAllowedReportType("invalid") {
		t.Fatal("expected invalid type rejected")
	}
}

func TestAllowedReportStatuses(t *testing.T) {
	allowed := []string{
		"open",
		"reviewing",
		"resolved",
		"dismissed",
	}

	for _, status := range allowed {
		if !isAllowedReportStatus(status) {
			t.Fatalf(
				"expected %q allowed",
				status,
			)
		}
	}

	if isAllowedReportStatus("invalid") {
		t.Fatal("expected invalid status rejected")
	}
}
