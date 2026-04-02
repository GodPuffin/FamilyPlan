package sessionutil

import (
	"familyplan/src/internal/domain"

	"github.com/labstack/echo/v5"
)

// Current returns the request session when the auth middleware has populated it.
func Current(c echo.Context) (domain.SessionData, bool) {
	session, ok := c.Get("session").(domain.SessionData)
	return session, ok
}
