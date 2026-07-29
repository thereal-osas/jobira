package companyaccounts

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"
)

type SQLRepository struct {
	db *sql.DB
}

func NewSQLRepository(db *sql.DB) *SQLRepository {
	return &SQLRepository{db: db}
}

func (r *SQLRepository) CreateCompany(ctx context.Context, company *Company) error {
	query := `
		INSERT INTO companies (
			owner_id,
			name,
			description,
			created_at,
			updated_at
		)
			VALUES ($1, $2, $3, $4, $4)
			RETURNING id, created_at, updated_at
	`

	now := time.Now()

	err := r.db.QueryRowContext(
		ctx,
		query,
		company.OwnerID,
		company.Name,
		company.Description,
		now,
	).Scan(
		&company.ID,
		&company.CreatedAt,
		&company.UpdatedAt,
	)

	if err != nil {
		return err
	}

	return nil
}

func (r *SQLRepository) AddMember(ctx context.Context, companyID uint, userID uint, role string) error {
	_ = role
	query := `
	 	INSERT INTO company_members (
			company_id,
			user_id,
			created_at
		)
			VALUES ($1, $2, $3)
	 `

	_, err := r.db.ExecContext(ctx, query, companyID, userID, time.Now())
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "duplicate key") {
			return ErrMemberExists
		}

		return err
	}
	return nil
}

func (r *SQLRepository) RemoveMember(ctx context.Context, companyID uint, userID uint) error {
	query := `
		DELETE FROM company_members
		WHERE company_id  = $1
		AND user_id = $2
	`

	result, err := r.db.ExecContext(ctx, query, companyID, userID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrCompanyNotFound
	}

	return nil
}

func (r *SQLRepository) ListMembers(ctx context.Context, companyID uint) ([]CompanyMember, error) {
	query := `
		SELECT
			id, 
			company_id,
			user_id,
			created_at
		FROM company_members
		WHERE company_id = $1 
		ORDER BY created_at DESC	
	`

	rows, err := r.db.QueryContext(ctx, query, companyID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrCompanyNotFound
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []CompanyMember

	for rows.Next() {
		var member CompanyMember

		err := rows.Scan(
			&member.ID,
			&member.CompanyID,
			&member.UserID,
			&member.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		members = append(members, member)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return members, nil
}

func (r *SQLRepository) CountMembers(ctx context.Context, companyID uint) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM company_members
		WHERE company_id = $1
	`

	var count int

	err := r.db.QueryRowContext(ctx, query, companyID).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *SQLRepository) GetCleanerSeatLimit(ctx context.Context, companyID uint) (int, error) {
	query := `
		SELECT sp.cleaner_seat_limit
		FROM companies c 
		JOIN user_subscriptions us
			ON us.user_id = c.owner_id 
		JOIN subscription_plans sp
			ON sp.id = us.plan_id 
		WHERE c.id = $1 
		AND sp.is_active = true 
		ORDER BY us.updated_at DESC 
		LIMIT 1 		
	`

	var seatLimit int

	err := r.db.QueryRowContext(ctx, query, companyID).Scan(&seatLimit)
	if err != nil {
		return 0, err
	}

	return seatLimit, nil
}

func (r *SQLRepository) GetByID(ctx context.Context, companyID uint) (*Company, error) {
	query := `
		SELECT 
			id,
			owner_id,
			name,
			description,
			created_at,
			updated_at
		FROM companies
		WHERE id = $1
		LIMIT 1		
	`

	var company Company

	err := r.db.QueryRowContext(ctx, query, companyID).Scan(
		&company.ID,
		&company.OwnerID,
		&company.Name,
		&company.Description,
		&company.CreatedAt,
		&company.UpdatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrCompanyNotFound
	}

	if err != nil {
		return nil, err
	}

	return &company, nil
}
