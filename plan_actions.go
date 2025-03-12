package main

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/daos"
	"github.com/pocketbase/pocketbase/models"
)

// Handler for requesting to join a plan
func handleRequestJoin(app *pocketbase.PocketBase) echo.HandlerFunc {
	return func(c echo.Context) error {
		session := c.Get("session").(SessionData)
		joinCode := c.PathParam("join_code")

		// Find the plan by join code
		plansCollection, err := app.Dao().FindCollectionByNameOrId("family_plans")
		if err != nil {
			return err
		}

		plan, err := app.Dao().FindFirstRecordByData(plansCollection.Id, "join_code", joinCode)
		if err != nil || plan == nil {
			// Plan not found
			return c.Redirect(http.StatusSeeOther, "/family-plans")
		}

		// Create a join request
		joinRequestsCollection, err := app.Dao().FindCollectionByNameOrId("join_requests")
		if err != nil {
			return err
		}

		// Check if a request already exists
		existingRequest, _ := app.Dao().FindFirstRecordByFilter(joinRequestsCollection.Id,
			fmt.Sprintf("plan_id = '%s' && user_id = '%s'", plan.Id, session.UserId))

		if existingRequest == nil {
			// Create new request
			newRequest := models.NewRecord(joinRequestsCollection)
			newRequest.Set("plan_id", plan.Id)
			newRequest.Set("user_id", session.UserId)

			if err := app.Dao().SaveRecord(newRequest); err != nil {
				return err
			}
		}

		// Redirect back to plan page
		return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/%s", joinCode))
	}
}

// Handler for approving a join request
func handleApproveRequest(app *pocketbase.PocketBase) echo.HandlerFunc {
	return func(c echo.Context) error {
		session := c.Get("session").(SessionData)
		joinCode := c.PathParam("join_code")
		userId := c.FormValue("user_id")

		// Find the plan by join code
		plansCollection, err := app.Dao().FindCollectionByNameOrId("family_plans")
		if err != nil {
			return err
		}

		plan, err := app.Dao().FindFirstRecordByData(plansCollection.Id, "join_code", joinCode)
		if err != nil || plan == nil {
			// Plan not found
			return c.Redirect(http.StatusSeeOther, "/family-plans")
		}

		// Verify user is the plan owner
		ownerSlice := plan.GetStringSlice("owner")
		if len(ownerSlice) == 0 || ownerSlice[0] != session.UserId {
			// Not authorized
			return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/%s", joinCode))
		}

		// Find the join request
		joinRequestsCollection, err := app.Dao().FindCollectionByNameOrId("join_requests")
		if err != nil {
			return err
		}

		request, err := app.Dao().FindFirstRecordByFilter(joinRequestsCollection.Id,
			fmt.Sprintf("plan_id = '%s' && user_id = '%s'", plan.Id, userId))
		if err != nil || request == nil {
			// Request not found
			return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/%s", joinCode))
		}

		// Add user to memberships
		membershipsCollection, err := app.Dao().FindCollectionByNameOrId("memberships")
		if err != nil {
			return err
		}

		// Check if membership already exists
		existingMembership, _ := app.Dao().FindFirstRecordByFilter(membershipsCollection.Id,
			fmt.Sprintf("plan_id = '%s' && user_id = '%s'", plan.Id, userId))

		if existingMembership == nil {
			// Create new membership
			newMembership := models.NewRecord(membershipsCollection)
			newMembership.Set("plan_id", plan.Id)
			newMembership.Set("user_id", userId)
			newMembership.Set("is_artificial", false)

			if err := app.Dao().SaveRecord(newMembership); err != nil {
				return err
			}
		}

		// Delete the join request
		if err := app.Dao().DeleteRecord(request); err != nil {
			return err
		}

		// Redirect back to plan page
		return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/%s", joinCode))
	}
}

