package timeutil

import (
	"testing"
	"time"
)

func TestWeekMondayBased(t *testing.T) {
	sun := Date(2026, 8, 23, 10, 0, 0) // Sunday
	mon := StartOfWeek(sun)
	if mon.Weekday() != 1 {
		t.Fatalf("weekday %v", mon.Weekday())
	}
	if FormatDate(mon) != "2026-08-17" {
		t.Fatal(FormatDate(mon))
	}
}

func TestClampAndEachDay(t *testing.T) {
	s := Date(2026, 8, 1, 0, 0, 0)
	e := Date(2026, 8, 3, 0, 0, 0)
	if FormatDate(ClampDate(Date(2026, 7, 1, 0, 0, 0), s, e)) != "2026-08-01" {
		t.Fatal("clamp low")
	}
	n := 0
	EachDay(s, e, func(time.Time) { n++ })
	if n != 3 {
		t.Fatal(n)
	}
}
