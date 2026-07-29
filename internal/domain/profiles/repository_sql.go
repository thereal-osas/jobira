package profiles

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"strings"
	"time"
)

type SQLRepository struct {
	db *sql.DB	
}

func NEWSQLRepository(db *sql.DB) *SQLRepository {
	return &SQLRepository{
		db: db,
	}
}

func (r *SQLRepository) Create(ctx context.Context, profile *CleanerProfile) error {
	query := `
		INSERT INTO cleaner_profiles (
			user_id,
			bio,
			country,
			city,
			region,
			postcode_area,
			availability_status,
			travel_radius_miles,
			location,
			years_experience,
			hourly_rate,
			services_offered,
			is_verified,
			verification_status,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		RETURNING id	
	`

	now := time.Now()

	profile.IsVerified = false
	profile.VerificationStatus = "pending"
	profile.CreatedAt = now
	profile.UpdatedAt = now

	err := r.db.QueryRowContext(
		ctx,
		query,
		profile.UserID,
		profile.Bio,
		profile.Country,
		profile.City,
		profile.Region,
		profile.PostcodeArea,
		profile.AvailabilityStatus,
		profile.TravelRadiusMiles,
		profile.Location,
		profile.YearsExperience,
		profile.HourlyRate,
		profile.ServicesOffered,
		profile.IsVerified,
		profile.VerificationStatus,
		profile.CreatedAt,
		profile.UpdatedAt,
	).Scan(&profile.ID)

	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "duplicate") {
			return ErrProfileAlreadyExists
		}
		
		return err
	}

	return nil
}

func (r *SQLRepository) GetByUserID(ctx context.Context, userID uint) (*CleanerProfile, error) {
	query := `
		SELECT 
			id,
			user_id,
			bio,
			location,
			country,
			city,
			region,
			postcode_area, 
			availability_status,
			travel_radius_miles,
			jobs_completed,
			jobs_cancelled,
			response_rate,
			reliability_score,
			Badge,
			years_experience,
			hourly_rate,
			services_offered,
			is_verified,
			verification_status,
			created_at,
			updated_at
		FROM cleaner_profiles
		WHERE user_id = $1
		LIMIT 1	
	`

	var profiles CleanerProfile

	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&profiles.ID,
		&profiles.UserID,
		&profiles.Bio,
		&profiles.Location,
		&profiles.Country,
		&profiles.City,
		&profiles.Region,
		&profiles.PostcodeArea,
		&profiles.AvailabilityStatus,
		&profiles.TravelRadiusMiles,
		&profiles.JobsCompleted,
		&profiles.JobsCancelled,
		&profiles.ResponseRate,
		&profiles.ReliabilityScore,
		&profiles.Badge,
		&profiles.YearsExperience,
		&profiles.HourlyRate,
		&profiles.ServicesOffered,
		&profiles.IsVerified,
		&profiles.VerificationStatus,
		&profiles.CreatedAt,
		&profiles.UpdatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrProfileNotFound
	}

	if err != nil {
		return nil, err
	}

	return &profiles, nil
}

func (r *SQLRepository) Update(ctx context.Context, profile *CleanerProfile) error {
	query := `
		UPDATE cleaner_profiles
		SET 
			bio = $1,
			location = $2,
			country = $3,
			city = $4,
			region = $5,
			postcode_area = $6,
			availability_status = $7,
			travel_radius_miles = $8, 
			years_experience = $9,
			hourly_rate = $10,
			services_offered = $11,
			reliability_score = $12,
			badge = $13,
			updated_at = $14
		WHERE user_id = $15    	
	`

	result, err := r.db.ExecContext(
		ctx, 
		query,
		profile.Bio,
		profile.Location,
		profile.Country,
		profile.City,
		profile.Region,
		profile.PostcodeArea,
		profile.AvailabilityStatus,
		profile.TravelRadiusMiles,
		profile.YearsExperience,
		profile.HourlyRate,
		profile.ServicesOffered,
		profile.ReliabilityScore,
		profile.Badge,
		time.Now(),
		profile.UserID,
	)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrProfileNotFound
	}

	return nil
}

