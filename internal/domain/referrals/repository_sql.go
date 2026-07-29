package referrals

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

func (r *SQLRepository) CreateCode(ctx context.Context, code *ReferralCode) error {
	query := `
	INSERT INTO referral_codes (
		user_id,
		code,
		reward_type,
		reward_value_pence,
		max_uses,
		times_used,
		is_active,
		created_at,
		updated_at
	)
	VALUES ($1, $2, $3, $4, $5, 0, true, $6, $6)
	RETURNING id, created_at, updated_at	
	`

	now := time.Now()

	err := r.db.QueryRowContext(
		ctx,
		query,
		code.UserID,
		code.Code,
		code.RewardType,
		code.RewardValuePence,
		code.MaxUses,
		now,
	).Scan(
		&code.ID,
		&code.CreatedAt,
		&code.UpdatedAt,
	)

	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "duplicate key") {
			return ErrReferralCodeExists
		}

		return err
	}

	return nil
}

func (r *SQLRepository) GetCodeByCode(ctx context.Context, code string) (*ReferralCode, error) {
	query := `
		SELECT
			id, 
			user_id,
			code,
			reward_type,
			reward_value_pence,
			max_uses,
			times_used,
			is_active,
			created_at,
			updated_at
		FROM referral_codes
		WHERE code = $1
		LIMIT 1	
	`

	var referralCode ReferralCode

	err := r.db.QueryRowContext(ctx, query, code).Scan(
		&referralCode.ID,
		&referralCode.UserID,
		&referralCode.Code,
		&referralCode.RewardType,
		&referralCode.RewardValuePence,
		&referralCode.MaxUses,
		&referralCode.TimesUsed,
		&referralCode.IsActive,
		&referralCode.CreatedAt,
		&referralCode.UpdatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrReferralNotFound
	}

	if err != nil {
		return nil, err
	}

	return &referralCode, nil
}

func (r *SQLRepository) ListCodesByUserID(ctx context.Context, userID uint) ([]ReferralCode, error) {
	query := `
		SELECT
			id, 
			user_id,
			code,
			reward_type,
			reward_value_pence,
			max_uses,
			times_used,
			is_active,
			created_at,
			updated_at
		FROM referral_codes
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var codes []ReferralCode

	for rows.Next() {
		var code ReferralCode

		err := rows.Scan(
			&code.ID,
			&code.UserID,
			&code.Code,
			&code.RewardType,
			&code.RewardValuePence,
			&code.MaxUses,
			&code.TimesUsed,
			&code.IsActive,
			&code.CreatedAt,
			&code.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		codes = append(codes, code)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return codes, nil
}

func (r *SQLRepository) CreateRedemption(ctx context.Context, redemption *ReferralRedemption) error {
	query := `
		INSERT INTO referral_redemptions (
			referral_code_id,
			referrer_id,
			referred_user_id,
			status,
			reward_applied,
			created_at,
			updated_at
		)
	VALUES ($1, $2, $3, $4, $5, $6, $6)
	RETURNING id, created_at, updated_at
	`

	now := time.Now()

	err := r.db.QueryRowContext(
		ctx,
		query,
		redemption.ReferralCodeID,
		redemption.ReferrerID,
		redemption.ReferredUserID,
		redemption.Status,
		redemption.RewardApplied,
		now,
	).Scan(
		&redemption.ID,
		&redemption.CreatedAt,
		&redemption.UpdatedAt,
	)

	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "duplicate key") {
			return ErrAlreadyRedeemed
		}

		return err
	}

	return nil
}

func (r *SQLRepository) IncrementCodeUse(ctx context.Context, referralCodeID uint) error {
	query := `
		UPDATE referral_codes
		SET
			times_used = times_used + 1,
			updated_at = $1
		WHERE id = $2	
	`

	_, err := r.db.ExecContext(ctx, query, time.Now(), referralCodeID)
	return err
}

func (r *SQLRepository) ListRedemptionsByReferrerID(ctx context.Context, referrerID uint) ([]ReferralRedemption, error) {
	query := `
		SELECT
			id,
			referral_code_id,
			referrer_id,
			referred_user_id,
			status,
			reward_applied,
			created_at,
			updated_at
		FROM referral_redemptions
		WHERE referrer_id = $1
		ORDER BY created_at DESC	  
	`

	rows, err := r.db.QueryContext(ctx, query, referrerID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var redemptions []ReferralRedemption

	for rows.Next() {
		var redemption ReferralRedemption

		err := rows.Scan(
			&redemption.ID,
			&redemption.ReferralCodeID,
			&redemption.ReferrerID,
			&redemption.ReferredUserID,
			&redemption.Status,
			&redemption.RewardApplied,
			&redemption.CreatedAt,
			&redemption.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		redemptions = append(redemptions, redemption)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return redemptions, nil
}
