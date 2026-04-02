package plans

import (
	"errors"
	"net/http"

	"familyplan/src/internal/money"
	"familyplan/src/internal/support/random"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/daos"
	pbmodels "github.com/pocketbase/pocketbase/models"
)

// HandleCreateFamilyPlan creates a new family plan and owner membership.
func HandleCreateFamilyPlan(app *pocketbase.PocketBase) echo.HandlerFunc {
	return func(c echo.Context) error {
		session, err := sessionOrRedirect(c)
		if err != nil {
			return err
		}

		name := c.FormValue("name")
		description := c.FormValue("description")
		costStr := c.FormValue("cost")
		individualCostStr := c.FormValue("individual_cost")

		if name == "" || costStr == "" || individualCostStr == "" {
			return c.Redirect(http.StatusSeeOther, "/family-plans")
		}

		cost, err := money.ParseAmount(costStr)
		if err != nil {
			return c.Redirect(http.StatusSeeOther, "/family-plans")
		}

		individualCost, err := money.ParseAmount(individualCostStr)
		if err != nil {
			return c.Redirect(http.StatusSeeOther, "/family-plans")
		}

		joinCode, err := random.GenerateJoinCode(6)
		if err != nil {
			return errors.New("failed to generate join code")
		}

		err = app.Dao().RunInTransaction(func(txDao *daos.Dao) error {
			plansCollection, err := txDao.FindCollectionByNameOrId("family_plans")
			if err != nil {
				return err
			}

			newPlan := pbmodels.NewRecord(plansCollection)
			newPlan.Set("name", name)
			newPlan.Set("description", description)
			newPlan.Set("cost", cost)
			newPlan.Set("individual_cost", individualCost)
			newPlan.Set("owner", []string{session.UserID})
			newPlan.Set("join_code", joinCode)

			if err := txDao.SaveRecord(newPlan); err != nil {
				return err
			}

			membershipsCollection, err := txDao.FindCollectionByNameOrId("memberships")
			if err != nil {
				return err
			}

			newMembership := pbmodels.NewRecord(membershipsCollection)
			newMembership.Set("plan_id", newPlan.Id)
			newMembership.Set("user_id", session.UserID)
			newMembership.Set("is_artificial", false)

			return txDao.SaveRecord(newMembership)
		})
		if err != nil {
			return err
		}

		return c.Redirect(http.StatusSeeOther, "/family-plans")
	}
}
