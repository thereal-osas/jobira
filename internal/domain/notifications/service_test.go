package notifications

import (
	"context"
	"errors"
	"testing"
)

type MockRepository struct {
	createErr  error
	listErr    error
	markErr    error
	markAllErr error
	countErr   error

	notification  *Notification
	notifications []Notification
	unreadCount   int
}

func (m *MockRepository) Create(ctx context.Context, notification *Notification) error {
	if m.createErr != nil {
		return m.createErr
	}

	notification.ID = 1
	m.notification = notification

	return nil
}

func (m *MockRepository) ListByUserID(ctx context.Context, userID uint) ([]Notification, error) {
	return m.notifications, m.listErr
}

func (m *MockRepository) ListUnreadByUserID(ctx context.Context, userID uint) ([]Notification, error) {
	return m.notifications, m.listErr
}

func (m *MockRepository) MarkAsRead(ctx context.Context, id uint, userID uint) error {
	return m.markErr
}

func (m *MockRepository) CountUnreadByUserID(ctx context.Context, userID uint) (int, error) {
	return m.unreadCount, m.countErr
}

func (m *MockRepository) MarkAllAsRead(ctx context.Context, userID uint) error {
	return m.markAllErr
}

func TestCreateNotification_Success(t *testing.T) {
	repo := &MockRepository{}
	service := NewService(repo)

	req := CreateNotificationsRequest{
		UserID:  1,
		Title:   " New message ",
		Message: " You have received a message ",
		Type:    " chat ",
	}

	notification, err := service.Create(context.Background(), req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if notification.ID != 1 {
		t.Errorf("expected notification ID 1, got %d", notification.ID)
	}

	if notification.UserID != 1 {
		t.Errorf("expected user ID 1, got %d", notification.UserID)
	}

	if notification.Title != "New message" {
		t.Errorf(
			"expected title %q, got %q",
			"New message",
			notification.Title,
		)
	}

	if notification.Message != "You have received a message" {
		t.Errorf(
			"expected message %q, got %q",
			"You have received a message",
			notification.Message,
		)
	}

	if notification.Type != "chat" {
		t.Errorf(
			"expected type %q, got %q",
			"chat",
			notification.Type,
		)
	}

	if notification.IsRead {
		t.Error("expected new notification to be unread")
	}
}

func TestCreateNotification_InvalidUserID(t *testing.T) {
	repo := &MockRepository{}
	service := NewService(repo)

	req := CreateNotificationsRequest{
		UserID:  0,
		Title:   "New message",
		Message: "You have received a message",
		Type:    "chat",
	}

	notification, err := service.Create(context.Background(), req)

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}

	if notification != nil {
		t.Errorf("expected nil notification, got %+v", notification)
	}
}

func TestCreateNotification_EmptyFields(t *testing.T) {
	tests := []struct {
		name string
		req  CreateNotificationsRequest
	}{
		{
			name: "empty title",
			req: CreateNotificationsRequest{
				UserID:  1,
				Title:   "",
				Message: "You have received a message",
				Type:    "chat",
			},
		},
		{
			name: "empty message",
			req: CreateNotificationsRequest{
				UserID:  1,
				Title:   "New message",
				Message: "",
				Type:    "chat",
			},
		},
		{
			name: "empty type",
			req: CreateNotificationsRequest{
				UserID:  1,
				Title:   "New message",
				Message: "You have received a message",
				Type:    "",
			},
		},
		{
			name: "whitespace only fields",
			req: CreateNotificationsRequest{
				UserID:  1,
				Title:   "   ",
				Message: "   ",
				Type:    "   ",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := &MockRepository{}
			service := NewService(repo)

			notification, err := service.Create(
				context.Background(),
				test.req,
			)

			if !errors.Is(err, ErrInvalidInput) {
				t.Fatalf(
					"expected ErrInvalidInput, got %v",
					err,
				)
			}

			if notification != nil {
				t.Errorf(
					"expected nil notification, got %+v",
					notification,
				)
			}
		})
	}
}

func TestCreateNotification_RepositoryError(t *testing.T) {
	repoErr := errors.New("database unavailable")

	repo := &MockRepository{
		createErr: repoErr,
	}

	service := NewService(repo)

	req := CreateNotificationsRequest{
		UserID:  1,
		Title:   "New message",
		Message: "You have received a message",
		Type:    "chat",
	}

	notification, err := service.Create(context.Background(), req)

	if !errors.Is(err, repoErr) {
		t.Fatalf("expected repository error, got %v", err)
	}

	if notification != nil {
		t.Errorf(
			"expected nil notification, got %+v",
			notification,
		)
	}
}

func TestListMine_Success(t *testing.T) {
	repo := &MockRepository{
		notifications: []Notification{
			{
				ID:      1,
				UserID:  5,
				Title:   "New message",
				Message: "You received a chat message",
				Type:    "chat",
				IsRead:  false,
			},
			{
				ID:      2,
				UserID:  5,
				Title:   "Application update",
				Message: "Your application was accepted",
				Type:    "application",
				IsRead:  true,
			},
		},
	}

	service := NewService(repo)

	notifications, err := service.ListMine(
		context.Background(),
		5,
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(notifications) != 2 {
		t.Fatalf(
			"expected 2 notifications, got %d",
			len(notifications),
		)
	}

	if notifications[0].ID != 1 {
		t.Errorf(
			"expected first notification ID 1, got %d",
			notifications[0].ID,
		)
	}

	if notifications[1].Type != "application" {
		t.Errorf(
			"expected second notification type %q, got %q",
			"application",
			notifications[1].Type,
		)
	}
}

func TestListMine_InvalidUserID(t *testing.T) {
	repo := &MockRepository{}
	service := NewService(repo)

	notifications, err := service.ListMine(
		context.Background(),
		0,
	)

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}

	if notifications != nil {
		t.Errorf(
			"expected nil notifications, got %+v",
			notifications,
		)
	}
}

