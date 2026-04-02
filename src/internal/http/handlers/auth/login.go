package auth

import (
	"net/http"
	"time"

	"familyplan/src/internal/http/sessionutil"
	"familyplan/src/internal/support/random"
	"familyplan/src/internal/view"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
)

// HandleLoginPage renders the login page.
func HandleLoginPage() echo.HandlerFunc {
	return func(c echo.Context) error {
		if session, ok := sessionutil.Current(c); ok && session.IsAuthenticated {
			return c.Redirect(http.StatusSeeOther, "/family-plans")
		}

		return view.RenderPage(c, "login.html", map[string]interface{}{
			"title": "Login - Family Plan Manager",
			"error": c.QueryParam("error"),
		})
	}
}

// HandleLoginSubmit authenticates a user and stores an auth token cookie.
func HandleLoginSubmit(app *pocketbase.PocketBase) echo.HandlerFunc {
	return func(c echo.Context) error {
		username := c.FormValue("username")
		password := c.FormValue("password")

		authCollection, err := app.Dao().FindCollectionByNameOrId("users")
		if err != nil {
			return c.Redirect(http.StatusSeeOther, "/login?error=Authentication+failed")
		}

		authRecord, err := app.Dao().FindAuthRecordByUsername(authCollection.Id, username)
		if err != nil {
			return c.Redirect(http.StatusSeeOther, "/login?error=Invalid+credentials")
		}

		if !authRecord.ValidatePassword(password) {
			return c.Redirect(http.StatusSeeOther, "/login?error=Invalid+credentials")
		}

		token, err := random.GenerateToken()
		if err != nil {
			return c.Redirect(http.StatusSeeOther, "/login?error=Authentication+failed")
		}
		authRecord.Set("tokenKey", token)
		if err := app.Dao().SaveRecord(authRecord); err != nil {
			return c.Redirect(http.StatusSeeOther, "/login?error=Authentication+failed")
		}

		c.SetCookie(newAuthCookie(token, time.Now().Add(30*24*time.Hour)))
		return c.Redirect(http.StatusSeeOther, "/family-plans")
	}
}
