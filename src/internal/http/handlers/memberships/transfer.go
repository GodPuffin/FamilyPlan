package memberships

import (
	"net/http"

	"familyplan/src/internal/domain"
	"familyplan/src/internal/planutil"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/daos"
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

		artificialMembershipFilter, err := planutil.BuildEqualsFilter(
			planutil.FilterTerm{Field: "plan_id", Value: planRecord.Id},
			planutil.FilterTerm{Field: "user_id", Value: artificialMemberID},
			planutil.FilterTerm{Field: "is_artificial", Value: "true", Literal: true},
		)
		if err != nil {
			return err
		}

		artificialMembership, err := app.Dao().FindFirstRecordByFilter(
			membershipsCollection.Id,
			artificialMembershipFilter,
		)
		if err != nil || artificialMembership == nil {
			return c.Redirect(http.StatusSeeOther, "/"+joinCode)
		}

		request, err := planutil.FindJoinRequest(app, planRecord.Id, realUserID)
		if err != nil || request == nil {
			return c.Redirect(http.StatusSeeOther, "/"+joinCode)
		}

		err = app.Dao().RunInTransaction(func(txDao *daos.Dao) error {
			paymentsCollection, err := txDao.FindCollectionByNameOrId("payments")
			if err != nil {
				return err
			}

			artificialPaymentsFilter, err := planutil.BuildEqualsFilter(
				planutil.FilterTerm{Field: "plan_id", Value: planRecord.Id},
				planutil.FilterTerm{Field: "user_id", Value: artificialMemberID},
			)
			if err != nil {
				return err
			}

			artificialPayments, err := txDao.FindRecordsByFilter(
				paymentsCollection.Id,
				artificialPaymentsFilter,
				"",
				-1,
				0,
			)
			if err != nil {
				return err
			}

			for _, payment := range artificialPayments {
				payment.Set("user_id", realUserID)
				if err := txDao.SaveRecord(payment); err != nil {
					return err
				}
			}

			if err := txDao.DeleteRecord(artificialMembership); err != nil {
				return err
			}

			newMembership := pbmodels.NewRecord(membershipsCollection)
			newMembership.Set("plan_id", planRecord.Id)
			newMembership.Set("user_id", realUserID)
			newMembership.Set("is_artificial", false)
			newMembership.Set("created", artificialMembership.GetDateTime("created"))
			if err := txDao.SaveRecord(newMembership); err != nil {
				return err
			}

			return txDao.DeleteRecord(request)
		})
		if err != nil {
			return err
		}

		return c.Redirect(http.StatusSeeOther, "/"+joinCode)
	}
}
