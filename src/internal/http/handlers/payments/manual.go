package payments

import (
	"net/http"
	"time"

	"familyplan/src/internal/billing"
	"familyplan/src/internal/money"
	"familyplan/src/internal/planutil"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/daos"
	pbmodels "github.com/pocketbase/pocketbase/models"
)

// HandleAddManualPayment adds an owner-entered approved payment.
func HandleAddManualPayment(app *pocketbase.PocketBase) echo.HandlerFunc {
	return func(c echo.Context) error {
		session, err := sessionOrRedirect(c)
		if err != nil {
			return err
		}
		joinCode := c.PathParam("join_code")
		userID := c.FormValue("user_id")

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

		membership, err := planutil.FindMembership(app, planRecord.Id, userID)
		if err != nil {
			return err
		}
		if membership == nil {
			return c.Redirect(http.StatusSeeOther, "/"+joinCode)
		}

		amount, err := money.ParseAmount(c.FormValue("amount"))
		if err != nil || amount <= 0 {
			return c.Redirect(http.StatusSeeOther, "/"+joinCode)
		}

		paymentsCollection, err := app.Dao().FindCollectionByNameOrId("payments")
		if err != nil {
			return err
		}

		payment := pbmodels.NewRecord(paymentsCollection)
		payment.Set("plan_id", planRecord.Id)
		payment.Set("user_id", userID)
		payment.Set("amount", amount)
		payment.Set("date", time.Now())
		payment.Set("status", "approved")
		notes, err := normalizeNotes(c.FormValue("notes"))
		if err != nil {
			return c.Redirect(http.StatusSeeOther, "/"+joinCode)
		}
		payment.Set("notes", notes)

		if forMonth := parseForMonth(c.FormValue("for_month")); forMonth != "" {
			payment.Set("for_month", forMonth)
		}

		err = app.Dao().RunInTransaction(func(txDao *daos.Dao) error {
			if err := txDao.SaveRecord(payment); err != nil {
				return err
			}

			return billing.EndMembershipIfSettledWithDao(txDao, planRecord.Id, userID, time.Now())
		})
		if err != nil {
			return err
		}

		return c.Redirect(http.StatusSeeOther, "/"+joinCode)
	}
}
