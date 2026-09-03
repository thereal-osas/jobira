package chat

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lib/pq"
)

func newChatRepositoryTest(
	t *testing.T,
) (*SQLRepository, sqlmock.Sqlmock, func()) {
	t.Helper()

	db, mock, err := sqlmock.New(
		sqlmock.QueryMatcherOption(
			sqlmock.QueryMatcherRegexp,
		),
	)
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}

	cleanup := func() {
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unmet SQL expectations: %v", err)
		}
	}

	return NewSQLRepository(db), mock, cleanup
}
func TestNewSQLRepository(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}

	repository := NewSQLRepository(db)

	if repository == nil {
		t.Fatal("expected repository")
	}

	if repository.db != db {
		t.Fatal("expected repository to contain supplied database")
	}
}

func TestSQLRepository_CreateConversation_Success(t *testing.T) {
	repository, mock, cleanup := newChatRepositoryTest(t)
	defer cleanup()

	now := time.Now()

	conversation := &Conversation{
		BookingID: 5,
		ClientID:  10,
		CleanerID: 11,
		Status:    "active",
	}

	query := regexp.QuoteMeta(`
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
	`)

	mock.ExpectQuery(query).
		WithArgs(
			conversation.BookingID,
			conversation.ClientID,
			conversation.CleanerID,
			conversation.Status,
			sqlmock.AnyArg(),
		).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{
					"id",
					"created_at",
					"updated_at",
				},
			).AddRow(
				1,
				now,
				now,
			),
		)

	err := repository.CreateConversation(
		context.Background(),
		conversation,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if conversation.ID != 1 {
		t.Fatalf(
			"expected conversation ID 1, got %d",
			conversation.ID,
		)
	}

	if conversation.CreatedAt.IsZero() {
		t.Fatal("expected CreatedAt to be populated")
	}

	if conversation.UpdatedAt.IsZero() {
		t.Fatal("expected UpdatedAt to be populated")
	}
}

func TestSQLRepository_CreateConversation_Duplicate(
	t *testing.T,
) {
	repository, mock, cleanup := newChatRepositoryTest(t)
	defer cleanup()

	conversation := &Conversation{
		BookingID: 5,
		ClientID:  10,
		CleanerID: 11,
		Status:    "active",
	}

	query := regexp.QuoteMeta(`
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
	`)

	mock.ExpectQuery(query).
		WithArgs(
			conversation.BookingID,
			conversation.ClientID,
			conversation.CleanerID,
			conversation.Status,
			sqlmock.AnyArg(),
		).
		WillReturnError(
			&pq.Error{
				Code: pq.ErrorCode("23505"),
			},
		)

	err := repository.CreateConversation(
		context.Background(),
		conversation,
	)

	if !errors.Is(err, ErrConversationExists) {
		t.Fatalf(
			"expected ErrConversationExists, got %v",
			err,
		)
	}
}

func TestSQLRepository_CreateConversation_DatabaseError(
	t *testing.T,
) {
	repository, mock, cleanup := newChatRepositoryTest(t)
	defer cleanup()

	databaseErr := errors.New("database error")

	conversation := &Conversation{
		BookingID: 5,
		ClientID:  10,
		CleanerID: 11,
		Status:    "active",
	}

	query := regexp.QuoteMeta(`
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
	`)

	mock.ExpectQuery(query).
		WithArgs(
			conversation.BookingID,
			conversation.ClientID,
			conversation.CleanerID,
			conversation.Status,
			sqlmock.AnyArg(),
		).
		WillReturnError(databaseErr)

	err := repository.CreateConversation(
		context.Background(),
		conversation,
	)

	if !errors.Is(err, databaseErr) {
		t.Fatalf(
			"expected database error, got %v",
			err,
		)
	}
}

