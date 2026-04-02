package plans

import (
	"errors"
	"net/http"
	"strconv"

	"familyplan/src/internal/domain"
	"familyplan/src/internal/support/random"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/daos"
	pbmodels "github.com/pocketbase/pocketbase/models"
)

// HandleCreateFamilyPlan creates a new family plan and owner membership.
func HandleCreateFamilyPlan(app *pocketbase.PocketBase) echo.HandlerFunc {
	return func(c echo.Context) error {
		session := c.Get("session").(domain.SessionData)

		name := c.FormValue("name")
		description := c.FormValue("description")
		costStr := c.FormValue("cost")
		individualCostStr := c.FormValue("individual_cost")

		if name == "" || costStr == "" || individualCostStr == "" {
			return c.Redirect(http.StatusSeeOther, "/family-plans")
		}

		cost, err := strconv.ParseFloat(costStr, 64)
		if err != nil {
			return c.Redirect(http.StatusSeeOther, "/family-plans")
		}

		individualCost, err := strconv.ParseFloat(individualCostStr, 64)
		if err != nil {
			return c.Redirect(http.StatusSeeOther, "/family-plans")
		}

		joinCode := random.GenerateJoinCode(6)
		if joinCode == "" {
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
