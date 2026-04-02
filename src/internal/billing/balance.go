package billing

import (
	"fmt"
	"time"

	"familyplan/src/internal/planutil"

	"github.com/pocketbase/pocketbase"
)

// CalculateMemberBalance calculates the current balance for a member.
func CalculateMemberBalance(app *pocketbase.PocketBase, planID, userID string) (float64, error) {
	plansCollection, err := app.Dao().FindCollectionByNameOrId("family_plans")
	if err != nil {
		return 0, err
	}

	plan, err := app.Dao().FindRecordById(plansCollection.Id, planID)
	if err != nil {
		return 0, err
	}

	monthlyCost := plan.GetFloat("cost")

	membership, err := planutil.FindMembership(app, planID, userID)
	if err != nil || membership == nil {
		return 0, fmt.Errorf("membership not found")
	}

	paymentsCollection, err := app.Dao().FindCollectionByNameOrId("payments")
	if err != nil {
		return 0, err
	}

	userPayments, err := app.Dao().FindRecordsByFilter(
		paymentsCollection.Id,
		fmt.Sprintf("plan_id = '%s' && user_id = '%s' && status = 'approved'", planID, userID),
		"",
		100,
		0,
	)
	if err != nil {
		return 0, err
	}

	totalPaid := 0.0
	paymentsByMonth := make(map[string]float64)

	for _, payment := range userPayments {
		amount := payment.GetFloat("amount")
		totalPaid += amount

		forMonth := payment.GetDateTime("for_month")
		if !forMonth.IsZero() {
			monthKey := forMonth.Time().Format("2006-01")
			paymentsByMonth[monthKey] += amount
		}
	}

	membershipStartDate := membership.GetDateTime("created").Time()
	currentDate := time.Now()

	membershipEndDate := membership.GetDateTime("date_ended")
	if !membershipEndDate.IsZero() {
		currentDate = membershipEndDate.Time()
	}

	startMonth := time.Date(
		membershipStartDate.Year(),
		membershipStartDate.Month(),
		1,
		0,
		0,
		0,
		0,
		membershipStartDate.Location(),
	)

	var endMonth time.Time
	if !membershipEndDate.IsZero() {
		endMonth = time.Date(
			membershipEndDate.Time().Year(),
			membershipEndDate.Time().Month(),
			1,
			0,
			0,
			0,
			0,
			membershipEndDate.Time().Location(),
		)
	} else {
		endMonth = time.Date(
			currentDate.Year(),
			currentDate.Month(),
			1,
			0,
			0,
			0,
			0,
			currentDate.Location(),
		)
	}

	amountDue := 0.0
	currentMonth := startMonth

	for !currentMonth.After(endMonth) {
		monthKey := currentMonth.Format("2006-01")

		activeMemberships, err := GetActiveMembershipsForMonth(app, planID, currentMonth)
		if err == nil {
			memberCount := 0
			ownerIncluded := false
			ownerID := planutil.OwnerID(plan)

			for _, activeMembership := range activeMemberships {
				if activeMembership.GetString("user_id") == ownerID {
					ownerIncluded = true
				}
				memberCount++
			}

			if !ownerIncluded {
				memberCount++
			}

			if memberCount > 0 {
				monthlyShare := monthlyCost / float64(memberCount)
				amountDue += monthlyShare
			}
		}

		if paidAmount, exists := paymentsByMonth[monthKey]; exists {
			amountDue -= paidAmount
			totalPaid -= paidAmount
		}

		currentMonth = time.Date(
			currentMonth.Year(),
			currentMonth.Month()+1,
			1,
			0,
			0,
			0,
			0,
			currentMonth.Location(),
		)
	}

	return totalPaid - amountDue, nil
}

// EndMembershipIfSettled ends a leave-requested membership once its balance is settled.
func EndMembershipIfSettled(app *pocketbase.PocketBase, planID, userID string, endedAt time.Time) error {
	membership, err := planutil.FindMembership(app, planID, userID)
	if err != nil || membership == nil {
		return err
	}

	if !membership.GetBool("leave_requested") {
		return nil
	}

	balance, err := CalculateMemberBalance(app, planID, userID)
	if err != nil || balance < 0 {
		return err
	}

	membership.Set("date_ended", endedAt)
	membership.Set("leave_requested", false)
	return app.Dao().SaveRecord(membership)
}
