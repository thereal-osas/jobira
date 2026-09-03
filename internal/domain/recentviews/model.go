package recentviews

import "time"

type RecentlyViewedCleaner struct {
	ID        uint      `json:"id"`
	ClientID  uint      `json:"client_id"`
	CleanerID uint      `json:"cleaner_id"`
	ViewedAt  time.Time `json:"viewed_at"`
}

type ProfileViewAnalytics struct {
	CleanerID       uint    `json:"cleaner_id"`
	TotalViews      int     `json:"total_views"`
	UniqueViewers   int     `json:"unique_viewers"`
	ViewsToday      int     `json:"views_today"`
	ViewsThisWeek   int     `json:"views_this_week"`
	ViewsLastWeek   int     `json:"views_last_week"`
	WeeklyChange    float64 `json:"weekly_change_precent"`
	HasWeeklyGrowth bool    `json:"has_weekly_growth"`
}
