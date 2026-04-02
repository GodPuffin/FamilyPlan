package memberships

import (
	"fmt"
	"net/http"

	"familyplan/src/internal/domain"
	"familyplan/src/internal/planutil"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
	pbmodels "github.com/pocketbase/pocketbase/models"
)

// HandleTransferMembership converts an artificial member into a real user membership.
func HandleTransferMembership(app *pocketbase.PocketBase) echo.HandlerFunc {
	return func(c echo.Context) error {
		session := c.Get("session").(domain.SessionData)
		joinCode := c.PathParam("join_code")
		realUserID := c.FormValue("user_id")
		artificialMemberID := c.FormValue("artificial_member_id")

		if realUserID == "" || artificialMemberID == "" {
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

		artificialMembership, err := app.Dao().FindFirstRecordByFilter(
			membershipsCollection.Id,
			fmt.Sprintf("plan_id = '%s' && user_id = '%s' && is_artificial = true", planRecord.Id, artificialMemberID),
		)
		if err != nil || artificialMembership == nil {
			return c.Redirect(http.StatusSeeOther, "/"+joinCode)
		}

		request, err := planutil.FindJoinRequest(app, planRecord.Id, realUserID)
		if err != nil || request == nil {
			return c.Redirect(http.StatusSeeOther, "/"+joinCode)
		}

		paymentsCollection, err := app.Dao().FindCollectionByNameOrId("payments")
		if err != nil {
			return err
		}

		artificialPayments, err := app.Dao().FindRecordsByFilter(
			paymentsCollection.Id,
			fmt.Sprintf("plan_id = '%s' && user_id = '%s'", planRecord.Id, artificialMemberID),
			"",
			100,
			0,
		)
		if err == nil {
			for _, payment := range artificialPayments {
				payment.Set("user_id", realUserID)
				_ = app.Dao().SaveRecord(payment)
			}
		}

		if err := app.Dao().DeleteRecord(artificialMembership); err != nil {
			return err
		}

		newMembership := pbmodels.NewRecord(membershipsCollection)
		newMembership.Set("plan_id", planRecord.Id)
		newMembership.Set("user_id", realUserID)
		newMembership.Set("is_artificial", false)
		newMembership.Set("created", artificialMembership.GetDateTime("created"))
		if err := app.Dao().SaveRecord(newMembership); err != nil {
			return err
		}

		if err := app.Dao().DeleteRecord(request); err != nil {
			return err
		}

		return c.Redirect(http.StatusSeeOther, "/"+joinCode)
	}
}
