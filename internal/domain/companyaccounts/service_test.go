package companyaccounts

import (
	"context"
	"errors"
	"testing"
	"time"
)

type mockRepository struct {
	createCompanyFn func(
		context.Context,
		*Company,
	) error

	addMemberFn func(
		context.Context,
		uint,
		uint,
		string,
	) error

	getByIDFn func(
		context.Context,
		uint,
	) (*Company, error)

	getMemberFn func(
		context.Context,
		uint,
		uint,
	) (*CompanyMember, error)

	removeMemberFn func(
		context.Context,
		uint,
		uint,
	) error

	listMembersFn func(
		context.Context,
		uint,
	) ([]CompanyMember, error)

	countActiveCleanerSeatsFn func(
		context.Context,
		uint,
	) (int, error)

	getCleanerSeatLimitFn func(
		context.Context,
		uint,
	) (int, error)

	updateMemberStatusFn func(
		context.Context,
		uint,
		uint,
		string,
	) error

	updateMemberRoleFn func(
		context.Context,
		uint,
		uint,
		string,
	) error

	countMembersByRolesAndStatusFn func(
		context.Context,
		uint,
		string,
		string,
	) (int, error)
}

func (m *mockRepository) CreateCompany(
	ctx context.Context,
	company *Company,
) error {
	if m.createCompanyFn != nil {
		return m.createCompanyFn(
			ctx,
			company,
		)
	}

	return nil
}

func (m *mockRepository) AddMember(
	ctx context.Context,
	companyID uint,
	userID uint,
	role string,
) error {
	if m.addMemberFn != nil {
		return m.addMemberFn(
			ctx,
			companyID,
			userID,
			role,
		)
	}

	return nil
}

func (m *mockRepository) GetByID(
	ctx context.Context,
	companyID uint,
) (*Company, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(
			ctx,
			companyID,
		)
	}

	return nil,
		ErrCompanyNotFound
}

func (m *mockRepository) GetMember(
	ctx context.Context,
	companyID uint,
	userID uint,
) (*CompanyMember, error) {
	if m.getMemberFn != nil {
		return m.getMemberFn(
			ctx,
			companyID,
			userID,
		)
	}

	return nil,
		ErrMemberNotFound
}

func (m *mockRepository) RemoveMember(
	ctx context.Context,
	companyID uint,
	userID uint,
) error {
	if m.removeMemberFn != nil {
		return m.removeMemberFn(
			ctx,
			companyID,
			userID,
		)
	}

	return nil
}

func (m *mockRepository) ListMembers(
	ctx context.Context,
	companyID uint,
) ([]CompanyMember, error) {
	if m.listMembersFn != nil {
		return m.listMembersFn(
			ctx,
			companyID,
		)
	}

	return []CompanyMember{},
		nil
}

func (m *mockRepository) CountActiveCleanerSeats(
	ctx context.Context,
	companyID uint,
) (int, error) {
	if m.countActiveCleanerSeatsFn != nil {
		return m.countActiveCleanerSeatsFn(
			ctx,
			companyID,
		)
	}

	return 0, nil
}

func (m *mockRepository) GetCleanerSeatLimit(
	ctx context.Context,
	companyID uint,
) (int, error) {
	if m.getCleanerSeatLimitFn != nil {
		return m.getCleanerSeatLimitFn(
			ctx,
			companyID,
		)
	}

	return 3, nil
}

func (m *mockRepository) UpdateMemberStatus(
	ctx context.Context,
	companyID uint,
	userID uint,
	status string,
) error {
	if m.updateMemberStatusFn != nil {
		return m.updateMemberStatusFn(
			ctx,
			companyID,
			userID,
			status,
		)
	}

	return nil
}

func (m *mockRepository) UpdateMemberRole(
	ctx context.Context,
	companyID uint,
	userID uint,
	role string,
) error {
	if m.updateMemberRoleFn != nil {
		return m.updateMemberRoleFn(
			ctx,
			companyID,
			userID,
			role,
		)
	}

	return nil
}

func (m *mockRepository) CountMembersByRolesAndStatus(
	ctx context.Context,
	companyID uint,
	role string,
	status string,
) (int, error) {
	if m.countMembersByRolesAndStatusFn != nil {
		return m.countMembersByRolesAndStatusFn(
			ctx,
			companyID,
			role,
			status,
		)
	}

	return 0, nil
}

