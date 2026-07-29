package analytics

import (
	"context"
	"database/sql"
)

type SQLRepository struct {
	db *sql.DB
}

func NewSQLRepository(db *sql.DB) *SQLRepository {
	return &SQLRepository{db: db}
}

func (r *SQLRepository) TotalUsers(ctx context.Context) (int, error) {
	var count int

	query := `
		SELECT COUNT (*)
		FROM users
	`

	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	return count, err
}

func (r *SQLRepository) UsersByRole(ctx context.Context, role string) (int, error) {
	var count int

	query := `
		SELECT COUNT (*)
		FROM users
		WHERE role = $1
	`

	err := r.db.QueryRowContext(ctx, query, role).Scan(&count)
	return count, err
}

func (r *SQLRepository) TotalCompanies(ctx context.Context) (int, error) {
	var count int

	query := `
		SELECT COUNT(*)
		FROM companies
	`

	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	return count, err
}

func (r *SQLRepository) ActiveJobs(ctx context.Context) (int, error) {
	var count int

	query := `
		SELECT COUNT (*)
		FROM jobs
		WHERE status = 'open'
	`

	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	return count, err
}

func (r *SQLRepository) CompletedBookings(ctx context.Context) (int, error) {
	var count int

	query := `
		SELECT COUNT (*)
		FROM bookings
		WHERE status = 'completed'
	`

	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	return count, err

}

func (r *SQLRepository) PendingVerifications(ctx context.Context) (int, error) {
	var count int

	query := `
		SELECT COUNT (*)
		FROM verification_requests
		WHERE status = 'pending'
	`

	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	return count, err
}

func (r *SQLRepository) OpenReports(ctx context.Context) (int, error) {
	var count int

	query := `
		SELECT COUNT (*)
		FROM reports
		WHERE status = 'open'
	`

	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	return count, err
}

func (r *SQLRepository) BookingsToday(ctx context.Context) (int, error) {
	var count int

	query := `
		SELECT COUNT (*)
		FROM bookings
		WHERE DATE(created_at) = CURRENT_DATE
	`

	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	return count, err
}
