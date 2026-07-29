package jobs

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func newJobsSQLMock(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}

	t.Cleanup(func() {
		_ = db.Close()
	})

	return db, mock
}

func TestNewSQLRepository(t *testing.T) {
	db, _ := newJobsSQLMock(t)

	repo := NewSQLRepository(db)

	if repo == nil {
		t.Fatal("expected repository")
	}

	if repo.db != db {
		t.Fatal("expected database to be assigned")
	}
}

func TestSQLRepository_Create_Success(t *testing.T) {
	db, mock := newJobsSQLMock(t)
	repo := NewSQLRepository(db)

	job := &Job{
		ClientID:    5,
		Title:       "Cleaner needed",
		Description: "Clean a flat",
		Location:    "London",
		JobType:     "domestic",
		ListingType: "shift",
		Budget:      80,
	}

	mock.ExpectQuery(regexp.QuoteMeta(`
	INSERT INTO jobs (
		client_id,
		title,
		description,
		location,
		job_type,
		listing_type,
		budget,
		status,
		created_at,
		updated_at
	)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	RETURNING id
	`)).
		WithArgs(
			job.ClientID,
			job.Title,
			job.Description,
			job.Location,
			job.JobType,
			job.ListingType,
			job.Budget,
			"open",
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
		).
		WillReturnRows(
			sqlmock.NewRows([]string{"id"}).AddRow(12),
		)

	err := repo.Create(context.Background(), job)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if job.ID != 12 {
		t.Fatalf("expected ID 12, got %d", job.ID)
	}

	if job.Status != "open" {
		t.Fatalf("expected open status, got %q", job.Status)
	}

	if job.CreatedAt.IsZero() {
		t.Fatal("expected created_at to be populated")
	}

	if job.UpdatedAt.IsZero() {
		t.Fatal("expected updated_at to be populated")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestSQLRepository_Create_Error(t *testing.T) {
	db, mock := newJobsSQLMock(t)
	repo := NewSQLRepository(db)

	expected := errors.New("insert failed")

	job := &Job{
		ClientID:    5,
		Title:       "Cleaner needed",
		Description: "Clean a flat",
		Location:    "London",
		JobType:     "domestic",
		ListingType: "shift",
		Budget:      80,
	}

	mock.ExpectQuery("INSERT INTO jobs").
		WithArgs(
			job.ClientID,
			job.Title,
			job.Description,
			job.Location,
			job.JobType,
			job.ListingType,
			job.Budget,
			"open",
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
		).
		WillReturnError(expected)

	err := repo.Create(context.Background(), job)

	if !errors.Is(err, expected) {
		t.Fatalf("expected %v, got %v", expected, err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestSQLRepository_GetByID_Success(t *testing.T) {
	db, mock := newJobsSQLMock(t)
	repo := NewSQLRepository(db)

	now := time.Now()

	rows := sqlmock.NewRows([]string{
		"id",
		"client_id",
		"title",
		"description",
		"location",
		"job_type",
		"listing_type",
		"budget",
		"status",
		"created_at",
		"updated_at",
	}).AddRow(
		12,
		5,
		"Cleaner needed",
		"Clean a flat",
		"London",
		"domestic",
		"shift",
		80,
		"open",
		now,
		now,
	)

	mock.ExpectQuery("SELECT").
		WithArgs(uint(12)).
		WillReturnRows(rows)

	job, err := repo.GetByID(context.Background(), 12)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if job.ID != 12 {
		t.Fatalf("expected ID 12, got %d", job.ID)
	}

	if job.ClientID != 5 {
		t.Fatalf("expected client ID 5, got %d", job.ClientID)
	}

	if job.Title != "Cleaner needed" {
		t.Fatalf("unexpected title %q", job.Title)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestSQLRepository_GetByID_NotFound(t *testing.T) {
	db, mock := newJobsSQLMock(t)
	repo := NewSQLRepository(db)

	mock.ExpectQuery("SELECT").
		WithArgs(uint(99)).
		WillReturnError(sql.ErrNoRows)

	job, err := repo.GetByID(context.Background(), 99)

	if job != nil {
		t.Fatalf("expected nil job, got %+v", job)
	}

	if !errors.Is(err, ErrJobNotFound) {
		t.Fatalf("expected ErrJobNotFound, got %v", err)
	}
}

func TestSQLRepository_GetByID_DatabaseError(t *testing.T) {
	db, mock := newJobsSQLMock(t)
	repo := NewSQLRepository(db)

	expected := errors.New("query failed")

	mock.ExpectQuery("SELECT").
		WithArgs(uint(12)).
		WillReturnError(expected)

	job, err := repo.GetByID(context.Background(), 12)

	if job != nil {
		t.Fatalf("expected nil job, got %+v", job)
	}

	if !errors.Is(err, expected) {
		t.Fatalf("expected %v, got %v", expected, err)
	}
}

func TestSQLRepository_GetByID_ScanError(t *testing.T) {
	db, mock := newJobsSQLMock(t)
	repo := NewSQLRepository(db)

	rows := sqlmock.NewRows([]string{
		"id",
		"client_id",
		"title",
		"description",
		"location",
		"job_type",
		"listing_type",
		"budget",
		"status",
		"created_at",
		"updated_at",
	}).AddRow(
		"not-an-integer",
		5,
		"Cleaner needed",
		"Clean a flat",
		"London",
		"domestic",
		"shift",
		80,
		"open",
		time.Now(),
		time.Now(),
	)

	mock.ExpectQuery("SELECT").
		WithArgs(uint(12)).
		WillReturnRows(rows)

	job, err := repo.GetByID(context.Background(), 12)

	if job != nil {
		t.Fatalf("expected nil job, got %+v", job)
	}

	if err == nil {
		t.Fatal("expected scan error")
	}
}

func TestSQLRepository_List_Success(t *testing.T) {
	db, mock := newJobsSQLMock(t)
	repo := NewSQLRepository(db)

	now := time.Now()

	rows := sqlmock.NewRows([]string{
		"id",
		"client_id",
		"title",
		"description",
		"location",
		"job_type",
		"listing_type",
		"budget",
		"status",
		"created_at",
		"updated_at",
	}).
		AddRow(
			1,
			5,
			"First job",
			"First description",
			"London",
			"domestic",
			"shift",
			50,
			"open",
			now,
			now,
		).
		AddRow(
			2,
			6,
			"Second job",
			"Second description",
			"Essex",
			"commercial",
			"job",
			100,
			"open",
			now,
			now,
		)

	mock.ExpectQuery("SELECT").
		WillReturnRows(rows)

	jobs, err := repo.List(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(jobs) != 2 {
		t.Fatalf("expected 2 jobs, got %d", len(jobs))
	}

	if jobs[0].ID != 1 {
		t.Fatalf("expected first ID 1, got %d", jobs[0].ID)
	}

	if jobs[1].ID != 2 {
		t.Fatalf("expected second ID 2, got %d", jobs[1].ID)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestSQLRepository_List_QueryError(t *testing.T) {
	db, mock := newJobsSQLMock(t)
	repo := NewSQLRepository(db)

	expected := errors.New("query failed")

	mock.ExpectQuery("SELECT").
		WillReturnError(expected)

	jobs, err := repo.List(context.Background())

	if jobs != nil {
		t.Fatalf("expected nil jobs, got %+v", jobs)
	}

	if !errors.Is(err, expected) {
		t.Fatalf("expected %v, got %v", expected, err)
	}
}

func TestSQLRepository_List_ScanError(t *testing.T) {
	db, mock := newJobsSQLMock(t)
	repo := NewSQLRepository(db)

	rows := sqlmock.NewRows([]string{
		"id",
		"client_id",
		"title",
		"description",
		"location",
		"job_type",
		"listing_type",
		"budget",
		"status",
		"created_at",
		"updated_at",
	}).AddRow(
		"invalid-id",
		5,
		"Cleaner needed",
		"Clean flat",
		"London",
		"domestic",
		"shift",
		50,
		"open",
		time.Now(),
		time.Now(),
	)

	mock.ExpectQuery("SELECT").
		WillReturnRows(rows)

	jobs, err := repo.List(context.Background())

	if jobs != nil {
		t.Fatalf("expected nil jobs, got %+v", jobs)
	}

	if err == nil {
		t.Fatal("expected scan error")
	}
}

func TestSQLRepository_List_RowsError(t *testing.T) {
	db, mock := newJobsSQLMock(t)
	repo := NewSQLRepository(db)

	now := time.Now()

	rows := sqlmock.NewRows([]string{
		"id",
		"client_id",
		"title",
		"description",
		"location",
		"job_type",
		"listing_type",
		"budget",
		"status",
		"created_at",
		"updated_at",
	}).
		AddRow(
			1,
			5,
			"Cleaner needed",
			"Clean flat",
			"London",
			"domestic",
			"shift",
			50,
			"open",
			now,
			now,
		).
		RowError(0, errors.New("row iteration failed"))

	mock.ExpectQuery("SELECT").
		WillReturnRows(rows)

	jobs, err := repo.List(context.Background())

	if jobs != nil {
		t.Fatalf("expected nil jobs, got %+v", jobs)
	}

	if err == nil {
		t.Fatal("expected rows error")
	}
}

func TestSQLRepository_ListByClientID_Success(t *testing.T) {
	db, mock := newJobsSQLMock(t)
	repo := NewSQLRepository(db)

	now := time.Now()

	rows := sqlmock.NewRows([]string{
		"id",
		"client_id",
		"title",
		"description",
		"location",
		"job_type",
		"listing_type",
		"budget",
		"status",
		"created_at",
		"updated_at",
	}).AddRow(
		1,
		5,
		"Cleaner needed",
		"Clean flat",
		"London",
		"domestic",
		"shift",
		50,
		"open",
		now,
		now,
	)

	mock.ExpectQuery("SELECT").
		WithArgs(uint(5)).
		WillReturnRows(rows)

	jobs, err := repo.ListByClientID(context.Background(), 5)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(jobs) != 1 {
		t.Fatalf("expected 1 job, got %d", len(jobs))
	}

	if jobs[0].ClientID != 5 {
		t.Fatalf("expected client ID 5, got %d", jobs[0].ClientID)
	}
}

func TestSQLRepository_ListByClientID_QueryError(t *testing.T) {
	db, mock := newJobsSQLMock(t)
	repo := NewSQLRepository(db)

	expected := errors.New("query failed")

	mock.ExpectQuery("SELECT").
		WithArgs(uint(5)).
		WillReturnError(expected)

	jobs, err := repo.ListByClientID(context.Background(), 5)

	if jobs != nil {
		t.Fatalf("expected nil jobs, got %+v", jobs)
	}

	if !errors.Is(err, expected) {
		t.Fatalf("expected %v, got %v", expected, err)
	}
}

func TestSQLRepository_ListByClientID_ScanError(t *testing.T) {
	db, mock := newJobsSQLMock(t)
	repo := NewSQLRepository(db)

	rows := sqlmock.NewRows([]string{
		"id",
		"client_id",
		"title",
		"description",
		"location",
		"job_type",
		"listing_type",
		"budget",
		"status",
		"created_at",
		"updated_at",
	}).AddRow(
		"invalid-id",
		5,
		"Cleaner needed",
		"Clean flat",
		"London",
		"domestic",
		"shift",
		50,
		"open",
		time.Now(),
		time.Now(),
	)

	mock.ExpectQuery("SELECT").
		WithArgs(uint(5)).
		WillReturnRows(rows)

	jobs, err := repo.ListByClientID(context.Background(), 5)

	if jobs != nil {
		t.Fatalf("expected nil jobs, got %+v", jobs)
	}

	if err == nil {
		t.Fatal("expected scan error")
	}
}

func TestSQLRepository_ListByClientID_RowsError(t *testing.T) {
	db, mock := newJobsSQLMock(t)
	repo := NewSQLRepository(db)

	now := time.Now()

	rows := sqlmock.NewRows([]string{
		"id",
		"client_id",
		"title",
		"description",
		"location",
		"job_type",
		"listing_type",
		"budget",
		"status",
		"created_at",
		"updated_at",
	}).
		AddRow(
			1,
			5,
			"Cleaner needed",
			"Clean flat",
			"London",
			"domestic",
			"shift",
			50,
			"open",
			now,
			now,
		).
		RowError(0, errors.New("row iteration failed"))

	mock.ExpectQuery("SELECT").
		WithArgs(uint(5)).
		WillReturnRows(rows)

	jobs, err := repo.ListByClientID(context.Background(), 5)

	if jobs != nil {
		t.Fatalf("expected nil jobs, got %+v", jobs)
	}

	if err == nil {
		t.Fatal("expected rows error")
	}
}

func TestSQLRepository_Update_Success(t *testing.T) {
	db, mock := newJobsSQLMock(t)
	repo := NewSQLRepository(db)

	job := &Job{
		ID:          12,
		Title:       "Updated title",
		Description: "Updated description",
		Location:    "London",
		JobType:     "domestic",
		ListingType: "shift",
		Budget:      100,
		Status:      "open",
	}

	mock.ExpectExec("UPDATE jobs").
		WithArgs(
			job.Title,
			job.Description,
			job.Location,
			job.JobType,
			job.ListingType,
			job.Budget,
			job.Status,
			sqlmock.AnyArg(),
			job.ID,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Update(context.Background(), job)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if job.UpdatedAt.IsZero() {
		t.Fatal("expected updated_at to be populated")
	}
}

func TestSQLRepository_Update_ExecError(t *testing.T) {
	db, mock := newJobsSQLMock(t)
	repo := NewSQLRepository(db)

	expected := errors.New("update failed")

	job := &Job{
		ID:          12,
		Title:       "Updated",
		Description: "Description",
		Location:    "London",
		JobType:     "domestic",
		ListingType: "shift",
		Budget:      80,
		Status:      "open",
	}

	mock.ExpectExec("UPDATE jobs").
		WithArgs(
			job.Title,
			job.Description,
			job.Location,
			job.JobType,
			job.ListingType,
			job.Budget,
			job.Status,
			sqlmock.AnyArg(),
			job.ID,
		).
		WillReturnError(expected)

	err := repo.Update(context.Background(), job)

	if !errors.Is(err, expected) {
		t.Fatalf("expected %v, got %v", expected, err)
	}
}

func TestSQLRepository_Update_RowsAffectedError(t *testing.T) {
	db, mock := newJobsSQLMock(t)
	repo := NewSQLRepository(db)

	expected := errors.New("rows affected failed")

	job := &Job{
		ID:          12,
		Title:       "Updated",
		Description: "Description",
		Location:    "London",
		JobType:     "domestic",
		ListingType: "shift",
		Budget:      80,
		Status:      "open",
	}

	mock.ExpectExec("UPDATE jobs").
		WillReturnResult(sqlmock.NewErrorResult(expected))

	err := repo.Update(context.Background(), job)

	if !errors.Is(err, expected) {
		t.Fatalf("expected %v, got %v", expected, err)
	}
}

func TestSQLRepository_Update_NotFound(t *testing.T) {
	db, mock := newJobsSQLMock(t)
	repo := NewSQLRepository(db)

	job := &Job{
		ID:          99,
		Title:       "Updated",
		Description: "Description",
		Location:    "London",
		JobType:     "domestic",
		ListingType: "shift",
		Budget:      80,
		Status:      "open",
	}

	mock.ExpectExec("UPDATE jobs").
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := repo.Update(context.Background(), job)

	if !errors.Is(err, ErrJobNotFound) {
		t.Fatalf("expected ErrJobNotFound, got %v", err)
	}
}

func TestSQLRepository_Delete_Success(t *testing.T) {
	db, mock := newJobsSQLMock(t)
	repo := NewSQLRepository(db)

	mock.ExpectExec("DELETE FROM jobs").
		WithArgs(uint(12)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Delete(context.Background(), 12)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestSQLRepository_Delete_ExecError(t *testing.T) {
	db, mock := newJobsSQLMock(t)
	repo := NewSQLRepository(db)

	expected := errors.New("delete failed")

	mock.ExpectExec("DELETE FROM jobs").
		WithArgs(uint(12)).
		WillReturnError(expected)

	err := repo.Delete(context.Background(), 12)

	if !errors.Is(err, expected) {
		t.Fatalf("expected %v, got %v", expected, err)
	}
}

func TestSQLRepository_Delete_RowsAffectedError(t *testing.T) {
	db, mock := newJobsSQLMock(t)
	repo := NewSQLRepository(db)

	expected := errors.New("rows affected failed")

	mock.ExpectExec("DELETE FROM jobs").
		WithArgs(uint(12)).
		WillReturnResult(sqlmock.NewErrorResult(expected))

	err := repo.Delete(context.Background(), 12)

	if !errors.Is(err, expected) {
		t.Fatalf("expected %v, got %v", expected, err)
	}
}

func TestSQLRepository_Delete_NotFound(t *testing.T) {
	db, mock := newJobsSQLMock(t)
	repo := NewSQLRepository(db)

	mock.ExpectExec("DELETE FROM jobs").
		WithArgs(uint(99)).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := repo.Delete(context.Background(), 99)

	if !errors.Is(err, ErrJobNotFound) {
		t.Fatalf("expected ErrJobNotFound, got %v", err)
	}
}

func TestSQLRepository_Search_Success(t *testing.T) {
	db, mock := newJobsSQLMock(t)
	repo := NewSQLRepository(db)

	now := time.Now()

	rows := sqlmock.NewRows([]string{
		"id",
		"client_id",
		"title",
		"description",
		"location",
		"job_type",
		"listing_type",
		"budget",
		"status",
		"created_at",
		"updated_at",
	}).AddRow(
		1,
		5,
		"Cleaner needed",
		"Clean flat",
		"London",
		"domestic",
		"shift",
		100,
		"open",
		now,
		now,
	)

	mock.ExpectQuery("SELECT").
		WithArgs(
			"%London%",
			"domestic",
			"shift",
			80,
		).
		WillReturnRows(rows)

	jobs, err := repo.Search(context.Background(), SearchJobRequest{
		Location:    "London",
		JobType:     "domestic",
		ListingType: "shift",
		MinBudget:   80,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(jobs) != 1 {
		t.Fatalf("expected 1 job, got %d", len(jobs))
	}

	if jobs[0].Budget != 100 {
		t.Fatalf("expected budget 100, got %d", jobs[0].Budget)
	}
}

func TestSQLRepository_Search_NoFilters(t *testing.T) {
	db, mock := newJobsSQLMock(t)
	repo := NewSQLRepository(db)

	mock.ExpectQuery("SELECT").
		WillReturnRows(sqlmock.NewRows([]string{
			"id",
			"client_id",
			"title",
			"description",
			"location",
			"job_type",
			"listing_type",
			"budget",
			"status",
			"created_at",
			"updated_at",
		}))

	jobs, err := repo.Search(context.Background(), SearchJobRequest{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(jobs) != 0 {
		t.Fatalf("expected no jobs, got %d", len(jobs))
	}
}

func TestSQLRepository_Search_QueryError(t *testing.T) {
	db, mock := newJobsSQLMock(t)
	repo := NewSQLRepository(db)

	expected := errors.New("search failed")

	mock.ExpectQuery("SELECT").
		WillReturnError(expected)

	jobs, err := repo.Search(context.Background(), SearchJobRequest{})

	if jobs != nil {
		t.Fatalf("expected nil jobs, got %+v", jobs)
	}

	if !errors.Is(err, expected) {
		t.Fatalf("expected %v, got %v", expected, err)
	}
}

func TestSQLRepository_Search_ScanError(t *testing.T) {
	db, mock := newJobsSQLMock(t)
	repo := NewSQLRepository(db)

	rows := sqlmock.NewRows([]string{
		"id",
		"client_id",
		"title",
		"description",
		"location",
		"job_type",
		"listing_type",
		"budget",
		"status",
		"created_at",
		"updated_at",
	}).AddRow(
		"invalid-id",
		5,
		"Cleaner needed",
		"Clean flat",
		"London",
		"domestic",
		"shift",
		100,
		"open",
		time.Now(),
		time.Now(),
	)

	mock.ExpectQuery("SELECT").
		WillReturnRows(rows)

	jobs, err := repo.Search(context.Background(), SearchJobRequest{})

	if jobs != nil {
		t.Fatalf("expected nil jobs, got %+v", jobs)
	}

	if err == nil {
		t.Fatal("expected scan error")
	}
}

func TestSQLRepository_Search_RowsError(t *testing.T) {
	db, mock := newJobsSQLMock(t)
	repo := NewSQLRepository(db)

	now := time.Now()

	rows := sqlmock.NewRows([]string{
		"id",
		"client_id",
		"title",
		"description",
		"location",
		"job_type",
		"listing_type",
		"budget",
		"status",
		"created_at",
		"updated_at",
	}).
		AddRow(
			1,
			5,
			"Cleaner needed",
			"Clean flat",
			"London",
			"domestic",
			"shift",
			100,
			"open",
			now,
			now,
		).
		RowError(0, errors.New("row iteration failed"))

	mock.ExpectQuery("SELECT").
		WillReturnRows(rows)

	jobs, err := repo.Search(context.Background(), SearchJobRequest{})

	if jobs != nil {
		t.Fatalf("expected nil jobs, got %+v", jobs)
	}

	if err == nil {
		t.Fatal("expected rows error")
	}
}
