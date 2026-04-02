package auth

import (
	"net/http"
	"time"

	"familyplan/src/internal/domain"
	"familyplan/src/internal/support/random"
	"familyplan/src/internal/view"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
	pbmodels "github.com/pocketbase/pocketbase/models"
)

// HandleRegisterPage renders the registration page.
func HandleRegisterPage() echo.HandlerFunc {
	return func(c echo.Context) error {
		session := c.Get("session").(domain.SessionData)
		if session.IsAuthenticated {
			return c.Redirect(http.StatusSeeOther, "/family-plans")
		}

		return view.RenderPage(c, "register.html", map[string]interface{}{
			"title": "Register - Family Plan Manager",
			"error": c.QueryParam("error"),
		})
	}
}

// HandleRegisterSubmit registers a user and signs them in immediately.
func HandleRegisterSubmit(app *pocketbase.PocketBase) echo.HandlerFunc {
	return func(c echo.Context) error {
		username := c.FormValue("username")
		password := c.FormValue("password")
		passwordConfirm := c.FormValue("passwordConfirm")

		if len(username) < 3 {
			return c.Redirect(http.StatusSeeOther, "/register?error=Username+must+be+at+least+3+characters")
		}
		if len(password) < 8 {
			return c.Redirect(http.StatusSeeOther, "/register?error=Password+must+be+at+least+8+characters")
		}
		if len(password) > 72 {
			return c.Redirect(http.StatusSeeOther, "/register?error=Password+must+be+no+more+than+72+characters")
		}
		if password != passwordConfirm {
			return c.Redirect(http.StatusSeeOther, "/register?error=Passwords+do+not+match")
		}

		authCollection, err := app.Dao().FindCollectionByNameOrId("users")
		if err != nil {
			return c.Redirect(http.StatusSeeOther, "/register?error=Registration+failed")
		}

		if _, err := app.Dao().FindAuthRecordByUsername(authCollection.Id, username); err == nil {
			return c.Redirect(http.StatusSeeOther, "/register?error=Username+already+exists")
		}

		record := pbmodels.NewRecord(authCollection)
		record.Set("username", username)

		if err := record.SetPassword(password); err != nil {
			return c.Redirect(http.StatusSeeOther, "/register?error=Password+setup+failed")
		}
		if err := record.SetVerified(true); err != nil {
			return c.Redirect(http.StatusSeeOther, "/register?error=Verification+failed")
		}
		if err := app.Dao().SaveRecord(record); err != nil {
			return c.Redirect(http.StatusSeeOther, "/register?error=Registration+failed")
		}

		token, err := random.GenerateToken()
		if err != nil {
			return c.Redirect(http.StatusSeeOther, "/register?error=Registration+failed")
		}
		record.Set("tokenKey", token)
		if err := app.Dao().SaveRecord(record); err != nil {
			return c.Redirect(http.StatusSeeOther, "/login?success=Registration+successful.+Please+login.")
		}

		c.SetCookie(newAuthCookie(token, time.Now().Add(30*24*time.Hour)))
		return c.Redirect(http.StatusSeeOther, "/family-plans")
	}
}
