package plans

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"familyplan/src/internal/domain"

	"github.com/labstack/echo/v5"
)

func TestSessionOrRedirectReturnsSession(t *testing.T) {
	t.Parallel()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/plan", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("session", domain.SessionData{IsAuthenticated: true, UserID: "user_123"})

	session, err := sessionOrRedirect(c)
	if err != nil {
		t.Fatalf("sessionOrRedirect returned error: %v", err)
	}
	if session.UserID != "user_123" {
		t.Fatalf("sessionOrRedirect returned %+v, want user_123", session)
	}
}

func TestSessionOrRedirectRedirectsWhenMissingSession(t *testing.T) {
	t.Parallel()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/plan", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	session, err := sessionOrRedirect(c)
	if err != nil {
		t.Fatalf("sessionOrRedirect returned error: %v", err)
	}

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusSeeOther)
	}
	if location := rec.Header().Get("Location"); location != "/login" {
		t.Fatalf("Location = %q, want %q", location, "/login")
	}

	if session != (domain.SessionData{}) {
		t.Fatalf("sessionOrRedirect returned %+v, want zero value", session)
	}
}
