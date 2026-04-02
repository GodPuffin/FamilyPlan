package sessionutil

import (
	"net/http/httptest"
	"testing"

	"familyplan/src/internal/domain"

	"github.com/labstack/echo/v5"
)

func TestCurrentReturnsSession(t *testing.T) {
	t.Parallel()

	e := echo.New()
	c := e.NewContext(httptest.NewRequest("GET", "/", nil), httptest.NewRecorder())
	want := domain.SessionData{
		IsAuthenticated: true,
		UserID:          "user_123",
		Username:        "marcus",
		Name:            "Marcus",
	}
	c.Set("session", want)

	got, ok := Current(c)
	if !ok {
		t.Fatal("expected session to be present")
	}

	if got != want {
		t.Fatalf("Current() = %+v, want %+v", got, want)
	}
}

func TestCurrentRejectsUnexpectedSessionValue(t *testing.T) {
	t.Parallel()

	e := echo.New()
	c := e.NewContext(httptest.NewRequest("GET", "/", nil), httptest.NewRecorder())
	c.Set("session", "invalid")

	got, ok := Current(c)
	if ok {
		t.Fatal("expected Current to reject unexpected session value")
	}

	if got != (domain.SessionData{}) {
		t.Fatalf("Current() = %+v, want zero value", got)
	}
}
