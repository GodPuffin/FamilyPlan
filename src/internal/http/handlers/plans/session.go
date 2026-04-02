package plans

import (
	"net/http"

	"familyplan/src/internal/domain"
	"familyplan/src/internal/http/sessionutil"

	"github.com/labstack/echo/v5"
)

func sessionOrRedirect(c echo.Context) (domain.SessionData, error) {
	session, ok := sessionutil.Current(c)
	if !ok {
		return domain.SessionData{}, c.Redirect(http.StatusSeeOther, "/login")
	}

	return session, nil
}
