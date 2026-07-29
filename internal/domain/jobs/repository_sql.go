package jobs

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"time"
)

type SQLRepository struct {
	db *sql.DB
}

func NewSQLRepository(db *sql.DB) *SQLRepository {
	return &SQLRepository{db: db}
}

func (r *SQLRepository) Create(ctx context.Context, job *Job) error {
	query := `
	INSERT INTO jobs (
		client_id,
		title,
		description,
		location,
		job_type,
		listing_type,
		budget,
		status,
		created_at,
		updated_at
	)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	RETURNING id
	`

	now := time.Now()

	job.Status = "open"
	job.CreatedAt = now
	job.UpdatedAt = now

	return r.db.QueryRowContext(
		ctx,
		query,
		job.ClientID,
		job.Title,
		job.Description,
		job.Location,
		job.JobType,
		job.ListingType,
		job.Budget,
		job.Status,
		job.CreatedAt,
		job.UpdatedAt,
	).Scan(&job.ID)
}

func (r *SQLRepository) GetByID(ctx context.Context, id uint) (*Job, error) {
	query := `
	SELECT
		id,
		client_id,
		title,
		description,
		location,
		job_type,
		listing_type,
		budget,
		status,
		created_at,
		updated_at
	FROM jobs
	WHERE id = $1
	LIMIT 1
`

	var job Job

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&job.ID,
		&job.ClientID,
		&job.Title,
		&job.Description,
		&job.Location,
		&job.JobType,
		&job.ListingType,
		&job.Budget,
		&job.Status,
		&job.CreatedAt,
		&job.UpdatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrJobNotFound
	}

	if err != nil {
		return nil, err
	}

	return &job, nil
}

func (r *SQLRepository) List(ctx context.Context) ([]Job, error) {
	query := `
		SELECT
			id,
			client_id,
			title,
			description, 
			location,
			job_type,
			listing_type,
			budget,
			status, 
			created_at,
			updated_at
		FROM jobs
		ORDER BY created_at DESC	
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var jobs []Job

	for rows.Next() {
		var job Job

		err := rows.Scan(
			&job.ID,
			&job.ClientID,
			&job.Title,
			&job.Description,
			&job.Location,
			&job.JobType,
			&job.ListingType,
			&job.Budget,
			&job.Status,
			&job.CreatedAt,
			&job.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		jobs = append(jobs, job)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return jobs, nil
}

func (r *SQLRepository) ListByClientID(ctx context.Context, clientID uint) ([]Job, error) {
	query := `
		SELECT
			id,
			client_id,
			title,
			description,
			location,
			job_type,
			listing_type,
			budget,
			status,
			created_at,
			updated_at
		FROM jobs
		WHERE client_id = $1
		ORDER BY created_at DESC	
	`

	rows, err := r.db.QueryContext(ctx, query, clientID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var jobs []Job

	for rows.Next() {
		var job Job

		err := rows.Scan(
			&job.ID,
			&job.ClientID,
			&job.Title,
			&job.Description,
			&job.Location,
			&job.JobType,
			&job.ListingType,
			&job.Budget,
			&job.Status,
			&job.CreatedAt,
			&job.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		jobs = append(jobs, job)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return jobs, nil
}

func (r *SQLRepository) Update(ctx context.Context, job *Job) error {
	query := `
		UPDATE jobs
		SET
			title = $1,
			description = $2,
			location = $3,
			job_type = $4,
			listing_type = $5,
			budget = $6,
			status = $7,
			updated_at = $8
		WHERE id = $9 	
	`

	job.UpdatedAt = time.Now()

	result, err := r.db.ExecContext(
		ctx,
		query,
		job.Title,
		job.Description,
		job.Location,
		job.JobType,
		job.ListingType,
		job.Budget,
		job.Status,
		job.UpdatedAt,
		job.ID,
	)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrJobNotFound
	}

	return nil
}

func (r *SQLRepository) Delete(ctx context.Context, id uint) error {
	query := `
		DELETE FROM jobs
		WHERE id = $1		
	`

	results, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := results.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrJobNotFound
	}

	return nil
}

func (r *SQLRepository) Search(ctx context.Context, req SearchJobRequest) ([]Job, error) {
	query := `
		SELECT
			id,
			client_id,
			title,
			description,
			location, 
			job_type, 
			listing_type,
			budget,
			status, 
			created_at,
			updated_at 
		FROM jobs 
		WHERE 1=1
	`

	args := []interface{}{}
	argPosition := 1

	if req.Location != "" {
		query += " AND LOWER(location) LIKE LOWER($" + strconv.Itoa(argPosition) + ")"
		args = append(args, "%"+req.Location+"%")
		argPosition++
	}

	if req.JobType != "" {
		query += " AND LOWER(job_type) LIKE LOWER($" + strconv.Itoa(argPosition) + ")"
		args = append(args, req.JobType)
		argPosition++
	}

	if req.ListingType != "" {
		query += " AND LOWER(listing_type) = LOWER($" + strconv.Itoa(argPosition) + ")"
		args = append(args, req.ListingType)
		argPosition++
	}

	if req.MinBudget > 0 {
		query += " AND budget >= $" + strconv.Itoa(argPosition)
		args = append(args, req.MinBudget)
		argPosition++
	}

	query += " ORDER BY created_at DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var jobs []Job

	for rows.Next() {
		var job Job

		err := rows.Scan(
			&job.ID,
			&job.ClientID,
			&job.Title,
			&job.Description,
			&job.Location,
			&job.JobType,
			&job.ListingType,
			&job.Budget,
			&job.Status,
			&job.CreatedAt,
			&job.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		jobs = append(jobs, job)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return jobs, nil
}
