package clientnotes

import (
	"context"
	"errors"
	"testing"
	"time"
)

type mockRepository struct {
	createFn         func(context.Context, *ClientCleanerNote) error
	getByCleanerIDFn func(context.Context, uint, uint) ([]ClientCleanerNote, error)
	updateFn         func(context.Context, uint, uint, string) error
	deleteFn         func(context.Context, uint, uint) error
}

func (m *mockRepository) Create(
	ctx context.Context,
	note *ClientCleanerNote,
) error {
	if m.createFn != nil {
		return m.createFn(ctx, note)
	}

	return nil
}

func (m *mockRepository) GetByCleanerID(
	ctx context.Context,
	clientID uint,
	cleanerID uint,
) ([]ClientCleanerNote, error) {
	if m.getByCleanerIDFn != nil {
		return m.getByCleanerIDFn(
			ctx,
			clientID,
			cleanerID,
		)
	}

	return nil, nil
}

func (m *mockRepository) Update(
	ctx context.Context,
	noteID uint,
	clientID uint,
	note string,
) error {
	if m.updateFn != nil {
		return m.updateFn(
			ctx,
			noteID,
			clientID,
			note,
		)
	}

	return nil
}

func (m *mockRepository) Delete(
	ctx context.Context,
	noteID uint,
	clientID uint,
) error {
	if m.deleteFn != nil {
		return m.deleteFn(
			ctx,
			noteID,
			clientID,
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
	createCalled := false

	repo := &mockRepository{
		createFn: func(
			ctx context.Context,
			note *ClientCleanerNote,
		) error {
			createCalled = true

			if note.ClientID != 5 {
				t.Fatalf(
					"expected client ID 5, got %d",
					note.ClientID,
				)
			}

			if note.CleanerID != 8 {
				t.Fatalf(
					"expected cleaner ID 8, got %d",
					note.CleanerID,
				)
			}

			if note.Note != "Reliable and punctual" {
				t.Fatalf(
					"expected trimmed note, got %q",
					note.Note,
				)
			}

			now := time.Now()

			note.ID = 12
			note.CreatedAt = now
			note.UpdatedAt = now

			return nil
		},
	}

	service := NewService(repo)

	note, err := service.Create(
		context.Background(),
		5,
		8,
		CreateNoteRequest{
			Note: "  Reliable and punctual  ",
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if note == nil {
		t.Fatal("expected note")
	}

	if note.ID != 12 {
		t.Fatalf(
			"expected ID 12, got %d",
			note.ID,
		)
	}

	if note.ClientID != 5 {
		t.Fatalf(
			"expected client ID 5, got %d",
			note.ClientID,
		)
	}

	if note.CleanerID != 8 {
		t.Fatalf(
			"expected cleaner ID 8, got %d",
			note.CleanerID,
		)
	}

	if note.Note != "Reliable and punctual" {
		t.Fatalf(
			"expected trimmed note, got %q",
			note.Note,
		)
	}

	if note.CreatedAt.IsZero() {
		t.Fatal("expected created_at")
	}

	if note.UpdatedAt.IsZero() {
		t.Fatal("expected updated_at")
	}

	if !createCalled {
		t.Fatal("expected Create repository method to be called")
	}
}

func TestService_Create_InvalidInput(t *testing.T) {
	service := NewService(&mockRepository{})

	tests := []struct {
		name      string
		clientID  uint
		cleanerID uint
		note      string
	}{
		{
			name:      "zero client ID",
			clientID:  0,
			cleanerID: 8,
			note:      "test",
		},
		{
			name:      "zero cleaner ID",
			clientID:  5,
			cleanerID: 0,
			note:      "test",
		},
		{
			name:      "blank note",
			clientID:  5,
			cleanerID: 8,
			note:      "   ",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			note, err := service.Create(
				context.Background(),
				test.clientID,
				test.cleanerID,
				CreateNoteRequest{
					Note: test.note,
				},
			)

			if note != nil {
				t.Fatalf(
					"expected nil note, got %+v",
					note,
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

func TestService_Create_RepositoryError(t *testing.T) {
	expectedErr := errors.New("create note failed")

	repo := &mockRepository{
		createFn: func(
			context.Context,
			*ClientCleanerNote,
		) error {
			return expectedErr
		},
	}

	service := NewService(repo)

	note, err := service.Create(
		context.Background(),
		5,
		8,
		CreateNoteRequest{
			Note: "Reliable",
		},
	)

	if note != nil {
		t.Fatalf(
			"expected nil note, got %+v",
			note,
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

func TestService_GetByCleanerID_Success(t *testing.T) {
	now := time.Now()

	repo := &mockRepository{
		getByCleanerIDFn: func(
			ctx context.Context,
			clientID uint,
			cleanerID uint,
		) ([]ClientCleanerNote, error) {
			if clientID != 5 {
				t.Fatalf(
					"expected client ID 5, got %d",
					clientID,
				)
			}

			if cleanerID != 8 {
				t.Fatalf(
					"expected cleaner ID 8, got %d",
					cleanerID,
				)
			}

			return []ClientCleanerNote{
				{
					ID:        1,
					ClientID:  5,
					CleanerID: 8,
					Note:      "Reliable",
					CreatedAt: now,
					UpdatedAt: now,
				},
				{
					ID:        2,
					ClientID:  5,
					CleanerID: 8,
					Note:      "Good communication",
					CreatedAt: now.Add(-time.Hour),
					UpdatedAt: now.Add(-time.Hour),
				},
			}, nil
		},
	}

	service := NewService(repo)

	notes, err := service.GetByCleanerID(
		context.Background(),
		5,
		8,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(notes) != 2 {
		t.Fatalf(
			"expected 2 notes, got %d",
			len(notes),
		)
	}

	if notes[0].Note != "Reliable" {
		t.Fatalf(
			"expected first note Reliable, got %q",
			notes[0].Note,
		)
	}

	if notes[1].Note != "Good communication" {
		t.Fatalf(
			"expected second note Good communication, got %q",
			notes[1].Note,
		)
	}
}

func TestService_GetByCleanerID_Empty(t *testing.T) {
	repo := &mockRepository{
		getByCleanerIDFn: func(
			context.Context,
			uint,
			uint,
		) ([]ClientCleanerNote, error) {
			return []ClientCleanerNote{}, nil
		},
	}

	service := NewService(repo)

	notes, err := service.GetByCleanerID(
		context.Background(),
		5,
		8,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(notes) != 0 {
		t.Fatalf(
			"expected no notes, got %d",
			len(notes),
		)
	}
}

func TestService_GetByCleanerID_InvalidInput(t *testing.T) {
	service := NewService(&mockRepository{})

	tests := []struct {
		name      string
		clientID  uint
		cleanerID uint
	}{
		{
			name:      "zero client ID",
			clientID:  0,
			cleanerID: 8,
		},
		{
			name:      "zero cleaner ID",
			clientID:  5,
			cleanerID: 0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			notes, err := service.GetByCleanerID(
				context.Background(),
				test.clientID,
				test.cleanerID,
			)

			if notes != nil {
				t.Fatalf(
					"expected nil notes, got %+v",
					notes,
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

func TestService_GetByCleanerID_RepositoryError(t *testing.T) {
	expectedErr := errors.New("list notes failed")

	repo := &mockRepository{
		getByCleanerIDFn: func(
			context.Context,
			uint,
			uint,
		) ([]ClientCleanerNote, error) {
			return nil, expectedErr
		},
	}

	service := NewService(repo)

	notes, err := service.GetByCleanerID(
		context.Background(),
		5,
		8,
	)

	if notes != nil {
		t.Fatalf(
			"expected nil notes, got %+v",
			notes,
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

func TestService_Update_Success(t *testing.T) {
	updateCalled := false

	repo := &mockRepository{
		updateFn: func(
			ctx context.Context,
			noteID uint,
			clientID uint,
			note string,
		) error {
			updateCalled = true

			if noteID != 12 {
				t.Fatalf(
					"expected note ID 12, got %d",
					noteID,
				)
			}

			if clientID != 5 {
				t.Fatalf(
					"expected client ID 5, got %d",
					clientID,
				)
			}

			if note != "Updated private note" {
				t.Fatalf(
					"expected trimmed note, got %q",
					note,
				)
			}

			return nil
		},
	}

	service := NewService(repo)

	err := service.Update(
		context.Background(),
		12,
		5,
		UpdateNoteRequest{
			Note: "  Updated private note  ",
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !updateCalled {
		t.Fatal("expected Update repository method to be called")
	}
}

func TestService_Update_InvalidInput(t *testing.T) {
	service := NewService(&mockRepository{})

	tests := []struct {
		name     string
		noteID   uint
		clientID uint
		note     string
	}{
		{
			name:     "zero note ID",
			noteID:   0,
			clientID: 5,
			note:     "test",
		},
		{
			name:     "zero client ID",
			noteID:   12,
			clientID: 0,
			note:     "test",
		},
		{
			name:     "blank note",
			noteID:   12,
			clientID: 5,
			note:     "   ",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := service.Update(
				context.Background(),
				test.noteID,
				test.clientID,
				UpdateNoteRequest{
					Note: test.note,
				},
			)

			if !errors.Is(err, ErrInvalidInput) {
				t.Fatalf(
					"expected ErrInvalidInput, got %v",
					err,
				)
			}
		})
	}
}

func TestService_Update_NotFound(t *testing.T) {
	repo := &mockRepository{
		updateFn: func(
			context.Context,
			uint,
			uint,
			string,
		) error {
			return ErrNoteNotFound
		},
	}

	service := NewService(repo)

	err := service.Update(
		context.Background(),
		12,
		5,
		UpdateNoteRequest{
			Note: "Updated note",
		},
	)

	if !errors.Is(err, ErrNoteNotFound) {
		t.Fatalf(
			"expected ErrNoteNotFound, got %v",
			err,
		)
	}
}

func TestService_Update_RepositoryError(t *testing.T) {
	expectedErr := errors.New("update note failed")

	repo := &mockRepository{
		updateFn: func(
			context.Context,
			uint,
			uint,
			string,
		) error {
			return expectedErr
		},
	}

	service := NewService(repo)

	err := service.Update(
		context.Background(),
		12,
		5,
		UpdateNoteRequest{
			Note: "Updated note",
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

func TestService_Delete_Success(t *testing.T) {
	deleteCalled := false

	repo := &mockRepository{
		deleteFn: func(
			ctx context.Context,
			noteID uint,
			clientID uint,
		) error {
			deleteCalled = true

			if noteID != 12 {
				t.Fatalf(
					"expected note ID 12, got %d",
					noteID,
				)
			}

			if clientID != 5 {
				t.Fatalf(
					"expected client ID 5, got %d",
					clientID,
				)
			}

			return nil
		},
	}

	service := NewService(repo)

	err := service.Delete(
		context.Background(),
		12,
		5,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !deleteCalled {
		t.Fatal("expected Delete repository method to be called")
	}
}

func TestService_Delete_InvalidInput(t *testing.T) {
	service := NewService(&mockRepository{})

	tests := []struct {
		name     string
		noteID   uint
		clientID uint
	}{
		{
			name:     "zero note ID",
			noteID:   0,
			clientID: 5,
		},
		{
			name:     "zero client ID",
			noteID:   12,
			clientID: 0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := service.Delete(
				context.Background(),
				test.noteID,
				test.clientID,
			)

			if !errors.Is(err, ErrInvalidInput) {
				t.Fatalf(
					"expected ErrInvalidInput, got %v",
					err,
				)
			}
		})
	}
}

func TestService_Delete_NotFound(t *testing.T) {
	repo := &mockRepository{
		deleteFn: func(
			context.Context,
			uint,
			uint,
		) error {
			return ErrNoteNotFound
		},
	}

	service := NewService(repo)

	err := service.Delete(
		context.Background(),
		12,
		5,
	)

	if !errors.Is(err, ErrNoteNotFound) {
		t.Fatalf(
			"expected ErrNoteNotFound, got %v",
			err,
		)
	}
}

func TestService_Delete_RepositoryError(t *testing.T) {
	expectedErr := errors.New("delete note failed")

	repo := &mockRepository{
		deleteFn: func(
			context.Context,
			uint,
			uint,
		) error {
			return expectedErr
		},
	}

	service := NewService(repo)

	err := service.Delete(
		context.Background(),
		12,
		5,
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected %v, got %v",
			expectedErr,
			err,
		)
	}
}
