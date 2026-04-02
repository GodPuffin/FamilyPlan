package billing

import (
	"testing"
	"time"
)

func TestMonthIterationCrossesYearBoundary(t *testing.T) {
	t.Parallel()

	start := time.Date(2025, time.December, 1, 0, 0, 0, 0, time.UTC)
	next := start.AddDate(0, 1, 0)

	if next.Year() != 2026 || next.Month() != time.January {
		t.Fatalf("next month = %s, want January 2026", next.Format(time.RFC3339))
	}
}

func TestMemberShareCentsDistributesRemainderDeterministically(t *testing.T) {
	t.Parallel()

	memberIDs := []string{"member-b", "member-a"}

	if got := memberShareCents(1001, memberIDs, "member-a"); got != 501 {
		t.Fatalf("member-a share = %d, want 501", got)
	}

	if got := memberShareCents(1001, memberIDs, "member-b"); got != 500 {
		t.Fatalf("member-b share = %d, want 500", got)
	}
}
