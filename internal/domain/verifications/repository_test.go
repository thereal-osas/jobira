package verifications

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func newVerificationRepositoryTest(
	t *testing.T,
) (*SQLRepository, sqlmock.Sqlmock) {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf(
			"failed creating sqlmock: %v",
			err,
		)
	}

	return NewSQLRepository(db), mock
}

func verificationColumns() []string {
	return []string{
		"id",
		"user_id",
		"verification_type",
		"document_url",
		"status",
		"admin_notes",
		"reviewed_by",
		"reviewed_at",
		"created_at",
		"updated_at",
	}
}

func TestNewSQLRepository(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}

	defer db.Close()

	repository := NewSQLRepository(db)

	if repository == nil {
		t.Fatal("expected repository")
	}

	if repository.db != db {
		t.Fatal(
			"expected supplied database",
		)
	}
}

func TestSQLRepository_Create_Success(t *testing.T) {
	repository, mock :=
		newVerificationRepositoryTest(t)

	request := &VerificationRequest{
		UserID:           5,
		VerificationType: "dbs_check",
		DocumentURL:      "https://test/dbs",
		Status:           "pending",
	}

	now := time.Now()

	mock.ExpectQuery(
		`(?s)INSERT INTO verification_requests.*RETURNING id, created_at, updated_at`,
	).
		WithArgs(
			request.UserID,
			request.VerificationType,
			request.DocumentURL,
			request.Status,
			sqlmock.AnyArg(),
		).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{
					"id",
					"created_at",
					"updated_at",
				},
			).AddRow(
				10,
				now,
				now,
			),
		)

	err := repository.Create(
		context.Background(),
		request,
	)

	if err != nil {
		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}

	if request.ID != 10 {
		t.Fatalf(
			"expected ID 10 got %d",
			request.ID,
		)
	}
}

func TestSQLRepository_Create_Error(t *testing.T) {
	repository, mock :=
		newVerificationRepositoryTest(t)

	createErr := errors.New("create failed")

	request := &VerificationRequest{
		UserID:           5,
		VerificationType: "id_check",
		DocumentURL:      "https://test/id",
		Status:           "pending",
	}

	mock.ExpectQuery(
		`(?s)INSERT INTO verification_requests.*RETURNING`,
	).
		WithArgs(
			request.UserID,
			request.VerificationType,
			request.DocumentURL,
			request.Status,
			sqlmock.AnyArg(),
		).
		WillReturnError(createErr)

	err := repository.Create(
		context.Background(),
		request,
	)

	if !errors.Is(err, createErr) {
		t.Fatalf(
			"expected create error got %v",
			err,
		)
	}
}

func TestSQLRepository_GetByID_Success(t *testing.T) {
	repository, mock :=
		newVerificationRepositoryTest(t)

	now := time.Now()
	reviewedAt := now.Add(-time.Hour)

	mock.ExpectQuery(
		`(?s)SELECT.*FROM verification_requests.*WHERE id = \$1`,
	).
		WithArgs(uint(10)).
		WillReturnRows(
			sqlmock.NewRows(
				verificationColumns(),
			).AddRow(
				10,
				5,
				"dbs_check",
				"https://test/dbs",
				"approved",
				"verified",
				99,
				reviewedAt,
				now,
				now,
			),
		)

	request, err := repository.GetByID(
		context.Background(),
		10,
	)

	if err != nil {
		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}

	if request.ID != 10 {
		t.Fatalf(
			"expected ID 10 got %d",
			request.ID,
		)
	}

	if request.ReviewedBy == nil ||
		*request.ReviewedBy != 99 {
		t.Fatal(
			"expected reviewed_by 99",
		)
	}

	if request.ReviewedAt == nil {
		t.Fatal(
			"expected reviewed_at",
		)
	}
}

func TestSQLRepository_GetByID_NullReviewFields(
	t *testing.T,
) {
	repository, mock :=
		newVerificationRepositoryTest(t)

	now := time.Now()

	mock.ExpectQuery(
		`(?s)SELECT.*FROM verification_requests.*WHERE id = \$1`,
	).
		WithArgs(uint(10)).
		WillReturnRows(
			sqlmock.NewRows(
				verificationColumns(),
			).AddRow(
				10,
				5,
				"id_check",
				"https://test/id",
				"pending",
				"",
				nil,
				nil,
				now,
				now,
			),
		)

	request, err := repository.GetByID(
		context.Background(),
		10,
	)

	if err != nil {
		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}

	if request.ReviewedBy != nil {
		t.Fatal(
			"expected nil ReviewedBy",
		)
	}

	if request.ReviewedAt != nil {
		t.Fatal(
			"expected nil ReviewedAt",
		)
	}
}

