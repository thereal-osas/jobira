package jobinvitations

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
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	if repo == nil {
		t.Fatal("expected repository")
	}

	if repo.db != db {
		t.Fatal("expected database assigned")
	}
}

func TestSQLRepository_Create_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)
	now := time.Now()

	mock.ExpectQuery(
		regexp.QuoteMeta(`
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
		`),
	).
		WithArgs(
			uint(7),
			uint(5),
			uint(8),
			"sent",
			"Please apply",
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
		).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"id",
				"created_at",
				"updated_at",
			}).AddRow(
				12,
				now,
				now,
			),
		)

	invitation := &JobInvitation{
		JobID:     7,
		ClientID:  5,
		CleanerID: 8,
		Message:   "Please apply",
	}

	err = repo.Create(
		context.Background(),
		invitation,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if invitation.ID != 12 {
		t.Fatalf(
			"expected ID 12, got %d",
			invitation.ID,
		)
	}

	if invitation.Status != "sent" {
		t.Fatalf(
			"expected status sent, got %q",
			invitation.Status,
		)
	}
}

func TestSQLRepository_Create_DatabaseError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)
	expectedErr := errors.New("create failed")

	mock.ExpectQuery(
		regexp.QuoteMeta(
			"INSERT INTO job_invitations",
		),
	).
		WillReturnError(expectedErr)

	err = repo.Create(
		context.Background(),
		&JobInvitation{},
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected %v, got %v",
			expectedErr,
			err,
		)
	}
}

func TestSQLRepository_Exists_True(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	mock.ExpectQuery(
		regexp.QuoteMeta(`
			SELECT EXISTS (
			SELECT 1
			FROM job_invitations
			WHERE job_id = $1
			AND cleaner_id = $2
			)
		`),
	).
		WithArgs(
			uint(7),
			uint(8),
		).
		WillReturnRows(
			sqlmock.NewRows([]string{"exists"}).
				AddRow(true),
		)

	exists, err := repo.Exists(
		context.Background(),
		7,
		8,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !exists {
		t.Fatal("expected exists to be true")
	}
}

func TestSQLRepository_Exists_False(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	mock.ExpectQuery(
		regexp.QuoteMeta("SELECT EXISTS"),
	).
		WithArgs(
			uint(7),
			uint(8),
		).
		WillReturnRows(
			sqlmock.NewRows([]string{"exists"}).
				AddRow(false),
		)

	exists, err := repo.Exists(
		context.Background(),
		7,
		8,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if exists {
		t.Fatal("expected exists to be false")
	}
}

func TestSQLRepository_Exists_DatabaseError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)
	expectedErr := errors.New("exists failed")

	mock.ExpectQuery(
		regexp.QuoteMeta("SELECT EXISTS"),
	).
		WithArgs(
			uint(7),
			uint(8),
		).
		WillReturnError(expectedErr)

	exists, err := repo.Exists(
		context.Background(),
		7,
		8,
	)

	if exists {
		t.Fatal("expected exists to be false")
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected %v, got %v",
			expectedErr,
			err,
		)
	}
}

func TestSQLRepository_GetJobClientID_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	mock.ExpectQuery(
		regexp.QuoteMeta(`
			SELECT client_id
			FROM jobs
			WHERE id = $1
			LIMIT 1
		`),
	).
		WithArgs(uint(7)).
		WillReturnRows(
			sqlmock.NewRows([]string{"client_id"}).
				AddRow(5),
		)

	clientID, err := repo.GetJobClientID(
		context.Background(),
		7,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if clientID != 5 {
		t.Fatalf(
			"expected client ID 5, got %d",
			clientID,
		)
	}
}

