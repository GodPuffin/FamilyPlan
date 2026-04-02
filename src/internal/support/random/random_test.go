package random

import (
	"regexp"
	"testing"

	"github.com/google/uuid"
)

func TestGenerateJoinCodeUsesExpectedCharset(t *testing.T) {
	t.Parallel()

	got, err := GenerateJoinCode(12)
	if err != nil {
		t.Fatalf("GenerateJoinCode returned error: %v", err)
	}

	if len(got) != 12 {
		t.Fatalf("len(code) = %d, want %d", len(got), 12)
	}

	if !regexp.MustCompile(`^[A-Z0-9]+$`).MatchString(got) {
		t.Fatalf("GenerateJoinCode() returned unexpected charset: %q", got)
	}
}

func TestGenerateTokenUsesExpectedCharset(t *testing.T) {
	t.Parallel()

	got, err := GenerateToken()
	if err != nil {
		t.Fatalf("GenerateToken returned error: %v", err)
	}

	if len(got) != 32 {
		t.Fatalf("len(token) = %d, want %d", len(got), 32)
	}

	if !regexp.MustCompile(`^[A-Za-z0-9]+$`).MatchString(got) {
		t.Fatalf("GenerateToken() returned unexpected charset: %q", got)
	}
}

func TestGenerateUUIDReturnsValidUUID(t *testing.T) {
	t.Parallel()

	got, err := GenerateUUID()
	if err != nil {
		t.Fatalf("GenerateUUID returned error: %v", err)
	}

	if _, err := uuid.Parse(got); err != nil {
		t.Fatalf("GenerateUUID() returned invalid UUID %q: %v", got, err)
	}
}

func TestSecureStringSupportsZeroLength(t *testing.T) {
	t.Parallel()

	got, err := secureString(0, "abc")
	if err != nil {
		t.Fatalf("secureString returned error: %v", err)
	}

	if got != "" {
		t.Fatalf("secureString() = %q, want empty string", got)
	}
}

func TestSecureStringRejectsEmptyCharset(t *testing.T) {
	t.Parallel()

	if _, err := secureString(5, ""); err == nil {
		t.Fatal("expected secureString to reject an empty charset")
	}
}