func TestService_CreateCompany_Success(
	t *testing.T,
) {
	ownerAdded := false

	repo := &mockRepository{
		createCompanyFn: func(
			ctx context.Context,
			company *Company,
		) error {
			company.ID = 10
			company.CreatedAt =
				time.Now().UTC()
			company.UpdatedAt =
				company.CreatedAt

			return nil
		},

		addMemberFn: func(
			ctx context.Context,
			companyID uint,
			userID uint,
			role string,
		) error {
			if companyID != 10 {
				t.Fatalf(
					"expected company ID 10, got %d",
					companyID,
				)
			}

			if userID != 5 {
				t.Fatalf(
					"expected owner ID 5, got %d",
					userID,
				)
			}

			if role != "owner" {
				t.Fatalf(
					"expected owner role, got %q",
					role,
				)
			}

			ownerAdded = true

			return nil
		},
	}

	service := NewService(repo)

	result, err :=
		service.CreateCompany(
			context.Background(),
			5,
			CreateCompanyRequest{
				Name: " Wembi Cleaning Ltd ",

				Description: " East London cleaning company ",
			},
		)

	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}

	if result == nil {
		t.Fatal(
			"expected company",
		)
	}

	if result.ID != 10 {
		t.Fatalf(
			"expected company ID 10, got %d",
			result.ID,
		)
	}

	if result.Name !=
		"Wembi Cleaning Ltd" {
		t.Fatalf(
			"unexpected company name %q",
			result.Name,
		)
	}

	if !ownerAdded {
		t.Fatal(
			"expected owner membership to be created",
		)
	}
}

func TestService_AddMember_CleanerSuccess(
	t *testing.T,
) {
	repo := &mockRepository{
		getMemberFn: func(
			ctx context.Context,
			companyID uint,
			userID uint,
		) (*CompanyMember, error) {
			return &CompanyMember{
				CompanyID: companyID,

				UserID: userID,

				Role: "owner",

				Status: "active",
			}, nil
		},

		getCleanerSeatLimitFn: func(
			ctx context.Context,
			companyID uint,
		) (int, error) {
			return 3, nil
		},

		countActiveCleanerSeatsFn: func(
			ctx context.Context,
			companyID uint,
		) (int, error) {
			return 2, nil
		},

		addMemberFn: func(
			ctx context.Context,
			companyID uint,
			userID uint,
			role string,
		) error {
			if userID != 8 {
				t.Fatalf(
					"expected user 8, got %d",
					userID,
				)
			}

			if role != "cleaner" {
				t.Fatalf(
					"expected cleaner role, got %q",
					role,
				)
			}

			return nil
		},
	}

	service := NewService(repo)

	err :=
		service.AddMember(
			context.Background(),
			10,
			5,
			AddMemberRequest{
				UserID: 8,
				Role:   "cleaner",
			},
		)

	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}
}

func TestService_AddMember_SeatLimitReached(
	t *testing.T,
) {
	repo := &mockRepository{
		getMemberFn: func(
			ctx context.Context,
			companyID uint,
			userID uint,
		) (*CompanyMember, error) {
			return &CompanyMember{
				CompanyID: companyID,

				UserID: userID,

				Role: "owner",

				Status: "active",
			}, nil
		},

		getCleanerSeatLimitFn: func(
			ctx context.Context,
			companyID uint,
		) (int, error) {
			return 3, nil
		},

		countActiveCleanerSeatsFn: func(
			ctx context.Context,
			companyID uint,
		) (int, error) {
			return 3, nil
		},
	}

	service := NewService(repo)

	err :=
		service.AddMember(
			context.Background(),
			10,
			5,
			AddMemberRequest{
				UserID: 8,
				Role:   "cleaner",
			},
		)

	if !errors.Is(
		err,
		ErrSeatLimitReached,
	) {
		t.Fatalf(
			"expected ErrSeatLimitReached, got %v",
			err,
		)
	}
}

func TestService_AddMember_AdminDoesNotUseCleanerSeat(
	t *testing.T,
) {
	seatCheckCalled := false

	repo := &mockRepository{
		getMemberFn: func(
			ctx context.Context,
			companyID uint,
			userID uint,
		) (*CompanyMember, error) {
			return &CompanyMember{
				CompanyID: companyID,

				UserID: userID,

				Role: "owner",

				Status: "active",
			}, nil
		},

		getCleanerSeatLimitFn: func(
			ctx context.Context,
			companyID uint,
		) (int, error) {
			seatCheckCalled = true

			return 3, nil
		},

		addMemberFn: func(
			ctx context.Context,
			companyID uint,
			userID uint,
			role string,
		) error {
			if role != "admin" {
				t.Fatalf(
					"expected admin, got %q",
					role,
				)
			}

			return nil
		},
	}

	service := NewService(repo)

	err :=
		service.AddMember(
			context.Background(),
			10,
			5,
			AddMemberRequest{
				UserID: 9,
				Role:   "admin",
			},
		)

	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}

	if seatCheckCalled {
		t.Fatal(
			"admin should not consume cleaner seat",
		)
	}
}

