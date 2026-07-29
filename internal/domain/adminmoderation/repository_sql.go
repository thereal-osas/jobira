package adminmoderation

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

func (r *SQLRepository) ListReports(ctx context.Context) ([]CleanerReportAdminView, error) {
	query := `
		SELECT 
			id,
			client_id,
			cleaner_id,
			reason,
			details,
			status,
			COALESCE(admin_notes, ''),
			resolved_by,
			resolved_at,
			created_at,
			updated_at 
		FROM cleaner_reports 
		ORDER BY created_at DESC 	
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reports []CleanerReportAdminView

	for rows.Next() {
		var report CleanerReportAdminView
		var resolvedAt sql.NullTime

		err := rows.Scan(
			&report.ID,
			&report.ClientID,
			&report.CleanerID,
			&report.Reason,
			&report.Details,
			&report.Status,
			&report.AdminNotes,
			&report.ResolvedBy,
			&resolvedAt,
			&report.CreatedAt,
			&report.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
	
	if resolvedAt.Valid {
		report.ResolvedAt = &resolvedAt.Time
	}
		reports = append(reports, report)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return reports, nil
}
	
func (r *SQLRepository) ListOpenReports(ctx context.Context) ([]CleanerReportAdminView, error) {
		query := `
		SELECT 
			id,
			client_id,
			cleaner_id,
			reason,
			details,
			status,
			COALESCE(admin_notes, ''),
			resolved_by,
			resolved_at,
			created_at,
			updated_at 
		FROM cleaner_reports
		WHERE status = 'open' 
		ORDER BY created_at DESC 	
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reports []CleanerReportAdminView
	

	for rows.Next() {
		var report CleanerReportAdminView
		var resolvedAt sql.NullTime

		err := rows.Scan(
			&report.ID,
			&report.ClientID,
			&report.CleanerID,
			&report.Reason,
			&report.Details,
			&report.Status,
			&report.AdminNotes,
			&report.ResolvedBy,
			&resolvedAt,
			&report.CreatedAt,
			&report.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if resolvedAt.Valid {
			report.ResolvedAt = &resolvedAt.Time
		}

		reports = append(reports, report)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return reports, nil
}
	
func (r *SQLRepository) UpdateReportStatus(ctx context.Context, reportID uint, adminID uint, status string, adminNote string) error {
	 query := `
	 	UPDATE cleaner_reports
		SET 
			status = $1, 
			admin_notes = $2, 
			resolved_by = $3, 
			resolved_at = $4, 
			updated_at = $5 
		WHERE id = $6 	
	 `

	 now := time.Now()

	 result, err  := r.db.ExecContext(
		ctx, 
		query,
		status, 
		adminNote,
		adminID, 
		now,
		now,
		reportID,
	 )
	 if err != nil {
		return err
	 }

	 rowsAffected, err := result.RowsAffected()
	 if err != nil {
		return err
	 }

	 if rowsAffected == 0 {
		return ErrReportNotFound
	 }

	 return nil
}
	
func (r *SQLRepository) ListBlockCleaners(ctx context.Context) ([]BlockedCleanerAdminView, error) {
	query := `
		SELECT 
			id, 
			client_id,
			cleaner_id, 
			reason, 
			created_at
		FROM blocked_cleaners 
		ORDER BY created_at DESC 	
	`	
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var blocks []BlockedCleanerAdminView

	for rows.Next() {
		var block BlockedCleanerAdminView

		err :=  rows.Scan(
			&block.ID,
			&block.ClientID,
			&block.CleanerID,
			&block.Reason,
			&block.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		blocks = append(blocks, block)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return blocks, nil
}