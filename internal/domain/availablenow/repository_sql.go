package availablenow

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/lib/pq"
)

type SQLRepository struct {
	db *sql.DB
}

func NewSQLRepository(db *sql.DB) *SQLRepository {
	return &SQLRepository{
		db: db,
	}
}

func (r *SQLRepository) Upsert(ctx context.Context, availability *CleanerAvailableNow) error {
	query := `
		INSERT INTO cleaner_available_now (
			cleaner_id,
			is_available,
			available_from,
			available_until,
			location,
			travel_radius_miles,
			job_types,
			created_at,
			updated_at
		)
	VALUES (
		$1, $2, $3, $4, $5, $6, $7, $8, $8
	)
		ON CONFLICT (cleaner_id)
		DO UPDATE SET
			is_available = EXCLUDED.is_available,
			available_from = EXCLUDED.available_from, 
			available_until = EXCLUDED.available_until,
			location = EXCLUDED.location, 
			travel_radius_miles = EXCLUDED.travel_radus_miles
			job_types = EXCLUDED.job_types, 
			updated_at = EXCLUDED.updated_at 
		RETURNING 
			id,
			created_at, 
			updated_at	
	`

	now := time.Now()

	return r.db.QueryRowContext(
		ctx,
		query,
		availability.CleanerID,
		availability.IsAvailable,
		availability.AvailableFrom,
		availability.AvailableUntil,
		availability.Location,
		availability.TravelRadiusMiles,
		pq.Array(availability.JobTypes),
		now,
	).Scan(
		&availability.ID,
		&availability.CreatedAt,
		&availability.UpdatedAt,
	)
}

func (r *SQLRepository) GetByCleanerID(ctx context.Context, cleanerID uint) (*CleanerAvailableNow, error) {
	query := `
		SELECT 
			id, 
			cleaner_id, 
			is_available,
			available_from, 
			available_until, 
			location, 
			travel_radius_miles,
			job_types,
			created_at, 
			updated_at
		FROM cleaner_available_now
		WHERE cleaner_id = $1
		LIMIT 1	
	`

	var availability CleanerAvailableNow

	err := r.db.QueryRowContext(
		ctx,
		query,
		cleanerID,
	).Scan(
		&availability.ID,
		&availability.CleanerID,
		&availability.IsAvailable,
		&availability.AvailableFrom,
		&availability.AvailableUntil,
		&availability.Location,
		&availability.TravelRadiusMiles,
		pq.Array(&availability.JobTypes),
		&availability.CreatedAt,
		&availability.UpdatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrAvailableNowNotFound
	}

	if err != nil {
		return nil, err
	}

	return &availability, nil
}

func (r *SQLRepository) Disable(ctx context.Context, cleanerID uint) error {
	query := `
		UPDATE cleaner_available_now
		SET 
			is_available = FALSE, 
			updated_at = $1 
		WHERE cleaner_id, = $2	
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		time.Now(),
		cleanerID,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrAvailableNowNotFound
	}

	return nil
}
func (r *SQLRepository) ListAvailableCleaners(
	ctx context.Context,
	search AvailableNowSearchRequest,
	now time.Time,
) ([]AvailableCleaner, error) {
	query := `
		SELECT
			can.cleaner_id,
			u.full_name,
			can.location,
			can.travel_radius_miles,
			can.available_until,
			can.job_types,

			COALESCE(
				cr.average_rating,
				0
			),

			COALESCE(
				cr.total_reviews,
				0
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

			COALESCE(
				cp.is_verified,
				FALSE
			),

			EXISTS (
				SELECT 1
				FROM verification_requests vr
				WHERE vr.user_id = can.cleaner_id
				AND vr.verification_type = 'dbs'
				AND vr.status = 'approved'
			) AS dbs_verified

			COALESCE(
				cr.average_response_minutes,
				0
			)

		FROM cleaner_available_now can

		JOIN users u
			ON u.id = can.cleaner_id

		LEFT JOIN cleaner_profiles cp
			ON cp.user_id = can.cleaner_id

		LEFT JOIN cleaner_reputation cr
			ON cr.cleaner_id = can.cleaner_id

		WHERE can.is_available = TRUE

		AND can.available_from <= $1

		AND can.available_until > $1

		AND (
			$2 = ''
			OR LOWER(can.location)
				LIKE LOWER(
					'%' || $2 || '%'
				)
		)

		AND (
			$3 = ''
			OR $3 = ANY(can.job_types)
		)

		AND (
			$4 = 0
			OR COALESCE(
				cr.average_rating,
				0
			) >= $4
		)

		AND (
			$5 = FALSE
			OR COALESCE(
				cp.is_verified,
				FALSE
			) = TRUE
		)

		AND (
			$6 = FALSE
			OR EXISTS (
				SELECT 1
				FROM verification_requests vr
				WHERE vr.user_id = can.cleaner_id
				AND vr.verification_type = 'dbs'
				AND vr.status = 'approved'
			)
		)

		ORDER BY
			COALESCE(
				cp.reliability_score,
				0
			) DESC,

			COALESCE(
				cr.average_rating,
				0
			) DESC,

			can.available_until ASC

		LIMIT $7
		OFFSET $8
	`

	rows, err := r.db.QueryContext(
		ctx,
		query,
		now,
		search.Location,
		search.JobType,
		search.MinimumRating,
		search.VerifiedOnly,
		search.DBSRequired,
		search.Limit,
		search.Offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cleaners := make(
		[]AvailableCleaner,
		0,
	)

	for rows.Next() {
		var cleaner AvailableCleaner

		if err := rows.Scan(
			&cleaner.CleanerID,
			&cleaner.FullName,
			&cleaner.Location,
			&cleaner.TravelRadiusMiles,
			&cleaner.AvailableUntil,
			pq.Array(
				&cleaner.JobTypes,
			),
			&cleaner.AverageRating,
			&cleaner.TotalReviews,
			&cleaner.ReliabilityScore,
			&cleaner.Badge,
			&cleaner.IsVerified,
			&cleaner.DBSVerified,
			&cleaner.UsuallyRespondsMinutes,
		); err != nil {
			return nil, err
		}

		cleaners = append(
			cleaners,
			cleaner,
		)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return cleaners, nil
}
