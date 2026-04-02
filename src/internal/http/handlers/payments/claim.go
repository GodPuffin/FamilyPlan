package payments

import (
	"net/http"
	"time"

	"familyplan/src/internal/money"
	"familyplan/src/internal/planutil"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
	pbmodels "github.com/pocketbase/pocketbase/models"
)

// HandleClaimPayment submits a pending payment claim.
func HandleClaimPayment(app *pocketbase.PocketBase) echo.HandlerFunc {
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

		existingMembership, err := planutil.FindMembership(app, planRecord.Id, session.UserID)
		if err != nil {
			return err
		}
		if !planutil.IsOwner(planRecord, session.UserID) && existingMembership == nil {
			return c.Redirect(http.StatusSeeOther, "/family-plans")
		}

		amount, err := money.ParseAmount(c.FormValue("amount"))
		if err != nil || amount <= 0 {
			return c.Redirect(http.StatusSeeOther, "/"+joinCode)
		}

		paymentsCollection, err := app.Dao().FindCollectionByNameOrId("payments")
		if err != nil {
			return err
		}

		notes, err := normalizeNotes(c.FormValue("notes"))
		if err != nil {
			return c.Redirect(http.StatusSeeOther, "/"+joinCode)
		}

		payment := pbmodels.NewRecord(paymentsCollection)
		payment.Set("plan_id", planRecord.Id)
		payment.Set("user_id", session.UserID)
		payment.Set("amount", amount)
		payment.Set("date", time.Now())
		payment.Set("status", "pending")
		payment.Set("notes", notes)

		if forMonth := parseForMonth(c.FormValue("for_month")); forMonth != "" {
			payment.Set("for_month", forMonth)
		}

		if err := app.Dao().SaveRecord(payment); err != nil {
			return err
		}

		return c.Redirect(http.StatusSeeOther, "/"+joinCode)
	}
}