// Handler for denying a join request
func handleDenyRequest(app *pocketbase.PocketBase) echo.HandlerFunc {
	return func(c echo.Context) error {
		session := c.Get("session").(SessionData)
		joinCode := c.PathParam("join_code")
		userId := c.FormValue("user_id")

		// Find the plan by join code
		plansCollection, err := app.Dao().FindCollectionByNameOrId("family_plans")
		if err != nil {
			return err
		}

		plan, err := app.Dao().FindFirstRecordByData(plansCollection.Id, "join_code", joinCode)
		if err != nil || plan == nil {
			// Plan not found
			return c.Redirect(http.StatusSeeOther, "/family-plans")
		}

		// Verify user is the plan owner
		ownerSlice := plan.GetStringSlice("owner")
		if len(ownerSlice) == 0 || ownerSlice[0] != session.UserId {
			// Not authorized
			return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/%s", joinCode))
		}

		// Find the join request
		joinRequestsCollection, err := app.Dao().FindCollectionByNameOrId("join_requests")
		if err != nil {
			return err
		}

		request, err := app.Dao().FindFirstRecordByFilter(joinRequestsCollection.Id,
			fmt.Sprintf("plan_id = '%s' && user_id = '%s'", plan.Id, userId))
		if err != nil || request == nil {
			// Request not found
			return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/%s", joinCode))
		}

		// Delete the join request
		if err := app.Dao().DeleteRecord(request); err != nil {
			return err
		}

		// Redirect back to plan page
		return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/%s", joinCode))
	}
}

// Handler for removing a member from a plan
func handleRemoveMember(app *pocketbase.PocketBase) echo.HandlerFunc {
	return func(c echo.Context) error {
		session := c.Get("session").(SessionData)
		joinCode := c.PathParam("join_code")
		memberId := c.PathParam("member_id")

		// Find the plan by join code
		plansCollection, err := app.Dao().FindCollectionByNameOrId("family_plans")
		if err != nil {
			return err
		}

		plan, err := app.Dao().FindFirstRecordByData(plansCollection.Id, "join_code", joinCode)
		if err != nil || plan == nil {
			// Plan not found
			return c.Redirect(http.StatusSeeOther, "/family-plans")
		}

		// Check if user is the owner
		isOwner := false
		ownerSlice := plan.GetStringSlice("owner")
		if len(ownerSlice) > 0 && ownerSlice[0] == session.UserId {
			isOwner = true
		}

		if !isOwner {
			// Only owner can remove members
			return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/%s", joinCode))
		}

		// Get the membership
		membershipsCollection, err := app.Dao().FindCollectionByNameOrId("memberships")
		if err != nil {
			return err
		}

		membership, err := app.Dao().FindFirstRecordByFilter(membershipsCollection.Id,
			fmt.Sprintf("plan_id = '%s' && user_id = '%s'", plan.Id, memberId))
		if err != nil || membership == nil {
			// Membership not found
			return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/%s", joinCode))
		}

		// Update membership with date_ended instead of deleting
		membership.Set("date_ended", time.Now())
		if err := app.Dao().SaveRecord(membership); err != nil {
			return err
		}

		return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/%s", joinCode))
	}
}

// Handler for claiming a payment
func handleClaimPayment(app *pocketbase.PocketBase) echo.HandlerFunc {
	return func(c echo.Context) error {
		session := c.Get("session").(SessionData)
		joinCode := c.PathParam("join_code")

		// Find the plan by join code
		plansCollection, err := app.Dao().FindCollectionByNameOrId("family_plans")
		if err != nil {
			return err
		}

		plan, err := app.Dao().FindFirstRecordByData(plansCollection.Id, "join_code", joinCode)
		if err != nil || plan == nil {
			// Plan not found
			return c.Redirect(http.StatusSeeOther, "/family-plans")
		}

		// Check if user is a member
		membershipsCollection, err := app.Dao().FindCollectionByNameOrId("memberships")
		if err != nil {
			return err
		}

		existingMembership, _ := app.Dao().FindFirstRecordByFilter(membershipsCollection.Id,
			fmt.Sprintf("plan_id = '%s' && user_id = '%s'", plan.Id, session.UserId))

		isOwner := false
		ownerSlice := plan.GetStringSlice("owner")
		if len(ownerSlice) > 0 && ownerSlice[0] == session.UserId {
			isOwner = true
		}

		if !isOwner && existingMembership == nil {
			// Not a member
			return c.Redirect(http.StatusSeeOther, "/family-plans")
		}

		// Parse form data
		if err := c.Request().ParseForm(); err != nil {
			return err
		}

		amountStr := c.FormValue("amount")
		amount, err := strconv.ParseFloat(amountStr, 64)
		if err != nil || amount <= 0 {
			// Invalid amount
			return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/%s", joinCode))
		}

		notes := c.FormValue("notes")
		forMonthStr := c.FormValue("for_month")
		var forMonth string
		if forMonthStr != "" {
			// Parse and format the date
			forMonthDate, err := time.Parse("2006-01", forMonthStr)
			if err == nil {
				forMonth = forMonthDate.Format(time.RFC3339)
			}
		}

		// Create payment record
		paymentsCollection, err := app.Dao().FindCollectionByNameOrId("payments")
		if err != nil {
			return err
		}

		payment := models.NewRecord(paymentsCollection)
		payment.Set("plan_id", plan.Id)
		payment.Set("user_id", session.UserId)
		payment.Set("amount", amount)
		payment.Set("date", time.Now())
		payment.Set("status", "pending")
		payment.Set("notes", notes)
		if forMonth != "" {
			payment.Set("for_month", forMonth)
		}

		if err := app.Dao().SaveRecord(payment); err != nil {
			return err
		}

		return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/%s", joinCode))
	}
}