func TestSQLRepository_GetConversationByID_Success(
	t *testing.T,
) {
	repository, mock, cleanup := newChatRepositoryTest(t)
	defer cleanup()

	now := time.Now()
	lastMessageAt := now.Add(-time.Minute)

	query := regexp.QuoteMeta(`
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
	`)

	mock.ExpectQuery(query).
		WithArgs(uint(1)).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{
					"id",
					"booking_id",
					"client_id",
					"cleaner_id",
					"status",
					"last_message_at",
					"created_at",
					"updated_at",
				},
			).AddRow(
				1,
				5,
				10,
				11,
				"active",
				lastMessageAt,
				now,
				now,
			),
		)

	conversation, err := repository.GetConversationByID(
		context.Background(),
		1,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if conversation == nil {
		t.Fatal("expected conversation")
	}

	if conversation.ID != 1 {
		t.Fatalf(
			"expected ID 1, got %d",
			conversation.ID,
		)
	}

	if conversation.BookingID != 5 {
		t.Fatalf(
			"expected booking ID 5, got %d",
			conversation.BookingID,
		)
	}

	if conversation.LastMessagesAt == nil {
		t.Fatal("expected last message time")
	}
}

func TestSQLRepository_GetConversationByID_NotFound(
	t *testing.T,
) {
	repository, mock, cleanup := newChatRepositoryTest(t)
	defer cleanup()

	query := regexp.QuoteMeta(`
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
	`)

	mock.ExpectQuery(query).
		WithArgs(uint(999)).
		WillReturnError(sql.ErrNoRows)

	conversation, err := repository.GetConversationByID(
		context.Background(),
		999,
	)

	if conversation != nil {
		t.Fatal("expected nil conversation")
	}

	if !errors.Is(err, ErrConversationNotFound) {
		t.Fatalf(
			"expected ErrConversationNotFound, got %v",
			err,
		)
	}
}

func TestSQLRepository_GetConversationByID_DatabaseError(
	t *testing.T,
) {
	repository, mock, cleanup := newChatRepositoryTest(t)
	defer cleanup()

	databaseErr := errors.New("database error")

	query := regexp.QuoteMeta(`
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
	`)

	mock.ExpectQuery(query).
		WithArgs(uint(1)).
		WillReturnError(databaseErr)

	_, err := repository.GetConversationByID(
		context.Background(),
		1,
	)

	if !errors.Is(err, databaseErr) {
		t.Fatalf(
			"expected database error, got %v",
			err,
		)
	}
}

func TestSQLRepository_GetConversationByBookingID_Success(
	t *testing.T,
) {
	repository, mock, cleanup := newChatRepositoryTest(t)
	defer cleanup()

	now := time.Now()

	query := regexp.QuoteMeta(`
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
	`)

	mock.ExpectQuery(query).
		WithArgs(uint(5)).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{
					"id",
					"booking_id",
					"client_id",
					"cleaner_id",
					"status",
					"last_message_at",
					"created_at",
					"updated_at",
				},
			).AddRow(
				1,
				5,
				10,
				11,
				"active",
				nil,
				now,
				now,
			),
		)

	conversation, err :=
		repository.GetConversationByBookingID(
			context.Background(),
			5,
		)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if conversation == nil {
		t.Fatal("expected conversation")
	}

	if conversation.BookingID != 5 {
		t.Fatalf(
			"expected booking ID 5, got %d",
			conversation.BookingID,
		)
	}

	if conversation.LastMessagesAt != nil {
		t.Fatal("expected nil last-message time")
	}
}

func TestSQLRepository_GetConversationByBookingID_NotFound(
	t *testing.T,
) {
	repository, mock, cleanup := newChatRepositoryTest(t)
	defer cleanup()

	query := regexp.QuoteMeta(`
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
	`)

	mock.ExpectQuery(query).
		WithArgs(uint(999)).
		WillReturnError(sql.ErrNoRows)

	conversation, err :=
		repository.GetConversationByBookingID(
			context.Background(),
			999,
		)

	if conversation != nil {
		t.Fatal("expected nil conversation")
	}

	if !errors.Is(err, ErrConversationNotFound) {
		t.Fatalf(
			"expected ErrConversationNotFound, got %v",
			err,
		)
	}
}

