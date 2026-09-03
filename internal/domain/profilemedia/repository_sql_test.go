package profilemedia

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func newMockSQLRepository(
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

	repo := NewSQLRepository(db)

	cleanup := func() {
		_ = db.Close()
	}

	return repo, mock, cleanup
}

func TestSQLRepository_Create_Success(t *testing.T) {
	repo, mock, cleanup := newMockSQLRepository(t)
	defer cleanup()

	now := time.Now()
	userID := uint(10)

	mock.ExpectQuery(
		regexp.QuoteMeta(`
		INSERT INTO profile_media (
			owner_user_id,
			company_id,
			media_type,
			url,
			caption,
			sort_order
		)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING
			id,
			created_at,
			updated_at
	`),
	).
		WithArgs(
			&userID,
			nil,
			MediaTypePortfolio,
			"https://example.com/work.jpg",
			"Kitchen clean",
			1,
		).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{
					"id",
					"created_at",
					"updated_at",
				},
			).AddRow(
				5,
				now,
				now,
			),
		)

	media := &ProfileMedia{
		OwnerUserID: &userID,
		MediaType:   MediaTypePortfolio,
		URL:         "https://example.com/work.jpg",
		Caption:     "Kitchen clean",
		SortOrder:   1,
	}

	err := repo.Create(
		context.Background(),
		media,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if media.ID != 5 {
		t.Fatalf(
			"expected ID 5 got %d",
			media.ID,
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf(
			"unmet SQL expectations: %v",
			err,
		)
	}
}

func TestSQLRepository_Create_Error(t *testing.T) {
	repo, mock, cleanup := newMockSQLRepository(t)
	defer cleanup()

	userID := uint(10)
	expectedErr := errors.New("database failed")

	mock.ExpectQuery(
		"INSERT INTO profile_media",
	).
		WithArgs(
			&userID,
			nil,
			MediaTypePortfolio,
			"https://example.com/work.jpg",
			"",
			0,
		).
		WillReturnError(expectedErr)

	media := &ProfileMedia{
		OwnerUserID: &userID,
		MediaType:   MediaTypePortfolio,
		URL:         "https://example.com/work.jpg",
	}

	err := repo.Create(
		context.Background(),
		media,
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected %v got %v",
			expectedErr,
			err,
		)
	}
}

func TestSQLRepository_GetByID_Success(t *testing.T) {
	repo, mock, cleanup := newMockSQLRepository(t)
	defer cleanup()

	now := time.Now()

	mock.ExpectQuery(
		regexp.QuoteMeta(`
		SELECT
			id,
			owner_user_id,
			company_id,
			media_type,
			url,
			caption,
			sort_order,
			created_at,
			updated_at
		FROM profile_media
		WHERE id = $1
	`),
	).
		WithArgs(uint(5)).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{
					"id",
					"owner_user_id",
					"company_id",
					"media_type",
					"url",
					"caption",
					"sort_order",
					"created_at",
					"updated_at",
				},
			).AddRow(
				5,
				10,
				nil,
				MediaTypePortfolio,
				"https://example.com/work.jpg",
				"Work photo",
				1,
				now,
				now,
			),
		)

	media, err := repo.GetByID(
		context.Background(),
		5,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if media.ID != 5 {
		t.Fatalf(
			"expected ID 5 got %d",
			media.ID,
		)
	}

	if media.OwnerUserID == nil ||
		*media.OwnerUserID != 10 {
		t.Fatal("expected owner user ID 10")
	}

	if media.CompanyID != nil {
		t.Fatal("expected company ID nil")
	}
}

func TestSQLRepository_GetByID_CompanyMedia(t *testing.T) {
	repo, mock, cleanup := newMockSQLRepository(t)
	defer cleanup()

	now := time.Now()

	mock.ExpectQuery(
		"SELECT(.|\\s)*FROM profile_media(.|\\s)*WHERE id = \\$1",
	).
		WithArgs(uint(8)).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{
					"id",
					"owner_user_id",
					"company_id",
					"media_type",
					"url",
					"caption",
					"sort_order",
					"created_at",
					"updated_at",
				},
			).AddRow(
				8,
				nil,
				7,
				MediaTypeCompanyLogo,
				"https://example.com/logo.jpg",
				"",
				0,
				now,
				now,
			),
		)

	media, err := repo.GetByID(
		context.Background(),
		8,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if media.CompanyID == nil ||
		*media.CompanyID != 7 {
		t.Fatal("expected company ID 7")
	}

	if media.OwnerUserID != nil {
		t.Fatal("expected owner user ID nil")
	}
}

