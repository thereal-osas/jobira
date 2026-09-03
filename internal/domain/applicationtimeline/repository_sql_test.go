package applicationtimeline

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestSQLRepository_Create_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	actorID := uint(7)
	now := time.Now()

	event := &Event{
		ApplicationID: 10,
		Status:        "pending",
		ActorUserID:   &actorID,
		Note:          "Application submitted",
		CreatedAt:     now,
	}

	query := regexp.QuoteMeta(`
		INSERT INTO application_timeline (
			application_id,
			status,
			actor_user_id,
			note,
			created_at
		)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at
	`)

	mock.ExpectQuery(query).
		WithArgs(
			uint(10),
			"pending",
			&actorID,
			"Application submitted",
			now,
		).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{
					"id",
					"created_at",
				},
			).AddRow(
				1,
				now,
			),
		)

	err = repo.Create(
		context.Background(),
		event,
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if event.ID != 1 {
		t.Fatalf(
			"expected id 1, got %d",
			event.ID,
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf(
			"unmet expectations: %v",
			err,
		)
	}
}

func TestSQLRepository_Create_SetsCreatedAt(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	actorID := uint(7)

	event := &Event{
		ApplicationID: 10,
		Status:        "pending",
		ActorUserID:   &actorID,
	}

	mock.ExpectQuery(
		"INSERT INTO application_timeline",
	).
		WithArgs(
			uint(10),
			"pending",
			&actorID,
			"",
			sqlmock.AnyArg(),
		).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{
					"id",
					"created_at",
				},
			).AddRow(
				1,
				time.Now(),
			),
		)

	err = repo.Create(
		context.Background(),
		event,
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if event.CreatedAt.IsZero() {
		t.Fatal("expected created_at to be set")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf(
			"unmet expectations: %v",
			err,
		)
	}
}

func TestSQLRepository_Create_NilEvent(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	err = repo.Create(
		context.Background(),
		nil,
	)

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf(
			"expected ErrInvalidInput, got %v",
			err,
		)
	}
}

func TestSQLRepository_Create_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	expectedErr := errors.New("insert error")

	event := &Event{
		ApplicationID: 10,
		Status:        "pending",
	}

	mock.ExpectQuery(
		"INSERT INTO application_timeline",
	).
		WithArgs(
			uint(10),
			"pending",
			nil,
			"",
			sqlmock.AnyArg(),
		).
		WillReturnError(expectedErr)

	err = repo.Create(
		context.Background(),
		event,
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected insert error, got %v",
			err,
		)
	}
}

func TestSQLRepository_ListByApplicationID_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	now := time.Now()

	rows := sqlmock.NewRows(
		[]string{
			"id",
			"application_id",
			"status",
			"actor_user_id",
			"note",
			"created_at",
		},
	).
		AddRow(
			1,
			10,
			"pending",
			20,
			"Application submitted",
			now,
		).
		AddRow(
			2,
			10,
			"shortlisted",
			30,
			"Application shortlisted",
			now.Add(time.Minute),
		)

	mock.ExpectQuery(
		"SELECT(.|\\s)*FROM application_timeline",
	).
		WithArgs(uint(10)).
		WillReturnRows(rows)

	events, err := repo.ListByApplicationID(
		context.Background(),
		10,
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(events) != 2 {
		t.Fatalf(
			"expected 2 events, got %d",
			len(events),
		)
	}

	if events[0].ActorUserID == nil {
		t.Fatal("expected actor user id")
	}

	if *events[0].ActorUserID != 20 {
		t.Fatalf(
			"expected actor user id 20, got %d",
			*events[0].ActorUserID,
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf(
			"unmet expectations: %v",
			err,
		)
	}
}

func TestSQLRepository_ListByApplicationID_NullActor(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	rows := sqlmock.NewRows(
		[]string{
			"id",
			"application_id",
			"status",
			"actor_user_id",
			"note",
			"created_at",
		},
	).AddRow(
		1,
		10,
		"pending",
		nil,
		"",
		time.Now(),
	)

	mock.ExpectQuery(
		"SELECT(.|\\s)*FROM application_timeline",
	).
		WithArgs(uint(10)).
		WillReturnRows(rows)

	events, err := repo.ListByApplicationID(
		context.Background(),
		10,
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(events) != 1 {
		t.Fatalf(
			"expected 1 event, got %d",
			len(events),
		)
	}

	if events[0].ActorUserID != nil {
		t.Fatal("expected nil actor user id")
	}
}

func TestSQLRepository_ListByApplicationID_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	mock.ExpectQuery(
		"SELECT(.|\\s)*FROM application_timeline",
	).
		WithArgs(uint(10)).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{
					"id",
					"application_id",
					"status",
					"actor_user_id",
					"note",
					"created_at",
				},
			),
		)

	_, err = repo.ListByApplicationID(
		context.Background(),
		10,
	)

	if !errors.Is(err, ErrTimelineNotFound) {
		t.Fatalf(
			"expected ErrTimelineNotFound, got %v",
			err,
		)
	}
}

