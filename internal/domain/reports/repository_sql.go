package reports

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

func (r *SQLRepository) Create(ctx context.Context, report *Report) error {
	query := `
		INSERT INTO reports (
			reporter_id,
			reported_user_id,
			job_id,
			booking_id,
			report_type,
			reason,
			details,
			status,
			created_at,
			updated_at 
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $9)
		RETURNING id, created_at, updated_at	
	`

	now := time.Now()

	err := r.db.QueryRowContext(
		ctx,
		query,
		report.ReporterID,
		report.ReportedUserID,
		report.JobID,
		report.BookingID,
		report.ReportType,
		report.Reason,
		report.Details,
		report.Status,
		now,
	).Scan(
		&report.ID,
		&report.CreatedAt,
		&report.UpdatedAt,
	)

	if err != nil {
		return err
	}

	return nil
}

func (r *SQLRepository) GetByID(ctx context.Context, id uint) (*Report, error) {
	query := `
		SELECT
			id,
			reporter_id,
			reported_user_id,
			job_id,
			booking_id,
			report_type,
			reason,
			COALESCE(details, ''),
			status,
			COALESCE(admin_notes, ''),
			reviewed_by,
			reviewed_at,
			created_at,
			updated_at
		FROM reports
		WHERE id = $1
		LIMIT 1	
	`

	var report Report
	var ReportedUserID sql.NullInt64
	var jobID sql.NullInt64
	var bookingID sql.NullInt64
	var reviewedBy sql.NullInt64
	var reviewedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&report.ID,
		&report.ReporterID,
		&ReportedUserID,
		&jobID,
		&bookingID,
		&report.ReportType,
		&report.Reason,
		&report.Details,
		&report.Status,
		&report.AdminNotes,
		&reviewedBy,
		&reviewedAt,
		&report.CreatedAt,
		&report.UpdatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrReportNotFound
	}

	if err != nil {
		return nil, err
	}

	if ReportedUserID.Valid {
		value := uint(ReportedUserID.Int64)
		report.ReportedUserID = &value
	}

	if jobID.Valid {
		value := uint(jobID.Int64)
		report.JobID = &value
	}

	if bookingID.Valid {
		value := uint(bookingID.Int64)
		report.BookingID = &value
	}

	if reviewedBy.Valid {
		value := uint(reviewedBy.Int64)
		report.ReviewedBy = &value
	}

	if reviewedAt.Valid {
		report.ReviewedAt = &reviewedAt.Time
	}

	return &report, nil
}

