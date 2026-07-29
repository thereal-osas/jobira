package applications

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func newRepositoryMock(t *testing.T) (*SQLRepository, sqlmock.Sqlmock) {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}

	t.Cleanup(func() {
		_ = db.Close()
	})

	return NewSQLRepository(db), mock
}

func applicationColumnNames() []string {
	return []string{
		"id",
		"job_id",
		"cleaner_id",
		"cover_message",
		"proposed_rate",
		"status",
		"created_at",
		"updated_at",
	}
}

func TestNewSQLRepository(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create SQL mock: %v", err)
	}
	defer db.Close()

	repository := NewSQLRepository(db)

	if repository == nil {
		t.Fatal("expected repository, got nil")
	}

	if repository.db != db {
		t.Fatal("expected repository to contain supplied database")
	}
}

func TestSQLRepository_Create_Success(t *testing.T) {
	repository, mock := newRepositoryMock(t)

	application := &Application{
		JobID:        10,
		CleanerID:    20,
		CoverMessage: "I am available for this job.",
		ProposedRate: 75,
	}

	query := regexp.QuoteMeta(`
		INSERT INTO applications (
			job_id,
			cleaner_id,
			cover_message,
			proposed_rate,
			status,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`)

	mock.ExpectQuery(query).
		WithArgs(
			application.JobID,
			application.CleanerID,
			application.CoverMessage,
			application.ProposedRate,
			"pending",
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
		).
		WillReturnRows(
			sqlmock.NewRows([]string{"id"}).AddRow(1),
		)

	err := repository.Create(context.Background(), application)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if application.ID != 1 {
		t.Fatalf("expected application ID 1, got %d", application.ID)
	}

	if application.Status != "pending" {
		t.Fatalf("expected pending status, got %q", application.Status)
	}

	if application.CreatedAt.IsZero() {
		t.Fatal("expected CreatedAt to be populated")
	}

	if application.UpdatedAt.IsZero() {
		t.Fatal("expected UpdatedAt to be populated")
	}

	if !application.CreatedAt.Equal(application.UpdatedAt) {
		t.Fatal("expected CreatedAt and UpdatedAt to use the same timestamp")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestSQLRepository_Create_AlreadyApplied(t *testing.T) {
	repository, mock := newRepositoryMock(t)

	application := &Application{
		JobID:        10,
		CleanerID:    20,
		CoverMessage: "Application message",
		ProposedRate: 75,
	}

	mock.ExpectQuery(regexp.QuoteMeta(`
		INSERT INTO applications (
			job_id,
			cleaner_id,
			cover_message,
			proposed_rate,
			status,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`)).
		WithArgs(
			application.JobID,
			application.CleanerID,
			application.CoverMessage,
			application.ProposedRate,
			"pending",
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
		).
		WillReturnError(
			errors.New("duplicate key value violates unique constraint"),
		)

	err := repository.Create(context.Background(), application)

	if !errors.Is(err, ErrAlreadyApplied) {
		t.Fatalf("expected ErrAlreadyApplied, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestSQLRepository_Create_DatabaseError(t *testing.T) {
	repository, mock := newRepositoryMock(t)

	expectedErr := errors.New("database unavailable")

	application := &Application{
		JobID:        10,
		CleanerID:    20,
		CoverMessage: "Application message",
		ProposedRate: 75,
	}

	mock.ExpectQuery(regexp.QuoteMeta(`
		INSERT INTO applications (
			job_id,
			cleaner_id,
			cover_message,
			proposed_rate,
			status,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`)).
		WithArgs(
			application.JobID,
			application.CleanerID,
			application.CoverMessage,
			application.ProposedRate,
			"pending",
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
		).
		WillReturnError(expectedErr)

	err := repository.Create(context.Background(), application)

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestSQLRepository_GetByID_Success(t *testing.T) {
	repository, mock := newRepositoryMock(t)

	createdAt := time.Now().Add(-time.Hour)
	updatedAt := time.Now()

	query := regexp.QuoteMeta(`
		SELECT
			id,
			job_id,
			cleaner_id,
			cover_message,
			proposed_rate,
			status,
			created_at,
			updated_at
		FROM applications
		WHERE id = $1
		LIMIT 1
	`)

	mock.ExpectQuery(query).
		WithArgs(uint(1)).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"id",
				"job_id",
				"cleaner_id",
				"cover_message",
				"proposed_rate",
				"status",
				"created_at",
				"updated_at",
			}).AddRow(
				1,
				10,
				20,
				"I am available for this job.",
				75,
				"pending",
				createdAt,
				updatedAt,
			),
		)

	application, err := repository.GetByID(context.Background(), 1)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if application == nil {
		t.Fatal("expected application, got nil")
	}

	if application.ID != 1 {
		t.Fatalf("expected application ID 1, got %d", application.ID)
	}

	if application.JobID != 10 {
		t.Fatalf("expected job ID 10, got %d", application.JobID)
	}

	if application.CleanerID != 20 {
		t.Fatalf("expected cleaner ID 20, got %d", application.CleanerID)
	}

	if application.CoverMessage != "I am available for this job." {
		t.Fatalf(
			"expected cover message %q, got %q",
			"I am available for this job.",
			application.CoverMessage,
		)
	}

	if application.ProposedRate != 75 {
		t.Fatalf("expected proposed rate 75, got %v", application.ProposedRate)
	}

	if application.Status != "pending" {
		t.Fatalf("expected pending status, got %q", application.Status)
	}

	if !application.CreatedAt.Equal(createdAt) {
		t.Fatalf(
			"expected CreatedAt %v, got %v",
			createdAt,
			application.CreatedAt,
		)
	}

	if !application.UpdatedAt.Equal(updatedAt) {
		t.Fatalf(
			"expected UpdatedAt %v, got %v",
			updatedAt,
			application.UpdatedAt,
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestSQLRepository_GetByID_NotFound(t *testing.T) {
	repository, mock := newRepositoryMock(t)

	query := regexp.QuoteMeta(`
		SELECT
			id,
			job_id,
			cleaner_id,
			cover_message,
			proposed_rate,
			status,
			created_at,
			updated_at
		FROM applications
		WHERE id = $1
		LIMIT 1
	`)

	mock.ExpectQuery(query).
		WithArgs(uint(999)).
		WillReturnError(sql.ErrNoRows)

	application, err := repository.GetByID(context.Background(), 999)

	if application != nil {
		t.Fatalf("expected nil application, got %+v", application)
	}

	if !errors.Is(err, ErrApplicationNotFound) {
		t.Fatalf("expected ErrApplicationNotFound, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestSQLRepository_GetByID_DatabaseError(t *testing.T) {
	repository, mock := newRepositoryMock(t)

	expectedErr := errors.New("database unavailable")

	query := regexp.QuoteMeta(`
		SELECT
			id,
			job_id,
			cleaner_id,
			cover_message,
			proposed_rate,
			status,
			created_at,
			updated_at
		FROM applications
		WHERE id = $1
		LIMIT 1
	`)

	mock.ExpectQuery(query).
		WithArgs(uint(1)).
		WillReturnError(expectedErr)

	application, err := repository.GetByID(context.Background(), 1)

	if application != nil {
		t.Fatalf("expected nil application, got %+v", application)
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestSQLRepository_ListByJobID_Success(t *testing.T) {
	repository, mock := newRepositoryMock(t)

	createdAt := time.Now().Add(-time.Hour)
	updatedAt := time.Now()

	query := regexp.QuoteMeta(`
		SELECT
			id,
			job_id,
			cleaner_id,
			cover_message,
			proposed_rate,
			status,
			created_at,
			updated_at
		FROM applications
		WHERE job_id = $1
		ORDER BY created_at DESC
	`)

	rows := sqlmock.NewRows(applicationColumnNames()).
		AddRow(
			1,
			10,
			20,
			"First application",
			75,
			"pending",
			createdAt,
			updatedAt,
		).
		AddRow(
			2,
			10,
			21,
			"Second application",
			80,
			"accepted",
			createdAt.Add(-time.Minute),
			updatedAt,
		)

	mock.ExpectQuery(query).
		WithArgs(uint(10)).
		WillReturnRows(rows)

	applications, err := repository.ListByJobID(context.Background(), 10)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if len(applications) != 2 {
		t.Fatalf("expected 2 applications, got %d", len(applications))
	}

	if applications[0].ID != 1 {
		t.Fatalf("expected first application ID 1, got %d", applications[0].ID)
	}

	if applications[0].JobID != 10 {
		t.Fatalf("expected job ID 10, got %d", applications[0].JobID)
	}

	if applications[1].CleanerID != 21 {
		t.Fatalf(
			"expected second cleaner ID 21, got %d",
			applications[1].CleanerID,
		)
	}

	if applications[1].Status != "accepted" {
		t.Fatalf(
			"expected accepted status, got %q",
			applications[1].Status,
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestSQLRepository_ListByJobID_Empty(t *testing.T) {
	repository, mock := newRepositoryMock(t)

	query := regexp.QuoteMeta(`
		SELECT
			id,
			job_id,
			cleaner_id,
			cover_message,
			proposed_rate,
			status,
			created_at,
			updated_at
		FROM applications
		WHERE job_id = $1
		ORDER BY created_at DESC
	`)

	mock.ExpectQuery(query).
		WithArgs(uint(10)).
		WillReturnRows(
			sqlmock.NewRows(applicationColumnNames()),
		)

	applications, err := repository.ListByJobID(context.Background(), 10)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if len(applications) != 0 {
		t.Fatalf("expected no applications, got %d", len(applications))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestSQLRepository_ListByJobID_QueryError(t *testing.T) {
	repository, mock := newRepositoryMock(t)

	expectedErr := errors.New("database unavailable")

	query := regexp.QuoteMeta(`
		SELECT
			id,
			job_id,
			cleaner_id,
			cover_message,
			proposed_rate,
			status,
			created_at,
			updated_at
		FROM applications
		WHERE job_id = $1
		ORDER BY created_at DESC
	`)

	mock.ExpectQuery(query).
		WithArgs(uint(10)).
		WillReturnError(expectedErr)

	applications, err := repository.ListByJobID(context.Background(), 10)

	if applications != nil {
		t.Fatalf("expected nil applications, got %+v", applications)
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestSQLRepository_ListByJobID_ScanError(t *testing.T) {
	repository, mock := newRepositoryMock(t)

	query := regexp.QuoteMeta(`
		SELECT
			id,
			job_id,
			cleaner_id,
			cover_message,
			proposed_rate,
			status,
			created_at,
			updated_at
		FROM applications
		WHERE job_id = $1
		ORDER BY created_at DESC
	`)

	rows := sqlmock.NewRows(applicationColumnNames()).
		AddRow(
			"invalid-id",
			10,
			20,
			"Application",
			75,
			"pending",
			time.Now(),
			time.Now(),
		)

	mock.ExpectQuery(query).
		WithArgs(uint(10)).
		WillReturnRows(rows)

	applications, err := repository.ListByJobID(context.Background(), 10)

	if applications != nil {
		t.Fatalf("expected nil applications, got %+v", applications)
	}

	if err == nil {
		t.Fatal("expected scan error, got nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestSQLRepository_ListByJobID_RowsError(t *testing.T) {
	repository, mock := newRepositoryMock(t)

	expectedErr := errors.New("row iteration failed")

	query := regexp.QuoteMeta(`
		SELECT
			id,
			job_id,
			cleaner_id,
			cover_message,
			proposed_rate,
			status,
			created_at,
			updated_at
		FROM applications
		WHERE job_id = $1
		ORDER BY created_at DESC
	`)

	rows := sqlmock.NewRows(applicationColumnNames()).
		AddRow(
			1,
			10,
			20,
			"First application",
			75,
			"pending",
			time.Now(),
			time.Now(),
		).
		AddRow(
			2,
			10,
			21,
			"Second application",
			80,
			"accepted",
			time.Now(),
			time.Now(),
		).
		RowError(1, expectedErr)

	mock.ExpectQuery(query).
		WithArgs(uint(10)).
		WillReturnRows(rows)

	applications, err := repository.ListByJobID(context.Background(), 10)

	if applications != nil {
		t.Fatalf("expected nil applications, got %+v", applications)
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestSQLRepository_ListByCleanerID_Success(t *testing.T) {
	repository, mock := newRepositoryMock(t)

	query := regexp.QuoteMeta(`
		SELECT
			id,
			job_id,
			cleaner_id,
			cover_message,
			proposed_rate,
			status,
			created_at,
			updated_at
		FROM applications
		WHERE cleaner_id = $1
		ORDER BY created_at DESC
	`)

	rows := sqlmock.NewRows(applicationColumnNames()).
		AddRow(1, 10, 20, "First application", 75, "pending", time.Now(), time.Now()).
		AddRow(2, 11, 20, "Second application", 90, "accepted", time.Now(), time.Now())

	mock.ExpectQuery(query).
		WithArgs(uint(20)).
		WillReturnRows(rows)

	applications, err := repository.ListByCleanerID(context.Background(), 20)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if len(applications) != 2 {
		t.Fatalf("expected 2 applications, got %d", len(applications))
	}

	if applications[0].CleanerID != 20 {
		t.Fatalf("expected cleaner ID 20, got %d", applications[0].CleanerID)
	}

	if applications[1].JobID != 11 {
		t.Fatalf("expected job ID 11, got %d", applications[1].JobID)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLRepository_ListByCleanerID_Empty(t *testing.T) {
	repository, mock := newRepositoryMock(t)

	query := regexp.QuoteMeta(`
		SELECT
			id,
			job_id,
			cleaner_id,
			cover_message,
			proposed_rate,
			status,
			created_at,
			updated_at
		FROM applications
		WHERE cleaner_id = $1
		ORDER BY created_at DESC
	`)

	mock.ExpectQuery(query).
		WithArgs(uint(20)).
		WillReturnRows(sqlmock.NewRows(applicationColumnNames()))

	applications, err := repository.ListByCleanerID(context.Background(), 20)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if len(applications) != 0 {
		t.Fatalf("expected empty slice, got %d", len(applications))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLRepository_ListByCleanerID_QueryError(t *testing.T) {
	repository, mock := newRepositoryMock(t)

	expectedErr := errors.New("database unavailable")

	query := regexp.QuoteMeta(`
		SELECT
			id,
			job_id,
			cleaner_id,
			cover_message,
			proposed_rate,
			status,
			created_at,
			updated_at
		FROM applications
		WHERE cleaner_id = $1
		ORDER BY created_at DESC
	`)

	mock.ExpectQuery(query).
		WithArgs(uint(20)).
		WillReturnError(expectedErr)

	_, err := repository.ListByCleanerID(context.Background(), 20)

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v got %v", expectedErr, err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLRepository_ListByCleanerID_ScanError(t *testing.T) {
	repository, mock := newRepositoryMock(t)

	query := regexp.QuoteMeta(`
		SELECT
			id,
			job_id,
			cleaner_id,
			cover_message,
			proposed_rate,
			status,
			created_at,
			updated_at
		FROM applications
		WHERE cleaner_id = $1
		ORDER BY created_at DESC
	`)

	rows := sqlmock.NewRows(applicationColumnNames()).
		AddRow("bad-id", 10, 20, "Application", 75, "pending", time.Now(), time.Now())

	mock.ExpectQuery(query).
		WithArgs(uint(20)).
		WillReturnRows(rows)

	_, err := repository.ListByCleanerID(context.Background(), 20)

	if err == nil {
		t.Fatal("expected scan error")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLRepository_ListByCleanerID_RowsError(t *testing.T) {
	repository, mock := newRepositoryMock(t)

	expectedErr := errors.New("rows error")

	query := regexp.QuoteMeta(`
		SELECT
			id,
			job_id,
			cleaner_id,
			cover_message,
			proposed_rate,
			status,
			created_at,
			updated_at
		FROM applications
		WHERE cleaner_id = $1
		ORDER BY created_at DESC
	`)

	rows := sqlmock.NewRows(applicationColumnNames()).
		AddRow(1, 10, 20, "Application", 75, "pending", time.Now(), time.Now()).
		RowError(0, expectedErr)

	mock.ExpectQuery(query).
		WithArgs(uint(20)).
		WillReturnRows(rows)

	_, err := repository.ListByCleanerID(context.Background(), 20)

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v got %v", expectedErr, err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLRepository_UpdateStatus_Success(t *testing.T) {
	repository, mock := newRepositoryMock(t)

	query := regexp.QuoteMeta(`
		UPDATE applications
		SET status = $1, updated_at = $2
		WHERE id = $3
	`)

	mock.ExpectExec(query).
		WithArgs("accepted", sqlmock.AnyArg(), uint(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repository.UpdateStatus(context.Background(), 1, "accepted")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestSQLRepository_UpdateStatus_NotFound(t *testing.T) {
	repository, mock := newRepositoryMock(t)

	query := regexp.QuoteMeta(`
		UPDATE applications
		SET status = $1, updated_at = $2
		WHERE id = $3
	`)

	mock.ExpectExec(query).
		WithArgs("rejected", sqlmock.AnyArg(), uint(999)).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := repository.UpdateStatus(context.Background(), 999, "rejected")

	if !errors.Is(err, ErrApplicationNotFound) {
		t.Fatalf("expected ErrApplicationNotFound, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestSQLRepository_UpdateStatus_ExecError(t *testing.T) {
	repository, mock := newRepositoryMock(t)

	expectedErr := errors.New("update failed")

	query := regexp.QuoteMeta(`
		UPDATE applications
		SET status = $1, updated_at = $2
		WHERE id = $3
	`)

	mock.ExpectExec(query).
		WithArgs("accepted", sqlmock.AnyArg(), uint(1)).
		WillReturnError(expectedErr)

	err := repository.UpdateStatus(context.Background(), 1, "accepted")

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestSQLRepository_UpdateStatus_RowsAffectedError(t *testing.T) {
	repository, mock := newRepositoryMock(t)

	expectedErr := errors.New("rows affected failed")

	query := regexp.QuoteMeta(`
		UPDATE applications
		SET status = $1, updated_at = $2
		WHERE id = $3
	`)

	mock.ExpectExec(query).
		WithArgs("accepted", sqlmock.AnyArg(), uint(1)).
		WillReturnResult(sqlmock.NewErrorResult(expectedErr))

	err := repository.UpdateStatus(context.Background(), 1, "accepted")

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestSQLRepository_GetJobClientID_Success(t *testing.T) {
	repository, mock := newRepositoryMock(t)

	query := regexp.QuoteMeta(`
		SELECT client_id
		FROM jobs
		WHERE id = $1
		LIMIT 1
	`)

	mock.ExpectQuery(query).
		WithArgs(uint(10)).
		WillReturnRows(
			sqlmock.NewRows([]string{"client_id"}).
				AddRow(50),
		)

	clientID, err := repository.GetJobClientID(context.Background(), 10)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if clientID != 50 {
		t.Fatalf("expected client ID 50, got %d", clientID)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestSQLRepository_GetJobClientID_NotFound(t *testing.T) {
	repository, mock := newRepositoryMock(t)

	query := regexp.QuoteMeta(`
		SELECT client_id
		FROM jobs
		WHERE id = $1
		LIMIT 1
	`)

	mock.ExpectQuery(query).
		WithArgs(uint(999)).
		WillReturnError(sql.ErrNoRows)

	clientID, err := repository.GetJobClientID(context.Background(), 999)

	if clientID != 0 {
		t.Fatalf("expected client ID 0, got %d", clientID)
	}

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestSQLRepository_GetJobClientID_DatabaseError(t *testing.T) {
	repository, mock := newRepositoryMock(t)

	expectedErr := errors.New("database unavailable")

	query := regexp.QuoteMeta(`
		SELECT client_id
		FROM jobs
		WHERE id = $1
		LIMIT 1
	`)

	mock.ExpectQuery(query).
		WithArgs(uint(10)).
		WillReturnError(expectedErr)

	clientID, err := repository.GetJobClientID(context.Background(), 10)

	if clientID != 0 {
		t.Fatalf("expected client ID 0, got %d", clientID)
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestSQLRepository_GetJobTitle_Success(t *testing.T) {
	repository, mock := newRepositoryMock(t)

	query := regexp.QuoteMeta(`
		SELECT title
		FROM jobs
		WHERE id = $1
		LIMIT 1
	`)

	mock.ExpectQuery(query).
		WithArgs(uint(10)).
		WillReturnRows(
			sqlmock.NewRows([]string{"title"}).
				AddRow("End of tenancy clean"),
		)

	title, err := repository.GetJobTitle(context.Background(), 10)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if title != "End of tenancy clean" {
		t.Fatalf("expected job title %q, got %q", "End of tenancy clean", title)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestSQLRepository_GetJobTitle_Error(t *testing.T) {
	repository, mock := newRepositoryMock(t)

	expectedErr := errors.New("title lookup failed")

	query := regexp.QuoteMeta(`
		SELECT title
		FROM jobs
		WHERE id = $1
		LIMIT 1
	`)

	mock.ExpectQuery(query).
		WithArgs(uint(10)).
		WillReturnError(expectedErr)

	title, err := repository.GetJobTitle(context.Background(), 10)

	if title != "" {
		t.Fatalf("expected empty title, got %q", title)
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestSQLRepository_GetUserEmail_Success(t *testing.T) {
	repository, mock := newRepositoryMock(t)

	query := regexp.QuoteMeta(`
		SELECT email
		FROM users
		WHERE id = $1
		LIMIT 1
	`)

	mock.ExpectQuery(query).
		WithArgs(uint(20)).
		WillReturnRows(
			sqlmock.NewRows([]string{"email"}).
				AddRow("cleaner@example.com"),
		)

	email, err := repository.GetUserEmail(context.Background(), 20)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if email != "cleaner@example.com" {
		t.Fatalf("expected cleaner@example.com, got %q", email)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestSQLRepository_GetUserEmail_Error(t *testing.T) {
	repository, mock := newRepositoryMock(t)

	expectedErr := errors.New("email lookup failed")

	query := regexp.QuoteMeta(`
		SELECT email
		FROM users
		WHERE id = $1
		LIMIT 1
	`)

	mock.ExpectQuery(query).
		WithArgs(uint(20)).
		WillReturnError(expectedErr)

	email, err := repository.GetUserEmail(context.Background(), 20)

	if email != "" {
		t.Fatalf("expected empty email, got %q", email)
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestSQLRepository_GetCleanerEmailByID_Success(t *testing.T) {
	repository, mock := newRepositoryMock(t)

	query := regexp.QuoteMeta(`
		SELECT email
		FROM users
		WHERE id = $1
		LIMIT 1
	`)

	mock.ExpectQuery(query).
		WithArgs(uint(20)).
		WillReturnRows(
			sqlmock.NewRows([]string{"email"}).
				AddRow("cleaner@example.com"),
		)

	email, err := repository.GetCleanerEmailByID(context.Background(), 20)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if email != "cleaner@example.com" {
		t.Fatalf("expected cleaner@example.com, got %q", email)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestSQLRepository_GetCleanerEmailByID_Error(t *testing.T) {
	repository, mock := newRepositoryMock(t)

	expectedErr := errors.New("cleaner email lookup failed")

	query := regexp.QuoteMeta(`
		SELECT email
		FROM users
		WHERE id = $1
		LIMIT 1
	`)

	mock.ExpectQuery(query).
		WithArgs(uint(20)).
		WillReturnError(expectedErr)

	email, err := repository.GetCleanerEmailByID(context.Background(), 20)

	if email != "" {
		t.Fatalf("expected empty email, got %q", email)
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}
