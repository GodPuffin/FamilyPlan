package plans

import (
	"time"

	"familyplan/src/internal/billing"
	"familyplan/src/internal/money"
	"familyplan/src/internal/planutil"

	"github.com/pocketbase/pocketbase"
	pbmodels "github.com/pocketbase/pocketbase/models"
)

func calculateTotalPayments(app *pocketbase.PocketBase, planID string) float64 {
	paymentsCollection, err := app.Dao().FindCollectionByNameOrId("payments")
	if err != nil {
		return 0
	}

	filter, err := planutil.BuildEqualsFilter(
		planutil.FilterTerm{Field: "plan_id", Value: planID},
		planutil.FilterTerm{Field: "status", Value: "approved"},
	)
	if err != nil {
		return 0
	}

	approvedPayments, err := app.Dao().FindRecordsByFilter(
		paymentsCollection.Id,
		filter.Expression,
		"",
		-1,
		0,
		filter.Params,
	)
	if err != nil {
		return 0
	}

	totalPaymentsCents := int64(0)
	for _, payment := range approvedPayments {
		totalPaymentsCents += money.ToCents(payment.GetFloat("amount"))
	}

	return money.FromCents(totalPaymentsCents)
}

func calculateTotalSavings(app *pocketbase.PocketBase, plan *pbmodels.Record) float64 {
	totalSavingsCents := int64(0)
	individualCostCents := money.ToCents(plan.GetFloat("individual_cost"))
	familyPlanCostCents := money.ToCents(plan.GetFloat("cost"))

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

		monthlySavingsCents := (individualCostCents * int64(memberCount)) - familyPlanCostCents
		if monthlySavingsCents > 0 {
			totalSavingsCents += monthlySavingsCents
		}
	}

	return money.FromCents(totalSavingsCents)
}

func calculatePlanAgeDays(plan *pbmodels.Record) int {
	return int(time.Since(plan.GetDateTime("created").Time()).Hours() / 24)
}
