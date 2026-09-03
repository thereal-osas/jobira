package cleaningteam

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func newCleaningTeamSQLMock(
	t *testing.T,
) (
	*SQLRepository,
	sqlmock.Sqlmock,
	func(),
) {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf(
			"failed to create sqlmock: %v",
			err,
		)
	}

	repo := NewSQLRepository(
		db,
	)

	cleanup := func() {
		db.Close()
	}

	return repo, mock, cleanup
}

func cleaningTeamColumns() []string {
	return []string{
		"cleaner_id",
		"cleaner_name",
		"location",
		"services_offered",
		"last_job_type",
		"availability_status",
		"is_verified",
		"is_favorite",
		"is_preferred",
		"private_note",
		"completed_jobs_together",
		"booking_id",
		"last_booked_at",
		"average_rating",
		"total_reviews",
		"reliability_score",
		"badge",
	}
}

func TestSQLRepository_ListTeamMembers_Success(
	t *testing.T,
) {
	repo, mock, cleanup :=
		newCleaningTeamSQLMock(t)

	defer cleanup()

	now := time.Now().UTC()

	rows := sqlmock.NewRows(
		cleaningTeamColumns(),
	).AddRow(
		20,
		"Maria Cleaner",
		"East London",
		"housekeeping, laundry, ironing",
		"housekeeping",
		"available",
		true,
		true,
		true,
		"Excellent with housekeeping",
		6,
		55,
		now,
		4.9,
		30,
		94,
		"Elite Cleaner",
	)

	mock.
		ExpectQuery(`WITH team_cleaners AS`).
		WithArgs(
			uint(5),
		).
		WillReturnRows(
			rows,
		)

	members, err := repo.ListTeamMembers(
		context.Background(),
		5,
	)
	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}

	if len(members) != 1 {
		t.Fatalf(
			"expected one member, got %d",
			len(members),
		)
	}

	member := members[0]

	if member.CleanerID != 20 {
		t.Fatalf(
			"expected cleaner ID 20, got %d",
			member.CleanerID,
		)
	}

	if member.CleanerName != "Maria Cleaner" {
		t.Fatalf(
			"unexpected cleaner name %q",
			member.CleanerName,
		)
	}

	if !member.IsFavorite {
		t.Fatal(
			"expected favorite",
		)
	}

	if !member.IsPreferred {
		t.Fatal(
			"expected preferred",
		)
	}

	if member.CompletedJobsTogether != 6 {
		t.Fatalf(
			"expected 6 completed jobs, got %d",
			member.CompletedJobsTogether,
		)
	}

	if member.LastBookingID == nil ||
		*member.LastBookingID != 55 {
		t.Fatalf(
			"expected booking ID 55, got %+v",
			member.LastBookingID,
		)
	}

	if member.ReliabilityScore != 94 {
		t.Fatalf(
			"expected reliability 94, got %d",
			member.ReliabilityScore,
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLRepository_ListTeamMembers_Empty(
	t *testing.T,
) {
	repo, mock, cleanup :=
		newCleaningTeamSQLMock(t)

	defer cleanup()

	rows := sqlmock.NewRows(
		cleaningTeamColumns(),
	)

	mock.
		ExpectQuery(`WITH team_cleaners AS`).
		WithArgs(
			uint(5),
		).
		WillReturnRows(
			rows,
		)

	members, err := repo.ListTeamMembers(
		context.Background(),
		5,
	)
	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}

	if len(members) != 0 {
		t.Fatalf(
			"expected no members, got %d",
			len(members),
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLRepository_ListTeamMembers_QueryError(
	t *testing.T,
) {
	repo, mock, cleanup :=
		newCleaningTeamSQLMock(t)

	defer cleanup()

	expectedErr := errors.New(
		"query failed",
	)

	mock.
		ExpectQuery(`WITH team_cleaners AS`).
		WithArgs(
			uint(5),
		).
		WillReturnError(
			expectedErr,
		)

	members, err := repo.ListTeamMembers(
		context.Background(),
		5,
	)

	if !errors.Is(
		err,
		expectedErr,
	) {
		t.Fatalf(
			"expected %v, got %v",
			expectedErr,
			err,
		)
	}

	if members != nil {
		t.Fatalf(
			"expected nil members, got %+v",
			members,
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLRepository_ListTeamMembers_ScanError(
	t *testing.T,
) {
	repo, mock, cleanup :=
		newCleaningTeamSQLMock(t)

	defer cleanup()

	rows := sqlmock.NewRows(
		cleaningTeamColumns(),
	).AddRow(
		"bad-id",
		"Maria Cleaner",
		"London",
		"housekeeping",
		"housekeeping",
		"available",
		true,
		true,
		false,
		"",
		1,
		55,
		time.Now(),
		4.9,
		10,
		90,
		"Top Rated",
	)

	mock.
		ExpectQuery(`WITH team_cleaners AS`).
		WithArgs(
			uint(5),
		).
		WillReturnRows(
			rows,
		)

	members, err := repo.ListTeamMembers(
		context.Background(),
		5,
	)

	if err == nil {
		t.Fatal(
			"expected scan error",
		)
	}

	if members != nil {
		t.Fatalf(
			"expected nil members, got %+v",
			members,
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLRepository_ListTeamMembers_RowsError(
	t *testing.T,
) {
	repo, mock, cleanup :=
		newCleaningTeamSQLMock(t)

	defer cleanup()

	expectedErr := errors.New(
		"rows failed",
	)

	rows := sqlmock.NewRows(
		cleaningTeamColumns(),
	).
		AddRow(
			20,
			"Maria Cleaner",
			"London",
			"housekeeping",
			"housekeeping",
			"available",
			true,
			true,
			false,
			"",
			1,
			55,
			time.Now(),
			4.9,
			10,
			90,
			"Top Rated",
		).
		RowError(
			0,
			expectedErr,
		)

	mock.
		ExpectQuery(`WITH team_cleaners AS`).
		WithArgs(
			uint(5),
		).
		WillReturnRows(
			rows,
		)

	members, err := repo.ListTeamMembers(
		context.Background(),
		5,
	)

	if !errors.Is(
		err,
		expectedErr,
	) {
		t.Fatalf(
			"expected %v, got %v",
			expectedErr,
			err,
		)
	}

	if members != nil {
		t.Fatalf(
			"expected nil members, got %+v",
			members,
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
