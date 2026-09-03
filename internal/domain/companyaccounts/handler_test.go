package companyaccounts

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/rodrigueghenda/jobira/internal/security/identity"
)

func companyRequestWithUser(
	method string,
	path string,
	body []byte,
	userID uint,
) *http.Request {
	req := httptest.NewRequest(
		method,
		path,
		bytes.NewReader(body),
	)

	ctx := identity.WithUser(
		req.Context(),
		identity.UserIdentity{
			UserID: userID,
		},
	)

	return req.WithContext(ctx)
}

func companyRequestWithParams(
	req *http.Request,
	params map[string]string,
) *http.Request {
	routeCtx := chi.NewRouteContext()

	for key, value := range params {
		routeCtx.URLParams.Add(
			key,
			value,
		)
	}

	ctx := context.WithValue(
		req.Context(),
		chi.RouteCtxKey,
		routeCtx,
	)

	return req.WithContext(ctx)
}

func TestHandler_CreateCompany_Success(
	t *testing.T,
) {
	repo := &mockRepository{
		createCompanyFn: func(
			ctx context.Context,
			company *Company,
		) error {
			company.ID = 10
			return nil
		},

		addMemberFn: func(
			ctx context.Context,
			companyID uint,
			userID uint,
			role string,
		) error {
			return nil
		},
	}

	handler := NewHandler(
		NewService(repo),
	)

	body := []byte(`{
		"name":"Jobira Cleaning Ltd",
		"description":"London cleaning company"
	}`)

	req := companyRequestWithUser(
		http.MethodPost,
		"/companies/",
		body,
		5,
	)

	recorder :=
		httptest.NewRecorder()

	handler.CreateCompany(
		recorder,
		req,
	)

	if recorder.Code !=
		http.StatusCreated {
		t.Fatalf(
			"expected status 201, got %d: %s",
			recorder.Code,
			recorder.Body.String(),
		)
	}

	var result Company

	if err := json.NewDecoder(
		recorder.Body,
	).Decode(
		&result,
	); err != nil {
		t.Fatalf(
			"failed to decode response: %v",
			err,
		)
	}

	if result.ID != 10 {
		t.Fatalf(
			"expected company ID 10, got %d",
			result.ID,
		)
	}
}

func TestHandler_CreateCompany_InvalidBody(
	t *testing.T,
) {
	handler := NewHandler(
		NewService(
			&mockRepository{},
		),
	)

	req := companyRequestWithUser(
		http.MethodPost,
		"/companies/",
		[]byte(`{invalid`),
		5,
	)

	recorder :=
		httptest.NewRecorder()

	handler.CreateCompany(
		recorder,
		req,
	)

	if recorder.Code !=
		http.StatusBadRequest {
		t.Fatalf(
			"expected status 400, got %d",
			recorder.Code,
		)
	}
}

