package auth

import (
	"net/http"
	"time"

	"familyplan/src/internal/http/sessionutil"
	"familyplan/src/internal/memberclaim"
	"familyplan/src/internal/support/random"
	"familyplan/src/internal/view"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
	pbmodels "github.com/pocketbase/pocketbase/models"
)

// HandleRegisterPage renders the registration page.
func HandleRegisterPage() echo.HandlerFunc {
	return func(c echo.Context) error {
		claimToken := currentClaimToken(c)
		if session, ok := sessionutil.Current(c); ok && session.IsAuthenticated {
			if claimToken != "" {
				return c.Redirect(http.StatusSeeOther, memberclaim.Path(claimToken))
			}
			return c.Redirect(http.StatusSeeOther, "/family-plans")
		}

		return view.RenderPage(c, "register.html", map[string]interface{}{
			"title": "Register - Family Plan Manager",
			"error": c.QueryParam("error"),
			"claim": claimToken,
		})
	}
}

// HandleRegisterSubmit registers a user and signs them in immediately.
func HandleRegisterSubmit(app *pocketbase.PocketBase) echo.HandlerFunc {
	return func(c echo.Context) error {
		claimToken := currentClaimToken(c)
		username := c.FormValue("username")
		password := c.FormValue("password")
		passwordConfirm := c.FormValue("passwordConfirm")

		if len(username) < 3 {
			return c.Redirect(http.StatusSeeOther, buildAuthPagePath("/register", claimToken, "Username must be at least 3 characters"))
		}
		if len(password) < 8 {
			return c.Redirect(http.StatusSeeOther, buildAuthPagePath("/register", claimToken, "Password must be at least 8 characters"))
		}
		if len(password) > 72 {
			return c.Redirect(http.StatusSeeOther, buildAuthPagePath("/register", claimToken, "Password must be no more than 72 characters"))
		}
		if password != passwordConfirm {
			return c.Redirect(http.StatusSeeOther, buildAuthPagePath("/register", claimToken, "Passwords do not match"))
		}

		authCollection, err := app.Dao().FindCollectionByNameOrId("users")
		if err != nil {
			return c.Redirect(http.StatusSeeOther, buildAuthPagePath("/register", claimToken, "Registration failed"))
		}

		if _, err := app.Dao().FindAuthRecordByUsername(authCollection.Id, username); err == nil {
			return c.Redirect(http.StatusSeeOther, buildAuthPagePath("/register", claimToken, "Username already exists"))
		}

		record := pbmodels.NewRecord(authCollection)
		record.Set("username", username)

		if err := record.SetPassword(password); err != nil {
			return c.Redirect(http.StatusSeeOther, buildAuthPagePath("/register", claimToken, "Password setup failed"))
		}
		if err := record.SetVerified(true); err != nil {
			return c.Redirect(http.StatusSeeOther, buildAuthPagePath("/register", claimToken, "Verification failed"))
		}
		if err := app.Dao().SaveRecord(record); err != nil {
			return c.Redirect(http.StatusSeeOther, buildAuthPagePath("/register", claimToken, "Registration failed"))
		}

		token, err := random.GenerateToken()
		if err != nil {
			return c.Redirect(http.StatusSeeOther, buildAuthPagePath("/register", claimToken, "Registration failed"))
		}
		record.Set("tokenKey", token)
		if err := app.Dao().SaveRecord(record); err != nil {
			return c.Redirect(http.StatusSeeOther, buildPathWithQuery("/login", mapSuccessAndClaim(claimToken, "Registration successful. Please login.")))
		}

		c.SetCookie(newAuthCookie(token, time.Now().Add(30*24*time.Hour)))
		return redirectAfterAuth(c, app, record.Id, claimToken)
	}
}
