package auth

import (
	"net/http"
	"net/url"
	"strings"

	"familyplan/src/internal/memberclaim"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
)

func currentClaimToken(c echo.Context) string {
	if token := strings.TrimSpace(c.FormValue("claim")); token != "" {
		return token
	}

	return strings.TrimSpace(c.QueryParam("claim"))
}

func buildAuthPagePath(path, claimToken, errorMessage string) string {
	values := url.Values{}
	if strings.TrimSpace(errorMessage) != "" {
		values.Set("error", errorMessage)
	}
	if strings.TrimSpace(claimToken) != "" {
		values.Set("claim", claimToken)
	}

	return buildPathWithQuery(path, values)
}

func mapSuccessAndClaim(claimToken, successMessage string) url.Values {
	values := url.Values{}
	if strings.TrimSpace(successMessage) != "" {
		values.Set("success", successMessage)
	}
	if strings.TrimSpace(claimToken) != "" {
		values.Set("claim", claimToken)
	}

	return values
}

func buildPathWithQuery(path string, values url.Values) string {
	if values == nil {
		return path
	}

	encoded := values.Encode()
	if encoded == "" {
		return path
	}

	return path + "?" + encoded
}

func redirectAfterAuth(c echo.Context, app *pocketbase.PocketBase, userID, claimToken string) error {
	if strings.TrimSpace(claimToken) == "" {
		return c.Redirect(http.StatusSeeOther, "/family-plans")
	}

	result, err := memberclaim.Claim(app, claimToken, userID)
	if err != nil {
		if message := memberclaim.ErrorMessage(err); message != "" {
			values := url.Values{}
			values.Set("error", message)
			return c.Redirect(http.StatusSeeOther, buildPathWithQuery(memberclaim.Path(claimToken), values))
		}

		return err
	}

	return c.Redirect(http.StatusSeeOther, "/"+result.JoinCode)
}
