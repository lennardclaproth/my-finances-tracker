package date

import "time"

func StartOfDayUTC(t time.Time) time.Time {
	t = t.UTC()
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

func EndOfDayUTC(t time.Time) time.Time {
	d := StartOfDayUTC(t)
	return d.AddDate(0, 0, 1).Add(-time.Nanosecond)
}

func SameDayUTC(a, b time.Time) bool {
	aa := StartOfDayUTC(a)
	bb := StartOfDayUTC(b)
	return aa.Equal(bb)
}

func LatestBusinessDate(now time.Time, loc *time.Location) time.Time {
	n := now.In(loc)
	// Target is yesterday (one-day lag).
	target := n.AddDate(0, 0, -1)
	// If target falls on weekend, roll back to Friday.
	switch target.Weekday() {
	case time.Saturday:
		target = target.AddDate(0, 0, -1)
	case time.Sunday:
		target = target.AddDate(0, 0, -2)
	}
	// Normalize to midnight local time (date-only semantics).
	y, m, d := target.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, loc)
}

func DateOnly(t time.Time, loc *time.Location) time.Time {
	tt := t.In(loc)
	y, m, d := tt.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, loc)
}
