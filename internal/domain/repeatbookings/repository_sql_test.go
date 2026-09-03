package repeatbookings

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
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	if repo == nil {
		t.Fatal("expected repository")
	}

	if repo.db != db {
		t.Fatal("expected database to be assigned")
	}
}

func TestSQLRepository_CreateJob_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	mock.ExpectQuery(
		regexp.QuoteMeta(`
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
			VALUES ($1, $2, $3, $4, $5, $6, $7, 'open', $8, $9)
			RETURNING id
		`),
	).
		WithArgs(
			uint(5),
			"Weekly domestic clean",
			"Clean a two-bedroom flat",
			"East London",
			"domestic",
			"shift",
			70,
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
		).
		WillReturnRows(
			sqlmock.NewRows([]string{"id"}).
				AddRow(12),
		)

	jobID, err := repo.CreateJob(
		context.Background(),
		5,
		CreateRepeatBookingRequest{
			Title:       "Weekly domestic clean",
			Description: "Clean a two-bedroom flat",
			Location:    "East London",
			JobType:     "domestic",
			ListingType: "shift",
			Budget:      70,
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if jobID != 12 {
		t.Fatalf("expected job ID 12, got %d", jobID)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestSQLRepository_CreateJob_DatabaseError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)
	expectedErr := errors.New("create job failed")

	mock.ExpectQuery(
		regexp.QuoteMeta("INSERT INTO jobs"),
	).
		WillReturnError(expectedErr)

	jobID, err := repo.CreateJob(
		context.Background(),
		5,
		CreateRepeatBookingRequest{},
	)

	if jobID != 0 {
		t.Fatalf("expected job ID 0, got %d", jobID)
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
}

func TestSQLRepository_CreateInvitation_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	mock.ExpectExec(
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
			VALUES ($1, $2, $3, 'sent', $4, $5, $6)
		`),
	).
		WithArgs(
			uint(12),
			uint(5),
			uint(8),
			"Please work with me again",
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.CreateInvitation(
		context.Background(),
		12,
		5,
		8,
		"Please work with me again",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestSQLRepository_CreateInvitation_DatabaseError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)
	expectedErr := errors.New("invitation failed")

	mock.ExpectExec(
		regexp.QuoteMeta("INSERT INTO job_invitations"),
	).
		WillReturnError(expectedErr)

	err = repo.CreateInvitation(
		context.Background(),
		12,
		5,
		8,
		"Again please",
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
}

func TestSQLRepository_CreateRepeatBooking_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)
	now := time.Now()

	mock.ExpectQuery(
		regexp.QuoteMeta(`
			INSERT INTO repeat_bookings (
				client_id,
				cleaner_id,
				job_id,
				message,
				status,
				created_at,
				updated_at
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			RETURNING id, created_at, updated_at
		`),
	).
		WithArgs(
			uint(5),
			uint(8),
			uint(12),
			"Please clean again",
			"sent",
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
		).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"id",
				"created_at",
				"updated_at",
			}).AddRow(
				20,
				now,
				now,
			),
		)

	booking := &RepeatBooking{
		ClientID:  5,
		CleanerID: 8,
		JobID:     12,
		Message:   "Please clean again",
		Status:    "sent",
	}

	err = repo.CreateRepeatBooking(
		context.Background(),
		booking,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if booking.ID != 20 {
		t.Fatalf("expected ID 20, got %d", booking.ID)
	}

	if booking.CreatedAt.IsZero() {
		t.Fatal("expected created_at")
	}

	if booking.UpdatedAt.IsZero() {
		t.Fatal("expected updated_at")
	}
}

func TestSQLRepository_CreateRepeatBooking_DatabaseError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)
	expectedErr := errors.New("repeat booking failed")

	mock.ExpectQuery(
		regexp.QuoteMeta("INSERT INTO repeat_bookings"),
	).
		WillReturnError(expectedErr)

	err = repo.CreateRepeatBooking(
		context.Background(),
		&RepeatBooking{},
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
}

func TestSQLRepository_ListByClientID_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)
	now := time.Now()

	rows := sqlmock.NewRows([]string{
		"id",
		"client_id",
		"cleaner_id",
		"job_id",
		"message",
		"status",
		"created_at",
		"updated_at",
	}).
		AddRow(
			1,
			5,
			8,
			12,
			"First repeat booking",
			"sent",
			now,
			now,
		).
		AddRow(
			2,
			5,
			9,
			13,
			"Second repeat booking",
			"sent",
			now,
			now,
		)

	mock.ExpectQuery(
		regexp.QuoteMeta(`
			SELECT
				id,
				client_id,
				cleaner_id,
				job_id,
				message,
				status,
				created_at,
				updated_at
			FROM repeat_bookings
			WHERE client_id = $1
			ORDER BY created_at DESC
		`),
	).
		WithArgs(uint(5)).
		WillReturnRows(rows)

	bookings, err := repo.ListByClientID(
		context.Background(),
		5,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(bookings) != 2 {
		t.Fatalf("expected 2 bookings, got %d", len(bookings))
	}

	if bookings[0].ID != 1 {
		t.Fatalf("expected first ID 1, got %d", bookings[0].ID)
	}
}

func TestSQLRepository_ListByClientID_Empty(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	mock.ExpectQuery(
		regexp.QuoteMeta("FROM repeat_bookings"),
	).
		WithArgs(uint(5)).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"id",
				"client_id",
				"cleaner_id",
				"job_id",
				"message",
				"status",
				"created_at",
				"updated_at",
			}),
		)

	bookings, err := repo.ListByClientID(
		context.Background(),
		5,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(bookings) != 0 {
		t.Fatalf("expected no bookings, got %d", len(bookings))
	}
}

func TestSQLRepository_ListByClientID_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)
	expectedErr := errors.New("query failed")

	mock.ExpectQuery(
		regexp.QuoteMeta("FROM repeat_bookings"),
	).
		WithArgs(uint(5)).
		WillReturnError(expectedErr)

	bookings, err := repo.ListByClientID(
		context.Background(),
		5,
	)

	if bookings != nil {
		t.Fatalf("expected nil bookings, got %+v", bookings)
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
}

func TestSQLRepository_ListByClientID_ScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)
	now := time.Now()

	rows := sqlmock.NewRows([]string{
		"id",
		"client_id",
		"cleaner_id",
		"job_id",
		"message",
		"status",
		"created_at",
		"updated_at",
	}).AddRow(
		"invalid-id",
		5,
		8,
		12,
		"Again",
		"sent",
		now,
		now,
	)

	mock.ExpectQuery(
		regexp.QuoteMeta("FROM repeat_bookings"),
	).
		WithArgs(uint(5)).
		WillReturnRows(rows)

	bookings, err := repo.ListByClientID(
		context.Background(),
		5,
	)

	if bookings != nil {
		t.Fatalf("expected nil bookings, got %+v", bookings)
	}

	if err == nil {
		t.Fatal("expected scan error")
	}
}

func TestSQLRepository_ListByClientID_RowsError(t *testing.T) {
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
		"client_id",
		"cleaner_id",
		"job_id",
		"message",
		"status",
		"created_at",
		"updated_at",
	}).
		AddRow(
			1,
			5,
			8,
			12,
			"Again",
			"sent",
			now,
			now,
		).
		RowError(0, expectedErr)

	mock.ExpectQuery(
		regexp.QuoteMeta("FROM repeat_bookings"),
	).
		WithArgs(uint(5)).
		WillReturnRows(rows)

	bookings, err := repo.ListByClientID(
		context.Background(),
		5,
	)

	if bookings != nil {
		t.Fatalf("expected nil bookings, got %+v", bookings)
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
}

func TestSQLRepository_GetOriginBooking_WithApplicationID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	mock.ExpectQuery(
		regexp.QuoteMeta(`
			SELECT
				id,
				job_id,
				application_id,
				client_id,
				cleaner_id,
				status
			FROM bookings
			WHERE id = $1
			LIMIT 1
		`),
	).
		WithArgs(uint(10)).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"id",
				"job_id",
				"application_id",
				"client_id",
				"cleaner_id",
				"status",
			}).AddRow(
				10,
				7,
				33,
				5,
				8,
				"closed",
			),
		)

	booking, err := repo.GetOriginBooking(
		context.Background(),
		10,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if booking == nil {
		t.Fatal("expected booking")
	}

	if booking.ApplicationID == nil {
		t.Fatal("expected application ID")
	}

	if *booking.ApplicationID != 33 {
		t.Fatalf(
			"expected application ID 33, got %d",
			*booking.ApplicationID,
		)
	}
}

func TestSQLRepository_GetOriginBooking_WithoutApplicationID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	mock.ExpectQuery(
		regexp.QuoteMeta("FROM bookings"),
	).
		WithArgs(uint(10)).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"id",
				"job_id",
				"application_id",
				"client_id",
				"cleaner_id",
				"status",
			}).AddRow(
				10,
				7,
				nil,
				5,
				8,
				"completed",
			),
		)

	booking, err := repo.GetOriginBooking(
		context.Background(),
		10,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if booking.ApplicationID != nil {
		t.Fatalf(
			"expected nil application ID, got %v",
			*booking.ApplicationID,
		)
	}
}

func TestSQLRepository_GetOriginBooking_DatabaseError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)
	expectedErr := errors.New("booking lookup failed")

	mock.ExpectQuery(
		regexp.QuoteMeta("FROM bookings"),
	).
		WithArgs(uint(10)).
		WillReturnError(expectedErr)

	booking, err := repo.GetOriginBooking(
		context.Background(),
		10,
	)

	if booking != nil {
		t.Fatalf("expected nil booking, got %+v", booking)
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
}

func TestSQLRepository_CreateRepeatBookings_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)
	applicationID := uint(33)
	scheduledAt := time.Date(
		2026,
		time.August,
		10,
		10,
		0,
		0,
		0,
		time.UTC,
	)

	scheduledEndAt := scheduledAt.Add(
		2 * time.Hour,
	)
	mock.ExpectQuery(
		regexp.QuoteMeta("INSERT INTO bookings"),
	).
		WithArgs(
			uint(7),
			applicationID,
			uint(5),
			uint(8),
			scheduledAt,
			scheduledEndAt,
			sqlmock.AnyArg(),
		).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{"id"},
			).AddRow(40),
		)

	newBookingID, err := repo.CreateRepeatBookings(
		context.Background(),
		&BookingSnapshot{
			JobID:         7,
			ApplicationID: &applicationID,
			ClientID:      5,
			CleanerID:     8,
		},
		scheduledAt,
		scheduledEndAt,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if newBookingID != 40 {
		t.Fatalf(
			"expected booking ID 40, got %d",
			newBookingID,
		)
	}
}

func TestSQLRepository_CreateRepeatBookings_DatabaseError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)
	expectedErr := errors.New("create booking failed")

	mock.ExpectQuery(
		regexp.QuoteMeta("INSERT INTO bookings"),
	).
		WillReturnError(expectedErr)

	scheduledAt := time.Now().Add(
		24 * time.Hour,
	)

	scheduledEndAt := scheduledAt.Add(
		2 * time.Hour,
	)

	newBookingID, err := repo.CreateRepeatBookings(
		context.Background(),
		&BookingSnapshot{},
		scheduledAt,
		scheduledEndAt,
	)

	if newBookingID != 0 {
		t.Fatalf(
			"expected booking ID 0, got %d",
			newBookingID,
		)
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
}

func TestSQLRepository_CreateBookAgainRequest_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)
	newBookingID := uint(40)
	now := time.Now()
	scheduledAt := time.Date(
		2026,
		time.August,
		10,
		10,
		0,
		0,
		0,
		time.UTC,
	)

	scheduledEndAt := scheduledAt.Add(
		2 * time.Hour,
	)

	mock.ExpectQuery(
		regexp.QuoteMeta(`
		INSERT INTO repeat_booking_requests (
			original_booking_id,
			new_booking_id,
			client_id,
			cleaner_id,
			job_id,
			scheduled_at,
			scheduled_end_at,
			status,
			message,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $10)
		RETURNING id, created_at, updated_at
	`),
	).
		WithArgs(
			uint(10),
			&newBookingID,
			uint(5),
			uint(8),
			uint(7),
			scheduledAt,
			&scheduledEndAt,
			"created",
			"Please return",
			sqlmock.AnyArg(),
		).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"id",
				"created_at",
				"updated_at",
			}).AddRow(
				50,
				now,
				now,
			),
		)

	request := &RepeatBookingRequest{
		OriginalBookingID: 10,
		NewBookingID:      &newBookingID,
		ClientID:          5,
		CleanerID:         8,
		JobID:             7,
		ScheduledAt:       scheduledAt,
		ScheduledEndAt:    &scheduledEndAt,
		Status:            "created",
		Message:           "Please return",
	}

	err = repo.CreateBookAgainRequest(
		context.Background(),
		request,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if request.ID != 50 {
		t.Fatalf("expected request ID 50, got %d", request.ID)
	}
}

func TestSQLRepository_CreateBookAgainRequest_DatabaseError(
	t *testing.T,
) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)
	expectedErr := errors.New("request insert failed")

	mock.ExpectQuery(
		regexp.QuoteMeta(
			"INSERT INTO repeat_booking_requests",
		),
	).
		WillReturnError(expectedErr)

	err = repo.CreateBookAgainRequest(
		context.Background(),
		&RepeatBookingRequest{},
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
}

func TestSQLRepository_CreateBookAgainTransaction_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	applicationID := uint(33)
	scheduledAt := time.Date(
		2026,
		time.August,
		10,
		10,
		0,
		0,
		0,
		time.UTC,
	)

	now := time.Now()

	mock.ExpectBegin()

	mock.ExpectQuery(
		regexp.QuoteMeta("INSERT INTO bookings"),
	).
		WithArgs(
			uint(7),
			&applicationID,
			uint(5),
			uint(8),
			scheduledAt,
			sqlmock.AnyArg(),
		).
		WillReturnRows(
			sqlmock.NewRows([]string{"id"}).
				AddRow(40),
		)

	mock.ExpectQuery(
		regexp.QuoteMeta(
			"INSERT INTO repeat_booking_requests",
		),
	).
		WithArgs(
			uint(10),
			sqlmock.AnyArg(),
			uint(5),
			uint(8),
			uint(7),
			scheduledAt,
			"created",
			"Please return",
			sqlmock.AnyArg(),
		).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"id",
				"created_at",
				"updated_at",
			}).AddRow(
				50,
				now,
				now,
			),
		)

	mock.ExpectCommit()

	request, err := repo.CreateBookAgainTransaction(
		context.Background(),
		BookAgainTransaction{
			OriginalBookingID: 10,
			ClientID:          5,
			CleanerID:         8,
			JobID:             7,
			ApplicationID:     &applicationID,
			ScheduledAt:       scheduledAt,
			Status:            "created",
			Message:           "Please return",
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if request == nil {
		t.Fatal("expected repeat-booking request")
	}

	if request.ID != 50 {
		t.Fatalf("expected request ID 50, got %d", request.ID)
	}

	if request.NewBookingID == nil {
		t.Fatal("expected new booking ID")
	}

	if *request.NewBookingID != 40 {
		t.Fatalf(
			"expected booking ID 40, got %d",
			*request.NewBookingID,
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestSQLRepository_CreateBookAgainTransaction_BeginError(
	t *testing.T,
) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)
	expectedErr := errors.New("begin failed")

	mock.ExpectBegin().
		WillReturnError(expectedErr)

	request, err := repo.CreateBookAgainTransaction(
		context.Background(),
		BookAgainTransaction{},
	)

	if request != nil {
		t.Fatalf("expected nil request, got %+v", request)
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
}

func TestSQLRepository_CreateBookAgainTransaction_BookingInsertError(
	t *testing.T,
) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)
	expectedErr := errors.New("booking insert failed")

	mock.ExpectBegin()

	mock.ExpectQuery(
		regexp.QuoteMeta("INSERT INTO bookings"),
	).
		WillReturnError(expectedErr)

	mock.ExpectRollback()

	request, err := repo.CreateBookAgainTransaction(
		context.Background(),
		BookAgainTransaction{},
	)

	if request != nil {
		t.Fatalf("expected nil request, got %+v", request)
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestSQLRepository_CreateBookAgainTransaction_RequestInsertError(
	t *testing.T,
) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	expectedErr := errors.New("request insert failed")
	scheduledAt := time.Date(
		2026,
		time.August,
		10,
		10,
		0,
		0,
		0,
		time.UTC,
	)

	mock.ExpectBegin()

	mock.ExpectQuery(
		regexp.QuoteMeta("INSERT INTO bookings"),
	).
		WillReturnRows(
			sqlmock.NewRows([]string{"id"}).
				AddRow(40),
		)

	mock.ExpectQuery(
		regexp.QuoteMeta(
			"INSERT INTO repeat_booking_requests",
		),
	).
		WillReturnError(expectedErr)

	mock.ExpectRollback()

	request, err := repo.CreateBookAgainTransaction(
		context.Background(),
		BookAgainTransaction{
			OriginalBookingID: 10,
			ClientID:          5,
			CleanerID:         8,
			JobID:             7,
			ScheduledAt:       scheduledAt,
			Status:            "created",
			Message:           "Please return",
		},
	)

	if request != nil {
		t.Fatalf("expected nil request, got %+v", request)
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestSQLRepository_CreateBookAgainTransaction_CommitError(
	t *testing.T,
) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	expectedErr := errors.New("commit failed")
	scheduledAt := time.Date(
		2026,
		time.August,
		10,
		10,
		0,
		0,
		0,
		time.UTC,
	)
	now := time.Now()

	mock.ExpectBegin()

	mock.ExpectQuery(
		regexp.QuoteMeta("INSERT INTO bookings"),
	).
		WillReturnRows(
			sqlmock.NewRows([]string{"id"}).
				AddRow(40),
		)

	mock.ExpectQuery(
		regexp.QuoteMeta(
			"INSERT INTO repeat_booking_requests",
		),
	).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"id",
				"created_at",
				"updated_at",
			}).AddRow(
				50,
				now,
				now,
			),
		)

	mock.ExpectCommit().
		WillReturnError(expectedErr)

	request, err := repo.CreateBookAgainTransaction(
		context.Background(),
		BookAgainTransaction{
			OriginalBookingID: 10,
			ClientID:          5,
			CleanerID:         8,
			JobID:             7,
			ScheduledAt:       scheduledAt,
			Status:            "created",
			Message:           "Please return",
		},
	)

	if request != nil {
		t.Fatalf("expected nil request, got %+v", request)
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
