package companyaccounts

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestSQLRepository_CreateCompany_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	now := time.Now().UTC()

	company := &Company{
		OwnerID:     5,
		Name:        "Wembi Cleaning Ltd",
		Description: "East London cleaning company",
	}

	mock.ExpectQuery(
		regexp.QuoteMeta(
			"INSERT INTO companies",
		),
	).
		WithArgs(
			uint(5),
			"Wembi Cleaning Ltd",
			"East London cleaning company",
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

	err = repo.CreateCompany(
		context.Background(),
		company,
	)
	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}

	if company.ID != 10 {
		t.Fatalf(
			"expected company ID 10, got %d",
			company.ID,
		)
	}
}

func TestSQLRepository_AddMember_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	mock.ExpectExec(
		regexp.QuoteMeta(
			"INSERT INTO company_members",
		),
	).
		WithArgs(
			uint(10),
			uint(8),
			"cleaner",
			sqlmock.AnyArg(),
		).
		WillReturnResult(
			sqlmock.NewResult(
				1,
				1,
			),
		)

	err = repo.AddMember(
		context.Background(),
		10,
		8,
		"cleaner",
	)
	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}
}

func TestSQLRepository_AddMember_Duplicate(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	mock.ExpectExec(
		regexp.QuoteMeta(
			"INSERT INTO company_members",
		),
	).
		WillReturnError(
			errors.New(
				"duplicate key value violates unique constraint",
			),
		)

	err = repo.AddMember(
		context.Background(),
		10,
		8,
		"cleaner",
	)

	if !errors.Is(
		err,
		ErrMemberExists,
	) {
		t.Fatalf(
			"expected ErrMemberExists, got %v",
			err,
		)
	}
}

func TestSQLRepository_GetByID_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	now := time.Now().UTC()

	mock.ExpectQuery(
		regexp.QuoteMeta("SELECT"),
	).
		WithArgs(uint(10)).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{
					"id",
					"owner_id",
					"name",
					"description",
					"created_at",
					"updated_at",
				},
			).AddRow(
				10,
				5,
				"Wembi Cleaning Ltd",
				"East London",
				now,
				now,
			),
		)

	result, err :=
		repo.GetByID(
			context.Background(),
			10,
		)
	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}

	if result.ID != 10 {
		t.Fatalf(
			"expected company ID 10, got %d",
			result.ID,
		)
	}

	if result.OwnerID != 5 {
		t.Fatalf(
			"expected owner ID 5, got %d",
			result.OwnerID,
		)
	}
}

func TestSQLRepository_GetByID_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	mock.ExpectQuery(
		regexp.QuoteMeta("SELECT"),
	).
		WithArgs(uint(10)).
		WillReturnError(
			sql.ErrNoRows,
		)

	result, err :=
		repo.GetByID(
			context.Background(),
			10,
		)

	if result != nil {
		t.Fatalf(
			"expected nil result, got %+v",
			result,
		)
	}

	if !errors.Is(
		err,
		ErrCompanyNotFound,
	) {
		t.Fatalf(
			"expected ErrCompanyNotFound, got %v",
			err,
		)
	}
}

func TestSQLRepository_GetMember_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	now := time.Now().UTC()

	mock.ExpectQuery(
		regexp.QuoteMeta("SELECT"),
	).
		WithArgs(
			uint(10),
			uint(8),
		).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{
					"id",
					"company_id",
					"user_id",
					"full_name",
					"role",
					"status",
					"is_verified",
					"reliability_score",
					"badge",
					"created_at",
				},
			).AddRow(
				2,
				10,
				8,
				"Sarah Cleaner",
				"cleaner",
				"active",
				true,
				94,
				"Elite Cleaner",
				now,
			),
		)

	result, err :=
		repo.GetMember(
			context.Background(),
			10,
			8,
		)
	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}

	if result.UserID != 8 {
		t.Fatalf(
			"expected user ID 8, got %d",
			result.UserID,
		)
	}

	if result.Role != "cleaner" {
		t.Fatalf(
			"expected cleaner role, got %q",
			result.Role,
		)
	}

	if result.ReliabilityScore != 94 {
		t.Fatalf(
			"expected reliability 94, got %d",
			result.ReliabilityScore,
		)
	}
}

func TestSQLRepository_GetMember_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	mock.ExpectQuery(
		regexp.QuoteMeta("SELECT"),
	).
		WithArgs(
			uint(10),
			uint(8),
		).
		WillReturnError(
			sql.ErrNoRows,
		)

	result, err :=
		repo.GetMember(
			context.Background(),
			10,
			8,
		)

	if result != nil {
		t.Fatalf(
			"expected nil result, got %+v",
			result,
		)
	}

	if !errors.Is(
		err,
		ErrMemberNotFound,
	) {
		t.Fatalf(
			"expected ErrMemberNotFound, got %v",
			err,
		)
	}
}

