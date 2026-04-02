package plans

import (
	"fmt"
	"time"

	"familyplan/src/internal/billing"

	"github.com/pocketbase/pocketbase"
	pbmodels "github.com/pocketbase/pocketbase/models"
)

func calculateTotalPayments(app *pocketbase.PocketBase, planID string) float64 {
	paymentsCollection, err := app.Dao().FindCollectionByNameOrId("payments")
	if err != nil {
		return 0
	}

	approvedPayments, err := app.Dao().FindRecordsByFilter(
		paymentsCollection.Id,
		fmt.Sprintf("plan_id = '%s' && status = 'approved'", planID),
		"",
		-1,
		0,
	)
	if err != nil {
		return 0
	}

	totalPayments := 0.0
	for _, payment := range approvedPayments {
		totalPayments += payment.GetFloat("amount")
	}

	return totalPayments
}

func calculateTotalSavings(app *pocketbase.PocketBase, plan *pbmodels.Record) float64 {
	totalSavings := 0.0
	individualCost := plan.GetFloat("individual_cost")
	familyPlanCost := plan.GetFloat("cost")

	planCreationTime := plan.GetDateTime("created").Time()
	currentTime := time.Now()
	startDate := time.Date(
		planCreationTime.Year(),
		planCreationTime.Month(),
		1,
		0,
		0,
		0,
		0,
		planCreationTime.Location(),
	)

	for currentDate := startDate; currentDate.Before(currentTime); currentDate = currentDate.AddDate(0, 1, 0) {
		activeMemberships, err := billing.GetActiveMembershipsForMonth(app, plan.Id, currentDate)
		if err != nil {
			continue
		}

		memberCount := len(activeMemberships)
		if memberCount == 0 {
			continue
		}

		monthlySavings := (individualCost * float64(memberCount)) - familyPlanCost
		if monthlySavings > 0 {
			totalSavings += monthlySavings
		}
	}

	return totalSavings
}

func calculatePlanAgeDays(plan *pbmodels.Record) int {
	return int(time.Since(plan.GetDateTime("created").Time()).Hours() / 24)
}