// Handler for approving a payment claim
func handleApprovePayment(app *pocketbase.PocketBase) echo.HandlerFunc {
	return func(c echo.Context) error {
		session := c.Get("session").(SessionData)
		joinCode := c.PathParam("join_code")
		paymentId := c.PathParam("payment_id")

		// Find the plan by join code
		plansCollection, err := app.Dao().FindCollectionByNameOrId("family_plans")
		if err != nil {
			return err
		}

		plan, err := app.Dao().FindFirstRecordByData(plansCollection.Id, "join_code", joinCode)
		if err != nil || plan == nil {
			// Plan not found
			return c.Redirect(http.StatusSeeOther, "/family-plans")
		}

		// Check if user is the owner
		isOwner := false
		ownerSlice := plan.GetStringSlice("owner")
		if len(ownerSlice) > 0 && ownerSlice[0] == session.UserId {
			isOwner = true
		}

		if !isOwner {
			// Only owner can approve payments
			return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/%s", joinCode))
		}

		// Get the payment
		paymentsCollection, err := app.Dao().FindCollectionByNameOrId("payments")
		if err != nil {
			return err
		}

		payment, err := app.Dao().FindRecordById(paymentsCollection.Id, paymentId)
		if err != nil || payment == nil {
			// Payment not found
			return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/%s", joinCode))
		}

		// Check if payment belongs to this plan
		if payment.GetString("plan_id") != plan.Id {
			// Payment doesn't belong to this plan
			return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/%s", joinCode))
		}

		// Check if payment is pending
		if payment.GetString("status") != "pending" {
			// Payment is not pending
			return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/%s", joinCode))
		}

		// Update payment status to approved
		payment.Set("status", "approved")
		if err := app.Dao().SaveRecord(payment); err != nil {
			return err
		}

		// If the user has requested to leave and their balance is now zero or positive,
		// update their date_ended
		userId := payment.GetString("user_id")
		balance, _ := calculateMemberBalance(app, plan.Id, userId)

		// Check if the user has requested to leave
		membershipsCollection, err := app.Dao().FindCollectionByNameOrId("memberships")
		if err != nil {
			return err
		}

		membership, err := app.Dao().FindFirstRecordByFilter(membershipsCollection.Id,
			fmt.Sprintf("plan_id = '%s' && user_id = '%s'", plan.Id, userId))
		if err == nil && membership != nil && membership.GetBool("leave_requested") && balance >= 0 {
			// Update membership with date_ended instead of deleting
			membership.Set("date_ended", time.Now())
			membership.Set("leave_requested", false) // Clear leave_requested flag
			if err := app.Dao().SaveRecord(membership); err != nil {
				return err
			}
		}

		return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/%s", joinCode))
	}
}

// Handler for rejecting a payment claim
func handleRejectPayment(app *pocketbase.PocketBase) echo.HandlerFunc {
	return func(c echo.Context) error {
		session := c.Get("session").(SessionData)
		joinCode := c.PathParam("join_code")

		// Find the plan by join code
		plansCollection, err := app.Dao().FindCollectionByNameOrId("family_plans")
		if err != nil {
			return err
		}

		plan, err := app.Dao().FindFirstRecordByData(plansCollection.Id, "join_code", joinCode)
		if err != nil || plan == nil {
			// Plan not found
			return c.Redirect(http.StatusSeeOther, "/family-plans")
		}

		// Check if user is the owner
		isOwner := false
		ownerSlice := plan.GetStringSlice("owner")
		if len(ownerSlice) > 0 && ownerSlice[0] == session.UserId {
			isOwner = true
		}

		if !isOwner {
			// Not the owner
			return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/%s", joinCode))
		}

		// Parse form data
		if err := c.Request().ParseForm(); err != nil {
			return err
		}

		paymentId := c.FormValue("payment_id")
		if paymentId == "" {
			return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/%s", joinCode))
		}

		// Find payment record
		paymentsCollection, err := app.Dao().FindCollectionByNameOrId("payments")
		if err != nil {
			return err
		}

		payment, err := app.Dao().FindRecordById(paymentsCollection.Id, paymentId)
		if err != nil || payment == nil {
			return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/%s", joinCode))
		}

		// Check if the payment belongs to this plan
		if payment.GetString("plan_id") != plan.Id {
			return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/%s", joinCode))
		}

		// Update payment status
		payment.Set("status", "rejected")
		if err := app.Dao().SaveRecord(payment); err != nil {
			return err
		}

		return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/%s", joinCode))
	}
}

