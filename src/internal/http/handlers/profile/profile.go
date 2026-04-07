package profile

import (
	"errors"
	"net/http"
	"strings"
	"unicode/utf8"

	"familyplan/src/internal/http/sessionutil"
	"familyplan/src/internal/userprofile"
	"familyplan/src/internal/view"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/forms"
	"github.com/pocketbase/pocketbase/tools/rest"
)

const maxDisplayNameLength = 80

// HandleProfilePage renders the profile page.
func HandleProfilePage(app *pocketbase.PocketBase) echo.HandlerFunc {
	return func(c echo.Context) error {
		session, ok := sessionutil.Current(c)
		if !ok {
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
			"title":     "Edit Profile - Family Plan Manager",
			"name":      authRecord.GetString("name"),
			"username":  authRecord.GetString("username"),
			"avatarURL": userprofile.AvatarURL(authRecord),
			"error":     c.QueryParam("error"),
			"success":   c.QueryParam("success"),
		})
	}
}

// HandleProfileUpdate updates the user's profile.
func HandleProfileUpdate(app *pocketbase.PocketBase) echo.HandlerFunc {
	return func(c echo.Context) error {
		session, ok := sessionutil.Current(c)
		if !ok {
			return c.JSON(http.StatusUnauthorized, map[string]interface{}{
				"success": false,
				"message": "Authentication required",
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

		name := strings.TrimSpace(c.FormValue("name"))
		if name == "" {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"success": false,
				"message": "Display name is required",
			})
		}

		if utf8.RuneCountInString(name) > maxDisplayNameLength {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"success": false,
				"message": "Display name must be 80 characters or fewer",
			})
		}

		form := forms.NewRecordUpsert(app, authRecord)
		if err := form.LoadData(map[string]any{"name": name}); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"success": false,
				"message": "Failed to update profile",
			})
		}

		if strings.HasPrefix(c.Request().Header.Get(echo.HeaderContentType), "multipart/form-data") {
			files, err := rest.FindUploadedFiles(c.Request(), "avatar")
			if err != nil && err != http.ErrMissingFile {
				return c.JSON(http.StatusBadRequest, map[string]interface{}{
					"success": false,
					"message": "Could not read uploaded profile picture",
				})
			}
			if len(files) > 0 {
				if err := form.AddFiles("avatar", files...); err != nil {
					return c.JSON(http.StatusBadRequest, map[string]interface{}{
						"success": false,
						"message": "Could not attach profile picture",
					})
				}
			}
		}

		if err := form.Submit(); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"success": false,
				"message": profileUpdateErrorMessage(err),
			})
		}

		updatedRecord, err := app.Dao().FindRecordById(authCollection.Id, session.UserID)
		if err != nil {
			updatedRecord = authRecord
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"success":   true,
			"message":   "Profile updated successfully",
			"name":      name,
			"username":  updatedRecord.GetString("username"),
			"avatarURL": userprofile.AvatarURL(updatedRecord),
		})
	}
}

func profileUpdateErrorMessage(err error) string {
	if msg := firstProfileErrorMessage(err); msg != "" {
		return msg
	}

	return "Failed to update profile"
}

func firstProfileErrorMessage(value any) string {
	switch v := value.(type) {
	case *apis.ApiError:
		if msg := firstProfileErrorMessage(v.RawData()); msg != "" {
			return msg
		}
		if msg := firstProfileErrorMessage(v.Data); msg != "" {
			return msg
		}
		return strings.TrimSpace(v.Message)
	case validation.Errors:
		for _, err := range v {
			if msg := firstProfileErrorMessage(err); msg != "" {
				return msg
			}
		}
	case map[string]any:
		if msg, ok := v["message"].(string); ok && strings.TrimSpace(msg) != "" {
			return strings.TrimSpace(msg)
		}
		for _, item := range v {
			if msg := firstProfileErrorMessage(item); msg != "" {
				return msg
			}
		}
	case map[string]string:
		if msg := strings.TrimSpace(v["message"]); msg != "" {
			return msg
		}
	case validation.Error:
		return strings.TrimSpace(v.Error())
	case error:
		var apiErr *apis.ApiError
		if errors.As(v, &apiErr) {
			return firstProfileErrorMessage(apiErr)
		}

		var validationErrs validation.Errors
		if errors.As(v, &validationErrs) {
			return firstProfileErrorMessage(validationErrs)
		}
	case string:
		return strings.TrimSpace(v)
	}

	return ""
}
