package payments

import (
	"net/http"
	"strconv"
	"time"

	"familyplan/src/internal/domain"
	"familyplan/src/internal/planutil"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
	pbmodels "github.com/pocketbase/pocketbase/models"
)

// HandleClaimPayment submits a pending payment claim.
func HandleClaimPayment(app *pocketbase.PocketBase) echo.HandlerFunc {
	return func(c echo.Context) error {
		session := c.Get("session").(domain.SessionData)
		joinCode := c.PathParam("join_code")

		planRecord, err := planutil.FindPlanByJoinCode(app, joinCode)
		if err != nil || planRecord == nil {
			return c.Redirect(http.StatusSeeOther, "/family-plans")
		}

		existingMembership, _ := planutil.FindMembership(app, planRecord.Id, session.UserID)
		if !planutil.IsOwner(planRecord, session.UserID) && existingMembership == nil {
			return c.Redirect(http.StatusSeeOther, "/family-plans")
		}

		if err := c.Request().ParseForm(); err != nil {
			return err
		}

		amount, err := strconv.ParseFloat(c.FormValue("amount"), 64)
		if err != nil || amount <= 0 {
			return c.Redirect(http.StatusSeeOther, "/"+joinCode)
		}

		paymentsCollection, err := app.Dao().FindCollectionByNameOrId("payments")
		if err != nil {
			return err
		}

		payment := pbmodels.NewRecord(paymentsCollection)
		payment.Set("plan_id", planRecord.Id)
		payment.Set("user_id", session.UserID)
		payment.Set("amount", amount)
		payment.Set("date", time.Now())
		payment.Set("status", "pending")
		payment.Set("notes", c.FormValue("notes"))

		if forMonth := parseForMonth(c.FormValue("for_month")); forMonth != "" {
			payment.Set("for_month", forMonth)
		}

		if err := app.Dao().SaveRecord(payment); err != nil {
			return err
		}

		return c.Redirect(http.StatusSeeOther, "/"+joinCode)
	}
}
