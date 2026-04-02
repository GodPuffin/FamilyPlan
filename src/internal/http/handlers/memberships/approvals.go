package memberships

import (
	"net/http"

	"familyplan/src/internal/domain"
	"familyplan/src/internal/planutil"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
	pbmodels "github.com/pocketbase/pocketbase/models"
)

// HandleApproveRequest approves a pending join request.
func HandleApproveRequest(app *pocketbase.PocketBase) echo.HandlerFunc {
	return func(c echo.Context) error {
		session := c.Get("session").(domain.SessionData)
		joinCode := c.PathParam("join_code")
		userID := c.FormValue("user_id")

		planRecord, err := planutil.FindPlanByJoinCode(app, joinCode)
		if err != nil || planRecord == nil {
			return c.Redirect(http.StatusSeeOther, "/family-plans")
		}

		if !planutil.IsOwner(planRecord, session.UserID) {
			return c.Redirect(http.StatusSeeOther, "/"+joinCode)
		}

		request, err := planutil.FindJoinRequest(app, planRecord.Id, userID)
		if err != nil || request == nil {
			return c.Redirect(http.StatusSeeOther, "/"+joinCode)
		}

		existingMembership, _ := planutil.FindMembership(app, planRecord.Id, userID)
		if existingMembership == nil {
			membershipsCollection, err := app.Dao().FindCollectionByNameOrId("memberships")
			if err != nil {
				return err
			}

			newMembership := pbmodels.NewRecord(membershipsCollection)
			newMembership.Set("plan_id", planRecord.Id)
			newMembership.Set("user_id", userID)
			newMembership.Set("is_artificial", false)
			if err := app.Dao().SaveRecord(newMembership); err != nil {
				return err
			}
		}

		if err := app.Dao().DeleteRecord(request); err != nil {
			return err
		}

		return c.Redirect(http.StatusSeeOther, "/"+joinCode)
	}
}

// HandleDenyRequest denies a pending join request.
func HandleDenyRequest(app *pocketbase.PocketBase) echo.HandlerFunc {
	return func(c echo.Context) error {
		session := c.Get("session").(domain.SessionData)
		joinCode := c.PathParam("join_code")
		userID := c.FormValue("user_id")

		planRecord, err := planutil.FindPlanByJoinCode(app, joinCode)
		if err != nil || planRecord == nil {
			return c.Redirect(http.StatusSeeOther, "/family-plans")
		}

		if !planutil.IsOwner(planRecord, session.UserID) {
			return c.Redirect(http.StatusSeeOther, "/"+joinCode)
		}

		request, err := planutil.FindJoinRequest(app, planRecord.Id, userID)
		if err != nil || request == nil {
			return c.Redirect(http.StatusSeeOther, "/"+joinCode)
		}

		if err := app.Dao().DeleteRecord(request); err != nil {
			return err
		}

		return c.Redirect(http.StatusSeeOther, "/"+joinCode)
	}
}