// Handler for adding manual payment
func handleAddManualPayment(app *pocketbase.PocketBase) echo.HandlerFunc {
	return func(c echo.Context) error {
		session := c.Get("session").(SessionData)
		joinCode := c.PathParam("join_code")
		userId := c.FormValue("user_id")

		// Find the plan by join code
		plansCollection, err := app.Dao().FindCollectionByNameOrId("family_plans")
		if err != nil {
			return err
		}

		plan, err := app.Dao().FindFirstRecordByData(plansCollection.Id, "join_code", joinCode)
		if err != nil || plan == nil {
			// Plan not found
			return c.Redirect(http.StatusSeeOther, "/family-plans")
		}

		// Check if user is the owner
		isOwner := false
		ownerSlice := plan.GetStringSlice("owner")
		if len(ownerSlice) > 0 && ownerSlice[0] == session.UserId {
			isOwner = true
		}

		if !isOwner {
			// Only owner can add manual payments
			return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/%s", joinCode))
		}

		// Verify user is a member of the plan (or an artificial member)
		membershipsCollection, err := app.Dao().FindCollectionByNameOrId("memberships")
		if err != nil {
			return err
		}

		membership, err := app.Dao().FindFirstRecordByFilter(membershipsCollection.Id,
			fmt.Sprintf("plan_id = '%s' && user_id = '%s'", plan.Id, userId))
		if err != nil || membership == nil {
			// User is not a member
			return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/%s", joinCode))
		}

		// Parse payment amount
		amountStr := c.FormValue("amount")
		amount, err := strconv.ParseFloat(amountStr, 64)
		if err != nil || amount == 0 {
			// Invalid amount (can't be zero, but can be negative or positive)
			return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/%s", joinCode))
		}

		notes := c.FormValue("notes")
		forMonthStr := c.FormValue("for_month")
		var forMonth string
		if forMonthStr != "" {
			// Parse and format the date
			forMonthDate, err := time.Parse("2006-01", forMonthStr)
			if err == nil {
				forMonth = forMonthDate.Format(time.RFC3339)
			}
		}

		// Create payment record
		paymentsCollection, err := app.Dao().FindCollectionByNameOrId("payments")
		if err != nil {
			return err
		}

		payment := models.NewRecord(paymentsCollection)
		payment.Set("plan_id", plan.Id)
		payment.Set("user_id", userId)
		payment.Set("amount", amount)
		payment.Set("date", time.Now())
		payment.Set("status", "approved") // Manual payments are automatically approved
		payment.Set("notes", notes)
		if forMonth != "" {
			payment.Set("for_month", forMonth)
		}

		if err := app.Dao().SaveRecord(payment); err != nil {
			return err
		}

		// If the user has requested to leave and their balance is now zero or positive,
		// remove them from the plan
		balance, _ := calculateMemberBalance(app, plan.Id, userId)

		// Check if the user has requested to leave
		membershipsCollection, err = app.Dao().FindCollectionByNameOrId("memberships")
		if err != nil {
			return err
		}

		membership, err = app.Dao().FindFirstRecordByFilter(membershipsCollection.Id,
			fmt.Sprintf("plan_id = '%s' && user_id = '%s'", plan.Id, userId))
		if err == nil && membership != nil && membership.GetBool("leave_requested") && balance >= 0 {
			// Mark the membership as ended instead of deleting it
			membership.Set("date_ended", time.Now())
			membership.Set("leave_requested", false) // Clear leave_requested flag
			if err := app.Dao().SaveRecord(membership); err != nil {
				return err
			}
		}

		return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/%s", joinCode))
	}
}