func TestSQLRepository_GetConversationByBookingID_Error(
	t *testing.T,
) {
	repository, mock, cleanup := newChatRepositoryTest(t)
	defer cleanup()

	databaseErr := errors.New("database error")

	query := regexp.QuoteMeta(`
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
	`)

	mock.ExpectQuery(query).
		WithArgs(uint(5)).
		WillReturnError(databaseErr)

	_, err := repository.GetConversationByBookingID(
		context.Background(),
		5,
	)

	if !errors.Is(err, databaseErr) {
		t.Fatalf(
			"expected database error, got %v",
			err,
		)
	}
}

func conversationSummaryColumns() []string {
	return []string{
		"id",
		"booking_id",
		"client_id",
		"cleaner_id",
		"status",
		"other_user_id",
		"other_user_name",
		"content",
		"sender_id",
		"last_message_at",
		"unread_count",
		"created_at",
	}
}

func TestSQLRepository_ListConversationByUserID_Success(
	t *testing.T,
) {
	repository, mock, cleanup := newChatRepositoryTest(t)
	defer cleanup()

	now := time.Now()

	mock.ExpectQuery(
		`(?s)SELECT.*FROM conversations c.*WHERE c\.client_id = \$1.*ORDER BY`,
	).
		WithArgs(uint(10)).
		WillReturnRows(
			sqlmock.NewRows(
				conversationSummaryColumns(),
			).AddRow(
				1,
				5,
				10,
				11,
				"active",
				11,
				"Cleaner One",
				"Hello",
				11,
				now,
				2,
				now,
			).AddRow(
				2,
				6,
				10,
				12,
				"active",
				12,
				"Cleaner Two",
				nil,
				nil,
				nil,
				0,
				now,
			),
		)

	conversations, err :=
		repository.ListConversationByUserID(
			context.Background(),
			10,
		)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(conversations) != 2 {
		t.Fatalf(
			"expected 2 conversations, got %d",
			len(conversations),
		)
	}

	if conversations[0].OtherUserName != "Cleaner One" {
		t.Fatalf(
			"expected Cleaner One, got %q",
			conversations[0].OtherUserName,
		)
	}

	if conversations[0].UnreadCount != 2 {
		t.Fatalf(
			"expected unread count 2, got %d",
			conversations[0].UnreadCount,
		)
	}

	if conversations[1].LastMessage != nil {
		t.Fatal("expected nil last message")
	}
}

func TestSQLRepository_ListConversationByUserID_Empty(
	t *testing.T,
) {
	repository, mock, cleanup := newChatRepositoryTest(t)
	defer cleanup()

	mock.ExpectQuery(
		`(?s)SELECT.*FROM conversations c.*WHERE c\.client_id = \$1.*ORDER BY`,
	).
		WithArgs(uint(10)).
		WillReturnRows(
			sqlmock.NewRows(
				conversationSummaryColumns(),
			),
		)

	conversations, err :=
		repository.ListConversationByUserID(
			context.Background(),
			10,
		)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if conversations == nil {
		t.Fatal("expected non-nil empty slice")
	}

	if len(conversations) != 0 {
		t.Fatalf(
			"expected empty list, got %d",
			len(conversations),
		)
	}
}

func TestSQLRepository_ListConversationByUserID_QueryError(
	t *testing.T,
) {
	repository, mock, cleanup := newChatRepositoryTest(t)
	defer cleanup()

	databaseErr := errors.New("query error")

	mock.ExpectQuery(
		`(?s)SELECT.*FROM conversations c.*WHERE c\.client_id = \$1.*ORDER BY`,
	).
		WithArgs(uint(10)).
		WillReturnError(databaseErr)

	_, err := repository.ListConversationByUserID(
		context.Background(),
		10,
	)

	if !errors.Is(err, databaseErr) {
		t.Fatalf(
			"expected query error, got %v",
			err,
		)
	}
}