func TestSQLRepository_GetByID_NotFound(t *testing.T) {
	repo, mock, cleanup := newMockSQLRepository(t)
	defer cleanup()

	mock.ExpectQuery(
		"SELECT(.|\\s)*FROM profile_media(.|\\s)*WHERE id = \\$1",
	).
		WithArgs(uint(404)).
		WillReturnError(sql.ErrNoRows)

	_, err := repo.GetByID(
		context.Background(),
		404,
	)

	if !errors.Is(err, ErrMediaNotFound) {
		t.Fatalf(
			"expected ErrMediaNotFound got %v",
			err,
		)
	}
}

func TestSQLRepository_GetByID_Error(t *testing.T) {
	repo, mock, cleanup := newMockSQLRepository(t)
	defer cleanup()

	expectedErr := errors.New("query failed")

	mock.ExpectQuery(
		"SELECT(.|\\s)*FROM profile_media(.|\\s)*WHERE id = \\$1",
	).
		WithArgs(uint(5)).
		WillReturnError(expectedErr)

	_, err := repo.GetByID(
		context.Background(),
		5,
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected %v got %v",
			expectedErr,
			err,
		)
	}
}

func TestSQLRepository_ListByUserID_Success(t *testing.T) {
	repo, mock, cleanup := newMockSQLRepository(t)
	defer cleanup()

	now := time.Now()

	rows := sqlmock.NewRows(
		[]string{
			"id",
			"owner_user_id",
			"company_id",
			"media_type",
			"url",
			"caption",
			"sort_order",
			"created_at",
			"updated_at",
		},
	).
		AddRow(
			1,
			10,
			nil,
			MediaTypeProfilePhoto,
			"https://example.com/profile.jpg",
			"",
			0,
			now,
			now,
		).
		AddRow(
			2,
			10,
			nil,
			MediaTypePortfolio,
			"https://example.com/work.jpg",
			"Work",
			1,
			now,
			now,
		)

	mock.ExpectQuery(
		"SELECT(.|\\s)*FROM profile_media(.|\\s)*WHERE owner_user_id = \\$1",
	).
		WithArgs(uint(10)).
		WillReturnRows(rows)

	result, err := repo.ListByUserID(
		context.Background(),
		10,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 2 {
		t.Fatalf(
			"expected 2 items got %d",
			len(result),
		)
	}
}

func TestSQLRepository_ListByCompanyID_Success(t *testing.T) {
	repo, mock, cleanup := newMockSQLRepository(t)
	defer cleanup()

	now := time.Now()

	rows := sqlmock.NewRows(
		[]string{
			"id",
			"owner_user_id",
			"company_id",
			"media_type",
			"url",
			"caption",
			"sort_order",
			"created_at",
			"updated_at",
		},
	).AddRow(
		1,
		nil,
		7,
		MediaTypeCompanyLogo,
		"https://example.com/logo.jpg",
		"",
		0,
		now,
		now,
	)

	mock.ExpectQuery(
		"SELECT(.|\\s)*FROM profile_media(.|\\s)*WHERE company_id = \\$1",
	).
		WithArgs(uint(7)).
		WillReturnRows(rows)

	result, err := repo.ListByCompanyID(
		context.Background(),
		7,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf(
			"expected 1 item got %d",
			len(result),
		)
	}
}

func TestSQLRepository_CountByUserAndType_Success(
	t *testing.T,
) {
	repo, mock, cleanup := newMockSQLRepository(t)
	defer cleanup()

	mock.ExpectQuery(
		"SELECT COUNT\\(\\*\\)(.|\\s)*FROM profile_media(.|\\s)*owner_user_id = \\$1(.|\\s)*media_type = \\$2",
	).
		WithArgs(
			uint(10),
			MediaTypePortfolio,
		).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{"count"},
			).AddRow(4),
		)

	count, err := repo.CountByUserAndType(
		context.Background(),
		10,
		MediaTypePortfolio,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if count != 4 {
		t.Fatalf(
			"expected count 4 got %d",
			count,
		)
	}
}

