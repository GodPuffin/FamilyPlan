package plans

import (
	"fmt"
	"net/http"

	"familyplan/src/internal/domain"

	"github.com/labstack/echo/v5"
	pbmodels "github.com/pocketbase/pocketbase/models"
)

func buildFamilyPlan(record *pbmodels.Record, membersCount int, balance float64) domain.FamilyPlan {
	return domain.FamilyPlan{
		ID:             record.Id,
		Name:           record.GetString("name"),
		Description:    record.GetString("description"),
		Cost:           record.GetFloat("cost"),
		IndividualCost: record.GetFloat("individual_cost"),
		Owner:          ownerID(record),
		JoinCode:       record.GetString("join_code"),
		CreatedAt:      record.GetDateTime("created").String(),
		MembersCount:   membersCount,
		Balance:        balance,
	}
}

func ownerID(plan *pbmodels.Record) string {
	ownerIDs := plan.GetStringSlice("owner")
	if len(ownerIDs) == 0 {
		return ""
	}

	return ownerIDs[0]
}

func activeMembershipCount(memberships []*pbmodels.Record) int {
	count := 0
	for _, membership := range memberships {
		dateEnded := membership.GetDateTime("date_ended")
		if !membership.GetBool("leave_requested") && dateEnded.IsZero() {
			count++
		}
	}

	return count
}

func buildPayment(record *pbmodels.Record, username, name string) domain.Payment {
	return domain.Payment{
		ID:       record.Id,
		PlanID:   record.GetString("plan_id"),
		UserID:   record.GetString("user_id"),
		Amount:   record.GetFloat("amount"),
		Date:     record.GetDateTime("date").String()[:10],
		Status:   record.GetString("status"),
		Notes:    record.GetString("notes"),
		ForMonth: formatForMonth(record),
		Username: username,
		Name:     name,
	}
}

func formatForMonth(record *pbmodels.Record) string {
	forMonth := record.GetDateTime("for_month")
	if forMonth.IsZero() {
		return ""
	}

	return forMonth.String()[:7]
}

func redirectToPlan(c echo.Context, joinCode string) error {
	return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/%s", joinCode))
}
