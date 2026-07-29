package messages

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

func (r *SQLRepository) Create(ctx context.Context, message *Message) error {
	query := `
	INSERT INTO messages (
		job_id,
		sender_id,
		receiver_id,
		content, 
		is_read, 
		created_at,
		updated_at
	)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
		`
	
	now := time.Now()
	
	message.IsRead = false
	message.CreatedAt = now
	message.UpdatedAt = now

	return r.db.QueryRowContext(
		ctx, 
		query, 
		message.JobID,
		message.SenderID,
		message.ReceiverID,
		message.Content,
		message.IsRead,
		message.CreatedAt,
		message.UpdatedAt,
	).Scan(&message.ID)
}

func (r *SQLRepository) ListByJobID(ctx context.Context, jobID uint) ([]Message, error) {
	query := `
		SELECT
			id,
			job_id,
			sender_id,
			receiver_id,
			content,
			is_read, 
			created_at, 
			updated_at
		FROM messages
		WHERE job_id = $1
		ORDER BY created_at ASC 		
	`

	rows, err := r.db.QueryContext(ctx, query, jobID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var messages []Message

	for rows.Next() {
		var message Message

		err := rows.Scan(
			&message.ID,
			&message.JobID,
			&message.SenderID,
			&message.ReceiverID,
			&message.Content,
			&message.IsRead,
			&message.CreatedAt,
			&message.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		messages = append(messages, message)
	}

	return messages, nil
}

func (r *SQLRepository) IsJobParticipant(ctx context.Context, jobID uint, userID uint) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT 1
			FROM jobs j 
			LEFT JOIN applications a ON a.job_id = j.id 
			WHERE j.id = $1 
			AND (
			
				j.client_id = $2
				OR a.cleaner_id = $2 
			)

		)
	`

	var exists bool

	err := r.db.QueryRowContext(ctx, query, jobID, userID).Scan(&exists)

	return exists, err
}