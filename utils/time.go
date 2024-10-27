package utils

import "time"

func EndOfWeek(now time.Time) time.Time {
	endOfWeek := now.AddDate(0, 0, (7-int(now.Weekday()))%7)

	return time.Date(endOfWeek.Year(), endOfWeek.Month(), endOfWeek.Day(), 23, 0, 0, 0, now.Location())
}
