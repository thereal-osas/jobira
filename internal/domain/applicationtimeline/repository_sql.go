package applicationtimeline

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
	return &SQLRepository{
		db: db,
	}
}

func (r *SQLRepository) Create(ctx context.Context, event *Event) error {
	if event == nil {
		return ErrInvalidInput
	}

	query := `
		INSERT INTO application_timeline (
			application_id,
			status,
			actor_user_id,
			note,
			created_at
		)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at	
	`

	if event.CreatedAt.IsZero() {
		event.CreatedAt = time.Now()
	}

	err := r.db.QueryRowContext(
		ctx,
		query,
		event.ApplicationID,
		event.Status,
		event.ActorUserID,
		event.Note,
		event.CreatedAt,
	).Scan(
		&event.ID,
		&event.CreatedAt,
	)

	if err != nil {
		return err
	}

	return nil
}

func (r *SQLRepository) ListByApplicationID(ctx context.Context, applicationID uint) ([]Event, error) {
	query := `
		SELECT
			id, 
			application_id,
			status,
			actor_user_id,
			note,
			created_at
		FROM application_timeline
		WHERE application_id = $1
		ORDER BY created_at ASC, id ASC	
	`
	rows, err := r.db.QueryContext(
		ctx,
		query,
		applicationID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	events := make([]Event, 0)

	for rows.Next() {
		var (
			event       Event
			ActorUserID sql.NullInt64
		)

		if err := rows.Scan(
			&event.ID,
			&event.ApplicationID,
			&event.Status,
			&ActorUserID,
			&event.Note,
			&event.CreatedAt,
		); err != nil {
			return nil, err
		}

		if ActorUserID.Valid {
			value := uint(ActorUserID.Int64)
			event.ActorUserID = &value
		}

		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(events) == 0 {
		return nil, ErrTimelineNotFound
	}

	return events, nil
}
func (r *SQLRepository) GetCurrentStatus(ctx context.Context, applicationID uint) (string, error) {
	query := `
		SELECT status 
		FROM applications
		WHERE id = $1
		LIMIT 1
	`

	var status string

	err := r.db.QueryRowContext(
		ctx,
		query,
		applicationID,
	).Scan(&status)

	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrApplicationNotFound
	}

	if err != nil {
		return "", err
	}

	return status, nil
}

func (r *SQLRepository) GetApplicationCleanerID(ctx context.Context, applicationID uint) (uint, error) {
	query := `
		SELECT cleaner_id
		FROM applications
		WHERE id = $1
		LIMITE 1
	`

	var cleanerID uint

	err := r.db.QueryRowContext(
		ctx,
		query,
		applicationID,
	).Scan(&cleanerID)

	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrApplicationNotFound
	}

	if err != nil {
		return 0, err
	}

	return cleanerID, nil
}

func (r *SQLRepository) GetApplicationClientID(ctx context.Context, applicationID uint) (uint, error) {
	query := `
		SELECT j.client_id
		FROM applications a 
		INNER JOIN jobs j
			ON j.id = a.job_id 
		WHERE a.id = $1 
		LIMIT 1	
	`

	var clientID uint

	err := r.db.QueryRowContext(
		ctx,
		query,
		applicationID,
	).Scan(&clientID)

	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrApplicationNotFound
	}

	if err != nil {
		return 0, err
	}

	return clientID, nil
}
