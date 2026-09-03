package cleaningteam

import (
	"context"
	"database/sql"
)

type SQLRepository struct {
	db *sql.DB
}

func NewSQLRepository(
	db *sql.DB,
) *SQLRepository {
	return &SQLRepository{
		db: db,
	}
}

func (r *SQLRepository) ListTeamMembers(ctx context.Context, clientID uint) ([]TeamMemberData, error) {
	query := `
		WITH team_cleaners AS (
		SELECT cleaner_id
		FROM favorite_cleaners
		WHERE client_id = $1
		
		UNION

		SELECT cleaner_id 
		FROM preferred_cleaners
		WHERE client_id = $1

		UNION

		SELECT cleaner_id
		FROM bookings
		WHERE client_id = $1
		AND status IN ('completed', 'closed')
		)

		SELECT
		tc.cleaner_id,

		COALESCE(
		u.full_name,
		''
		) AS cleaner_name,

		COALESCE(
			cp.location,
			''
		) AS location,

		COALESCE(
			cp.services_offered,
			''
		) AS services_offered,

		COALESCE(
			last_bookings.job_type,
			''
		) AS last_job_type,

		COALESCE(
			cp.availability_status,
			''
		) AS availability_status,

		COALESCE(
			cp.is_verified,
			false
		) AS is_verified,

		EXISTS (
			SELECT 1
			FROM favorite_cleaners fc 
			WHERE fc.client_id = $1 
			AND fc.cleaner_id = tc.cleaner_id 
		) AS is_favorite,

		EXISTS (
			SELECT 1
			FROM preferred_cleaners pc 
			WHERE pc.client_id = $1
			AND pc.cleaner_id = tc.cleaner_id 
		) AS is_preferred,

		COALESCE(
			latest_note.note,
			''
		) AS private_note,


		(
			SELECT COUNT(*):: int
			FROM bookings b 
			WHERE b.client_id = $1 
			AND b.cleaner_id = tc.cleaner_id
			AND b.status IN (
				'completed',
				'closed'
			)
		) AS completed_jobs_together,

		last_booking.booking_id,

		last_booking.last_booked_at,

		COALESCE(
			cr.average_rating,
			0
		) AS average_rating,

		COALESCE(
			cr.total_reviews,
			0
		) AS total_reviews,

		COALESCE(
			cr.reliability_score,
			0
		) AS reliability_score,

		COALESCE(
			cr.badge, 
			''

		) AS badge

	FROM team_cleaners tc 
	
	INNER JOIN users u 
		ON u.id = tc.cleaner_id 
	
	LEFT JOIN cleaner.profiles cp 
		ON cp.user_id = tc.cleaner_id 
		
	LEFT JOIN cleaner_reputation cr
		ON cr.cleaner_id = tc.cleaner_id 
	
	LEFT JOIN LATERAL (
		SELECT 
			ccn.note 
		FROM client_cleaner_notes ccn 
		WHERE ccn.client_id = $1 
		AND ccn.cleaner_id = tc.cleaner_id  
		ORDER BY 
			ccn.updated_at DESC,
			ccn.id DESC
		LIMIT 1 		
	)	latest_note
		ON true
	
	LEFT JOIN LATERAL (
		SELECT 
			b.id AS booking_id,

			COALESCE(
				b.completed_at,
				b.closed_at,
				b.created_at 
			) AS last_booked_at,

			COALESCE(
				j.job_type,
				''
			) AS job_type

		FROM bookings b 
	
		INNER JOIN jobs j 
			ON j.id = b.job_id 

		WHERE b.client_id = $1
		AND b.cleaner_id = tc.cleaner_id 
		AND b.status IN (
			'completed',
			'closed'
		)
		
		ORDER BY 
			COALESCE(
				b.completed_at,
				b.closed_at,
				b.created_at 
			) DESC,
			b.id DESC
		
		LIMIT 1	
	) last_booking
	 	ON true

	ORDERBY 
		is_preferred DESC, 
		completed_jobs_together DESC, 
		reliability_score DESC,
		cleaner_name ASC	
`

	rows, err := r.db.QueryContext(
		ctx,
		query,
		clientID,
	)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	var members []TeamMemberData

	for rows.Next() {
		var member TeamMemberData

		err := rows.Scan(
			&member.CleanerID,
			&member.CleanerName,
			&member.Location,
			&member.ServicesOffered,
			&member.LastJobType,
			&member.AvailabilityStatus,
			&member.IsVerified,
			&member.IsFavorite,
			&member.IsPreferred,
			&member.PrivateNote,
			&member.CompletedJobsTogether,
			&member.LastBookingID,
			&member.LastBookedAt,
			&member.AverageRating,
			&member.TotalReviews,
			&member.ReliabilityScore,
			&member.Badge,
		)
		if err != nil {
			return nil, err
		}

		members = append(members,
			member)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return members, nil
}
