package memberships

import (
	"net/http"
	"time"

	"familyplan/src/internal/billing"
	"familyplan/src/internal/domain"
	"familyplan/src/internal/planutil"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
)

// HandleLeavePlan either leaves immediately or marks leave_requested.
func HandleLeavePlan(app *pocketbase.PocketBase) echo.HandlerFunc {
	return func(c echo.Context) error {
		session := c.Get("session").(domain.SessionData)
		joinCode := c.PathParam("join_code")

		planRecord, err := planutil.FindPlanByJoinCode(app, joinCode)
		if err != nil || planRecord == nil {
			return c.Redirect(http.StatusSeeOther, "/family-plans")
		}

		if planutil.IsOwner(planRecord, session.UserID) {
			return c.Redirect(http.StatusSeeOther, "/"+joinCode)
		}

		existingMembership, _ := planutil.FindMembership(app, planRecord.Id, session.UserID)
		if existingMembership == nil {
			return c.Redirect(http.StatusSeeOther, "/family-plans")
		}

		balance, err := billing.CalculateMemberBalance(app, planRecord.Id, session.UserID)
		if err == nil && balance >= 0 {
			existingMembership.Set("date_ended", time.Now())
			existingMembership.Set("leave_requested", false)
			if err := app.Dao().SaveRecord(existingMembership); err != nil {
				return err
			}
		} else {
			existingMembership.Set("leave_requested", true)
			if err := app.Dao().SaveRecord(existingMembership); err != nil {
				return err
			}
		}

		return c.Redirect(http.StatusSeeOther, "/family-plans")
	}
}
