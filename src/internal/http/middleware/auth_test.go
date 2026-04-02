package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"familyplan/src/internal/domain"
	"familyplan/src/internal/http/sessionutil"

	"github.com/labstack/echo/v5"
)

func TestSetupAuthDefaultsToAnonymousSessionWithoutCookie(t *testing.T) {
	t.Parallel()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := SetupAuth(nil)(func(c echo.Context) error {
		session, ok := sessionutil.Current(c)
		if !ok {
			t.Fatal("expected middleware to populate session")
		}

		if session.IsAuthenticated {
			t.Fatalf("session = %+v, want anonymous user", session)
		}

		return c.NoContent(http.StatusNoContent)
	})

	if err := handler(c); err != nil {
		t.Fatalf("handler returned error: %v", err)
	}

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
}

func TestRequireAuthRedirectsAnonymousUsers(t *testing.T) {
	t.Parallel()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/family-plans", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("session", domain.SessionData{})

	called := false
	err := RequireAuth(func(c echo.Context) error {
		called = true
		return c.NoContent(http.StatusOK)
	})(c)
	if err != nil {
		t.Fatalf("handler returned error: %v", err)
	}

	if called {
		t.Fatal("expected next handler not to be called")
	}

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusSeeOther)
	}

	if location := rec.Header().Get("Location"); location != "/login" {
		t.Fatalf("Location = %q, want %q", location, "/login")
	}
}

func TestRequireAuthAllowsAuthenticatedUsers(t *testing.T) {
	t.Parallel()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/family-plans", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("session", domain.SessionData{IsAuthenticated: true, UserID: "user_123"})

	called := false
	err := RequireAuth(func(c echo.Context) error {
		called = true
		return c.NoContent(http.StatusNoContent)
	})(c)
	if err != nil {
		t.Fatalf("handler returned error: %v", err)
	}

	if !called {
		t.Fatal("expected next handler to be called")
	}

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
}
