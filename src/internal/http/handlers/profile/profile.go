package profile

import (
	"net/http"

	"familyplan/src/internal/domain"
	"familyplan/src/internal/view"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
)

// HandleProfilePage renders the profile page.
func HandleProfilePage(app *pocketbase.PocketBase) echo.HandlerFunc {
	return func(c echo.Context) error {
		session := c.Get("session").(domain.SessionData)
		if !session.IsAuthenticated {
			return c.Redirect(http.StatusSeeOther, "/login")
		}

		authCollection, err := app.Dao().FindCollectionByNameOrId("users")
		if err != nil {
			return err
		}

		authRecord, err := app.Dao().FindRecordById(authCollection.Id, session.UserID)
		if err != nil {
			return err
		}

		return view.RenderPage(c, "profile.html", map[string]interface{}{
			"title":   "Edit Profile - Family Plan Manager",
			"name":    authRecord.GetString("name"),
			"error":   c.QueryParam("error"),
			"success": c.QueryParam("success"),
		})
	}
}

// HandleProfileUpdate updates the user's display name.
func HandleProfileUpdate(app *pocketbase.PocketBase) echo.HandlerFunc {
	return func(c echo.Context) error {
		session := c.Get("session").(domain.SessionData)
		if !session.IsAuthenticated {
			return c.JSON(http.StatusUnauthorized, map[string]interface{}{
				"success": false,
				"message": "Not authenticated",
			})
		}

		authCollection, err := app.Dao().FindCollectionByNameOrId("users")
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"success": false,
				"message": "Failed to update profile",
			})
		}

		authRecord, err := app.Dao().FindRecordById(authCollection.Id, session.UserID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"success": false,
				"message": "Failed to update profile",
			})
		}

		name := c.FormValue("name")
		authRecord.Set("name", name)
		if err := app.Dao().SaveRecord(authRecord); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"success": false,
				"message": "Failed to update profile",
			})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"success": true,
			"message": "Profile updated successfully",
			"name":    name,
		})
	}
}