func TestSQLRepository_ListConversationByUserID_ScanError(
	t *testing.T,
) {
	repository, mock, cleanup := newChatRepositoryTest(t)
	defer cleanup()

	now := time.Now()

	mock.ExpectQuery(
		`(?s)SELECT.*FROM conversations c.*WHERE c\.client_id = \$1.*ORDER BY`,
	).
		WithArgs(uint(10)).
		WillReturnRows(
			sqlmock.NewRows(
				conversationSummaryColumns(),
			).AddRow(
				"invalid-id",
				5,
				10,
				11,
				"active",
				11,
				"Cleaner One",
				"Hello",
				11,
				now,
				2,
				now,
			),
		)

	_, err := repository.ListConversationByUserID(
		context.Background(),
		10,
	)

	if err == nil {
		t.Fatal("expected scan error")
	}
}

func TestSQLRepository_ListConversationByUserID_RowsError(
	t *testing.T,
) {
	repository, mock, cleanup := newChatRepositoryTest(t)
	defer cleanup()

	rowsErr := errors.New("rows error")
	now := time.Now()

	rows := sqlmock.NewRows(
		conversationSummaryColumns(),
	).
		AddRow(
			1,
			5,
			10,
			11,
			"active",
			11,
			"Cleaner One",
			"Hello",
			11,
			now,
			2,
			now,
		).
		RowError(0, rowsErr)

	mock.ExpectQuery(
		`(?s)SELECT.*FROM conversations c.*WHERE c\.client_id = \$1.*ORDER BY`,
	).
		WithArgs(uint(10)).
		WillReturnRows(rows)

	_, err := repository.ListConversationByUserID(
		context.Background(),
		10,
	)

	if !errors.Is(err, rowsErr) {
		t.Fatalf(
			"expected rows error, got %v",
			err,
		)
	}
}

func messageColumns() []string {
	return []string{
		"id",
		"conversation_id",
		"sender_id",
		"message_type",
		"content",
		"is_read",
		"read_at",
		"created_at",
		"updated_at",
	}
}

func TestSQLRepository_CreateMessage_Success(t *testing.T) {
	repository, mock, cleanup := newChatRepositoryTest(t)
	defer cleanup()

	now := time.Now()

	message := &Message{
		ConversationID: 1,
		SenderID:       10,
		MessageType:    "text",
		Content:        "Hello cleaner",
	}

	mock.ExpectBegin()

	mock.ExpectQuery(
		`(?s)INSERT INTO messages.*RETURNING.*id.*is_read.*read_at.*created_at.*updated_at`,
	).
		WithArgs(
			message.ConversationID,
			message.SenderID,
			message.MessageType,
			message.Content,
			sqlmock.AnyArg(),
		).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{
					"id",
					"is_read",
					"read_at",
					"created_at",
					"updated_at",
				},
			).AddRow(
				1,
				false,
				nil,
				now,
				now,
			),
		)

	mock.ExpectExec(
		`(?s)UPDATE conversations.*last_message_at = \$1.*WHERE id = \$2`,
	).
		WithArgs(
			sqlmock.AnyArg(),
			message.ConversationID,
		).
		WillReturnResult(
			sqlmock.NewResult(0, 1),
		)

	mock.ExpectCommit()

	err := repository.CreateMessage(
		context.Background(),
		message,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if message.ID != 1 {
		t.Fatalf(
			"expected message ID 1, got %d",
			message.ID,
		)
	}

	if message.IsRead {
		t.Fatal("expected message to be unread")
	}
}

func TestSQLRepository_CreateMessage_BeginFails(t *testing.T) {
	repository, mock, cleanup := newChatRepositoryTest(t)
	defer cleanup()

	beginErr := errors.New("begin transaction failed")

	mock.ExpectBegin().WillReturnError(beginErr)

	err := repository.CreateMessage(
		context.Background(),
		&Message{},
	)

	if !errors.Is(err, beginErr) {
		t.Fatalf(
			"expected begin error, got %v",
			err,
		)
	}
}

