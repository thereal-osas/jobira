package cleanerdashboard

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func newCleanerDashboardTestDB(
	t *testing.T,
) (*sql.DB, sqlmock.Sqlmock) {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf(
			"failed to create sqlmock: %v",
			err,
		)
	}

	t.Cleanup(func() {
		_ = db.Close()
	})

	return db, mock
}

func TestSQLRepository_CountFavourites_Success(t *testing.T) {
	db, mock := newCleanerDashboardTestDB(t)

	repo := NewSQLRepository(db)

	mock.ExpectQuery(
		"SELECT COUNT",
	).
		WithArgs(uint(7)).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{"count"},
			).AddRow(8),
		)

	count, err := repo.CountFavourites(
		context.Background(),
		7,
	)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if count != 8 {
		t.Fatalf(
			"expected count 8, got %d",
			count,
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf(
			"unmet expectations: %v",
			err,
		)
	}
}

func TestSQLRepository_CountFavourites_Error(t *testing.T) {
	db, mock := newCleanerDashboardTestDB(t)

	repo := NewSQLRepository(db)

	expectedErr := errors.New("query failed")

	mock.ExpectQuery(
		"SELECT COUNT",
	).
		WithArgs(uint(7)).
		WillReturnError(expectedErr)

	count, err := repo.CountFavourites(
		context.Background(),
		7,
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected query error, got %v",
			err,
		)
	}

	if count != 0 {
		t.Fatalf(
			"expected count 0, got %d",
			count,
		)
	}
}

func TestSQLRepository_CountPreferredClients_Success(t *testing.T) {
	db, mock := newCleanerDashboardTestDB(t)

	repo := NewSQLRepository(db)

	mock.ExpectQuery(
		"SELECT COUNT",
	).
		WithArgs(uint(7)).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{"count"},
			).AddRow(4),
		)

	count, err := repo.CountPreferredClients(
		context.Background(),
		7,
	)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if count != 4 {
		t.Fatalf(
			"expected count 4, got %d",
			count,
		)
	}
}

func TestSQLRepository_CountPreferredClients_Error(t *testing.T) {
	db, mock := newCleanerDashboardTestDB(t)

	repo := NewSQLRepository(db)

	expectedErr := errors.New("query failed")

	mock.ExpectQuery(
		"SELECT COUNT",
	).
		WithArgs(uint(7)).
		WillReturnError(expectedErr)

	count, err := repo.CountPreferredClients(
		context.Background(),
		7,
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected query error, got %v",
			err,
		)
	}

	if count != 0 {
		t.Fatalf(
			"expected count 0, got %d",
			count,
		)
	}
}

func TestSQLRepository_ListUpcomingBookings_Success(t *testing.T) {
	db, mock := newCleanerDashboardTestDB(t)

	repo := NewSQLRepository(db)

	scheduledOne := time.Now().Add(24 * time.Hour)
	scheduledTwo := time.Now().Add(48 * time.Hour)

	mock.ExpectQuery(
		"SELECT[\\s\\S]*FROM bookings",
	).
		WithArgs(
			uint(7),
			5,
		).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{
					"id",
					"job_id",
					"client_id",
					"status",
					"scheduled_at",
				},
			).
				AddRow(
					1,
					10,
					20,
					"confirmed",
					scheduledOne,
				).
				AddRow(
					2,
					11,
					21,
					"pending",
					scheduledTwo,
				),
		)

	bookings, err := repo.ListUpcomingBookings(
		context.Background(),
		7,
		5,
	)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if len(bookings) != 2 {
		t.Fatalf(
			"expected 2 bookings, got %d",
			len(bookings),
		)
	}

	if bookings[0].ID != 1 {
		t.Fatalf(
			"expected first booking ID 1, got %d",
			bookings[0].ID,
		)
	}

	if bookings[0].JobID != 10 {
		t.Fatalf(
			"expected first job ID 10, got %d",
			bookings[0].JobID,
		)
	}

	if bookings[0].ClientID != 20 {
		t.Fatalf(
			"expected first client ID 20, got %d",
			bookings[0].ClientID,
		)
	}

	if bookings[0].Status != "confirmed" {
		t.Fatalf(
			"expected confirmed, got %s",
			bookings[0].Status,
		)
	}

	if !bookings[0].ScheduledAt.Equal(scheduledOne) {
		t.Fatal("expected scheduled time to match")
	}
}

func TestSQLRepository_ListUpcomingBookings_Empty(t *testing.T) {
	db, mock := newCleanerDashboardTestDB(t)

	repo := NewSQLRepository(db)

	mock.ExpectQuery(
		"SELECT[\\s\\S]*FROM bookings",
	).
		WithArgs(
			uint(7),
			5,
		).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{
					"id",
					"job_id",
					"client_id",
					"status",
					"scheduled_at",
				},
			),
		)

	bookings, err := repo.ListUpcomingBookings(
		context.Background(),
		7,
		5,
	)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if len(bookings) != 0 {
		t.Fatalf(
			"expected zero bookings, got %d",
			len(bookings),
		)
	}
}

func TestSQLRepository_ListUpcomingBookings_QueryError(t *testing.T) {
	db, mock := newCleanerDashboardTestDB(t)

	repo := NewSQLRepository(db)

	expectedErr := errors.New("query failed")

	mock.ExpectQuery(
		"SELECT[\\s\\S]*FROM bookings",
	).
		WithArgs(
			uint(7),
			5,
		).
		WillReturnError(expectedErr)

	bookings, err := repo.ListUpcomingBookings(
		context.Background(),
		7,
		5,
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected query error, got %v",
			err,
		)
	}

	if bookings != nil {
		t.Fatal("expected nil bookings")
	}
}

func TestSQLRepository_ListUpcomingBookings_ScanError(t *testing.T) {
	db, mock := newCleanerDashboardTestDB(t)

	repo := NewSQLRepository(db)

	mock.ExpectQuery(
		"SELECT[\\s\\S]*FROM bookings",
	).
		WithArgs(
			uint(7),
			5,
		).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{
					"id",
					"job_id",
					"client_id",
					"status",
					"scheduled_at",
				},
			).AddRow(
				"invalid-id",
				10,
				20,
				"confirmed",
				time.Now(),
			),
		)

	bookings, err := repo.ListUpcomingBookings(
		context.Background(),
		7,
		5,
	)

	if err == nil {
		t.Fatal("expected scan error")
	}

	if bookings != nil {
		t.Fatal("expected nil bookings")
	}
}

func TestSQLRepository_ListUpcomingBookings_RowsError(t *testing.T) {
	db, mock := newCleanerDashboardTestDB(t)

	repo := NewSQLRepository(db)

	expectedErr := errors.New("rows failure")

	rows := sqlmock.NewRows(
		[]string{
			"id",
			"job_id",
			"client_id",
			"status",
			"scheduled_at",
		},
	).
		AddRow(
			1,
			10,
			20,
			"confirmed",
			time.Now(),
		).
		RowError(
			0,
			expectedErr,
		)

	mock.ExpectQuery(
		"SELECT[\\s\\S]*FROM bookings",
	).
		WithArgs(
			uint(7),
			5,
		).
		WillReturnRows(rows)

	bookings, err := repo.ListUpcomingBookings(
		context.Background(),
		7,
		5,
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected rows error, got %v",
			err,
		)
	}

	if bookings != nil {
		t.Fatal("expected nil bookings")
	}
}