func TestSQLRepository_GetByID_NotFound(
	t *testing.T,
) {
	repository, mock :=
		newVerificationRepositoryTest(t)

	mock.ExpectQuery(
		`(?s)SELECT.*FROM verification_requests.*WHERE id = \$1`,
	).
		WithArgs(uint(99)).
		WillReturnError(sql.ErrNoRows)

	_, err := repository.GetByID(
		context.Background(),
		99,
	)

	if !errors.Is(
		err,
		ErrVerificationNotFound,
	) {
		t.Fatalf(
			"expected not found got %v",
			err,
		)
	}
}

func TestSQLRepository_GetByID_DatabaseError(
	t *testing.T,
) {
	repository, mock :=
		newVerificationRepositoryTest(t)

	databaseErr := errors.New(
		"database failed",
	)

	mock.ExpectQuery(
		`(?s)SELECT.*FROM verification_requests.*WHERE id = \$1`,
	).
		WithArgs(uint(10)).
		WillReturnError(databaseErr)

	_, err := repository.GetByID(
		context.Background(),
		10,
	)

	if !errors.Is(err, databaseErr) {
		t.Fatalf(
			"expected database error got %v",
			err,
		)
	}
}

func TestSQLRepository_ListByUserID_Success(
	t *testing.T,
) {
	repository, mock :=
		newVerificationRepositoryTest(t)

	now := time.Now()

	mock.ExpectQuery(
		`(?s)SELECT.*FROM verification_requests.*WHERE user_id = \$1.*ORDER BY created_at DESC`,
	).
		WithArgs(uint(5)).
		WillReturnRows(
			sqlmock.NewRows(
				verificationColumns(),
			).AddRow(
				1,
				5,
				"id_check",
				"https://test/id",
				"pending",
				"",
				nil,
				nil,
				now,
				now,
			),
		)

	requests, err :=
		repository.ListByUserID(
			context.Background(),
			5,
		)

	if err != nil {
		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}

	if len(requests) != 1 {
		t.Fatalf(
			"expected one request got %d",
			len(requests),
		)
	}
}

func TestSQLRepository_ListByUserID_QueryError(
	t *testing.T,
) {
	repository, mock :=
		newVerificationRepositoryTest(t)

	queryErr := errors.New("query failed")

	mock.ExpectQuery(
		`(?s)SELECT.*FROM verification_requests.*WHERE user_id = \$1`,
	).
		WithArgs(uint(5)).
		WillReturnError(queryErr)

	_, err := repository.ListByUserID(
		context.Background(),
		5,
	)

	if !errors.Is(err, queryErr) {
		t.Fatalf(
			"expected query error got %v",
			err,
		)
	}
}

func TestSQLRepository_ListByUserID_ScanError(
	t *testing.T,
) {
	repository, mock :=
		newVerificationRepositoryTest(t)

	now := time.Now()

	mock.ExpectQuery(
		`(?s)SELECT.*FROM verification_requests.*WHERE user_id = \$1`,
	).
		WithArgs(uint(5)).
		WillReturnRows(
			sqlmock.NewRows(
				verificationColumns(),
			).AddRow(
				"bad-id",
				5,
				"id_check",
				"https://test/id",
				"pending",
				"",
				nil,
				nil,
				now,
				now,
			),
		)

	_, err := repository.ListByUserID(
		context.Background(),
		5,
	)

	if err == nil {
		t.Fatal("expected scan error")
	}
}

