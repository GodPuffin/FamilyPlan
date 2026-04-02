package plans

import (
	"net/http"
	"strconv"

	"familyplan/src/internal/domain"
	"familyplan/src/internal/support/random"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
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

		plansCollection, err := app.Dao().FindCollectionByNameOrId("family_plans")
		if err != nil {
			return err
		}

		newPlan := pbmodels.NewRecord(plansCollection)
		newPlan.Set("name", name)
		newPlan.Set("description", description)
		newPlan.Set("cost", cost)
		newPlan.Set("individual_cost", individualCost)
		newPlan.Set("owner", []string{session.UserID})
		newPlan.Set("join_code", random.GenerateJoinCode(6))

		if err := app.Dao().SaveRecord(newPlan); err != nil {
			return err
		}

		membershipsCollection, err := app.Dao().FindCollectionByNameOrId("memberships")
		if err != nil {
			return err
		}

		newMembership := pbmodels.NewRecord(membershipsCollection)
		newMembership.Set("plan_id", newPlan.Id)
		newMembership.Set("user_id", session.UserID)
		newMembership.Set("is_artificial", false)
		_ = app.Dao().SaveRecord(newMembership)

		return c.Redirect(http.StatusSeeOther, "/family-plans")
	}
}
