package jobinvitations

import (
	"context"
	"database/sql"
	"time"
)

type SQLRepository struct {
	db *sql.DB
}

func NewSQLRepository(db *sql.DB) *SQLRepository {
	return &SQLRepository{
		db: db,
	}
}

func (r *SQLRepository) Create(ctx context.Context, invitation *JobInvitation) error {
	query := `
		INSERT INTO job_invitations (
			job_id,
			client_id,
			cleaner_id,
			status,
			message,
			created_at,
			updated_at 
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at 	
	`

	now := time.Now()

	invitation.Status = "sent"

	return r.db.QueryRowContext(
		ctx, 
		query,
		invitation.JobID,
		invitation.ClientID,
		invitation.CleanerID,
		invitation.Status,
		invitation.Message,
		now,
		now,
	).Scan(
		&invitation.ID,
		&invitation.CreatedAt,
		&invitation.UpdatedAt,
	)
}

func (r *SQLRepository) Exists(ctx context.Context, jobID uint, cleanerID uint) (bool, error) {
	query := `
		SELECT EXISTS (
		SELECT 1
		FROM job_invitations
		WHERE job_id = $1 
		AND cleaner_id = $2 
		)
	`

	var exists bool

	err := r.db.QueryRowContext(
		ctx,
		query,
		jobID,
		cleanerID,
	).Scan(&exists)

	return exists, err
}

func (r *SQLRepository) GetJobClientID(ctx context.Context, jobID uint) (uint, error) {
	query := `
		SELECT client_id
		FROM jobs 
		WHERE id = $1 
		LIMIT 1
	`

	var clientID uint

	err := r.db.QueryRowContext(ctx, query, jobID).Scan(&clientID)

	return clientID, err
}

func (r *SQLRepository) ListSentByClientID(ctx context.Context, clientID uint) ([]JobInvitation, error) {
	query := `
		SELECT 
			id, 
			job_id,
			client_id,
			cleaner_id,
			status,
			message,
			created_at, 
			updated_at 
		FROM job_invitations 
		WHERE client_id = $1 
		ORDER BY created_at DESC 	
	`

	rows, err := r.db.QueryContext(ctx, query, clientID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var invitations []JobInvitation

	for rows.Next() {
		var invitation JobInvitation

		err := rows.Scan(
			&invitation.ID,
			&invitation.JobID,
			&invitation.ClientID,
			&invitation.CleanerID,
			&invitation.Status,
			&invitation.Message,
			&invitation.CreatedAt,
			&invitation.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		invitations = append(invitations, invitation)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return invitations, nil
}

func (r *SQLRepository) ListReceivedByCleanerID(ctx context.Context, cleanerID uint) ([]JobInvitation, error) {
	query := `
		SELECT 
			id, 
			job_id,
			client_id, 
			cleaner_id,
			status, 
			message, 
			created_at,
			updated_at 
		FROM job_invitations 
		WHERE cleaner_id = $1 
		ORDER BY created_at DESC 	
	`

	rows, err := r.db.QueryContext(ctx, query, cleanerID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var invitations []JobInvitation

	for rows.Next() {
		var invitation JobInvitation

		err := rows.Scan(
			&invitation.ID,
			&invitation.JobID,
			&invitation.ClientID,
			&invitation.CleanerID,
			&invitation.Status,
			&invitation.Message,
			&invitation.CreatedAt,
			&invitation.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		invitations = append(invitations, invitation)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return invitations, nil
}