package timeutil

import "time"

// Beijing is the sole civil timezone for persistence and display (GMT+8).
var Beijing = time.FixedZone("CST", 8*3600)

func Now() time.Time {
	return time.Now().In(Beijing)
}

func Date(year int, month time.Month, day, hour, min, sec int) time.Time {
	return time.Date(year, month, day, hour, min, sec, 0, Beijing)
}

func StartOfDay(t time.Time) time.Time {
	t = t.In(Beijing)
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, Beijing)
}

func EndOfDay(t time.Time) time.Time {
	return StartOfDay(t).Add(24*time.Hour - time.Nanosecond)
}

func ParseDate(s string) (time.Time, error) {
	return time.ParseInLocation("2006-01-02", s, Beijing)
}

func FormatDate(t time.Time) string {
	return t.In(Beijing).Format("2006-01-02")
}

func FormatDateTime(t time.Time) string {
	return t.In(Beijing).Format("2006-01-02 15:04:05")
}

func DaysBetween(from, to time.Time) int {
	a := StartOfDay(from)
	b := StartOfDay(to)
	d := int(b.Sub(a).Hours() / 24)
	if d < 1 {
		return 1
	}
	return d
}

func ElapsedDays(start, now time.Time) int {
	d := int(StartOfDay(now).Sub(StartOfDay(start)).Hours() / 24)
	if d < 0 {
		return 0
	}
	return d
}

func ToBeijing(t time.Time) time.Time {
	return t.In(Beijing)
}
