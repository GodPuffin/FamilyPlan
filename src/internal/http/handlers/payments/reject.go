package payments

import (
	"net/http"

	"familyplan/src/internal/domain"
	"familyplan/src/internal/planutil"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
)

// HandleRejectPayment rejects a pending payment.
func HandleRejectPayment(app *pocketbase.PocketBase) echo.HandlerFunc {
	return func(c echo.Context) error {
		session := c.Get("session").(domain.SessionData)
		joinCode := c.PathParam("join_code")

		planRecord, err := planutil.FindPlanByJoinCode(app, joinCode)
		if err != nil || planRecord == nil {
			return c.Redirect(http.StatusSeeOther, "/family-plans")
		}

		if !planutil.IsOwner(planRecord, session.UserID) {
			return c.Redirect(http.StatusSeeOther, "/"+joinCode)
		}

		if err := c.Request().ParseForm(); err != nil {
			return err
		}

		paymentID := c.FormValue("payment_id")
		if paymentID == "" {
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

		if payment.GetString("plan_id") != planRecord.Id {
			return c.Redirect(http.StatusSeeOther, "/"+joinCode)
		}

		payment.Set("status", "rejected")
		if err := app.Dao().SaveRecord(payment); err != nil {
			return err
		}

		return c.Redirect(http.StatusSeeOther, "/"+joinCode)
	}
}
