package cleanerprogression

import (
	"context"
	"database/sql"
	"errors"

	reputationdomain "github.com/rodrigueghenda/jobira/internal/domain/reputation"
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

func (r *SQLRepository) GetCleanerReputation(ctx context.Context, cleanerID uint) (*reputationdomain.CleanerReputation, error) {
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

	var reputation reputationdomain.CleanerReputation

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
		return nil,
			reputationdomain.ErrReputationNotFound
	}

	if err != nil {
		return nil, err
	}

	return &reputation, nil

}
