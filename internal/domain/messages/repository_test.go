package messages

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func newMessagesRepositoryTest(
	t *testing.T,
) (*SQLRepository, sqlmock.Sqlmock, func()) {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf(
			"failed to create sqlmock: %v",
			err,
		)
	}
cleanup := func() {
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf(
			"unmet SQL expectations: %v",
			err,
		)
	}
}

	return NewSQLRepository(db), mock, cleanup
}

func TestNewSQLRepository(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf(
			"failed creating database: %v",
			err,
		)
	}
	defer db.Close()

	repository := NewSQLRepository(db)

	if repository == nil {
		t.Fatal("expected repository")
	}

	if repository.db != db {
		t.Fatal("expected supplied database")
	}
}

func TestSQLRepository_Create_Success(t *testing.T) {
	repository, mock, cleanup :=
		newMessagesRepositoryTest(t)
	defer cleanup()

	message := &Message{
		JobID:      10,
		SenderID:   1,
		ReceiverID: 2,
		Content:    "hello",
	}

	mock.ExpectQuery(
		`(?s)INSERT INTO messages.*RETURNING id`,
	).
		WithArgs(
			message.JobID,
			message.SenderID,
			message.ReceiverID,
			message.Content,
			false,
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
		).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{"id"},
			).AddRow(5),
		)

	err := repository.Create(
		context.Background(),
		message,
	)
	if err != nil {
		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}

	if message.ID != 5 {
		t.Fatalf(
			"expected ID 5 got %d",
			message.ID,
		)
	}

	if message.IsRead {
		t.Fatal(
			"expected new message to be unread",
		)
	}

	if message.CreatedAt.IsZero() {
		t.Fatal("expected CreatedAt")
	}

	if message.UpdatedAt.IsZero() {
		t.Fatal("expected UpdatedAt")
	}
}

func TestSQLRepository_Create_Error(t *testing.T) {
	repository, mock, cleanup :=
		newMessagesRepositoryTest(t)
	defer cleanup()

	createErr := errors.New("insert failed")

	message := &Message{
		JobID:      10,
		SenderID:   1,
		ReceiverID: 2,
		Content:    "hello",
	}

	mock.ExpectQuery(
		`(?s)INSERT INTO messages.*RETURNING id`,
	).
		WithArgs(
			message.JobID,
			message.SenderID,
			message.ReceiverID,
			message.Content,
			false,
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
		).
		WillReturnError(createErr)

	err := repository.Create(
		context.Background(),
		message,
	)

	if !errors.Is(err, createErr) {
		t.Fatalf(
			"expected create error, got %v",
			err,
		)
	}
}

func messageColumns() []string {
	return []string{
		"id",
		"job_id",
		"sender_id",
		"receiver_id",
		"content",
		"is_read",
		"created_at",
		"updated_at",
	}
}

func TestSQLRepository_ListByJobID_Success(
	t *testing.T,
) {
	repository, mock, cleanup :=
		newMessagesRepositoryTest(t)
	defer cleanup()

	now := time.Now()

	mock.ExpectQuery(
		`(?s)SELECT.*FROM messages.*WHERE job_id = \$1.*ORDER BY created_at ASC`,
	).
		WithArgs(uint(10)).
		WillReturnRows(
			sqlmock.NewRows(messageColumns()).
				AddRow(
					1,
					10,
					1,
					2,
					"hello",
					false,
					now,
					now,
				).
				AddRow(
					2,
					10,
					2,
					1,
					"hi",
					true,
					now,
					now,
				),
		)

	messages, err := repository.ListByJobID(
		context.Background(),
		10,
	)
	if err != nil {
		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}

	if len(messages) != 2 {
		t.Fatalf(
			"expected 2 messages got %d",
			len(messages),
		)
	}

	if messages[0].Content != "hello" {
		t.Fatalf(
			"expected hello got %q",
			messages[0].Content,
		)
	}
}

func TestSQLRepository_ListByJobID_Empty(
	t *testing.T,
) {
	repository, mock, cleanup :=
		newMessagesRepositoryTest(t)
	defer cleanup()

	mock.ExpectQuery(
		`(?s)SELECT.*FROM messages.*WHERE job_id = \$1.*ORDER BY created_at ASC`,
	).
		WithArgs(uint(10)).
		WillReturnRows(
			sqlmock.NewRows(messageColumns()),
		)

	messages, err := repository.ListByJobID(
		context.Background(),
		10,
	)

	if err != nil {
		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}

	if len(messages) != 0 {
		t.Fatalf(
			"expected no messages got %d",
			len(messages),
		)
	}
}

