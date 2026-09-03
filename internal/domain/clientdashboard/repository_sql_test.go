package clientdashboard

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func newClientDashboardTestDB(
	t *testing.T,
) (*SQLRepository, sqlmock.Sqlmock, func()) {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}

	repo := NewSQLRepository(db)

	cleanup := func() {
		db.Close()
	}

	return repo, mock, cleanup
}

func TestSQLRepository_ListActiveJobs_Success(t *testing.T) {
	repo, mock, cleanup := newClientDashboardTestDB(t)
	defer cleanup()

	now := time.Now().UTC()

	rows := sqlmock.NewRows([]string{
		"id",
		"title",
		"location",
		"job_type",
		"budget",
		"status",
		"application_count",
		"created_at",
	}).
		AddRow(
			11,
			"End of tenancy clean",
			"East London",
			"end_of_tenancy",
			150.00,
			"open",
			6,
			now,
		).
		AddRow(
			12,
			"Airbnb turnover",
			"London",
			"airbnb",
			80.00,
			"open",
			3,
			now,
		)

	mock.ExpectQuery(`SELECT[\s\S]*FROM jobs j`).
		WithArgs(uint(7), 5).
		WillReturnRows(rows)

	got, err := repo.ListActiveJobs(
		context.Background(),
		7,
		5,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(got) != 2 {
		t.Fatalf("expected 2 jobs, got %d", len(got))
	}

	if got[0].ID != 11 {
		t.Fatalf("expected first job ID 11, got %d", got[0].ID)
	}

	if got[0].ApplicationCount != 6 {
		t.Fatalf(
			"expected application count 6, got %d",
			got[0].ApplicationCount,
		)
	}

	if got[0].Budget != 150 {
		t.Fatalf(
			"expected budget 150, got %.2f",
			got[0].Budget,
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestSQLRepository_ListActiveJobs_Empty(t *testing.T) {
	repo, mock, cleanup := newClientDashboardTestDB(t)
	defer cleanup()

	rows := sqlmock.NewRows([]string{
		"id",
		"title",
		"location",
		"job_type",
		"budget",
		"status",
		"application_count",
		"created_at",
	})

	mock.ExpectQuery(`SELECT[\s\S]*FROM jobs j`).
		WithArgs(uint(7), 5).
		WillReturnRows(rows)

	got, err := repo.ListActiveJobs(
		context.Background(),
		7,
		5,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(got) != 0 {
		t.Fatalf("expected 0 jobs, got %d", len(got))
	}
}

func TestSQLRepository_ListActiveJobs_QueryError(t *testing.T) {
	repo, mock, cleanup := newClientDashboardTestDB(t)
	defer cleanup()

	expectedErr := errors.New("query failure")

	mock.ExpectQuery(`SELECT[\s\S]*FROM jobs j`).
		WithArgs(uint(7), 5).
		WillReturnError(expectedErr)

	got, err := repo.ListActiveJobs(
		context.Background(),
		7,
		5,
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}

	if got != nil {
		t.Fatal("expected nil jobs")
	}
}

func TestSQLRepository_ListActiveJobs_ScanError(t *testing.T) {
	repo, mock, cleanup := newClientDashboardTestDB(t)
	defer cleanup()

	rows := sqlmock.NewRows([]string{
		"id",
		"title",
		"location",
		"job_type",
		"budget",
		"status",
		"application_count",
		"created_at",
	}).AddRow(
		"invalid-id",
		"Test job",
		"London",
		"domestic",
		50.00,
		"open",
		2,
		time.Now(),
	)

	mock.ExpectQuery(`SELECT[\s\S]*FROM jobs j`).
		WithArgs(uint(7), 5).
		WillReturnRows(rows)

	got, err := repo.ListActiveJobs(
		context.Background(),
		7,
		5,
	)

	if err == nil {
		t.Fatal("expected scan error")
	}

	if got != nil {
		t.Fatal("expected nil jobs")
	}
}

func TestSQLRepository_ListActiveJobs_RowsError(t *testing.T) {
	repo, mock, cleanup := newClientDashboardTestDB(t)
	defer cleanup()

	expectedErr := errors.New("rows failure")

	rows := sqlmock.NewRows([]string{
		"id",
		"title",
		"location",
		"job_type",
		"budget",
		"status",
		"application_count",
		"created_at",
	}).
		AddRow(
			11,
			"Test job",
			"London",
			"domestic",
			50.00,
			"open",
			2,
			time.Now(),
		).
		RowError(0, expectedErr)

	mock.ExpectQuery(`SELECT[\s\S]*FROM jobs j`).
		WithArgs(uint(7), 5).
		WillReturnRows(rows)

	got, err := repo.ListActiveJobs(
		context.Background(),
		7,
		5,
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}

	if got != nil {
		t.Fatal("expected nil jobs")
	}
}

func TestSQLRepository_ListUpcomingBookings_Success(t *testing.T) {
	repo, mock, cleanup := newClientDashboardTestDB(t)
	defer cleanup()

	scheduledAt := time.Now().UTC().Add(24 * time.Hour)

	rows := sqlmock.NewRows([]string{
		"id",
		"job_id",
		"cleaner_id",
		"status",
		"scheduled_at",
	}).AddRow(
		21,
		11,
		20,
		"confirmed",
		scheduledAt,
	)

	mock.ExpectQuery(`SELECT[\s\S]*FROM bookings`).
		WithArgs(uint(7), 5).
		WillReturnRows(rows)

	got, err := repo.ListUpcomingBookings(
		context.Background(),
		7,
		5,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(got) != 1 {
		t.Fatalf("expected 1 booking, got %d", len(got))
	}

	if got[0].ID != 21 {
		t.Fatalf(
			"expected booking ID 21, got %d",
			got[0].ID,
		)
	}

	if got[0].CleanerID != 20 {
		t.Fatalf(
			"expected cleaner ID 20, got %d",
			got[0].CleanerID,
		)
	}
}

func TestSQLRepository_ListUpcomingBookings_Empty(t *testing.T) {
	repo, mock, cleanup := newClientDashboardTestDB(t)
	defer cleanup()

	rows := sqlmock.NewRows([]string{
		"id",
		"job_id",
		"cleaner_id",
		"status",
		"scheduled_at",
	})

	mock.ExpectQuery(`SELECT[\s\S]*FROM bookings`).
		WithArgs(uint(7), 5).
		WillReturnRows(rows)

	got, err := repo.ListUpcomingBookings(
		context.Background(),
		7,
		5,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(got) != 0 {
		t.Fatalf(
			"expected 0 bookings, got %d",
			len(got),
		)
	}
}

func TestSQLRepository_ListUpcomingBookings_QueryError(t *testing.T) {
	repo, mock, cleanup := newClientDashboardTestDB(t)
	defer cleanup()

	expectedErr := errors.New("query failure")

	mock.ExpectQuery(`SELECT[\s\S]*FROM bookings`).
		WithArgs(uint(7), 5).
		WillReturnError(expectedErr)

	got, err := repo.ListUpcomingBookings(
		context.Background(),
		7,
		5,
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}

	if got != nil {
		t.Fatal("expected nil bookings")
	}
}

func TestSQLRepository_ListUpcomingBookings_ScanError(t *testing.T) {
	repo, mock, cleanup := newClientDashboardTestDB(t)
	defer cleanup()

	rows := sqlmock.NewRows([]string{
		"id",
		"job_id",
		"cleaner_id",
		"status",
		"scheduled_at",
	}).AddRow(
		"invalid-id",
		11,
		20,
		"confirmed",
		time.Now(),
	)

	mock.ExpectQuery(`SELECT[\s\S]*FROM bookings`).
		WithArgs(uint(7), 5).
		WillReturnRows(rows)

	got, err := repo.ListUpcomingBookings(
		context.Background(),
		7,
		5,
	)

	if err == nil {
		t.Fatal("expected scan error")
	}

	if got != nil {
		t.Fatal("expected nil bookings")
	}
}

func TestSQLRepository_ListUpcomingBookings_RowsError(t *testing.T) {
	repo, mock, cleanup := newClientDashboardTestDB(t)
	defer cleanup()

	expectedErr := errors.New("rows failure")

	rows := sqlmock.NewRows([]string{
		"id",
		"job_id",
		"cleaner_id",
		"status",
		"scheduled_at",
	}).
		AddRow(
			21,
			11,
			20,
			"confirmed",
			time.Now(),
		).
		RowError(0, expectedErr)

	mock.ExpectQuery(`SELECT[\s\S]*FROM bookings`).
		WithArgs(uint(7), 5).
		WillReturnRows(rows)

	got, err := repo.ListUpcomingBookings(
		context.Background(),
		7,
		5,
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}

	if got != nil {
		t.Fatal("expected nil bookings")
	}
}

func TestSQLRepository_CountApplicationsReceived_Success(t *testing.T) {
	repo, mock, cleanup := newClientDashboardTestDB(t)
	defer cleanup()

	mock.ExpectQuery(
		regexp.QuoteMeta(`
			SELECT COUNT(a.id)
			FROM applications a
			INNER JOIN jobs j ON j.id = a.job_id
			WHERE j.client_id = $1
		`),
	).
		WithArgs(uint(7)).
		WillReturnRows(
			sqlmock.NewRows([]string{"count"}).
				AddRow(19),
		)

	got, err := repo.CountApplicationsReceived(
		context.Background(),
		7,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got != 19 {
		t.Fatalf("expected 19, got %d", got)
	}
}

func TestSQLRepository_CountApplicationsReceived_Error(t *testing.T) {
	repo, mock, cleanup := newClientDashboardTestDB(t)
	defer cleanup()

	expectedErr := errors.New("database failure")

	mock.ExpectQuery(`SELECT COUNT\(a\.id\)`).
		WithArgs(uint(7)).
		WillReturnError(expectedErr)

	got, err := repo.CountApplicationsReceived(
		context.Background(),
		7,
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}

	if got != 0 {
		t.Fatalf("expected 0, got %d", got)
	}
}

func TestSQLRepository_CountCompletedBookings_Success(t *testing.T) {
	repo, mock, cleanup := newClientDashboardTestDB(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT COUNT \(\*\)[\s\S]*FROM bookings`).
		WithArgs(uint(7)).
		WillReturnRows(
			sqlmock.NewRows([]string{"count"}).
				AddRow(14),
		)

	got, err := repo.CountCompletedBookings(
		context.Background(),
		7,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got != 14 {
		t.Fatalf("expected 14, got %d", got)
	}
}

func TestSQLRepository_CountCompletedBookings_Error(t *testing.T) {
	repo, mock, cleanup := newClientDashboardTestDB(t)
	defer cleanup()

	expectedErr := errors.New("database failure")

	mock.ExpectQuery(`SELECT COUNT \(\*\)[\s\S]*FROM bookings`).
		WithArgs(uint(7)).
		WillReturnError(expectedErr)

	got, err := repo.CountCompletedBookings(
		context.Background(),
		7,
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}

	if got != 0 {
		t.Fatalf("expected 0, got %d", got)
	}
}

func TestSQLRepository_CountFavouriteCleaners_Success(t *testing.T) {
	repo, mock, cleanup := newClientDashboardTestDB(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT COUNT\(\*\)[\s\S]*FROM favorite_cleaners`).
		WithArgs(uint(7)).
		WillReturnRows(
			sqlmock.NewRows([]string{"count"}).
				AddRow(8),
		)

	got, err := repo.CountFavouriteCleaners(
		context.Background(),
		7,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got != 8 {
		t.Fatalf("expected 8, got %d", got)
	}
}

func TestSQLRepository_CountFavouriteCleaners_Error(t *testing.T) {
	repo, mock, cleanup := newClientDashboardTestDB(t)
	defer cleanup()

	expectedErr := errors.New("database failure")

	mock.ExpectQuery(`SELECT COUNT\(\*\)[\s\S]*FROM favorite_cleaners`).
		WithArgs(uint(7)).
		WillReturnError(expectedErr)

	got, err := repo.CountFavouriteCleaners(
		context.Background(),
		7,
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}

	if got != 0 {
		t.Fatalf("expected 0, got %d", got)
	}
}

func TestSQLRepository_CountPreferredCleaners_Success(t *testing.T) {
	repo, mock, cleanup := newClientDashboardTestDB(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT COUNT\(\*\)[\s\S]*FROM preferred_cleaners`).
		WithArgs(uint(7)).
		WillReturnRows(
			sqlmock.NewRows([]string{"count"}).
				AddRow(4),
		)

	got, err := repo.CountPreferredCleaners(
		context.Background(),
		7,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got != 4 {
		t.Fatalf("expected 4, got %d", got)
	}
}

func TestSQLRepository_CountPreferredCleaners_Error(t *testing.T) {
	repo, mock, cleanup := newClientDashboardTestDB(t)
	defer cleanup()

	expectedErr := errors.New("database failure")

	mock.ExpectQuery(`SELECT COUNT\(\*\)[\s\S]*FROM preferred_cleaners`).
		WithArgs(uint(7)).
		WillReturnError(expectedErr)

	got, err := repo.CountPreferredCleaners(
		context.Background(),
		7,
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}

	if got != 0 {
		t.Fatalf("expected 0, got %d", got)
	}
}

func TestSQLRepository_CountRepeatBookings_Success(t *testing.T) {
	repo, mock, cleanup := newClientDashboardTestDB(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT COUNT\(\*\)[\s\S]*FROM repeat_booking_requests`).
		WithArgs(uint(7)).
		WillReturnRows(
			sqlmock.NewRows([]string{"count"}).
				AddRow(6),
		)

	got, err := repo.CountRepeatBookings(
		context.Background(),
		7,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got != 6 {
		t.Fatalf("expected 6, got %d", got)
	}
}

func TestSQLRepository_CountRepeatBookings_Error(t *testing.T) {
	repo, mock, cleanup := newClientDashboardTestDB(t)
	defer cleanup()

	expectedErr := errors.New("database failure")

	mock.ExpectQuery(`SELECT COUNT\(\*\)[\s\S]*FROM repeat_booking_requests`).
		WithArgs(uint(7)).
		WillReturnError(expectedErr)

	got, err := repo.CountRepeatBookings(
		context.Background(),
		7,
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}

	if got != 0 {
		t.Fatalf("expected 0, got %d", got)
	}
}