func TestService_AddMember_ForbiddenCleanerManager(
	t *testing.T,
) {
	repo := &mockRepository{
		getMemberFn: func(
			ctx context.Context,
			companyID uint,
			userID uint,
		) (*CompanyMember, error) {
			return &CompanyMember{
				CompanyID: companyID,

				UserID: userID,

				Role: "cleaner",

				Status: "active",
			}, nil
		},
	}

	service := NewService(repo)

	err :=
		service.AddMember(
			context.Background(),
			10,
			8,
			AddMemberRequest{
				UserID: 9,
				Role:   "cleaner",
			},
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

func TestService_RemoveMember_Success(
	t *testing.T,
) {
	removeCalled := false

	repo := &mockRepository{
		getMemberFn: func(
			ctx context.Context,
			companyID uint,
			userID uint,
		) (*CompanyMember, error) {
			if userID == 5 {
				return &CompanyMember{
					CompanyID: companyID,

					UserID: 5,

					Role: "owner",

					Status: "active",
				}, nil
			}

			return &CompanyMember{
				CompanyID: companyID,

				UserID: userID,

				Role: "cleaner",

				Status: "active",
			}, nil
		},

		getByIDFn: func(
			ctx context.Context,
			companyID uint,
		) (*Company, error) {
			return &Company{
				ID: companyID,

				OwnerID: 5,
			}, nil
		},

		removeMemberFn: func(
			ctx context.Context,
			companyID uint,
			userID uint,
		) error {
			removeCalled = true
			return nil
		},
	}

	service := NewService(repo)

	err :=
		service.RemoveMember(
			context.Background(),
			10,
			5,
			8,
		)

	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}

	if !removeCalled {
		t.Fatal(
			"expected member removal",
		)
	}
}

func TestService_RemoveMember_CannotRemoveOwner(
	t *testing.T,
) {
	repo := &mockRepository{
		getMemberFn: func(
			ctx context.Context,
			companyID uint,
			userID uint,
		) (*CompanyMember, error) {
			return &CompanyMember{
				CompanyID: companyID,

				UserID: userID,

				Role: "owner",

				Status: "active",
			}, nil
		},

		getByIDFn: func(
			ctx context.Context,
			companyID uint,
		) (*Company, error) {
			return &Company{
				ID:      companyID,
				OwnerID: 5,
			}, nil
		},
	}

	service := NewService(repo)

	err :=
		service.RemoveMember(
			context.Background(),
			10,
			5,
			5,
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

func TestService_ListMembers_Success(
	t *testing.T,
) {
	repo := &mockRepository{
		getMemberFn: func(
			ctx context.Context,
			companyID uint,
			userID uint,
		) (*CompanyMember, error) {
			return &CompanyMember{
				CompanyID: companyID,
				UserID:    userID,
				Role:      "cleaner",
				Status:    "active",
			}, nil
		},

		listMembersFn: func(
			ctx context.Context,
			companyID uint,
		) ([]CompanyMember, error) {
			return []CompanyMember{
				{
					ID:        1,
					CompanyID: companyID,
					UserID:    5,
					FullName:  "Company Owner",
					Role:      "owner",
					Status:    "active",
				},
				{
					ID:        2,
					CompanyID: companyID,
					UserID:    8,
					FullName:  "Cleaner One",
					Role:      "cleaner",
					Status:    "active",
				},
			}, nil
		},
	}

	service := NewService(repo)

	result, err :=
		service.ListMembers(
			context.Background(),
			10,
			8,
		)

	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}

	if len(result) != 2 {
		t.Fatalf(
			"expected 2 members, got %d",
			len(result),
		)
	}
}

func TestService_UpdateMemberStatus_DeactivateCleaner(
	t *testing.T,
) {
	statusUpdated := false

	repo := &mockRepository{
		getMemberFn: func(
			ctx context.Context,
			companyID uint,
			userID uint,
		) (*CompanyMember, error) {
			if userID == 5 {
				return &CompanyMember{
					CompanyID: companyID,

					UserID: 5,

					Role: "owner",

					Status: "active",
				}, nil
			}

			return &CompanyMember{
				CompanyID: companyID,

				UserID: userID,

				Role: "cleaner",

				Status: "active",
			}, nil
		},

		getByIDFn: func(
			ctx context.Context,
			companyID uint,
		) (*Company, error) {
			return &Company{
				ID:      companyID,
				OwnerID: 5,
			}, nil
		},

		updateMemberStatusFn: func(
			ctx context.Context,
			companyID uint,
			userID uint,
			status string,
		) error {
			if status != "inactive" {
				t.Fatalf(
					"expected inactive, got %q",
					status,
				)
			}

			statusUpdated = true

			return nil
		},
	}

	service := NewService(repo)

	err :=
		service.UpdateMemberStatus(
			context.Background(),
			10,
			5,
			8,
			UpdateMemberStatusRequest{
				Status: "inactive",
			},
		)

	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}

	if !statusUpdated {
		t.Fatal(
			"expected status update",
		)
	}
}

func TestService_UpdateMemberStatus_ReactivationSeatLimit(
	t *testing.T,
) {
	repo := &mockRepository{
		getMemberFn: func(
			ctx context.Context,
			companyID uint,
			userID uint,
		) (*CompanyMember, error) {
			if userID == 5 {
				return &CompanyMember{
					CompanyID: companyID,

					UserID: 5,

					Role: "owner",

					Status: "active",
				}, nil
			}

			return &CompanyMember{
				CompanyID: companyID,

				UserID: userID,

				Role: "cleaner",

				Status: "inactive",
			}, nil
		},

		getByIDFn: func(
			ctx context.Context,
			companyID uint,
		) (*Company, error) {
			return &Company{
				ID:      companyID,
				OwnerID: 5,
			}, nil
		},

		getCleanerSeatLimitFn: func(
			ctx context.Context,
			companyID uint,
		) (int, error) {
			return 3, nil
		},

		countActiveCleanerSeatsFn: func(
			ctx context.Context,
			companyID uint,
		) (int, error) {
			return 3, nil
		},
	}

	service := NewService(repo)

	err :=
		service.UpdateMemberStatus(
			context.Background(),
			10,
			5,
			8,
			UpdateMemberStatusRequest{
				Status: "active",
			},
		)

	if !errors.Is(
		err,
		ErrSeatLimitReached,
	) {
		t.Fatalf(
			"expected ErrSeatLimitReached, got %v",
			err,
		)
	}
}

func TestService_GetDashboard_Success(
	t *testing.T,
) {
	repo := &mockRepository{
		getMemberFn: func(
			ctx context.Context,
			companyID uint,
			userID uint,
		) (*CompanyMember, error) {
			return &CompanyMember{
				CompanyID: companyID,

				UserID: userID,

				Role: "owner",

				Status: "active",
			}, nil
		},

		getByIDFn: func(
			ctx context.Context,
			companyID uint,
		) (*Company, error) {
			return &Company{
				ID: companyID,

				OwnerID: 5,

				Name: "Wembi Cleaning Ltd",
			}, nil
		},

		listMembersFn: func(
			ctx context.Context,
			companyID uint,
		) ([]CompanyMember, error) {
			return []CompanyMember{
				{
					UserID: 5,
					Role:   "owner",
					Status: "active",
				},
				{
					UserID: 8,
					Role:   "cleaner",
					Status: "active",
				},
				{
					UserID: 9,
					Role:   "cleaner",
					Status: "inactive",
				},
			}, nil
		},

		getCleanerSeatLimitFn: func(
			ctx context.Context,
			companyID uint,
		) (int, error) {
			return 3, nil
		},

		countMembersByRolesAndStatusFn: func(
			ctx context.Context,
			companyID uint,
			role string,
			status string,
		) (int, error) {
			switch {
			case role == "cleaner" &&
				status == "active":
				return 1, nil

			case role == "cleaner" &&
				status == "inactive":
				return 1, nil

			case role == "admin":
				return 0, nil

			default:
				return 0, nil
			}
		},
	}

	service := NewService(repo)

	result, err :=
		service.GetDashboard(
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

	if result.TotalMembers != 3 {
		t.Fatalf(
			"expected 3 members, got %d",
			result.TotalMembers,
		)
	}

	if result.ActiveCleaners != 1 {
		t.Fatalf(
			"expected 1 active cleaner, got %d",
			result.ActiveCleaners,
		)
	}

	if result.SeatUsage.SeatLimit != 3 {
		t.Fatalf(
			"expected seat limit 3, got %d",
			result.SeatUsage.SeatLimit,
		)
	}

	if result.SeatUsage.SeatsUsed != 1 {
		t.Fatalf(
			"expected 1 used seat, got %d",
			result.SeatUsage.SeatsUsed,
		)
	}

	if result.SeatUsage.SeatsRemaining != 2 {
		t.Fatalf(
			"expected 2 remaining seats, got %d",
			result.SeatUsage.SeatsRemaining,
		)
	}
}
