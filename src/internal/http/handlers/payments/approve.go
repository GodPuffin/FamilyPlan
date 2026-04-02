package payments

import (
	"net/http"
	"time"

	"familyplan/src/internal/billing"
	"familyplan/src/internal/domain"
	"familyplan/src/internal/planutil"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
)

// HandleApprovePayment approves a pending payment.
func HandleApprovePayment(app *pocketbase.PocketBase) echo.HandlerFunc {
	return func(c echo.Context) error {
		session := c.Get("session").(domain.SessionData)
		joinCode := c.PathParam("join_code")

		if err := c.Request().ParseForm(); err != nil {
			return err
		}
		paymentID := c.FormValue("payment_id")

		planRecord, err := planutil.FindPlanByJoinCode(app, joinCode)
		if err != nil || planRecord == nil {
			return c.Redirect(http.StatusSeeOther, "/family-plans")
		}

		if !planutil.IsOwner(planRecord, session.UserID) {
			return c.Redirect(http.StatusSeeOther, "/"+joinCode)
		}

		paymentsCollection, err := app.Dao().FindCollectionByNameOrId("payments")
		if err != nil {
			return err
		}

		payment, err := app.Dao().FindRecordById(paymentsCollection.Id, paymentID)
		if err != nil || payment == nil {
			return c.Redirect(http.StatusSeeOther, "/"+joinCode)
		}

		if payment.GetString("plan_id") != planRecord.Id || payment.GetString("status") != "pending" {
			return c.Redirect(http.StatusSeeOther, "/"+joinCode)
		}

		payment.Set("status", "approved")
		if err := app.Dao().SaveRecord(payment); err != nil {
			return err
		}

		if err := billing.EndMembershipIfSettled(app, planRecord.Id, payment.GetString("user_id"), time.Now()); err != nil {
			return err
		}

		return c.Redirect(http.StatusSeeOther, "/"+joinCode)
	}
}