func TestSQLRepository_CreateMessage_InsertFails(t *testing.T) {
	repository, mock, cleanup := newChatRepositoryTest(t)
	defer cleanup()

	insertErr := errors.New("insert message failed")

	message := &Message{
		ConversationID: 1,
		SenderID:       10,
		MessageType:    "text",
		Content:        "Hello",
	}

	mock.ExpectBegin()

	mock.ExpectQuery(
		`(?s)INSERT INTO messages.*RETURNING.*id.*is_read.*read_at.*created_at.*updated_at`,
	).
		WithArgs(
			message.ConversationID,
			message.SenderID,
			message.MessageType,
			message.Content,
			sqlmock.AnyArg(),
		).
		WillReturnError(insertErr)

	mock.ExpectRollback()

	err := repository.CreateMessage(
		context.Background(),
		message,
	)

	if !errors.Is(err, insertErr) {
		t.Fatalf(
			"expected insert error, got %v",
			err,
		)
	}
}

func TestSQLRepository_CreateMessage_UpdateFails(t *testing.T) {
	repository, mock, cleanup := newChatRepositoryTest(t)
	defer cleanup()

	now := time.Now()
	updateErr := errors.New("update conversation failed")

	message := &Message{
		ConversationID: 1,
		SenderID:       10,
		MessageType:    "text",
		Content:        "Hello",
	}

	mock.ExpectBegin()

	mock.ExpectQuery(
		`(?s)INSERT INTO messages.*RETURNING.*id.*is_read.*read_at.*created_at.*updated_at`,
	).
		WithArgs(
			message.ConversationID,
			message.SenderID,
			message.MessageType,
			message.Content,
			sqlmock.AnyArg(),
		).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{
					"id",
					"is_read",
					"read_at",
					"created_at",
					"updated_at",
				},
			).AddRow(
				1,
				false,
				nil,
				now,
				now,
			),
		)

	mock.ExpectExec(
		`(?s)UPDATE conversations.*last_message_at = \$1.*WHERE id = \$2`,
	).
		WithArgs(
			sqlmock.AnyArg(),
			message.ConversationID,
		).
		WillReturnError(updateErr)

	mock.ExpectRollback()

	err := repository.CreateMessage(
		context.Background(),
		message,
	)

	if !errors.Is(err, updateErr) {
		t.Fatalf(
			"expected update error, got %v",
			err,
		)
	}
}

func TestSQLRepository_CreateMessage_RowsAffectedFails(
	t *testing.T,
) {
	repository, mock, cleanup := newChatRepositoryTest(t)
	defer cleanup()

	now := time.Now()
	rowsErr := errors.New("rows affected failed")

	message := &Message{
		ConversationID: 1,
		SenderID:       10,
		MessageType:    "text",
		Content:        "Hello",
	}

	mock.ExpectBegin()

	mock.ExpectQuery(
		`(?s)INSERT INTO messages.*RETURNING.*id.*is_read.*read_at.*created_at.*updated_at`,
	).
		WithArgs(
			message.ConversationID,
			message.SenderID,
			message.MessageType,
			message.Content,
			sqlmock.AnyArg(),
		).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{
					"id",
					"is_read",
					"read_at",
					"created_at",
					"updated_at",
				},
			).AddRow(
				1,
				false,
				nil,
				now,
				now,
			),
		)

	mock.ExpectExec(
		`(?s)UPDATE conversations.*last_message_at = \$1.*WHERE id = \$2`,
	).
		WithArgs(
			sqlmock.AnyArg(),
			message.ConversationID,
		).
		WillReturnResult(
			sqlmock.NewErrorResult(rowsErr),
		)

	mock.ExpectRollback()

	err := repository.CreateMessage(
		context.Background(),
		message,
	)

	if !errors.Is(err, rowsErr) {
		t.Fatalf(
			"expected rows-affected error, got %v",
			err,
		)
	}
}

