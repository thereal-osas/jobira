package availability

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestService_GetByID_InvalidInput_Edge(t *testing.T) {
	service := NewService(&mockRepository{})

	tests := []struct {
		name   string
		id     uint
		userID uint
	}{
		{
			name:   "zero availability ID",
			id:     0,
			userID: 8,
		},
		{
			name:   "zero user ID",
			id:     1,
			userID: 0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			record, err := service.GetByID(
				context.Background(),
				test.id,
				test.userID,
				"cleaner",
			)

			if record != nil {
				t.Fatalf("expected nil record, got %+v", record)
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

func TestService_GetByID_RepositoryError_Edge(t *testing.T) {
	expectedErr := errors.New("lookup failed")

	repo := &mockRepository{
		getByIDFn: func(
			context.Context,
			uint,
		) (*CleanerAvailability, error) {
			return nil, expectedErr
		},
	}

	service := NewService(repo)

	record, err := service.GetByID(
		context.Background(),
		1,
		8,
		"cleaner",
	)

	if record != nil {
		t.Fatalf("expected nil record, got %+v", record)
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
}

func TestService_ListByCleanerID_InvalidInput_Edge(t *testing.T) {
	service := NewService(&mockRepository{})

	records, err := service.ListByCleanerID(
		context.Background(),
		0,
	)

	if records != nil {
		t.Fatalf("expected nil records, got %+v", records)
	}

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestService_ListByCleanerID_RepositoryError_Edge(t *testing.T) {
	expectedErr := errors.New("list failed")

	repo := &mockRepository{
		listByCleanerIDFn: func(
			context.Context,
			uint,
		) ([]CleanerAvailability, error) {
			return nil, expectedErr
		},
	}

	service := NewService(repo)

	records, err := service.ListByCleanerID(
		context.Background(),
		8,
	)

	if records != nil {
		t.Fatalf("expected nil records, got %+v", records)
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
}

func TestService_ListMine_InvalidInput_Edge(t *testing.T) {
	service := NewService(&mockRepository{})

	records, err := service.ListMine(
		context.Background(),
		0,
	)

	if records != nil {
		t.Fatalf("expected nil records, got %+v", records)
	}

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestService_ListMine_RepositoryError_Edge(t *testing.T) {
	expectedErr := errors.New("list mine failed")

	repo := &mockRepository{
		listByCleanerIDFn: func(
			context.Context,
			uint,
		) ([]CleanerAvailability, error) {
			return nil, expectedErr
		},
	}

	service := NewService(repo)

	records, err := service.ListMine(
		context.Background(),
		8,
	)

	if records != nil {
		t.Fatalf("expected nil records, got %+v", records)
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
}

func TestService_Update_InvalidIDs_Edge(t *testing.T) {
	service := NewService(&mockRepository{})

	tests := []struct {
		name   string
		id     uint
		userID uint
	}{
		{
			name:   "zero availability ID",
			id:     0,
			userID: 8,
		},
		{
			name:   "zero user ID",
			id:     1,
			userID: 0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			record, err := service.Update(
				context.Background(),
				test.id,
				test.userID,
				"cleaner",
				UpdateAvailabilityRequest{},
			)

			if record != nil {
				t.Fatalf("expected nil record, got %+v", record)
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

func TestService_Update_GetByIDError_Edge(t *testing.T) {
	expectedErr := errors.New("lookup failed")

	repo := &mockRepository{
		getByIDFn: func(
			context.Context,
			uint,
		) (*CleanerAvailability, error) {
			return nil, expectedErr
		},
	}

	service := NewService(repo)

	record, err := service.Update(
		context.Background(),
		1,
		8,
		"cleaner",
		UpdateAvailabilityRequest{},
	)

	if record != nil {
		t.Fatalf("expected nil record, got %+v", record)
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
}

func TestService_Update_Forbidden_Edge(t *testing.T) {
	repo := &mockRepository{
		getByIDFn: func(
			context.Context,
			uint,
		) (*CleanerAvailability, error) {
			return &CleanerAvailability{
				ID:        1,
				CleanerID: 8,
			}, nil
		},
	}

	service := NewService(repo)

	record, err := service.Update(
		context.Background(),
		1,
		99,
		"cleaner",
		UpdateAvailabilityRequest{},
	)

	if record != nil {
		t.Fatalf("expected nil record, got %+v", record)
	}

	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestService_Update_InvalidPayload_Edge(t *testing.T) {
	repo := &mockRepository{
		getByIDFn: func(
			context.Context,
			uint,
		) (*CleanerAvailability, error) {
			return &CleanerAvailability{
				ID:        1,
				CleanerID: 8,
			}, nil
		},
	}

	service := NewService(repo)

	tests := []struct {
		name string
		req  UpdateAvailabilityRequest
		err  error
	}{
		{
			name: "missing required fields",
			req:  UpdateAvailabilityRequest{},
			err:  ErrInvalidInput,
		},
		{
			name: "invalid date",
			req: UpdateAvailabilityRequest{
				AvailableDate: "tomorrow",
				StartTime:     "09:00",
				EndTime:       "17:00",
				Status:        "available",
			},
			err: ErrInvalidInput,
		},
		{
			name: "invalid status",
			req: UpdateAvailabilityRequest{
				AvailableDate: "2026-08-10",
				StartTime:     "09:00",
				EndTime:       "17:00",
				Status:        "asleep",
			},
			err: ErrInvalidStatus,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			record, err := service.Update(
				context.Background(),
				1,
				8,
				"cleaner",
				test.req,
			)

			if record != nil {
				t.Fatalf("expected nil record, got %+v", record)
			}

			if !errors.Is(err, test.err) {
				t.Fatalf("expected %v, got %v", test.err, err)
			}
		})
	}
}

func TestService_Update_ConflictCheckError_Edge(t *testing.T) {
	expectedErr := errors.New("conflict lookup failed")

	repo := &mockRepository{
		getByIDFn: func(
			context.Context,
			uint,
		) (*CleanerAvailability, error) {
			return &CleanerAvailability{
				ID:        1,
				CleanerID: 8,
			}, nil
		},
		hasConflictFn: func(
			context.Context,
			uint,
			string,
			string,
			string,
			uint,
		) (bool, error) {
			return false, expectedErr
		},
	}

	service := NewService(repo)

	record, err := service.Update(
		context.Background(),
		1,
		8,
		"cleaner",
		UpdateAvailabilityRequest{
			AvailableDate: "2026-08-10",
			StartTime:     "09:00",
			EndTime:       "17:00",
			Status:        "available",
		},
	)

	if record != nil {
		t.Fatalf("expected nil record, got %+v", record)
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
}

func TestService_Update_RepositoryError_Edge(t *testing.T) {
	expectedErr := errors.New("update failed")

	repo := &mockRepository{
		getByIDFn: func(
			context.Context,
			uint,
		) (*CleanerAvailability, error) {
			return &CleanerAvailability{
				ID:        1,
				CleanerID: 8,
			}, nil
		},
		updateFn: func(
			context.Context,
			*CleanerAvailability,
		) error {
			return expectedErr
		},
	}

	service := NewService(repo)

	record, err := service.Update(
		context.Background(),
		1,
		8,
		"cleaner",
		UpdateAvailabilityRequest{
			AvailableDate: "2026-08-10",
			StartTime:     "09:00",
			EndTime:       "17:00",
			Status:        "available",
		},
	)

	if record != nil {
		t.Fatalf("expected nil record, got %+v", record)
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
}

func TestService_Update_FinalGetByIDError_Edge(t *testing.T) {
	getCalls := 0
	expectedErr := errors.New("final lookup failed")

	repo := &mockRepository{
		getByIDFn: func(
			context.Context,
			uint,
		) (*CleanerAvailability, error) {
			getCalls++

			if getCalls == 1 {
				return &CleanerAvailability{
					ID:        1,
					CleanerID: 8,
				}, nil
			}

			return nil, expectedErr
		},
	}

	service := NewService(repo)

	record, err := service.Update(
		context.Background(),
		1,
		8,
		"cleaner",
		UpdateAvailabilityRequest{
			AvailableDate: "2026-08-10",
			StartTime:     "09:00",
			EndTime:       "17:00",
			Status:        "available",
		},
	)

	if record != nil {
		t.Fatalf("expected nil record, got %+v", record)
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
}

func TestService_Delete_InvalidInput_Edge(t *testing.T) {
	service := NewService(&mockRepository{})

	err := service.Delete(
		context.Background(),
		0,
		8,
		"cleaner",
	)

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestService_Delete_GetByIDError_Edge(t *testing.T) {
	expectedErr := errors.New("lookup failed")

	repo := &mockRepository{
		getByIDFn: func(
			context.Context,
			uint,
		) (*CleanerAvailability, error) {
			return nil, expectedErr
		},
	}

	service := NewService(repo)

	err := service.Delete(
		context.Background(),
		1,
		8,
		"cleaner",
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
}

func TestService_Delete_Forbidden_Edge(t *testing.T) {
	repo := &mockRepository{
		getByIDFn: func(
			context.Context,
			uint,
		) (*CleanerAvailability, error) {
			return &CleanerAvailability{
				ID:        1,
				CleanerID: 8,
			}, nil
		},
	}

	service := NewService(repo)

	err := service.Delete(
		context.Background(),
		1,
		99,
		"cleaner",
	)

	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestService_Delete_RepositoryError_Edge(t *testing.T) {
	expectedErr := errors.New("delete failed")

	repo := &mockRepository{
		getByIDFn: func(
			context.Context,
			uint,
		) (*CleanerAvailability, error) {
			return &CleanerAvailability{
				ID:        1,
				CleanerID: 8,
			}, nil
		},
		deleteFn: func(
			context.Context,
			uint,
		) error {
			return expectedErr
		},
	}

	service := NewService(repo)

	err := service.Delete(
		context.Background(),
		1,
		8,
		"cleaner",
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
}

func TestService_CreateBlock_InvalidEndTime_Edge(t *testing.T) {
	service := NewService(&mockRepository{})

	block, err := service.CreateBlock(
		context.Background(),
		8,
		CreateAvailabilityBlockRequest{
			StartAt: "2026-08-10T09:00:00Z",
			EndAt:   "not-a-time",
			Reason:  "Personal",
		},
	)

	if block != nil {
		t.Fatalf("expected nil block, got %+v", block)
	}

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestService_CreateBlock_BlankReason_Edge(t *testing.T) {
	service := NewService(&mockRepository{})

	block, err := service.CreateBlock(
		context.Background(),
		8,
		CreateAvailabilityBlockRequest{
			StartAt: "2026-08-10T09:00:00Z",
			EndAt:   "2026-08-10T17:00:00Z",
			Reason:  "   ",
		},
	)

	if block != nil {
		t.Fatalf("expected nil block, got %+v", block)
	}

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestService_ListBlocks_InvalidInput_Edge(t *testing.T) {
	service := NewService(&mockRepository{})

	blocks, err := service.ListBlocks(
		context.Background(),
		0,
	)

	if blocks != nil {
		t.Fatalf("expected nil blocks, got %+v", blocks)
	}

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestService_ListBlocks_RepositoryError_Edge(t *testing.T) {
	expectedErr := errors.New("list blocks failed")

	repo := &mockRepository{
		listBlocksByCleanerIDFn: func(
			context.Context,
			uint,
		) ([]AvailabilityBlock, error) {
			return nil, expectedErr
		},
	}

	service := NewService(repo)

	blocks, err := service.ListBlocks(
		context.Background(),
		8,
	)

	if blocks != nil {
		t.Fatalf("expected nil blocks, got %+v", blocks)
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
}

func TestHandler_GetByID_Unauthorized_Edge(t *testing.T) {
	handler := newHandlerForTest(&mockRepository{})

	req := httptest.NewRequest(
		http.MethodGet,
		"/availability/1",
		nil,
	)

	recorder := httptest.NewRecorder()

	handler.GetByID(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected %d, got %d",
			http.StatusUnauthorized,
			recorder.Code,
		)
	}
}

func TestHandler_GetByID_InvalidID_Edge(t *testing.T) {
	handler := newHandlerForTest(&mockRepository{})

	req := httptest.NewRequest(
		http.MethodGet,
		"/availability/nope",
		nil,
	)
	req = requestWithURLParam(req, "availabilityID", "nope")
	req = requestWithUser(req, 8, "cleaner")

	recorder := httptest.NewRecorder()

	handler.GetByID(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected %d, got %d",
			http.StatusBadRequest,
			recorder.Code,
		)
	}
}

func TestHandler_GetByID_NotFound_Edge(t *testing.T) {
	repo := &mockRepository{
		getByIDFn: func(
			context.Context,
			uint,
		) (*CleanerAvailability, error) {
			return nil, ErrAvailabilityNotFound
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodGet,
		"/availability/1",
		nil,
	)
	req = requestWithURLParam(req, "availabilityID", "1")
	req = requestWithUser(req, 8, "cleaner")

	recorder := httptest.NewRecorder()

	handler.GetByID(recorder, req)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf(
			"expected %d, got %d: %s",
			http.StatusForbidden,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_GetByID_InternalServerError_Edge(t *testing.T) {
	expectedErr := errors.New("lookup failed")

	repo := &mockRepository{
		getByIDFn: func(
			context.Context,
			uint,
		) (*CleanerAvailability, error) {
			return nil, expectedErr
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodGet,
		"/availability/1",
		nil,
	)
	req = requestWithURLParam(req, "availabilityID", "1")
	req = requestWithUser(req, 8, "cleaner")

	recorder := httptest.NewRecorder()

	handler.GetByID(recorder, req)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected %d, got %d: %s",
			http.StatusInternalServerError,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_ListMine_Unauthorized_Edge(t *testing.T) {
	handler := newHandlerForTest(&mockRepository{})

	req := httptest.NewRequest(
		http.MethodGet,
		"/availability/me",
		nil,
	)

	recorder := httptest.NewRecorder()

	handler.ListMine(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected %d, got %d",
			http.StatusUnauthorized,
			recorder.Code,
		)
	}
}

func TestHandler_ListMine_InternalServerError_Edge(t *testing.T) {
	expectedErr := errors.New("list failed")

	repo := &mockRepository{
		listByCleanerIDFn: func(
			context.Context,
			uint,
		) ([]CleanerAvailability, error) {
			return nil, expectedErr
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodGet,
		"/availability/me",
		nil,
	)
	req = requestWithUser(req, 8, "cleaner")

	recorder := httptest.NewRecorder()

	handler.ListMine(recorder, req)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected %d, got %d: %s",
			http.StatusInternalServerError,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_ListByCleaner_InvalidID_Edge(t *testing.T) {
	handler := newHandlerForTest(&mockRepository{})

	req := httptest.NewRequest(
		http.MethodGet,
		"/availability/cleaners/nope",
		nil,
	)
	req = requestWithURLParam(req, "cleanerID", "nope")

	recorder := httptest.NewRecorder()

	handler.ListByCleaner(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected %d, got %d",
			http.StatusBadRequest,
			recorder.Code,
		)
	}
}

func TestHandler_ListByCleaner_InternalServerError_Edge(t *testing.T) {
	expectedErr := errors.New("list failed")

	repo := &mockRepository{
		listByCleanerIDFn: func(
			context.Context,
			uint,
		) ([]CleanerAvailability, error) {
			return nil, expectedErr
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodGet,
		"/availability/cleaners/8",
		nil,
	)
	req = requestWithURLParam(req, "cleanerID", "8")

	recorder := httptest.NewRecorder()

	handler.ListByCleaner(recorder, req)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected %d, got %d: %s",
			http.StatusInternalServerError,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Update_Unauthorized_Edge(t *testing.T) {
	handler := newHandlerForTest(&mockRepository{})

	req := httptest.NewRequest(
		http.MethodPut,
		"/availability/1",
		strings.NewReader(`{}`),
	)

	recorder := httptest.NewRecorder()

	handler.Update(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected %d, got %d",
			http.StatusUnauthorized,
			recorder.Code,
		)
	}
}

func TestHandler_Update_InvalidID_Edge(t *testing.T) {
	handler := newHandlerForTest(&mockRepository{})

	req := httptest.NewRequest(
		http.MethodPut,
		"/availability/nope",
		strings.NewReader(`{}`),
	)
	req = requestWithURLParam(req, "availabilityID", "nope")
	req = requestWithUser(req, 8, "cleaner")

	recorder := httptest.NewRecorder()

	handler.Update(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected %d, got %d",
			http.StatusBadRequest,
			recorder.Code,
		)
	}
}

func TestHandler_Update_InvalidJSON_Edge(t *testing.T) {
	handler := newHandlerForTest(&mockRepository{})

	req := httptest.NewRequest(
		http.MethodPut,
		"/availability/1",
		strings.NewReader(`{"available_date":`),
	)
	req = requestWithURLParam(req, "availabilityID", "1")
	req = requestWithUser(req, 8, "cleaner")

	recorder := httptest.NewRecorder()

	handler.Update(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected %d, got %d",
			http.StatusBadRequest,
			recorder.Code,
		)
	}
}

func TestHandler_Update_Forbidden_Edge(t *testing.T) {
	repo := &mockRepository{
		getByIDFn: func(
			context.Context,
			uint,
		) (*CleanerAvailability, error) {
			return &CleanerAvailability{
				ID:        1,
				CleanerID: 8,
			}, nil
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodPut,
		"/availability/1",
		strings.NewReader(`{
			"available_date":"2026-08-10",
			"start_time":"09:00",
			"end_time":"17:00",
			"status":"available"
		}`),
	)
	req = requestWithURLParam(req, "availabilityID", "1")
	req = requestWithUser(req, 99, "cleaner")

	recorder := httptest.NewRecorder()

	handler.Update(recorder, req)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf(
			"expected %d, got %d: %s",
			http.StatusForbidden,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Update_InternalServerError_Edge(t *testing.T) {
	expectedErr := errors.New("update failed")

	repo := &mockRepository{
		getByIDFn: func(
			context.Context,
			uint,
		) (*CleanerAvailability, error) {
			return &CleanerAvailability{
				ID:        1,
				CleanerID: 8,
			}, nil
		},
		updateFn: func(
			context.Context,
			*CleanerAvailability,
		) error {
			return expectedErr
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodPut,
		"/availability/1",
		strings.NewReader(`{
			"available_date":"2026-08-10",
			"start_time":"09:00",
			"end_time":"17:00",
			"status":"available"
		}`),
	)
	req = requestWithURLParam(req, "availabilityID", "1")
	req = requestWithUser(req, 8, "cleaner")

	recorder := httptest.NewRecorder()

	handler.Update(recorder, req)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected %d, got %d: %s",
			http.StatusInternalServerError,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Delete_Unauthorized_Edge(t *testing.T) {
	handler := newHandlerForTest(&mockRepository{})

	req := httptest.NewRequest(
		http.MethodDelete,
		"/availability/1",
		nil,
	)

	recorder := httptest.NewRecorder()

	handler.Delete(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected %d, got %d",
			http.StatusUnauthorized,
			recorder.Code,
		)
	}
}

func TestHandler_Delete_InvalidID_Edge(t *testing.T) {
	handler := newHandlerForTest(&mockRepository{})

	req := httptest.NewRequest(
		http.MethodDelete,
		"/availability/nope",
		nil,
	)
	req = requestWithURLParam(req, "availabilityID", "nope")
	req = requestWithUser(req, 8, "cleaner")

	recorder := httptest.NewRecorder()

	handler.Delete(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected %d, got %d",
			http.StatusBadRequest,
			recorder.Code,
		)
	}
}

func TestHandler_Delete_Forbidden_Edge(t *testing.T) {
	repo := &mockRepository{
		getByIDFn: func(
			context.Context,
			uint,
		) (*CleanerAvailability, error) {
			return &CleanerAvailability{
				ID:        1,
				CleanerID: 8,
			}, nil
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodDelete,
		"/availability/1",
		nil,
	)
	req = requestWithURLParam(req, "availabilityID", "1")
	req = requestWithUser(req, 99, "cleaner")

	recorder := httptest.NewRecorder()

	handler.Delete(recorder, req)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf(
			"expected %d, got %d: %s",
			http.StatusForbidden,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_Delete_InternalServerError_Edge(t *testing.T) {
	expectedErr := errors.New("delete failed")

	repo := &mockRepository{
		getByIDFn: func(
			context.Context,
			uint,
		) (*CleanerAvailability, error) {
			return &CleanerAvailability{
				ID:        1,
				CleanerID: 8,
			}, nil
		},
		deleteFn: func(
			context.Context,
			uint,
		) error {
			return expectedErr
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodDelete,
		"/availability/1",
		nil,
	)
	req = requestWithURLParam(req, "availabilityID", "1")
	req = requestWithUser(req, 8, "cleaner")

	recorder := httptest.NewRecorder()

	handler.Delete(recorder, req)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected %d, got %d: %s",
			http.StatusInternalServerError,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_CreateBlock_Unauthorized_Edge(t *testing.T) {
	handler := newHandlerForTest(&mockRepository{})

	req := httptest.NewRequest(
		http.MethodPost,
		"/availability/blocks",
		strings.NewReader(`{}`),
	)

	recorder := httptest.NewRecorder()

	handler.CreateBlock(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected %d, got %d",
			http.StatusUnauthorized,
			recorder.Code,
		)
	}
}

func TestHandler_CreateBlock_InvalidInput_Edge(t *testing.T) {
	handler := newHandlerForTest(&mockRepository{})

	req := httptest.NewRequest(
		http.MethodPost,
		"/availability/blocks",
		strings.NewReader(`{
			"start_at":"",
			"end_at":"",
			"reason":""
		}`),
	)
	req = requestWithUser(req, 8, "cleaner")

	recorder := httptest.NewRecorder()

	handler.CreateBlock(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected %d, got %d: %s",
			http.StatusBadRequest,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_CreateBlock_InternalServerError_Edge(t *testing.T) {
	expectedErr := errors.New("create block failed")

	repo := &mockRepository{
		createBlockFn: func(
			context.Context,
			*AvailabilityBlock,
		) error {
			return expectedErr
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodPost,
		"/availability/blocks",
		strings.NewReader(`{
			"start_at":"2026-08-10T09:00:00Z",
			"end_at":"2026-08-10T17:00:00Z",
			"reason":"Holiday"
		}`),
	)
	req = requestWithUser(req, 8, "cleaner")

	recorder := httptest.NewRecorder()

	handler.CreateBlock(recorder, req)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected %d, got %d: %s",
			http.StatusInternalServerError,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_ListBlocks_Unauthorized_Edge(t *testing.T) {
	handler := newHandlerForTest(&mockRepository{})

	req := httptest.NewRequest(
		http.MethodGet,
		"/availability/blocks/me",
		nil,
	)

	recorder := httptest.NewRecorder()

	handler.ListBlocks(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected %d, got %d",
			http.StatusUnauthorized,
			recorder.Code,
		)
	}
}

func TestHandler_ListBlocks_InternalServerError_Edge(t *testing.T) {
	expectedErr := errors.New("list blocks failed")

	repo := &mockRepository{
		listBlocksByCleanerIDFn: func(
			context.Context,
			uint,
		) ([]AvailabilityBlock, error) {
			return nil, expectedErr
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodGet,
		"/availability/blocks/me",
		nil,
	)
	req = requestWithUser(req, 8, "cleaner")

	recorder := httptest.NewRecorder()

	handler.ListBlocks(recorder, req)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected %d, got %d: %s",
			http.StatusInternalServerError,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestHandler_DeleteBlock_Unauthorized_Edge(t *testing.T) {
	handler := newHandlerForTest(&mockRepository{})

	req := httptest.NewRequest(
		http.MethodDelete,
		"/availability/blocks/1",
		nil,
	)

	recorder := httptest.NewRecorder()

	handler.DeleteBlock(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected %d, got %d",
			http.StatusUnauthorized,
			recorder.Code,
		)
	}
}

func TestHandler_DeleteBlock_InvalidID_Edge(t *testing.T) {
	handler := newHandlerForTest(&mockRepository{})

	req := httptest.NewRequest(
		http.MethodDelete,
		"/availability/blocks/nope",
		nil,
	)
	req = requestWithURLParam(req, "blockID", "nope")
	req = requestWithUser(req, 8, "cleaner")

	recorder := httptest.NewRecorder()

	handler.DeleteBlock(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected %d, got %d",
			http.StatusBadRequest,
			recorder.Code,
		)
	}
}

func TestHandler_DeleteBlock_InternalServerError_Edge(t *testing.T) {
	expectedErr := errors.New("delete block failed")

	repo := &mockRepository{
		deleteBlockFn: func(
			context.Context,
			uint,
			uint,
		) error {
			return expectedErr
		},
	}

	handler := newHandlerForTest(repo)

	req := httptest.NewRequest(
		http.MethodDelete,
		"/availability/blocks/1",
		nil,
	)
	req = requestWithURLParam(req, "blockID", "1")
	req = requestWithUser(req, 8, "cleaner")

	recorder := httptest.NewRecorder()

	handler.DeleteBlock(recorder, req)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected %d, got %d: %s",
			http.StatusInternalServerError,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}
