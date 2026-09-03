package wokrproof

import (
	"context"
	"database/sql"
	"errors"
)

type SQLRepository struct {
	db *sql.DB
}

func NewSQLRepository(
	db *sql.DB,
) *SQLRepository {
	return &SQLRepository{
		db: db,
	}
}

func (r *SQLRepository) GetBooking(ctx context.Context, bookingID uint) (*BookingSnapshot, error) {
	query := `
		SELECT
			id,
			job_id,
			client_id,
			cleaner_id,
			status,
		FROM bookings
		WHERE id = $1 
		LIMIT 1 	
	`

	var booking BookingSnapshot

	err := r.db.QueryRowContext(
		ctx,
		query,
		bookingID,
	).Scan(
		&booking.ID,
		&booking.JobID,
		&booking.ClientID,
		&booking.CleanerID,
		&booking.Status,
	)

	if errors.Is(
		err,
		sql.ErrNoRows,
	) {
		return nil,
			ErrBookingNotFound
	}

	if err != nil {
		return nil, err
	}

	return &booking, nil
}

func (r *SQLRepository) Create(ctx context.Context, proof *WorkProof) error {
	query := `
		INSERT INTO work_proofs (
			booking_id,
			cleaner_id,
			proof_type,
			photo_url,
			caption,
			created_at
		)
		VALUES (
			$1, $2, $3, $4, $5, NOW()
		)	
		RETURNING 
			id,
			created_at	
	`

	return r.db.QueryRowContext(
		ctx,
		query,
		proof.BookingID,
		proof.CleanerID,
		proof.ProofType,
		proof.PhotoURL,
		proof.Caption,
	).Scan(
		&proof.ID,
		&proof.CreatedAt,
	)
}

func (r *SQLRepository) GetByID(ctx context.Context, proofID uint) (*WorkProof, error) {
	query := `
		SELECT 
			id, 
			booking_id,
			cleaner_id,
			proof_type,
			photo_url,
			caption,
			created_at
		FROM work_proofs
		WHERE id = $1 
		LIMIT 1			
	`

	var proof WorkProof

	err := r.db.QueryRowContext(
		ctx,
		query,
		proofID,
	).Scan(
		&proof.ID,
		&proof.BookingID,
		&proof.CleanerID,
		&proof.ProofType,
		&proof.PhotoURL,
		&proof.Caption,
		&proof.CreatedAt,
	)

	if errors.Is(
		err,
		sql.ErrNoRows,
	) {
		return nil,
			ErrProofNotFound
	}

	if err != nil {
		return nil, err
	}

	return &proof, nil
}

func (r *SQLRepository) ListByBookingID(ctx context.Context, bookingID uint) ([]WorkProof, error) {
	query := `
		SELECT 
			id,
			booking_id,
			cleaner_id,
			proof_type,
			photo_url,
			caption,
			created_at
		FROM work_proofs
		WHERE booking_id = $1 
		ORDER BY created_at ASC 	
	`

	rows, err := r.db.QueryContext(
		ctx,
		query,
		bookingID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	proofs := make(
		[]WorkProof,
		0,
	)

	for rows.Next() {
		var proof WorkProof

		if err := rows.Scan(
			&proof.ID,
			&proof.BookingID,
			&proof.CleanerID,
			&proof.ProofType,
			&proof.PhotoURL,
			&proof.Caption,
			&proof.CreatedAt,
		); err != nil {
			return nil, err
		}

		proofs = append(
			proofs,
			proof,
		)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return proofs, nil

}

func (r *SQLRepository) CountByBookingAndType(ctx context.Context, bookingID uint, proofType ProofType) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM work_proofs
		WHERE booking_id = $1 
		AND proof_type = $2
	`

	var count int

	err := r.db.QueryRowContext(
		ctx,
		query,
		bookingID,
		proofType,
	).Scan(
		&count,
	)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *SQLRepository) Delete(ctx context.Context, proofID uint, cleanerID uint) error {
	query := `
		DELETE FROM work_proofs
		WHERE id = $1 
		AND cleaner_id = $2	
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		proofID,
		cleanerID,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrProofNotFound
	}

	return nil

}

func (r *SQLRepository) GetVerifiedWorkSummary(ctx context.Context, cleanerID uint) (*VerifiedWorkSummary, error) {
	query := `
		SELECT 
			$1,

			COUNT(
				DISTINCT CASE
					WHEN EXISTS (
						SELECT 1 
						FROM work_proofs before_proof
						WHERE before_proof.booking_id = wp.booking_id
						AND before_proof.cleaner_id = $1 
						AND before_proof.proof_type = 'before'
					)
					AND EXISTS (
						SELECT 1 
						FROM work_proofs after_proof
						WHERE after_proof.booking_id = wp.booking_id 
						AND after_proof.cleaner_id = $1
						AND after_proof.proof_type = 'after'
					)	
					THEN wp.booking_id 
				END 		
			),

			COUNT(*) FILTER (
				WHERE wp.proof_type = 'before'
			),

			COUNT(*) FILTER (
				WHERE wp.proof_type = 'before'
			),

			COUNT(*)FILTER (
				WHERE wp.proof_type = 'after'
			),

			COUNT(*)

		FROM work_proofs wp
		WHERE wp.cleaner_id = $1	
	`

	var summary VerifiedWorkSummary

	err := r.db.QueryRowContext(
		ctx,
		query,
		cleanerID,
	).Scan(
		&summary.CleanerID,
		&summary.VerifiedJobs,
		&summary.BeforePhotos,
		&summary.AfterPhotos,
		&summary.TotalPhotos,
	)
	if err != nil {
		return nil, err
	}

	return &summary, nil
}
