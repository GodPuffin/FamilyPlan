package money

import "testing"

func TestParseCents(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    int64
		wantErr bool
	}{
		{name: "whole amount", input: "12", want: 1200},
		{name: "single decimal", input: "12.3", want: 1230},
		{name: "two decimals", input: "12.34", want: 1234},
		{name: "trim spaces", input: " 9.99 ", want: 999},
		{name: "too many decimals", input: "1.999", wantErr: true},
		{name: "empty", input: "", wantErr: true},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got, err := ParseCents(test.input)
			if test.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got != test.want {
				t.Fatalf("ParseCents(%q) = %d, want %d", test.input, got, test.want)
			}
		})
	}
}

func TestSplitEvenly(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		total int64
		parts int
		want  int64
	}{
		{name: "exact split", total: 999, parts: 3, want: 333},
		{name: "rounded split", total: 1000, parts: 3, want: 333},
		{name: "single part", total: 1000, parts: 1, want: 1000},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if got := SplitEvenly(test.total, test.parts); got != test.want {
				t.Fatalf("SplitEvenly(%d, %d) = %d, want %d", test.total, test.parts, got, test.want)
			}
		})
	}
}
