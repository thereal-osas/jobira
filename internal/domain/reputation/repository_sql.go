package reputation

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
	return &SQLRepository{db: db}
}

func (r *SQLRepository) EnsureCleaner(ctx context.Context, cleanerID uint) error {
	query := `
		INSERT INTO cleaner_reputation (
			cleaner_id, 
			created_at,
			updated_at
		)
		VALUES ($1, $2, $2)
		ON CONFLICT (cleaner_id)
		DO NOTHING
	`

	_, err := r.db.ExecContext(ctx, query, cleanerID, time.Now())
	return err
}
func (r *SQLRepository) Refresh(ctx context.Context, cleanerID uint) error {
	if err := r.EnsureCleaner(ctx, cleanerID); err != nil {
		return err
	}

	query := `
		UPDATE cleaner_reputation
		SET
			average_rating = COALESCE((
				SELECT ROUND(AVG(rv.rating)::numeric, 2)
				FROM reviews rv
				WHERE rv.cleaner_id = $1
			), 0),

			total_reviews = (
				SELECT COUNT(*)
				FROM reviews rv
				WHERE rv.cleaner_id = $1
			),

			completed_jobs = (
				SELECT COUNT(*)
				FROM bookings b
				WHERE b.cleaner_id = $1
				AND b.status IN ('completed', 'closed')
			),

			repeat_clients = (
				SELECT COUNT(*)
				FROM (
					SELECT b.client_id
					FROM bookings b
					WHERE b.cleaner_id = $1
					AND b.status IN ('completed', 'closed')
					GROUP BY b.client_id
					HAVING COUNT(*) > 1
				) AS repeated_clients
			),

			would_hire_again_count = (
				SELECT COUNT(*)
				FROM bookings b
				WHERE b.cleaner_id = $1
				AND b.status = 'closed'
				AND b.client_would_hire_again = true
			),

			recommendation_percentage = (
				SELECT
					CASE
						WHEN COUNT(*) = 0 THEN 0
						ELSE ROUND(
							COUNT(*) FILTER (
								WHERE b.client_would_hire_again = true
							)::numeric
							/
							COUNT(*)::numeric
							* 100
						)::int
					END
				FROM bookings b
				WHERE b.cleaner_id = $1
				AND b.status = 'closed'
				AND b.client_would_hire_again IS NOT NULL
			),

			updated_at = $2
		WHERE cleaner_id = $1
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		cleanerID,
		time.Now(),
	)

	return err
}

func (r *SQLRepository) GetByCleanerID(ctx context.Context, cleanerID uint) (*CleanerReputation, error) {
	query := `
		SELECT
			cleaner_id,
			average_rating,
			total_reviews,
			completed_jobs,
			repeat_clients,
			would_hire_again_count,
			recommendation_percentage,
			badge,
			created_at,
			updated_at
		FROM cleaner_reputation
		WHERE cleaner_id = $1
		LIMIT 1	
	`

	var reputation CleanerReputation

	err := r.db.QueryRowContext(ctx, query, cleanerID).Scan(
		&reputation.CleanerID,
		&reputation.AverageRating,
		&reputation.TotalReviews,
		&reputation.CompletedJobs,
		&reputation.RepeatClients,
		&reputation.WouldHireAgainCount,
		&reputation.RecommendationPercentage,
		&reputation.Badge,
		&reputation.CreatedAt,
		&reputation.UpdatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrReputationNotFound
	}

	if err != nil {
		return nil, err
	}

	return &reputation, nil
}

func (r *SQLRepository) UpdateBadge(ctx context.Context, cleanerID uint, badge string) error {
	query := `
		UPDATE cleaner_reputation
		SET 
			badge = $1, 
			updated_at = $2
		WHERE cleaner_id = $3	
	`

	_, err := r.db.ExecContext(ctx, query, badge, time.Now(), cleanerID)

	return err
}
