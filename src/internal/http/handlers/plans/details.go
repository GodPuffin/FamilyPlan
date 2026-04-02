package plans

import (
	"familyplan/src/internal/billing"
	"familyplan/src/internal/domain"
	"familyplan/src/internal/planutil"
	"familyplan/src/internal/view"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
)

// HandlePlanDetails renders the plan detail page.
func HandlePlanDetails(app *pocketbase.PocketBase) echo.HandlerFunc {
	return func(c echo.Context) error {
		session, err := sessionOrRedirect(c)
		if err != nil {
			return err
		}
		joinCode := c.PathParam("join_code")

		planRecord, err := planutil.FindPlanByJoinCode(app, joinCode)
		if err != nil || planRecord == nil {
			return view.RenderPage(c, "plan_details.html", map[string]interface{}{
				"title":     "Plan Not Found",
				"not_found": true,
				"plan":      nil,
			})
		}

		isOwner := planutil.IsOwner(planRecord, session.UserID)
		existingMembership, _ := planutil.FindMembership(app, planRecord.Id, session.UserID)
		isMember := existingMembership != nil || isOwner

		pendingRequest := false
		if !isMember {
			existingRequest, _ := planutil.FindJoinRequest(app, planRecord.Id, session.UserID)
			pendingRequest = existingRequest != nil
		}

		familyPlan := buildFamilyPlan(planRecord, 0, 0)

		members := []domain.Member{}
		totalMembers := 0
		if isMember {
			members, totalMembers, err = loadMembers(app, familyPlan)
			if err != nil {
				return err
			}
		}

		joinRequests := []domain.JoinRequest{}
		if isOwner {
			joinRequests, err = loadJoinRequests(app, planRecord.Id)
			if err != nil {
				return err
			}
		}

		pendingPayments := []domain.Payment{}
		userPayments := []domain.Payment{}
		allPayments := []domain.Payment{}
		if isMember {
			if isOwner {
				pendingPayments, err = loadPendingPayments(app, planRecord.Id)
				if err != nil {
					return err
				}

				allPayments, err = loadAllPayments(app, planRecord.Id)
				if err != nil {
					return err
				}
			}

			userPayments, err = loadUserPayments(app, planRecord.Id, session.UserID)
			if err != nil {
				return err
			}
		}

		totalSavings := calculateTotalSavings(app, planRecord)
		planAgeDays := calculatePlanAgeDays(planRecord)

		userBalance := 0.0
		if isMember && !isOwner {
			userBalance, _ = billing.CalculateMemberBalance(app, planRecord.Id, session.UserID)
		}

		return view.RenderPage(c, "plan_details.html", map[string]interface{}{
			"title":              familyPlan.Name,
			"plan":               familyPlan,
			"is_owner":           isOwner,
			"is_member":          isMember,
			"members":            members,
			"total_members":      totalMembers,
			"join_requests":      joinRequests,
			"pending_request":    pendingRequest,
			"pending_payments":   pendingPayments,
			"user_payments":      userPayments,
			"user_balance":       userBalance,
			"existingMembership": existingMembership,
			"all_payments":       allPayments,
			"total_payments":     calculateTotalPayments(app, planRecord.Id),
			"total_savings":      totalSavings,
			"plan_age_days":      planAgeDays,
		})
	}
}
