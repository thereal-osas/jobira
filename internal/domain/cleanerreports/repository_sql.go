package cleanerreports

import (
	"context"
	"database/sql"
	"time"
)

type SQLRepository struct {
	db *sql.DB
}

func NewSQLRepository(db *sql.DB) *SQLRepository {
	return &SQLRepository{db: db}
}

func (r *SQLRepository) Create(ctx context.Context, report *CleanerReport) error {
	query := `
		INSERT INTO cleaner_reports (
			client_id,
			cleaner_id,
			reason,
			details,
			status,
			created_at,
			updated_at 
		)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			RETURNING id, status, created_at, updated_at 
	`

	now := time.Now()

	report.Status = "open"

	return r.db.QueryRowContext(
		ctx, 
		query, 
		report.ClientID,
		report.CleanerID,
		report.Reason,
		report.Details,
		report.Status,
		now,
		now,
	).Scan(
		&report.ID,
		&report.Status,
		&report.CreatedAt,
		&report.UpdatedAt,
	)
}

func (r *SQLRepository) ListByClientID(ctx context.Context, clientID uint) ([]CleanerReport, error) {
	query := `
		SELECT
			id,
			client_id,
			cleaner_id,
			reason,
			details,
			status,
			created_at,
			updated_at 
		FROM cleaner_reports 
		WHERE client_id = $1 
		ORDER BY created_at DESC 		
	`

	rows, err := r.db.QueryContext(ctx, query, clientID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var reports []CleanerReport

	for rows.Next() {
		var report CleanerReport

		err := rows.Scan(
			&report.ID,
			&report.ClientID,
			&report.CleanerID,
			&report.Reason,
			&report.Details,
			&report.Status,
			&report.CreatedAt,
			&report.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		reports = append(reports, report)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return reports, nil
}