func TestListMine_RepositoryError(t *testing.T) {
	repoErr := errors.New("database unavailable")

	repo := &MockRepository{
		listErr: repoErr,
	}

	service := NewService(repo)

	notifications, err := service.ListMine(
		context.Background(),
		5,
	)

	if !errors.Is(err, repoErr) {
		t.Fatalf("expected repository error, got %v", err)
	}

	if notifications != nil {
		t.Errorf(
			"expected nil notifications, got %+v",
			notifications,
		)
	}
}

func TestListUnreadMine_Success(t *testing.T) {
	repo := &MockRepository{
		notifications: []Notification{
			{
				ID:      1,
				UserID:  5,
				Title:   "New message",
				Message: "Unread notification",
				Type:    "chat",
				IsRead:  false,
			},
		},
	}

	service := NewService(repo)

	notifications, err := service.ListUnreadMine(
		context.Background(),
		5,
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(notifications) != 1 {
		t.Fatalf("expected 1 notification, got %d", len(notifications))
	}

	if notifications[0].IsRead {
		t.Error("expected unread notification")
	}
}

func TestListUnreadMine_InvalidUserID(t *testing.T) {
	repo := &MockRepository{}
	service := NewService(repo)

	notifications, err := service.ListUnreadMine(
		context.Background(),
		0,
	)

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}

	if notifications != nil {
		t.Errorf("expected nil notifications, got %+v", notifications)
	}
}

func TestListUnreadMine_RepositoryError(t *testing.T) {
	repoErr := errors.New("database unavailable")

	repo := &MockRepository{
		listErr: repoErr,
	}

	service := NewService(repo)

	notifications, err := service.ListUnreadMine(
		context.Background(),
		5,
	)

	if !errors.Is(err, repoErr) {
		t.Fatalf("expected repository error, got %v", err)
	}

	if notifications != nil {
		t.Errorf("expected nil notifications, got %+v", notifications)
	}
}

func TestCountUnreadMine_Success(t *testing.T) {
	repo := &MockRepository{
		unreadCount: 4,
	}

	service := NewService(repo)

	count, err := service.CountUnreadMine(
		context.Background(),
		5,
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if count != 4 {
		t.Errorf("expected count 4, got %d", count)
	}
}

func TestCountUnreadMine_InvalidUserID(t *testing.T) {
	repo := &MockRepository{}
	service := NewService(repo)

	count, err := service.CountUnreadMine(
		context.Background(),
		0,
	)

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}

	if count != 0 {
		t.Errorf("expected count 0, got %d", count)
	}
}

func TestCountUnreadMine_RepositoryError(t *testing.T) {
	repoErr := errors.New("database unavailable")

	repo := &MockRepository{
		countErr: repoErr,
	}

	service := NewService(repo)

	_, err := service.CountUnreadMine(
		context.Background(),
		5,
	)

	if !errors.Is(err, repoErr) {
		t.Fatalf("expected repository error, got %v", err)
	}
}

func TestMarkAsRead_Success(t *testing.T) {
	repo := &MockRepository{}
	service := NewService(repo)

	err := service.MarkAsRead(
		context.Background(),
		10,
		5,
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestMarkAsRead_InvalidInput(t *testing.T) {
	tests := []struct {
		name           string
		notificationID uint
		userID         uint
	}{
		{
			name:           "zero notification ID",
			notificationID: 0,
			userID:         5,
		},
		{
			name:           "zero user ID",
			notificationID: 10,
			userID:         0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := &MockRepository{}
			service := NewService(repo)

			err := service.MarkAsRead(
				context.Background(),
				test.notificationID,
				test.userID,
			)

			if !errors.Is(err, ErrInvalidInput) {
				t.Fatalf("expected ErrInvalidInput, got %v", err)
			}
		})
	}
}

func TestMarkAsRead_RepositoryError(t *testing.T) {
	repoErr := errors.New("database unavailable")

	repo := &MockRepository{
		markErr: repoErr,
	}

	service := NewService(repo)

	err := service.MarkAsRead(
		context.Background(),
		10,
		5,
	)

	if !errors.Is(err, repoErr) {
		t.Fatalf("expected repository error, got %v", err)
	}
}

func TestMarkAllAsRead_Success(t *testing.T) {
	repo := &MockRepository{}
	service := NewService(repo)

	err := service.MarkAllAsRead(
		context.Background(),
		5,
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestMarkAllAsRead_InvalidUserID(t *testing.T) {
	repo := &MockRepository{}
	service := NewService(repo)

	err := service.MarkAllAsRead(
		context.Background(),
		0,
	)

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestMarkAllAsRead_RepositoryError(t *testing.T) {
	repoErr := errors.New("database unavailable")

	repo := &MockRepository{
		markAllErr: repoErr,
	}

	service := NewService(repo)

	err := service.MarkAllAsRead(
		context.Background(),
		5,
	)

	if !errors.Is(err, repoErr) {
		t.Fatalf("expected repository error, got %v", err)
	}
}
