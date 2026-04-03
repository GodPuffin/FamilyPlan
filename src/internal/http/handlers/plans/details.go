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
		if err != nil {
			return err
		}
		if planRecord == nil {
			return view.RenderPage(c, "plan_details.html", map[string]interface{}{
				"title":     "Plan Not Found",
				"not_found": true,
				"plan":      nil,
			})
		}

		isOwner := planutil.IsOwner(planRecord, session.UserID)
		existingMembership, err := planutil.FindMembership(app, planRecord.Id, session.UserID)
		if err != nil {
			return err
		}
		isMember := existingMembership != nil || isOwner

		pendingRequest := false
		if !isMember {
			existingRequest, err := planutil.FindJoinRequest(app, planRecord.Id, session.UserID)
			if err != nil {
				return err
			}
			pendingRequest = existingRequest != nil
		}

		familyPlan := buildFamilyPlan(planRecord, 0, 0)

		members := []domain.Member{}
		totalMembers := 0
		claimLinks := map[string]string{}
		if isMember {
			members, totalMembers, err = loadMembers(app, familyPlan)
			if err != nil {
				return err
			}

			if isOwner {
				claimLinks, err = loadMemberClaimLinks(app, planRecord.Id, c.Scheme(), c.Request().Host)
				if err != nil {
					return err
				}
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
		memberPaymentsPagination := buildMemberPaymentsPagination(1, false)
		if isMember {
			if isOwner {
				pendingPayments, err = loadPendingPayments(app, planRecord.Id)
				if err != nil {
					return err
				}

				allPayments, memberPaymentsPagination, err = loadAllPaymentsPage(
					app,
					planRecord.Id,
					memberPaymentsPage(c.QueryParam(memberPaymentsPageParam)),
					memberPaymentsPageSize,
				)
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
			"title":                      familyPlan.Name,
			"plan":                       familyPlan,
			"is_owner":                   isOwner,
			"is_member":                  isMember,
			"members":                    members,
			"claim_links":                claimLinks,
			"total_members":              totalMembers,
			"join_requests":              joinRequests,
			"pending_request":            pendingRequest,
			"pending_payments":           pendingPayments,
			"user_payments":              userPayments,
			"user_balance":               userBalance,
			"existingMembership":         existingMembership,
			"all_payments":               allPayments,
			"member_payments_pagination": memberPaymentsPagination,
			"total_payments":             calculateTotalPayments(app, planRecord.Id),
			"total_savings":              totalSavings,
			"plan_age_days":              planAgeDays,
		})
	}
}
