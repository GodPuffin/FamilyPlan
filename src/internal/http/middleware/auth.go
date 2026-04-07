package middleware

import (
	"net/http"

	"familyplan/src/internal/domain"
	"familyplan/src/internal/http/sessionutil"
	"familyplan/src/internal/userprofile"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
)

// SetupAuth populates session data for each request.
func SetupAuth(app *pocketbase.PocketBase) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			session := domain.SessionData{
				IsAuthenticated: false,
			}

			cookie, err := c.Cookie("auth_token")
			if err == nil && cookie.Value != "" {
				authCollection, err := app.Dao().FindCollectionByNameOrId("users")
				if err == nil {
					record, err := app.Dao().FindFirstRecordByData(authCollection.Id, "tokenKey", cookie.Value)
					if err == nil && record != nil && record.Verified() {
						session.IsAuthenticated = true
						session.UserID = record.Id
						session.Username = record.GetString("username")
						session.Name = record.GetString("name")
						session.AvatarURL = userprofile.AvatarURL(record)
					}
				}
			}

			c.Set("session", session)
			return next(c)
		}
	}
}

// RequireAuth redirects anonymous users to login.
func RequireAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		session, ok := sessionutil.Current(c)
		if !ok || !session.IsAuthenticated {
			return c.Redirect(http.StatusSeeOther, "/login")
		}
		return next(c)
	}
}
