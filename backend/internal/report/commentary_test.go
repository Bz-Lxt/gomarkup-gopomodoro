package report

import (
	"strings"
	"testing"

	"gopomodoro/internal/timeutil"
)

func TestCommentaryBranches(t *testing.T) {
	now := timeutil.Date(2026, 8, 23, 0, 0, 0)
	due := timeutil.Date(2026, 9, 30, 0, 0, 0)
	done := Triple(0, 2, due, now)
	if !strings.Contains(Commentary(done, 0, 0), "归零") {
		t.Fatal("done")
	}
	ok := Triple(4, 2, due, now)
	c := Commentary(ok, 0.35, 4)
	if !strings.Contains(c, "废弃率") {
		t.Fatal(c)
	}
	stuck := Triple(10, 0, due, now)
	if !strings.Contains(Commentary(stuck, 0, 10), "无法预测") {
		t.Fatal("stuck")
	}
}
