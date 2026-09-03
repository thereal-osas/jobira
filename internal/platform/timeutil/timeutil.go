package timeutil

import "time"

func NowTC() time.Time {
	return time.Now().UTC()
}

func FormatISO(t time.Time) string {
	return t.UTC().Format(time.RFC3339)
}
