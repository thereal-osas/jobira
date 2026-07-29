package notifications

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestNewSQLRepository(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sql mock: %v", err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	if repo == nil {
		t.Fatal("expected repository, got nil")
	}

	if repo.db != db {
		t.Fatal("expected repository to use supplied database")
	}
}

func TestSQLRepository_Create_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sql mock: %v", err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	notification := &Notification{
		UserID:  5,
		Title:   "New job application",
		Message: "A cleaner applied for your job",
		Type:    "application",
	}

	query := regexp.QuoteMeta(`
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
	`)

	mock.ExpectQuery(query).
		WithArgs(
			notification.UserID,
			notification.Title,
			notification.Message,
			notification.Type,
			false,
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
		).
		WillReturnRows(
			sqlmock.NewRows([]string{"id"}).
				AddRow(10),
		)

	err = repo.Create(context.Background(), notification)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if notification.ID != 10 {
		t.Fatalf("expected notification ID 10, got %d", notification.ID)
	}

	if notification.IsRead {
		t.Fatal("expected notification to be unread")
	}

	if notification.CreatedAt.IsZero() {
		t.Fatal("expected CreatedAt to be set")
	}

	if notification.UpdatedAt.IsZero() {
		t.Fatal("expected UpdatedAt to be set")
	}

	if !notification.CreatedAt.Equal(notification.UpdatedAt) {
		t.Fatal("expected CreatedAt and UpdatedAt to match")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestSQLRepository_Create_DatabaseError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sql mock: %v", err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	notification := &Notification{
		UserID:  5,
		Title:   "New job application",
		Message: "A cleaner applied for your job",
		Type:    "application",
	}

	repoErr := errors.New("database unavailable")

	query := regexp.QuoteMeta(`
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
	`)

	mock.ExpectQuery(query).
		WithArgs(
			notification.UserID,
			notification.Title,
			notification.Message,
			notification.Type,
			false,
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
		).
		WillReturnError(repoErr)

	err = repo.Create(context.Background(), notification)

	if !errors.Is(err, repoErr) {
		t.Fatalf("expected database error, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}


func TestSQLRepository_ListByUserID_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create SQL mock: %v", err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	now := time.Now()

	query := regexp.QuoteMeta(`
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
	`)

	rows := sqlmock.NewRows([]string{
		"id",
		"user_id",
		"title",
		"message",
		"type",
		"is_read",
		"created_at",
		"updated_at",
	}).
		AddRow(
			2,
			5,
			"Second notification",
			"Second message",
			"application",
			false,
			now,
			now,
		).
		AddRow(
			1,
			5,
			"First notification",
			"First message",
			"booking",
			true,
			now.Add(-time.Hour),
			now.Add(-time.Hour),
		)

	mock.ExpectQuery(query).
		WithArgs(uint(5)).
		WillReturnRows(rows)

	notifications, err := repo.ListByUserID(
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

	if notifications[0].ID != 2 {
		t.Fatalf(
			"expected first notification ID 2, got %d",
			notifications[0].ID,
		)
	}

	if notifications[0].UserID != 5 {
		t.Fatalf(
			"expected user ID 5, got %d",
			notifications[0].UserID,
		)
	}

	if notifications[0].Title != "Second notification" {
		t.Fatalf(
			"expected title %q, got %q",
			"Second notification",
			notifications[0].Title,
		)
	}

	if notifications[0].IsRead {
		t.Fatal("expected first notification to be unread")
	}

	if !notifications[1].IsRead {
		t.Fatal("expected second notification to be read")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestSQLRepository_ListByUserID_Empty(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create SQL mock: %v", err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	query := regexp.QuoteMeta(`
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
	`)

	mock.ExpectQuery(query).
		WithArgs(uint(5)).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"id",
				"user_id",
				"title",
				"message",
				"type",
				"is_read",
				"created_at",
				"updated_at",
			}),
		)

	notifications, err := repo.ListByUserID(
		context.Background(),
		5,
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(notifications) != 0 {
		t.Fatalf(
			"expected no notifications, got %d",
			len(notifications),
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestSQLRepository_ListByUserID_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create SQL mock: %v", err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)
	repoErr := errors.New("database unavailable")

	query := regexp.QuoteMeta(`
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
	`)

	mock.ExpectQuery(query).
		WithArgs(uint(5)).
		WillReturnError(repoErr)

	notifications, err := repo.ListByUserID(
		context.Background(),
		5,
	)

	if notifications != nil {
		t.Fatalf(
			"expected nil notifications, got %#v",
			notifications,
		)
	}

	if !errors.Is(err, repoErr) {
		t.Fatalf("expected database error, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestSQLRepository_ListByUserID_ScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create SQL mock: %v", err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	now := time.Now()

	query := regexp.QuoteMeta(`
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
	`)

	rows := sqlmock.NewRows([]string{
		"id",
		"user_id",
		"title",
		"message",
		"type",
		"is_read",
		"created_at",
		"updated_at",
	}).AddRow(
		"invalid-id",
		5,
		"Notification",
		"Message",
		"application",
		false,
		now,
		now,
	)

	mock.ExpectQuery(query).
		WithArgs(uint(5)).
		WillReturnRows(rows)

	notifications, err := repo.ListByUserID(
		context.Background(),
		5,
	)

	if err == nil {
		t.Fatal("expected scan error, got nil")
	}

	if notifications != nil {
		t.Fatalf(
			"expected nil notifications, got %#v",
			notifications,
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestSQLRepository_ListByUserID_RowsError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create SQL mock: %v", err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)
	rowsErr := errors.New("rows iteration failed")
	now := time.Now()

	query := regexp.QuoteMeta(`
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
	`)

	rows := sqlmock.NewRows([]string{
		"id",
		"user_id",
		"title",
		"message",
		"type",
		"is_read",
		"created_at",
		"updated_at",
	}).
		AddRow(
			1,
			5,
			"First notification",
			"First message",
			"application",
			false,
			now,
			now,
		).
		AddRow(
			2,
			5,
			"Second notification",
			"Second message",
			"booking",
			false,
			now,
			now,
		).
		RowError(1, rowsErr)

	mock.ExpectQuery(query).
		WithArgs(uint(5)).
		WillReturnRows(rows)

	notifications, err := repo.ListByUserID(
		context.Background(),
		5,
	)

	if notifications != nil {
		t.Fatalf(
			"expected nil notifications, got %#v",
			notifications,
		)
	}

	if !errors.Is(err, rowsErr) {
		t.Fatalf("expected rows error, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestSQLRepository_ListUnreadByUserID_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create SQL mock: %v", err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)
	now := time.Now()

	query := regexp.QuoteMeta(`
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
	`)

	rows := sqlmock.NewRows([]string{
		"id",
		"user_id",
		"title",
		"message",
		"type",
		"is_read",
		"created_at",
		"updated_at",
	}).
		AddRow(
			3,
			5,
			"Unread notification",
			"Unread message",
			"application",
			false,
			now,
			now,
		).
		AddRow(
			2,
			5,
			"Another unread notification",
			"Another unread message",
			"booking",
			false,
			now.Add(-time.Hour),
			now.Add(-time.Hour),
		)

	mock.ExpectQuery(query).
		WithArgs(uint(5)).
		WillReturnRows(rows)

	notifications, err := repo.ListUnreadByUserID(
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

	for _, notification := range notifications {
		if notification.IsRead {
			t.Fatalf(
				"expected notification %d to be unread",
				notification.ID,
			)
		}
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestSQLRepository_ListUnreadByUserID_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create SQL mock: %v", err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)
	repoErr := errors.New("database unavailable")

	query := regexp.QuoteMeta(`
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
	`)

	mock.ExpectQuery(query).
		WithArgs(uint(5)).
		WillReturnError(repoErr)

	notifications, err := repo.ListUnreadByUserID(
		context.Background(),
		5,
	)

	if notifications != nil {
		t.Fatalf(
			"expected nil notifications, got %#v",
			notifications,
		)
	}

	if !errors.Is(err, repoErr) {
		t.Fatalf("expected database error, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestSQLRepository_ListUnreadByUserID_ScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create SQL mock: %v", err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)
	now := time.Now()

	query := regexp.QuoteMeta(`
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
	`)

	rows := sqlmock.NewRows([]string{
		"id",
		"user_id",
		"title",
		"message",
		"type",
		"is_read",
		"created_at",
		"updated_at",
	}).AddRow(
		1,
		5,
		"Notification",
		"Message",
		"application",
		"not-a-boolean",
		now,
		now,
	)

	mock.ExpectQuery(query).
		WithArgs(uint(5)).
		WillReturnRows(rows)

	notifications, err := repo.ListUnreadByUserID(
		context.Background(),
		5,
	)

	if err == nil {
		t.Fatal("expected scan error, got nil")
	}

	if notifications != nil {
		t.Fatalf(
			"expected nil notifications, got %#v",
			notifications,
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestSQLRepository_ListUnreadByUserID_RowsError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create SQL mock: %v", err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)
	rowsErr := errors.New("rows iteration failed")
	now := time.Now()

	query := regexp.QuoteMeta(`
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
	`)

	rows := sqlmock.NewRows([]string{
		"id",
		"user_id",
		"title",
		"message",
		"type",
		"is_read",
		"created_at",
		"updated_at",
	}).
		AddRow(
			1,
			5,
			"First notification",
			"First message",
			"application",
			false,
			now,
			now,
		).
		AddRow(
			2,
			5,
			"Second notification",
			"Second message",
			"booking",
			false,
			now,
			now,
		).
		RowError(1, rowsErr)

	mock.ExpectQuery(query).
		WithArgs(uint(5)).
		WillReturnRows(rows)

	notifications, err := repo.ListUnreadByUserID(
		context.Background(),
		5,
	)

	if notifications != nil {
		t.Fatalf(
			"expected nil notifications, got %#v",
			notifications,
		)
	}

	if !errors.Is(err, rowsErr) {
		t.Fatalf("expected rows error, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestSQLRepository_CountUnreadByUserID_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create SQL mock: %v", err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	query := regexp.QuoteMeta(`
		SELECT COUNT(*)
		FROM notifications
		WHERE user_id = $1 
		AND is_read = false 
	`)

	mock.ExpectQuery(query).
		WithArgs(uint(5)).
		WillReturnRows(
			sqlmock.NewRows([]string{"count"}).
				AddRow(4),
		)

	count, err := repo.CountUnreadByUserID(
		context.Background(),
		5,
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if count != 4 {
		t.Fatalf("expected unread count 4, got %d", count)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestSQLRepository_CountUnreadByUserID_DatabaseError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create SQL mock: %v", err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)
	repoErr := errors.New("database unavailable")

	query := regexp.QuoteMeta(`
		SELECT COUNT(*)
		FROM notifications
		WHERE user_id = $1 
		AND is_read = false 
	`)

	mock.ExpectQuery(query).
		WithArgs(uint(5)).
		WillReturnError(repoErr)

	count, err := repo.CountUnreadByUserID(
		context.Background(),
		5,
	)

	if count != 0 {
		t.Fatalf("expected count 0, got %d", count)
	}

	if !errors.Is(err, repoErr) {
		t.Fatalf("expected database error, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestSQLRepository_MarkAsRead_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create SQL mock: %v", err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	query := regexp.QuoteMeta(`
		UPDATE notifications
		SET 	
			is_read = true, 
			updated_at = $1
		WHERE id = $2 
		AND user_id = $3 	
	`)

	mock.ExpectExec(query).
		WithArgs(
			sqlmock.AnyArg(),
			uint(10),
			uint(5),
		).
		WillReturnResult(
			sqlmock.NewResult(0, 1),
		)

	err = repo.MarkAsRead(
		context.Background(),
		10,
		5,
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestSQLRepository_MarkAsRead_DatabaseError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create SQL mock: %v", err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)
	repoErr := errors.New("database unavailable")

	query := regexp.QuoteMeta(`
		UPDATE notifications
		SET 	
			is_read = true, 
			updated_at = $1
		WHERE id = $2 
		AND user_id = $3 	
	`)

	mock.ExpectExec(query).
		WithArgs(
			sqlmock.AnyArg(),
			uint(10),
			uint(5),
		).
		WillReturnError(repoErr)

	err = repo.MarkAsRead(
		context.Background(),
		10,
		5,
	)

	if !errors.Is(err, repoErr) {
		t.Fatalf("expected database error, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestSQLRepository_MarkAsRead_RowsAffectedError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create SQL mock: %v", err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)
	resultErr := errors.New("rows affected unavailable")

	query := regexp.QuoteMeta(`
		UPDATE notifications
		SET 	
			is_read = true, 
			updated_at = $1
		WHERE id = $2 
		AND user_id = $3 	
	`)

	mock.ExpectExec(query).
		WithArgs(
			sqlmock.AnyArg(),
			uint(10),
			uint(5),
		).
		WillReturnResult(
			sqlmock.NewErrorResult(resultErr),
		)

	err = repo.MarkAsRead(
		context.Background(),
		10,
		5,
	)

	if !errors.Is(err, resultErr) {
		t.Fatalf("expected rows affected error, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestSQLRepository_MarkAsRead_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create SQL mock: %v", err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	query := regexp.QuoteMeta(`
		UPDATE notifications
		SET 	
			is_read = true, 
			updated_at = $1
		WHERE id = $2 
		AND user_id = $3 	
	`)

	mock.ExpectExec(query).
		WithArgs(
			sqlmock.AnyArg(),
			uint(999),
			uint(5),
		).
		WillReturnResult(
			sqlmock.NewResult(0, 0),
		)

	err = repo.MarkAsRead(
		context.Background(),
		999,
		5,
	)

	if !errors.Is(err, ErrNotificationNotFound) {
		t.Fatalf(
			"expected ErrNotificationNotFound, got %v",
			err,
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestSQLRepository_MarkAllAsRead_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create SQL mock: %v", err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	query := regexp.QuoteMeta(`
		UPDATE notifications
		SET
			is_read = true, 
			updated_at = $1 
		WHERE user_id = $2 
		AND is_read = false 	
	`)

	mock.ExpectExec(query).
		WithArgs(
			sqlmock.AnyArg(),
			uint(5),
		).
		WillReturnResult(
			sqlmock.NewResult(0, 3),
		)

	err = repo.MarkAllAsRead(
		context.Background(),
		5,
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestSQLRepository_MarkAllAsRead_DatabaseError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create SQL mock: %v", err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)
	repoErr := errors.New("database unavailable")

	query := regexp.QuoteMeta(`
		UPDATE notifications
		SET
			is_read = true, 
			updated_at = $1 
		WHERE user_id = $2 
		AND is_read = false 	
	`)

	mock.ExpectExec(query).
		WithArgs(
			sqlmock.AnyArg(),
			uint(5),
		).
		WillReturnError(repoErr)

	err = repo.MarkAllAsRead(
		context.Background(),
		5,
	)

	if !errors.Is(err, repoErr) {
		t.Fatalf("expected database error, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}