// Function to calculate a member's balance
func calculateMemberBalance(app *pocketbase.PocketBase, planId, userId string) (float64, error) {
	// Get plan record to calculate monthly costs
	plansCollection, err := app.Dao().FindCollectionByNameOrId("family_plans")
	if err != nil {
		return 0, err
	}

	plan, err := app.Dao().FindRecordById(plansCollection.Id, planId)
	if err != nil {
		return 0, err
	}

	// Get the current plan cost
	monthlyCost := plan.GetFloat("cost")

	// Get membership to know when the user joined and if they've left
	membershipsCollection, err := app.Dao().FindCollectionByNameOrId("memberships")
	if err != nil {
		return 0, err
	}

	membership, err := app.Dao().FindFirstRecordByFilter(membershipsCollection.Id,
		fmt.Sprintf("plan_id = '%s' && user_id = '%s'", planId, userId))
	if err != nil || membership == nil {
		return 0, fmt.Errorf("membership not found")
	}

	// Get all user's approved payments
	paymentsCollection, err := app.Dao().FindCollectionByNameOrId("payments")
	if err != nil {
		return 0, err
	}

	userPayments, err := app.Dao().FindRecordsByFilter(paymentsCollection.Id,
		fmt.Sprintf("plan_id = '%s' && user_id = '%s' && status = 'approved'", planId, userId), "", 100, 0)
	if err != nil {
		return 0, err
	}

	// Calculate total amount user has already paid
	totalPaid := 0.0

	// Create a map to track payments by month for more accurate accounting
	paymentsByMonth := make(map[string]float64)

	for _, payment := range userPayments {
		amount := payment.GetFloat("amount")
		totalPaid += amount

		// If payment has a for_month, record it for detailed tracking
		forMonth := payment.GetDateTime("for_month")
		if !forMonth.IsZero() {
			monthKey := forMonth.Time().Format("2006-01")
			paymentsByMonth[monthKey] += amount
		}
	}

	// Calculate starting and ending months
	membershipStartDate := membership.GetDateTime("created").Time()
	currentDate := time.Now()

	// Check if membership has ended
	membershipEndDate := membership.GetDateTime("date_ended")
	if !membershipEndDate.IsZero() {
		// Use the end date instead of current date if membership has ended
		currentDate = membershipEndDate.Time()
	}

	// Initialize the start and end months
	// If they join partway through a month, they pay for the whole month
	startMonth := time.Date(membershipStartDate.Year(), membershipStartDate.Month(), 1, 0, 0, 0, 0, membershipStartDate.Location())

	// If they leave partway through a month, they pay for the whole month
	var endMonth time.Time
	if !membershipEndDate.IsZero() {
		// If they've left, use the month they left
		endMonth = time.Date(membershipEndDate.Time().Year(), membershipEndDate.Time().Month(), 1, 0, 0, 0, 0, membershipEndDate.Time().Location())
	} else {
		// Otherwise use the current month
		endMonth = time.Date(currentDate.Year(), currentDate.Month(), 1, 0, 0, 0, 0, currentDate.Location())
	}

	// Calculate total amount due, month by month
	amountDue := 0.0
	currentMonth := startMonth

	// Process each month
	for !currentMonth.After(endMonth) {
		monthKey := currentMonth.Format("2006-01")

		// Get active memberships for this month
		activeMemberships, err := getActiveMembershipsForMonth(app, planId, currentMonth)
		if err == nil {
			// Get the owner ID
			plan, err := app.Dao().FindRecordById(plansCollection.Id, planId)
			if err != nil {
				continue
			}
			ownerSlice := plan.GetStringSlice("owner")
			ownerId := ""
			if len(ownerSlice) > 0 {
				ownerId = ownerSlice[0]
			}

			// Count members but avoid double-counting the owner
			memberCount := 0
			ownerIncluded := false

			// Check if owner is already included in memberships
			for _, m := range activeMemberships {
				if m.GetString("user_id") == ownerId {
					ownerIncluded = true
				}
				memberCount++
			}

			// Only add 1 for owner if not already included
			if !ownerIncluded {
				memberCount++
			}

			if memberCount > 0 {
				// Calculate monthly share - every active member pays the full share for the month
				// regardless of when they joined or left during the month
				monthlyShare := monthlyCost / float64(memberCount)
				amountDue += monthlyShare
			}
		}

		// Subtract any payments specifically for this month
		if paidAmount, exists := paymentsByMonth[monthKey]; exists {
			amountDue -= paidAmount
			// Remove this from totalPaid since we've accounted for it
			totalPaid -= paidAmount
		}

		// Move to the next month
		currentMonth = time.Date(currentMonth.Year(), currentMonth.Month()+1, 1, 0, 0, 0, 0, currentMonth.Location())
	}

	// Final balance calculation
	// Positive balance means user has paid more than they owe
	// Negative balance means user still owes money
	return totalPaid - amountDue, nil
}

