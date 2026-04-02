package plans

import (
	"fmt"
	"net/http"
	"strconv"

	"familyplan/src/internal/domain"
	"familyplan/src/internal/planutil"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/daos"
)

// HandleDeletePlan deletes a plan and its related records.
func HandleDeletePlan(app *pocketbase.PocketBase) echo.HandlerFunc {
	return func(c echo.Context) error {
		session := c.Get("session").(domain.SessionData)
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
				membershipFilter,
				"",
				-1,
				0,
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
				joinRequestFilter,
				"",
				-1,
				0,
			)
			if err == nil {
				for _, request := range joinRequests {
					if err := txDao.DeleteRecord(request); err != nil {
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
		session := c.Get("session").(domain.SessionData)
		joinCode := c.PathParam("join_code")

		planRecord, err := planutil.FindPlanByJoinCode(app, joinCode)
		if err != nil || planRecord == nil {
			return c.Redirect(http.StatusSeeOther, "/family-plans")
		}

		if !planutil.IsOwner(planRecord, session.UserID) {
			return c.Redirect(http.StatusSeeOther, "/family-plans")
		}

		name := c.FormValue("name")
		description := c.FormValue("description")
		costStr := c.FormValue("cost")
		individualCostStr := c.FormValue("individual_cost")

		cost, err := strconv.ParseFloat(costStr, 64)
		if err != nil {
			cost = planRecord.GetFloat("cost")
		}

		individualCost, err := strconv.ParseFloat(individualCostStr, 64)
		if err != nil {
			individualCost = planRecord.GetFloat("individual_cost")
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