func TestSQLRepository_GetJobClientID_DatabaseError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)
	expectedErr := errors.New("job lookup failed")

	mock.ExpectQuery(
		regexp.QuoteMeta("FROM jobs"),
	).
		WithArgs(uint(7)).
		WillReturnError(expectedErr)

	clientID, err := repo.GetJobClientID(
		context.Background(),
		7,
	)

	if clientID != 0 {
		t.Fatalf(
			"expected client ID 0, got %d",
			clientID,
		)
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected %v, got %v",
			expectedErr,
			err,
		)
	}
}

func TestSQLRepository_ListSentByClientID_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)
	now := time.Now()

	rows := sqlmock.NewRows([]string{
		"id",
		"job_id",
		"client_id",
		"cleaner_id",
		"status",
		"message",
		"created_at",
		"updated_at",
	}).
		AddRow(
			1,
			7,
			5,
			8,
			"sent",
			"Please apply",
			now,
			now,
		).
		AddRow(
			2,
			9,
			5,
			10,
			"sent",
			"Another invitation",
			now,
			now,
		)

	mock.ExpectQuery(regexp.QuoteMeta(`
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
	`)).
		WithArgs(uint(5)).
		WillReturnRows(rows)

	invitations, err := repo.ListSentByClientID(
		context.Background(),
		5,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(invitations) != 2 {
		t.Fatalf(
			"expected 2 invitations, got %d",
			len(invitations),
		)
	}

	if invitations[0].ID != 1 {
		t.Fatalf(
			"expected first ID 1, got %d",
			invitations[0].ID,
		)
	}
}

func TestSQLRepository_ListSentByClientID_Empty(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	mock.ExpectQuery(
		regexp.QuoteMeta("FROM job_invitations"),
	).
		WithArgs(uint(5)).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"id",
				"job_id",
				"client_id",
				"cleaner_id",
				"status",
				"message",
				"created_at",
				"updated_at",
			}),
		)

	invitations, err := repo.ListSentByClientID(
		context.Background(),
		5,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(invitations) != 0 {
		t.Fatalf(
			"expected no invitations, got %d",
			len(invitations),
		)
	}
}

func TestSQLRepository_ListSentByClientID_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)
	expectedErr := errors.New("query failed")

	mock.ExpectQuery(
		regexp.QuoteMeta("FROM job_invitations"),
	).
		WithArgs(uint(5)).
		WillReturnError(expectedErr)

	invitations, err := repo.ListSentByClientID(
		context.Background(),
		5,
	)

	if invitations != nil {
		t.Fatalf(
			"expected nil invitations, got %+v",
			invitations,
		)
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
}

func TestSQLRepository_ListSentByClientID_ScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)
	now := time.Now()

	rows := sqlmock.NewRows([]string{
		"id",
		"job_id",
		"client_id",
		"cleaner_id",
		"status",
		"message",
		"created_at",
		"updated_at",
	}).AddRow(
		"invalid-id",
		7,
		5,
		8,
		"sent",
		"Please apply",
		now,
		now,
	)

	mock.ExpectQuery(
		regexp.QuoteMeta("FROM job_invitations"),
	).
		WithArgs(uint(5)).
		WillReturnRows(rows)

	invitations, err := repo.ListSentByClientID(
		context.Background(),
		5,
	)

	if invitations != nil {
		t.Fatalf(
			"expected nil invitations, got %+v",
			invitations,
		)
	}

	if err == nil {
		t.Fatal("expected scan error")
	}
}

func TestSQLRepository_ListSentByClientID_RowsError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)
	now := time.Now()
	expectedErr := errors.New("rows failed")

	rows := sqlmock.NewRows([]string{
		"id",
		"job_id",
		"client_id",
		"cleaner_id",
		"status",
		"message",
		"created_at",
		"updated_at",
	}).
		AddRow(
			1,
			7,
			5,
			8,
			"sent",
			"Please apply",
			now,
			now,
		).
		RowError(0, expectedErr)

	mock.ExpectQuery(
		regexp.QuoteMeta("FROM job_invitations"),
	).
		WithArgs(uint(5)).
		WillReturnRows(rows)

	invitations, err := repo.ListSentByClientID(
		context.Background(),
		5,
	)

	if invitations != nil {
		t.Fatalf(
			"expected nil invitations, got %+v",
			invitations,
		)
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
}

