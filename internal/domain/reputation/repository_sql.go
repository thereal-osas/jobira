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

func (r *SQLRepository) Refresh(
	ctx context.Context,
	cleanerID uint,
) error {
	if err := r.EnsureCleaner(
		ctx,
		cleanerID,
	); err != nil {
		return err
	}

	query := `
		WITH review_stats AS (
			SELECT
				COALESCE(
					ROUND(AVG(rv.rating)::numeric, 2),
					0
				) AS average_rating,
				COUNT(*)::int AS total_reviews
			FROM reviews rv
			WHERE rv.cleaner_id = $1
		),

		booking_stats AS (
			SELECT
				COUNT(*) FILTER (
					WHERE b.status IN (
						'completed',
						'closed',
						'cancelled'
					)
				)::int AS total_bookings,

				COUNT(*) FILTER (
					WHERE b.status IN (
						'completed',
						'closed'
					)
				)::int AS completed_jobs,

				COUNT(*) FILTER (
					WHERE b.status = 'cancelled'
					AND b.cancelled_by = b.cleaner_id
				)::int AS cleaner_cancellations,

				COUNT(*) FILTER (
					WHERE b.status = 'closed'
					AND b.client_would_hire_again = true
				)::int AS would_hire_again_count,

				CASE
					WHEN COUNT(*) FILTER (
						WHERE b.status = 'closed'
						AND b.client_would_hire_again IS NOT NULL
					) = 0
					THEN 0

					ELSE ROUND(
						COUNT(*) FILTER (
							WHERE b.status = 'closed'
							AND b.client_would_hire_again = true
						)::numeric
						/
						COUNT(*) FILTER (
							WHERE b.status = 'closed'
							AND b.client_would_hire_again IS NOT NULL
						)::numeric
						* 100
					)::int
				END AS recommendation_percentage

			FROM bookings b
			WHERE b.cleaner_id = $1
		),

		repeat_stats AS (
			SELECT
				COUNT(*)::int AS repeat_clients
			FROM (
				SELECT b.client_id
				FROM bookings b
				WHERE b.cleaner_id = $1
				AND b.status IN (
					'completed',
					'closed'
				)
				GROUP BY b.client_id
				HAVING COUNT(*) > 1
			) repeated_clients
		),

		ordered_messages AS (
			SELECT
				m.sender_id,
				m.created_at,
				c.client_id,
				c.cleaner_id,

				LEAD(m.sender_id) OVER (
					PARTITION BY m.conversation_id
					ORDER BY
						m.created_at ASC,
						m.id ASC
				) AS next_sender_id,

				LEAD(m.created_at) OVER (
					PARTITION BY m.conversation_id
					ORDER BY
						m.created_at ASC,
						m.id ASC
				) AS next_created_at

			FROM messages m

			INNER JOIN conversations c
				ON c.id = m.conversation_id

			WHERE c.cleaner_id = $1
			AND m.message_type <> 'system'
		),

		response_stats AS (
			SELECT
				COUNT(*) FILTER (
					WHERE sender_id = client_id
					AND (
						next_sender_id IS NULL
						OR next_sender_id <> client_id
					)
				)::int AS eligible_response_messages,

				COUNT(*) FILTER (
					WHERE sender_id = client_id
					AND next_sender_id = cleaner_id
				)::int AS responded_messages,

				COALESCE(
					ROUND(
						AVG(
							EXTRACT(
								EPOCH FROM (
									next_created_at -
									created_at
								)
							) / 60.0
						) FILTER (
							WHERE sender_id = client_id
							AND next_sender_id = cleaner_id
						)
					)::int,
					0
				) AS average_response_minutes

			FROM ordered_messages
		)

		UPDATE cleaner_reputation
		SET
			average_rating =
				review_stats.average_rating,

			total_reviews =
				review_stats.total_reviews,

			completed_jobs =
				booking_stats.completed_jobs,

			repeat_clients =
				repeat_stats.repeat_clients,

			would_hire_again_count =
				booking_stats.would_hire_again_count,

			recommendation_percentage =
				booking_stats.recommendation_percentage,

			total_bookings =
				booking_stats.total_bookings,

			cleaner_cancellations =
				booking_stats.cleaner_cancellations,

			eligible_response_messages =
				response_stats.eligible_response_messages,

			responded_messages =
				response_stats.responded_messages,

			average_response_minutes =
				response_stats.average_response_minutes,

			updated_at = $2

		FROM
			review_stats,
			booking_stats,
			repeat_stats,
			response_stats

		WHERE cleaner_reputation.cleaner_id = $1
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		cleanerID,
		time.Now(),
	)

	return err
}

func (r *SQLRepository) GetByCleanerID(
	ctx context.Context,
	cleanerID uint,
) (*CleanerReputation, error) {
	query := `
		SELECT
			cleaner_id,
			average_rating,
			total_reviews,
			completed_jobs,
			repeat_clients,
			would_hire_again_count,
			recommendation_percentage,

			total_bookings,
			cleaner_cancellations,

			eligible_response_messages,
			responded_messages,
			average_response_minutes,

			badge,
			created_at,
			updated_at

		FROM cleaner_reputation
		WHERE cleaner_id = $1
		LIMIT 1
	`

	var reputation CleanerReputation

	err := r.db.QueryRowContext(
		ctx,
		query,
		cleanerID,
	).Scan(
		&reputation.CleanerID,
		&reputation.AverageRating,
		&reputation.TotalReviews,
		&reputation.CompletedJobs,
		&reputation.RepeatClients,
		&reputation.WouldHireAgainCount,
		&reputation.RecommendationPercentage,

		&reputation.TotalBookings,
		&reputation.CleanerCancellations,

		&reputation.EligibleResponseMessages,
		&reputation.RespondedMessages,
		&reputation.AverageResponseMinutes,

		&reputation.Badge,
		&reputation.CreatedAt,
		&reputation.UpdatedAt,
	)

	if errors.Is(
		err,
		sql.ErrNoRows,
	) {
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
