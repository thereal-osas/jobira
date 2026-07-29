package chat

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"time"

	"github.com/lib/pq"
)

type SQLRepository struct {
	db *sql.DB
}

func NewSQLRepository(db *sql.DB) *SQLRepository {
	return &SQLRepository{
		db: db,
	}
}

func (r *SQLRepository) CreateConversation(ctx context.Context, conversation *Conversation) error {
	query := `
		INSERT INTO conversations (
			booking_id,
			client_id,
			cleaner_id,
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
		conversation.BookingID,
		conversation.ClientID,
		conversation.CleanerID,
		conversation.Status,
		now,
	).Scan(
		&conversation.ID,
		&conversation.CreatedAt,
		&conversation.UpdatedAt,
	)

	if err != nil {
		var pqErr *pq.Error

		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return ErrConversationExists
		}

		return err
	}

	return nil
}

func (r *SQLRepository) GetConversationByID(ctx context.Context, conversationID uint) (*Conversation, error) {
	query := `
		SELECT 
			id, 
			booking_id, 
			client_id,
			cleaner_id,
			status, 
			last_message_at,
			created_at,
			updated_at
		FROM conversations
		WHERE id = $1	
	`

	conversation := &Conversation{}

	err := r.db.QueryRowContext(
		ctx,
		query,
		conversationID,
	).Scan(
		&conversation.ID,
		&conversation.BookingID,
		&conversation.ClientID,
		&conversation.CleanerID,
		&conversation.Status,
		&conversation.LastMessagesAt,
		&conversation.CreatedAt,
		&conversation.UpdatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrConversationNotFound
	}

	if err != nil {
		return nil, err
	}

	return conversation, nil
}

func (r *SQLRepository) GetConversationByBookingID(ctx context.Context, bookingID uint) (*Conversation, error) {
	query := `
		SELECT
			id, 
			booking_id,
			client_id,
			cleaner_id, 
			status, 
			last_message_at, 
			created_at, 
			updated_at 
		FROM conversations 
		WHERE booking_id = $1	
	`

	conversation := &Conversation{}

	err := r.db.QueryRowContext(
		ctx,
		query,
		bookingID,
	).Scan(
		&conversation.ID,
		&conversation.BookingID,
		&conversation.ClientID,
		&conversation.CleanerID,
		&conversation.Status,
		&conversation.LastMessagesAt,
		&conversation.CreatedAt,
		&conversation.UpdatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrConversationNotFound
	}

	if err != nil {
		return nil, err
	}

	return conversation, nil
}
func (r *SQLRepository) ListConversationByUserID(
	ctx context.Context,
	userID uint,
) ([]ConversationSummary, error) {
	query := `
		SELECT
			c.id,
			c.booking_id,
			c.client_id,
			c.cleaner_id,
			c.status,

			CASE
				WHEN c.client_id = $1 THEN c.cleaner_id
				ELSE c.client_id
			END AS other_user_id,

			CASE
				WHEN c.client_id = $1 THEN cleaner.full_name
				ELSE client.full_name
			END AS other_user_name,

			last_message.content,
			last_message.sender_id,
			c.last_message_at,
			COUNT(unread.id) AS unread_count,
			c.created_at

		FROM conversations c

		INNER JOIN users client
			ON client.id = c.client_id

		INNER JOIN users cleaner
			ON cleaner.id = c.cleaner_id

		LEFT JOIN LATERAL (
			SELECT
				m.content,
				m.sender_id
			FROM messages m
			WHERE m.conversation_id = c.id
			ORDER BY m.created_at DESC, m.id DESC
			LIMIT 1
		) AS last_message ON true

		LEFT JOIN messages unread
			ON unread.conversation_id = c.id
			AND unread.sender_id <> $1
			AND unread.is_read = false

		WHERE c.client_id = $1
			OR c.cleaner_id = $1

		GROUP BY
			c.id,
			c.booking_id,
			c.client_id,
			c.cleaner_id,
			c.status,
			client.full_name,
			cleaner.full_name,
			last_message.content,
			last_message.sender_id,
			c.last_message_at,
			c.created_at

		ORDER BY
			c.last_message_at DESC NULLS LAST,
			c.created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		log.Printf("ListConversationByUserID SQL error: %v", err)
		return nil, err
	}
	defer rows.Close()

	conversations := make([]ConversationSummary, 0)

	for rows.Next() {
		var conversation ConversationSummary

		err := rows.Scan(
			&conversation.ID,
			&conversation.BookingID,
			&conversation.ClientID,
			&conversation.CleanerID,
			&conversation.Status,
			&conversation.OtherUserID,
			&conversation.OtherUserName,
			&conversation.LastMessage,
			&conversation.LastMessageSender,
			&conversation.LastMessageAt,
			&conversation.UnreadCount,
			&conversation.CreatedAt,
		)
		if err != nil {
			log.Printf("ListConversationByUserID scan error: %v", err)
			return nil, err
		}

		conversations = append(conversations, conversation)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return conversations, nil
}

