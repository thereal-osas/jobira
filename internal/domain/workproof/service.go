package wokrproof

import (
	"context"
	"strings"
	"time"
)

const maxPhotosPerType = 10

type Service struct {
	repo Repository
}

func NewService(
	repo Repository,
) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) Create(ctx context.Context, bookingID uint, cleanerID uint, req CreateWorkProofRequest) (*WorkProof, error) {
	if bookingID == 0 || cleanerID == 0 {
		return nil, ErrInvalidInput
	}

	req.ProofType =
		strings.ToLower(
			strings.TrimSpace(
				req.ProofType,
			),
		)

	req.PhotoURL =
		strings.TrimSpace(
			req.PhotoURL,
		)

	req.Caption =
		strings.TrimSpace(
			req.Caption,
		)

	if req.ProofType !=
		string(ProofTypeBefore) &&
		req.ProofType !=
			string(ProofTypeAfter) {
		return nil,
			ErrInvalidProofType
	}

	if req.PhotoURL == "" {
		return nil,
			ErrInvalidInput
	}

	if len(req.Caption) > 500 {
		return nil,
			ErrInvalidInput
	}

	booking, err :=
		s.repo.GetBooking(
			ctx,
			bookingID,
		)
	if err != nil {
		return nil, err
	}

	if booking == nil {
		return nil,
			ErrBookingNotFound
	}

	if booking.CleanerID !=
		cleanerID {
		return nil, ErrForbidden
	}

	if !bookingAllowsProof(
		booking.Status,
	) {
		return nil,
			ErrBookingNotEligible
	}

	ProofType :=
		ProofType(
			req.ProofType,
		)

	count, err :=
		s.repo.CountByBookingAndType(
			ctx,
			bookingID,
			ProofType,
		)

	if err != nil {
		return nil, err
	}

	if count >=
		maxPhotosPerType {
		return nil,
			ErrProofLimitReached
	}

	proof :=
		&WorkProof{
			BookingID: bookingID,

			CleanerID: cleanerID,

			ProofType: ProofType,

			PhotoURL: req.PhotoURL,
			Caption:  req.Caption,
		}

	if err := s.repo.Create(
		ctx,
		proof,
	); err != nil {
		return nil, err
	}

	return proof, nil
}

func (s *Service) GetBookingProof(ctx context.Context, bookingID uint, userID uint) (*BookingProofSummary, error) {
	if bookingID == 0 || userID == 0 {
		return nil, ErrInvalidInput
	}

	booking, err := s.repo.GetBooking(
		ctx,
		bookingID,
	)
	if err != nil {
		return nil, err
	}

	if booking == nil {
		return nil,
			ErrBookingNotFound
	}

	if booking.ClientID != userID &&
		booking.CleanerID != userID {
		return nil, ErrForbidden
	}

	proofs, err := s.repo.ListByBookingID(
		ctx,
		bookingID,
	)
	if err != nil {
		return nil, err
	}

	beforePhotos := make([]WorkProof, 0)

	afterPhotos := make([]WorkProof, 0)

	var LastUploadedAt = (*time.Time)(nil)

	for _, proof := range proofs {
		switch proof.ProofType {
		case ProofTypeBefore:
			beforePhotos =
				append(
					beforePhotos,
					proof,
				)

		case ProofTypeAfter:
			afterPhotos =
				append(
					afterPhotos,
					proof,
				)
		}

		if LastUploadedAt == nil ||
			proof.CreatedAt.After(
				*LastUploadedAt,
			) {
			uploadedAt :=
				proof.CreatedAt

			LastUploadedAt =
				&uploadedAt
		}
	}

	hasBefore :=
		len(beforePhotos) > 0

	hasAfter :=
		len(afterPhotos) > 0

	return &BookingProofSummary{
		BookingID: bookingID,

		CleanerID: booking.CleanerID,

		BeforePhotos: beforePhotos,

		AfterPhotos: afterPhotos,

		BeforeCount: len(beforePhotos),

		AfterCount: len(afterPhotos),

		HasBeforeProof: hasBefore,

		HasAfterProof: hasAfter,

		IsVerifiedWork: hasBefore &&
			hasAfter,

		LastUploadedAt: LastUploadedAt,
	}, nil
}

func (s *Service) Delete(ctx context.Context, proofID uint, cleanerID uint) error {
	if proofID == 0 || cleanerID == 0 {
		return ErrInvalidInput
	}

	proof, err := s.repo.GetByID(
		ctx,
		proofID,
	)
	if err != nil {
		return err
	}

	if proof.CleanerID != cleanerID {
		return ErrForbidden
	}

	return s.repo.Delete(ctx, proofID, cleanerID)
}

func (s *Service) GetVerifiedWork(ctx context.Context, cleanerID uint) (*VerifiedWorkSummary, error) {
	if cleanerID == 0 {
		return nil, ErrInvalidInput
	}

	return s.repo.GetVerifiedWorkSummary(
		ctx,
		cleanerID,
	)
}

func bookingAllowsProof(
	status string,
) bool {
	switch strings.ToLower(
		strings.TrimSpace(
			status,
		),
	) {
	case "confirmed",
		"in_progress",
		"completed",
		"closed":
		return true

	default:
		return false
	}
}