func (r *SQLRepository) ListByReporterID(ctx context.Context, reporterID uint) ([]Report, error) {
	query := `
		SELECT 
			id, 
			reporter_id,
			reported_user_id,
			job_id,
			booking_id,
			report_type, 
			reason,
			COALESCE(details, ''),
			status,
			COALESCE(admin_notes, ''),
			reviewed_by,
			reviewed_at,
			created_at,
			updated_at
		FROM reports
		WHERE reporter_id = $1 
		ORDER BY created_at DESC 	
	`

	rows, err := r.db.QueryContext(ctx, query, reporterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reports []Report

	for rows.Next() {
		var report Report
		var ReportedUserID sql.NullInt64
		var jobID sql.NullInt64
		var bookingID sql.NullInt64
		var reviewedBy sql.NullInt64
		var reviewedAt sql.NullTime

		err := rows.Scan(
			&report.ID,
			&report.ReporterID,
			&ReportedUserID,
			&jobID,
			&bookingID,
			&report.ReportType,
			&report.Reason,
			&report.Details,
			&report.Status,
			&report.AdminNotes,
			&reviewedBy,
			&reviewedAt,
			&report.CreatedAt,
			&report.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if ReportedUserID.Valid {
			value := uint(ReportedUserID.Int64)
			report.ReportedUserID = &value
		}

		if jobID.Valid {
			value := uint(jobID.Int64)
			report.JobID = &value
		}

		if bookingID.Valid {
			value := uint(bookingID.Int64)
			report.BookingID = &value
		}

		if reviewedBy.Valid {
			value := uint(reviewedBy.Int64)
			report.ReviewedBy = &value
		}

		if reviewedAt.Valid {
			report.ReviewedAt = &reviewedAt.Time
		}

		reports = append(reports, report)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return reports, nil
}

func (r *SQLRepository) ListAll(ctx context.Context) ([]Report, error) {
	query := `
		SELECT 
			id, 
			reporter_id,
			reported_user_id,
			job_id,
			booking_id,
			report_type, 
			reason,
			COALESCE(details, ''),
			status,
			COALESCE(admin_notes, ''),
			reviewed_by,
			reviewed_at,
			created_at,
			updated_at
		FROM reports 
		ORDER BY created_at DESC 	
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reports []Report

	for rows.Next() {
		var report Report
		var ReportedUserID sql.NullInt64
		var jobID sql.NullInt64
		var bookingID sql.NullInt64
		var reviewedBy sql.NullInt64
		var reviewedAt sql.NullTime

		err := rows.Scan(
			&report.ID,
			&report.ReporterID,
			&ReportedUserID,
			&jobID,
			&bookingID,
			&report.ReportType,
			&report.Reason,
			&report.Details,
			&report.Status,
			&report.AdminNotes,
			&reviewedBy,
			&reviewedAt,
			&report.CreatedAt,
			&report.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if ReportedUserID.Valid {
			value := uint(ReportedUserID.Int64)
			report.ReportedUserID = &value
		}

		if jobID.Valid {
			value := uint(jobID.Int64)
			report.JobID = &value
		}

		if bookingID.Valid {
			value := uint(bookingID.Int64)
			report.BookingID = &value
		}

		if reviewedBy.Valid {
			value := uint(reviewedBy.Int64)
			report.ReviewedBy = &value
		}

		if reviewedAt.Valid {
			report.ReviewedAt = &reviewedAt.Time
		}

		reports = append(reports, report)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return reports, nil

}
func (r *SQLRepository) ListOpen(ctx context.Context) ([]Report, error) {
	query := `
		SELECT 
			id, 
			reporter_id,
			reported_user_id,
			job_id,
			booking_id,
			report_type, 
			reason,
			COALESCE(details, ''),
			status,
			COALESCE(admin_notes, ''),
			reviewed_by,
			reviewed_at,
			created_at,
			updated_at
		FROM reports 
		WHERE status = 'open'
		ORDER BY created_at DESC 	
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reports []Report

	for rows.Next() {
		var report Report
		var ReportedUserID sql.NullInt64
		var jobID sql.NullInt64
		var bookingID sql.NullInt64
		var reviewedBy sql.NullInt64 
		var reviewedAt sql.NullTime

		err := rows.Scan(
			&report.ID,
			&report.ReporterID,
			&ReportedUserID,
			&jobID,
			&bookingID,
			&report.ReportType,
			&report.Reason,
			&report.Details,
			&report.Status,
			&report.AdminNotes,
			&reviewedBy,
			&reviewedAt,
			&report.CreatedAt,
			&report.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		if ReportedUserID.Valid {
			value := uint(ReportedUserID.Int64)
			report.ReportedUserID = &value
		}

		if jobID.Valid {
			value := uint(jobID.Int64)
			report.JobID = &value
		}

		if bookingID.Valid {
			value := uint(bookingID.Int64)
			report.BookingID = &value
		}

		if reviewedBy.Valid {
			value := uint(reviewedBy.Int64)
			report.ReviewedBy = &value
		}

		if reviewedAt.Valid {
			report.ReviewedAt = &reviewedAt.Time
		}

		reports = append(reports, report)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return reports, nil

}
func (r *SQLRepository) Review(ctx context.Context, reportID uint, status string, adminNotes string, adminID uint) error {
	query := `
		UPDATE reports
		SET	
			status = $1,
			admin_notes = $2,
			reviewed_by = $3,
			reviewed_at = $4,
			updated_at = $4
		WHERE id = $5	
	`

	now := time.Now()

	result, err := r.db.ExecContext(
		ctx,
		query,
		status,
		adminNotes,
		adminID,
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
