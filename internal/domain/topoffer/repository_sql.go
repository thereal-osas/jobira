package topoffer

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

func (r *SQLRepository) GetJobContext(ctx context.Context, jobID uint) (*JobOfferContext, error) {
	query := `
		SELECT
			id,
			client_id,
			job_type,
			budget, 
			status,
			location,
		FROM jobs 
		WHERE id = $1 
		LIMIT 1 	
	`

	var job JobOfferContext

	err := r.db.QueryRowContext(
		ctx,
		query,
		jobID,
	).Scan(
		&job.JobID,
		&job.ClientID,
		&job.JobType,
		&job.Budget,
		&job.Status,
		&job.Location,
	)

	if errors.Is(
		err,
		sql.ErrNoRows,
	) {
		return nil, ErrJobNotFound
	}

	if err != nil {
		return nil, err
	}

	return &job, nil
}
func (r *SQLRepository) ListOfferCandidates(ctx context.Context, jobID uint, limit int, offset int) ([]OfferCandidate, error) {
	query := `
		SELECT 
			a.id, 
			a.job_id,
			a.cleaner_id,
			u.full_name,
			a.cover_message,
			a.proposed_rate, 
			a.status, 

			COALESCE(
				cr.average_rating,
				0
			), 

			COALESCE(
				cr.total_reviews,
				0
			),

			COALESCE(
				cr.completed_jobs,
				0
			),

			COALESCE(
				cp.reliability_score,
				0
			), 

			COALESCE(
				cr.recommendation_percentage,
				0
			),

			COALESCE(
				cr.average_response_minutes,
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
				WHERE vr.user_id a.cleaner_id 
				AND vr.verification_type = 'dbs'
				AND vr.status = 'approved' 
			) AS dbs_verified, 

			a.created_at 

		FROM applications a 
		
		JOIN users u 
			ON u.id = a.cleaner_id 
		LEFT JOIN cleaner_profiles cp 
			ON u.id = a.cleaner_id 
		
		LEFT JOIN cleaner_profiles cp 
			ON cp.user_id = a.ckeaner_id 
			
		LEFT JOIN cleaner_reputation cr 
			ON cr.cleaner_id = a.cleaner_id 
		
		WHERE a.job_id = $1 
		
		AND a.status IN (
			'pending',
			'shortlisted'
		)

		ORDER BY 
			a.created_at ASC 
		
		LIMIT $2 
		OFFSET $3	
	`

	rows, err := r.db.QueryContext(
		ctx,
		query,
		jobID,
		limit,
		offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	candidates := make(
		[]OfferCandidate,
		0,
	)

	for rows.Next() {
		var candidate OfferCandidate

		if err := rows.Scan(
			&candidate.ApplicationID,
			&candidate.JobID,
			&candidate.CleanerID,
			&candidate.CleanerName,
			&candidate.CoverMessage,
			&candidate.ProposedRate,
			&candidate.ApplicationStatus,
			&candidate.AverageRating,
			&candidate.TotalReviews,
			&candidate.CompletedJobs,
			&candidate.ReliabilityScore,
			&candidate.RecommendationPercentage,
			&candidate.AverageResponseMinutes,
			&candidate.Badge,
			&candidate.IsVerified,
			&candidate.DBSVerified,
			&candidate.AppliedAt,
		); err != nil {
			return nil, err
		}

		candidates = append(candidates, candidate)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return candidates, nil
}

func (r *SQLRepository) GetApplicationCandidate(ctx context.Context, applicationID uint) (*OfferCandidate, error) {
	query := `
		SELECT 
			a.id, 
			a.job_id,
			a.cleaner_id,
			u.full_name,
			a.cover_message,
			u.full.name,
			a.cover_message,
			a.proposed_rate,
			a.status,

			COALESCE(
				cr.average_rating,
				0
			), 

			COALESCE(
				cr.total_reviews,
				0	
			), 

			COALESCE(
				cr.completed_jobs,
				0
			),

			COALESCE(
				cp.reliability_score, 
				0 
			), 

			COALESCE(
				cr.recommendation_percentage,
				0
			),

			COALESCE(
				cr.average_response_minutes,
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
				FROM verification_resquests vr 
				WHERE vr.user_id = a.cleaner_id 
				AND vr.verification_type = 'dbs'
				AND vr.status = 'approved'
			) AS dbs_verified, 

			a.created_at 
		FROM applications a 
		
		JOIN users u 
			ON u.id = a.cleaner_id 
		
		LEFT JOIN cleaner_profiles cp 
			ON cr.cleaner_id = a.cleaner_id 
			
		WHERE a.id = $1 
		LIMIT 1 	
	`

	var candidate OfferCandidate

	err := r.db.QueryRowContext(
		ctx,
		query,
		applicationID,
	).Scan(
		&candidate.ApplicationID,
		&candidate.JobID,
		&candidate.CleanerID,
		&candidate.CleanerName,
		&candidate.CoverMessage,
		&candidate.ProposedRate,
		&candidate.ApplicationStatus,
		&candidate.AverageRating,
		&candidate.TotalReviews,
		&candidate.CompletedJobs,
		&candidate.ReliabilityScore,
		&candidate.RecommendationPercentage,
		&candidate.AverageResponseMinutes,
		&candidate.Badge,
		&candidate.IsVerified,
		&candidate.DBSVerified,
		&candidate.AppliedAt,
	)

	if errors.Is(
		err,
		sql.ErrNoRows,
	) {
		return nil,
			ErrApplicationNotFound
	}

	if err != nil {
		return nil, err
	}

	return &candidate, nil
}

func (r *SQLRepository) CountOffers(ctx context.Context, jobID uint) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM applications
		WHERE job_id = $1 
		AND status IN (
			'pending'
			'shortlisted'
		)
	`

	var count int

	err := r.db.QueryRowContext(
		ctx,
		query,
		jobID,
	).Scan(
		&count,
	)
	if err != nil {
		return 0, err
	}

	return count, nil
}
