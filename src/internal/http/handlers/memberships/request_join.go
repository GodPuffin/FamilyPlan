package memberships

import (
	"net/http"

	"familyplan/src/internal/planutil"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
	pbmodels "github.com/pocketbase/pocketbase/models"
)

// HandleRequestJoin creates a join request for the current user.
func HandleRequestJoin(app *pocketbase.PocketBase) echo.HandlerFunc {
	return func(c echo.Context) error {
		session, err := sessionOrRedirect(c)
		if err != nil {
			return err
		}
		joinCode := c.PathParam("join_code")

		planRecord, err := planutil.FindPlanByJoinCode(app, joinCode)
		if err != nil {
			return err
		}
		if planRecord == nil {
			return c.Redirect(http.StatusSeeOther, "/family-plans")
		}

		existingRequest, err := planutil.FindJoinRequest(app, planRecord.Id, session.UserID)
		if err != nil {
			return err
		}
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

		return c.Redirect(http.StatusSeeOther, "/"+joinCode)
	}
}