func TestSQLRepository_ListByUserID_RowsError(
	t *testing.T,
) {
	repository, mock :=
		newVerificationRepositoryTest(t)

	now := time.Now()
	rowsErr := errors.New("rows failed")

	rows := sqlmock.NewRows(
		verificationColumns(),
	).
		AddRow(
			1,
			5,
			"id_check",
			"https://test/id",
			"pending",
			"",
			nil,
			nil,
			now,
			now,
		).
		RowError(0, rowsErr)

	mock.ExpectQuery(
		`(?s)SELECT.*FROM verification_requests.*WHERE user_id = \$1`,
	).
		WithArgs(uint(5)).
		WillReturnRows(rows)

	_, err := repository.ListByUserID(
		context.Background(),
		5,
	)

	if !errors.Is(err, rowsErr) {
		t.Fatalf(
			"expected rows error got %v",
			err,
		)
	}
}

func TestSQLRepository_ListAll_Success(t *testing.T) {
	repository, mock :=
		newVerificationRepositoryTest(t)

	now := time.Now()

	mock.ExpectQuery(
		`(?s)SELECT.*FROM verification_requests.*ORDER BY created_at DESC`,
	).
		WillReturnRows(
			sqlmock.NewRows(
				verificationColumns(),
			).AddRow(
				1,
				5,
				"insurance",
				"https://test/insurance",
				"approved",
				"ok",
				99,
				now,
				now,
				now,
			),
		)

	requests, err := repository.ListAll(
		context.Background(),
	)

	if err != nil {
		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}

	if len(requests) != 1 {
		t.Fatalf(
			"expected one request got %d",
			len(requests),
		)
	}
}

func TestSQLRepository_ListPending_Success(
	t *testing.T,
) {
	repository, mock :=
		newVerificationRepositoryTest(t)

	now := time.Now()

	mock.ExpectQuery(
		`(?s)SELECT.*FROM verification_requests.*WHERE status = 'pending'.*ORDER BY created_at DESC`,
	).
		WillReturnRows(
			sqlmock.NewRows(
				verificationColumns(),
			).AddRow(
				1,
				5,
				"dbs_check",
				"https://test/dbs",
				"pending",
				"",
				nil,
				nil,
				now,
				now,
			),
		)

	requests, err :=
		repository.ListPending(
			context.Background(),
		)

	if err != nil {
		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}

	if len(requests) != 1 {
		t.Fatalf(
			"expected one pending request got %d",
			len(requests),
		)
	}
}

func TestSQLRepository_Review_Success(t *testing.T) {
	repository, mock :=
		newVerificationRepositoryTest(t)

	mock.ExpectExec(
		`(?s)UPDATE verification_requests.*status = \$1.*WHERE id = \$5`,
	).
		WithArgs(
			"approved",
			"verified",
			uint(99),
			sqlmock.AnyArg(),
			uint(10),
		).
		WillReturnResult(
			sqlmock.NewResult(0, 1),
		)

	err := repository.Review(
		context.Background(),
		10,
		"approved",
		"verified",
		99,
	)

	if err != nil {
		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}
}

func TestSQLRepository_Review_NotFound(t *testing.T) {
	repository, mock :=
		newVerificationRepositoryTest(t)

	mock.ExpectExec(
		`(?s)UPDATE verification_requests.*WHERE id = \$5`,
	).
		WithArgs(
			"approved",
			"verified",
			uint(99),
			sqlmock.AnyArg(),
			uint(999),
		).
		WillReturnResult(
			sqlmock.NewResult(0, 0),
		)

	err := repository.Review(
		context.Background(),
		999,
		"approved",
		"verified",
		99,
	)

	if !errors.Is(
		err,
		ErrVerificationNotFound,
	) {
		t.Fatalf(
			"expected not found got %v",
			err,
		)
	}
}

func TestSQLRepository_Review_DatabaseError(
	t *testing.T,
) {
	repository, mock :=
		newVerificationRepositoryTest(t)

	databaseErr := errors.New(
		"update failed",
	)

	mock.ExpectExec(
		`(?s)UPDATE verification_requests.*WHERE id = \$5`,
	).
		WithArgs(
			"approved",
			"verified",
			uint(99),
			sqlmock.AnyArg(),
			uint(10),
		).
		WillReturnError(databaseErr)

	err := repository.Review(
		context.Background(),
		10,
		"approved",
		"verified",
		99,
	)

	if !errors.Is(err, databaseErr) {
		t.Fatalf(
			"expected database error got %v",
			err,
		)
	}
}

var _ Repository = (*SQLRepository)(nil)