// Helper function to get all active memberships for a specific month
func getActiveMembershipsForMonth(app *pocketbase.PocketBase, planId string, targetMonth time.Time) ([]*models.Record, error) {
	membershipsCollection, err := app.Dao().FindCollectionByNameOrId("memberships")
	if err != nil {
		return nil, err
	}

	// Get all memberships for this plan
	allMemberships, err := app.Dao().FindRecordsByFilter(membershipsCollection.Id,
		fmt.Sprintf("plan_id = '%s'", planId), "", 100, 0)
	if err != nil {
		return nil, err
	}

	// Filter to include only memberships that were active during the target month
	activeMemberships := make([]*models.Record, 0)

	// Calculate the first day of the target month
	monthStart := time.Date(targetMonth.Year(), targetMonth.Month(), 1, 0, 0, 0, 0, targetMonth.Location())

	for _, membership := range allMemberships {
		// Get membership creation date
		createdAt := membership.GetDateTime("created").Time()

		// Get the first day of the month the membership was created
		membershipMonth := time.Date(createdAt.Year(), createdAt.Month(), 1, 0, 0, 0, 0, createdAt.Location())

		// Skip if the membership was created in a month after the target month
		if membershipMonth.After(monthStart) {
			continue
		}

		// Check if this membership has ended
		dateEndedTime := membership.GetDateTime("date_ended").Time()
		hasEnded := !membership.GetDateTime("date_ended").IsZero()

		// If membership has ended, get the first day of the month it ended
		var endedMonth time.Time
		if hasEnded {
			endedMonth = time.Date(dateEndedTime.Year(), dateEndedTime.Month(), 1, 0, 0, 0, 0, dateEndedTime.Location())

			// If the membership ended before the target month, it wasn't active
			if endedMonth.Before(monthStart) {
				continue
			}
		}

		// If we got here, the membership was active during the target month
		activeMemberships = append(activeMemberships, membership)
	}

	return activeMemberships, nil
}

// Handler for requesting to leave a plan (modified to handle balance)
func handleLeavePlan(app *pocketbase.PocketBase) echo.HandlerFunc {
	return func(c echo.Context) error {
		session := c.Get("session").(SessionData)
		joinCode := c.PathParam("join_code")

		// Find the plan by join code
		plansCollection, err := app.Dao().FindCollectionByNameOrId("family_plans")
		if err != nil {
			return err
		}

		plan, err := app.Dao().FindFirstRecordByData(plansCollection.Id, "join_code", joinCode)
		if err != nil || plan == nil {
			// Plan not found
			return c.Redirect(http.StatusSeeOther, "/family-plans")
		}

		// Check if user is a member
		membershipsCollection, err := app.Dao().FindCollectionByNameOrId("memberships")
		if err != nil {
			return err
		}

		// Check if the user is the owner
		isOwner := false
		ownerSlice := plan.GetStringSlice("owner")
		if len(ownerSlice) > 0 && ownerSlice[0] == session.UserId {
			isOwner = true
		}

		if isOwner {
			// Owner cannot leave their own plan
			return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/%s", joinCode))
		}

		existingMembership, _ := app.Dao().FindFirstRecordByFilter(membershipsCollection.Id,
			fmt.Sprintf("plan_id = '%s' && user_id = '%s'", plan.Id, session.UserId))

		if existingMembership == nil {
			// Not a member
			return c.Redirect(http.StatusSeeOther, "/family-plans")
		}

		// Calculate user's balance
		balance, err := calculateMemberBalance(app, plan.Id, session.UserId)

		if err == nil && balance >= 0 {
			// If balance is zero or positive, mark as left with date_ended
			existingMembership.Set("date_ended", time.Now())
			existingMembership.Set("leave_requested", false) // Clear leave_requested flag
			if err := app.Dao().SaveRecord(existingMembership); err != nil {
				return err
			}
		} else {
			// Otherwise mark as requested to leave (negative balance)
			existingMembership.Set("leave_requested", true)
			if err := app.Dao().SaveRecord(existingMembership); err != nil {
				return err
			}
		}

		return c.Redirect(http.StatusSeeOther, "/family-plans")
	}
}

