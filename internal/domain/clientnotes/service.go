package clientnotes

import (
	"context"
	"strings"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, clientID uint, cleanerID uint, req CreateNoteRequest) (*ClientCleanerNote, error) {
	req.Note = strings.TrimSpace(req.Note)

	if clientID == 0 || cleanerID == 0 || req.Note == "" {
		return nil, ErrInvalidInput
	}

	note := &ClientCleanerNote{
		ClientID: clientID,
		CleanerID: cleanerID,
		Note: req.Note,
	}

	if err := s.repo.Create(ctx, note); err != nil { 
		return nil, err
	}

	return note, nil
}

func (s *Service) GetByCleanerID(ctx context.Context, clientID uint, cleanerID uint) ([] ClientCleanerNote, error) {
	if clientID == 0 || cleanerID == 0 {
		return nil, ErrInvalidInput
	}

	return s.repo.GetByCleanerID(ctx, clientID, cleanerID)
}

func (s *Service) Update(ctx context.Context, noteID uint, clientID uint, req UpdateNoteRequest) error {
	req.Note = strings.TrimSpace(req.Note)

	if noteID == 0 || clientID == 0 || req.Note == "" {
		return ErrInvalidInput
	}

	return s.repo.Update(ctx, noteID, clientID, req.Note)
}

func (s *Service) Delete(ctx context.Context, noteID uint, clientID uint) error {
	if noteID == 0 || clientID == 0 {
		return ErrInvalidInput
	}

	return s.repo.Delete(ctx, noteID, clientID)
}