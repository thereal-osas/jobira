package jobinvitations

import (
	"context"
	"errors"
	"testing"
	"time"
)

type mockRepository struct {
	createFn                  func(context.Context, *JobInvitation) error
	existsFn                  func(context.Context, uint, uint) (bool, error)
	getJobClientIDFn          func(context.Context, uint) (uint, error)
	listSentByClientIDFn      func(context.Context, uint) ([]JobInvitation, error)
	listReceivedByCleanerIDFn func(context.Context, uint) ([]JobInvitation, error)
}

func (m *mockRepository) Create(
	ctx context.Context,
	invitation *JobInvitation,
) error {
	if m.createFn != nil {
		return m.createFn(ctx, invitation)
	}

	return nil
}

func (m *mockRepository) Exists(
	ctx context.Context,
	jobID uint,
	cleanerID uint,
) (bool, error) {
	if m.existsFn != nil {
		return m.existsFn(ctx, jobID, cleanerID)
	}

	return false, nil
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

func (m *mockRepository) ListSentByClientID(
	ctx context.Context,
	clientID uint,
) ([]JobInvitation, error) {
	if m.listSentByClientIDFn != nil {
		return m.listSentByClientIDFn(ctx, clientID)
	}

	return nil, nil
}

func (m *mockRepository) ListReceivedByCleanerID(
	ctx context.Context,
	cleanerID uint,
) ([]JobInvitation, error) {
	if m.listReceivedByCleanerIDFn != nil {
		return m.listReceivedByCleanerIDFn(
			ctx,
			cleanerID,
		)
	}

	return nil, nil
}

type mockBlockChecker struct {
	isBlockedFn func(
		context.Context,
		uint,
		uint,
	) (bool, error)
}

func (m *mockBlockChecker) IsBlocked(
	ctx context.Context,
	clientID uint,
	cleanerID uint,
) (bool, error) {
	if m.isBlockedFn != nil {
		return m.isBlockedFn(
			ctx,
			clientID,
			cleanerID,
		)
	}

	return false, nil
}

func TestNewService(t *testing.T) {
	repo := &mockRepository{}
	blockChecker := &mockBlockChecker{}

	service := NewService(
		repo,
		nil,
		blockChecker,
	)

	if service == nil {
		t.Fatal("expected service")
	}

	if service.repo != repo {
		t.Fatal("expected repository assigned")
	}

	if service.notificationsService != nil {
		t.Fatal("expected nil notifications service")
	}

	if service.blockChecker != blockChecker {
		t.Fatal("expected block checker assigned")
	}
}

func TestService_Create_Success(t *testing.T) {
	ownershipChecked := false
	existsChecked := false
	blockChecked := false
	createCalled := false

	repo := &mockRepository{
		getJobClientIDFn: func(
			ctx context.Context,
			jobID uint,
		) (uint, error) {
			ownershipChecked = true

			if jobID != 7 {
				t.Fatalf(
					"expected job ID 7, got %d",
					jobID,
				)
			}

			return 5, nil
		},
		existsFn: func(
			ctx context.Context,
			jobID uint,
			cleanerID uint,
		) (bool, error) {
			existsChecked = true

			if jobID != 7 {
				t.Fatalf(
					"expected job ID 7, got %d",
					jobID,
				)
			}

			if cleanerID != 8 {
				t.Fatalf(
					"expected cleaner ID 8, got %d",
					cleanerID,
				)
			}

			return false, nil
		},
		createFn: func(
			ctx context.Context,
			invitation *JobInvitation,
		) error {
			createCalled = true

			if invitation.JobID != 7 {
				t.Fatalf(
					"expected job ID 7, got %d",
					invitation.JobID,
				)
			}

			if invitation.ClientID != 5 {
				t.Fatalf(
					"expected client ID 5, got %d",
					invitation.ClientID,
				)
			}

			if invitation.CleanerID != 8 {
				t.Fatalf(
					"expected cleaner ID 8, got %d",
					invitation.CleanerID,
				)
			}

			if invitation.Status != "sent" {
				t.Fatalf(
					"expected status sent, got %q",
					invitation.Status,
				)
			}

			invitation.ID = 12
			invitation.CreatedAt = time.Now()
			invitation.UpdatedAt = invitation.CreatedAt

			return nil
		},
	}

	blockChecker := &mockBlockChecker{
		isBlockedFn: func(
			context.Context,
			uint,
			uint,
		) (bool, error) {
			blockChecked = true
			return false, nil
		},
	}

	service := NewService(
		repo,
		nil,
		blockChecker,
	)

	invitation, err := service.Create(
		context.Background(),
		7,
		5,
		CreateJobInvitationRequest{
			CleanerID: 8,
			Message:   "  Please apply for this job.  ",
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if invitation == nil {
		t.Fatal("expected invitation")
	}

	if invitation.ID != 12 {
		t.Fatalf(
			"expected ID 12, got %d",
			invitation.ID,
		)
	}

	if invitation.Message != "Please apply for this job." {
		t.Fatalf(
			"expected trimmed message, got %q",
			invitation.Message,
		)
	}

	if !ownershipChecked {
		t.Fatal("expected ownership check")
	}

	if !existsChecked {
		t.Fatal("expected duplicate check")
	}

	if !blockChecked {
		t.Fatal("expected block check")
	}

	if !createCalled {
		t.Fatal("expected create call")
	}
}

func TestService_Create_WithoutBlockChecker(t *testing.T) {
	repo := &mockRepository{
		getJobClientIDFn: func(
			context.Context,
			uint,
		) (uint, error) {
			return 5, nil
		},
		existsFn: func(
			context.Context,
			uint,
			uint,
		) (bool, error) {
			return false, nil
		},
		createFn: func(
			context.Context,
			*JobInvitation,
		) error {
			return nil
		},
	}

	service := NewService(repo, nil, nil)

	invitation, err := service.Create(
		context.Background(),
		7,
		5,
		CreateJobInvitationRequest{
			CleanerID: 8,
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if invitation == nil {
		t.Fatal("expected invitation")
	}
}

func TestService_Create_InvalidInput(t *testing.T) {
	service := NewService(
		&mockRepository{},
		nil,
		nil,
	)

	tests := []struct {
		name      string
		jobID     uint
		clientID  uint
		cleanerID uint
	}{
		{
			name:      "zero job ID",
			jobID:     0,
			clientID:  5,
			cleanerID: 8,
		},
		{
			name:      "zero client ID",
			jobID:     7,
			clientID:  0,
			cleanerID: 8,
		},
		{
			name:      "zero cleaner ID",
			jobID:     7,
			clientID:  5,
			cleanerID: 0,
		},
		{
			name:      "same client and cleaner",
			jobID:     7,
			clientID:  5,
			cleanerID: 5,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			invitation, err := service.Create(
				context.Background(),
				test.jobID,
				test.clientID,
				CreateJobInvitationRequest{
					CleanerID: test.cleanerID,
				},
			)

			if invitation != nil {
				t.Fatalf(
					"expected nil invitation, got %+v",
					invitation,
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

func TestService_Create_ExistsError(t *testing.T) {
	expectedErr := errors.New("exists check failed")

	repo := &mockRepository{
		getJobClientIDFn: func(
			context.Context,
			uint,
		) (uint, error) {
			return 5, nil
		},
		existsFn: func(
			context.Context,
			uint,
			uint,
		) (bool, error) {
			return false, expectedErr
		},
	}

	service := NewService(repo, nil, nil)

	invitation, err := service.Create(
		context.Background(),
		7,
		5,
		CreateJobInvitationRequest{
			CleanerID: 8,
		},
	)

	if invitation != nil {
		t.Fatalf(
			"expected nil invitation, got %+v",
			invitation,
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

func TestService_Create_InvitationExists(t *testing.T) {
	repo := &mockRepository{
		getJobClientIDFn: func(
			context.Context,
			uint,
		) (uint, error) {
			return 5, nil
		},
		existsFn: func(
			context.Context,
			uint,
			uint,
		) (bool, error) {
			return true, nil
		},
	}

	service := NewService(repo, nil, nil)

	invitation, err := service.Create(
		context.Background(),
		7,
		5,
		CreateJobInvitationRequest{
			CleanerID: 8,
		},
	)

	if invitation != nil {
		t.Fatalf(
			"expected nil invitation, got %+v",
			invitation,
		)
	}

	if !errors.Is(err, ErrInvitationExists) {
		t.Fatalf(
			"expected ErrInvitationExists, got %v",
			err,
		)
	}
}

func TestService_Create_BlockCheckerError(t *testing.T) {
	expectedErr := errors.New("block check failed")

	repo := &mockRepository{
		getJobClientIDFn: func(
			context.Context,
			uint,
		) (uint, error) {
			return 5, nil
		},
		existsFn: func(
			context.Context,
			uint,
			uint,
		) (bool, error) {
			return false, nil
		},
	}

	blockChecker := &mockBlockChecker{
		isBlockedFn: func(
			context.Context,
			uint,
			uint,
		) (bool, error) {
			return false, expectedErr
		},
	}

	service := NewService(
		repo,
		nil,
		blockChecker,
	)

	invitation, err := service.Create(
		context.Background(),
		7,
		5,
		CreateJobInvitationRequest{
			CleanerID: 8,
		},
	)

	if invitation != nil {
		t.Fatalf(
			"expected nil invitation, got %+v",
			invitation,
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

func TestService_Create_Blocked(t *testing.T) {
	repo := &mockRepository{
		getJobClientIDFn: func(
			context.Context,
			uint,
		) (uint, error) {
			return 5, nil
		},
		existsFn: func(
			context.Context,
			uint,
			uint,
		) (bool, error) {
			return false, nil
		},
	}

	blockChecker := &mockBlockChecker{
		isBlockedFn: func(
			context.Context,
			uint,
			uint,
		) (bool, error) {
			return true, nil
		},
	}

	service := NewService(
		repo,
		nil,
		blockChecker,
	)

	invitation, err := service.Create(
		context.Background(),
		7,
		5,
		CreateJobInvitationRequest{
			CleanerID: 8,
		},
	)

	if invitation != nil {
		t.Fatalf(
			"expected nil invitation, got %+v",
			invitation,
		)
	}

	if !errors.Is(err, ErrForbidden) {
		t.Fatalf(
			"expected ErrForbidden, got %v",
			err,
		)
	}
}

func TestService_Create_RepositoryError(t *testing.T) {
	expectedErr := errors.New("create failed")

	repo := &mockRepository{
		getJobClientIDFn: func(
			context.Context,
			uint,
		) (uint, error) {
			return 5, nil
		},
		existsFn: func(
			context.Context,
			uint,
			uint,
		) (bool, error) {
			return false, nil
		},
		createFn: func(
			context.Context,
			*JobInvitation,
		) error {
			return expectedErr
		},
	}

	service := NewService(repo, nil, nil)

	invitation, err := service.Create(
		context.Background(),
		7,
		5,
		CreateJobInvitationRequest{
			CleanerID: 8,
		},
	)

	if invitation != nil {
		t.Fatalf(
			"expected nil invitation, got %+v",
			invitation,
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

func TestService_ListSent_Success(t *testing.T) {
	repo := &mockRepository{
		listSentByClientIDFn: func(
			ctx context.Context,
			clientID uint,
		) ([]JobInvitation, error) {
			if clientID != 5 {
				t.Fatalf(
					"expected client ID 5, got %d",
					clientID,
				)
			}

			return []JobInvitation{
				{
					ID:        1,
					JobID:     7,
					ClientID:  5,
					CleanerID: 8,
				},
			}, nil
		},
	}

	service := NewService(repo, nil, nil)

	invitations, err := service.ListSent(
		context.Background(),
		5,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(invitations) != 1 {
		t.Fatalf(
			"expected 1 invitation, got %d",
			len(invitations),
		)
	}
}

func TestService_ListSent_InvalidInput(t *testing.T) {
	service := NewService(
		&mockRepository{},
		nil,
		nil,
	)

	invitations, err := service.ListSent(
		context.Background(),
		0,
	)

	if invitations != nil {
		t.Fatalf(
			"expected nil invitations, got %+v",
			invitations,
		)
	}

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf(
			"expected ErrInvalidInput, got %v",
			err,
		)
	}
}

func TestService_ListSent_RepositoryError(t *testing.T) {
	expectedErr := errors.New("list sent failed")

	repo := &mockRepository{
		listSentByClientIDFn: func(
			context.Context,
			uint,
		) ([]JobInvitation, error) {
			return nil, expectedErr
		},
	}

	service := NewService(repo, nil, nil)

	invitations, err := service.ListSent(
		context.Background(),
		5,
	)

	if invitations != nil {
		t.Fatalf(
			"expected nil invitations, got %+v",
			invitations,
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

func TestService_ListReceived_Success(t *testing.T) {
	repo := &mockRepository{
		listReceivedByCleanerIDFn: func(
			ctx context.Context,
			cleanerID uint,
		) ([]JobInvitation, error) {
			if cleanerID != 8 {
				t.Fatalf(
					"expected cleaner ID 8, got %d",
					cleanerID,
				)
			}

			return []JobInvitation{
				{
					ID:        1,
					JobID:     7,
					ClientID:  5,
					CleanerID: 8,
				},
			}, nil
		},
	}

	service := NewService(repo, nil, nil)

	invitations, err := service.ListReceived(
		context.Background(),
		8,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(invitations) != 1 {
		t.Fatalf(
			"expected 1 invitation, got %d",
			len(invitations),
		)
	}
}

func TestService_ListReceived_InvalidInput(t *testing.T) {
	service := NewService(
		&mockRepository{},
		nil,
		nil,
	)

	invitations, err := service.ListReceived(
		context.Background(),
		0,
	)

	if invitations != nil {
		t.Fatalf(
			"expected nil invitations, got %+v",
			invitations,
		)
	}

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf(
			"expected ErrInvalidInput, got %v",
			err,
		)
	}
}

func TestService_ListReceived_RepositoryError(t *testing.T) {
	expectedErr := errors.New("list received failed")

	repo := &mockRepository{
		listReceivedByCleanerIDFn: func(
			context.Context,
			uint,
		) ([]JobInvitation, error) {
			return nil, expectedErr
		},
	}

	service := NewService(repo, nil, nil)

	invitations, err := service.ListReceived(
		context.Background(),
		8,
	)

	if invitations != nil {
		t.Fatalf(
			"expected nil invitations, got %+v",
			invitations,
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

func TestService_Create_GetJobClientIDError(t *testing.T) {
	expectedErr := errors.New("job lookup failed")

	repo := &mockRepository{
		getJobClientIDFn: func(
			context.Context,
			uint,
		) (uint, error) {
			return 0, expectedErr
		},
	}

	service := NewService(repo, nil, nil)

	invitation, err := service.Create(
		context.Background(),
		7,
		5,
		CreateJobInvitationRequest{
			CleanerID: 8,
		},
	)

	if invitation != nil {
		t.Fatalf(
			"expected nil invitation, got %+v",
			invitation,
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

func TestService_Create_ForbiddenOwner(t *testing.T) {
	repo := &mockRepository{
		getJobClientIDFn: func(
			context.Context,
			uint,
		) (uint, error) {
			return 99, nil
		},
	}

	service := NewService(repo, nil, nil)

	invitation, err := service.Create(
		context.Background(),
		7,
		5,
		CreateJobInvitationRequest{
			CleanerID: 8,
		},
	)

	if invitation != nil {
		t.Fatalf(
			"expected nil invitation, got %+v",
			invitation,
		)
	}

	if !errors.Is(err, ErrForbidden) {
		t.Fatalf(
			"expected ErrForbidden, got %v",
			err,
		)
	}
}
