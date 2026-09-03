package matchscore

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func newMatchScoreSQLMock(
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

func TestSQLRepository_JobBelongsToClient_True(
	t *testing.T,
) {
	repo, mock, cleanup :=
		newMatchScoreSQLMock(t)

	defer cleanup()

	mock.
		ExpectQuery(`SELECT EXISTS`).
		WithArgs(
			uint(10),
			uint(5),
		).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{"exists"},
			).AddRow(
				true,
			),
		)

	exists, err := repo.JobBelongsToClient(
		context.Background(),
		10,
		5,
	)
	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}

	if !exists {
		t.Fatal(
			"expected job to belong to client",
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLRepository_JobBelongsToClient_False(
	t *testing.T,
) {
	repo, mock, cleanup :=
		newMatchScoreSQLMock(t)

	defer cleanup()

	mock.
		ExpectQuery(`SELECT EXISTS`).
		WithArgs(
			uint(10),
			uint(5),
		).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{"exists"},
			).AddRow(
				false,
			),
		)

	exists, err := repo.JobBelongsToClient(
		context.Background(),
		10,
		5,
	)
	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}

	if exists {
		t.Fatal(
			"did not expect job to belong to client",
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLRepository_JobBelongsToClient_Error(
	t *testing.T,
) {
	repo, mock, cleanup :=
		newMatchScoreSQLMock(t)

	defer cleanup()

	expectedErr := errors.New(
		"database failed",
	)

	mock.
		ExpectQuery(`SELECT EXISTS`).
		WithArgs(
			uint(10),
			uint(5),
		).
		WillReturnError(
			expectedErr,
		)

	exists, err := repo.JobBelongsToClient(
		context.Background(),
		10,
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

	if exists {
		t.Fatal(
			"expected false result",
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLRepository_ListCandidates_Success(
	t *testing.T,
) {
	repo, mock, cleanup :=
		newMatchScoreSQLMock(t)

	defer cleanup()

	rows := sqlmock.NewRows(
		[]string{
			"application_id",
			"job_id",
			"client_id",
			"cleaner_id",
			"cleaner_name",
			"job_type",
			"job_location",
			"cleaner_location",
			"cleaner_postcode_area",
			"services_offered",
			"availability_status",
			"is_verified",
			"reliability_score",
			"response_rate",
			"average_response_minutes",
			"average_rating",
			"total_reviews",
			"badge",
		},
	).AddRow(
		1,
		10,
		5,
		20,
		"Maria Cleaner",
		"end_of_tenancy",
		"East London",
		"East London",
		"E14",
		"domestic, end of tenancy",
		"available",
		true,
		94,
		95,
		12,
		4.9,
		25,
		"Elite Cleaner",
	)

	mock.
		ExpectQuery(`SELECT\s+a\.id`).
		WithArgs(
			uint(10),
			uint(5),
		).
		WillReturnRows(
			rows,
		)

	candidates, err := repo.ListCandidates(
		context.Background(),
		10,
		5,
	)
	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}

	if len(candidates) != 1 {
		t.Fatalf(
			"expected 1 candidate, got %d",
			len(candidates),
		)
	}

	candidate := candidates[0]

	if candidate.ApplicationID != 1 {
		t.Fatalf(
			"expected application ID 1, got %d",
			candidate.ApplicationID,
		)
	}

	if candidate.CleanerID != 20 {
		t.Fatalf(
			"expected cleaner ID 20, got %d",
			candidate.CleanerID,
		)
	}

	if candidate.CleanerName !=
		"Maria Cleaner" {
		t.Fatalf(
			"unexpected cleaner name %q",
			candidate.CleanerName,
		)
	}

	if candidate.ReliabilityScore != 94 {
		t.Fatalf(
			"expected reliability 94, got %d",
			candidate.ReliabilityScore,
		)
	}

	if candidate.ResponseRate != 95 {
		t.Fatalf(
			"expected response rate 95, got %d",
			candidate.ResponseRate,
		)
	}

	if candidate.AverageResponseMinutes != 12 {
		t.Fatalf(
			"expected average response minutes 12, got %d",
			candidate.AverageResponseMinutes,
		)
	}

	if !candidate.IsVerified {
		t.Fatal(
			"expected verified cleaner",
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLRepository_ListCandidates_Empty(
	t *testing.T,
) {
	repo, mock, cleanup :=
		newMatchScoreSQLMock(t)

	defer cleanup()

	rows := sqlmock.NewRows(
		[]string{
			"application_id",
			"job_id",
			"client_id",
			"cleaner_id",
			"cleaner_name",
			"job_type",
			"job_location",
			"cleaner_location",
			"cleaner_postcode_area",
			"services_offered",
			"availability_status",
			"is_verified",
			"reliability_score",
			"response_rate",
			"average_response_minutes",
			"average_rating",
			"total_reviews",
			"badge",
		},
	)

	mock.
		ExpectQuery(`SELECT\s+a\.id`).
		WithArgs(
			uint(10),
			uint(5),
		).
		WillReturnRows(
			rows,
		)

	candidates, err := repo.ListCandidates(
		context.Background(),
		10,
		5,
	)
	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}

	if len(candidates) != 0 {
		t.Fatalf(
			"expected zero candidates, got %d",
			len(candidates),
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLRepository_ListCandidates_QueryError(
	t *testing.T,
) {
	repo, mock, cleanup :=
		newMatchScoreSQLMock(t)

	defer cleanup()

	expectedErr := errors.New(
		"candidate query failed",
	)

	mock.
		ExpectQuery(`SELECT\s+a\.id`).
		WithArgs(
			uint(10),
			uint(5),
		).
		WillReturnError(
			expectedErr,
		)

	candidates, err := repo.ListCandidates(
		context.Background(),
		10,
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

	if candidates != nil {
		t.Fatalf(
			"expected nil candidates, got %+v",
			candidates,
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLRepository_ListCandidates_ScanError(
	t *testing.T,
) {
	repo, mock, cleanup :=
		newMatchScoreSQLMock(t)

	defer cleanup()

	rows := sqlmock.NewRows(
		[]string{
			"application_id",
			"job_id",
			"client_id",
			"cleaner_id",
			"cleaner_name",
			"job_type",
			"job_location",
			"cleaner_location",
			"cleaner_postcode_area",
			"services_offered",
			"availability_status",
			"is_verified",
			"reliability_score",
			"response_rate",
			"average_response_minutes",
			"average_rating",
			"total_reviews",
			"badge",
		},
	).AddRow(
		"invalid-application-id",
		10,
		5,
		20,
		"Maria Cleaner",
		"domestic",
		"London",
		"London",
		"E14",
		"domestic",
		"available",
		true,
		90,
		95,
		10,
		4.9,
		20,
		"Top Rated",
	)

	mock.
		ExpectQuery(`SELECT\s+a\.id`).
		WithArgs(
			uint(10),
			uint(5),
		).
		WillReturnRows(
			rows,
		)

	candidates, err := repo.ListCandidates(
		context.Background(),
		10,
		5,
	)

	if err == nil {
		t.Fatal(
			"expected scan error",
		)
	}

	if candidates != nil {
		t.Fatalf(
			"expected nil candidates, got %+v",
			candidates,
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLRepository_ListCandidates_RowsError(
	t *testing.T,
) {
	repo, mock, cleanup :=
		newMatchScoreSQLMock(t)

	defer cleanup()

	expectedErr := errors.New(
		"rows failed",
	)

	rows := sqlmock.NewRows(
		[]string{
			"application_id",
			"job_id",
			"client_id",
			"cleaner_id",
			"cleaner_name",
			"job_type",
			"job_location",
			"cleaner_location",
			"cleaner_postcode_area",
			"services_offered",
			"availability_status",
			"is_verified",
			"reliability_score",
			"response_rate",
			"average_response_minutes",
			"average_rating",
			"total_reviews",
			"badge",
		},
	).
		AddRow(
			1,
			10,
			5,
			20,
			"Maria Cleaner",
			"domestic",
			"London",
			"London",
			"E14",
			"domestic",
			"available",
			true,
			90,
			95,
			10,
			4.9,
			20,
			"Top Rated",
		).
		RowError(
			0,
			expectedErr,
		)

	mock.
		ExpectQuery(`SELECT\s+a\.id`).
		WithArgs(
			uint(10),
			uint(5),
		).
		WillReturnRows(
			rows,
		)

	candidates, err := repo.ListCandidates(
		context.Background(),
		10,
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

	if candidates != nil {
		t.Fatalf(
			"expected nil candidates, got %+v",
			candidates,
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