func TestSQLRepository_CreateMessage_ConversationNotFound(
	t *testing.T,
) {
	repository, mock, cleanup := newChatRepositoryTest(t)
	defer cleanup()

	now := time.Now()

	message := &Message{
		ConversationID: 999,
		SenderID:       10,
		MessageType:    "text",
		Content:        "Hello",
	}

	mock.ExpectBegin()

	mock.ExpectQuery(
		`(?s)INSERT INTO messages.*RETURNING.*id.*is_read.*read_at.*created_at.*updated_at`,
	).
		WithArgs(
			message.ConversationID,
			message.SenderID,
			message.MessageType,
			message.Content,
			sqlmock.AnyArg(),
		).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{
					"id",
					"is_read",
					"read_at",
					"created_at",
					"updated_at",
				},
			).AddRow(
				1,
				false,
				nil,
				now,
				now,
			),
		)

	mock.ExpectExec(
		`(?s)UPDATE conversations.*last_message_at = \$1.*WHERE id = \$2`,
	).
		WithArgs(
			sqlmock.AnyArg(),
			message.ConversationID,
		).
		WillReturnResult(
			sqlmock.NewResult(0, 0),
		)

	mock.ExpectRollback()

	err := repository.CreateMessage(
		context.Background(),
		message,
	)

	if !errors.Is(err, ErrConversationNotFound) {
		t.Fatalf(
			"expected ErrConversationNotFound, got %v",
			err,
		)
	}
}

func TestSQLRepository_CreateMessage_CommitFails(t *testing.T) {
	repository, mock, cleanup := newChatRepositoryTest(t)
	defer cleanup()

	now := time.Now()
	commitErr := errors.New("commit failed")

	message := &Message{
		ConversationID: 1,
		SenderID:       10,
		MessageType:    "text",
		Content:        "Hello",
	}

	mock.ExpectBegin()

	mock.ExpectQuery(
		`(?s)INSERT INTO messages.*RETURNING.*id.*is_read.*read_at.*created_at.*updated_at`,
	).
		WithArgs(
			message.ConversationID,
			message.SenderID,
			message.MessageType,
			message.Content,
			sqlmock.AnyArg(),
		).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{
					"id",
					"is_read",
					"read_at",
					"created_at",
					"updated_at",
				},
			).AddRow(
				1,
				false,
				nil,
				now,
				now,
			),
		)

	mock.ExpectExec(
		`(?s)UPDATE conversations.*last_message_at = \$1.*WHERE id = \$2`,
	).
		WithArgs(
			sqlmock.AnyArg(),
			message.ConversationID,
		).
		WillReturnResult(
			sqlmock.NewResult(0, 1),
		)

	mock.ExpectCommit().WillReturnError(commitErr)

	// database/sql considers a failed commit to have completed the
	// transaction, so the deferred rollback does not issue another
	// driver call.
	err := repository.CreateMessage(
		context.Background(),
		message,
	)

	if !errors.Is(err, commitErr) {
		t.Fatalf(
			"expected commit error, got %v",
			err,
		)
	}
}

