package main

import (
	"embed"
	"html/template"
	"net/http"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/models"
)

// Setup auth middleware
func setupAuthMiddleware(app *pocketbase.PocketBase, templatesFS embed.FS) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Default session data (not authenticated)
			session := SessionData{
				IsAuthenticated: false,
			}

			// Get auth cookie
			cookie, err := c.Cookie("auth_token")

			// If we have a valid auth cookie, get the user info
			if err == nil && cookie.Value != "" {
				// Try to find the auth record by tokenKey
				authCollection, err := app.Dao().FindCollectionByNameOrId("users")
				if err == nil {
					// Find user with matching tokenKey
					record, err := app.Dao().FindFirstRecordByData(authCollection.Id, "tokenKey", cookie.Value)
					if err == nil && record != nil {
						// Ensure the user is verified
						if record.Verified() {
							session.IsAuthenticated = true
							session.UserId = record.Id
							session.Username = record.GetString("username")
							session.Name = record.GetString("name")
						}
					}
				}
			}

			// Add session data to context
			c.Set("session", session)
			return next(c)
		}
	}
}

// Require authentication middleware
func requireAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		session := c.Get("session").(SessionData)
		if !session.IsAuthenticated {
			return c.Redirect(http.StatusSeeOther, "/login")
		}
		return next(c)
	}
}

// Home page handler
func handleHome(templatesFS embed.FS) echo.HandlerFunc {
	return func(c echo.Context) error {
		session := c.Get("session").(SessionData)

		tmpl, err := template.ParseFS(templatesFS, "templates/layout.html", "templates/index.html")
		if err != nil {
			return err
		}

		return tmpl.ExecuteTemplate(c.Response().Writer, "layout", map[string]interface{}{
			"title":           "Family Plan Manager - Home",
			"isAuthenticated": session.IsAuthenticated,
			"username":        session.Username,
			"name":            session.Name,
		})
	}
}

// Login page handler
func handleLoginPage(templatesFS embed.FS) echo.HandlerFunc {
	return func(c echo.Context) error {
		session := c.Get("session").(SessionData)
		if session.IsAuthenticated {
			return c.Redirect(http.StatusSeeOther, "/family-plans")
		}

		tmpl, err := template.ParseFS(templatesFS, "templates/layout.html", "templates/login.html")
		if err != nil {
			return err
		}

		return tmpl.ExecuteTemplate(c.Response().Writer, "layout", map[string]interface{}{
			"title":           "Login - Family Plan Manager",
			"error":           c.QueryParam("error"),
			"isAuthenticated": session.IsAuthenticated,
			"username":        session.Username,
			"name":            session.Name,
		})
	}
}

// Login form submission handler
func handleLoginSubmit(app *pocketbase.PocketBase) echo.HandlerFunc {
	return func(c echo.Context) error {
		username := c.FormValue("username")
		password := c.FormValue("password")

		// Use PocketBase's auth collection
		authCollection, err := app.Dao().FindCollectionByNameOrId("users")
		if err != nil {
			return c.Redirect(http.StatusSeeOther, "/login?error=Authentication+failed")
		}

		// Find the user record
		authRecord, err := app.Dao().FindAuthRecordByUsername(authCollection.Id, username)
		if err != nil {
			return c.Redirect(http.StatusSeeOther, "/login?error=Invalid+username+or+password")
		}

		// Make sure the record is verified
		if !authRecord.Verified() {
			// Verify the account automatically
			if err := authRecord.SetVerified(true); err != nil {
				return c.Redirect(http.StatusSeeOther, "/login?error=Account+verification+failed")
			}
			if err := app.Dao().SaveRecord(authRecord); err != nil {
				return c.Redirect(http.StatusSeeOther, "/login?error=Account+verification+failed")
			}
		}

		// Validate password
		if !authRecord.ValidatePassword(password) {
			// If password validation fails, try to log more details (in a real app, don't log passwords)
			return c.Redirect(http.StatusSeeOther, "/login?error=Invalid+password")
		}

		// Generate a token
		token := generateRandomToken()

		// Store the token in the auth record
		authRecord.Set("tokenKey", token)
		if err := app.Dao().SaveRecord(authRecord); err != nil {
			return c.Redirect(http.StatusSeeOther, "/login?error=Authentication+failed")
		}

		// Set auth cookie
		cookie := &http.Cookie{
			Name:     "auth_token",
			Value:    token,
			Path:     "/",
			Expires:  time.Now().Add(30 * 24 * time.Hour), // 30 days
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
			Secure:   c.Scheme() == "https",
		}
		c.SetCookie(cookie)

		// Redirect to family plans page
		return c.Redirect(http.StatusSeeOther, "/family-plans")
	}
}

// Register page handler
func handleRegisterPage(templatesFS embed.FS) echo.HandlerFunc {
	return func(c echo.Context) error {
		session := c.Get("session").(SessionData)
		if session.IsAuthenticated {
			return c.Redirect(http.StatusSeeOther, "/family-plans")
		}

		tmpl, err := template.ParseFS(templatesFS, "templates/layout.html", "templates/register.html")
		if err != nil {
			return err
		}

		return tmpl.ExecuteTemplate(c.Response().Writer, "layout", map[string]interface{}{
			"title":           "Register - Family Plan Manager",
			"error":           c.QueryParam("error"),
			"isAuthenticated": session.IsAuthenticated,
			"username":        session.Username,
			"name":            session.Name,
		})
	}
}

