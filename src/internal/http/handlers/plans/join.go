package plans

import (
	"net/http"

	"familyplan/src/internal/planutil"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
	pbmodels "github.com/pocketbase/pocketbase/models"
)

// HandleJoinPlan creates a join request for an existing plan.
func HandleJoinPlan(app *pocketbase.PocketBase) echo.HandlerFunc {
	return func(c echo.Context) error {
		session, err := sessionOrRedirect(c)
		if err != nil {
			return err
		}
		joinCode := c.FormValue("join_code")
		if joinCode == "" {
			return c.Redirect(http.StatusSeeOther, "/family-plans")
		}

		planRecord, err := planutil.FindPlanByJoinCode(app, joinCode)
		if err != nil || planRecord == nil {
			return c.Redirect(http.StatusSeeOther, "/family-plans")
		}

		if planutil.IsOwner(planRecord, session.UserID) {
			return redirectToPlan(c, joinCode)
		}

		existingMembership, _ := planutil.FindMembership(app, planRecord.Id, session.UserID)
		if existingMembership != nil {
			return redirectToPlan(c, joinCode)
		}

		existingRequest, _ := planutil.FindJoinRequest(app, planRecord.Id, session.UserID)
		if existingRequest == nil {
			joinRequestsCollection, err := app.Dao().FindCollectionByNameOrId("join_requests")
			if err != nil {
				return err
			}

			newRequest := pbmodels.NewRecord(joinRequestsCollection)
			newRequest.Set("plan_id", planRecord.Id)
			newRequest.Set("user_id", session.UserID)
			if err := app.Dao().SaveRecord(newRequest); err != nil {
				return err
			}
		}

		return redirectToPlan(c, joinCode)
	}
}
