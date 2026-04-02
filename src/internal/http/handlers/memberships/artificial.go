package memberships

import (
	"net/http"

	"familyplan/src/internal/planutil"
	"familyplan/src/internal/support/random"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
	pbmodels "github.com/pocketbase/pocketbase/models"
)

// HandleAddArtificialMember creates an artificial member record.
func HandleAddArtificialMember(app *pocketbase.PocketBase) echo.HandlerFunc {
	return func(c echo.Context) error {
		session, err := sessionOrRedirect(c)
		if err != nil {
			return err
		}
		joinCode := c.PathParam("join_code")
		memberName := c.FormValue("name")

		if memberName == "" {
			return c.Redirect(http.StatusSeeOther, "/"+joinCode)
		}

		planRecord, err := planutil.FindPlanByJoinCode(app, joinCode)
		if err != nil || planRecord == nil {
			return c.Redirect(http.StatusSeeOther, "/family-plans")
		}

		if !planutil.IsOwner(planRecord, session.UserID) {
			return c.Redirect(http.StatusSeeOther, "/"+joinCode)
		}

		membershipsCollection, err := app.Dao().FindCollectionByNameOrId("memberships")
		if err != nil {
			return err
		}

		artificialUserID, err := random.GenerateUUID()
		if err != nil {
			return err
		}

		newMembership := pbmodels.NewRecord(membershipsCollection)
		newMembership.Set("plan_id", planRecord.Id)
		newMembership.Set("user_id", artificialUserID)
		newMembership.Set("is_artificial", true)
		newMembership.Set("name", memberName)
		if err := app.Dao().SaveRecord(newMembership); err != nil {
			return err
		}

		return c.Redirect(http.StatusSeeOther, "/"+joinCode)
	}
}
