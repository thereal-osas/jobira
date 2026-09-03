package profiles

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func newProfilesRepositoryTest(
	t *testing.T,
) (*SQLRepository, sqlmock.Sqlmock, func()) {
	t.Helper()

	db, mock, err := sqlmock.New(
		sqlmock.QueryMatcherOption(
			sqlmock.QueryMatcherRegexp,
		),
	)
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}

	cleanup := func() {
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unmet SQL expectations: %v", err)
		}
	}

	return NewSQLRepository(db), mock, cleanup
}

func profileColumns() []string {
	return []string{
		"id",
		"user_id",
		"bio",
		"location",
		"country",
		"city",
		"region",
		"postcode_area",
		"availability_status",
		"travel_radius_miles",
		"jobs_completed",
		"jobs_cancelled",
		"response_rate",
		"reliability_score",
		"badge",
		"years_experience",
		"hourly_rate",
		"services_offered",
		"is_verified",
		"verification_status",
		"created_at",
		"updated_at",
	}
}

func TestNewSQLRepository(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}

	repository := NewSQLRepository(db)

	if repository == nil {
		t.Fatal("expected repository")
	}

	if repository.db != db {
		t.Fatal("expected supplied database")
	}
}

func TestSQLRepository_Create_Success(t *testing.T) {
	repository, mock, cleanup := newProfilesRepositoryTest(t)
	defer cleanup()

	profile := &CleanerProfile{
		UserID:             10,
		Bio:                "Cleaner",
		Country:            "UK",
		City:               "London",
		Region:             "East London",
		PostcodeArea:       "E14",
		AvailabilityStatus: "available_immediately",
		TravelRadiusMiles:  10,
		Location:           "Canary Wharf",
		YearsExperience:    5,
		HourlyRate:         25,
		ServicesOffered:    "Domestic",
	}

	mock.ExpectQuery(
		`(?s)INSERT INTO cleaner_profiles.*RETURNING id`,
	).
		WithArgs(
			profile.UserID,
			profile.Bio,
			profile.Country,
			profile.City,
			profile.Region,
			profile.PostcodeArea,
			profile.AvailabilityStatus,
			profile.TravelRadiusMiles,
			profile.Location,
			profile.YearsExperience,
			profile.HourlyRate,
			profile.ServicesOffered,
			false,
			"pending",
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
		profile,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if profile.ID != 5 {
		t.Fatalf(
			"expected ID 5, got %d",
			profile.ID,
		)
	}

	if profile.IsVerified {
		t.Fatal("expected unverified profile")
	}

	if profile.VerificationStatus != "pending" {
		t.Fatalf(
			"expected pending status, got %q",
			profile.VerificationStatus,
		)
	}

	if profile.CreatedAt.IsZero() {
		t.Fatal("expected CreatedAt")
	}

	if profile.UpdatedAt.IsZero() {
		t.Fatal("expected UpdatedAt")
	}
}

func TestSQLRepository_Create_Duplicate(t *testing.T) {
	repository, mock, cleanup := newProfilesRepositoryTest(t)
	defer cleanup()

	profile := &CleanerProfile{
		UserID:             10,
		Bio:                "Cleaner",
		AvailabilityStatus: "available_immediately",
		Location:           "London",
		ServicesOffered:    "Domestic",
	}

	mock.ExpectQuery(
		`(?s)INSERT INTO cleaner_profiles.*RETURNING id`,
	).
		WithArgs(
			profile.UserID,
			profile.Bio,
			profile.Country,
			profile.City,
			profile.Region,
			profile.PostcodeArea,
			profile.AvailabilityStatus,
			profile.TravelRadiusMiles,
			profile.Location,
			profile.YearsExperience,
			profile.HourlyRate,
			profile.ServicesOffered,
			false,
			"pending",
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
		).
		WillReturnError(
			errors.New("duplicate key value violates unique constraint"),
		)

	err := repository.Create(
		context.Background(),
		profile,
	)

	if !errors.Is(err, ErrProfileAlreadyExists) {
		t.Fatalf(
			"expected ErrProfileAlreadyExists, got %v",
			err,
		)
	}
}

func TestSQLRepository_Create_DatabaseError(t *testing.T) {
	repository, mock, cleanup := newProfilesRepositoryTest(t)
	defer cleanup()

	databaseErr := errors.New("database error")

	profile := &CleanerProfile{
		UserID: 1,
	}

	mock.ExpectQuery(
		`(?s)INSERT INTO cleaner_profiles.*RETURNING id`,
	).
		WithArgs(
			profile.UserID,
			profile.Bio,
			profile.Country,
			profile.City,
			profile.Region,
			profile.PostcodeArea,
			profile.AvailabilityStatus,
			profile.TravelRadiusMiles,
			profile.Location,
			profile.YearsExperience,
			profile.HourlyRate,
			profile.ServicesOffered,
			false,
			"pending",
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
		).
		WillReturnError(databaseErr)

	err := repository.Create(
		context.Background(),
		profile,
	)

	if !errors.Is(err, databaseErr) {
		t.Fatalf(
			"expected database error, got %v",
			err,
		)
	}
}

func TestSQLRepository_GetByUserID_Success(t *testing.T) {
	repository, mock, cleanup := newProfilesRepositoryTest(t)
	defer cleanup()

	now := time.Now()

	mock.ExpectQuery(
		`(?s)SELECT.*FROM cleaner_profiles.*WHERE user_id = \$1`,
	).
		WithArgs(uint(10)).
		WillReturnRows(
			sqlmock.NewRows(profileColumns()).
				AddRow(
					5,
					10,
					"Cleaner",
					"Canary Wharf",
					"UK",
					"London",
					"East London",
					"E14",
					"available_immediately",
					15,
					12,
					1,
					90,
					85,
					"Trusted Cleaner",
					5,
					25,
					"Domestic",
					true,
					"verified",
					now,
					now,
				),
		)

	profile, err := repository.GetByUserID(
		context.Background(),
		10,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if profile.ID != 5 {
		t.Fatalf(
			"expected profile ID 5, got %d",
			profile.ID,
		)
	}

	if profile.UserID != 10 {
		t.Fatalf(
			"expected user ID 10, got %d",
			profile.UserID,
		)
	}
}

func TestSQLRepository_GetByUserID_NotFound(t *testing.T) {
	repository, mock, cleanup := newProfilesRepositoryTest(t)
	defer cleanup()

	mock.ExpectQuery(
		`(?s)SELECT.*FROM cleaner_profiles.*WHERE user_id = \$1`,
	).
		WithArgs(uint(999)).
		WillReturnError(sql.ErrNoRows)

	profile, err := repository.GetByUserID(
		context.Background(),
		999,
	)

	if profile != nil {
		t.Fatal("expected nil profile")
	}

	if !errors.Is(err, ErrProfileNotFound) {
		t.Fatalf(
			"expected ErrProfileNotFound, got %v",
			err,
		)
	}
}

func TestSQLRepository_GetByUserID_DatabaseError(t *testing.T) {
	repository, mock, cleanup := newProfilesRepositoryTest(t)
	defer cleanup()

	databaseErr := errors.New("database error")

	mock.ExpectQuery(
		`(?s)SELECT.*FROM cleaner_profiles.*WHERE user_id = \$1`,
	).
		WithArgs(uint(10)).
		WillReturnError(databaseErr)

	_, err := repository.GetByUserID(
		context.Background(),
		10,
	)

	if !errors.Is(err, databaseErr) {
		t.Fatalf(
			"expected database error, got %v",
			err,
		)
	}
}

func TestSQLRepository_Update_Success(t *testing.T) {
	repository, mock, cleanup := newProfilesRepositoryTest(t)
	defer cleanup()

	profile := &CleanerProfile{
		UserID:             10,
		Bio:                "Updated",
		Location:           "London",
		Country:            "UK",
		City:               "London",
		Region:             "East",
		PostcodeArea:       "E14",
		AvailabilityStatus: "weekends_only",
		TravelRadiusMiles:  20,
		YearsExperience:    7,
		HourlyRate:         30,
		ServicesOffered:    "Domestic",
		ReliabilityScore:   85,
		Badge:              "Trusted Cleaner",
	}

	mock.ExpectExec(
		`(?s)UPDATE cleaner_profiles.*SET.*bio = \$1.*WHERE user_id = \$15`,
	).
		WithArgs(
			profile.Bio,
			profile.Location,
			profile.Country,
			profile.City,
			profile.Region,
			profile.PostcodeArea,
			profile.AvailabilityStatus,
			profile.TravelRadiusMiles,
			profile.YearsExperience,
			profile.HourlyRate,
			profile.ServicesOffered,
			profile.ReliabilityScore,
			profile.Badge,
			sqlmock.AnyArg(),
			profile.UserID,
		).
		WillReturnResult(
			sqlmock.NewResult(0, 1),
		)

	err := repository.Update(
		context.Background(),
		profile,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSQLRepository_Update_NotFound(t *testing.T) {
	repository, mock, cleanup := newProfilesRepositoryTest(t)
	defer cleanup()

	profile := &CleanerProfile{
		UserID: 999,
	}

	mock.ExpectExec(
		`(?s)UPDATE cleaner_profiles.*WHERE user_id = \$15`,
	).
		WithArgs(
			profile.Bio,
			profile.Location,
			profile.Country,
			profile.City,
			profile.Region,
			profile.PostcodeArea,
			profile.AvailabilityStatus,
			profile.TravelRadiusMiles,
			profile.YearsExperience,
			profile.HourlyRate,
			profile.ServicesOffered,
			profile.ReliabilityScore,
			profile.Badge,
			sqlmock.AnyArg(),
			profile.UserID,
		).
		WillReturnResult(
			sqlmock.NewResult(0, 0),
		)

	err := repository.Update(
		context.Background(),
		profile,
	)

	if !errors.Is(err, ErrProfileNotFound) {
		t.Fatalf(
			"expected ErrProfileNotFound, got %v",
			err,
		)
	}
}

func TestSQLRepository_Update_DatabaseError(t *testing.T) {
	repository, mock, cleanup := newProfilesRepositoryTest(t)
	defer cleanup()

	databaseErr := errors.New("update failed")

	profile := &CleanerProfile{
		UserID: 1,
	}

	mock.ExpectExec(
		`(?s)UPDATE cleaner_profiles.*WHERE user_id = \$15`,
	).
		WithArgs(
			profile.Bio,
			profile.Location,
			profile.Country,
			profile.City,
			profile.Region,
			profile.PostcodeArea,
			profile.AvailabilityStatus,
			profile.TravelRadiusMiles,
			profile.YearsExperience,
			profile.HourlyRate,
			profile.ServicesOffered,
			profile.ReliabilityScore,
			profile.Badge,
			sqlmock.AnyArg(),
			profile.UserID,
		).
		WillReturnError(databaseErr)

	err := repository.Update(
		context.Background(),
		profile,
	)

	if !errors.Is(err, databaseErr) {
		t.Fatalf(
			"expected database error, got %v",
			err,
		)
	}
}

func TestSQLRepository_UpdateVerificationStatus_Success(
	t *testing.T,
) {
	repository, mock, cleanup := newProfilesRepositoryTest(t)
	defer cleanup()

	mock.ExpectExec(
		`(?s)UPDATE cleaner_profiles.*verification_status = \$1.*WHERE user_id = \$4`,
	).
		WithArgs(
			"verified",
			true,
			sqlmock.AnyArg(),
			uint(10),
		).
		WillReturnResult(
			sqlmock.NewResult(0, 1),
		)

	err := repository.UpdateVerificationStatus(
		context.Background(),
		10,
		"verified",
		true,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSQLRepository_UpdateVerificationStatus_NotFound(
	t *testing.T,
) {
	repository, mock, cleanup := newProfilesRepositoryTest(t)
	defer cleanup()

	mock.ExpectExec(
		`(?s)UPDATE cleaner_profiles.*verification_status = \$1.*WHERE user_id = \$4`,
	).
		WithArgs(
			"verified",
			true,
			sqlmock.AnyArg(),
			uint(999),
		).
		WillReturnResult(
			sqlmock.NewResult(0, 0),
		)

	err := repository.UpdateVerificationStatus(
		context.Background(),
		999,
		"verified",
		true,
	)

	if !errors.Is(err, ErrProfileNotFound) {
		t.Fatalf(
			"expected ErrProfileNotFound, got %v",
			err,
		)
	}
}

func TestSQLRepository_Search_Success(t *testing.T) {
	repository, mock, cleanup := newProfilesRepositoryTest(t)
	defer cleanup()

	now := time.Now()
	verified := true

	req := SearchProfilesRequest{
		ClientID:           10,
		Country:            "UK",
		City:               "London",
		Region:             "East",
		PostcodeArea:       "E14",
		AvailabilityStatus: "weekends_only",
		ServicesOffered:    "airbnb",
		MinExperience:      3,
		MaxHourlyRate:      30,
		MaxTravelRadius:    20,
		IsVerified:         &verified,
	}

	mock.ExpectQuery(
		`(?s)SELECT.*FROM cleaner_profiles.*blocked_cleaners.*travel_radius_miles.*ORDER BY`,
	).
		WithArgs(
			req.ClientID,
			req.Country,
			req.City,
			req.Region,
			req.PostcodeArea,
			req.AvailabilityStatus,
			"%"+req.ServicesOffered+"%",
			req.MinExperience,
			req.MaxHourlyRate,
			req.MaxTravelRadius,
			true,
		).
		WillReturnRows(
			sqlmock.NewRows(profileColumns()).
				AddRow(
					5,
					11,
					"Cleaner",
					"London",
					"UK",
					"London",
					"East",
					"E14",
					"weekends_only",
					10,
					20,
					0,
					95,
					100,
					"Elite Cleaner",
					6,
					25,
					"Airbnb",
					true,
					"verified",
					now,
					now,
				),
		)

	profiles, err := repository.Search(
		context.Background(),
		req,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(profiles) != 1 {
		t.Fatalf(
			"expected one profile, got %d",
			len(profiles),
		)
	}

	if profiles[0].ID != 5 {
		t.Fatalf(
			"expected ID 5, got %d",
			profiles[0].ID,
		)
	}
}

func TestSQLRepository_Search_QueryError(t *testing.T) {
	repository, mock, cleanup := newProfilesRepositoryTest(t)
	defer cleanup()

	queryErr := errors.New("search query failed")

	mock.ExpectQuery(
		`(?s)SELECT.*FROM cleaner_profiles.*ORDER BY`,
	).
		WillReturnError(queryErr)

	_, err := repository.Search(
		context.Background(),
		SearchProfilesRequest{},
	)

	if !errors.Is(err, queryErr) {
		t.Fatalf(
			"expected query error, got %v",
			err,
		)
	}
}

func TestSQLRepository_IncrementJobsCompleted_Success(
	t *testing.T,
) {
	repository, mock, cleanup := newProfilesRepositoryTest(t)
	defer cleanup()

	mock.ExpectExec(
		`(?s)UPDATE cleaner_profiles.*jobs_completed = jobs_completed \+ 1.*WHERE user_id = \$2`,
	).
		WithArgs(
			sqlmock.AnyArg(),
			uint(10),
		).
		WillReturnResult(
			sqlmock.NewResult(0, 1),
		)

	err := repository.IncrementJobsCompleted(
		context.Background(),
		10,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSQLRepository_IncrementJobsCompleted_Error(
	t *testing.T,
) {
	repository, mock, cleanup := newProfilesRepositoryTest(t)
	defer cleanup()

	databaseErr := errors.New("increment failed")

	mock.ExpectExec(
		`(?s)UPDATE cleaner_profiles.*jobs_completed = jobs_completed \+ 1.*WHERE user_id = \$2`,
	).
		WithArgs(
			sqlmock.AnyArg(),
			uint(10),
		).
		WillReturnError(databaseErr)

	err := repository.IncrementJobsCompleted(
		context.Background(),
		10,
	)

	if !errors.Is(err, databaseErr) {
		t.Fatalf(
			"expected database error, got %v",
			err,
		)
	}
}

func TestSQLRepository_IncrementJobsCancelled_Success(
	t *testing.T,
) {
	repository, mock, cleanup := newProfilesRepositoryTest(t)
	defer cleanup()

	mock.ExpectExec(
		`(?s)UPDATE cleaner_profiles.*jobs_cancelled = jobs_cancelled \+ 1.*WHERE user_id = \$2`,
	).
		WithArgs(
			sqlmock.AnyArg(),
			uint(10),
		).
		WillReturnResult(
			sqlmock.NewResult(0, 1),
		)

	err := repository.IncrementJobsCancelled(
		context.Background(),
		10,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSQLRepository_RecalculateReputation_Success(
	t *testing.T,
) {
	repository, mock, cleanup := newProfilesRepositoryTest(t)
	defer cleanup()

	now := time.Now()

	mock.ExpectQuery(
		`(?s)SELECT.*FROM cleaner_profiles.*WHERE user_id = \$1`,
	).
		WithArgs(uint(10)).
		WillReturnRows(
			sqlmock.NewRows(profileColumns()).
				AddRow(
					5,
					10,
					"Cleaner",
					"London",
					"UK",
					"London",
					"East",
					"E14",
					"available_immediately",
					10,
					20,
					0,
					95,
					0,
					"",
					5,
					25,
					"Domestic",
					true,
					"verified",
					now,
					now,
				),
		)

	mock.ExpectExec(
		`(?s)UPDATE cleaner_profiles.*reliability_score = \$1.*badge = \$2.*WHERE user_id = \$4`,
	).
		WithArgs(
			100,
			"Fast Responder",
			sqlmock.AnyArg(),
			uint(10),
		).
		WillReturnResult(
			sqlmock.NewResult(0, 1),
		)

	err := repository.RecalculateReputation(
		context.Background(),
		10,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSQLRepository_RecalculateReputation_ProfileError(
	t *testing.T,
) {
	repository, mock, cleanup := newProfilesRepositoryTest(t)
	defer cleanup()

	lookupErr := errors.New("lookup failed")

	mock.ExpectQuery(
		`(?s)SELECT.*FROM cleaner_profiles.*WHERE user_id = \$1`,
	).
		WithArgs(uint(10)).
		WillReturnError(lookupErr)

	err := repository.RecalculateReputation(
		context.Background(),
		10,
	)

	if !errors.Is(err, lookupErr) {
		t.Fatalf(
			"expected lookup error, got %v",
			err,
		)
	}
}

func historyColumns() []string {
	return []string{
		"id",
		"job_id",
		"title",
		"status",
		"location",
		"created_at",
	}
}

func TestSQLRepository_GetCompletedJobs_Success(t *testing.T) {
	repository, mock, cleanup := newProfilesRepositoryTest(t)
	defer cleanup()

	now := time.Now()

	mock.ExpectQuery(
		`(?s)SELECT.*FROM applications a.*a\.status = 'completed'.*ORDER BY`,
	).
		WithArgs(uint(10)).
		WillReturnRows(
			sqlmock.NewRows(historyColumns()).
				AddRow(
					1,
					20,
					"Office clean",
					"completed",
					"London",
					now,
				),
		)

	history, err := repository.GetCompletedJobs(
		context.Background(),
		10,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(history) != 1 {
		t.Fatalf(
			"expected one history item, got %d",
			len(history),
		)
	}
}

func TestSQLRepository_GetCancelledJobs_Success(t *testing.T) {
	repository, mock, cleanup := newProfilesRepositoryTest(t)
	defer cleanup()

	now := time.Now()

	mock.ExpectQuery(
		`(?s)SELECT.*FROM applications a.*a\.status = 'cancelled'.*ORDER BY`,
	).
		WithArgs(uint(10)).
		WillReturnRows(
			sqlmock.NewRows(historyColumns()).
				AddRow(
					1,
					20,
					"Office clean",
					"cancelled",
					"London",
					now,
				),
		)

	history, err := repository.GetCancelledJobs(
		context.Background(),
		10,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(history) != 1 {
		t.Fatalf(
			"expected one history item, got %d",
			len(history),
		)
	}
}

func TestSQLRepository_GetFullHistory_Success(t *testing.T) {
	repository, mock, cleanup := newProfilesRepositoryTest(t)
	defer cleanup()

	now := time.Now()

	mock.ExpectQuery(
		`(?s)SELECT.*FROM applications a.*WHERE a\.cleaner_id = \$1.*ORDER BY`,
	).
		WithArgs(uint(10)).
		WillReturnRows(
			sqlmock.NewRows(historyColumns()).
				AddRow(
					1,
					20,
					"Office clean",
					"completed",
					"London",
					now,
				).
				AddRow(
					2,
					21,
					"Airbnb clean",
					"cancelled",
					"London",
					now,
				),
		)

	history, err := repository.GetFullHistory(
		context.Background(),
		10,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(history) != 2 {
		t.Fatalf(
			"expected two history items, got %d",
			len(history),
		)
	}
}

func TestSQLRepository_GetCompletedJobs_RowsError(t *testing.T) {
	repository, mock, cleanup := newProfilesRepositoryTest(t)
	defer cleanup()

	rowsErr := errors.New("rows error")
	now := time.Now()

	rows := sqlmock.NewRows(historyColumns()).
		AddRow(
			1,
			20,
			"Office clean",
			"completed",
			"London",
			now,
		).
		RowError(0, rowsErr)

	mock.ExpectQuery(
		`(?s)SELECT.*FROM applications a.*a\.status = 'completed'.*ORDER BY`,
	).
		WithArgs(uint(10)).
		WillReturnRows(rows)

	_, err := repository.GetCompletedJobs(
		context.Background(),
		10,
	)

	if !errors.Is(err, rowsErr) {
		t.Fatalf(
			"expected rows error, got %v",
			err,
		)
	}
}

var _ Repository = (*SQLRepository)(nil)

// Keep regexp imported for easy exact-query additions in the
// coverage-polish batch.
var _ = regexp.QuoteMeta