func (r *SQLRepository) CreateMessage(ctx context.Context, message *Message) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	committed := false

	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	now := time.Now()

	insertMessageQuery := `
		INSERT INTO messages (
			conversation_id,
			sender_id,
			message_type,
			content,
			is_read,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, false, $5, $5)
		RETURNING
			id,
			is_read,
			read_at,
			created_at,
			updated_at
	`

	err = tx.QueryRowContext(
		ctx,
		insertMessageQuery,
		message.ConversationID,
		message.SenderID,
		message.MessageType,
		message.Content,
		now,
	).Scan(
		&message.ID,
		&message.IsRead,
		&message.ReadAt,
		&message.CreatedAt,
		&message.UpdatedAt,
	)
	if err != nil {
		log.Printf("CreateMessage SQL error: %v", err)
		return err
	}
	updateConversationQuery := `
		UPDATE conversations
		SET
			last_message_at = $1,
			updated_at = $1
		WHERE id = $2
	`

	result, err := tx.ExecContext(
		ctx,
		updateConversationQuery,
		now,
		message.ConversationID,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrConversationNotFound
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	committed = true

	return nil
}
func (r *SQLRepository) ListMessage(
	ctx context.Context,
	conversationID uint,
	limit int,
	offset int,
) ([]Message, error) {
	query := `
		SELECT 
			id,
			conversation_id,
			sender_id,
			message_type,
			content,
			is_read,
			read_at,
			created_at,
			updated_at
		FROM messages
		WHERE conversation_id = $1
		ORDER BY created_at ASC, id ASC
		LIMIT $2
		OFFSET $3
	`

	rows, err := r.db.QueryContext(
		ctx,
		query,
		conversationID,
		limit,
		offset,
	)
	if err != nil {
		log.Printf("ListMessage SQL error: %v", err)
		return nil, err
	}
	defer rows.Close()

	messages := make([]Message, 0)

	for rows.Next() {
		var message Message

		err := rows.Scan(
			&message.ID,
			&message.ConversationID,
			&message.SenderID,
			&message.MessageType,
			&message.Content,
			&message.IsRead,
			&message.ReadAt,
			&message.CreatedAt,
			&message.UpdatedAt,
		)
		if err != nil {
			log.Printf("ListMessage scan error: %v", err)
			return nil, err
		}

		messages = append(messages, message)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return messages, nil
}

func (r *SQLRepository) MarkConversationAsRead(ctx context.Context, conversationID uint, userID uint) error {
	query := `
		UPDATE messages
		SET 
			is_read = true, 
			read_at = NOW(), 
			updated_at = NOW()
		WHERE conversation_id = $1
		AND sender_id <> $2
		AND is_read = false	
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		conversationID,
		userID,
	)
	if err != nil {
		return err
	}

	return nil
}