func TestSQLRepository_ListByJobID_QueryError(
	t *testing.T,
) {
	repository, mock, cleanup :=
		newMessagesRepositoryTest(t)
	defer cleanup()

	queryErr := errors.New("query failed")

	mock.ExpectQuery(
		`(?s)SELECT.*FROM messages.*WHERE job_id = \$1`,
	).
		WithArgs(uint(10)).
		WillReturnError(queryErr)

	_, err := repository.ListByJobID(
		context.Background(),
		10,
	)

	if !errors.Is(err, queryErr) {
		t.Fatalf(
			"expected query error, got %v",
			err,
		)
	}
}

func TestSQLRepository_ListByJobID_ScanError(
	t *testing.T,
) {
	repository, mock, cleanup :=
		newMessagesRepositoryTest(t)
	defer cleanup()

	now := time.Now()

	mock.ExpectQuery(
		`(?s)SELECT.*FROM messages.*WHERE job_id = \$1`,
	).
		WithArgs(uint(10)).
		WillReturnRows(
			sqlmock.NewRows(messageColumns()).
				AddRow(
					"invalid-id",
					10,
					1,
					2,
					"hello",
					false,
					now,
					now,
				),
		)

	_, err := repository.ListByJobID(
		context.Background(),
		10,
	)

	if err == nil {
		t.Fatal("expected scan error")
	}
}

func TestSQLRepository_ListByJobID_RowsError(
	t *testing.T,
) {
	repository, mock, cleanup :=
		newMessagesRepositoryTest(t)
	defer cleanup()

	rowsErr := errors.New("rows failed")
	now := time.Now()

	rows := sqlmock.NewRows(
		messageColumns(),
	).
		AddRow(
			1,
			10,
			1,
			2,
			"hello",
			false,
			now,
			now,
		).
		RowError(0, rowsErr)

	mock.ExpectQuery(
		`(?s)SELECT.*FROM messages.*WHERE job_id = \$1`,
	).
		WithArgs(uint(10)).
		WillReturnRows(rows)

	_, err := repository.ListByJobID(
		context.Background(),
		10,
	)

	if !errors.Is(err, rowsErr) {
		t.Fatalf(
			"expected rows error got %v",
			err,
		)
	}
}

func TestSQLRepository_IsJobParticipant_True(
	t *testing.T,
) {
	repository, mock, cleanup :=
		newMessagesRepositoryTest(t)
	defer cleanup()

	mock.ExpectQuery(
		`(?s)SELECT EXISTS.*FROM jobs j.*LEFT JOIN applications`,
	).
		WithArgs(
			uint(10),
			uint(1),
		).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{"exists"},
			).AddRow(true),
		)

	exists, err := repository.IsJobParticipant(
		context.Background(),
		10,
		1,
	)

	if err != nil {
		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}

	if !exists {
		t.Fatal(
			"expected participant to exist",
		)
	}
}

func TestSQLRepository_IsJobParticipant_False(
	t *testing.T,
) {
	repository, mock, cleanup :=
		newMessagesRepositoryTest(t)
	defer cleanup()

	mock.ExpectQuery(
		`(?s)SELECT EXISTS.*FROM jobs j.*LEFT JOIN applications`,
	).
		WithArgs(
			uint(10),
			uint(99),
		).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{"exists"},
			).AddRow(false),
		)

	exists, err := repository.IsJobParticipant(
		context.Background(),
		10,
		99,
	)

	if err != nil {
		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}

	if exists {
		t.Fatal(
			"expected participant to be false",
		)
	}
}

func TestSQLRepository_IsJobParticipant_Error(
	t *testing.T,
) {
	repository, mock, cleanup :=
		newMessagesRepositoryTest(t)
	defer cleanup()

	queryErr := errors.New("participant query failed")

	mock.ExpectQuery(
		`(?s)SELECT EXISTS.*FROM jobs j.*LEFT JOIN applications`,
	).
		WithArgs(
			uint(10),
			uint(1),
		).
		WillReturnError(queryErr)

	_, err := repository.IsJobParticipant(
		context.Background(),
		10,
		1,
	)

	if !errors.Is(err, queryErr) {
		t.Fatalf(
			"expected query error got %v",
			err,
		)
	}
}

var _ Repository = (*SQLRepository)(nil)

