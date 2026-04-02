package planutil

import "testing"

func TestBuildEqualsFilterUsesBoundParams(t *testing.T) {
	t.Parallel()

	filter, err := BuildEqualsFilter(
		FilterTerm{Field: "username", Value: "o'hara"},
		FilterTerm{Field: "verified", Value: true},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got, want := filter.Expression, "username = {:p0} && verified = {:p1}"; got != want {
		t.Fatalf("expression = %q, want %q", got, want)
	}

	if got := filter.Params["p0"]; got != "o'hara" {
		t.Fatalf("param p0 = %#v, want quoted username", got)
	}

	if got := filter.Params["p1"]; got != true {
		t.Fatalf("param p1 = %#v, want true", got)
	}
}

func TestBuildContainsFilterRejectsInvalidField(t *testing.T) {
	t.Parallel()

	if _, err := BuildContainsFilter("owner.id", "abc123"); err == nil {
		t.Fatalf("expected invalid field error")
	}
}