// Handler for deleting a plan (for owner)
func handleDeletePlan(app *pocketbase.PocketBase) echo.HandlerFunc {
	return func(c echo.Context) error {
		session := c.Get("session").(SessionData)
		joinCode := c.PathParam("join_code")

		// Find the plan by join code
		plansCollection, err := app.Dao().FindCollectionByNameOrId("family_plans")
		if err != nil {
			return err
		}

		plan, err := app.Dao().FindFirstRecordByData(plansCollection.Id, "join_code", joinCode)
		if err != nil || plan == nil {
			// Plan not found
			return c.Redirect(http.StatusSeeOther, "/family-plans")
		}

		// Verify user is the plan owner
		ownerSlice := plan.GetStringSlice("owner")
		if len(ownerSlice) == 0 || ownerSlice[0] != session.UserId {
			// Not authorized
			return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/%s", joinCode))
		}

		// Use a transaction to ensure all deletions succeed or fail together
		err = app.Dao().RunInTransaction(func(txDao *daos.Dao) error {
			// Delete all related memberships first
			membershipsCollection, err := txDao.FindCollectionByNameOrId("memberships")
			if err != nil {
				return err
			}

			// Use -1 for no limit to ensure we get all records
			memberships, err := txDao.FindRecordsByFilter(membershipsCollection.Id,
				fmt.Sprintf("plan_id = '%s'", plan.Id), "", -1, 0)
			if err == nil {
				for _, membership := range memberships {
					if err := txDao.DeleteRecord(membership); err != nil {
						return err
					}
				}
			}

			// Delete all related join requests
			joinRequestsCollection, err := txDao.FindCollectionByNameOrId("join_requests")
			if err != nil {
				return err
			}

			// Use -1 for no limit to ensure we get all records
			joinRequests, err := txDao.FindRecordsByFilter(joinRequestsCollection.Id,
				fmt.Sprintf("plan_id = '%s'", plan.Id), "", -1, 0)
			if err == nil {
				for _, request := range joinRequests {
					if err := txDao.DeleteRecord(request); err != nil {
						return err
					}
				}
			}

			// Finally delete the plan itself
			return txDao.DeleteRecord(plan)
		})

		if err != nil {
			// If there was an error, redirect with error message
			return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/%s?error=Failed+to+delete+plan", joinCode))
		}

		// Redirect to family plans page
		return c.Redirect(http.StatusSeeOther, "/family-plans")
	}
}

// Handler for updating plan details
func handleUpdatePlan(app *pocketbase.PocketBase) echo.HandlerFunc {
	return func(c echo.Context) error {
		session := c.Get("session").(SessionData)
		joinCode := c.PathParam("join_code")

		// Get the family plan
		plansCollection, err := app.Dao().FindCollectionByNameOrId("family_plans")
		if err != nil || plansCollection == nil {
			return c.Redirect(http.StatusSeeOther, "/family-plans")
		}

		plan, err := app.Dao().FindFirstRecordByData(plansCollection.Id, "join_code", joinCode)
		if err != nil || plan == nil {
			return c.Redirect(http.StatusSeeOther, "/family-plans")
		}

		// Check if user is the owner
		ownerSlice := plan.GetStringSlice("owner")
		if len(ownerSlice) == 0 || ownerSlice[0] != session.UserId {
			return c.Redirect(http.StatusSeeOther, "/family-plans")
		}

		// Get form values
		name := c.FormValue("name")
		description := c.FormValue("description")
		costStr := c.FormValue("cost")
		individualCostStr := c.FormValue("individual_cost")

		// Parse cost
		cost, err := strconv.ParseFloat(costStr, 64)
		if err != nil {
			// Use current cost if error
			cost = plan.GetFloat("cost")
		}

		// Parse individual cost
		individualCost, err := strconv.ParseFloat(individualCostStr, 64)
		if err != nil {
			// Use current individual cost if error
			individualCost = plan.GetFloat("individual_cost")
		}

		// Update plan
		plan.Set("name", name)
		plan.Set("description", description)
		plan.Set("cost", cost)
		plan.Set("individual_cost", individualCost)

		// Save changes
		if err := app.Dao().SaveRecord(plan); err != nil {
			return err
		}

		// Redirect back to plan details
		return c.Redirect(http.StatusSeeOther, "/"+joinCode)
	}
}

