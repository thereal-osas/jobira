package verifications

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type SQLRepository struct {
	db *sql.DB
}

func NewSQLRepository(db *sql.DB) *SQLRepository {
	return &SQLRepository{db: db}
}

func (r *SQLRepository) Create(ctx context.Context, request *VerificationRequest) error {
	query := `
	 	INSERT INTO verification_requests (
			user_id, 
			verification_type,
			document_url,
			status,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $5)
		RETURNING id, created_at, updated_at
	`

	now := time.Now()

	err := r.db.QueryRowContext(
		ctx,
		query,
		request.UserID,
		request.VerificationType,
		request.DocumentURL,
		request.Status,
		now,
	).Scan(
		&request.ID,
		&request.CreatedAt,
		&request.UpdatedAt,
	)

	if err != nil {
		return err
	}

	return nil
}

func (r *SQLRepository) GetByID(ctx context.Context, id uint) (*VerificationRequest, error) {
	query := `
		SELECT 
			id,
			user_id,
			verification_type,
			document_url,
			status,
			COALESCE(admin_notes, ''),
			reviewed_by,
			reviewed_at,
			created_at,
			updated_at
		FROM verification_requests
		WHERE id = $1 
		LIMIT 1 	
	`

	var request VerificationRequest
	var reviewedBy sql.NullInt64
	var reviewedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&request.ID,
		&request.UserID,
		&request.VerificationType,
		&request.DocumentURL,
		&request.Status,
		&request.AdminNotes,
		&reviewedBy,
		&reviewedAt,
		&request.CreatedAt,
		&request.UpdatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrVerificationNotFound
	}

	if err != nil {
		return nil, err
	}

	if reviewedBy.Valid {
		value := uint(reviewedBy.Int64)
		request.ReviewedBy = &value
	}

	if reviewedAt.Valid {
		request.ReviewedAt = &reviewedAt.Time
	}

	return &request, nil
}

func (r *SQLRepository) ListByUserID(ctx context.Context, userID uint) ([]VerificationRequest, error) {
	query := `
		SELECT
			id, 
			user_id,
			verification_type,
			document_url,
			status, 
			COALESCE(admin_notes, ''),
			reviewed_by,
			reviewed_at,
			created_at,
			updated_at
		FROM verification_requests
		WHERE user_id = $1
		ORDER BY created_at DESC	
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []VerificationRequest

	for rows.Next() {
		var request VerificationRequest
		var reviewedBy sql.NullInt64
		var reviewedAt sql.NullTime

		err := rows.Scan(
			&request.ID,
			&request.UserID,
			&request.VerificationType,
			&request.DocumentURL,
			&request.Status,
			&request.AdminNotes,
			&reviewedBy,
			&reviewedAt,
			&request.CreatedAt,
			&request.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if reviewedBy.Valid {
			value := uint(reviewedBy.Int64)
			request.ReviewedBy = &value
		}

		if reviewedAt.Valid {
			request.ReviewedAt = &reviewedAt.Time
		}

		requests = append(requests, request)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return requests, nil

}

func (r *SQLRepository) ListAll(ctx context.Context) ([]VerificationRequest, error) {
	query := `
		SELECT 
			id,
			user_id,
			verification_type,
			document_url,
			status,
			COALESCE(admin_notes, ''),
			reviewed_by,
			reviewed_at,
			created_at,
			updated_at
		FROM verification_requests
		ORDER BY created_at DESC	
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []VerificationRequest

	for rows.Next() {
		var request VerificationRequest
		var reviewedBy sql.NullInt64
		var reviewedAt sql.NullTime

		err := rows.Scan(
			&request.ID,
			&request.UserID,
			&request.VerificationType,
			&request.DocumentURL,
			&request.Status,
			&request.AdminNotes,
			&reviewedBy,
			&reviewedAt,
			&request.CreatedAt,
			&request.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if reviewedBy.Valid {
			value := uint(reviewedBy.Int64)
			request.ReviewedBy = &value
		}

		if reviewedAt.Valid {
			request.ReviewedAt = &reviewedAt.Time
		}

		requests = append(requests, request)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return requests, nil
}

func (r *SQLRepository) ListPending(ctx context.Context) ([]VerificationRequest, error) {
	query := `
		SELECT 
			id,
			user_id,
			verification_type,
			document_url,
			status,
			COALESCE(admin_notes, ''),
			reviewed_by,
			reviewed_at,
			created_at,
			updated_at
		FROM verification_requests
		WHERE status = 'pending'
		ORDER BY created_at DESC	
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []VerificationRequest

	for rows.Next() {
		var request VerificationRequest
		var reviewedBy sql.NullInt64
		var reviewedAt sql.NullTime

		err := rows.Scan(
			&request.ID,
			&request.UserID,
			&request.VerificationType,
			&request.DocumentURL,
			&request.Status,
			&request.AdminNotes,
			&reviewedBy,
			&reviewedAt,
			&request.CreatedAt,
			&request.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if reviewedBy.Valid {
			value := uint(reviewedBy.Int64)
			request.ReviewedBy = &value
		}

		if reviewedAt.Valid {
			request.ReviewedAt = &reviewedAt.Time
		}

		requests = append(requests, request)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return requests, nil

}

func (r *SQLRepository) Review(ctx context.Context, requestID uint, status string, adminNotes string, adminID uint) error {
	query := `
		UPDATE verification_requests
		SET
			status = $1,
			admin_notes = $2,
			reviewed_by = $3,
			reviewed_at = $4,
			updated_at = $4
		WHERE id = $5	
	`

	now := time.Now()

	result, err := r.db.ExecContext(
		ctx,
		query,
		status,
		adminNotes,
		adminID,
		now,
		requestID,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrVerificationNotFound
	}

	return nil
}
