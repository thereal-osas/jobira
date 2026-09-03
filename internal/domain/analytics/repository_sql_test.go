package analytics

import (
	"context"
	"errors"
	"testing"

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

func TestSQLRepository_TotalUsers_Success(
	t *testing.T,
) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	mock.ExpectQuery(
		`SELECT COUNT \(\*\).*FROM users`,
	).WillReturnRows(
		sqlmock.NewRows([]string{"count"}).
			AddRow(100),
	)

	count, err := repo.TotalUsers(
		context.Background(),
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if count != 100 {
		t.Fatalf(
			"expected 100, got %d",
			count,
		)
	}
}

func TestSQLRepository_TotalUsers_Error(
	t *testing.T,
) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)
	expectedErr := errors.New("query failed")

	mock.ExpectQuery(
		`SELECT COUNT \(\*\).*FROM users`,
	).WillReturnError(expectedErr)

	count, err := repo.TotalUsers(
		context.Background(),
	)

	if count != 0 {
		t.Fatalf(
			"expected 0, got %d",
			count,
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

func TestSQLRepository_UsersByRole_Success(
	t *testing.T,
) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	mock.ExpectQuery(
		`SELECT COUNT \(\*\).*FROM users.*WHERE role = \$1`,
	).
		WithArgs("cleaner").
		WillReturnRows(
			sqlmock.NewRows([]string{"count"}).
				AddRow(50),
		)

	count, err := repo.UsersByRole(
		context.Background(),
		"cleaner",
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if count != 50 {
		t.Fatalf(
			"expected 50, got %d",
			count,
		)
	}
}

func TestSQLRepository_UsersByRole_Error(
	t *testing.T,
) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)
	expectedErr := errors.New("query failed")

	mock.ExpectQuery(
		`SELECT COUNT \(\*\).*FROM users.*WHERE role = \$1`,
	).
		WithArgs("client").
		WillReturnError(expectedErr)

	_, err = repo.UsersByRole(
		context.Background(),
		"client",
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected %v, got %v",
			expectedErr,
			err,
		)
	}
}

func TestSQLRepository_TotalCompanies_Success(
	t *testing.T,
) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	mock.ExpectQuery(
		`SELECT COUNT\(\*\).*FROM companies`,
	).WillReturnRows(
		sqlmock.NewRows([]string{"count"}).
			AddRow(12),
	)

	count, err := repo.TotalCompanies(
		context.Background(),
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if count != 12 {
		t.Fatalf(
			"expected 12, got %d",
			count,
		)
	}
}

func TestSQLRepository_TotalCompanies_Error(
	t *testing.T,
) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)
	expectedErr := errors.New("query failed")

	mock.ExpectQuery(
		`SELECT COUNT\(\*\).*FROM companies`,
	).WillReturnError(expectedErr)

	_, err = repo.TotalCompanies(
		context.Background(),
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected %v, got %v",
			expectedErr,
			err,
		)
	}
}

func TestSQLRepository_ActiveJobs_Success(
	t *testing.T,
) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	mock.ExpectQuery(
		`SELECT COUNT \(\*\).*FROM jobs.*WHERE status = 'open'`,
	).WillReturnRows(
		sqlmock.NewRows([]string{"count"}).
			AddRow(25),
	)

	count, err := repo.ActiveJobs(
		context.Background(),
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if count != 25 {
		t.Fatalf(
			"expected 25, got %d",
			count,
		)
	}
}

func TestSQLRepository_ActiveJobs_Error(
	t *testing.T,
) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)
	expectedErr := errors.New("query failed")

	mock.ExpectQuery(
		`SELECT COUNT \(\*\).*FROM jobs.*WHERE status = 'open'`,
	).WillReturnError(expectedErr)

	_, err = repo.ActiveJobs(
		context.Background(),
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected %v, got %v",
			expectedErr,
			err,
		)
	}
}

func TestSQLRepository_CompletedBookings_Success(
	t *testing.T,
) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	mock.ExpectQuery(
		`SELECT COUNT \(\*\).*FROM bookings.*WHERE status = 'completed'`,
	).WillReturnRows(
		sqlmock.NewRows([]string{"count"}).
			AddRow(80),
	)

	count, err := repo.CompletedBookings(
		context.Background(),
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if count != 80 {
		t.Fatalf(
			"expected 80, got %d",
			count,
		)
	}
}

func TestSQLRepository_CompletedBookings_Error(
	t *testing.T,
) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)
	expectedErr := errors.New("query failed")

	mock.ExpectQuery(
		`SELECT COUNT \(\*\).*FROM bookings.*WHERE status = 'completed'`,
	).WillReturnError(expectedErr)

	_, err = repo.CompletedBookings(
		context.Background(),
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected %v, got %v",
			expectedErr,
			err,
		)
	}
}

func TestSQLRepository_PendingVerifications_Success(
	t *testing.T,
) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	mock.ExpectQuery(
		`SELECT COUNT \(\*\).*FROM verification_requests.*WHERE status = 'pending'`,
	).WillReturnRows(
		sqlmock.NewRows([]string{"count"}).
			AddRow(7),
	)

	count, err := repo.PendingVerifications(
		context.Background(),
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if count != 7 {
		t.Fatalf(
			"expected 7, got %d",
			count,
		)
	}
}

func TestSQLRepository_PendingVerifications_Error(
	t *testing.T,
) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)
	expectedErr := errors.New("query failed")

	mock.ExpectQuery(
		`SELECT COUNT \(\*\).*FROM verification_requests.*WHERE status = 'pending'`,
	).WillReturnError(expectedErr)

	_, err = repo.PendingVerifications(
		context.Background(),
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected %v, got %v",
			expectedErr,
			err,
		)
	}
}

func TestSQLRepository_OpenReports_Success(
	t *testing.T,
) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	mock.ExpectQuery(
		`SELECT COUNT \(\*\).*FROM cleaner_reports.*WHERE status = 'open'`,
	).WillReturnRows(
		sqlmock.NewRows([]string{"count"}).
			AddRow(3),
	)

	count, err := repo.OpenReports(
		context.Background(),
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if count != 3 {
		t.Fatalf(
			"expected 3, got %d",
			count,
		)
	}
}

func TestSQLRepository_OpenReports_Error(
	t *testing.T,
) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)
	expectedErr := errors.New("query failed")

	mock.ExpectQuery(
		`SELECT COUNT \(\*\).*FROM cleaner_reports.*WHERE status = 'open'`,
	).WillReturnError(expectedErr)
	_, err = repo.OpenReports(
		context.Background(),
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected %v, got %v",
			expectedErr,
			err,
		)
	}
}

func TestSQLRepository_BookingsToday_Success(
	t *testing.T,
) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	mock.ExpectQuery(
		`SELECT COUNT \(\*\).*FROM bookings.*WHERE DATE\(created_at\) = CURRENT_DATE`,
	).WillReturnRows(
		sqlmock.NewRows([]string{"count"}).
			AddRow(12),
	)

	count, err := repo.BookingsToday(
		context.Background(),
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if count != 12 {
		t.Fatalf(
			"expected 12, got %d",
			count,
		)
	}
}

func TestSQLRepository_BookingsToday_Error(
	t *testing.T,
) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)
	expectedErr := errors.New("query failed")

	mock.ExpectQuery(
		`SELECT COUNT \(\*\).*FROM bookings.*WHERE DATE\(created_at\) = CURRENT_DATE`,
	).WillReturnError(expectedErr)

	_, err = repo.BookingsToday(
		context.Background(),
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected %v, got %v",
			expectedErr,
			err,
		)
	}
}
