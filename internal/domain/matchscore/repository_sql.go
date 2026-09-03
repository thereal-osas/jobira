package matchscore

import (
	"context"
	"database/sql"
	"errors"
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

func (r *SQLRepository) JobBelongsToClient(ctx context.Context, jobID uint, clientID uint) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT 1 
			FROM jobs 
			WHERE id = $1
			AND client_id = $2
		)
	`

	var exists bool

	err := r.db.QueryRowContext(
		ctx,
		query,
		jobID,
		clientID,
	).Scan(
		&exists,
	)

	if err != nil {
		return false, err
	}

	return exists, nil
}

func (r *SQLRepository) ListCandidates(ctx context.Context, jobID uint, clientID uint) ([]CandidateData, error) {
	query := `
		SELECT 
			a.id,
			a.job_id,
			j.client_id, 
			a.cleaner_id,

			COALESCE(u.full_name, ''), 

			COALESCE(j.job_type, ''), 
			COALESCE(j.location, ''), 

			COALESCE(cp.location, ''), 
			COALESCE(cp.postcode_area, ''), 
			COALESCE(cp.services_offered, ''), 
			COALESCE(cp.availability_status, ''), 

			COALESCE(cp.is_verified, false), 

			COALESCE(cr.reliability_score, 0), 
			COALESCE(cr.response_rate, 0), 
			COALESCE(cr.average_response_minutes, 0), 

			COALESCE(cr.average_rating, 0), 
			COALESCE(cr.total_reviews, 0),


			COALESCE(cr.badge, '')
		
		FROM applications a 
		
		INNER JOIN jobs j 
			ON j.id = a.job_id 

		INNER JOIN users u 
			ON u.id = a.cleaner_id  
		
		LEFT JOIN cleaner_profiles cp 
			ON u.id = a.cleaner_id 
		
		LEFT JOIN cleaner_reputation cr 
			ON cr.cleaner_id = a.cleaner_id 
		
		WHERE a.job_id = $1 
		AND j.client_id = $2 
		AND a.status IN (
			'pending'
			'shortlisted'
		)	

		ORDER BY a.created_at ASC 
	`

	rows, err := r.db.QueryContext(
		ctx,
		query,
		jobID,
		clientID,
	)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var candidates []CandidateData

	for rows.Next() {
		var candidate CandidateData

		err := rows.Scan(
			&candidate.ApplicationID,
			&candidate.JobID,
			&candidate.ClientID,
			&candidate.CleanerID,

			&candidate.CleanerName,

			&candidate.JobType,
			&candidate.JobLocation,

			&candidate.CleanerLocation,
			&candidate.CleanerPostcodeArea,
			&candidate.ServicesOffered,
			&candidate.AvailabilityStatus,

			&candidate.IsVerified,

			&candidate.ReliabilityScore,
			&candidate.ResponseRate,
			&candidate.AverageResponseMinutes,

			&candidate.AverageRating,
			&candidate.TotalReviews,

			&candidate.Badge,
		)
		if err != nil {
			return nil, err
		}

		candidates = append(candidates, candidate)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return candidates, nil
}

func scanBoolResult(
	row *sql.Row,
) (bool, error) {
	var value bool

	err := row.Scan(
		&value,
	)

	if errors.Is(
		err, sql.ErrNoRows,
	) {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	return value, nil
}
