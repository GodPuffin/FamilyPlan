package payments

import (
	"strings"
	"testing"
)

func TestParseForMonth(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value string
		want  string
	}{
		{name: "empty value", value: "", want: ""},
		{name: "invalid month", value: "2026-13", want: ""},
		{name: "wrong format", value: "2026/04", want: ""},
		{name: "valid month", value: "2026-04", want: "2026-04-01"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseForMonth(tt.value); got != tt.want {
				t.Fatalf("parseForMonth(%q) = %q, want %q", tt.value, got, tt.want)
			}
		})
	}
}

func TestNormalizeNotes(t *testing.T) {
	t.Parallel()

	got, err := normalizeNotes("  paid for March  ")
	if err != nil {
		t.Fatalf("normalizeNotes returned error: %v", err)
	}
	if got != "paid for March" {
		t.Fatalf("normalizeNotes() = %q, want %q", got, "paid for March")
	}

	tooLong := strings.Repeat("界", maxPaymentNotesLength+1)
	_, err = normalizeNotes(tooLong)
	if err == nil {
		t.Fatal("expected normalizeNotes to reject oversized notes")
	}
}