func (r *SQLRepository) UpdateVerificationStatus(ctx context.Context, userID uint, status string, verified bool) error {
	query := `
		UPDATE cleaner_profiles
		SET 
			verification_status = $1,
			is_verified = $2,
			updated_at = $3
		WHERE user_id = $4 	
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		status,
		verified,
		time.Now(),
		userID,
	)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrProfileNotFound
	}

	return nil
}

func (r *SQLRepository) Search(ctx context.Context, req SearchProfilesRequest) ([]CleanerProfile, error) {
	query := `
		SELECT
			id, 
			user_id,
			bio,
			location,
			country,
			city,
			region,
			postcode_area,
			availability_status,
			travel_radius_miles,
			Jobs_Completed,
			Jobs_Cancelled,
			Response_Rate,
			Reliability_Score,
			Badge,
			years_experience,
			hourly_rate,
			services_offered,
			is_verified,
			verification_status,
			created_at,
			updated_at 
		FROM cleaner_profiles 
		WHERE 1=1	
	`
	
	args := []interface{}{}
	argPosition := 1 

	if req.ClientID > 0 {
	query += `
		AND  user_id NOT IN (
			SELECT cleaner_id
			FROM blocked_cleaners
			WHERE client_id = $` + strconv.Itoa(argPosition) + `
		)
	`
		args = append(args, req.ClientID)
		argPosition++
	}

	if req.Country != "" {
		query += " AND LOWER(country) = LOWER($" + strconv.Itoa(argPosition) + ")"
		args = append(args, req.Country)
		argPosition++
	}

	if req.City != "" {
		query += " AND LOWER(city) = LOWER($" + strconv.Itoa(argPosition) + ")"
		args = append(args, req.City)
		argPosition++
	}

	if req.Region != "" {
		query += " AND LOWER(region) = LOWER($" + strconv.Itoa(argPosition) + ")"
		args = append(args, req.Region)
		argPosition++
	}

	if req.PostcodeArea != "" {
		query += " AND LOWER(postcode_area) = LOWER($" + strconv.Itoa(argPosition) + ")"
		args = append(args, req.PostcodeArea)
		argPosition++
	}

	if req.AvailabilityStatus != "" {
		query += " AND LOWER(availability_status) = LOWER($" + strconv.Itoa(argPosition) + ")"
		args = append(args, req.AvailabilityStatus)
		argPosition++
	}

	if req.ServicesOffered != "" {
		query += " AND LOWER(services_offered) LIKE LOWER($" + strconv.Itoa(argPosition) + ")"
		args = append(args, "%"+req.ServicesOffered+"%")
		argPosition++
	} 

	if req.MinExperience > 0 {
		query += " AND years_experience >= $" + strconv.Itoa(argPosition)
		args = append(args, req.MinExperience)
		argPosition++
	}

	if req.MaxHourlyRate > 0 {
		query += " AND hourly_rate <= $" + strconv.Itoa(argPosition)
		args = append(args, req.MaxHourlyRate)
		argPosition++
	}

	if req.IsVerified != nil {
		query += " AND is_verified = $" + strconv.Itoa(argPosition)
		args = append(args, *req.IsVerified)
		argPosition++
	}

	query += " ORDER BY is_verified DESC, years_experience DESC, created_at DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var profiles []CleanerProfile

	for rows.Next() {
		var profile CleanerProfile

		err := rows.Scan(
			&profile.ID,
			&profile.UserID,
			&profile.Bio,
			&profile.Location,
			&profile.Country,
			&profile.City,
			&profile.Region,
			&profile.PostcodeArea,
			&profile.AvailabilityStatus,
			&profile.TravelRadiusMiles,
			&profile.JobsCompleted,
			&profile.JobsCancelled,
			&profile.ResponseRate,
			&profile.ReliabilityScore,
			&profile.Badge,
			&profile.YearsExperience,
			&profile.HourlyRate,
			&profile.ServicesOffered,
			&profile.IsVerified,
			&profile.VerificationStatus, 
			&profile.CreatedAt,
			&profile.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		profiles = append(profiles, profile)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return profiles, nil
}

func (r *SQLRepository) IncrementJobsCompleted(ctx context.Context, userID uint) error {
	query := `
		UPDATE cleaner_profiles
		SET 
			jobs_completed = jobs_completed + 1,
			updated_at = $1 
		WHERE user_id = $2	
	`

	_, err := r.db.ExecContext(ctx, query, time.Now(), userID)

	return err
}

func (r *SQLRepository) IncrementJobsCancelled(ctx context.Context, userID uint) error {
	query := `
		UPDATE cleaner_profiles
		SET 
			jobs_cancelled = jobs_cancelled + 1, 
			updated_at = $1 
		WHERE user_id = $2 	
	`

	_, err :=r.db.ExecContext(ctx, query, time.Now(), userID)

	return err
}

func (r *SQLRepository) RecalculateReputation(ctx context.Context, userID uint) error {
	profile, err := r.GetByUserID(ctx, userID)
	if err != nil {
		return err
	}

	profile.ReliabilityScore = calculateReliabilityScore(profile)
	profile.Badge = calculateCleanerBadge(profile)

	query := `
		UPDATE cleaner_profiles
		SET 
			reliability_score = $1, 
			badge = $2, 
			updated_at = $3 
		WHERE user_id = $4 	
	`

	_, err = r.db.ExecContext(
		ctx,
		query,
		profile.ReliabilityScore,
		profile.Badge,
		time.Now(),
		userID,
	)

	return err
}

func (r *SQLRepository)GetCompletedJobs(ctx context.Context, userID uint) ([]JobHistoryItem, error) {
	query := `
		SELECT 
			a.id,
			a.job_id,
			j.title,
			a.status,
			j.location,
			a.created_at
		FROM applications a
		INNER JOIN jobs j ON j.id = a.job_id 
		WHERE a.cleaner_id = $1 
		AND a.status = 'completed'
		ORDER BY a.created_at DESC 	
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	} 

	defer rows.Close()

	var history []JobHistoryItem

	for rows.Next() {
		var item JobHistoryItem

		err := rows.Scan(
			&item.ApplicationID,
			&item.JobID,
			&item.JobTitle,
			&item.Status,
			&item.Location,
			&item.CreatedAt,
		)

		if err != nil {
			return nil, err
		}

		history = append(history, item)
	}

	return history, nil
}
func (r *SQLRepository)	GetCancelledJobs(ctx context.Context, userID uint) ([]JobHistoryItem, error) {
	query := `
		SELECT 
			a.id,
			a.job_id,
			j.title,
			a.status,
			j.location,
			a.created_at
		FROM applications a 
		INNER JOIN jobs j ON j.id = a.job_id 
		WHERE a.cleaner_id = $1 
		AND a.status = 'cancelled' 
		ORDER BY a.created_at DESC 	
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var history []JobHistoryItem

	for rows.Next() {
		var item JobHistoryItem

		err := rows.Scan(
			&item.ApplicationID,
			&item.JobID,
			&item.JobTitle,
			&item.Status,
			&item.Location,
			&item.CreatedAt,
		)

		if err != nil {
			return nil, err
		}

		history = append(history, item)
	}

	return history, nil
}


func (r *SQLRepository)	GetFullHistory(ctx context.Context, userID uint) ([]JobHistoryItem, error) {
	query := `
		SELECT 
			a.id,
			a.job_id,
			j.title,
			a.status,
			j.location,
			a.created_at 
		FROM applications a 
		INNER JOIN jobs j ON j.id = a.job_id 
		WHERE a.cleaner_id = $1 
		ORDER BY a.created_at DESC 	
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var history []JobHistoryItem

	for rows.Next() {
		var item JobHistoryItem

		err := rows.Scan(
			&item.ApplicationID,
			&item.JobID,
			&item.JobTitle,
			&item.Status,
			&item.Location,
			&item.CreatedAt,
		)

		if err != nil {
			return nil, err
		}

		history = append(history, item)
	}

	return history, nil
}