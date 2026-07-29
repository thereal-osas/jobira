package profiles

import (
	"context"
	"sort"
	"strings"

	
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) Create(ctx context.Context, userID uint, req CreateProfileRequest) (*CleanerProfile, error) {
	req.Bio = strings.TrimSpace(req.Bio)
	req.Location = strings.TrimSpace(req.Location)
	req.ServicesOffered = strings.TrimSpace(req.ServicesOffered)

	if userID == 0 {
		return nil, ErrInvalidInput
	}

	if req.Bio == "" || req.Location == "" || req.ServicesOffered == "" {
		return nil, ErrInvalidInput
	}

	profile := &CleanerProfile{
		UserID:             userID,
		Bio:                req.Bio,
		Country: 			req.Country,
		City: 				req.City,
		Region: 			req.Region,
		PostcodeArea: 		req.PostcodeArea,	
		Location:           req.Location,
		YearsExperience:    req.YearsExperience,
		HourlyRate:         req.HourlyRate,
		ServicesOffered:    req.ServicesOffered,
		IsVerified:         false,
		VerificationStatus: "pending",
	}

	if err := s.repo.Create(ctx, profile); err != nil {
		return nil, err
	} 

	return profile, nil
}

func (s *Service) GetMine(ctx context.Context, userID uint) (*CleanerProfile, error) {
	if userID == 0 {
		return nil, ErrInvalidInput
	}

	return s.repo.GetByUserID(ctx, userID)
}

func (s *Service) Update(ctx context.Context, userID uint, req UpdateProfileRequest) (*CleanerProfile, error) {
	req.Bio = strings.TrimSpace(req.Bio)
	 req.Bio = strings.TrimSpace(req.Bio)
	 req.Location = strings.TrimSpace(req.Location)
	 req.Country = strings.TrimSpace(req.Country)
	 req.City = strings.TrimSpace(req.City)
	 req.Region = strings.TrimSpace(req.Region)
	 req.PostcodeArea = strings.TrimSpace(req.PostcodeArea)
	 req.AvailabilityStatus = strings.TrimSpace(req.AvailabilityStatus)
	 req.ServicesOffered = strings.TrimSpace(req.ServicesOffered)

	 if req.Country == "" {
		req.Country = "UK"
	 }

	 if req.City == "" {
		req.City = "London"
	 }

	 if req.Region == "" {
		req.Region = "unknown" 
	 }

	 if req.PostcodeArea == "" {
		req.PostcodeArea = "unknown"
	 }

	 if req.AvailabilityStatus == "" {
		req.AvailabilityStatus = "available_immediately" 
	 }

	 if req.TravelRadiusMiles <= 0 {
		req.TravelRadiusMiles = 10
	 }
	 

	 if !isAllowedAvailabilityStatus(req.AvailabilityStatus) {
		return nil, ErrInvalidAvailabilityStatus 
	 }

	 foundProfile, err := s.repo.GetByUserID(ctx, userID)
	 if err != nil {
		return nil, err
	 }

	 foundProfile.Bio = req.Bio
	 foundProfile.Location = req.Location
	 foundProfile.Country = req.Country
	 foundProfile.City = req.City
	 foundProfile.Region = req.Region
	 foundProfile.PostcodeArea = req.PostcodeArea
	 foundProfile.AvailabilityStatus = req.AvailabilityStatus
	 foundProfile.TravelRadiusMiles = req.TravelRadiusMiles
	 foundProfile.ReliabilityScore = calculateReliabilityScore(foundProfile)
	 foundProfile.Badge = calculateCleanerBadge(foundProfile)
	 foundProfile.YearsExperience = req.YearsExperience
	 foundProfile.HourlyRate = req.HourlyRate
	 foundProfile.ServicesOffered = req.ServicesOffered

	 if err := s.repo.Update(ctx, foundProfile); err != nil {
		return nil, err
	 }

	 return s.repo.GetByUserID(ctx, userID) 
}

func isAllowedAvailabilityStatus(status string) bool {
	if status == "available_immediately" {
		return true
	}

	if status == "weekends_only" {
		return true
	}

	if status == "evening_only" {
		return true
	}

	if status == "weekdays_only" {
		return true
	}

	if status == "part-time" {
		return true
	}

	if status == "full_time" {
		return true
	}

	if status == "not_available" {
		return true
	}

	return false
}

func (s *Service) UpdateVerificationStatus(ctx context.Context, userID uint, status string) (*CleanerProfile, error) {
	status = strings.TrimSpace(status)

	if !isAllowedVerificationStatus(status) {
		return nil, ErrInvalidVerificationState
	}

	verified := false

	if status == "verified" {
		verified = true
	}

	if err := s.repo.UpdateVerificationStatus(ctx, userID, status, verified); err != nil {
		return nil, err
	}

	return s.repo.GetByUserID(ctx, userID)
}

