package plans

import "testing"

func TestMemberPaymentsPageDefaultsToOne(t *testing.T) {
	t.Parallel()

	tests := []string{"", "0", "-2", "abc"}
	for _, raw := range tests {
		if got := memberPaymentsPage(raw); got != 1 {
			t.Fatalf("memberPaymentsPage(%q) = %d, want 1", raw, got)
		}
	}
}

func TestMemberPaymentsPageParsesPositiveValue(t *testing.T) {
	t.Parallel()

	if got := memberPaymentsPage("3"); got != 3 {
		t.Fatalf("memberPaymentsPage(3) = %d, want 3", got)
	}
}

func TestBuildMemberPaymentsPagination(t *testing.T) {
	t.Parallel()

	t.Run("first page", func(t *testing.T) {
		got := buildMemberPaymentsPagination(1, false)

		if got.CurrentPage != 1 {
			t.Fatalf("CurrentPage = %d, want 1", got.CurrentPage)
		}
		if got.HasPrev {
			t.Fatalf("HasPrev = %t, want false", got.HasPrev)
		}
		if got.PrevPage != 1 {
			t.Fatalf("PrevPage = %d, want 1", got.PrevPage)
		}
		if got.HasNext {
			t.Fatalf("HasNext = %t, want false", got.HasNext)
		}
		if got.NextPage != 2 {
			t.Fatalf("NextPage = %d, want 2", got.NextPage)
		}
	})

	t.Run("middle page", func(t *testing.T) {
		got := buildMemberPaymentsPagination(3, true)

		if got.CurrentPage != 3 {
			t.Fatalf("CurrentPage = %d, want 3", got.CurrentPage)
		}
		if !got.HasPrev || got.PrevPage != 2 {
			t.Fatalf("unexpected prev page data: %+v", got)
		}
		if !got.HasNext || got.NextPage != 4 {
			t.Fatalf("unexpected next page data: %+v", got)
		}
	})
}
