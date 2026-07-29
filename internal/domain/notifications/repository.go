package notifications

import "context"

type Repository interface {
	Create(ctx context.Context, notification *Notification) error
	ListByUserID(ctx context.Context, userID uint) ([]Notification, error)
	ListUnreadByUserID(ctx context.Context, userID uint) ([]Notification, error)
	MarkAsRead(ctx context.Context, id uint, userID uint) error
	CountUnreadByUserID(ctx context.Context, userID uint) (int, error)
	MarkAllAsRead(ctx context.Context, userID uint) error
}
