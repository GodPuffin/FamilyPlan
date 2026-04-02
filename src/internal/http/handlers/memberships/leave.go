package memberships

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

// HandleLeavePlan either leaves immediately or marks leave_requested.
func HandleLeavePlan(app *pocketbase.PocketBase) echo.HandlerFunc {
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

		if planutil.IsOwner(planRecord, session.UserID) {
			return c.Redirect(http.StatusSeeOther, "/"+joinCode)
		}

		membershipNotFound := errors.New("membership not found")
		err = app.Dao().RunInTransaction(func(txDao *daos.Dao) error {
			existingMembership, err := planutil.FindMembershipWithDao(txDao, planRecord.Id, session.UserID)
			if err != nil {
				return err
			}
			if existingMembership == nil {
				return membershipNotFound
			}

			balance, err := billing.CalculateMemberBalanceWithDao(txDao, planRecord.Id, session.UserID)
			if err != nil {
				return err
			}

			if balance >= 0 {
				existingMembership.Set("date_ended", time.Now())
				existingMembership.Set("leave_requested", false)
			} else {
				existingMembership.Set("leave_requested", true)
			}

			return txDao.SaveRecord(existingMembership)
		})
		if err != nil {
			if errors.Is(err, membershipNotFound) {
				return c.Redirect(http.StatusSeeOther, "/family-plans")
			}
			return err
		}

		return c.Redirect(http.StatusSeeOther, "/family-plans")
	}
}
