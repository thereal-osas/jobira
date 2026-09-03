package jobpulse

import (
	"context"
	"database/sql"
	"errors"
	"time"
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

func (r *SQLRepository) GetSnapshot(ctx context.Context, jobID uint) (*JobPulseSnapshot, error) {
	query := `
		SELECT 
			J.id,
			j.status,
			j.created_at,

			COUNT(
				DISTINCT a.id
			) AS application_count,

			COUNT(
				DISTINCT CASE
					WHEN a.created_at >= $2
					THEN a.id
				END	
			) AS recent_applicatoin_count,

			COUNT(
				DISTINCT b.id 
			) AS booking_count 

		FROM jobs j 
		
		LEFT JOIN applications a 
			ON a.job_id = j.id 

		LEFT JOIN bookings b 
			ON b.job_id = j.id 
			AND b.status NOT IN (
				'cancelled'
			)	

		WHERE j.id = $1 
		
		GROUP BY 
			j.id,
			j.status,
			j.created_at 
	`

	recentSince :=
		time.Now().UTC().
			Add(-24 * time.Hour)

	var snapshot JobPulseSnapshot

	err := r.db.QueryRowContext(
		ctx,
		query,
		jobID,
		recentSince,
	).Scan(
		&snapshot.JobID,
		&snapshot.JobStatus,
		&snapshot.CreatedAt,
		&snapshot.ApplicationCount,
		&snapshot.RecentApplicationCount,
		&snapshot.BookingCount,
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

	return &snapshot, nil
}