// Register form submission handler
func handleRegisterSubmit(app *pocketbase.PocketBase) echo.HandlerFunc {
	return func(c echo.Context) error {
		username := c.FormValue("username")
		password := c.FormValue("password")
		passwordConfirm := c.FormValue("passwordConfirm")

		// Validate inputs
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

		// Use PocketBase's auth collection
		authCollection, err := app.Dao().FindCollectionByNameOrId("users")
		if err != nil {
			return c.Redirect(http.StatusSeeOther, "/register?error=Registration+failed")
		}

		// Check if username already exists
		_, err = app.Dao().FindAuthRecordByUsername(authCollection.Id, username)
		if err == nil {
			return c.Redirect(http.StatusSeeOther, "/register?error=Username+already+exists")
		}

		// Create the auth record with proper password handling
		record := models.NewRecord(authCollection)

		// Set required fields
		record.Set("username", username)

		// Use the proper method to set password
		if err := record.SetPassword(password); err != nil {
			return c.Redirect(http.StatusSeeOther, "/register?error=Password+setup+failed")
		}

		if err := record.SetVerified(true); err != nil { // Ensure the account is verified
			return c.Redirect(http.StatusSeeOther, "/register?error=Verification+failed")
		}

		// Save the record
		if err := app.Dao().SaveRecord(record); err != nil {
			return c.Redirect(http.StatusSeeOther, "/register?error=Registration+failed")
		}

		// Generate a token for immediate authentication
		token := generateRandomToken()

		// Store the token in the auth record
		record.Set("tokenKey", token)
		if err := app.Dao().SaveRecord(record); err != nil {
			return c.Redirect(http.StatusSeeOther, "/login?success=Registration+successful.+Please+login.")
		}

		// Set auth cookie
		cookie := &http.Cookie{
			Name:     "auth_token",
			Value:    token,
			Path:     "/",
			Expires:  time.Now().Add(30 * 24 * time.Hour), // 30 days
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		}
		c.SetCookie(cookie)

		// Redirect to family plans page
		return c.Redirect(http.StatusSeeOther, "/family-plans")
	}
}

// Logout handler
func handleLogout() echo.HandlerFunc {
	return func(c echo.Context) error {
		// Delete the auth cookie
		cookie := &http.Cookie{
			Name:     "auth_token",
			Value:    "",
			Path:     "/",
			Expires:  time.Now().Add(-1 * time.Hour), // In the past
			HttpOnly: true,
		}
		c.SetCookie(cookie)

		// Redirect to home page
		return c.Redirect(http.StatusSeeOther, "/")
	}
}

// Profile page handler
func handleProfilePage(app *pocketbase.PocketBase, templatesFS embed.FS) echo.HandlerFunc {
	return func(c echo.Context) error {
		session := c.Get("session").(SessionData)
		if !session.IsAuthenticated {
			return c.Redirect(http.StatusSeeOther, "/login")
		}

		// Get the user record
		authCollection, err := app.Dao().FindCollectionByNameOrId("users")
		if err != nil {
			return err
		}

		// Find the user record
		authRecord, err := app.Dao().FindRecordById(authCollection.Id, session.UserId)
		if err != nil {
			return err
		}

		// Get the current display name
		name := authRecord.GetString("name")

		tmpl, err := template.ParseFS(templatesFS, "templates/layout.html", "templates/profile.html")
		if err != nil {
			return err
		}

		return tmpl.ExecuteTemplate(c.Response().Writer, "layout", map[string]interface{}{
			"title":           "Edit Profile - Family Plan Manager",
			"isAuthenticated": session.IsAuthenticated,
			"username":        session.Username,
			"name":            name,
			"error":           c.QueryParam("error"),
			"success":         c.QueryParam("success"),
		})
	}
}

// Profile update handler
func handleProfileUpdate(app *pocketbase.PocketBase) echo.HandlerFunc {
	return func(c echo.Context) error {
		session := c.Get("session").(SessionData)
		if !session.IsAuthenticated {
			return c.JSON(http.StatusUnauthorized, map[string]interface{}{
				"success": false,
				"message": "Not authenticated",
			})
		}

		// Get the display name from the form
		name := c.FormValue("name")

		// Get the user record
		authCollection, err := app.Dao().FindCollectionByNameOrId("users")
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"success": false,
				"message": "Failed to update profile",
			})
		}

		// Find the user record
		authRecord, err := app.Dao().FindRecordById(authCollection.Id, session.UserId)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"success": false,
				"message": "Failed to update profile",
			})
		}

		// Update the name field
		authRecord.Set("name", name)

		// Save the record
		if err := app.Dao().SaveRecord(authRecord); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"success": false,
				"message": "Failed to update profile",
			})
		}

		// Return success JSON response
		return c.JSON(http.StatusOK, map[string]interface{}{
			"success": true,
			"message": "Profile updated successfully",
			"name":    name,
		})
	}
}
