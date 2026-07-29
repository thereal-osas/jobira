package recentviews

import "time"

type RecentlyViewedCleaner struct {
	ID uint	`json:"id"`
	ClientID 	uint	`json:"client_id"`
	CleanerID 	uint	`json:"cleaner_id"`
	ViewedAt	time.Time	`json:"viewed_at"`
}