package plans

import (
	"fmt"
	"net/http"
	"strings"

	"familyplan/src/internal/money"
	"familyplan/src/internal/planutil"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/daos"
)

// HandleDeletePlan deletes a plan and its related records.
func HandleDeletePlan(app *pocketbase.PocketBase) echo.HandlerFunc {
	return func(c echo.Context) error {
		session, err := sessionOrRedirect(c)
		if err != nil {
			return err
		}
		joinCode := c.PathParam("join_code")

		planRecord, err := planutil.FindPlanByJoinCode(app, joinCode)
		if err != nil || planRecord == nil {
			return c.Redirect(http.StatusSeeOther, "/family-plans")
		}

		if !planutil.IsOwner(planRecord, session.UserID) {
			return redirectToPlan(c, joinCode)
		}

		err = app.Dao().RunInTransaction(func(txDao *daos.Dao) error {
			membershipsCollection, err := txDao.FindCollectionByNameOrId("memberships")
			if err != nil {
				return err
			}

			membershipFilter, err := planutil.BuildEqualsFilter(
				planutil.FilterTerm{Field: "plan_id", Value: planRecord.Id},
			)
			if err != nil {
				return err
			}

			memberships, err := txDao.FindRecordsByFilter(
				membershipsCollection.Id,
				membershipFilter.Expression,
				"",
				-1,
				0,
				membershipFilter.Params,
			)
			if err == nil {
				for _, membership := range memberships {
					if err := txDao.DeleteRecord(membership); err != nil {
						return err
					}
				}
			}

			joinRequestsCollection, err := txDao.FindCollectionByNameOrId("join_requests")
			if err != nil {
				return err
			}

			joinRequestFilter, err := planutil.BuildEqualsFilter(
				planutil.FilterTerm{Field: "plan_id", Value: planRecord.Id},
			)
			if err != nil {
				return err
			}

			joinRequests, err := txDao.FindRecordsByFilter(
				joinRequestsCollection.Id,
				joinRequestFilter.Expression,
				"",
				-1,
				0,
				joinRequestFilter.Params,
			)
			if err == nil {
				for _, request := range joinRequests {
					if err := txDao.DeleteRecord(request); err != nil {
						return err
					}
				}
			}

			paymentsCollection, err := txDao.FindCollectionByNameOrId("payments")
			if err != nil {
				return err
			}

			paymentsFilter, err := planutil.BuildEqualsFilter(
				planutil.FilterTerm{Field: "plan_id", Value: planRecord.Id},
			)
			if err != nil {
				return err
			}

			payments, err := txDao.FindRecordsByFilter(
				paymentsCollection.Id,
				paymentsFilter.Expression,
				"",
				-1,
				0,
				paymentsFilter.Params,
			)
			if err == nil {
				for _, payment := range payments {
					if err := txDao.DeleteRecord(payment); err != nil {
						return err
					}
				}
			}

			return txDao.DeleteRecord(planRecord)
		})
		if err != nil {
			return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/%s?error=Failed+to+delete+plan", joinCode))
		}

		return c.Redirect(http.StatusSeeOther, "/family-plans")
	}
}

// HandleUpdatePlan updates editable plan fields.
func HandleUpdatePlan(app *pocketbase.PocketBase) echo.HandlerFunc {
	return func(c echo.Context) error {
		session, err := sessionOrRedirect(c)
		if err != nil {
			return err
		}
		joinCode := c.PathParam("join_code")

		planRecord, err := planutil.FindPlanByJoinCode(app, joinCode)
		if err != nil || planRecord == nil {
			return c.Redirect(http.StatusSeeOther, "/family-plans")
		}

		if !planutil.IsOwner(planRecord, session.UserID) {
			return c.Redirect(http.StatusSeeOther, "/family-plans")
		}

		name := strings.TrimSpace(c.FormValue("name"))
		description := c.FormValue("description")
		costStr := c.FormValue("cost")
		individualCostStr := c.FormValue("individual_cost")

		if name == "" {
			return redirectToPlan(c, joinCode)
		}

		cost, err := money.ParseAmount(costStr)
		if err != nil {
			cost = money.Normalize(planRecord.GetFloat("cost"))
		}

		individualCost, err := money.ParseAmount(individualCostStr)
		if err != nil {
			individualCost = money.Normalize(planRecord.GetFloat("individual_cost"))
		}

		planRecord.Set("name", name)
		planRecord.Set("description", description)
		planRecord.Set("cost", cost)
		planRecord.Set("individual_cost", individualCost)

		if err := app.Dao().SaveRecord(planRecord); err != nil {
			return err
		}

		return redirectToPlan(c, joinCode)
	}
}