func TestSQLRepository_CountByCompanyAndType_Success(
	t *testing.T,
) {
	repo, mock, cleanup := newMockSQLRepository(t)
	defer cleanup()

	mock.ExpectQuery(
		"SELECT COUNT\\(\\*\\)(.|\\s)*FROM profile_media(.|\\s)*company_id = \\$1(.|\\s)*media_type = \\$2",
	).
		WithArgs(
			uint(7),
			MediaTypePortfolio,
		).
		WillReturnRows(
			sqlmock.NewRows(
				[]string{"count"},
			).AddRow(6),
		)

	count, err := repo.CountByCompanyAndType(
		context.Background(),
		7,
		MediaTypePortfolio,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if count != 6 {
		t.Fatalf(
			"expected count 6 got %d",
			count,
		)
	}
}

func TestSQLRepository_Update_Success(t *testing.T) {
	repo, mock, cleanup := newMockSQLRepository(t)
	defer cleanup()

	mock.ExpectExec(
		"UPDATE profile_media",
	).
		WithArgs(
			"Updated caption",
			2,
			sqlmock.AnyArg(),
			uint(5),
		).
		WillReturnResult(
			sqlmock.NewResult(0, 1),
		)

	media := &ProfileMedia{
		ID:        5,
		Caption:   "Updated caption",
		SortOrder: 2,
	}

	err := repo.Update(
		context.Background(),
		media,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if media.UpdatedAt.IsZero() {
		t.Fatal(
			"expected UpdatedAt to be populated",
		)
	}
}

func TestSQLRepository_Update_NotFound(t *testing.T) {
	repo, mock, cleanup := newMockSQLRepository(t)
	defer cleanup()

	mock.ExpectExec(
		"UPDATE profile_media",
	).
		WithArgs(
			"",
			0,
			sqlmock.AnyArg(),
			uint(404),
		).
		WillReturnResult(
			sqlmock.NewResult(0, 0),
		)

	err := repo.Update(
		context.Background(),
		&ProfileMedia{
			ID: 404,
		},
	)

	if !errors.Is(err, ErrMediaNotFound) {
		t.Fatalf(
			"expected ErrMediaNotFound got %v",
			err,
		)
	}
}

func TestSQLRepository_Update_Error(t *testing.T) {
	repo, mock, cleanup := newMockSQLRepository(t)
	defer cleanup()

	expectedErr := errors.New("update failed")

	mock.ExpectExec(
		"UPDATE profile_media",
	).
		WithArgs(
			"",
			0,
			sqlmock.AnyArg(),
			uint(5),
		).
		WillReturnError(expectedErr)

	err := repo.Update(
		context.Background(),
		&ProfileMedia{
			ID: 5,
		},
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected %v got %v",
			expectedErr,
			err,
		)
	}
}

func TestSQLRepository_Delete_Success(t *testing.T) {
	repo, mock, cleanup := newMockSQLRepository(t)
	defer cleanup()

	mock.ExpectExec(
		regexp.QuoteMeta(`
		DELETE FROM profile_media
		WHERE id = $1
	`),
	).
		WithArgs(uint(5)).
		WillReturnResult(
			sqlmock.NewResult(0, 1),
		)

	err := repo.Delete(
		context.Background(),
		5,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSQLRepository_Delete_NotFound(t *testing.T) {
	repo, mock, cleanup := newMockSQLRepository(t)
	defer cleanup()

	mock.ExpectExec(
		"DELETE FROM profile_media",
	).
		WithArgs(uint(404)).
		WillReturnResult(
			sqlmock.NewResult(0, 0),
		)

	err := repo.Delete(
		context.Background(),
		404,
	)

	if !errors.Is(err, ErrMediaNotFound) {
		t.Fatalf(
			"expected ErrMediaNotFound got %v",
			err,
		)
	}
}

func TestSQLRepository_Delete_Error(t *testing.T) {
	repo, mock, cleanup := newMockSQLRepository(t)
	defer cleanup()

	expectedErr := errors.New("delete failed")

	mock.ExpectExec(
		"DELETE FROM profile_media",
	).
		WithArgs(uint(5)).
		WillReturnError(expectedErr)

	err := repo.Delete(
		context.Background(),
		5,
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected %v got %v",
			expectedErr,
			err,
		)
	}
}

func TestSQLRepository_ListByUserID_QueryError(
	t *testing.T,
) {
	repo, mock, cleanup := newMockSQLRepository(t)
	defer cleanup()

	expectedErr := errors.New("list failed")

	mock.ExpectQuery(
		"SELECT(.|\\s)*FROM profile_media(.|\\s)*WHERE owner_user_id = \\$1",
	).
		WithArgs(uint(10)).
		WillReturnError(expectedErr)

	_, err := repo.ListByUserID(
		context.Background(),
		10,
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected %v got %v",
			expectedErr,
			err,
		)
	}
}

func TestSQLRepository_CountByUserAndType_Error(
	t *testing.T,
) {
	repo, mock, cleanup := newMockSQLRepository(t)
	defer cleanup()

	expectedErr := errors.New("count failed")

	mock.ExpectQuery(
		"SELECT COUNT\\(\\*\\)",
	).
		WithArgs(
			uint(10),
			MediaTypePortfolio,
		).
		WillReturnError(expectedErr)

	_, err := repo.CountByUserAndType(
		context.Background(),
		10,
		MediaTypePortfolio,
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected %v got %v",
			expectedErr,
			err,
		)
	}
}
