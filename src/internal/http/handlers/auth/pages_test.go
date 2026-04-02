package auth

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"familyplan/src/internal/domain"

	"github.com/labstack/echo/v5"
)

func TestHandleHomeRendersLandingPage(t *testing.T) {
	t.Parallel()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := HandleHome()(c); err != nil {
		t.Fatalf("HandleHome returned error: %v", err)
	}

	if !strings.Contains(rec.Body.String(), "Family Plan Manager - Home") {
		t.Fatalf("expected landing page title in response, got %q", rec.Body.String())
	}
}

func TestHandleLoginPageRedirectsAuthenticatedUsers(t *testing.T) {
	t.Parallel()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/login", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("session", domain.SessionData{IsAuthenticated: true})

	if err := HandleLoginPage()(c); err != nil {
		t.Fatalf("HandleLoginPage returned error: %v", err)
	}

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusSeeOther)
	}
	if location := rec.Header().Get("Location"); location != "/family-plans" {
		t.Fatalf("Location = %q, want %q", location, "/family-plans")
	}
}

func TestHandleLoginPageRendersForAnonymousUsers(t *testing.T) {
	t.Parallel()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/login?error=bad+credentials", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := HandleLoginPage()(c); err != nil {
		t.Fatalf("HandleLoginPage returned error: %v", err)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "Login - Family Plan Manager") || !strings.Contains(body, "bad credentials") {
		t.Fatalf("expected login page content in response, got %q", body)
	}
}

func TestHandleRegisterPageRedirectsAuthenticatedUsers(t *testing.T) {
	t.Parallel()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/register", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("session", domain.SessionData{IsAuthenticated: true})

	if err := HandleRegisterPage()(c); err != nil {
		t.Fatalf("HandleRegisterPage returned error: %v", err)
	}

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusSeeOther)
	}
	if location := rec.Header().Get("Location"); location != "/family-plans" {
		t.Fatalf("Location = %q, want %q", location, "/family-plans")
	}
}

func TestHandleRegisterPageRendersForAnonymousUsers(t *testing.T) {
	t.Parallel()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/register?error=name+taken", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := HandleRegisterPage()(c); err != nil {
		t.Fatalf("HandleRegisterPage returned error: %v", err)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "Register - Family Plan Manager") || !strings.Contains(body, "name taken") {
		t.Fatalf("expected register page content in response, got %q", body)
	}
}

func TestHandleLogoutClearsCookieAndRedirectsHome(t *testing.T) {
	t.Parallel()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/logout", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := HandleLogout()(c); err != nil {
		t.Fatalf("HandleLogout returned error: %v", err)
	}

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusSeeOther)
	}
	if location := rec.Header().Get("Location"); location != "/" {
		t.Fatalf("Location = %q, want %q", location, "/")
	}

	cookies := rec.Result().Cookies()
	if len(cookies) == 0 || cookies[0].Name != "auth_token" || cookies[0].Value != "" {
		t.Fatalf("expected cleared auth cookie, got %+v", cookies)
	}
}