func TestSQLRepository_ListMembers_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	now := time.Now().UTC()

	mock.ExpectQuery(
		regexp.QuoteMeta("SELECT"),
	).
		WithArgs(uint(10)).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{
					"id",
					"company_id",
					"user_id",
					"full_name",
					"role",
					"status",
					"is_verified",
					"reliability_score",
					"badge",
					"created_at",
				},
			).
				AddRow(
					1,
					10,
					5,
					"Company Owner",
					"owner",
					"active",
					false,
					0,
					"New Cleaner",
					now,
				).
				AddRow(
					2,
					10,
					8,
					"Sarah Cleaner",
					"cleaner",
					"active",
					true,
					94,
					"Elite Cleaner",
					now,
				),
		)

	results, err :=
		repo.ListMembers(
			context.Background(),
			10,
		)
	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}

	if len(results) != 2 {
		t.Fatalf(
			"expected 2 members, got %d",
			len(results),
		)
	}

	if results[0].Role != "owner" {
		t.Fatalf(
			"expected owner first, got %q",
			results[0].Role,
		)
	}
}

func TestSQLRepository_RemoveMember_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	mock.ExpectExec(
		regexp.QuoteMeta(
			"DELETE FROM company_members",
		),
	).
		WithArgs(
			uint(10),
			uint(8),
		).
		WillReturnResult(
			sqlmock.NewResult(
				0,
				1,
			),
		)

	err = repo.RemoveMember(
		context.Background(),
		10,
		8,
	)
	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}
}

func TestSQLRepository_RemoveMember_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	mock.ExpectExec(
		regexp.QuoteMeta(
			"DELETE FROM company_members",
		),
	).
		WithArgs(
			uint(10),
			uint(8),
		).
		WillReturnResult(
			sqlmock.NewResult(
				0,
				0,
			),
		)

	err = repo.RemoveMember(
		context.Background(),
		10,
		8,
	)

	if !errors.Is(
		err,
		ErrMemberNotFound,
	) {
		t.Fatalf(
			"expected ErrMemberNotFound, got %v",
			err,
		)
	}
}

func TestSQLRepository_CountActiveCleanerSeats_Success(
	t *testing.T,
) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	mock.ExpectQuery(
		regexp.QuoteMeta(
			"SELECT COUNT(*)",
		),
	).
		WithArgs(uint(10)).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{"count"},
			).AddRow(
				2,
			),
		)

	count, err :=
		repo.CountActiveCleanerSeats(
			context.Background(),
			10,
		)
	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}

	if count != 2 {
		t.Fatalf(
			"expected 2 active cleaner seats, got %d",
			count,
		)
	}
}

func TestSQLRepository_GetCleanerSeatLimit_Success(
	t *testing.T,
) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	mock.ExpectQuery(
		regexp.QuoteMeta("SELECT"),
	).
		WithArgs(uint(10)).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{
					"cleaner_seat_limit",
				},
			).AddRow(
				3,
			),
		)

	limit, err :=
		repo.GetCleanerSeatLimit(
			context.Background(),
			10,
		)
	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}

	if limit != 3 {
		t.Fatalf(
			"expected seat limit 3, got %d",
			limit,
		)
	}
}

func TestSQLRepository_GetCleanerSeatLimit_NotFound(
	t *testing.T,
) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	mock.ExpectQuery(
		regexp.QuoteMeta("SELECT"),
	).
		WithArgs(uint(10)).
		WillReturnError(
			sql.ErrNoRows,
		)

	limit, err :=
		repo.GetCleanerSeatLimit(
			context.Background(),
			10,
		)

	if limit != 0 {
		t.Fatalf(
			"expected limit 0, got %d",
			limit,
		)
	}

	if !errors.Is(
		err,
		ErrSeatPlanNotFound,
	) {
		t.Fatalf(
			"expected ErrSeatPlanNotFound, got %v",
			err,
		)
	}
}

func TestSQLRepository_UpdateMemberStatus_Success(
	t *testing.T,
) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	mock.ExpectExec(
		regexp.QuoteMeta(
			"UPDATE company_members",
		),
	).
		WithArgs(
			"inactive",
			uint(10),
			uint(8),
		).
		WillReturnResult(
			sqlmock.NewResult(
				0,
				1,
			),
		)

	err = repo.UpdateMemberStatus(
		context.Background(),
		10,
		8,
		"inactive",
	)
	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}
}

func TestSQLRepository_UpdateMemberRole_Success(
	t *testing.T,
) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	mock.ExpectExec(
		regexp.QuoteMeta(
			"UPDATE company_members",
		),
	).
		WithArgs(
			"admin",
			uint(10),
			uint(8),
		).
		WillReturnResult(
			sqlmock.NewResult(
				0,
				1,
			),
		)

	err = repo.UpdateMemberRole(
		context.Background(),
		10,
		8,
		"admin",
	)
	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}
}

func TestSQLRepository_CountMembersByRoleAndStatus_Success(
	t *testing.T,
) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	mock.ExpectQuery(
		regexp.QuoteMeta(
			"SELECT COUNT(*)",
		),
	).
		WithArgs(
			uint(10),
			"cleaner",
			"active",
		).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{"count"},
			).AddRow(
				3,
			),
		)

	count, err :=
		repo.CountMembersByRolesAndStatus(
			context.Background(),
			10,
			"cleaner",
			"active",
		)
	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}

	if count != 3 {
		t.Fatalf(
			"expected count 3, got %d",
			count,
		)
	}
}

