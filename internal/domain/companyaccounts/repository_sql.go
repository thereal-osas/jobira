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

func (r *SQLRepository) AddMember(
	ctx context.Context,
	companyID uint,
	userID uint,
	role string,
) error {
	query := `
		INSERT INTO company_members (
			company_id,
			user_id,
			role,
			status,
			created_at
		)
		VALUES (
			$1,
			$2,
			$3,
			'active',
			$4
		)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		companyID,
		userID,
		role,
		time.Now(),
	)

	if err != nil {
		if strings.Contains(
			strings.ToLower(err.Error()),
			"duplicate key",
		) {
			return ErrMemberExists
		}

		return err
	}

	return nil
}

func (r *SQLRepository) RemoveMember(
	ctx context.Context,
	companyID uint,
	userID uint,
) error {
	query := `
		DELETE FROM company_members
		WHERE company_id = $1
		AND user_id = $2
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		companyID,
		userID,
	)
	if err != nil {
		return err
	}

	rowsAffected, err :=
		result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrMemberNotFound
	}

	return nil
}

func (r *SQLRepository) ListMembers(
	ctx context.Context,
	companyID uint,
) ([]CompanyMember, error) {
	query := `
		SELECT
			cm.id,
			cm.company_id,
			cm.user_id,
			u.full_name,
			cm.role,
			cm.status,

			COALESCE(
				cp.is_verified,
				FALSE
			),

			COALESCE(
				cp.reliability_score,
				0
			),

			COALESCE(
				cp.badge,
				cr.badge,
				'New Cleaner'
			),

			cm.created_at

		FROM company_members cm

		JOIN users u
			ON u.id = cm.user_id

		LEFT JOIN cleaner_profiles cp
			ON cp.user_id = cm.user_id

		LEFT JOIN cleaner_reputation cr
			ON cr.cleaner_id = cm.user_id

		WHERE cm.company_id = $1

		ORDER BY
			CASE cm.role
				WHEN 'owner' THEN 1
				WHEN 'admin' THEN 2
				ELSE 3
			END,
			cm.status DESC,
			u.full_name ASC
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

	members := make(
		[]CompanyMember,
		0,
	)

	for rows.Next() {
		var member CompanyMember

		if err := rows.Scan(
			&member.ID,
			&member.CompanyID,
			&member.UserID,
			&member.FullName,
			&member.Role,
			&member.Status,
			&member.IsVerified,
			&member.ReliabilityScore,
			&member.Badge,
			&member.CreatedAt,
		); err != nil {
			return nil, err
		}

		members = append(
			members,
			member,
		)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return members, nil
}

func (r *SQLRepository) GetCleanerSeatLimit(ctx context.Context, companyID uint) (int, error) {
	query := `
		SELECT
			sp.cleaner_seat_limit
		FROM companies c

		JOIN user_subscriptions us
			ON us.user_id = c.owner_id

		JOIN subscription_plans sp
			ON sp.id = us.plan_id

		WHERE c.id = $1
		AND sp.is_active = TRUE

		ORDER BY us.updated_at DESC
		LIMIT 1
	`

	var seatLimit int

	err := r.db.QueryRowContext(
		ctx,
		query,
		companyID,
	).Scan(
		&seatLimit,
	)

	if errors.Is(
		err,
		sql.ErrNoRows,
	) {
		return 0,
			ErrSeatPlanNotFound
	}

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

func (r *SQLRepository) GetMember(ctx context.Context, companyID uint, userID uint) (*CompanyMember, error) {
	query := `
		SELECT 
			cm.id,
			cm.company_id,
			cm.user_id,
			u.full_name,
			cm.role,
			cm.status,

			COALESCE(
			cp.is_verified,
			FALSE
			),

			COALESCE(
				cp.reliability_score,
				0
			),

			COALESCE(
			cp.badge,
			cp.badge,
			'New Cleaner'
			),

			cm.created_at

		FROM company_members cm 
		
		JOIN users u 
			ON u.id = cm.user_id 

		LEFT JOIN cleaner_profiles cp 
			ON cp.user_id = cm.user_id
		
		LEFT JOIN cleaner_reputation cr 
			ON cr.cleaner_id = cm.user_id 
		
		WHERE cm.company_id = $1 
		AND cm.user_id = $2 
		LIMIT 1	
	`

	var member CompanyMember

	err := r.db.QueryRowContext(
		ctx,
		query,
		companyID,
		userID,
	).Scan(
		&member.ID,
		&member.CompanyID,
		&member.UserID,
		&member.FullName,
		&member.Role,
		&member.Status,
		&member.IsVerified,
		&member.ReliabilityScore,
		&member.Badge,
		&member.CreatedAt,
	)

	if errors.Is(
		err,
		sql.ErrNoRows,
	) {
		return nil,
			ErrMemberNotFound
	}

	if err != nil {
		return nil, err
	}

	return &member, nil
}

func (r *SQLRepository) CountActiveCleanerSeats(ctx context.Context, companyID uint) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM compnay_members
		WHERE company_id = $1 
		AND role = 'cleaner'
		AND status = 'active'
	`

	var count int

	err := r.db.QueryRowContext(
		ctx,
		query,
		companyID,
	).Scan(
		&count,
	)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *SQLRepository) UpdateMemberRole(ctx context.Context, companyID uint, userID uint, role string) error {
	query := `
		UPDATE company_members
		SET role = $1
		WHERE company_id = $2
		AND user_id = $3
	`

	results, err := r.db.ExecContext(
		ctx,
		query,
		role,
		companyID,
		userID,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := results.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrMemberNotFound
	}

	return nil
}

func (r *SQLRepository) CountMembersByRolesAndStatus(ctx context.Context, companyID uint, role string, status string) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM company_members
		WHERE company_id = $1
		AND ($2 = '' OR role = $2)
		AND ($3 = '' OR status = $3)
	`

	var count int

	err := r.db.QueryRowContext(
		ctx,
		query,
		companyID,
		role,
		status,
	).Scan(
		&count,
	)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *SQLRepository) UpdateMemberStatus(
	ctx context.Context,
	companyID uint,
	userID uint,
	status string,
) error {
	query := `
		UPDATE company_members
		SET status = $1
		WHERE company_id = $2
		AND user_id = $3
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		status,
		companyID,
		userID,
	)
	if err != nil {
		return err
	}

	rowsAffected, err :=
		result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrMemberNotFound
	}

	return nil
}