func isAllowedVerificationStatus(status string) bool {
	if status == "pending" {
		return true
	}

	if status == "verified" {
		return true
	}

	if status == "rejected" {
		return true
	}

	return false
}

func (s *Service) Search(ctx context.Context, req SearchProfilesRequest) ([]SearchProfilesResult, error) {
	req.Country =  strings.TrimSpace(req.Country)
	req.City = strings.TrimSpace(req.City)
	req.Region = strings.TrimSpace(req.Region)
	req.PostcodeArea = strings.TrimSpace(req.PostcodeArea)
	req.AvailabilityStatus = strings.TrimSpace(req.AvailabilityStatus)
	req.ServicesOffered = strings.TrimSpace(req.ServicesOffered)

	profiles, err := s.repo.Search(ctx, req)
	if err != nil {
		return nil, err
	}

	results := make([]SearchProfilesResult, 0, len(profiles))

	for _, profile := range profiles {
		score := calculateMatchScore(profile, req)

		results = append(results, SearchProfilesResult{
			CleanerProfile: profile,
			MatchScore: score,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].MatchScore > results[j].MatchScore
	})

	return results, nil
}

func calculateMatchScore(profile CleanerProfile, req SearchProfilesRequest) int {
	score := 0 

	if req.PostcodeArea != "" && strings.EqualFold(profile.PostcodeArea, req.PostcodeArea) {
		score += 30
	}

	if req.Region != "" && strings.EqualFold(profile.Region, req.Region) {
		score += 20
	}

	if req.City != "" && strings.EqualFold(profile.City, req.City) {
		score += 15
	}

	if req.AvailabilityStatus != "" && strings.EqualFold(profile.AvailabilityStatus, req.AvailabilityStatus) {
		score += 15
	}

	if req.ServicesOffered != "" && strings.Contains(
		strings.ToLower(profile.ServicesOffered),
		strings.ToLower(req.ServicesOffered),
	) {
		score += 15 
	}

	if profile.IsVerified {
		score += 15 
	}

	if profile.YearsExperience >= 3 {
		score += 5
	}

	if req.MaxHourlyRate > 0 && profile.HourlyRate <= req.MaxHourlyRate {
		score += 5
	}

	return score
}

func calculateReliabilityScore(profile *CleanerProfile) int {
	score := 0 

	if profile.IsVerified {
		score += 25
	}

	if profile.JobsCompleted >= 1 {
		score += 20 
	}

	if profile.JobsCompleted >= 10 {
		score += 20 
	}

	if profile.JobsCancelled == 0 {
		score += 20
	}

	if profile.ResponseRate >= 80 {
		score += 20 
	}

	if score > 100 {
		score = 100
	}

	return score
}

func calculateCleanerBadge(profile *CleanerProfile) string {
	if profile.JobsCompleted >= 100 && profile.ReliabilityScore >= 90 {
		return "Elite Cleaner"
	}

	if profile.JobsCompleted >= 25 && profile.ReliabilityScore >= 80 {
		return "Trusted Cleaner"
	}

	if profile.ResponseRate >= 90 {
		return "Fast Responder"
	}

	if strings.Contains(strings.ToLower(profile.ServicesOffered), "airbnb") {
		return "Airbnb Specialist"
	}

	if profile.IsVerified {
		return "Verified Cleaner"
	}

	return "New Cleaner"
}

func  (s *Service) IncrementJobsCompleted(ctx context.Context, userID uint) error {
	if userID == 0 {
		return ErrInvalidInput
	}

	if err := s.repo.IncrementJobsCompleted(ctx, userID); err != nil {
		return err
	}

	return s.repo.RecalculateReputation(ctx, userID)
}

func (s *Service) IncrementJobsCancelled(ctx context.Context, userID uint) error {
	if userID == 0 {
		return ErrInvalidInput
	}

	if err := s.repo.IncrementJobsCancelled(ctx, userID); err != nil {
		return err
	}

	return s.repo.RecalculateReputation(ctx, userID)
}

func (s *Service) GetCompletedJobs(ctx context.Context, userID uint) ([]JobHistoryItem, error) {
	if userID == 0 {
		return nil, ErrInvalidInput
	}

	return s.repo.GetCompletedJobs(ctx, userID)
}

func (s *Service) GetCancelledJobs(ctx context.Context, userID uint) ([]JobHistoryItem, error) {
	if userID == 0 {
		return nil, ErrInvalidInput
	}

	return s.repo.GetCancelledJobs(ctx, userID)
}

func (s *Service) GetFullHistory(ctx context.Context, userID uint) ([]JobHistoryItem, error) {
	if userID == 0 {
		return nil, ErrInvalidInput
	}

	return s.repo.GetFullHistory(ctx, userID)
}