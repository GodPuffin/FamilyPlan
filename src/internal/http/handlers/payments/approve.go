package payments

import (
	"errors"
	"net/http"
	"time"

	"familyplan/src/internal/billing"
	"familyplan/src/internal/planutil"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/daos"
)

// HandleApprovePayment approves a pending payment.
func HandleApprovePayment(app *pocketbase.PocketBase) echo.HandlerFunc {
	return func(c echo.Context) error {
		session, err := sessionOrRedirect(c)
		if err != nil {
			return err
		}
		joinCode := c.PathParam("join_code")
		paymentID := c.FormValue("payment_id")

		planRecord, err := planutil.FindPlanByJoinCode(app, joinCode)
		if err != nil {
			return err
		}
		if planRecord == nil {
			return c.Redirect(http.StatusSeeOther, "/family-plans")
		}

		if !planutil.IsOwner(planRecord, session.UserID) {
			return c.Redirect(http.StatusSeeOther, "/"+joinCode)
		}

		paymentsCollection, err := app.Dao().FindCollectionByNameOrId("payments")
		if err != nil {
			return err
		}

		paymentNotApprovable := errors.New("payment is not approvable")
		err = app.Dao().RunInTransaction(func(txDao *daos.Dao) error {
			payment, err := txDao.FindRecordById(paymentsCollection.Id, paymentID)
			if err != nil || payment == nil {
				return paymentNotApprovable
			}

			if payment.GetString("plan_id") != planRecord.Id || payment.GetString("status") != "pending" {
				return paymentNotApprovable
			}

			payment.Set("status", "approved")
			if err := txDao.SaveRecord(payment); err != nil {
				return err
			}

			return billing.EndMembershipIfSettledWithDao(txDao, planRecord.Id, payment.GetString("user_id"), time.Now())
		})
		if err != nil {
			if errors.Is(err, paymentNotApprovable) {
				return c.Redirect(http.StatusSeeOther, "/"+joinCode)
			}
			return err
		}

		return c.Redirect(http.StatusSeeOther, "/"+joinCode)
	}
}
