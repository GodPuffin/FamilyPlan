package profile

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase/apis"
)

func TestHandleProfilePageRedirectsWhenSessionMissing(t *testing.T) {
	t.Parallel()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/profile", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := HandleProfilePage(nil)(c); err != nil {
		t.Fatalf("HandleProfilePage returned error: %v", err)
	}

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusSeeOther)
	}
	if location := rec.Header().Get("Location"); location != "/login" {
		t.Fatalf("Location = %q, want %q", location, "/login")
	}
}

func TestHandleProfileUpdateRejectsMissingSession(t *testing.T) {
	t.Parallel()

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/profile", strings.NewReader("name=Marcus"))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := HandleProfileUpdate(nil)(c); err != nil {
		t.Fatalf("HandleProfileUpdate returned error: %v", err)
	}

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
	if !strings.Contains(rec.Body.String(), "Authentication required") {
		t.Fatalf("expected auth error in response, got %q", rec.Body.String())
	}
}

func TestProfileUpdateErrorMessageUsesValidationMessage(t *testing.T) {
	t.Parallel()

	err := validation.Errors{
		"avatar": validation.NewError("validation_invalid_mime_type", "Avatar must be a JPG or PNG."),
	}

	if got := profileUpdateErrorMessage(err); got != "Avatar must be a JPG or PNG." {
		t.Fatalf("profileUpdateErrorMessage() = %q, want %q", got, "Avatar must be a JPG or PNG.")
	}
}

func TestProfileUpdateErrorMessageUsesAPIValidationMessage(t *testing.T) {
	t.Parallel()

	err := apis.NewBadRequestError("Failed to update profile", validation.Errors{
		"avatar": validation.NewError("validation_invalid_mime_type", "Avatar must be a JPG or PNG."),
	})

	if got := profileUpdateErrorMessage(err); got != "Avatar must be a JPG or PNG." {
		t.Fatalf("profileUpdateErrorMessage() = %q, want %q", got, "Avatar must be a JPG or PNG.")
	}
}

func TestProfileUpdateErrorMessageFallsBackForGenericErrors(t *testing.T) {
	t.Parallel()

	if got := profileUpdateErrorMessage(errors.New("disk exploded")); got != "Failed to update profile" {
		t.Fatalf("profileUpdateErrorMessage() = %q, want %q", got, "Failed to update profile")
	}
}
