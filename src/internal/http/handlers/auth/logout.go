package auth

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v5"
)

// HandleLogout clears the auth cookie and returns the user to the home page.
func HandleLogout() echo.HandlerFunc {
	return func(c echo.Context) error {
		c.SetCookie(newAuthCookie("", time.Now().Add(-time.Hour)))
		return c.Redirect(http.StatusSeeOther, "/")
	}
}
