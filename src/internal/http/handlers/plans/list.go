package plans

import (
	"fmt"

	"familyplan/src/internal/billing"
	"familyplan/src/internal/domain"
	"familyplan/src/internal/view"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
	pbmodels "github.com/pocketbase/pocketbase/models"
)

// HandleFamilyPlansList renders the current user's plans.
func HandleFamilyPlansList(app *pocketbase.PocketBase) echo.HandlerFunc {
	return func(c echo.Context) error {
		session := c.Get("session").(domain.SessionData)

		plansCollection, err := app.Dao().FindCollectionByNameOrId("family_plans")
		if err != nil {
			return err
		}

		ownedPlanRecords, err := app.Dao().FindRecordsByFilter(
			plansCollection.Id,
			fmt.Sprintf("owner ~ '%s'", session.UserID),
			"",
			100,
			0,
		)
		if err != nil {
			return err
		}

		membershipsCollection, err := app.Dao().FindCollectionByNameOrId("memberships")
		if err != nil {
			return err
		}

		memberships, err := app.Dao().FindRecordsByFilter(
			membershipsCollection.Id,
			fmt.Sprintf("user_id = '%s'", session.UserID),
			"",
			100,
			0,
		)
		if err != nil {
			return err
		}

		planMap := make(map[string]*pbmodels.Record)
		membershipMap := make(map[string]*pbmodels.Record)

		for _, record := range ownedPlanRecords {
			planMap[record.Id] = record
		}

		for _, membership := range memberships {
			planID := membership.GetString("plan_id")
			membershipMap[planID] = membership

			if _, exists := planMap[planID]; exists {
				continue
			}

			planRecord, err := app.Dao().FindRecordById(plansCollection.Id, planID)
			if err == nil {
				planMap[planRecord.Id] = planRecord
			}
		}

		plansList := make([]domain.FamilyPlan, 0, len(planMap))
		for _, planRecord := range planMap {
			membershipRecords, err := app.Dao().FindRecordsByFilter(
				membershipsCollection.Id,
				fmt.Sprintf("plan_id = '%s'", planRecord.Id),
				"",
				100,
				0,
			)

			membersCount := 0
			if err == nil {
				membersCount = activeMembershipCount(membershipRecords)
			}

			balance := 0.0
			isOwner := ownerID(planRecord) == session.UserID
			if !isOwner && membershipMap[planRecord.Id] != nil {
				balanceAmount, err := billing.CalculateMemberBalance(app, planRecord.Id, session.UserID)
				if err == nil {
					balance = balanceAmount
				}
			}

			plansList = append(plansList, buildFamilyPlan(planRecord, membersCount, balance))
		}

		return view.RenderPage(c, "family_plans.html", map[string]interface{}{
			"title": "My Family Plans",
			"plans": plansList,
		})
	}
}