func TestSQLRepository_ListByApplicationID_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	expectedErr := errors.New("query error")

	mock.ExpectQuery(
		"SELECT(.|\\s)*FROM application_timeline",
	).
		WithArgs(uint(10)).
		WillReturnError(expectedErr)

	_, err = repo.ListByApplicationID(
		context.Background(),
		10,
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected query error, got %v",
			err,
		)
	}
}

func TestSQLRepository_GetCurrentStatus_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	mock.ExpectQuery(
		"SELECT status(.|\\s)*FROM applications",
	).
		WithArgs(uint(10)).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{"status"},
			).AddRow("accepted"),
		)

	status, err := repo.GetCurrentStatus(
		context.Background(),
		10,
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if status != "accepted" {
		t.Fatalf(
			"expected accepted, got %s",
			status,
		)
	}
}

func TestSQLRepository_GetCurrentStatus_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	mock.ExpectQuery(
		"SELECT status(.|\\s)*FROM applications",
	).
		WithArgs(uint(10)).
		WillReturnError(sql.ErrNoRows)

	_, err = repo.GetCurrentStatus(
		context.Background(),
		10,
	)

	if !errors.Is(err, ErrApplicationNotFound) {
		t.Fatalf(
			"expected ErrApplicationNotFound, got %v",
			err,
		)
	}
}

func TestSQLRepository_GetCurrentStatus_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	expectedErr := errors.New("database error")

	mock.ExpectQuery(
		"SELECT status(.|\\s)*FROM applications",
	).
		WithArgs(uint(10)).
		WillReturnError(expectedErr)

	_, err = repo.GetCurrentStatus(
		context.Background(),
		10,
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected database error, got %v",
			err,
		)
	}
}

func TestSQLRepository_GetApplicationCleanerID_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	mock.ExpectQuery(
		"SELECT cleaner_id(.|\\s)*FROM applications",
	).
		WithArgs(uint(10)).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{"cleaner_id"},
			).AddRow(20),
		)

	cleanerID, err := repo.GetApplicationCleanerID(
		context.Background(),
		10,
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cleanerID != 20 {
		t.Fatalf(
			"expected cleaner id 20, got %d",
			cleanerID,
		)
	}
}

func TestSQLRepository_GetApplicationCleanerID_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	mock.ExpectQuery(
		"SELECT cleaner_id(.|\\s)*FROM applications",
	).
		WithArgs(uint(10)).
		WillReturnError(sql.ErrNoRows)

	_, err = repo.GetApplicationCleanerID(
		context.Background(),
		10,
	)

	if !errors.Is(err, ErrApplicationNotFound) {
		t.Fatalf(
			"expected ErrApplicationNotFound, got %v",
			err,
		)
	}
}

func TestSQLRepository_GetApplicationCleanerID_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	expectedErr := errors.New("database error")

	mock.ExpectQuery(
		"SELECT cleaner_id(.|\\s)*FROM applications",
	).
		WithArgs(uint(10)).
		WillReturnError(expectedErr)

	_, err = repo.GetApplicationCleanerID(
		context.Background(),
		10,
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected database error, got %v",
			err,
		)
	}
}

func TestSQLRepository_GetApplicationClientID_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	mock.ExpectQuery(
		"SELECT j.client_id(.|\\s)*FROM applications a",
	).
		WithArgs(uint(10)).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{"client_id"},
			).AddRow(30),
		)

	clientID, err := repo.GetApplicationClientID(
		context.Background(),
		10,
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if clientID != 30 {
		t.Fatalf(
			"expected client id 30, got %d",
			clientID,
		)
	}
}

func TestSQLRepository_GetApplicationClientID_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	mock.ExpectQuery(
		"SELECT j.client_id(.|\\s)*FROM applications a",
	).
		WithArgs(uint(10)).
		WillReturnError(sql.ErrNoRows)

	_, err = repo.GetApplicationClientID(
		context.Background(),
		10,
	)

	if !errors.Is(err, ErrApplicationNotFound) {
		t.Fatalf(
			"expected ErrApplicationNotFound, got %v",
			err,
		)
	}
}

func TestSQLRepository_GetApplicationClientID_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	expectedErr := errors.New("database error")

	mock.ExpectQuery(
		"SELECT j.client_id(.|\\s)*FROM applications a",
	).
		WithArgs(uint(10)).
		WillReturnError(expectedErr)

	_, err = repo.GetApplicationClientID(
		context.Background(),
		10,
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected database error, got %v",
			err,
		)
	}
}
