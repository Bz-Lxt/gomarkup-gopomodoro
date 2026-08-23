package timeutil

import "testing"

func TestDaysBetweenMinOne(t *testing.T) {
	a := Date(2026, 8, 23, 10, 0, 0)
	if DaysBetween(a, a) != 1 {
		t.Fatal("same day must be 1")
	}
	if DaysBetween(a, Date(2026, 8, 25, 1, 0, 0)) != 2 {
		t.Fatal("two day gap")
	}
}

func TestFormatBeijing(t *testing.T) {
	d := Date(2026, 8, 23, 16, 7, 0)
	if FormatDateTime(d) != "2026-08-23 16:07:00" {
		t.Fatal(FormatDateTime(d))
	}
}
