package billing

import (
	"fmt"
	"sort"
	"time"

	"familyplan/src/internal/money"
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

	monthlyCostCents := money.ToCents(plan.GetFloat("cost"))

	membership, err := planutil.FindMembership(app, planID, userID)
	if err != nil || membership == nil {
		return 0, fmt.Errorf("membership not found")
	}

	paymentsCollection, err := app.Dao().FindCollectionByNameOrId("payments")
	if err != nil {
		return 0, err
	}

	filter, err := planutil.BuildEqualsFilter(
		planutil.FilterTerm{Field: "plan_id", Value: planID},
		planutil.FilterTerm{Field: "user_id", Value: userID},
		planutil.FilterTerm{Field: "status", Value: "approved"},
	)
	if err != nil {
		return 0, err
	}

	userPayments, err := app.Dao().FindRecordsByFilter(
		paymentsCollection.Id,
		filter.Expression,
		"",
		-1,
		0,
		filter.Params,
	)
	if err != nil {
		return 0, err
	}

	totalPaidCents := int64(0)
	paymentsByMonth := make(map[string]int64)

	for _, payment := range userPayments {
		amountCents := money.ToCents(payment.GetFloat("amount"))
		totalPaidCents += amountCents

		forMonth := payment.GetDateTime("for_month")
		if !forMonth.IsZero() {
			monthKey := forMonth.Time().Format("2006-01")
			paymentsByMonth[monthKey] += amountCents
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

	amountDueCents := int64(0)
	currentMonth := startMonth

	for !currentMonth.After(endMonth) {
		monthKey := currentMonth.Format("2006-01")

		activeMemberships, err := GetActiveMembershipsForMonth(app, planID, currentMonth)
		if err == nil {
			memberIDs := make([]string, 0, len(activeMemberships)+1)
			ownerIncluded := false
			ownerID := planutil.OwnerID(plan)

			for _, activeMembership := range activeMemberships {
				memberID := activeMembership.GetString("user_id")
				if memberID == ownerID {
					ownerIncluded = true
				}
				memberIDs = append(memberIDs, memberID)
			}

			if !ownerIncluded {
				memberIDs = append(memberIDs, ownerID)
			}

			if len(memberIDs) > 0 {
				amountDueCents += memberShareCents(monthlyCostCents, memberIDs, userID)
			}
		}

		if paidAmount, exists := paymentsByMonth[monthKey]; exists {
			amountDueCents -= paidAmount
			totalPaidCents -= paidAmount
		}

		currentMonth = currentMonth.AddDate(0, 1, 0)
	}

	return money.FromCents(totalPaidCents - amountDueCents), nil
}

func memberShareCents(totalCents int64, memberIDs []string, userID string) int64 {
	if len(memberIDs) == 0 {
		return 0
	}

	sortedIDs := append([]string(nil), memberIDs...)
	sort.Strings(sortedIDs)

	baseShare := totalCents / int64(len(sortedIDs))
	remainder := totalCents % int64(len(sortedIDs))

	for i, memberID := range sortedIDs {
		if memberID != userID {
			continue
		}

		if int64(i) < remainder {
			return baseShare + 1
		}

		return baseShare
	}

	return 0
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
