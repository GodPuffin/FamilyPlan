package auth

import (
	"familyplan/src/internal/view"

	"github.com/labstack/echo/v5"
)

// HandleHome renders the public landing page.
func HandleHome() echo.HandlerFunc {
	return func(c echo.Context) error {
		return view.RenderPage(c, "index.html", map[string]interface{}{
			"title": "Family Plan Manager - Home",
		})
	}
}