// Handler for adding an artificial member
func handleAddArtificialMember(app *pocketbase.PocketBase) echo.HandlerFunc {
	return func(c echo.Context) error {
		session := c.Get("session").(SessionData)
		joinCode := c.PathParam("join_code")
		memberName := c.FormValue("name")

		if memberName == "" {
			// Name is required
			return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/%s", joinCode))
		}

		// Find the plan by join code
		plansCollection, err := app.Dao().FindCollectionByNameOrId("family_plans")
		if err != nil {
			return err
		}

		plan, err := app.Dao().FindFirstRecordByData(plansCollection.Id, "join_code", joinCode)
		if err != nil || plan == nil {
			// Plan not found
			return c.Redirect(http.StatusSeeOther, "/family-plans")
		}

		// Check if user is the owner
		isOwner := false
		ownerSlice := plan.GetStringSlice("owner")
		if len(ownerSlice) > 0 && ownerSlice[0] == session.UserId {
			isOwner = true
		}

		if !isOwner {
			// Only owner can add artificial members
			return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/%s", joinCode))
		}

		// Create a new artificial membership
		membershipsCollection, err := app.Dao().FindCollectionByNameOrId("memberships")
		if err != nil {
			return err
		}

		// Generate a unique ID for the artificial member
		artificialUserId := fmt.Sprintf("artificial_%s_%d", plan.Id, time.Now().UnixNano())

		// Create the membership record
		newMembership := models.NewRecord(membershipsCollection)
		newMembership.Set("plan_id", plan.Id)
		newMembership.Set("user_id", artificialUserId)
		newMembership.Set("is_artificial", true)
		newMembership.Set("name", memberName)

		if err := app.Dao().SaveRecord(newMembership); err != nil {
			return err
		}

		// Redirect back to plan page
		return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/%s", joinCode))
	}
}

// Handler for transferring membership from real user to artificial user
func handleTransferMembership(app *pocketbase.PocketBase) echo.HandlerFunc {
	return func(c echo.Context) error {
		session := c.Get("session").(SessionData)
		joinCode := c.PathParam("join_code")
		realUserId := c.FormValue("user_id")
		artificialMemberId := c.FormValue("artificial_member_id")

		if realUserId == "" || artificialMemberId == "" {
			// Required fields
			return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/%s", joinCode))
		}

		// Find the plan by join code
		plansCollection, err := app.Dao().FindCollectionByNameOrId("family_plans")
		if err != nil {
			return err
		}

		plan, err := app.Dao().FindFirstRecordByData(plansCollection.Id, "join_code", joinCode)
		if err != nil || plan == nil {
			// Plan not found
			return c.Redirect(http.StatusSeeOther, "/family-plans")
		}

		// Check if user is the owner
		isOwner := false
		ownerSlice := plan.GetStringSlice("owner")
		if len(ownerSlice) > 0 && ownerSlice[0] == session.UserId {
			isOwner = true
		}

		if !isOwner {
			// Only owner can transfer memberships
			return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/%s", joinCode))
		}

		// Get memberships collection
		membershipsCollection, err := app.Dao().FindCollectionByNameOrId("memberships")
		if err != nil {
			return err
		}

		// Get the artificial membership
		artificialMembership, err := app.Dao().FindFirstRecordByFilter(membershipsCollection.Id,
			fmt.Sprintf("plan_id = '%s' && user_id = '%s' && is_artificial = true", plan.Id, artificialMemberId))
		if err != nil || artificialMembership == nil {
			// Artificial membership not found
			return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/%s", joinCode))
		}

		// Find the join request
		joinRequestsCollection, err := app.Dao().FindCollectionByNameOrId("join_requests")
		if err != nil {
			return err
		}

		request, err := app.Dao().FindFirstRecordByFilter(joinRequestsCollection.Id,
			fmt.Sprintf("plan_id = '%s' && user_id = '%s'", plan.Id, realUserId))
		if err != nil || request == nil {
			// Request not found
			return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/%s", joinCode))
		}

		// Get payments collection
		paymentsCollection, err := app.Dao().FindCollectionByNameOrId("payments")
		if err != nil {
			return err
		}

		// Update all payments from artificial member to real member
		artificialPayments, err := app.Dao().FindRecordsByFilter(paymentsCollection.Id,
			fmt.Sprintf("plan_id = '%s' && user_id = '%s'", plan.Id, artificialMemberId), "", 100, 0)
		if err == nil {
			for _, payment := range artificialPayments {
				// Update payment to new user ID
				payment.Set("user_id", realUserId)
				_ = app.Dao().SaveRecord(payment)
			}
		}

		// Delete the artificial membership
		if err := app.Dao().DeleteRecord(artificialMembership); err != nil {
			return err
		}

		// Create new membership for the real user
		newMembership := models.NewRecord(membershipsCollection)
		newMembership.Set("plan_id", plan.Id)
		newMembership.Set("user_id", realUserId)
		newMembership.Set("is_artificial", false)

		if err := app.Dao().SaveRecord(newMembership); err != nil {
			return err
		}

		// Delete the join request
		if err := app.Dao().DeleteRecord(request); err != nil {
			return err
		}

		// Redirect back to plan page
		return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/%s", joinCode))
	}
}