func TestSQLRepository_ListReceivedByCleanerID_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)
	now := time.Now()

	rows := sqlmock.NewRows([]string{
		"id",
		"job_id",
		"client_id",
		"cleaner_id",
		"status",
		"message",
		"created_at",
		"updated_at",
	}).
		AddRow(
			1,
			7,
			5,
			8,
			"sent",
			"Please apply",
			now,
			now,
		)

	mock.ExpectQuery(regexp.QuoteMeta(`
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
	`)).
		WithArgs(uint(8)).
		WillReturnRows(rows)

	invitations, err := repo.ListReceivedByCleanerID(
		context.Background(),
		8,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(invitations) != 1 {
		t.Fatalf(
			"expected 1 invitation, got %d",
			len(invitations),
		)
	}

	if invitations[0].CleanerID != 8 {
		t.Fatalf(
			"expected cleaner ID 8, got %d",
			invitations[0].CleanerID,
		)
	}
}

func TestSQLRepository_ListReceivedByCleanerID_Empty(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	mock.ExpectQuery(
		regexp.QuoteMeta("FROM job_invitations"),
	).
		WithArgs(uint(8)).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"id",
				"job_id",
				"client_id",
				"cleaner_id",
				"status",
				"message",
				"created_at",
				"updated_at",
			}),
		)

	invitations, err := repo.ListReceivedByCleanerID(
		context.Background(),
		8,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(invitations) != 0 {
		t.Fatalf(
			"expected no invitations, got %d",
			len(invitations),
		)
	}
}

func TestSQLRepository_ListReceivedByCleanerID_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)
	expectedErr := errors.New("query failed")

	mock.ExpectQuery(
		regexp.QuoteMeta("FROM job_invitations"),
	).
		WithArgs(uint(8)).
		WillReturnError(expectedErr)

	invitations, err := repo.ListReceivedByCleanerID(
		context.Background(),
		8,
	)

	if invitations != nil {
		t.Fatalf(
			"expected nil invitations, got %+v",
			invitations,
		)
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
}

func TestSQLRepository_ListReceivedByCleanerID_ScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)
	now := time.Now()

	rows := sqlmock.NewRows([]string{
		"id",
		"job_id",
		"client_id",
		"cleaner_id",
		"status",
		"message",
		"created_at",
		"updated_at",
	}).AddRow(
		"invalid-id",
		7,
		5,
		8,
		"sent",
		"Please apply",
		now,
		now,
	)

	mock.ExpectQuery(
		regexp.QuoteMeta("FROM job_invitations"),
	).
		WithArgs(uint(8)).
		WillReturnRows(rows)

	invitations, err := repo.ListReceivedByCleanerID(
		context.Background(),
		8,
	)

	if invitations != nil {
		t.Fatalf(
			"expected nil invitations, got %+v",
			invitations,
		)
	}

	if err == nil {
		t.Fatal("expected scan error")
	}
}

func TestSQLRepository_ListReceivedByCleanerID_RowsError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)
	now := time.Now()
	expectedErr := errors.New("rows failed")

	rows := sqlmock.NewRows([]string{
		"id",
		"job_id",
		"client_id",
		"cleaner_id",
		"status",
		"message",
		"created_at",
		"updated_at",
	}).
		AddRow(
			1,
			7,
			5,
			8,
			"sent",
			"Please apply",
			now,
			now,
		).
		RowError(0, expectedErr)

	mock.ExpectQuery(
		regexp.QuoteMeta("FROM job_invitations"),
	).
		WithArgs(uint(8)).
		WillReturnRows(rows)

	invitations, err := repo.ListReceivedByCleanerID(
		context.Background(),
		8,
	)

	if invitations != nil {
		t.Fatalf(
			"expected nil invitations, got %+v",
			invitations,
		)
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
}
