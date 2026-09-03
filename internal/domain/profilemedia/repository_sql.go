package profilemedia

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type SQLRepository struct {
	db *sql.DB
}

func NewSQLRepository(db *sql.DB) *SQLRepository {
	return &SQLRepository{
		db: db,
	}
}

func (r *SQLRepository) Create(ctx context.Context, media *ProfileMedia) error {
	query := `
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
	`

	err := r.db.QueryRowContext(
		ctx,
		query,
		media.OwnerUserID,
		media.CompanyID,
		media.MediaType,
		media.URL,
		media.Caption,
		media.SortOrder,
	).Scan(
		&media.ID,
		&media.CreatedAt,
		&media.UpdatedAt,
	)

	return err
}

func (r *SQLRepository) GetByID(ctx context.Context, id uint) (*ProfileMedia, error) {
	query := `
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
	`

	media := &ProfileMedia{}

	var OwnerUserID sql.NullInt64
	var companyID sql.NullInt64

	err := r.db.QueryRowContext(
		ctx,
		query,
		id,
	).Scan(
		&media.ID,
		&OwnerUserID,
		&companyID,
		&media.MediaType,
		&media.URL,
		&media.Caption,
		&media.SortOrder,
		&media.CreatedAt,
		&media.UpdatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrMediaNotFound
	}

	if err != nil {
		return nil, err
	}

	if OwnerUserID.Valid {
		value := uint(OwnerUserID.Int64)
		media.OwnerUserID = &value
	}

	if companyID.Valid {
		value := uint(companyID.Int64)
		media.CompanyID = &value
	}

	return media, nil
}

func (r *SQLRepository) ListByUserID(ctx context.Context, userID uint) ([]ProfileMedia, error) {
	query := `
		SELECT
			id,
			owner_user_id,
			company_id,
			media_type,
			url,
			caption,
			sort_order,
			created_at,
			updated_at,
		FROM profile_media
		WHERE owner_user_id = $1
		ORDER BY
			media_type ASC, 
			sort_order ASC,
			created_at ASC	
	`

	rows, err := r.db.QueryContext(
		ctx,
		query,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanMediaRows(rows)

}

func (r *SQLRepository) ListByCompanyID(ctx context.Context, companyID uint) ([]ProfileMedia, error) {
	query := `
		SELECT
			id,
			owner_user_id,
			company_id,
			media_type,
			url,
			caption,
			sort_order,
			created_at,
			updated_at,
		FROM profile_media
		WHERE company_id = $1
		ORDER BY
			media_type ASC, 
			sort_order ASC,
			created_at ASC	
	`

	rows, err := r.db.QueryContext(
		ctx,
		query,
		companyID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanMediaRows(rows)
}

func (r *SQLRepository) CountByUserAndType(ctx context.Context, userID uint, mediaType MediaType) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM profile_media
		WHERE owner_user_id = $1
			AND media_type = $2
	`

	var count int
	err := r.db.QueryRowContext(
		ctx,
		query,
		userID,
		mediaType,
	).Scan(&count)

	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *SQLRepository) CountPortfolioPhotos(ctx context.Context,userID uint,) (int, error) {
	if userID == 0 {
		return 0, ErrInvalidInput
	}

	return r.CountByUserAndType(
		ctx,
		userID,
		MediaTypePortfolio,
	)
}

func (r *SQLRepository) CountByCompanyAndType(ctx context.Context, companyID uint, mediaType MediaType) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM profile_media
		WHERE company_id = $1
			AND media_type = $2
	`

	var count int

	err := r.db.QueryRowContext(
		ctx,
		query,
		companyID,
		mediaType,
	).Scan(&count)

	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *SQLRepository) CanManageCompany(ctx context.Context, userID uint, companyID uint) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT 1
			FROM companies c
			LEFT JOIN company_members cm
				ON cm.company_id = c.id
				AND cm.user_id = $1
				AND cm.status = 'active'
			WHERE c.id = $2
			  AND (
				c.owner_id = $1
				OR cm.role IN ('owner', 'admin')
			  )
		)
	`

	var canManage bool

	err := r.db.QueryRowContext(
		ctx,
		query,
		userID,
		companyID,
	).Scan(&canManage)
	if err != nil {
		return false, err
	}

	return canManage, nil
}

func (r *SQLRepository) Update(ctx context.Context, media *ProfileMedia) error {
	query := `
		UPDATE profile_media
		SET
			caption = $1,
			sort_order = $2,
			updated_at = $3
		WHERE id = $4	
	`

	now := time.Now()

	result, err := r.db.ExecContext(
		ctx,
		query,
		media.Caption,
		media.SortOrder,
		now,
		media.ID,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrMediaNotFound
	}

	media.UpdatedAt = now

	return nil
}

func (r *SQLRepository) Delete(ctx context.Context, id uint) error {
	query := `
		DELETE FROM profile_media
		WHERE id = $1
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		id,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrMediaNotFound
	}

	return nil
}

func scanMediaRows(rows *sql.Rows) ([]ProfileMedia, error) {
	mediaItems := make([]ProfileMedia, 0)

	for rows.Next() {
		var media ProfileMedia

		var OwnerUserID sql.NullInt64
		var companyID sql.NullInt64

		err := rows.Scan(
			&media.ID,
			&OwnerUserID,
			&companyID,
			&media.MediaType,
			&media.URL,
			&media.Caption,
			&media.SortOrder,
			&media.CreatedAt,
			&media.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if OwnerUserID.Valid {
			value := uint(OwnerUserID.Int64)
			media.OwnerUserID = &value
		}

		if companyID.Valid {
			value := uint(companyID.Int64)
			media.CompanyID = &value
		}

		mediaItems = append(
			mediaItems,
			media,
		)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return mediaItems, nil
}

var _ Repository = (*SQLRepository)(nil)
