package notifications

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

func (r *SQLRepository) Create(ctx context.Context, notification *Notification) error {
	query := `
		INSERT INTO notifications (
			user_id,
			title,
			message,
			type,
			is_read,
			created_at,
			updated_at 
		)
			VALUES ($1,$2,$3,$4,$5,$6,$7)
			RETURNING id 
	`

	now := time.Now()

	notification.IsRead = false
	notification.CreatedAt = now
	notification.UpdatedAt = now

	return r.db.QueryRowContext(
		ctx,
		query,
		notification.UserID,
		notification.Title,
		notification.Message,
		notification.Type,
		notification.IsRead,
		notification.CreatedAt,
		notification.UpdatedAt,
	).Scan(&notification.ID)
}

func (r *SQLRepository) ListByUserID(ctx context.Context, userID uint) ([]Notification, error) {
	query := `
		SELECT
			id, 
			user_id,
			title,
			message,
			type,
			is_read, 
			created_at,
			updated_at 
		FROM notifications 
		WHERE user_id = $1 
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var notifications []Notification

	for rows.Next() {
		var notification Notification

		err := rows.Scan(
			&notification.ID,
			&notification.UserID,
			&notification.Title,
			&notification.Message,
			&notification.Type,
			&notification.IsRead,
			&notification.CreatedAt,
			&notification.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		notifications = append(notifications, notification)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return notifications, nil
}

func (r *SQLRepository) ListUnreadByUserID(ctx context.Context, userID uint) ([]Notification, error) {
	query := `
		SELECT
			id, 
			user_id,
			title,
			message,
			type,
			is_read, 
			created_at,
			updated_at 
		FROM notifications 
		WHERE user_id = $1 
		AND is_read = false
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var notifications []Notification

	for rows.Next() {
		var notification Notification

		err := rows.Scan(
			&notification.ID,
			&notification.UserID,
			&notification.Title,
			&notification.Message,
			&notification.Type,
			&notification.IsRead,
			&notification.CreatedAt,
			&notification.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		notifications = append(notifications, notification)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return notifications, nil
}

func (r *SQLRepository) MarkAsRead(ctx context.Context, id uint, userID uint) error {
	query := `
		UPDATE notifications
		SET 	
			is_read = true, 
			updated_at = $1
		WHERE id = $2 
		AND user_id = $3 	
	`

	result, err := r.db.ExecContext(ctx, query, time.Now(), id, userID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrNotificationNotFound
	}

	return nil
}

func (r *SQLRepository) CountUnreadByUserID(ctx context.Context, userID uint) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM notifications
		WHERE user_id = $1 
		AND is_read = false 
	`

	var count int

	err := r.db.QueryRowContext(ctx, query, userID).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *SQLRepository) MarkAllAsRead(ctx context.Context, userID uint) error {
	query := `
		UPDATE notifications
		SET
			is_read = true, 
			updated_at = $1 
		WHERE user_id = $2 
		AND is_read = false 	
	`

	_, err := r.db.ExecContext(ctx, query, time.Now(), userID)

	return err
}