func TestHandler_AddMember_Success(
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
				Role:      "owner",
				Status:    "active",
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
			return 1, nil
		},

		addMemberFn: func(
			ctx context.Context,
			companyID uint,
			userID uint,
			role string,
		) error {
			return nil
		},
	}

	handler := NewHandler(
		NewService(repo),
	)

	body := []byte(`{
		"user_id":8,
		"role":"cleaner"
	}`)

	req := companyRequestWithUser(
		http.MethodPost,
		"/companies/10/members",
		body,
		5,
	)

	req = companyRequestWithParams(
		req,
		map[string]string{
			"companyID": "10",
		},
	)

	recorder :=
		httptest.NewRecorder()

	handler.AddMember(
		recorder,
		req,
	)

	if recorder.Code !=
		http.StatusCreated {
		t.Fatalf(
			"expected status 201, got %d: %s",
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_AddMember_SeatLimitReached(
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
				Role:      "owner",
				Status:    "active",
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

	handler := NewHandler(
		NewService(repo),
	)

	body := []byte(`{
		"user_id":8,
		"role":"cleaner"
	}`)

	req := companyRequestWithUser(
		http.MethodPost,
		"/companies/10/members",
		body,
		5,
	)

	req = companyRequestWithParams(
		req,
		map[string]string{
			"companyID": "10",
		},
	)

	recorder :=
		httptest.NewRecorder()

	handler.AddMember(
		recorder,
		req,
	)

	if recorder.Code !=
		http.StatusPaymentRequired {
		t.Fatalf(
			"expected status 402, got %d: %s",
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_ListMembers_Success(
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

	handler := NewHandler(
		NewService(repo),
	)

	req := companyRequestWithUser(
		http.MethodGet,
		"/companies/10/members",
		nil,
		8,
	)

	req = companyRequestWithParams(
		req,
		map[string]string{
			"companyID": "10",
		},
	)

	recorder :=
		httptest.NewRecorder()

	handler.ListMembers(
		recorder,
		req,
	)

	if recorder.Code !=
		http.StatusOK {
		t.Fatalf(
			"expected status 200, got %d: %s",
			recorder.Code,
			recorder.Body.String(),
		)
	}

	var result []CompanyMember

	if err := json.NewDecoder(
		recorder.Body,
	).Decode(
		&result,
	); err != nil {
		t.Fatalf(
			"failed to decode response: %v",
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

func TestHandler_RemoveMember_Success(
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
					UserID:    5,
					Role:      "owner",
					Status:    "active",
				}, nil
			}

			return &CompanyMember{
				CompanyID: companyID,
				UserID:    userID,
				Role:      "cleaner",
				Status:    "active",
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

		removeMemberFn: func(
			ctx context.Context,
			companyID uint,
			userID uint,
		) error {
			return nil
		},
	}

	handler := NewHandler(
		NewService(repo),
	)

	req := companyRequestWithUser(
		http.MethodDelete,
		"/companies/10/members/8",
		nil,
		5,
	)

	req = companyRequestWithParams(
		req,
		map[string]string{
			"companyID": "10",
			"userID":    "8",
		},
	)

	recorder :=
		httptest.NewRecorder()

	handler.RemoveMember(
		recorder,
		req,
	)

	if recorder.Code !=
		http.StatusOK {
		t.Fatalf(
			"expected status 200, got %d: %s",
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_UpdateMemberStatus_Success(
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
					UserID:    5,
					Role:      "owner",
					Status:    "active",
				}, nil
			}

			return &CompanyMember{
				CompanyID: companyID,
				UserID:    userID,
				Role:      "cleaner",
				Status:    "active",
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

			return nil
		},
	}

	handler := NewHandler(
		NewService(repo),
	)

	req := companyRequestWithUser(
		http.MethodPatch,
		"/companies/10/members/8/status",
		[]byte(`{"status":"inactive"}`),
		5,
	)

	req = companyRequestWithParams(
		req,
		map[string]string{
			"companyID": "10",
			"userID":    "8",
		},
	)

	recorder :=
		httptest.NewRecorder()

	handler.UpdateMemberStatus(
		recorder,
		req,
	)

	if recorder.Code !=
		http.StatusOK {
		t.Fatalf(
			"expected status 200, got %d: %s",
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_UpdateMemberRole_Success(
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
					UserID:    5,
					Role:      "owner",
					Status:    "active",
				}, nil
			}

			return &CompanyMember{
				CompanyID: companyID,
				UserID:    userID,
				Role:      "cleaner",
				Status:    "active",
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

		updateMemberRoleFn: func(
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

	handler := NewHandler(
		NewService(repo),
	)

	req := companyRequestWithUser(
		http.MethodPatch,
		"/companies/10/members/8/role",
		[]byte(`{"role":"admin"}`),
		5,
	)

	req = companyRequestWithParams(
		req,
		map[string]string{
			"companyID": "10",
			"userID":    "8",
		},
	)

	recorder :=
		httptest.NewRecorder()

	handler.UpdateMemberRole(
		recorder,
		req,
	)

	if recorder.Code !=
		http.StatusOK {
		t.Fatalf(
			"expected status 200, got %d: %s",
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_GetDashboard_Success(
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
				Role:      "owner",
				Status:    "active",
			}, nil
		},

		getByIDFn: func(
			ctx context.Context,
			companyID uint,
		) (*Company, error) {
			return &Company{
				ID:      companyID,
				OwnerID: 5,
				Name:    "Jobira Cleaning Ltd",
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
				return 0, nil

			case role == "admin":
				return 0, nil
			}

			return 0, nil
		},
	}

	handler := NewHandler(
		NewService(repo),
	)

	req := companyRequestWithUser(
		http.MethodGet,
		"/companies/10/dashboard",
		nil,
		5,
	)

	req = companyRequestWithParams(
		req,
		map[string]string{
			"companyID": "10",
		},
	)

	recorder :=
		httptest.NewRecorder()

	handler.GetDashboard(
		recorder,
		req,
	)

	if recorder.Code !=
		http.StatusOK {
		t.Fatalf(
			"expected status 200, got %d: %s",
			recorder.Code,
			recorder.Body.String(),
		)
	}

	var result CompanyDashboard

	if err := json.NewDecoder(
		recorder.Body,
	).Decode(
		&result,
	); err != nil {
		t.Fatalf(
			"failed to decode response: %v",
			err,
		)
	}

	if result.SeatUsage.SeatLimit != 3 {
		t.Fatalf(
			"expected seat limit 3, got %d",
			result.SeatUsage.SeatLimit,
		)
	}

	if result.SeatUsage.SeatsRemaining != 2 {
		t.Fatalf(
			"expected 2 seats remaining, got %d",
			result.SeatUsage.SeatsRemaining,
		)
	}
}