func TestSQLRepository_ListMessage_Success(t *testing.T) {
	repository, mock, cleanup := newChatRepositoryTest(t)
	defer cleanup()

	now := time.Now()

	query := regexp.QuoteMeta(`
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
	`)

	mock.ExpectQuery(query).
		WithArgs(
			uint(1),
			20,
			0,
		).
		WillReturnRows(
			sqlmock.NewRows(
				messageColumns(),
			).AddRow(
				1,
				1,
				10,
				"text",
				"Hello",
				false,
				nil,
				now,
				now,
			).AddRow(
				2,
				1,
				11,
				"text",
				"Hi",
				true,
				now,
				now,
				now,
			),
		)

	messages, err := repository.ListMessage(
		context.Background(),
		1,
		20,
		0,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(messages) != 2 {
		t.Fatalf(
			"expected 2 messages, got %d",
			len(messages),
		)
	}

	if messages[0].Content != "Hello" {
		t.Fatalf(
			"expected Hello, got %q",
			messages[0].Content,
		)
	}

	if !messages[1].IsRead {
		t.Fatal("expected second message to be read")
	}
}

func TestSQLRepository_ListMessage_Empty(t *testing.T) {
	repository, mock, cleanup := newChatRepositoryTest(t)
	defer cleanup()

	query := regexp.QuoteMeta(`
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
	`)

	mock.ExpectQuery(query).
		WithArgs(
			uint(1),
			20,
			0,
		).
		WillReturnRows(
			sqlmock.NewRows(messageColumns()),
		)

	messages, err := repository.ListMessage(
		context.Background(),
		1,
		20,
		0,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if messages == nil {
		t.Fatal("expected non-nil empty slice")
	}

	if len(messages) != 0 {
		t.Fatalf(
			"expected empty list, got %d",
			len(messages),
		)
	}
}

func TestSQLRepository_ListMessage_QueryError(t *testing.T) {
	repository, mock, cleanup := newChatRepositoryTest(t)
	defer cleanup()

	queryErr := errors.New("query error")

	query := regexp.QuoteMeta(`
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
	`)

	mock.ExpectQuery(query).
		WithArgs(
			uint(1),
			20,
			0,
		).
		WillReturnError(queryErr)

	_, err := repository.ListMessage(
		context.Background(),
		1,
		20,
		0,
	)

	if !errors.Is(err, queryErr) {
		t.Fatalf(
			"expected query error, got %v",
			err,
		)
	}
}

func TestSQLRepository_ListMessage_ScanError(t *testing.T) {
	repository, mock, cleanup := newChatRepositoryTest(t)
	defer cleanup()

	now := time.Now()

	query := regexp.QuoteMeta(`
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
	`)

	mock.ExpectQuery(query).
		WithArgs(
			uint(1),
			20,
			0,
		).
		WillReturnRows(
			sqlmock.NewRows(
				messageColumns(),
			).AddRow(
				"invalid-id",
				1,
				10,
				"text",
				"Hello",
				false,
				nil,
				now,
				now,
			),
		)

	_, err := repository.ListMessage(
		context.Background(),
		1,
		20,
		0,
	)

	if err == nil {
		t.Fatal("expected scan error")
	}
}

func TestSQLRepository_ListMessage_RowsError(t *testing.T) {
	repository, mock, cleanup := newChatRepositoryTest(t)
	defer cleanup()

	now := time.Now()
	rowsErr := errors.New("rows error")

	rows := sqlmock.NewRows(messageColumns()).
		AddRow(
			1,
			1,
			10,
			"text",
			"Hello",
			false,
			nil,
			now,
			now,
		).
		RowError(0, rowsErr)

	query := regexp.QuoteMeta(`
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
	`)

	mock.ExpectQuery(query).
		WithArgs(
			uint(1),
			20,
			0,
		).
		WillReturnRows(rows)

	_, err := repository.ListMessage(
		context.Background(),
		1,
		20,
		0,
	)

	if !errors.Is(err, rowsErr) {
		t.Fatalf(
			"expected rows error, got %v",
			err,
		)
	}
}

func TestSQLRepository_MarkConversationAsRead_Success(
	t *testing.T,
) {
	repository, mock, cleanup := newChatRepositoryTest(t)
	defer cleanup()

	query := regexp.QuoteMeta(`
		UPDATE messages
		SET 
			is_read = true, 
			read_at = NOW(), 
			updated_at = NOW()
		WHERE conversation_id = $1
		AND sender_id <> $2
		AND is_read = false	
	`)

	mock.ExpectExec(query).
		WithArgs(
			uint(1),
			uint(10),
		).
		WillReturnResult(
			sqlmock.NewResult(0, 2),
		)

	err := repository.MarkConversationAsRead(
		context.Background(),
		1,
		10,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSQLRepository_MarkConversationAsRead_Error(
	t *testing.T,
) {
	repository, mock, cleanup := newChatRepositoryTest(t)
	defer cleanup()

	databaseErr := errors.New("update failed")

	query := regexp.QuoteMeta(`
		UPDATE messages
		SET 
			is_read = true, 
			read_at = NOW(), 
			updated_at = NOW()
		WHERE conversation_id = $1
		AND sender_id <> $2
		AND is_read = false	
	`)

	mock.ExpectExec(query).
		WithArgs(
			uint(1),
			uint(10),
		).
		WillReturnError(databaseErr)

	err := repository.MarkConversationAsRead(
		context.Background(),
		1,
		10,
	)

	if !errors.Is(err, databaseErr) {
		t.Fatalf(
			"expected database error, got %v",
			err,
		)
	}
}

var _ Repository = (*SQLRepository)(nil)
