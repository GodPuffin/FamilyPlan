package billing

import (
	"fmt"
	"sort"
	"time"

	"familyplan/src/internal/money"
	"familyplan/src/internal/planutil"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/daos"
)

// CalculateMemberBalance calculates the current balance for a member.
func CalculateMemberBalance(app *pocketbase.PocketBase, planID, userID string) (float64, error) {
	return CalculateMemberBalanceWithDao(app.Dao(), planID, userID)
}

// CalculateMemberBalanceWithDao calculates the current balance using the provided dao.
func CalculateMemberBalanceWithDao(dao *daos.Dao, planID, userID string) (float64, error) {
	plansCollection, err := dao.FindCollectionByNameOrId("family_plans")
	if err != nil {
		return 0, err
	}

	plan, err := dao.FindRecordById(plansCollection.Id, planID)
	if err != nil {
		return 0, err
	}

	monthlyCostCents := money.ToCents(plan.GetFloat("cost"))

	membership, err := planutil.FindMembershipWithDao(dao, planID, userID)
	if err != nil {
		return 0, err
	}
	if membership == nil {
		return 0, fmt.Errorf("membership not found")
	}

	paymentsCollection, err := dao.FindCollectionByNameOrId("payments")
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

	userPayments, err := dao.FindRecordsByFilter(
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
	membershipEndDate := membership.GetDateTime("date_ended")

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
		now := time.Now()
		endMonth = time.Date(
			now.Year(),
			now.Month(),
			1,
			0,
			0,
			0,
			0,
			now.Location(),
		)
	}

	amountDueCents := int64(0)
	currentMonth := startMonth

	for !currentMonth.After(endMonth) {
		monthKey := currentMonth.Format("2006-01")

		activeMemberships, err := getActiveMembershipsForMonth(dao, planID, currentMonth)
		if err != nil {
			return 0, err
		}

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

		if paidAmount, exists := paymentsByMonth[monthKey]; exists {
			totalPaidCents, amountDueCents = applyAttributedPayment(totalPaidCents, amountDueCents, paidAmount)
		}

		currentMonth = currentMonth.AddDate(0, 1, 0)
	}

	return money.FromCents(totalPaidCents - amountDueCents), nil
}

func applyAttributedPayment(totalPaidCents, amountDueCents, paidAmount int64) (int64, int64) {
	// Month-attributed payments settle that month's charge and stop counting as unallocated credit.
	return totalPaidCents - paidAmount, amountDueCents - paidAmount
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
	return EndMembershipIfSettledWithDao(app.Dao(), planID, userID, endedAt)
}

// EndMembershipIfSettledWithDao ends a leave-requested membership using the provided dao.
func EndMembershipIfSettledWithDao(dao *daos.Dao, planID, userID string, endedAt time.Time) error {
	membership, err := planutil.FindMembershipWithDao(dao, planID, userID)
	if err != nil {
		return err
	}
	if membership == nil {
		return fmt.Errorf("membership not found")
	}

	if !membership.GetBool("leave_requested") {
		return nil
	}

	balance, err := CalculateMemberBalanceWithDao(dao, planID, userID)
	if err != nil {
		return err
	}
	if balance < 0 {
		return nil
	}

	membership.Set("date_ended", endedAt)
	membership.Set("leave_requested", false)
	return dao.SaveRecord(membership)
}
