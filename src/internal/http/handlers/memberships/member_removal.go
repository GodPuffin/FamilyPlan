package memberships

import (
	"net/http"
	"time"

	"familyplan/src/internal/domain"
	"familyplan/src/internal/planutil"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
)

// HandleRemoveMember ends another member's membership.
func HandleRemoveMember(app *pocketbase.PocketBase) echo.HandlerFunc {
	return func(c echo.Context) error {
		session := c.Get("session").(domain.SessionData)
		joinCode := c.PathParam("join_code")
		memberID := c.FormValue("user_id")

		planRecord, err := planutil.FindPlanByJoinCode(app, joinCode)
		if err != nil || planRecord == nil {
			return c.Redirect(http.StatusSeeOther, "/family-plans")
		}

		if !planutil.IsOwner(planRecord, session.UserID) {
			return c.Redirect(http.StatusSeeOther, "/"+joinCode)
		}

		membership, err := planutil.FindMembership(app, planRecord.Id, memberID)
		if err != nil || membership == nil {
			return c.Redirect(http.StatusSeeOther, "/"+joinCode)
		}

		membership.Set("date_ended", time.Now())
		if err := app.Dao().SaveRecord(membership); err != nil {
			return err
		}

		return c.Redirect(http.StatusSeeOther, "/"+joinCode)
	}
}
