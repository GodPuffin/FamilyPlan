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

	got := buildMemberPaymentsPagination(3, true)

	if got["CurrentPage"] != 3 {
		t.Fatalf("CurrentPage = %v, want 3", got["CurrentPage"])
	}
	if got["HasPrev"] != true || got["PrevPage"] != 2 {
		t.Fatalf("unexpected prev page data: %+v", got)
	}
	if got["HasNext"] != true || got["NextPage"] != 4 {
		t.Fatalf("unexpected next page data: %+v", got)
	}
}
