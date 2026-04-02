package auth

import (
	"net/http"
	"testing"
	"time"
)

func TestSecureCookiesEnabledDefaultsToTrue(t *testing.T) {
	t.Setenv("FAMILYPLAN_COOKIE_SECURE", "")

	if !secureCookiesEnabled() {
		t.Fatal("expected secure cookies to default to enabled")
	}
}

func TestSecureCookiesEnabledUsesEnvOverride(t *testing.T) {
	t.Setenv("FAMILYPLAN_COOKIE_SECURE", "false")

	if secureCookiesEnabled() {
		t.Fatal("expected secure cookies to be disabled by env override")
	}
}

func TestSecureCookiesEnabledFallsBackToTrueForInvalidValues(t *testing.T) {
	t.Setenv("FAMILYPLAN_COOKIE_SECURE", "definitely-not-a-bool")

	if !secureCookiesEnabled() {
		t.Fatal("expected invalid env values to keep secure cookies enabled")
	}
}

func TestNewAuthCookieUsesExpectedSettings(t *testing.T) {
	t.Setenv("FAMILYPLAN_COOKIE_SECURE", "false")
	expires := time.Date(2026, time.April, 3, 12, 0, 0, 0, time.UTC)

	cookie := newAuthCookie("token-123", expires)

	if cookie.Name != "auth_token" {
		t.Fatalf("Name = %q, want %q", cookie.Name, "auth_token")
	}
	if cookie.Value != "token-123" {
		t.Fatalf("Value = %q, want %q", cookie.Value, "token-123")
	}
	if cookie.Path != "/" {
		t.Fatalf("Path = %q, want %q", cookie.Path, "/")
	}
	if !cookie.Expires.Equal(expires) {
		t.Fatalf("Expires = %v, want %v", cookie.Expires, expires)
	}
	if !cookie.HttpOnly {
		t.Fatal("expected cookie to be HttpOnly")
	}
	if cookie.SameSite != http.SameSiteLaxMode {
		t.Fatalf("SameSite = %v, want %v", cookie.SameSite, http.SameSiteLaxMode)
	}
	if cookie.Secure {
		t.Fatal("expected cookie Secure flag to follow env override")
	}
}
